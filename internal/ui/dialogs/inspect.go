package dialogs

import (
	"fmt"
	"io"
	"os/exec"
	"runtime"

	"github.com/gdamore/tcell/v2"
	"github.com/jessym/d4s/internal/ui/common"
	"github.com/jessym/d4s/internal/ui/styles"
	"github.com/rivo/tview"
)

func ShowInspectModal(app common.AppController, title, content string) {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf("[green]%s", content)).
		SetScrollable(true)
	
	tv.SetBorder(true).SetTitle(fmt.Sprintf(" Inspect %s ", title)).SetTitleColor(styles.ColorTitle)
	tv.SetBackgroundColor(styles.ColorBg)
	
	pages := app.GetPages()
	tviewApp := app.GetTviewApp()
	
	// Navigation for Inspect
	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.RemovePage("inspect")
			// Restore focus
			tviewApp.SetFocus(pages)
			return nil
		}
		if event.Rune() == 'c' {
			// Copy to clipboard (Cross-platform)
			var cmd *exec.Cmd
			switch runtime.GOOS {
			case "darwin":
				cmd = exec.Command("pbcopy")
			case "windows":
				cmd = exec.Command("clip")
			default: // linux
				// Try xclip, fallback to xsel? Just xclip for now
				cmd = exec.Command("xclip", "-selection", "clipboard")
			}

			if cmd == nil {
				app.SetFlashText("[red]Clipboard not supported on this OS")
				return nil
			}

			stdin, err := cmd.StdinPipe()
			if err != nil {
				app.SetFlashText("[red]Copy error: stdin pipe")
				return nil
			}
			go func() {
				defer stdin.Close()
				io.WriteString(stdin, content)
			}()
			if err := cmd.Run(); err != nil {
				app.SetFlashText(fmt.Sprintf("[red]Copy error: %v (install xclip/pbcopy?)", err))
			} else {
				app.SetFlashText("[green]Copied to clipboard!")
			}
			return nil
		}
		return event
	})

	pages.AddPage("inspect", tv, true, true)
	tviewApp.SetFocus(tv)
}
