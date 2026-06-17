package storage

import (
	"strings"
)

// CheckItem is a checkbox line parsed from packing.md.
type CheckItem struct {
	LineIndex int
	Text      string
	Checked   bool
	RawLine   string
}

// ParseChecklist extracts `- [ ]` and `- [x]` items from markdown.
func ParseChecklist(content string) []CheckItem {
	lines := strings.Split(content, "\n")
	var items []CheckItem
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- [") {
			continue
		}
		closeIdx := strings.Index(trimmed, "]")
		if closeIdx < 4 {
			continue
		}
		mark := trimmed[3]
		if mark != ' ' && mark != 'x' && mark != 'X' {
			continue
		}
		text := strings.TrimSpace(trimmed[closeIdx+1:])
		items = append(items, CheckItem{
			LineIndex: i,
			Text:      text,
			Checked:   mark == 'x' || mark == 'X',
			RawLine:   line,
		})
	}
	return items
}

// ToggleChecklistItem flips the checkbox at item index and writes the file.
func ToggleChecklistItem(raceID string, itemIndex int) ([]CheckItem, error) {
	content, err := ReadFile(raceID, "packing")
	if err != nil {
		return nil, err
	}
	items := ParseChecklist(content)
	if itemIndex < 0 || itemIndex >= len(items) {
		return items, nil
	}
	lines := strings.Split(content, "\n")
	lines[items[itemIndex].LineIndex] = toggleCheckboxLine(lines[items[itemIndex].LineIndex])
	newContent := strings.Join(lines, "\n")
	if err := WriteFile(raceID, "packing", newContent); err != nil {
		return nil, err
	}
	return ParseChecklist(newContent), nil
}

func toggleCheckboxLine(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	indent := line[:len(line)-len(trimmed)]
	if !strings.HasPrefix(trimmed, "- [") {
		return line
	}
	closeIdx := strings.Index(trimmed, "]")
	if closeIdx < 0 {
		return line
	}
	text := strings.TrimSpace(trimmed[closeIdx+1:])
	mark := trimmed[3]
	checked := mark == 'x' || mark == 'X'
	if checked {
		return indent + "- [ ] " + text
	}
	return indent + "- [x] " + text
}
