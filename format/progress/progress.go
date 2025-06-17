package progress

import (
	"context"
	"io"
)

type progressKeyType struct{}

var progressKey progressKeyType

func WithProgressWriter(ctx context.Context, w io.Writer) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if w == nil {
		w = io.Discard
	}
	return context.WithValue(ctx, progressKey, w)
}

func ProgressWriter(ctx context.Context) io.Writer {
	w, ok := ctx.Value(progressKey).(io.Writer)
	if !ok {
		return io.Discard
	}
	return w
}
