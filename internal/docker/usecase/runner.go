package usecase

import (
	"archive/tar"
	"bytes"
	"check_system/internal/domain"
	"context"
	"fmt"
	"io"
	"path/filepath"
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

func (r *DockerRunner) SaveFile(path string, data string) error {
	// Create a tar stream from the text content
	_, tarWriter := io.Pipe()
	go func() {
		defer tarWriter.Close()
		tw := tar.NewWriter(tarWriter)
		defer tw.Close()
		header := &tar.Header{
			Name: filepath.Base(path),
			Mode: 0600,
			Size: int64(len(data)),
		}
		if err := tw.WriteHeader(header); err != nil {
			// TODO make errors here
			return 
		}
		if _, err := io.Copy(tw, bytes.NewReader([]byte(data))); err != nil {
			// TODO make errors here
			return
		}
	}()
	// Create the exec configuration
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"sh", "-c", fmt.Sprintf("cat > %s", path)},
	}

	// Create the exec instance
	resp, err := r.cli.ContainerExecCreate(context.Background(), r.idContainer, execConfig)
	if err != nil {
		return err
	}

	// Start the exec instance with the tar stream as input
	err = r.cli.ContainerExecStart(context.Background(), resp.ID, types.ExecStartCheck{})
	if err != nil {
		return err
	}

	// Attach to the exec instance for handling input/output
	hijackedResp, err := r.cli.ContainerExecAttach(context.Background(), resp.ID, types.ExecStartCheck{})
	if err != nil {
		return err
	}
	defer hijackedResp.Close()

	// Wait for the exec instance to finish
	_, err = r.cli.ContainerExecInspect(context.Background(), resp.ID)
	if err != nil {
		return err
	}

	return nil
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

func (r *DockerRunner) getContainerId(cli *client.Client) (string, error) {
	containerInfo, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return "", err
	}

	id := containerInfo[0].ID
	return id, nil
} 
