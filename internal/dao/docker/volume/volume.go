package volume

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	volTypes "github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/gdamore/tcell/v2"
	"github.com/jr-k/d4s/internal/dao/common"
	"github.com/jr-k/d4s/internal/ui/styles"
)

type Manager struct {
	cli *client.Client
	ctx context.Context
}

func NewManager(cli *client.Client, ctx context.Context) *Manager {
	return &Manager{cli: cli, ctx: ctx}
}

// Volume Model
type Volume struct {
	Name    string
	Driver  string
	Mount   string
	Created string
	Size    string
	Scope   string
}

func (v Volume) GetID() string { return v.Name }
func (v Volume) GetCells() []string {
	return []string{v.Name, v.Driver, v.Scope, v.Mount, v.Created, v.Size}
}

func (v Volume) GetStatusColor() (tcell.Color, tcell.Color) {
	return styles.ColorIdle, tcell.ColorBlack
}

func (m *Manager) List() ([]common.Resource, error) {
	// 1. Get List of all volumes (fast & reliable)
	list, err := m.cli.VolumeList(m.ctx, volTypes.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 2. Try to get Usage Data (optional / might fail or be partial)
	sizes := make(map[string]string)

	// Use a timeout context for DiskUsage as it can be very slow
	ctx, cancel := context.WithTimeout(m.ctx, 2*time.Second)
	defer cancel()

	du, err := m.cli.DiskUsage(ctx, types.DiskUsageOptions{})
	if err == nil {
		for _, v := range du.Volumes {
			if v.UsageData != nil {
				sizes[v.Name] = common.FormatBytes(v.UsageData.Size)
			}
		}
	}

	var res []common.Resource
	for _, v := range list.Volumes {
		created := "-"
		if v.CreatedAt != "" {
			t, err := time.Parse(time.RFC3339, v.CreatedAt)
			if err == nil {
				created = common.FormatTime(t.Unix())
			}
		}

		size := "-"
		if s, ok := sizes[v.Name]; ok {
			size = s
		}

		res = append(res, Volume{
			Name:    v.Name,
			Driver:  v.Driver,
			Mount:   common.ShortenPath(v.Mountpoint),
			Created: created,
			Size:    size,
			Scope:   v.Scope,
		})
	}
	return res, nil
}

func (m *Manager) Create(name string) error {
	_, err := m.cli.VolumeCreate(m.ctx, volTypes.CreateOptions{
		Name: name,
	})
	return err
}

func (m *Manager) Remove(id string, force bool) error {
	return m.cli.VolumeRemove(m.ctx, id, force)
}

func (m *Manager) Prune() error {
	_, err := m.cli.VolumesPrune(m.ctx, filters.NewArgs())
	return err
}
