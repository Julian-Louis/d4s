package header

import (
	"fmt"

	"github.com/jessym/d4s/internal/dao"
	"github.com/jessym/d4s/internal/ui/styles"
	"github.com/rivo/tview"
)

type HeaderComponent struct {
	View *tview.Table
	LastStats dao.HostStats
}

func NewHeaderComponent() *HeaderComponent {
	h := tview.NewTable().SetBorders(false)
	h.SetBackgroundColor(styles.ColorBg)
	return &HeaderComponent{
		View: h,
	}
}

func (h *HeaderComponent) Update(stats dao.HostStats, shortcuts []string) {
	// Merge with existing stats to avoid flickering "..."
	// If new stats have "...", check if we have better old values
	if stats.CPUPercent == "..." && h.LastStats.CPUPercent != "" && h.LastStats.CPUPercent != "..." {
		stats.CPUPercent = h.LastStats.CPUPercent
	}
	if stats.MemPercent == "..." && h.LastStats.MemPercent != "" && h.LastStats.MemPercent != "..." {
		stats.MemPercent = h.LastStats.MemPercent
	}
	
	// Save for next time
	h.LastStats = stats

	h.View.Clear()
	h.View.SetBackgroundColor(styles.ColorBg) // Ensure no black block
	
	logo := []string{
		"[#ffb86c]    ____  __ __ ____",
		"[#ffb86c]   / __ \\/ // // __/",
		"[#ffb86c]  / /_/ / // /_\\ \\ ",
		"[#ffb86c] /_____/_//_/____/ ",
		"",
		"",
	}
	
	// Build CPU display with cores and percentage
	cpuDisplay := fmt.Sprintf("%s cores", stats.CPU)
	if stats.CPUPercent != "" && stats.CPUPercent != "N/A" && stats.CPUPercent != "..." {
		cpuDisplay += fmt.Sprintf(" (%s)", stats.CPUPercent)
	} else if stats.CPUPercent == "..." {
		cpuDisplay += " [dim](...)"
	}
	
	// Build Mem display with total and percentage
	memDisplay := stats.Mem
	if stats.MemPercent != "" && stats.MemPercent != "N/A" && stats.MemPercent != "..." {
		memDisplay += fmt.Sprintf(" (%s)", stats.MemPercent)
	} else if stats.MemPercent == "..." {
		memDisplay += " [dim](...)"
	}
	
	lines := []string{
		fmt.Sprintf("[#8be9fd]Host:    [white]%s", stats.Hostname),
		fmt.Sprintf("[#8be9fd]D4s Rev: [white]v%s", stats.D4SVersion),
		fmt.Sprintf("[#8be9fd]User:    [white]%s", stats.User),
		fmt.Sprintf("[#8be9fd]Engine:  [white]%s [dim](v%s)", stats.Name, stats.Version),
		fmt.Sprintf("[#8be9fd]CPU:     [white]%s", cpuDisplay),
		fmt.Sprintf("[#8be9fd]Mem:     [white]%s", memDisplay),
	}

	// Layout Header
	// Col 0: Stats
	for i, line := range lines {
		// Add padding to the right of stats
		cell := tview.NewTableCell(line).
			SetBackgroundColor(styles.ColorBg).
			SetAlign(tview.AlignLeft).
			SetExpansion(0) // Fixed width
		h.View.SetCell(i, 0, cell)
	}
	
	// Spacer Column (between Stats and Shortcuts)
	// A fixed width column to separate them nicely (tripled size ~21 spaces)
	spacerWidth := "                     " 
	for i := 0; i < 6; i++ {
		h.View.SetCell(i, 1, tview.NewTableCell(spacerWidth).SetBackgroundColor(styles.ColorBg))
	}
	
	// Center Columns: Shortcuts
	// Max 6 per column (matches header height)
	const maxPerCol = 6
	
	colIndex := 2 // Start at 2 (0=Stats, 1=Spacer)
	for i := 0; i < len(shortcuts); i += maxPerCol {
		end := i + maxPerCol
		if end > len(shortcuts) {
			end = len(shortcuts)
		}
		
		chunk := shortcuts[i:end]
		
		// Fill all 6 rows for this column to ensure background color
		for row := 0; row < maxPerCol; row++ {
			text := ""
			if row < len(chunk) {
				text = chunk[row] + "  " // Content + padding
			}
			
			cell := tview.NewTableCell(text).
				SetAlign(tview.AlignLeft).
				SetExpansion(0). // Compact columns
				SetBackgroundColor(styles.ColorBg)
			h.View.SetCell(row, colIndex, cell)
		}
		colIndex++
	}
	
	// Flexible Spacer Column (pushes logo to right)
	// Use an empty cell with Expansion 1. Need to set it on at least one row.
	// Set on all rows to be safe with background
	for i := 0; i < 6; i++ {
		h.View.SetCell(i, colIndex, tview.NewTableCell("").SetExpansion(1).SetBackgroundColor(styles.ColorBg))
	}
	colIndex++

	// Right Column: Logo
	for i, line := range logo {
		cell := tview.NewTableCell(line).
			SetAlign(tview.AlignRight).
			SetBackgroundColor(styles.ColorBg).
			SetExpansion(0) // Fixed width
		h.View.SetCell(i, colIndex, cell)
	}
}
