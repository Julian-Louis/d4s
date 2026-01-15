package dialogs

import (
	"github.com/gdamore/tcell/v2"
	"github.com/jessym/d4s/internal/ui/common"
	"github.com/jessym/d4s/internal/ui/styles"
	"github.com/rivo/tview"
)

// This logic was in app.go, moving here.
// But it depends on DAO.
// For now, we will assume PerformEnv is called on AppController which delegates to this?
// Or AppController implements PerformEnv (which is in app.go).
// Wait, details.go contained `PerformEnv`? Yes.
// But `PerformEnv` calls `a.Docker.GetContainerEnv`.
// So it needs access to Docker Client.
// `AppController` doesn't expose DockerClient.
// We should add `GetDockerClient()` to `AppController`?
// Or keep `PerformEnv` in `app.go` and just use `ShowTextView` from here?
// The file `details.go` had `PerformEnv`, `PerformStats`, `ShowTextView`.
// If we want to keep logic here, we need access to Docker.

// Let's create `ShowTextView` here and import it in `app.go`.
// And keep `PerformEnv` in `app.go` (Controller Logic) while `ShowTextView` is View Logic.

func ShowTextView(app common.AppController, title, content string) {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetScrollable(true).
		SetText(content)
	
	tv.SetBorder(true).SetTitle(title).SetTitleColor(styles.ColorTitle)
	tv.SetBackgroundColor(styles.ColorBg)
	
	pages := app.GetPages()
	tviewApp := app.GetTviewApp()
	
	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.RemovePage("textview")
			tviewApp.SetFocus(pages)
			return nil
		}
		return event
	})
	
	pages.AddPage("textview", tv, true, true)
	tviewApp.SetFocus(tv)
}
