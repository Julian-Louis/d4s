package dialogs

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/jr-k/d4s/internal/ui/common"
	"github.com/jr-k/d4s/internal/ui/styles"
	"github.com/rivo/tview"
)

type EnvItem struct {
	Key      string
	Value    string
	Selected bool
}

func ShowEnvEditor(app common.AppController, title string, items []EnvItem, onConfirm func(envVars []string)) {
	dialogWidth := 70
	dialogHeight := 10 + len(items)
	if dialogHeight > 28 {
		dialogHeight = 28
	}
	if dialogHeight < 14 {
		dialogHeight = 14
	}

	pages := app.GetPages()
	tviewApp := app.GetTviewApp()

	// Sort items by key
	sort.Slice(items, func(i, j int) bool {
		return items[i].Key < items[j].Key
	})

	// Track selections
	selections := make([]bool, len(items))
	for i, item := range items {
		selections[i] = item.Selected
	}

	currentIndex := 0

	// --- Add new env form ---
	placeholderStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(140, 140, 160)).
		Background(tcell.NewRGBColor(50, 52, 68))

	nameInput := tview.NewInputField().
		SetFieldBackgroundColor(styles.ColorSelectBg).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabel(" Name: ").
		SetLabelColor(styles.ColorWhite).
		SetPlaceholder("MY_VAR").
		SetPlaceholderStyle(placeholderStyle)
	nameInput.SetBackgroundColor(styles.ColorBlack)

	valueInput := tview.NewInputField().
		SetFieldBackgroundColor(styles.ColorSelectBg).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabel(" Value: ").
		SetLabelColor(styles.ColorWhite).
		SetPlaceholder("my_value").
		SetPlaceholderStyle(placeholderStyle)
	valueInput.SetBackgroundColor(styles.ColorBlack)

	addButton := tview.NewButton("  Add  ").
		SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(styles.ColorStatusBlue)).
		SetActivatedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(styles.ColorStatusGreen))
	addButton.SetBackgroundColor(styles.ColorBlack)

	spacer := tview.NewBox().SetBackgroundColor(styles.ColorBlack)

	addFormRow := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nameInput, 0, 1, true).
		AddItem(valueInput, 0, 1, false).
		AddItem(spacer, 2, 0, false).
		AddItem(addButton, 10, 0, false)

	separator := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[" + styles.TagDim + "]" + strings.Repeat("─", 66) + "[-]").
		SetTextAlign(tview.AlignCenter)
	separator.SetBackgroundColor(styles.ColorBlack)

	// --- Env list with checkboxes ---
	list := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	list.SetBackgroundColor(styles.ColorBlack)

	updateList := func() {
		if len(items) == 0 {
			list.SetText(fmt.Sprintf("[%s]  No environment variables[-]", styles.TagDim))
			return
		}
		var content string
		for i, item := range items {
			checkbox := "[ ]"
			if selections[i] {
				checkbox = "[" + "✔" + "]"
			}

			color := fmt.Sprintf("[%s]", styles.TagFg)
			if i == currentIndex {
				color = fmt.Sprintf("[%s]", styles.TagAccent)
				checkbox = "> " + checkbox
			} else {
				checkbox = "  " + checkbox
			}

			display := fmt.Sprintf("%s=%s", item.Key, item.Value)
			if len(display) > 60 {
				display = display[:57] + "..."
			}

			content += fmt.Sprintf("%s%s %s[-]\n", color, checkbox, display)
		}
		list.SetText(content)
	}
	updateList()

	// Help text
	helpText := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf("[%s]tab switch focus • ↑/↓ navigate • space toggle • enter/esc confirm", styles.TagDim)).
		SetTextAlign(tview.AlignCenter)
	helpText.SetBackgroundColor(styles.ColorBlack)

	// Layout
	content := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(addFormRow, 1, 0, true).
		AddItem(separator, 1, 0, false).
		AddItem(list, 0, 1, false).
		AddItem(helpText, 1, 0, false)

	content.SetBorder(true).
		SetTitle(" " + title + " ").
		SetTitleColor(styles.ColorTitle).
		SetBorderColor(styles.ColorTitle).
		SetBackgroundColor(styles.ColorBlack)

	// Center on screen
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(content, dialogHeight, 1, true).
			AddItem(nil, 0, 1, false), dialogWidth, 1, true).
		AddItem(nil, 0, 1, false)

	closeModal := func() {
		pages.RemovePage("env_editor")
		tviewApp.SetFocus(pages)
		app.UpdateShortcuts()
	}

	collectResults := func() {
		var envVars []string
		for i, item := range items {
			if selections[i] {
				envVars = append(envVars, fmt.Sprintf("%s=%s", item.Key, item.Value))
			}
		}
		closeModal()
		onConfirm(envVars)
	}

	addEnvVar := func() {
		name := strings.TrimSpace(nameInput.GetText())
		value := valueInput.GetText()
		if name == "" {
			return
		}

		// Check if key already exists, if so update it
		found := false
		for i, item := range items {
			if item.Key == name {
				items[i].Value = value
				selections[i] = true
				found = true
				break
			}
		}

		if !found {
			items = append(items, EnvItem{Key: name, Value: value, Selected: true})
			selections = append(selections, true)
		}

		// Re-sort
		type indexedItem struct {
			item     EnvItem
			selected bool
		}
		indexed := make([]indexedItem, len(items))
		for i := range items {
			indexed[i] = indexedItem{items[i], selections[i]}
		}
		sort.Slice(indexed, func(i, j int) bool {
			return indexed[i].item.Key < indexed[j].item.Key
		})
		for i := range indexed {
			items[i] = indexed[i].item
			selections[i] = indexed[i].selected
		}

		nameInput.SetText("")
		valueInput.SetText("")
		updateList()

		// Recalculate dialog height
		newHeight := 10 + len(items)
		if newHeight > 28 {
			newHeight = 28
		}
		if newHeight < 14 {
			newHeight = 14
		}
		_ = newHeight

		tviewApp.SetFocus(nameInput)
	}

	// Focus management: 0=nameInput, 1=valueInput, 2=addButton, 3=list
	focusTarget := 0

	setFocus := func(target int) {
		focusTarget = target
		switch target {
		case 0:
			tviewApp.SetFocus(nameInput)
		case 1:
			tviewApp.SetFocus(valueInput)
		case 2:
			tviewApp.SetFocus(addButton)
		case 3:
			tviewApp.SetFocus(list)
		}
	}

	addOrConfirm := func() {
		name := strings.TrimSpace(nameInput.GetText())
		if name == "" {
			collectResults()
		} else {
			addEnvVar()
		}
	}

	// Input handlers
	nameInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			collectResults()
			return nil
		}
		if event.Key() == tcell.KeyTab || event.Key() == tcell.KeyDown {
			setFocus(1)
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			if len(items) > 0 {
				setFocus(3)
			}
			return nil
		}
		if event.Key() == tcell.KeyEnter {
			addOrConfirm()
			return nil
		}
		return event
	})

	valueInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			collectResults()
			return nil
		}
		if event.Key() == tcell.KeyTab || event.Key() == tcell.KeyDown {
			setFocus(2)
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			setFocus(0)
			return nil
		}
		if event.Key() == tcell.KeyEnter {
			addOrConfirm()
			return nil
		}
		return event
	})

	addButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			collectResults()
			return nil
		}
		if event.Key() == tcell.KeyTab || event.Key() == tcell.KeyDown {
			if len(items) > 0 {
				setFocus(3)
			} else {
				setFocus(0)
			}
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			setFocus(1)
			return nil
		}
		if event.Key() == tcell.KeyEnter {
			addOrConfirm()
			return nil
		}
		return event
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			collectResults()
			return nil
		case tcell.KeyEnter:
			collectResults()
			return nil
		case tcell.KeyTab:
			setFocus(0)
			return nil
		case tcell.KeyBacktab:
			setFocus(2)
			return nil
		case tcell.KeyUp:
			if currentIndex > 0 {
				currentIndex--
				updateList()
			} else {
				setFocus(2)
			}
			return nil
		case tcell.KeyDown:
			if currentIndex < len(items)-1 {
				currentIndex++
				updateList()
			}
			return nil
		}

		if event.Rune() == ' ' {
			if len(items) > 0 {
				selections[currentIndex] = !selections[currentIndex]
				updateList()
			}
			return nil
		}

		return event
	})

	pages.AddPage("env_editor", modal, true, true)
	setFocus(0)
	_ = focusTarget
	app.UpdateShortcuts()
}
