package slogdiscard

import (
	"context"
	"log/slog"
)

type DiscardHandler struct {
}

func NewDiscardLogger() *slog.Logger {
	return slog.New(NewDiscardHandler())
}

func NewDiscardHandler() *DiscardHandler {
	return &DiscardHandler{}
}

func (d *DiscardHandler) Enabled(ctx context.Context, level slog.Level) bool {
	_ = ctx
	_ = level
	return false
}

func (d *DiscardHandler) Handle(ctx context.Context, record slog.Record) error {
	_ = ctx
	_ = record
	return nil
}

func (d *DiscardHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	_ = attrs
	return d
}

func (d *DiscardHandler) WithGroup(name string) slog.Handler {
	_ = name
	return d
}
