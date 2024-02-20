package main

import (
	consts "check_system/config"
	runners_delivery "check_system/internal/code_runner/delivery"
	"check_system/internal/docker/usecase"
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
	ctx := context.WithValue(context.Background(), consts.LoggerCtxName, logger)

	system, err := usecase.NewDockerSystem(consts.Languages, ctx) 
	if err != nil {
		logger.Error("can't init system", zap.Error(err))
		return
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		l := FromContext(r.Context())
		l.Info("A")
		
		data := `package main

import "fmt"

func main() {
	fmt.Print("Hello world!")
}`

		out, errs, err := runners_delivery.RunGo(data, r.Context(), system)
		if err != nil {
			logger.Error("err in rungo function", zap.Error(err))
		} else {
			logger.Info("output: ", zap.String("output", out), zap.String("errs", errs))
		}
		system.SetMin(0)
		system.SetMax(1)
		w.Write([]byte("welcome"))
	})

	logger.Info(fmt.Sprintf("Serve port %d", consts.Port))
	http.ListenAndServe(fmt.Sprintf(":%d", consts.Port), r)
}