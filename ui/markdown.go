package ui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
)

var (
	glamourRenderer *glamour.TermRenderer
	glamourWidth    int
	glamourMu       sync.Mutex
)

func renderMarkdown(source string, width int) string {
	if width < 20 {
		width = 20
	}
	source = strings.TrimSpace(source)
	if source == "" {
		return mutedStyle.Render("(empty — press e to edit)")
	}

	glamourMu.Lock()
	defer glamourMu.Unlock()

	if glamourRenderer == nil || glamourWidth != width {
		r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width),
		)
		if err != nil {
			return fmt.Sprintf("render error: %v\n\n%s", err, source)
		}
		glamourRenderer = r
		glamourWidth = width
	}

	out, err := glamourRenderer.Render(source)
	if err != nil {
		return fmt.Sprintf("render error: %v\n\n%s", err, source)
	}
	return strings.TrimRight(out, "\n")
}

func invalidateMarkdownRenderer() {
	glamourMu.Lock()
	defer glamourMu.Unlock()
	glamourRenderer = nil
	glamourWidth = 0
}
