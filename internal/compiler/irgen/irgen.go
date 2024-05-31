package irgen

import (
	"errors"
	"fmt"

	"github.com/nevalang/neva/internal/compiler"
	src "github.com/nevalang/neva/internal/compiler/sourcecode"
	"github.com/nevalang/neva/internal/compiler/sourcecode/core"
	"github.com/nevalang/neva/internal/runtime/ir"
)

var ErrNodeUsageNotFound = errors.New("node usage not found")

type Generator struct{}

type (
	nodeContext struct {
		path       []string
		node       src.Node
		portsUsage portsUsage
	}

	portsUsage struct {
		in  map[relPortAddr]struct{}
		out map[relPortAddr]struct{}
	}

	relPortAddr struct {
		Port string
		Idx  uint8
	}
)

func (g Generator) Generate(
	build src.Build,
	mainPkgName string,
) (*ir.Program, *compiler.Error) {
	scope := src.Scope{
		Build: build,
		Location: src.Location{
			ModRef:   build.EntryModRef,
			PkgName:  mainPkgName,
			FileName: "",
		},
	}

	result := &ir.Program{
		Ports:       map[ir.PortAddr]struct{}{},
		Connections: map[ir.PortAddr]map[ir.PortAddr]struct{}{},
		Funcs:       []ir.FuncCall{},
	}

	rootNodeCtx := nodeContext{
		path: []string{},
		node: src.Node{
			EntityRef: core.EntityRef{
				Pkg:  "",
				Name: "Main",
			},
		},
		portsUsage: portsUsage{
			in: map[relPortAddr]struct{}{
				{Port: "start"}: {},
			},
			out: map[relPortAddr]struct{}{
				{Port: "stop"}: {},
			},
		},
	}

	if err := g.processNode(rootNodeCtx, scope, result); err != nil {
		return nil, compiler.Error{
			Location: &scope.Location,
		}.Wrap(err)
	}

	reducedPorts, reducerNet := reduceGraph(result)

	return &ir.Program{
		Ports:       reducedPorts,
		Connections: reducerNet,
		Funcs:       result.Funcs,
	}, nil
}

func (g Generator) processNode(
	nodeCtx nodeContext,
	scope src.Scope,
	result *ir.Program,
) *compiler.Error {
	flowEntity, foundLocation, err := scope.Entity(nodeCtx.node.EntityRef)
	if err != nil {
		return &compiler.Error{
			Err:      err,
			Location: &scope.Location,
		}
	}

	flow := flowEntity.Flow

	runtimeFuncRef, err := getFuncRef(flow, nodeCtx.node.TypeArgs)
	if err != nil {
		return &compiler.Error{
			Err:      err,
			Location: &foundLocation,
			Meta:     &flow.Meta,
		}
	}

	// if flow uses #extern, then we only need ports and func call
	// ports are already created, so it's time to create func call
	if runtimeFuncRef != "" {
		// use prev location, not the location where runtime func was found
		runtimeFuncMsg, err := getCfgMsg(nodeCtx.node, scope)
		if err != nil {
			return &compiler.Error{
				Err:      err,
				Location: &scope.Location,
			}
		}

		// for in we only use parent ctx cuz all inports are used
		inportAddrs := g.getFuncInports(nodeCtx)
		//  for out we use both parent ctx and interface
		outportAddrs := g.getFuncOutports(nodeCtx)

		result.Funcs = append(result.Funcs, ir.FuncCall{
			Ref: runtimeFuncRef,
			IO: ir.FuncIO{
				In:  inportAddrs,
				Out: outportAddrs,
			},
			Msg: runtimeFuncMsg,
		})

		return nil
	}

	newScope := scope.WithLocation(foundLocation) // only use new location if that's not builtin

	// We use network as a source of true about how subnodes ports instead subnodes interface definitions.
	// We cannot rely on them because there's no information about how many array slots are used (in case of array ports).
	// On the other hand, we believe network has everything we need because program' correctness is verified by analyzer.
	subnodesPortsUsage, err := g.processNetwork(
		flow.Net,
		nodeCtx,
		result,
	)
	if err != nil {
		return &compiler.Error{
			Err:      err,
			Location: &newScope.Location,
		}
	}

	for nodeName, node := range flow.Nodes {
		nodePortsUsage, ok := subnodesPortsUsage[nodeName]
		if !ok {
			return &compiler.Error{
				Err:      fmt.Errorf("%w: %v", ErrNodeUsageNotFound, nodeName),
				Location: &foundLocation,
				Meta:     &flow.Meta,
			}
		}

		subNodeCtx := nodeContext{
			path:       append(nodeCtx.path, nodeName),
			portsUsage: nodePortsUsage,
			node:       node,
		}

		scopeToUse := getSubnodeScope(
			nodeCtx.node.Deps,
			nodeName,
			subNodeCtx,
			scope,
			newScope,
		)

		if err := g.processNode(subNodeCtx, scopeToUse, result); err != nil {
			return &compiler.Error{
				Err:      fmt.Errorf("%w: node '%v'", err, nodeName),
				Location: &foundLocation,
				Meta:     &flow.Meta,
			}
		}
	}

	return nil
}

func getSubnodeScope(
	deps map[string]src.Node,
	nodeName string,
	subNodeCtx nodeContext,
	scope src.Scope,
	newScope src.Scope,
) src.Scope {
	var scopeToUseThisTime src.Scope
	if dep, ok := deps[nodeName]; ok { // is interface node
		subNodeCtx.node = dep
		scopeToUseThisTime = scope
	} else {
		scopeToUseThisTime = newScope
	}
	return scopeToUseThisTime
}

func New() Generator {
	return Generator{}
}
