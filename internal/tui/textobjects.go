package tui

import "strings"

func FindParagraph(lines []string, line int) (start, end int) {
	if line < 0 || line >= len(lines) {
		return 0, 0
	}

	if strings.TrimSpace(lines[line]) == "" {
		return line, line
	}

	start = line
	for start > 0 && strings.TrimSpace(lines[start-1]) != "" {
		start--
	}

	end = line
	for end < len(lines)-1 && strings.TrimSpace(lines[end+1]) != "" {
		end++
	}

	return start, end
}

func FindCodeBlock(lines []string, line int) (start, end int) {
	if line < 0 || line >= len(lines) {
		return -1, -1
	}

	type block struct {
		start, end int
	}
	var blocks []block

	inBlock := false
	blockStart := 0
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "```") {
			if !inBlock {
				inBlock = true
				blockStart = i
			} else {
				blocks = append(blocks, block{start: blockStart, end: i})
				inBlock = false
			}
		}
	}

	for _, b := range blocks {
		if line >= b.start && line <= b.end {
			return b.start, b.end
		}
	}

	return -1, -1
}

func FindSection(lines []string, line int) (start, end int) {
	if line < 0 || line >= len(lines) {
		return 0, len(lines) - 1
	}

	headingLevel := func(l string) int {
		trimmed := strings.TrimLeft(l, " \t")
		count := 0
		for _, r := range trimmed {
			if r == '#' {
				count++
			} else {
				break
			}
		}
		if count > 0 && count <= 6 && len(trimmed) > count && trimmed[count] == ' ' {
			return count
		}
		return 0
	}

	sectionStart := 0
	sectionLevel := 0

	for i := line; i >= 0; i-- {
		lvl := headingLevel(lines[i])
		if lvl > 0 {
			sectionStart = i
			sectionLevel = lvl
			break
		}
		if i == 0 {
			sectionStart = 0
		}
	}

	end = len(lines) - 1
	for i := line + 1; i < len(lines); i++ {
		lvl := headingLevel(lines[i])
		if lvl > 0 {
			if sectionLevel == 0 || lvl <= sectionLevel {
				end = i - 1
				break
			}
		}
	}

	if sectionLevel == 0 {
		for i := line + 1; i < len(lines); i++ {
			lvl := headingLevel(lines[i])
			if lvl > 0 {
				end = i - 1
				break
			}
		}
	}

	return sectionStart, end
}
