package footer

import (
	"github.com/gdamore/tcell/v2"
	"github.com/jessym/d4s/internal/ui/styles"
	"github.com/rivo/tview"
)

type FooterComponent struct {
	View *tview.TextView
}

func NewFooterComponent() *FooterComponent {
	f := tview.NewTextView()
	f.SetDynamicColors(true).SetBackgroundColor(styles.ColorBg)
	return &FooterComponent{View: f}
}

func (f *FooterComponent) SetText(text string) {
	f.View.SetText(text)
}

type FlashComponent struct {
	View *tview.TextView
}

func NewFlashComponent() *FlashComponent {
	f := tview.NewTextView()
	f.SetTextColor(tcell.NewRGBColor(95, 135, 255)).SetBackgroundColor(styles.ColorBg) // Royal Blueish
	return &FlashComponent{View: f}
}

func (f *FlashComponent) SetText(text string) {
	f.View.SetText(text)
}

func (f *FlashComponent) Clear() {
	f.View.SetText("")
}
