package dialogs

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/jessym/d4s/internal/ui/common"
	"github.com/jessym/d4s/internal/ui/styles"
	"github.com/rivo/tview"
)

// ShowConfirmation shows a modal asking to type "Yes Please!" and allows forcing
func ShowConfirmation(app common.AppController, actionName, item string, onConfirm func(force bool)) {
	// Center the dialog
	dialogWidth := 60
	dialogHeight := 16 
	
	pages := app.GetPages()
	tviewApp := app.GetTviewApp()

	text := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("\n[red::b] DANGER ZONE \n\n[white::-]You are about to %s:\n[yellow]%s[white]\n\nType exactly: [red::b]Yes Please![white::-]", actionName, item))
	text.SetBackgroundColor(tcell.ColorBlack)
	
	input := tview.NewInputField().
		SetFieldBackgroundColor(styles.ColorSelectBg).
		SetFieldTextColor(tcell.ColorRed).
		SetLabel("Confirmation: ").
		SetLabelColor(tcell.ColorWhite)
	input.SetBackgroundColor(tcell.ColorBlack)
	
	// Force Checkbox
	force := false
	checkbox := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[white][ ] Force (Tab to focus, Space to toggle)")
	checkbox.SetBackgroundColor(tcell.ColorBlack)

	updateCheckbox := func(focused bool) {
		prefix := "[ ]"
		if force {
			prefix = "[red][X]"
		}
		
		color := "[white]"
		if focused {
			color = "[#ffb86c]" // Orange focus
		}

		checkbox.SetText(fmt.Sprintf("%s%s Force (Tab to focus, Space to toggle)", color, prefix))
	}

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(text, 0, 1, false).
		AddItem(input, 3, 1, true).
		AddItem(checkbox, 1, 1, false) // 1 line for checkbox
	
	flex.SetBorder(true).
		SetTitle(" Are you sure? ").
		SetTitleColor(tcell.ColorRed).
		SetBorderColor(tcell.ColorRed).
		SetBackgroundColor(tcell.ColorBlack)

	// Center on screen
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(flex, dialogHeight, 1, true).
			AddItem(nil, 0, 1, false), dialogWidth, 1, true).
		AddItem(nil, 0, 1, false)

	// Restore focus helper
	closeModal := func() {
		pages.RemovePage("confirm")
		// We assume we want to focus back on the table or pages
		tviewApp.SetFocus(pages) 
	}

	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			if input.GetText() == "Yes Please!" {
				closeModal()
				onConfirm(force)
			} else {
				app.SetFlashText("[red]Confirmation mismatch. Action cancelled.")
				closeModal()
			}
		} else if key == tcell.KeyEsc {
			closeModal()
		} else if key == tcell.KeyTab {
			// Switch to Checkbox
			updateCheckbox(true)
			tviewApp.SetFocus(checkbox)
		}
	})

	checkbox.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			// Switch back to Input
			updateCheckbox(false)
			tviewApp.SetFocus(input)
			return nil
		}
		if event.Rune() == ' ' {
			force = !force
			updateCheckbox(true)
			return nil
		}
		if event.Key() == tcell.KeyEsc {
			closeModal()
			return nil
		}
		if event.Key() == tcell.KeyEnter {
			if input.GetText() == "Yes Please!" {
				closeModal()
				onConfirm(force)
			}
			return nil
		}
		return event
	})

	pages.AddPage("confirm", modal, true, true)
	tviewApp.SetFocus(input)
}
