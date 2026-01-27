package tui

import "strings"

func decodeAnnText(s string) string {
	return strings.ReplaceAll(s, "\\n", "\n")
}

func encodeAnnText(s string) string {
	return strings.ReplaceAll(s, "\n", "\\n")
}

func wrapLine(s string, width int) []string {
	if width <= 0 || len(s) <= width {
		return []string{s}
	}

	var result []string
	runes := []rune(s)
	for len(runes) > width {
		result = append(result, string(runes[:width]))
		runes = runes[width:]
	}
	if len(runes) > 0 {
		result = append(result, string(runes))
	}
	return result
}
