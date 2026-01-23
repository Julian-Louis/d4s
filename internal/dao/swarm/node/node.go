package node

import (
	"strings"

	dt "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/gdamore/tcell/v2"
	"github.com/jr-k/d4s/internal/dao/common"
	"github.com/jr-k/d4s/internal/ui/styles"
	"golang.org/x/net/context"
)

type Manager struct {
	cli *client.Client
	ctx context.Context
}

func NewManager(cli *client.Client, ctx context.Context) *Manager {
	return &Manager{cli: cli, ctx: ctx}
}

// Node Model
type Node struct {
	ID       string
	Hostname string
	Status   string
	Avail    string
	Role     string
	Version  string
	Created  string
}

func (n Node) GetID() string { return n.ID }
func (n Node) GetCells() []string {
	id := n.ID
	if len(id) > 12 {
		id = id[:12]
	}
	return []string{id, n.Hostname, n.Status, n.Avail, n.Role, n.Version, n.Created}
}

func (n Node) GetStatusColor() (tcell.Color, tcell.Color) {
	status := strings.ToLower(n.Status)
	// avail := strings.ToLower(n.Avail)

	if strings.Contains(status, "ready") || strings.Contains(status, "active") {
		return styles.ColorStatusGreen, styles.ColorBlack
	}
	if strings.Contains(status, "down") || strings.Contains(status, "drain") || strings.Contains(status, "disconnected") {
		return styles.ColorStatusRed, styles.ColorBlack
	}
	if strings.Contains(status, "unknown") {
		return styles.ColorStatusOrange, styles.ColorBlack
	}
	
	return styles.ColorIdle, styles.ColorBlack
}

func (n Node) GetColumnValue(column string) string {
	switch strings.ToLower(column) {
	case "id":
		return n.ID
	case "hostname":
		return n.Hostname
	case "status":
		return n.Status
	case "avail":
		return n.Avail
	case "role":
		return n.Role
	case "version":
		return n.Version
	case "created":
		return n.Created
	}
	return ""
}

func (n Node) GetDefaultColumn() string {
	return "Hostname"
}

func (n Node) GetDefaultSortColumn() string {
	return "Hostname"
}

func (m *Manager) List() ([]common.Resource, error) {
	list, err := m.cli.NodeList(m.ctx, dt.NodeListOptions{})
	if err != nil {
		return nil, err
	}

	var res []common.Resource
	for _, n := range list {
		res = append(res, Node{
			ID:       n.ID,
			Hostname: n.Description.Hostname,
			Status:   string(n.Status.State),
			Avail:    string(n.Spec.Availability),
			Role:     string(n.Spec.Role),
			Version:  n.Description.Engine.EngineVersion,
			Created:  common.FormatTime(n.CreatedAt.Unix()),
		})
	}
	return res, nil
}

func (m *Manager) Remove(id string, force bool) error {
	// Force remove
	return m.cli.NodeRemove(m.ctx, id, swarm.NodeRemoveOptions{Force: force})
}

func (m *Manager) ListTasks(nodeID string) ([]swarm.Task, error) {
	filter := filters.NewArgs()
	filter.Add("node", nodeID)
	filter.Add("desired-state", "running")
	
	return m.cli.TaskList(m.ctx, dt.TaskListOptions{Filters: filter})
}
