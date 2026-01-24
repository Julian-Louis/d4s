package images

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/jr-k/d4s/internal/dao"
	"github.com/jr-k/d4s/internal/ui/common"
	"github.com/jr-k/d4s/internal/ui/components/inspect"
	"github.com/jr-k/d4s/internal/ui/components/view"
	"github.com/jr-k/d4s/internal/ui/dialogs"
	"github.com/jr-k/d4s/internal/ui/styles"
)

var Headers = []string{"ID", "TAGS", "SIZE", "CONTAINERS", "CREATED"}

func Fetch(app common.AppController, v *view.ResourceView) ([]dao.Resource, error) {
	return app.GetDocker().ListImages()
}

func GetShortcuts() []string {
	return []string{
		common.FormatSCHeader("d", "Describe"),
		common.FormatSCHeader("p", "Prune"),
		common.FormatSCHeader("r", "Pull"),
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
	case 'r':
		PullAction(app, v)
		return nil
	case 'p':
		PruneAction(app)
		return nil
	case 'd':
		app.InspectCurrentSelection()
		return nil
	}
	
	if event.Key() == tcell.KeyEnter {
		EnterAction(app, v)
		return nil
	}

	return event
}

func EnterAction(app common.AppController, v *view.ResourceView) {
	id, err := v.GetSelectedID()
	if err != nil {
		return
	}

	// Find the full image object just to be sure what we have
	// Or just use the ID. The container view needs to know how to filter.
	// We'll pass the Image ID as scope.
	
	// Get Name if possible for nicer display label
	label := id
	for _, it := range v.Data {
		if it.GetID() == id {
			if im, ok := it.(dao.Image); ok {
				if im.Tags != "<none>" && im.Tags != "" {
					label = im.Tags
				}
			}
			break
		}
	}

	// We'll use a special scope type 'image'
	// But we need to make sure containers view supports it (already added beforehand)
	scope := &common.Scope{
		Type:       "image",
		Value:      id, // Use ID for robust filtering
		Label:      fmt.Sprintf("Image: %s", label),
		OriginView: "Images",
		Parent:     app.GetActiveScope(),
	}
	app.SetActiveScope(scope)
	app.SwitchTo(styles.TitleContainers)
}

func PruneAction(app common.AppController) {
	dialogs.ShowConfirmation(app, "PRUNE", "Images", func(force bool) {
		app.SetFlashPending("pruning images...")
		app.RunInBackground(func() {
			err := Prune(app)
			app.GetTviewApp().QueueUpdateDraw(func() {
				if err != nil {
					app.SetFlashError(fmt.Sprintf("%v", err))
				} else {
					app.SetFlashSuccess("pruned images")
					app.RefreshCurrentView()
				}
			})
		})
	})
}

func PullAction(app common.AppController, v *view.ResourceView) {
	ids, err := v.GetSelectedIDs()
	if err != nil || len(ids) == 0 {
		return
	}

	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	count := 0
	for _, item := range v.Data {
		if idMap[item.GetID()] {
			if img, ok := item.(dao.Image); ok {
				if img.RepoTag != "" && img.RepoTag != "<none>" {
					count++
					tag := img.RepoTag

					app.RunInBackground(func() {
						err := app.GetDocker().PullImage(tag)
						app.GetTviewApp().QueueUpdateDraw(func() {
							if err != nil {
								app.SetFlashError(fmt.Sprintf("Pull failed: %v", err))
							}
							app.RefreshCurrentView()
						})
					})
				}
			}
		}
	}

	if count > 0 {
		app.SetFlashPending(fmt.Sprintf("Pulling %d image(s)...", count))
		// Force refresh to show status
		go func() {
			time.Sleep(100 * time.Millisecond)
			app.GetTviewApp().QueueUpdateDraw(func() {
				app.RefreshCurrentView()
			})
		}()
	}
}

func Inspect(app common.AppController, id string) {
	content, err := app.GetDocker().Inspect("image", id)
	if err != nil {
		app.SetFlashError(fmt.Sprintf("%v", err))
		return
	}

	subject := id
	if len(id) > 12 {
		subject = id[:12]
	}

	// Resolve Tags from List
	images, err := app.GetDocker().ListImages()
	if err == nil {
		for _, item := range images {
			// dao.Image ID usually matches trimmed?
			// dao.Image GetID returns trimmed. 'id' passed here is usually full or trimmed?
			// app.InspectCurrentSelection passes resource.GetID().
			// Which is trimmed in dao.Image.List().
			// Double check? dao/docker/image/image.go: ID: strings.TrimPrefix(i.ID, "sha256:")
			// So it is full hex without prefix, likely 64 chars.
			if item.GetID() == id {
				if img, ok := item.(dao.Image); ok {
					if img.Tags != "" && img.Tags != "<none>" {
						subject = fmt.Sprintf("%s@%s", img.Tags, subject)
					}
				}
				break
			}
		}
	}

	app.OpenInspector(inspect.NewTextInspector("Describe image", subject, content, "json"))
}

func DeleteAction(app common.AppController, v *view.ResourceView) {
	ids, err := v.GetSelectedIDs()
	if err != nil { return }

	label := ids[0]
	if len(ids) == 1 {
		row, _ := v.Table.GetSelection()
		if row > 0 && row <= len(v.Data) {
			item := v.Data[row-1]
			if item.GetID() == ids[0] {
				cells := item.GetCells()
				if len(cells) > 1 {
					// Inside Confirmation Modal
					label = fmt.Sprintf("%s ([#00ffff]%s[yellow])", label, cells[1])
				}
			}
		}
	} else {
		label = fmt.Sprintf("%d items", len(ids))
	}

	dialogs.ShowConfirmation(app, "DELETE", label, func(force bool) {
		simpleAction := func(id string) error {
			return Remove(id, force, app)
		}
		app.PerformAction(simpleAction, "deleting", styles.ColorStatusRed)
	})
}

func Prune(app common.AppController) error {
	return app.GetDocker().PruneImages()
}

func Remove(id string, force bool, app common.AppController) error {
	return app.GetDocker().RemoveImage(id, force)
}
