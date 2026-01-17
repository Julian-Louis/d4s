package footer

import (
	"github.com/jr-k/d4s/internal/ui/styles"
	"github.com/rivo/tview"
)

type FlashComponent struct {
	View         *tview.TextView
	mainText     string
	appendedText string
}

func NewFlashComponent() *FlashComponent {
	f := tview.NewTextView()
	f.SetDynamicColors(true).SetTextColor(styles.ColorFlashFg).SetBackgroundColor(styles.ColorFlashBg)
	return &FlashComponent{View: f}
}

func (f *FlashComponent) SetText(text string) {
	f.mainText = text
	f.render()
}

// Appends text to existing content temporarily.
// It uses a dedicated slot, so multiple appends overwrite each other (last one wins),
// but it stays separate from the main text (breadcrumb/status).
func (f *FlashComponent) Append(text string) {
	f.appendedText = text
	f.render()
}

func (f *FlashComponent) ClearAppend() {
	f.appendedText = ""
	f.render()
}

func (f *FlashComponent) render() {
	full := f.mainText
	if f.appendedText != "" {
		// Ensure separation
		full += " " + f.appendedText
	}
	f.View.SetText(full)
}
