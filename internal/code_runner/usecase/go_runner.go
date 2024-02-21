package usecase

import (
	consts "check_system/config"
	"check_system/internal/domain"
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

type GoRunner struct {
	language string
	runner domain.Runner
	path string
	log *zap.Logger
}

func FromContext(ctx context.Context) *zap.Logger {
    logger := ctx.Value(consts.LoggerCtxName)
    if l, ok := logger.(*zap.Logger); ok {
        return l
    }
    return zap.Must(zap.NewDevelopment())
}

func NewGoRunner(language string, runner domain.Runner, ctx context.Context) (domain.ProgramRunner) {
	log := FromContext(ctx)
	
	goRunner := &GoRunner{
		language: language,
		runner: runner,
		path: "/go",
		log: log,
	}

	return goRunner
}

func (g *GoRunner) RunProgram(nameOfFile, program string) (outStr, errOut string, err error) {
	defer func() {
		g.log.Info("close runner in GoRun")
		g.runner.CloseRunner()
	}()

	err = g.beforeRun(g.path, nameOfFile, program)
	if err != nil {
		g.log.Error("before run", zap.Error(err))
		return "", "", err
	}

	com := strings.Split(fmt.Sprintf("go run %s//%s", g.path, nameOfFile), " ")
	outStr, errOut, err = g.runner.RunCommand(com)
	if err != nil {
		g.log.Error("run command", zap.Error(err))
		return "", "", err
	}

	err = g.afterRun(g.path, nameOfFile)
	if err != nil {
		g.log.Error("after run", zap.Error(err))
		return "", "", err
	}

	return 
}

func (g *GoRunner) beforeRun(path, nameOfFile, program string) (error) {
	return g.runner.SaveFile(path, nameOfFile, program)
}

func (g *GoRunner) afterRun(path, fileName string) (error) {
	com := strings.Split(fmt.Sprintf("rm %s//%s", path, fileName), " ")
	_, _, err := g.runner.RunCommand(com)
	return err
}