package domain

import "context"

// Runner for docker, terminal etc
// Can save to local dir and run file with needed language
type Runner interface {
	GetLanguage() string
	// Get stdout and stderror
	RunCommand([]string) (string, string, error)
	SaveFile(path string, data string) error
	CloseRunner()
}

// System that gives containers to run at
type DockerSystem interface {
	GetContainer(language string, ctx context.Context) (Runner, error)
	SetMaxContainers(maxCont int, ctx context.Context) (error)
	SetMinContainers(minCont int, ctx context.Context) (error)
}

// Key is language name and value is image in docker container
type LanguageRec map[string]string