package dao

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/context/docker"
	"github.com/docker/cli/cli/context/store"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/jr-k/d4s/internal/dao/common"
	"github.com/jr-k/d4s/internal/dao/compose"
	"github.com/jr-k/d4s/internal/dao/docker/container"
	"github.com/jr-k/d4s/internal/dao/docker/image"
	"github.com/jr-k/d4s/internal/dao/docker/network"
	"github.com/jr-k/d4s/internal/dao/docker/volume"
	"github.com/jr-k/d4s/internal/dao/swarm/node"
	"github.com/jr-k/d4s/internal/dao/swarm/service"
)

// Re-export types for backward compatibility / convenience
type Resource = common.Resource
type HostStats = common.HostStats
type Container = container.Container
type Image = image.Image
type Volume = volume.Volume
type Network = network.Network
type Service = service.Service
type Node = node.Node
type ComposeProject = compose.ComposeProject

type DockerClient struct {
	Cli *client.Client
	Ctx context.Context
	
	// Managers
	Container *container.Manager
	Image     *image.Manager
	Volume    *volume.Manager
	Network   *network.Manager
	Service   *service.Manager
	Node      *node.Manager
	Compose   *compose.Manager
}

func NewDockerClient() (*DockerClient, error) {
	// Setup debug logging
	f, err := os.OpenFile("/tmp/d4s_debug_dao.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	var logger *log.Logger
	if err == nil {
		// Note: defer f.Close() here is slightly risky if function returns early and we want to keep writing?
		// Actually typical logger usage keeps file open. But NewDockerClient returns quickly.
		// So defer is fine.
		defer f.Close()
		logger = log.New(f, "d4s-dao: ", log.LstdFlags)
	} else {
		logger = log.New(io.Discard, "", 0)
	}

	opts := []client.Opt{
		client.WithAPIVersionNegotiation(),
	}

	// 1. DOCKER_HOST takes precedence (standard Docker behavior)
	if h := os.Getenv("DOCKER_HOST"); h != "" {
		logger.Printf("DOCKER_HOST set to %s, using FromEnv", h)
		opts = append(opts, client.FromEnv)
	} else {
		// 2. Identify Target Context
		targetCtx := "default"
		if envCtx := os.Getenv("DOCKER_CONTEXT"); envCtx != "" {
			targetCtx = envCtx
			logger.Printf("DOCKER_CONTEXT set to %s", targetCtx)
		} else {
			// Load config
			if cfg, err := config.Load(config.Dir()); err == nil {
				if cfg.CurrentContext != "" {
					targetCtx = cfg.CurrentContext
					logger.Printf("Loaded CurrentContext from config: %s", targetCtx)
				}
			} else {
				logger.Printf("Failed to load config: %v", err)
			}
		}

		if targetCtx == "default" {
			logger.Println("Context is default, using FromEnv")
			opts = append(opts, client.FromEnv)
		} else {
			// 3. Load Specific Context STRICTLY
			// We do NOT fallback to FromEnv if loading fails.
			logger.Printf("Loading context: %s", targetCtx)
			s := store.New(config.ContextStoreDir(), store.NewConfig(
				func() interface{} {
					return &map[string]interface{}{}
				},
				store.EndpointTypeGetter(docker.DockerEndpoint, func() interface{} { return &docker.EndpointMeta{} }),
			))
			meta, err := s.GetMetadata(targetCtx)
			if err != nil {
				logger.Printf("Error getting metadata for %s: %v", targetCtx, err)
				return nil, fmt.Errorf("failed to load docker context '%s': %v", targetCtx, err)
			}

			if epMeta, err := docker.EndpointFromContext(meta); err == nil {
				ep, err := docker.WithTLSData(s, targetCtx, epMeta)
				if err != nil {
					// Fallback to meta only if TLS loading fails
					logger.Printf("TLS data loading failed (non-critical): %v", err)
					ep = docker.Endpoint{EndpointMeta: epMeta}
				}

				logger.Printf("Using Host: %s", ep.Host)
				opts = append(opts, client.WithHost(ep.Host))
				if ep.TLSData != nil {
					tlsConfig := &tls.Config{
						InsecureSkipVerify: ep.SkipTLSVerify,
					}
					if ep.TLSData.CA != nil {
						certPool := x509.NewCertPool()
						certPool.AppendCertsFromPEM(ep.TLSData.CA)
						tlsConfig.RootCAs = certPool
					}
					if ep.TLSData.Cert != nil && ep.TLSData.Key != nil {
						cert, err := tls.X509KeyPair(ep.TLSData.Cert, ep.TLSData.Key)
						if err == nil {
							tlsConfig.Certificates = []tls.Certificate{cert}
						}
					}
					transport := &http.Transport{
						TLSClientConfig: tlsConfig,
					}
					opts = append(opts, client.WithHTTPClient(&http.Client{Transport: transport}))
				}
			} else {
				logger.Printf("EndpointFromContext failed for %s: %v", targetCtx, err)
				return nil, fmt.Errorf("failed to parse endpoint for context '%s': %v", targetCtx, err)
			}
		}
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	
	return &DockerClient{
		Cli:       cli,
		Ctx:       ctx,
		Container: container.NewManager(cli, ctx),
		Image:     image.NewManager(cli, ctx),
		Volume:    volume.NewManager(cli, ctx),
		Network:   network.NewManager(cli, ctx),
		Service:   service.NewManager(cli, ctx),
		Node:      node.NewManager(cli, ctx),
		Compose:   compose.NewManager(cli, ctx),
	}, nil
}

func (d *DockerClient) ListContainers() ([]common.Resource, error) {
	return d.Container.List()
}

func (d *DockerClient) ListImages() ([]common.Resource, error) {
	return d.Image.List()
}

func (d *DockerClient) ListVolumes() ([]common.Resource, error) {
	return d.Volume.List()
}

func (d *DockerClient) ListNetworks() ([]common.Resource, error) {
	return d.Network.List()
}

func (d *DockerClient) ListServices() ([]common.Resource, error) {
	return d.Service.List()
}

func (d *DockerClient) ListNodes() ([]common.Resource, error) {
	return d.Node.List()
}

func (d *DockerClient) ListCompose() ([]common.Resource, error) {
	return d.Compose.List()
}

// Actions wrappers
func (d *DockerClient) StopContainer(id string) error {
	return d.Container.Stop(id)
}

func (d *DockerClient) StartContainer(id string) error {
	return d.Container.Start(id)
}

func (d *DockerClient) RestartContainer(id string) error {
	return d.Container.Restart(id)
}

func (d *DockerClient) RemoveContainer(id string, force bool) error {
	return d.Container.Remove(id, force)
}

func (d *DockerClient) PruneContainers() error {
	return d.Container.Prune()
}

func (d *DockerClient) RemoveImage(id string, force bool) error {
	return d.Image.Remove(id, force)
}

func (d *DockerClient) PruneImages() error {
	return d.Image.Prune()
}

func (d *DockerClient) CreateVolume(name string) error {
	return d.Volume.Create(name)
}

func (d *DockerClient) RemoveVolume(id string, force bool) error {
	return d.Volume.Remove(id, force)
}

func (d *DockerClient) PruneVolumes() error {
	return d.Volume.Prune()
}

func (d *DockerClient) CreateNetwork(name string) error {
	return d.Network.Create(name)
}

func (d *DockerClient) RemoveNetwork(id string) error {
	return d.Network.Remove(id)
}

func (d *DockerClient) PruneNetworks() error {
	return d.Network.Prune()
}

func (d *DockerClient) ScaleService(id string, replicas uint64) error {
	return d.Service.Scale(id, replicas)
}

func (d *DockerClient) RemoveService(id string) error {
	return d.Service.Remove(id)
}

func (d *DockerClient) RemoveNode(id string, force bool) error {
	return d.Node.Remove(id, force)
}

func (d *DockerClient) StopComposeProject(projectName string) error {
	return d.Compose.Stop(projectName)
}

func (d *DockerClient) RestartComposeProject(projectName string) error {
	return d.Compose.Restart(projectName)
}

func (d *DockerClient) GetComposeConfig(projectName string) (string, error) {
	return d.Compose.GetConfig(projectName)
}

// Common/Stats wrappers
func (d *DockerClient) GetHostStats() (common.HostStats, error) {
	return common.GetHostStats(d.Cli, d.Ctx)
}

func (d *DockerClient) GetHostStatsWithUsage() (common.HostStats, error) {
	return common.GetHostStatsWithUsage(d.Cli, d.Ctx)
}

func (d *DockerClient) Inspect(resourceType, id string) (string, error) {
	return common.Inspect(d.Cli, d.Ctx, resourceType, id)
}

func (d *DockerClient) GetContainerStats(id string) (string, error) {
	return common.GetContainerStats(d.Cli, d.Ctx, id)
}

func (d *DockerClient) GetContainerEnv(id string) ([]string, error) {
	return d.Container.GetEnv(id)
}

func (d *DockerClient) HasTTY(id string) (bool, error) {
	return common.HasTTY(d.Cli, d.Ctx, id)
}

func (d *DockerClient) GetContainerLogs(id string, since string, tail string, timestamps bool) (io.ReadCloser, error) {
	return d.Container.Logs(id, since, tail, timestamps)
}

func (d *DockerClient) GetServiceLogs(id string, since string, tail string, timestamps bool) (io.ReadCloser, error) {
	return d.Service.Logs(id, since, tail, timestamps)
}

func (d *DockerClient) ListTasksForNode(nodeID string) ([]swarm.Task, error) {
	return d.Node.ListTasks(nodeID)
}

func (d *DockerClient) ListVolumesForContainer(id string) ([]common.Resource, error) {
	// Inspect container to get mounts
	json, err := d.Cli.ContainerInspect(d.Ctx, id)
	if err != nil {
		return nil, err
	}
	
	names := make(map[string]bool)
	for _, m := range json.Mounts {
		if m.Type == "volume" {
			names[m.Name] = true
		}
	}
	
	// List all volumes and filter
	all, err := d.Volume.List()
	if err != nil {
		return nil, err
	}
	
	var filtered []common.Resource
	for _, r := range all {
		if v, ok := r.(volume.Volume); ok {
			if names[v.Name] {
				filtered = append(filtered, r)
			}
		}
	}
	
	return filtered, nil
}

func (d *DockerClient) ListNetworksForContainer(id string) ([]common.Resource, error) {
	// Inspect container to get networks
	json, err := d.Cli.ContainerInspect(d.Ctx, id)
	if err != nil {
		return nil, err
	}
	
	ids := make(map[string]bool)
	for _, n := range json.NetworkSettings.Networks {
		ids[n.NetworkID] = true
	}
	
	// List all networks and filter
	all, err := d.Network.List()
	if err != nil {
		return nil, err
	}
	
	var filtered []common.Resource
	for _, r := range all {
		if n, ok := r.(network.Network); ok {
			if ids[n.ID] {
				filtered = append(filtered, r)
			}
		}
	}
	
	return filtered, nil
}
