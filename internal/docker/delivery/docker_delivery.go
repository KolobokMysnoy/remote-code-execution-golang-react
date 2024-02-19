package delivery

import (
	consts "check_system/config"
	"check_system/internal/docker/usecase"
	"context"
	"log"
	"strings"

	"go.uber.org/zap"
)

var loggg, err = zap.NewDevelopment() 
var system, _ = usecase.NewDockerSystem(consts.Languages, context.WithValue(context.Background(), consts.LoggerCtxName, loggg))

func RunCommand(command string, ctx context.Context) {
	runner, err := system.GetRunner("golang", ctx)
	if err != nil {
		return
	}
	defer runner.CloseRunner()

	// output, errors, err := runner.RunCommand(strings.Split(command, " "))
	// if err != nil {
	// 	return
	// }

	data := `package main

import "fmt"

func main() {
	fmt.Print("Hello world!")
}`

	
	if err := runner.SaveFile("main.go", data); err != nil {
		log.Default().Print("Troubles save", err)
		return 
	}

	cmd := "ls -la"
	output, errors, err := runner.RunCommand(strings.Split(cmd, " "))
	if err != nil {
		return
	}

	log.Default().Print("Output ", output)
	log.Default().Print("Errors ", errors)

	
	cmd = "cat main.go"
	output, errors, err = runner.RunCommand(strings.Split(cmd, " "))
	if err != nil {
		return
	}
	log.Default().Print("Output ", output)
	log.Default().Print("Errors ", errors)

	cmd = "od -c main.go"
	output, errors, err = runner.RunCommand(strings.Split(cmd, " "))
	if err != nil {
		return
	}
	log.Default().Print("Output ", output)
	log.Default().Print("Errors ", errors)

	cmd = "go run main.go"
	output, errors, err = runner.RunCommand(strings.Split(cmd, " "))
	if err != nil {
		return
	}
	log.Default().Print("Output ", output)
	log.Default().Print("Errors ", errors)
	
}