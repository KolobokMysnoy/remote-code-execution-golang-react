package main

import (
	consts "check_system/config"
	"check_system/internal/docker/delivery"
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type key struct{}

// WithContext returns a new context with the logger added.
func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
    return context.WithValue(ctx, key{}, logger)
}

// FromContext returns the logger in the context if it exists, otherwise a new logger is returned.
func FromContext(ctx context.Context) *zap.Logger {
    logger := ctx.Value(key{})
    if l, ok := logger.(*zap.Logger); ok {
        return l
    }
    return zap.Must(zap.NewDevelopment())
}

func loggerMiddleware(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), consts.LoggerCtxName, l)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func main() {
	logger := zap.Must(zap.NewDevelopment())

	r := chi.NewRouter()
	r.Use(loggerMiddleware(logger))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		l := FromContext(r.Context())
		l.Info("A")
		
		delivery.RunCommand("ls -la", r.Context())
		w.Write([]byte("welcome"))
	})

	logger.Info(fmt.Sprintf("Serve port %d", consts.Port))
	http.ListenAndServe(fmt.Sprintf(":%d", consts.Port), r)
}