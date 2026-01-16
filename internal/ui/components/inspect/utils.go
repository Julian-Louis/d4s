package inspect

import (
	"fmt"
	"strings"
)

// FormatInspectorTitle generates the standard title string for inspectors
// Format: Action(subject) [Mode] <Search>
// Colors:
// - Action, brackets, parenthesis: Blue
// - Subject, Search text: Orange
// - Mode, Counters: White
func FormatInspectorTitle(action, subject, mode, filter string, matchIndex, matchCount int) string {
	// Special handling for @ separator in subject to make it white
	if strings.Contains(subject, "@") {
		subject = strings.ReplaceAll(subject, "@", "[white] @ [orange]")
	}
	
	title := fmt.Sprintf("[#8be9fd]%s([orange]%s[#8be9fd])", action, subject)
	modeStr := fmt.Sprintf(" [#8be9fd][[white]%s[#8be9fd]]", mode)
	
	search := ""
	if filter != "" {
		idx := 0
		if matchCount > 0 {
			idx = matchIndex + 1
		}
		
		search = fmt.Sprintf(" [#8be9fd]</[orange]%s [-][[white]%d[-]:[white]%d[-]][#8be9fd]>", filter, idx, matchCount)
	}

	return fmt.Sprintf(" %s%s%s ", title, modeStr, search)
}
