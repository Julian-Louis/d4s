package command

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/jessym/d4s/internal/ui/common"
	"github.com/jessym/d4s/internal/ui/styles"
	"github.com/rivo/tview"
)

type CommandComponent struct {
	View *tview.InputField
	App  common.AppController
}

func NewCommandComponent(app common.AppController) *CommandComponent {
	c := tview.NewInputField().
		SetFieldBackgroundColor(styles.ColorBg).
		SetLabelColor(tcell.ColorWhite).
		SetFieldTextColor(styles.ColorFg).
		SetLabel("[#ffb86c::b]VIEW> [-:-:-]")
	
	c.SetBorder(true).
		SetBorderColor(tcell.NewRGBColor(144, 238, 144)). // Light green
		SetBackgroundColor(styles.ColorBg)
	
	comp := &CommandComponent{View: c, App: app}
	comp.setupHandlers()
	return comp
}

func (c *CommandComponent) setupHandlers() {
	c.View.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			c.Reset()
			c.App.SetActiveFilter("")
			
			c.App.RefreshCurrentView()
			c.App.SetFlashText("")
			
			// Restore focus and hide cmdline
			c.App.SetCmdLineVisible(false)
			c.App.RestoreFocus()
			return nil
		}
		return event
	})

	c.View.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			cmd := c.View.GetText()
			if strings.HasPrefix(cmd, "/") {
				if len(cmd) > 1 {
					filter := strings.TrimPrefix(cmd, "/")
					c.App.SetActiveFilter(filter)
					c.App.RefreshCurrentView()
					c.App.SetFlashText(fmt.Sprintf("Filter: %s", filter))
				}
			} else {
				c.App.ExecuteCmd(cmd)
			}
			
			c.Reset()
			
			// Restore focus and hide cmdline
			c.App.SetCmdLineVisible(false)
			c.App.RestoreFocus()
		}
	})
}

func (c *CommandComponent) Activate(initial string) {
	label := "[#ffb86c::b]CMD> [-:-:-]" // Orange for Command
	if strings.HasPrefix(initial, "/") {
		label = "[#ffb86c::b]FILTER> [-:-:-]" // Orange for Filter
	}
	c.View.SetLabel(label)
	c.View.SetText(initial)
	c.App.GetTviewApp().SetFocus(c.View)
}

func (c *CommandComponent) Reset() {
	c.View.SetText("")
	c.View.SetLabel("[#ffb86c::b]VIEW> [-:-:-]")
}

func (c *CommandComponent) HasFocus() bool {
	return c.View.HasFocus()
}

func (c *CommandComponent) SetFilter(filter string) {
	c.View.SetLabel("[#ffb86c::b]FILTER> [-:-:-]")
	c.View.SetText(filter)
}
