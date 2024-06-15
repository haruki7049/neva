package funcs

import (
	"context"
	"os"

	"github.com/nevalang/neva/internal/runtime"
)

type writeAll struct{}

func (c writeAll) Create(rio runtime.FuncIO, msg runtime.Msg) (func(ctx context.Context), error) {
	filename, err := rio.In.Single("filename")
	if err != nil {
		return nil, err
	}

	dataPort, err := rio.In.Single("data")
	if err != nil {
		return nil, err
	}

	sig, err := rio.Out.SingleOutport("sig")
	if err != nil {
		return nil, err
	}

	errPort, err := rio.Out.SingleOutport("err")
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context) {
		for {
			name, ok := filename.Receive(ctx)
			if !ok {
				return
			}

			data, ok := dataPort.Receive(ctx)
			if !ok {
				return
			}

			err := os.WriteFile(name.Str(), []byte(data.Str()), 0755)
			if err != nil {
				if !errPort.Send(ctx, errFromErr(err)) {
					return
				}
				continue
			}

			if !sig.Send(ctx, nil) {
				return
			}
		}
	}, nil
}
