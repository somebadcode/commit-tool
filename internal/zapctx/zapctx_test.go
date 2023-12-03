package zapctx_test

import (
	"context"
	"reflect"
	"testing"

	"go.uber.org/zap"

	"github.com/somebadcode/commit-tool/internal/zapctx"
)

func TestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	want := zap.NewNop()

	ctxWithLogger := zapctx.Context(ctx, want)

	if got := zapctx.L(ctxWithLogger); !reflect.DeepEqual(got, want) {
		t.Errorf("Context() = %v, want %v", got, want)
	}
}

func TestContextNop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	if got := zapctx.L(ctx); !reflect.DeepEqual(got, zap.NewNop()) {
		t.Errorf("Context() = %v, want %v", got, zap.NewNop())
	}
}
