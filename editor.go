package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Editor is a simple text editor using textarea
type Editor struct {
	width   int
	height  int
	focused bool

	path     string
	textarea textarea.Model
	modified bool
	err      error
	status   string
}

func NewEditor() *Editor {
	ta := textarea.New()
	ta.ShowLineNumbers = true
	ta.CharLimit = 0 // unlimited
	return &Editor{
		textarea: ta,
	}
}

func (e *Editor) Init() tea.Cmd {
	return nil
}

func (e *Editor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case EditorOpenMsg:
		e.path = msg.Path
		e.textarea.SetValue(msg.Content)
		e.textarea.Focus()
		e.modified = false
		e.err = msg.Err
		e.status = ""
		e.updateSize()
		return e, nil

	case tea.KeyMsg:
		if !e.focused {
			return e, nil
		}

		key := msg.String()

		// Handle commands
		switch key {
		case "ctrl+s":
			return e, e.save()
		case "esc":
			return e, e.cancel()
		}

		// Check for :w command (vim-style save)
		if key == ":" {
			// We'd need a command mode for proper :w support
			// For now, just pass through
		}

		// Update textarea
		var cmd tea.Cmd
		e.textarea, cmd = e.textarea.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		e.modified = true

	default:
		var cmd tea.Cmd
		e.textarea, cmd = e.textarea.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return e, tea.Batch(cmds...)
}

func (e *Editor) View() string {
	if e.path == "" {
		return e.centerText("No file open")
	}
	if e.err != nil {
		return e.centerText("Error: " + e.err.Error())
	}

	// Header with filename and modified indicator
	name := filepath.Base(e.path)
	if e.modified {
		name += " [+]"
	}
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render(name)

	// Status bar
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))
	status := statusStyle.Render("Ctrl+S: save | Esc: cancel")
	if e.status != "" {
		status = statusStyle.Render(e.status)
	}

	return header + "\n" + e.textarea.View() + "\n" + status
}

func (e *Editor) SetSize(width, height int) {
	e.width = width
	e.height = height
	e.updateSize()
}

func (e *Editor) updateSize() {
	// Account for header and status line
	e.textarea.SetWidth(e.width)
	e.textarea.SetHeight(e.height - 3) // -1 header, -1 status, -1 padding
}

func (e *Editor) Focused() bool {
	return e.focused
}

func (e *Editor) SetFocused(focused bool) {
	e.focused = focused
	if focused {
		e.textarea.Focus()
	} else {
		e.textarea.Blur()
	}
}

func (e *Editor) save() tea.Cmd {
	return func() tea.Msg {
		content := e.textarea.Value()
		err := os.WriteFile(e.path, []byte(content), 0644)
		return EditorSavedMsg{Path: e.path, Err: err}
	}
}

func (e *Editor) cancel() tea.Cmd {
	return func() tea.Msg {
		return EditorCancelledMsg{Path: e.path}
	}
}

func (e *Editor) centerText(text string) string {
	style := lipgloss.NewStyle().
		Width(e.width).
		Height(e.height).
		Align(lipgloss.Center, lipgloss.Center)
	return style.Render(text)
}

// Open prepares the editor to edit a file
func (e *Editor) Open(path string) tea.Cmd {
	e.path = path
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		return EditorOpenMsg{
			Path:    path,
			Content: string(content),
			Err:     err,
		}
	}
}

// EditorOpenMsg is sent when a file is ready for editing
type EditorOpenMsg struct {
	Path    string
	Content string
	Err     error
}

// EditorSavedMsg is sent when a file has been saved
type EditorSavedMsg struct {
	Path string
	Err  error
}

// EditorCancelledMsg is sent when editing is cancelled
type EditorCancelledMsg struct {
	Path string
}

// Helper to check if file is likely text (for edit routing)
func isTextFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	textExts := map[string]bool{
		".txt": true, ".md": true, ".markdown": true,
		".go": true, ".py": true, ".js": true, ".ts": true,
		".json": true, ".yaml": true, ".yml": true, ".toml": true,
		".html": true, ".css": true, ".xml": true,
		".sh": true, ".bash": true, ".zsh": true,
		".c": true, ".h": true, ".cpp": true, ".hpp": true,
		".rs": true, ".rb": true, ".java": true,
		".sql": true, ".graphql": true,
		".conf": true, ".cfg": true, ".ini": true,
		".gitignore": true, ".dockerignore": true,
		"": true, // no extension - might be text
	}
	return textExts[ext]
}
