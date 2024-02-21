package helpers

import (
	consts "check_system/config"
	"context"

	"go.uber.org/zap"
)

// WithContext returns a new context with the logger added.
func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
    return context.WithValue(ctx, consts.LoggerCtxName, logger)
}

// FromContext returns the logger in the context if it exists, otherwise a new logger is returned.
func FromContext(ctx context.Context) *zap.Logger {
    logger := ctx.Value(consts.LoggerCtxName)
    if l, ok := logger.(*zap.Logger); ok {
        return l
    }
    return zap.Must(zap.NewDevelopment())
}