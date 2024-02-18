package usecase

import (
	"archive/tar"
	"bytes"
	"check_system/internal/domain"
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)


type DockerRunner struct {
	language string
	cli *client.Client
	releaseFunc func()
	id string
}

func TransformConnToRunner(conn *client.Client, language string, releaseFunc func(), ) (domain.Runner, error) {
	runner := &DockerRunner{
		language: language,
		releaseFunc: releaseFunc,
		cli: conn,
	}

	// Get container id
	containerInfo, err := conn.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return nil, err
	}

	id := containerInfo[len(containerInfo)-1].ID
	runner.id = id

	return runner, nil
}

func (r *DockerRunner) GetLanguage() string {
	return r.language
}

func (r *DockerRunner) RunCommand(command []string) (string, error) {
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          command,
	}
	

	resp, err := r.cli.ContainerExecCreate(context.Background(), r.id, execConfig)
	if err != nil {
		return "", err
	}

	hijackedResp, err := r.cli.ContainerExecAttach(context.Background(), resp.ID, types.ExecStartCheck{})
	if err != nil {
		return "", err
	}
	defer hijackedResp.Close()

	output, err := io.ReadAll(hijackedResp.Reader)
	if err != nil {
		return "", err
	}

	return string(output), nil
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
	resp, err := r.cli.ContainerExecCreate(context.Background(), r.id, execConfig)
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
