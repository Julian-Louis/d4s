package container

import (
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/jessym/d4s/internal/dao/common"
	"golang.org/x/net/context"
)

type Manager struct {
	cli *client.Client
	ctx context.Context
}

func NewManager(cli *client.Client, ctx context.Context) *Manager {
	return &Manager{cli: cli, ctx: ctx}
}

// Container Model
type Container struct {
	ID          string
	Names       string
	Image       string
	Status      string
	State       string
	Age         string
	Ports       string
	Created     string
	Compose     string
	ProjectName string
	CPU         string
	Mem         string
}

func (c Container) GetID() string { return c.ID }
func (c Container) GetCells() []string {
	id := c.ID
	if len(id) > 12 {
		id = id[:12]
	}
	return []string{id, c.Names, c.Image, c.Status, c.Age, c.Ports, c.CPU, c.Mem, c.Compose, c.Created}
}

func (m *Manager) List() ([]common.Resource, error) {
	list, err := m.cli.ContainerList(m.ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	var res []common.Resource
	for _, c := range list {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		
		ports := ""
		if len(c.Ports) > 0 {
			ports = fmt.Sprintf("%d->%d", c.Ports[0].PublicPort, c.Ports[0].PrivatePort)
		}

		compose := ""
		if cf, ok := c.Labels["com.docker.compose.project.config_files"]; ok {
			compose = "📄 " + common.ShortenPath(cf)
		} else if proj, ok := c.Labels["com.docker.compose.project"]; ok {
			compose = "📦 " + proj
		}
		
		projectName := c.Labels["com.docker.compose.project"]

		status, age := common.ParseStatus(c.Status)

		res = append(res, Container{
			ID:          c.ID,
			Names:       name,
			Image:       c.Image,
			Status:      status,
			Age:         age,
			State:       c.State,
			Ports:       ports,
			Created:     common.FormatTime(c.Created),
			Compose:     compose,
			ProjectName: projectName,
			CPU:         "0%", // Mock until async stats implemented
			Mem:         "0% ([#6272a4]0 B[-])", // Mock
		})
	}
	return res, nil
}

func (m *Manager) Stop(id string) error {
	timeout := 10 // seconds
	return m.cli.ContainerStop(m.ctx, id, container.StopOptions{Timeout: &timeout})
}

func (m *Manager) Start(id string) error {
	return m.cli.ContainerStart(m.ctx, id, container.StartOptions{})
}

func (m *Manager) Restart(id string) error {
	timeout := 10 // seconds
	return m.cli.ContainerRestart(m.ctx, id, container.StopOptions{Timeout: &timeout})
}

func (m *Manager) Remove(id string, force bool) error {
	return m.cli.ContainerRemove(m.ctx, id, container.RemoveOptions{Force: force})
}

func (m *Manager) Logs(id string, timestamps bool) (io.ReadCloser, error) {
	opts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "200", // Enough to fill screen, optimized start
		Timestamps: timestamps,
	}
	return m.cli.ContainerLogs(m.ctx, id, opts)
}

func (m *Manager) GetEnv(id string) ([]string, error) {
	c, err := m.cli.ContainerInspect(m.ctx, id)
	if err != nil {
		return nil, err
	}
	return c.Config.Env, nil
}
