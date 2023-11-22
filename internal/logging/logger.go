package logging

import (
	"context"
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

var loggerContextKey contextKey = "logging"

func New(w io.Writer, level zapcore.Level) (*zap.Logger, zap.AtomicLevel) {
	atomicLevel := zap.NewAtomicLevelAt(level)

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(w),
		atomicLevel,
	)

	logger := zap.New(core)

	return logger, atomicLevel
}

func WithValue(ctx context.Context, logger *zap.Logger) context.Context {
	loggerCtx := context.WithValue(ctx, loggerContextKey, logger)
	return loggerCtx
}

func FromContext(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(loggerContextKey).(*zap.Logger)
	if !ok || logger == nil {
		return zap.NewNop()
	}

	return logger
}
