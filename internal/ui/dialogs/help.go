package dialogs

import (
	"github.com/gdamore/tcell/v2"
	"github.com/jr-k/d4s/internal/ui/common"
	"github.com/jr-k/d4s/internal/ui/styles"
	"github.com/rivo/tview"
)

func NewHelpView(app common.AppController) tview.Primitive {
	helpTable := tview.NewTable()
	helpTable.SetBorders(false)
	helpTable.SetBackgroundColor(tcell.ColorBlack)

	type helpRow struct {
		label1, alias1 string
		label2, alias2 string
	}

	rows := []helpRow{
		{label1: "[orange::b]GLOBAL"},
		{label1: "Command", alias1: "[#5f87ff]:[-]", label2: "Help", alias2: "[#5f87ff]?[-]"},
		{label1: "Filter", alias1: "[#5f87ff]/[-]", label2: "Back/Clear", alias2: "[#5f87ff]esc[-]"},
		{label1: "Copy", alias1: "[#5f87ff]c[-]", label2: "Unselect All", alias2: "[#5f87ff]u[-]"},
		{},
		{label1: "[orange::b]DOCKER"},
		{label1: "Containers", alias1: "[#5f87ff]:c[-]", label2: "Images", alias2: "[#5f87ff]:i[-]"},
		{label1: "Volumes", alias1: "[#5f87ff]:v[-]", label2: "Networks", alias2: "[#5f87ff]:n[-]"},
		{label1: "Compose", alias1: "[#5f87ff]:p[-]"},
		{},
		{label1: "[orange::b]SWARM"},
		{label1: "Services", alias1: "[#5f87ff]:s[-]", label2: "Nodes", alias2: "[#5f87ff]:no[-]"},
		{},
		{label1: "[orange::b]NAVIGATION"},
		{label1: "Navigate", alias1: "[#5f87ff]←/→[-], [#5f87ff]j/k[-]", label2: "Drill Down", alias2: "[#5f87ff]enter[-]"},
		{label1: "Sort Column", alias1: "[#5f87ff]shift ←/→[-]", label2: "Toggle Order", alias2: "[#5f87ff]shift ↑/↓[-]"},
	}

	for i, row := range rows {
		cells := []struct {
			text      string
			align     int
			expansion int
		}{
			{text: row.label1, align: tview.AlignLeft, expansion: 0},
			{text: row.alias1, align: tview.AlignRight, expansion: 0},
			{text: "", align: tview.AlignLeft, expansion: 1}, // Column spacer
			{text: row.label2, align: tview.AlignLeft, expansion: 0},
			{text: row.alias2, align: tview.AlignRight, expansion: 0},
		}

		for j, cellData := range cells {
			cell := tview.NewTableCell(cellData.text).
				SetTextColor(tcell.ColorWhite).
				SetAlign(cellData.align).
				SetExpansion(cellData.expansion)

			helpTable.SetCell(i, j, cell)
		}
	}

	helpBox := tview.NewFrame(helpTable).
		SetBorders(1, 1, 1, 1, 0, 0).
		AddText(" Help ", true, tview.AlignCenter, styles.ColorTitle).
		SetBorderPadding(0, 0, 2, 2)
	helpBox.SetBorder(true).SetBorderColor(styles.ColorTitle).SetBackgroundColor(tcell.ColorBlack)

	// Center Modal
	helpFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(helpBox, 30, 1, true).
			AddItem(nil, 0, 1, false), 90, 1, true).
		AddItem(nil, 0, 1, false)

	helpFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Rune() == 'q' {
			app.GetPages().RemovePage("help")
			// Restore focus
			app.GetTviewApp().SetFocus(app.GetPages())
			app.UpdateShortcuts()
			return nil
		}
		return event
	})

	return helpFlex
}
