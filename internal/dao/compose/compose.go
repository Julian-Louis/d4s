package compose

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
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

// ComposeProject Model
type ComposeProject struct {
	Name        string
	Status      string
	ConfigFiles string
}

func (cp ComposeProject) GetID() string { return cp.Name }
func (cp ComposeProject) GetCells() []string {
	return []string{cp.Name, cp.Status, cp.ConfigFiles}
}

func (m *Manager) List() ([]common.Resource, error) {
	list, err := m.cli.ContainerList(m.ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	type projData struct {
		total   int
		running int
		config  string
	}
	projects := make(map[string]*projData)

	for _, c := range list {
		proj := c.Labels["com.docker.compose.project"]
		if proj == "" {
			continue
		}

		if _, ok := projects[proj]; !ok {
			config := ""
			if cf, ok := c.Labels["com.docker.compose.project.config_files"]; ok {
				config = common.ShortenPath(cf)
			}
			projects[proj] = &projData{
				config: config,
			}
		}

		projects[proj].total++
		if c.State == "running" {
			projects[proj].running++
		}
	}

	var res []common.Resource
	for name, data := range projects {
		res = append(res, ComposeProject{
			Name:        name,
			Status:      fmt.Sprintf("Running (%d/%d)", data.running, data.total),
			ConfigFiles: data.config,
		})
	}
	return res, nil
}

func (m *Manager) Stop(projectName string) error {
	// Find all containers with this project name
	args := filters.NewArgs()
	args.Add("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))
	
	containers, err := m.cli.ContainerList(m.ctx, container.ListOptions{Filters: args, All: true})
	if err != nil {
		return err
	}
	
	if len(containers) == 0 {
		return fmt.Errorf("no containers found for project %s", projectName)
	}

	// Stop them all (sequentially for now, or parallel if needed)
	timeout := 10
	var errs []string
	for _, c := range containers {
		if err := m.cli.ContainerStop(m.ctx, c.ID, container.StopOptions{Timeout: &timeout}); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", c.Names[0], err))
		}
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("errors stopping containers: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (m *Manager) Restart(projectName string) error {
	// Find all containers with this project name
	args := filters.NewArgs()
	args.Add("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))
	
	containers, err := m.cli.ContainerList(m.ctx, container.ListOptions{Filters: args, All: true})
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		return fmt.Errorf("no containers found for project %s", projectName)
	}

	timeout := 10
	var errs []string
	for _, c := range containers {
		if err := m.cli.ContainerRestart(m.ctx, c.ID, container.StopOptions{Timeout: &timeout}); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", c.Names[0], err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors restarting containers: %s", strings.Join(errs, "; "))
	}
	return nil
}
