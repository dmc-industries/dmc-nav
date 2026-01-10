package main

import (
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Viewer is the interface for file content viewers
type Viewer interface {
	Pane
	CanView(path string) bool
	Load(path string) tea.Cmd
}

// FileLoadedMsg is sent when a file has been loaded
type FileLoadedMsg struct {
	Path    string
	Content string
	Err     error
}

// ViewerRouter selects the appropriate viewer for a file
type ViewerRouter struct {
	viewers []Viewer
	current Viewer
	width   int
	height  int
	focused bool
}

func NewViewerRouter() *ViewerRouter {
	md := NewMarkdownViewer()
	jsonv := NewJSONViewer()
	text := NewTextViewer()
	return &ViewerRouter{
		viewers: []Viewer{md, jsonv, text}, // order matters: specific viewers before fallback
		current: text,
	}
}

func (r *ViewerRouter) Init() tea.Cmd {
	return nil
}

func (r *ViewerRouter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if r.current == nil {
		return r, nil
	}
	m, cmd := r.current.Update(msg)
	r.current = m.(Viewer)
	return r, cmd
}

func (r *ViewerRouter) View() string {
	if r.current == nil {
		return "No viewer"
	}
	return r.current.View()
}

func (r *ViewerRouter) SetSize(width, height int) {
	r.width = width
	r.height = height
	for _, v := range r.viewers {
		v.SetSize(width, height)
	}
}

func (r *ViewerRouter) Focused() bool {
	return r.focused
}

func (r *ViewerRouter) SetFocused(focused bool) {
	r.focused = focused
	if r.current != nil {
		r.current.SetFocused(focused)
	}
}

// OpenFile selects appropriate viewer and loads the file
func (r *ViewerRouter) OpenFile(path string) tea.Cmd {
	// Find first viewer that can handle this file
	for _, v := range r.viewers {
		if v.CanView(path) {
			r.current = v
			r.current.SetSize(r.width, r.height)
			r.current.SetFocused(r.focused)
			return r.current.Load(path)
		}
	}
	// Fallback to first viewer (text)
	r.current = r.viewers[0]
	return r.current.Load(path)
}

// TextViewer displays plain text files
type TextViewer struct {
	width   int
	height  int
	focused bool

	path    string
	content string
	lines   []string
	offset  int
	err     error
}

func NewTextViewer() *TextViewer {
	return &TextViewer{}
}

func (t *TextViewer) Init() tea.Cmd {
	return nil
}

func (t *TextViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FileLoadedMsg:
		if msg.Path == t.path {
			t.content = msg.Content
			t.lines = strings.Split(msg.Content, "\n")
			t.offset = 0
			t.err = msg.Err
		}

	case tea.KeyMsg:
		if !t.focused {
			return t, nil
		}
		switch msg.String() {
		case "j", "down":
			t.scroll(1)
		case "k", "up":
			t.scroll(-1)
		case "d", "ctrl+d":
			t.scroll(t.height / 2)
		case "u", "ctrl+u":
			t.scroll(-t.height / 2)
		case "g":
			t.offset = 0
		case "G":
			t.offset = max(0, len(t.lines)-t.height+2)
		}
	}

	return t, nil
}

func (t *TextViewer) View() string {
	if t.path == "" {
		return t.centerText("Select a file to view")
	}
	if t.err != nil {
		return t.centerText("Error: " + t.err.Error())
	}

	var visible []string
	end := t.offset + t.height - 1
	if end > len(t.lines) {
		end = len(t.lines)
	}

	for i := t.offset; i < end; i++ {
		line := t.lines[i]
		// Truncate long lines
		if len(line) > t.width-2 {
			line = line[:t.width-5] + "..."
		}
		visible = append(visible, line)
	}

	// Header with filename
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render(filepath.Base(t.path))

	lines := append([]string{header}, visible...)

	// Pad to full height to pin content to top
	for len(lines) < t.height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func (t *TextViewer) SetSize(width, height int) {
	t.width = width
	t.height = height
}

func (t *TextViewer) Focused() bool {
	return t.focused
}

func (t *TextViewer) SetFocused(focused bool) {
	t.focused = focused
}

func (t *TextViewer) CanView(path string) bool {
	// TextViewer is the fallback, handles everything
	return true
}

func (t *TextViewer) Load(path string) tea.Cmd {
	t.path = path
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		return FileLoadedMsg{
			Path:    path,
			Content: string(content),
			Err:     err,
		}
	}
}

func (t *TextViewer) scroll(delta int) {
	t.offset += delta
	if t.offset < 0 {
		t.offset = 0
	}
	maxOffset := len(t.lines) - t.height + 2
	if maxOffset < 0 {
		maxOffset = 0
	}
	if t.offset > maxOffset {
		t.offset = maxOffset
	}
}

func (t *TextViewer) centerText(text string) string {
	style := lipgloss.NewStyle().
		Width(t.width).
		Height(t.height).
		Align(lipgloss.Center, lipgloss.Center)
	return style.Render(text)
}
