package services

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jessym/d4s/internal/dao"
	"github.com/jessym/d4s/internal/ui/common"
	"github.com/jessym/d4s/internal/ui/dialogs"
)

var Headers = []string{"ID", "NAME", "IMAGE", "MODE", "REPLICAS", "PORTS"}

func Fetch(app common.AppController) ([]dao.Resource, error) {
	services, err := app.GetDocker().ListServices()
	if err != nil {
		return nil, err
	}

	// Filter by Node Scope
	scope := app.GetActiveScope()
	if scope != nil && scope.Type == "node" {
		nodeID := scope.Value
		var filtered []dao.Resource
		
		// We need to check which services have tasks on this node
		// This requires an extra call to list tasks for this node
		tasks, err := app.GetDocker().ListTasksForNode(nodeID)
		if err == nil {
			serviceIDs := make(map[string]bool)
			for _, task := range tasks {
				serviceIDs[task.ServiceID] = true
			}
			
			for _, s := range services {
				if serviceIDs[s.GetID()] {
					filtered = append(filtered, s)
				}
			}
			return filtered, nil
		}
	}
	
	return services, nil
}

func Remove(id string, force bool, app common.AppController) error {
	return app.GetDocker().RemoveService(id)
}

func Scale(app common.AppController, id string, currentReplicas string) {
	if parts := strings.Split(currentReplicas, "/"); len(parts) == 2 {
		currentReplicas = parts[1]
	}

	dialogs.ShowInput(app, "Scale Service", "Replicas:", currentReplicas, func(text string) {
		replicas, err := strconv.ParseUint(text, 10, 64)
		if err != nil {
			app.SetFlashText("[red]Invalid number")
			return
		}
		
		app.SetFlashText(fmt.Sprintf("[yellow]Scaling %s to %d...", id, replicas))
		
		go func() {
			err := app.GetDocker().ScaleService(id, replicas)
			app.GetTviewApp().QueueUpdateDraw(func() {
				if err != nil {
					app.SetFlashText(fmt.Sprintf("[red]Scale Error: %v", err))
				} else {
					app.SetFlashText(fmt.Sprintf("[green]Service scaled to %d", replicas))
					app.RefreshCurrentView()
				}
			})
		}()
	})
}

