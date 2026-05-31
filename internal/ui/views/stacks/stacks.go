package stacks

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/jr-k/d4s/internal/dao"
	"github.com/jr-k/d4s/internal/ui/common"
	"github.com/jr-k/d4s/internal/ui/components/inspect"
	"github.com/jr-k/d4s/internal/ui/components/view"
	"github.com/jr-k/d4s/internal/ui/dialogs"
	"github.com/jr-k/d4s/internal/ui/styles"
)

var Headers = []string{"NAME", "READY", "STATUS"}

func Fetch(app common.AppController, v *view.ResourceView) ([]dao.Resource, error) {
	return app.GetDocker().ListStacks()
}

func GetShortcuts() []string {
	return []string{
		common.FormatSCHeader("enter", "Services"),
		common.FormatSCHeader("p", "Ps"),
		common.FormatSCHeader("ctrl-d", "Delete"),
	}
}

func InputHandler(v *view.ResourceView, event *tcell.EventKey) *tcell.EventKey {
	app := v.App

	if event.Key() == tcell.KeyCtrlD {
		DeleteAction(app, v)
		return nil
	}

	switch event.Rune() {
	case 'p':
		app.InspectCurrentSelection()
		return nil
	}

	if event.Key() == tcell.KeyEnter {
		NavigateToServices(app, v)
		return nil
	}

	return event
}

func NavigateToServices(app common.AppController, v *view.ResourceView) {
	id, err := v.GetSelectedID()
	if err != nil {
		return
	}

	app.SetActiveScope(&common.Scope{
		Type:       "stack",
		Value:      id,
		Label:      id,
		OriginView: styles.TitleStacks,
	})

	app.SwitchTo(styles.TitleServices)
}

func DeleteAction(app common.AppController, v *view.ResourceView) {
	ids, err := v.GetSelectedIDs()
	if err != nil || len(ids) == 0 {
		return
	}

	label := ids[0]
	if len(ids) > 1 {
		label = fmt.Sprintf("%d items", len(ids))
	}

	dialogs.ShowConfirmation(app, "DELETE", label, func(force bool) {
		simpleAction := func(id string) error {
			return app.GetDocker().RemoveStack(id)
		}
		app.PerformAction(simpleAction, "deleting", styles.ColorStatusRed)
	})
}

func Inspect(app common.AppController, id string) {
	inspector := inspect.NewTextInspector("Stack ps", id, fmt.Sprintf(" [%s]Loading stack tasks...\n", styles.TagAccent), "text")
	app.OpenInspector(inspector)

	app.RunInBackground(func() {
		tasks, err := app.GetDocker().ListTasks()
		if err != nil {
			app.GetTviewApp().QueueUpdateDraw(func() {
				inspector.Viewer.Update(fmt.Sprintf("Error: %v", err), "text")
			})
			return
		}

		nodes, _ := app.GetDocker().ListNodes()
		nodeNames := make(map[string]string)
		for _, n := range nodes {
			if node, ok := n.(dao.Node); ok {
				nodeNames[node.ID] = node.Hostname
			}
		}

		var lines []string
		var filtered []dao.Task
		for _, t := range tasks {
			if task, ok := t.(dao.Task); ok {
				if strings.HasPrefix(task.Name, id+"_") || strings.HasPrefix(task.Name, id+".") {
					filtered = append(filtered, task)
				}
			}
		}

		if len(filtered) == 0 {
			lines = append(lines, "# No tasks for this stack")
		} else {
			lines = append(lines, fmt.Sprintf("%-14s %-30s %-20s %-15s %-15s %s", "ID", "NAME", "NODE", "DESIRED STATE", "CURRENT STATE", "ERROR"))
			lines = append(lines, strings.Repeat("-", 120))

			for _, t := range filtered {
				taskID := t.ID
				if len(taskID) > 12 {
					taskID = taskID[:12]
				}
				lines = append(lines, fmt.Sprintf("%-14s %-30s %-20s %-15s %-15s %s", taskID, t.Name, t.Node, t.DesiredState, t.CurrentState, t.Error))
			}
		}

		app.GetTviewApp().QueueUpdateDraw(func() {
			inspector.Viewer.Update(strings.Join(lines, "\n"), "text")
		})
	})
}
