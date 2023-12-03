package zapctx

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

var contextKeyLogger contextKey = "zap-logger"

func Context(ctx context.Context, logger *zap.Logger) context.Context {
	loggerCtx := context.WithValue(ctx, contextKeyLogger, logger)
	return loggerCtx
}

func Value(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(contextKeyLogger).(*zap.Logger)
	if !ok || logger == nil {
		return zap.NewNop()
	}

	return logger
}

func L(ctx context.Context) *zap.Logger {
	return Value(ctx)
}
