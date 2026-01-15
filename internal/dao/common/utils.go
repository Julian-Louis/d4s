package common

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Resource is a generic interface for displayable items
// Already defined in common.go but Go allows multiple files per package.
// If common.go defines Resource, I don't need to redefine it here if they are in same package.
// BUT, if I change package name to common, I must ensure common.go is also package common.

func ShortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err == nil && home != "" && strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

func ParseStatus(s string) (status, age string) {
	s = strings.TrimSpace(s)
	
	if strings.HasPrefix(s, "Up") {
		status = "Up"
		rest := strings.TrimPrefix(s, "Up ")
		age = ShortenDuration(rest)
	} else if strings.HasPrefix(s, "Exited") {
		// "Exited (0) 5 minutes ago"
		parts := strings.SplitN(s, ") ", 2)
		if len(parts) == 2 {
			status = parts[0] + ")"
			age = ShortenDuration(strings.TrimSuffix(parts[1], " ago"))
		} else {
			status = "Exited"
			age = s
		}
	} else if strings.HasPrefix(s, "Created") {
		status = "Created"
		age = "-"
	} else if strings.HasPrefix(s, "Paused") {
		// "Up 2 hours (Paused)"
		if strings.Contains(s, "(Paused)") {
			status = "Paused"
			rest := strings.TrimPrefix(s, "Up ")
			rest = strings.TrimSuffix(rest, " (Paused)")
			age = ShortenDuration(rest)
			return
		}
		status = "Paused"
		age = "-"
	} else if strings.HasPrefix(s, "Exiting") { // Explicitly handle Exiting
		status = "Exiting"
		age = "-"
	} else if strings.Contains(strings.ToLower(s), "starting") {
		status = "Starting"
		age = "-"
	} else {
		status = s
		age = "-"
	}
	return
}

func ShortenDuration(d string) string {
	d = strings.ToLower(d)
	if strings.Contains(d, "less than") {
		return "1s"
	}
	
	// Clean up verbose words
	d = strings.ReplaceAll(d, "about ", "")
	d = strings.ReplaceAll(d, "an ", "1 ")
	d = strings.ReplaceAll(d, "a ", "1 ")
	d = strings.TrimSuffix(d, " ago")
	
	parts := strings.Fields(d)
	if len(parts) >= 2 {
		val := parts[0]
		unit := parts[1]
		
		if val == "0" && strings.HasPrefix(unit, "second") {
			return "1s"
		}

		if strings.HasPrefix(unit, "second") { return val + "s" }
		if strings.HasPrefix(unit, "minute") { return val + "m" }
		if strings.HasPrefix(unit, "hour") { return val + "h" }
		if strings.HasPrefix(unit, "day") { return val + "d" }
		if strings.HasPrefix(unit, "week") { return val + "w" }
		if strings.HasPrefix(unit, "month") { return val + "mo" }
		if strings.HasPrefix(unit, "year") { return val + "y" }
	}
	return d
}

func FormatTime(ts int64) string {
	t := time.Unix(ts, 0)
	return t.Format("2006-01-02 15:04")
}

func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

