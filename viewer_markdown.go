package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MarkdownViewer renders markdown files with glamour
type MarkdownViewer struct {
	width   int
	height  int
	focused bool

	path     string
	rendered string
	lines    []string
	offset   int
	err      error
}

func NewMarkdownViewer() *MarkdownViewer {
	return &MarkdownViewer{}
}

func (m *MarkdownViewer) Init() tea.Cmd {
	return nil
}

func (m *MarkdownViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case MarkdownLoadedMsg:
		if msg.Path == m.path {
			m.rendered = msg.Content
			m.lines = strings.Split(msg.Content, "\n")
			m.offset = 0
			m.err = msg.Err
		}

	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}
		switch msg.String() {
		case "j", "down":
			m.scroll(1)
		case "k", "up":
			m.scroll(-1)
		case "d", "ctrl+d":
			m.scroll(m.height / 2)
		case "u", "ctrl+u":
			m.scroll(-m.height / 2)
		case "g":
			m.offset = 0
		case "G":
			m.offset = max(0, len(m.lines)-m.height+2)
		}
	}

	return m, nil
}

func (m *MarkdownViewer) View() string {
	if m.path == "" {
		return m.centerText("Select a markdown file to view")
	}
	if m.err != nil {
		return m.centerText("Error: " + m.err.Error())
	}

	var visible []string
	end := m.offset + m.height - 1
	if end > len(m.lines) {
		end = len(m.lines)
	}

	for i := m.offset; i < end; i++ {
		visible = append(visible, m.lines[i])
	}

	// Header with filename
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render(filepath.Base(m.path))

	lines := append([]string{header}, visible...)

	// Pad to full height to pin content to top
	for len(lines) < m.height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func (m *MarkdownViewer) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *MarkdownViewer) Focused() bool {
	return m.focused
}

func (m *MarkdownViewer) SetFocused(focused bool) {
	m.focused = focused
}

func (m *MarkdownViewer) CanView(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".markdown"
}

func (m *MarkdownViewer) Load(path string) tea.Cmd {
	m.path = path
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		if err != nil {
			return MarkdownLoadedMsg{Path: path, Err: err}
		}

		// Render markdown with glamour
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(80),
		)
		if err != nil {
			return MarkdownLoadedMsg{Path: path, Err: err}
		}

		rendered, err := renderer.Render(string(content))
		if err != nil {
			return MarkdownLoadedMsg{Path: path, Err: err}
		}

		return MarkdownLoadedMsg{Path: path, Content: rendered}
	}
}

func (m *MarkdownViewer) scroll(delta int) {
	m.offset += delta
	if m.offset < 0 {
		m.offset = 0
	}
	maxOffset := len(m.lines) - m.height + 2
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
}

func (m *MarkdownViewer) centerText(text string) string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)
	return style.Render(text)
}

// MarkdownLoadedMsg is sent when markdown has been rendered
type MarkdownLoadedMsg struct {
	Path    string
	Content string
	Err     error
}
