package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func truncateText(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen > 3 {
		return s[:maxLen-3] + "..."
	}
	return s[:maxLen]
}

func colorTags(tags []string) string {
	var b strings.Builder
	for i, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(tagColors[i%len(tagColors)]).
			Padding(0, 1).
			MarginRight(1)
		b.WriteString(style.Render(t))
	}
	return b.String()
}
