package slogctx_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/somebadcode/commit-tool/slogctx"
	"github.com/somebadcode/commit-tool/slognop"
)

func TestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	want := slognop.New()

	ctxWithLogger := slogctx.Context(ctx, want)

	if got := slogctx.L(ctxWithLogger); !reflect.DeepEqual(got, want) {
		t.Errorf("Context() = %v, want %v", got, want)
	}
}

func TestContextNop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	if got := slogctx.L(ctx); !reflect.DeepEqual(got, slognop.New()) {
		t.Errorf("Context() = %v, want %v", got, slognop.New())
	}
}
