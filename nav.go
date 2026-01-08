package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FileEntry represents a file or directory in the tree
type FileEntry struct {
	Name     string
	Path     string
	IsDir    bool
	Expanded bool
	Depth    int
}

// NavPane is the file tree navigation component
type NavPane struct {
	width   int
	height  int
	focused bool

	root    string        // root directory path
	entries []FileEntry   // flattened visible entries
	cursor  int           // current selection index
	offset  int           // scroll offset for viewport
}

func NewNavPane(root string) *NavPane {
	n := &NavPane{
		root:   root,
		cursor: 0,
	}
	n.loadEntries()
	return n
}

func (n *NavPane) Init() tea.Cmd {
	return nil
}

func (n *NavPane) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !n.focused {
		return n, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			n.moveCursor(1)
		case "k", "up":
			n.moveCursor(-1)
		case "g":
			n.cursor = 0
			n.offset = 0
		case "G":
			n.cursor = len(n.entries) - 1
			n.adjustOffset()
		case "enter", "l":
			n.toggleOrOpen()
		case "h", "backspace":
			n.goToParent()
		}
	}

	return n, nil
}

func (n *NavPane) View() string {
	if len(n.entries) == 0 {
		return "Empty directory"
	}

	var lines []string
	visibleHeight := n.height - 2 // leave room for header/footer

	// Header showing current directory
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render(filepath.Base(n.root))
	lines = append(lines, header, "")

	// File entries
	end := n.offset + visibleHeight
	if end > len(n.entries) {
		end = len(n.entries)
	}

	for i := n.offset; i < end; i++ {
		entry := n.entries[i]
		line := n.renderEntry(entry, i == n.cursor)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (n *NavPane) SetSize(width, height int) {
	n.width = width
	n.height = height
}

func (n *NavPane) Focused() bool {
	return n.focused
}

func (n *NavPane) SetFocused(focused bool) {
	n.focused = focused
}

// SelectedPath returns the path of the currently selected entry
func (n *NavPane) SelectedPath() string {
	if n.cursor >= 0 && n.cursor < len(n.entries) {
		return n.entries[n.cursor].Path
	}
	return ""
}

func (n *NavPane) loadEntries() {
	n.entries = nil
	n.loadDir(n.root, 0)
}

func (n *NavPane) loadDir(dir string, depth int) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	// Sort: directories first, then alphabetically
	sort.Slice(files, func(i, j int) bool {
		iDir := files[i].IsDir()
		jDir := files[j].IsDir()
		if iDir != jDir {
			return iDir
		}
		return strings.ToLower(files[i].Name()) < strings.ToLower(files[j].Name())
	})

	for _, f := range files {
		name := f.Name()
		// Skip hidden files (except .git for now)
		if strings.HasPrefix(name, ".") && name != ".git" {
			continue
		}

		path := filepath.Join(dir, name)
		entry := FileEntry{
			Name:  name,
			Path:  path,
			IsDir: f.IsDir(),
			Depth: depth,
		}
		n.entries = append(n.entries, entry)

		// If directory is expanded, load its contents
		if entry.IsDir && entry.Expanded {
			n.loadDir(path, depth+1)
		}
	}
}

func (n *NavPane) renderEntry(entry FileEntry, selected bool) string {
	indent := strings.Repeat("  ", entry.Depth)

	style := lipgloss.NewStyle()
	if selected {
		style = style.
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Bold(true)
	} else if entry.IsDir {
		style = style.Foreground(lipgloss.Color("12"))
	}

	name := entry.Name
	if entry.IsDir {
		name += "/"
	}

	line := indent + name

	// Pad to width for selection highlight
	if selected && n.width > 0 {
		padding := n.width - lipgloss.Width(line)
		if padding > 0 {
			line += strings.Repeat(" ", padding)
		}
	}

	return style.Render(line)
}

func (n *NavPane) moveCursor(delta int) {
	n.cursor += delta
	if n.cursor < 0 {
		n.cursor = 0
	}
	if n.cursor >= len(n.entries) {
		n.cursor = len(n.entries) - 1
	}
	n.adjustOffset()
}

func (n *NavPane) adjustOffset() {
	visibleHeight := n.height - 2
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	// Scroll up if cursor above viewport
	if n.cursor < n.offset {
		n.offset = n.cursor
	}
	// Scroll down if cursor below viewport
	if n.cursor >= n.offset+visibleHeight {
		n.offset = n.cursor - visibleHeight + 1
	}
}

func (n *NavPane) toggleOrOpen() {
	if n.cursor < 0 || n.cursor >= len(n.entries) {
		return
	}

	entry := &n.entries[n.cursor]
	if entry.IsDir {
		entry.Expanded = !entry.Expanded
		n.loadEntries()
		// Try to keep cursor on same entry after reload
		for i, e := range n.entries {
			if e.Path == entry.Path {
				n.cursor = i
				break
			}
		}
	}
	// TODO: For files, emit message to open in viewer
}

func (n *NavPane) goToParent() {
	parent := filepath.Dir(n.root)
	if parent != n.root {
		oldRoot := n.root
		n.root = parent
		n.loadEntries()
		// Try to select the directory we came from
		for i, e := range n.entries {
			if e.Path == oldRoot {
				n.cursor = i
				n.adjustOffset()
				break
			}
		}
	}
}
