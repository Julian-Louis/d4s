package stack

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/gdamore/tcell/v2"
	"github.com/jr-k/d4s/internal/dao/common"
	"github.com/jr-k/d4s/internal/ui/styles"
)

type Manager struct {
	cli         *client.Client
	ctx         context.Context
	contextName string
}

func NewManager(cli *client.Client, ctx context.Context, contextName string) *Manager {
	return &Manager{cli: cli, ctx: ctx, contextName: contextName}
}

// dockerCmd builds a docker CLI invocation pinned to the manager's
// docker context, so commands hit the right daemon (e.g. over SSH).
func (m *Manager) dockerCmd(args ...string) *exec.Cmd {
	if m.contextName != "" && m.contextName != "default" {
		args = append([]string{"--context", m.contextName}, args...)
	}
	return exec.Command("docker", args...)
}

type Stack struct {
	Name     string
	Services int
	Running  int
	Status   string
}

func (s Stack) GetID() string { return s.Name }
func (s Stack) GetCells() []string {
	ready := fmt.Sprintf("%d/%d", s.Running, s.Services)
	return []string{s.Name, ready, s.Status}
}

func (s Stack) GetStatusColor() (tcell.Color, tcell.Color) {
	if s.Services == 0 {
		return styles.ColorStatusGray, styles.ColorBlack
	}
	if s.Running < s.Services {
		return styles.ColorStatusRed, styles.ColorBlack
	}
	return styles.ColorIdle, styles.ColorBlack
}

func (s Stack) GetColumnValue(column string) string {
	switch strings.ToLower(column) {
	case "name":
		return s.Name
	case "ready":
		return fmt.Sprintf("%d/%d", s.Running, s.Services)
	case "status":
		return s.Status
	}
	return ""
}

func (s Stack) GetDefaultColumn() string {
	return "Name"
}

func (s Stack) GetDefaultSortColumn() string {
	return "Name"
}

func (m *Manager) List() ([]common.Resource, error) {
	services, err := m.cli.ServiceList(m.ctx, types.ServiceListOptions{})
	if err != nil {
		return nil, err
	}

	type stackData struct {
		services int
		running  int
	}
	stacks := make(map[string]*stackData)

	for _, svc := range services {
		ns := svc.Spec.Labels["com.docker.stack.namespace"]
		if ns == "" {
			continue
		}
		if _, ok := stacks[ns]; !ok {
			stacks[ns] = &stackData{}
		}
		stacks[ns].services++

		// Count running tasks for this service
		taskFilter := filters.NewArgs()
		taskFilter.Add("service", svc.ID)
		taskFilter.Add("desired-state", "running")
		tasks, err := m.cli.TaskList(m.ctx, types.TaskListOptions{Filters: taskFilter})
		if err == nil {
			for _, t := range tasks {
				if t.Status.State == "running" {
					stacks[ns].running++
				}
			}
		}
	}

	var res []common.Resource
	for name, data := range stacks {
		status := "Ready"
		if data.running < data.services {
			status = "Degraded"
		}
		if data.services == 0 {
			status = "Empty"
		}

		res = append(res, Stack{
			Name:     name,
			Services: data.services,
			Running:  data.running,
			Status:   status,
		})
	}
	return res, nil
}

func (m *Manager) Remove(name string) error {
	cmd := m.dockerCmd("stack", "rm", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error removing stack: %v, output: %s", err, string(output))
	}
	return nil
}

func (m *Manager) Deploy(name string, composeFile string) error {
	cmd := m.dockerCmd("stack", "deploy", "-c", composeFile, name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error deploying stack: %v, output: %s", err, string(output))
	}
	return nil
}

func (m *Manager) PS(name string) (string, error) {
	cmd := m.dockerCmd("stack", "ps", name, "--no-trunc")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error listing stack tasks: %v, output: %s", err, string(output))
	}
	return string(output), nil
}
