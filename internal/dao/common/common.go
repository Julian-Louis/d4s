package common

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

// Re-export common types
type Resource interface {
	GetID() string
	GetCells() []string
}

// HostStats represents basic host metrics
type HostStats struct {
	CPU        string
	CPUPercent string
	Mem        string
	MemPercent string
	Name       string
	Version    string
	Context    string
	User       string
	Hostname   string
	D4SVersion string
}

func GetHostStats(cli *client.Client, ctx context.Context) (HostStats, error) {
	info, err := cli.Info(ctx)
	if err != nil {
		return HostStats{}, err
	}
	
	memTotal := FormatBytes(info.MemTotal)
	
	// Get current user
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME") // Windows fallback
	}
	if user == "" {
		user = "unknown"
	}
	
	// Get hostname
	hostname, _ := os.Hostname()

	return HostStats{
		CPU:        fmt.Sprintf("%d", info.NCPU),
		CPUPercent: "...", // Placeholder
		Mem:        memTotal,
		MemPercent: "...", // Placeholder
		Name:       info.Name,
		Version:    info.ServerVersion,
		Context:    "default",
		User:       user,
		Hostname:   hostname,
		D4SVersion: "1.0.0",
	}, nil
}

func GetHostStatsWithUsage(cli *client.Client, ctx context.Context) (HostStats, error) {
	// First get basic stats
	stats, err := GetHostStats(cli, ctx)
	if err != nil {
		return stats, err
	}
	
	// Then calculate usage stats asynchronously
	info, _ := cli.Info(ctx)
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: false})
	if err != nil || len(containers) == 0 {
		return stats, nil
	}
	
	var totalCPU float64
	var totalMem uint64
	validStats := 0
	
	// Collect stats from first few containers (to avoid blocking too long)
	maxContainers := len(containers)
	if maxContainers > 10 {
		maxContainers = 10 // Limit to 10 containers for performance
	}
	
	for i := 0; i < maxContainers; i++ {
		c := containers[i]
		statsResp, err := cli.ContainerStats(ctx, c.ID, false)
		if err != nil {
			continue
		}
		
		var v map[string]interface{}
		if err := json.NewDecoder(statsResp.Body).Decode(&v); err != nil {
			statsResp.Body.Close()
			continue
		}
		statsResp.Body.Close()
		
		// Extract CPU stats
		if cpuStats, ok := v["cpu_stats"].(map[string]interface{}); ok {
			if preCPUStats, ok := v["precpu_stats"].(map[string]interface{}); ok {
				if cpuUsage, ok := cpuStats["cpu_usage"].(map[string]interface{}); ok {
					if preCPUUsage, ok := preCPUStats["cpu_usage"].(map[string]interface{}); ok {
						if totalUsage, ok := cpuUsage["total_usage"].(float64); ok {
							if preTotalUsage, ok := preCPUUsage["total_usage"].(float64); ok {
								if systemUsage, ok := cpuStats["system_cpu_usage"].(float64); ok {
									if preSystemUsage, ok := preCPUStats["system_cpu_usage"].(float64); ok {
										cpuDelta := totalUsage - preTotalUsage
										systemDelta := systemUsage - preSystemUsage
										if systemDelta > 0 && cpuDelta > 0 {
											if percpu, ok := cpuUsage["percpu_usage"].([]interface{}); ok {
												cpuPct := (cpuDelta / systemDelta) * float64(len(percpu)) * 100.0
												totalCPU += cpuPct
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		
		// Extract memory stats
		if memStats, ok := v["memory_stats"].(map[string]interface{}); ok {
			if usage, ok := memStats["usage"].(float64); ok {
				totalMem += uint64(usage)
				validStats++
			}
		}
	}
	
	// Format CPU percentage
	if validStats > 0 && totalCPU > 0 {
		stats.CPUPercent = fmt.Sprintf("%.1f%%", totalCPU)
	} else {
		stats.CPUPercent = "0%"
	}
	
	// Format Memory percentage
	if info.MemTotal > 0 && totalMem > 0 {
		memPct := float64(totalMem) / float64(info.MemTotal) * 100.0
		stats.MemPercent = fmt.Sprintf("%.1f%%", memPct)
	} else {
		stats.MemPercent = "0%"
	}
	
	return stats, nil
}

func Inspect(cli *client.Client, ctx context.Context, resourceType, id string) (string, error) {
	var data interface{}
	var err error

	switch resourceType {
	case "container":
		data, err = cli.ContainerInspect(ctx, id)
	case "image":
		data, _, err = cli.ImageInspectWithRaw(ctx, id)
	case "volume":
		data, err = cli.VolumeInspect(ctx, id)
	case "network":
		data, err = cli.NetworkInspect(ctx, id, network.InspectOptions{})
	case "service":
		data, _, err = cli.ServiceInspectWithRaw(ctx, id, swarm.ServiceInspectOptions{})
	case "node":
		data, _, err = cli.NodeInspectWithRaw(ctx, id)
	default:
		return "", fmt.Errorf("unknown resource type: %s", resourceType)
	}

	if err != nil {
		return "", err
	}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func GetContainerStats(cli *client.Client, ctx context.Context, id string) (string, error) {
	resp, err := cli.ContainerStats(ctx, id, false)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var v interface{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return "", err
	}
	
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func HasTTY(cli *client.Client, ctx context.Context, id string) (bool, error) {
	c, err := cli.ContainerInspect(ctx, id)
	if err != nil {
		return false, err
	}
	return c.Config.Tty, nil
}

