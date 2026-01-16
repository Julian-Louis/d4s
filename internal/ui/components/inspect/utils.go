package inspect

import (
	"fmt"
	"strings"
)

// FormatInspectorTitle generates the standard title string for inspectors
// Format: Action(subject) [Mode] <Search>
func FormatInspectorTitle(action, subject, mode, filter string, matchIndex, matchCount int) string {
	// Special handling for @ separator in subject to make it white
	if strings.Contains(subject, "@") {
		subject = strings.ReplaceAll(subject, "@", "[white] @ [#ff00ff]")
	}
	
	title := fmt.Sprintf("[#00ffff::b]%s([#ff00ff]%s[#00ffff])", action, subject)
	modeStr := fmt.Sprintf(" [#00ffff::b][[white]%s[#00ffff]]", mode)
	
	search := ""
	if filter != "" {
		idx := 0
		if matchCount > 0 {
			idx = matchIndex + 1
		}
		
		search = fmt.Sprintf(" [#00ffff::b]</[#ff00ff]%s [-][[white]%d[-]:[white]%d[-]][#00ffff]>", filter, idx, matchCount)
	}

	return fmt.Sprintf(" %s%s%s ", title, modeStr, search)
}
