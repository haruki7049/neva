package analyzer

import (
	"errors"
	"fmt"

	src "github.com/nevalang/neva/internal/compiler/sourcecode"
)

var (
	ErrEmptyConst         = errors.New("Constant must either have value or reference to another constant")
	ErrEntityNotConst     = errors.New("Constant refers to an entity that is not constant")
	ErrResolveConstType   = errors.New("Cannot resolve constant type")
	ErrConstSeveralValues = errors.New("Constant cannot have several values at once")
)

//nolint:funlen
func (a Analyzer) analyzeConst(constant src.Const, scope src.Scope) (src.Const, *Error) {
	if constant.Value == nil && constant.Ref == nil {
		return src.Const{}, &Error{
			Err:      ErrEmptyConst,
			Location: &scope.Location,
			Meta:     &constant.Meta,
		}
	}

	if constant.Value == nil { // is ref
		entity, location, err := scope.Entity(*constant.Ref)
		if err != nil {
			return src.Const{}, &Error{
				Err:      err,
				Location: &location,
				Meta:     entity.Meta(),
			}
		}

		if entity.Kind != src.ConstEntity {
			return src.Const{}, &Error{
				Err:      fmt.Errorf("%w: entity kind %v", ErrEntityNotConst, entity.Kind),
				Location: &location,
				Meta:     entity.Meta(),
			}
		}
	}

	resolvedType, err := a.analyzeTypeExpr(constant.Value.TypeExpr, scope)
	if err != nil {
		return src.Const{}, Error{
			Err:      ErrResolveConstType,
			Location: &scope.Location,
			Meta:     &constant.Meta,
		}.Merge(err)
	}

	switch resolvedType.Inst.Ref.String() {
	case "bool":
		if constant.Value.Int != 0 || constant.Value.Float != 0 || constant.Value.Str != "" {
			return src.Const{}, &Error{
				Err:      ErrConstSeveralValues,
				Location: &scope.Location,
				Meta:     &constant.Meta,
			}
		}
	case "int":
		if constant.Value.Bool != false || constant.Value.Float != 0 || constant.Value.Str != "" {
			return src.Const{}, &Error{
				Err:      ErrConstSeveralValues,
				Location: &scope.Location,
				Meta:     &constant.Meta,
			}
		}
	case "float":
		if constant.Value.Bool != false || constant.Value.Int != 0 || constant.Value.Str != "" {
			return src.Const{}, &Error{
				Err:      ErrConstSeveralValues,
				Location: &scope.Location,
				Meta:     &constant.Meta,
			}
		}
	case "str":
		if constant.Value.Bool != false || constant.Value.Int != 0 || constant.Value.Float != 0 {
			return src.Const{}, &Error{
				Err:      ErrConstSeveralValues,
				Location: &scope.Location,
				Meta:     &constant.Meta,
			}
		}
	}

	valueCopy := *constant.Value
	valueCopy.TypeExpr = resolvedType

	return src.Const{
		Value: &valueCopy,
	}, nil
}
