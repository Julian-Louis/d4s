package styles

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/lucasb-eyer/go-colorful"
)

// Indigo / Dracula-like / K9s Color Palette (Restored)
var (
	// Main Background (Indigo/Dark Blue)
	ColorBg          = tcell.Color16 // Dark Indigo
	ColorFg          = tcell.ColorWhite
	ColorTableBorder = tcell.NewRGBColor(137, 206, 250) // Blue

	ColorBlack = tcell.Color16
	ColorWhite = tcell.NewRGBColor(255, 255, 255)

	ColorHeader = tcell.NewRGBColor(255, 255, 255) // White
	ColorHeaderFocus = tcell.NewRGBColor(255, 184, 108) // Orange

	// Header
	ColorTitle       = tcell.NewRGBColor(189, 147, 249) // Purple
	ColorIdle	 	 = tcell.NewRGBColor(137, 206, 250) // Blue
	ColorTeal 		 = tcell.NewRGBColor(94, 175, 175)  // Teal/Cyan
	
	// Footer
	ColorFooterBg    = tcell.NewRGBColor(68, 71, 90)    // Selection Gray
	ColorFooterFg    = tcell.NewRGBColor(248, 248, 242) // White

	// Flash
	ColorFlashFg 	 = tcell.NewRGBColor(95, 135, 255) // Royal Blueish
	ColorFlashBg 	 = tcell.Color16 // Dark Indigo
	
	// Table
	ColorSelectBg    = tcell.NewRGBColor(68, 71, 90)    // Selection Gray
	ColorSelectFg    = tcell.ColorWhite
	ColorValue       = tcell.ColorWhite
	
	// Added for compatibility with view.go
	ColorSelect = tcell.NewRGBColor(153, 251, 152) // Green
	
	// Text Colors
	ColorDim         = tcell.ColorDimGray  // Comment/Dim
	ColorAccent      = tcell.NewRGBColor(255, 165, 3) // Orange
	ColorAccentLight = tcell.NewRGBColor(255, 184, 108) // Light Orange
	
	// Status
	ColorLogo        = tcell.NewRGBColor(255, 184, 108) // Orange
	ColorError       = tcell.NewRGBColor(255, 85, 85)   // Red
	ColorInfo        = tcell.NewRGBColor(80, 250, 123)  // Green
	
	// Rows Status
	ColorStatusGreen = tcell.NewRGBColor(80, 250, 123)  // Green
	ColorStatusRed   = tcell.NewRGBColor(255, 85, 85)   // Red
	ColorStatusGray  = tcell.NewRGBColor(119, 136, 153)  // Gray
	ColorStatusYellow = tcell.NewRGBColor(241, 250, 140) // Yellow
	ColorStatusOrange = tcell.NewRGBColor(255, 140, 3) // Orange
	ColorStatusBlue = tcell.NewRGBColor(1, 123, 255)  // Blue (lighter)
	ColorStatusPurple = tcell.NewRGBColor(103, 35, 186) // Purple

	ColorStatusRedDarkBg = tcell.NewRGBColor(46, 30, 30) // Red
	ColorStatusGreenDarkBg = tcell.NewRGBColor(32, 46, 30)  // Green
	ColorStatusGrayDarkBg   = tcell.NewRGBColor(60, 64, 90)    // Darker Gray/Bluish
	ColorStatusYellowDarkBg = tcell.NewRGBColor(46, 46, 30)  // Darker Yellow
	ColorStatusOrangeDarkBg = tcell.NewRGBColor(46, 39, 30)   // Darker Orange/Brown
	ColorStatusBlueDarkBg   = tcell.NewRGBColor(30, 37, 46)    // Darker Blue
	ColorStatusPurpleDarkBg = tcell.NewRGBColor(38, 30, 46)    // Darker Purple
)

// Tview markup-compatible color hex strings.
// These track the tcell.Color variables above so that format strings
// like fmt.Sprintf("[%s]text", styles.TagFg) produce the correct color
// even after InvertColors() is called.
var (
	TagFg     = colorToTag(ColorFg)       // replaces [white]
	TagBg     = colorToTag(ColorBg)       // replaces [black]
	TagAccent = colorToTag(ColorAccent)   // replaces [orange]
	TagAccentLight = colorToTag(ColorAccentLight) // replaces [light orange]
	TagIdle   = colorToTag(ColorIdle)     // replaces [blue] (info blue)
	TagDim    = colorToTag(ColorDim)      // replaces [gray] / [dim]
	TagError  = colorToTag(ColorError)    // replaces [red]
	TagInfo   = colorToTag(ColorInfo)     // replaces [green]
	TagTitle  = colorToTag(ColorTitle)    // purple title
	TagCyan   = "#00ffff"                 // breadcrumb/title cyan
	TagPink   = "#ff00ff"                 // scope label pink
	TagFilter = "#bd93f9"                 // filter badge purple
	TagSCKey       = "#2090ff"                 // shortcut key blue
)

// colorToTag converts a tcell.Color to a tview-compatible hex tag like "#rrggbb".
func colorToTag(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// refreshTags re-derives all Tag* strings from the current Color* variables.
// standalone colors (cyan, pink, etc.) are inverted manually.
func refreshTags(invert bool) {
	TagFg = colorToTag(ColorFg)
	TagBg = colorToTag(ColorBg)
	TagAccent = colorToTag(ColorAccent)
	TagAccentLight = colorToTag(ColorAccentLight)
	TagIdle = colorToTag(ColorIdle)
	TagDim = colorToTag(ColorDim)
	TagError = colorToTag(ColorError)
	TagInfo = colorToTag(ColorInfo)
	TagTitle = colorToTag(ColorTitle)
	TagFilter = colorToTag(ColorTitle)
	if invert {
		TagCyan = colorToTag(invertColor(tcell.NewRGBColor(0, 255, 255)))
		TagPink = colorToTag(invertColor(tcell.NewRGBColor(255, 0, 255)))
		TagSCKey = colorToTag(invertColor(tcell.NewRGBColor(32, 144, 255)))
	}
}

const (
	TitleContainers = "Containers"
	TitleImages     = "Images"
	TitleVolumes    = "Volumes"
	TitleNetworks   = "Networks"
	TitleServices   = "Services"
	TitleNodes      = "Nodes"
	TitleCompose    = "Compose"
	TitleAliases    = "Aliases"
	TitleSecrets    = "Secrets"
)

// invertColor inverts a tcell.Color by flipping its lightness while preserving hue and saturation.
func invertColor(c tcell.Color) tcell.Color {
	if c == tcell.Color16 {
		// Special case: dark terminal background -> light
		return tcell.NewRGBColor(240, 240, 240)
	}
	r, g, b := c.RGB()
	col := colorful.Color{R: float64(r) / 255.0, G: float64(g) / 255.0, B: float64(b) / 255.0}
	h, s, l := col.Hsl()
	inverted := colorful.Hsl(h, s, 1.0-l)
	ir, ig, ib := inverted.RGB255()
	return tcell.NewRGBColor(int32(ir), int32(ig), int32(ib))
}

// InvertColors flips all theme colors from dark to light or vice versa.
func InvertColors() {
	ColorBg = invertColor(ColorBg)
	ColorFg = invertColor(ColorFg)
	ColorTableBorder = invertColor(ColorTableBorder)
	ColorBlack = invertColor(ColorBlack)
	ColorWhite = invertColor(ColorWhite)
	ColorHeader = invertColor(ColorHeader)
	ColorHeaderFocus = invertColor(ColorHeaderFocus)
	ColorTitle = invertColor(ColorTitle)
	ColorIdle = invertColor(ColorIdle)
	ColorTeal = invertColor(ColorTeal)
	ColorFooterBg = invertColor(ColorFooterBg)
	ColorFooterFg = invertColor(ColorFooterFg)
	ColorFlashFg = invertColor(ColorFlashFg)
	ColorFlashBg = invertColor(ColorFlashBg)
	ColorSelectBg = invertColor(ColorSelectBg)
	ColorSelectFg = invertColor(ColorSelectFg)
	ColorValue = invertColor(ColorValue)
	ColorSelect = invertColor(ColorSelect)
	ColorDim = invertColor(ColorDim)
	ColorAccent = invertColor(ColorAccent)
	ColorLogo = invertColor(ColorLogo)
	ColorError = invertColor(ColorError)
	ColorInfo = invertColor(ColorInfo)
	ColorStatusGreen = invertColor(ColorStatusGreen)
	ColorStatusRed = invertColor(ColorStatusRed)
	ColorStatusGray = invertColor(ColorStatusGray)
	ColorStatusYellow = invertColor(ColorStatusYellow)
	ColorStatusOrange = invertColor(ColorStatusOrange)
	ColorStatusBlue = invertColor(ColorStatusBlue)
	ColorStatusPurple = invertColor(ColorStatusPurple)
	ColorStatusRedDarkBg = invertColor(ColorStatusRedDarkBg)
	ColorStatusGreenDarkBg = invertColor(ColorStatusGreenDarkBg)
	ColorStatusGrayDarkBg = invertColor(ColorStatusGrayDarkBg)
	ColorStatusYellowDarkBg = invertColor(ColorStatusYellowDarkBg)
	ColorStatusOrangeDarkBg = invertColor(ColorStatusOrangeDarkBg)
	ColorStatusBlueDarkBg = invertColor(ColorStatusBlueDarkBg)
	ColorStatusPurpleDarkBg = invertColor(ColorStatusPurpleDarkBg)

	// Refresh all tag strings to match the inverted colors
	refreshTags(true)
}
