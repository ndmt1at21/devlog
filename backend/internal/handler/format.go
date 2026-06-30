package handler

import (
	"fmt"
	"strings"
	"time"
)

// formatVNDate renders a date as "24 Th6, 2026", matching the design.
func formatVNDate(t time.Time) string {
	return fmt.Sprintf("%02d Th%d, %d", t.Day(), int(t.Month()), t.Year())
}

// relativeVN renders a Vietnamese relative time label like the design's comments.
func relativeVN(t time.Time) string {
	diff := time.Since(t)
	switch {
	case diff < time.Minute:
		return "Vừa xong"
	case diff < time.Hour:
		return fmt.Sprintf("%d phút trước", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%d giờ trước", int(diff.Hours()))
	case diff < 30*24*time.Hour:
		return fmt.Sprintf("%d ngày trước", int(diff.Hours()/24))
	default:
		return formatVNDate(t)
	}
}

// initial returns the uppercased first rune of a name, or "?".
func initial(name string) string {
	r := []rune(strings.TrimSpace(name))
	if len(r) == 0 {
		return "?"
	}
	return strings.ToUpper(string(r[0]))
}
