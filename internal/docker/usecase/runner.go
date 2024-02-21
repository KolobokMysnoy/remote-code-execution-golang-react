package usecase

import (
	"archive/tar"
	"bytes"
	"check_system/internal/domain"
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type DockerRunner struct {
	language string
	idContainer string
	cli *client.Client
	releaseFunc func()
}

// Create new docker runner
func NewDockerRunner(language string, cli *client.Client, releaseFunc func()) (domain.Runner, error) {
	runner := &DockerRunner{
		language: language,
		cli: cli,
		releaseFunc: releaseFunc,
	}

	id, err := runner.getContainerId(cli)
	if err != nil {
		return nil, err
	}
	runner.idContainer = id
	
	return runner, nil
}

// Return language of runner
func (r *DockerRunner) GetLanguage() string {
	return r.language
}

func (r *DockerRunner) getContainerId(cli *client.Client) (string, error) {
	containerInfo, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return "", err
	}

	id := containerInfo[0].ID
	return id, nil
}

func (r *DockerRunner) RunCommand(command []string) (string, string, error) {
	// https://stackoverflow.com/questions/52774830/docker-exec-command-from-golang-api
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          command,
	}

	ctx := context.Background()

	resp, err := r.cli.ContainerExecCreate(ctx, r.idContainer, execConfig)
	if err != nil {
		return "", "", err
	}

	hijackedResp, err := r.cli.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{})
	if err != nil {
		return "", "", err
	}
	defer hijackedResp.Close()

	var stdout, stderr strings.Builder
	_, err = stdcopy.StdCopy(&stdout, &stderr, hijackedResp.Reader)
	stdoutStr := stdout.String()
	stderrStr := stderr.String()
	
	return stdoutStr, stderrStr, nil
}

func (r *DockerRunner) CloseRunner() {
	r.releaseFunc()
}

func (r *DockerRunner) closeApp(cli client.Client, containerID, processName string) error {
	execCreateResp, err := cli.ContainerExecCreate(context.Background(), containerID, types.ExecConfig{
		Cmd: []string{"pkill", "-SIGTERM", processName},
	})
	if err != nil {
		return fmt.Errorf("Failed to create exec in container: %v", err)
	}

	err = cli.ContainerExecStart(context.Background(), execCreateResp.ID, types.ExecStartCheck{})
	if err != nil {
		return fmt.Errorf("Failed to start exec in container: %v", err)
	}

	return nil
}

func (r *DockerRunner) SaveFile(path, nameOfFile, data string) error {
	// Create a tar stream from the text content
	tarR, err := r.getTarStream(nameOfFile, data)
	if err != nil {
		log.Println("Tar error:", err)
		return err
	}

	// Use the tar stream directly as an io.Reader
	err = r.cli.CopyToContainer(context.Background(), r.idContainer, path, tarR, types.CopyToContainerOptions{})
	if err != nil {
		log.Println("Copy error:", err)
		return err
	}

	return nil
}

func (r *DockerRunner) getTarStream(path, data string) (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	hdr := &tar.Header{
		Name: path,
		Mode: 0644,
		Size: int64(len(data)),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}

	if _, err := tw.Write([]byte(data)); err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	// Use the buffer as an io.Reader
	return &buf, nil
}