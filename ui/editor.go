package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type editorDoneMsg struct {
	err error
}

// OpenEditorCmd launches $EDITOR (nano/vim fallback) on path via tea.ExecProcess.
func OpenEditorCmd(path string, configuredEditor string) tea.Cmd {
	editor := resolveEditor(configuredEditor)
	return tea.ExecProcess(exec.Command(editor, path), func(err error) tea.Msg {
		return editorDoneMsg{err: err}
	})
}

func resolveEditor(configured string) string {
	if e := strings.TrimSpace(configured); e != "" {
		return e
	}
	if e := strings.TrimSpace(os.Getenv("TAPER_EDITOR")); e != "" {
		return e
	}
	if e := strings.TrimSpace(os.Getenv("EDITOR")); e != "" {
		return e
	}
	if _, err := exec.LookPath("nano"); err == nil {
		return "nano"
	}
	return "vim"
}

func editorName(configured string) string {
	e := resolveEditor(configured)
	if i := strings.LastIndex(e, "/"); i >= 0 {
		return e[i+1:]
	}
	return e
}

func formatEditorErr(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("editor: %v", err)
}
