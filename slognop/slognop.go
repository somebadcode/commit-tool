package slognop

import (
	"context"
	"log/slog"
)

type Handler struct{}

var _ slog.Handler = (*Handler)(nil)

func New() *slog.Logger {
	return slog.New(Handler{})
}

func (n Handler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}

func (n Handler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

func (n Handler) WithAttrs(_ []slog.Attr) slog.Handler {
	return n
}

func (n Handler) WithGroup(_ string) slog.Handler {
	return n
}
