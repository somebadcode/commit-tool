package slogctx

import (
	"context"
	"log/slog"

	"github.com/somebadcode/commit-tool/slognop"
)

type contextKey string

const contextKeyLogger contextKey = "slog"

func Context(ctx context.Context, logger *slog.Logger) context.Context {
	loggerCtx := context.WithValue(ctx, contextKeyLogger, logger)
	return loggerCtx
}

func Logger(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(contextKeyLogger).(*slog.Logger)
	if !ok || logger == nil {
		return slog.New(slognop.Handler{})
	}

	return logger
}

func L(ctx context.Context) *slog.Logger {
	return Logger(ctx)
}
