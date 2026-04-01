package adapters

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerBaseAdapter struct {
	cli *client.Client
}

func NewDockerBaseAdapter() (*DockerBaseAdapter, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerBaseAdapter{cli: cli}, nil
}

func (a *DockerBaseAdapter) IsContainerRunning(ctx context.Context, name string) (bool, error) {
	inspect, err := a.cli.ContainerInspect(ctx, name)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return inspect.State.Running, nil
}

func (a *DockerBaseAdapter) StopAndRemove(ctx context.Context, name string) error {
	_ = a.cli.ContainerStop(ctx, name, container.StopOptions{})
	return a.cli.ContainerRemove(ctx, name, container.RemoveOptions{Force: true})
}

func (a *DockerBaseAdapter) ExecuteInContainer(ctx context.Context, name string, cmd []string) (string, error) {
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	}
	resp, err := a.cli.ContainerExecCreate(ctx, name, execConfig)
	if err != nil {
		return "", err
	}

	attach, err := a.cli.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{})
	if err != nil {
		return "", err
	}
	defer attach.Close()

	data, err := io.ReadAll(attach.Reader)
	return string(data), err
}
