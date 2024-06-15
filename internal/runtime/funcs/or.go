package funcs

import (
	"context"

	"github.com/nevalang/neva/internal/runtime"
)

type or struct{}

func (p or) Create(io runtime.FuncIO, _ runtime.Msg) (func(ctx context.Context), error) {
	aIn, err := io.In.Single("a")
	if err != nil {
		return nil, err
	}

	bIn, err := io.In.Single("b")
	if err != nil {
		return nil, err
	}

	resOut, err := io.Out.SingleOutport("res")
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context) {
		for {
			aMsg, ok := aIn.Receive(ctx)
			if !ok {
				return
			}

			bMsg, ok := bIn.Receive(ctx)
			if !ok {
				return
			}

			if !resOut.Send(
				ctx,
				runtime.NewBoolMsg(
					aMsg.Bool() || bMsg.Bool(),
				),
			) {
				return
			}
		}
	}, nil
}
