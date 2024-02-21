package domain

import (
	"context"

	"github.com/docker/docker/client"
)

// Runner for docker, terminal etc
// Can save to local dir and run file with needed language
type Runner interface {
	GetLanguage() string
	// Get stdout and stderror
	RunCommand([]string) (string, string, error)
	SaveFile(path, nameOfFile, data string) error
	CloseRunner()
}

// System that gives containers to run at
type DockerSystem interface {
	GetRunner(language string, ctx context.Context) (Runner, error)
	ReturnRunner(runner Runner)
	SetMax(int)
	SetMin(int)
}

type DockerPool interface {
	AddContainer() (error) 
	ReleaseExtraContainers(desiredSizeOfPool int) (error) 
	ReturnContainer(*client.Client) 
	// Timeout to wait untill raise error
	GetContainer(timeout int) (*client.Client, error) 
	SetImage(string) 
	Size() int
	Active() int
}

type ProgramRunner interface {
	RunProgram(nameOfFile, program string) (outStr, errOut string, err error)
}

// Key is language name and value is image in docker container
type LanguageRec map[string]string