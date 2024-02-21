package delivery

import (
	consts "check_system/config"
	"check_system/internal/code_runner/usecase"
	"check_system/internal/domain"
	"context"

	"go.uber.org/zap"
)

// XXX
func FromContext(ctx context.Context) *zap.Logger {
    logger := ctx.Value(consts.LoggerCtxName)
    if l, ok := logger.(*zap.Logger); ok {
        return l
    }
    return zap.Must(zap.NewDevelopment())
}

func RunGo(program string, ctx context.Context, system domain.DockerSystem) (out, errOut string, err error) {
	lg := FromContext(ctx)

	lg.Info("Run go function")

	lang := "golang"
	runner, err := system.GetRunner(lang, ctx)
	if err != nil {
		lg.Error("error can't get runner", zap.Error(err))
		return "", "", err
	}
	
	goRunner := usecase.NewGoRunner(lang, runner, ctx)
	out, errOut, err = goRunner.RunProgram("main.go", program)
	if err != nil {
		lg.Error("error can't get go runner", zap.Error(err))
		return "", "", err
	}

	return 
}