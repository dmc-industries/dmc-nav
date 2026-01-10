package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// JSONNode represents a node in the JSON tree
type JSONNode struct {
	Key      string
	Value    any
	Children []*JSONNode
	Expanded bool
	Depth    int
	IsArray  bool
}

// JSONViewer displays JSON files as a collapsible tree
type JSONViewer struct {
	width   int
	height  int
	focused bool

	path   string
	root   *JSONNode
	cursor int
	offset int
	err    error
}

func NewJSONViewer() *JSONViewer {
	return &JSONViewer{}
}

func (j *JSONViewer) Init() tea.Cmd {
	return nil
}

func (j *JSONViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case JSONLoadedMsg:
		if msg.Path == j.path {
			j.root = msg.Root
			j.cursor = 0
			j.offset = 0
			j.err = msg.Err
		}

	case tea.KeyMsg:
		if !j.focused {
			return j, nil
		}
		visible := j.visibleNodes()
		switch msg.String() {
		case "j", "down":
			if j.cursor < len(visible)-1 {
				j.cursor++
				j.ensureVisible()
			}
		case "k", "up":
			if j.cursor > 0 {
				j.cursor--
				j.ensureVisible()
			}
		case "enter", "l", "right":
			if j.cursor < len(visible) {
				node := visible[j.cursor]
				if len(node.Children) > 0 {
					node.Expanded = !node.Expanded
				}
			}
		case "h", "left":
			// Collapse current node or go to parent
			if j.cursor < len(visible) {
				node := visible[j.cursor]
				if node.Expanded && len(node.Children) > 0 {
					node.Expanded = false
				}
			}
		case "d", "ctrl+d":
			j.cursor += j.height / 2
			if j.cursor >= len(visible) {
				j.cursor = len(visible) - 1
			}
			j.ensureVisible()
		case "u", "ctrl+u":
			j.cursor -= j.height / 2
			if j.cursor < 0 {
				j.cursor = 0
			}
			j.ensureVisible()
		case "g":
			j.cursor = 0
			j.offset = 0
		case "G":
			j.cursor = len(visible) - 1
			j.ensureVisible()
		}
	}

	return j, nil
}

func (j *JSONViewer) View() string {
	if j.path == "" {
		return j.centerText("Select a JSON file to view")
	}
	if j.err != nil {
		return j.centerText("Error: " + j.err.Error())
	}
	if j.root == nil {
		return j.centerText("Loading...")
	}

	visible := j.visibleNodes()
	viewHeight := j.height - 2 // -1 for header, -1 for padding

	// Header with filename
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render(filepath.Base(j.path))

	var lines []string
	lines = append(lines, header)

	end := j.offset + viewHeight
	if end > len(visible) {
		end = len(visible)
	}

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("81"))
	stringStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("178"))
	boolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("168"))
	nullStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	cursorStyle := lipgloss.NewStyle().Background(lipgloss.Color("237"))

	for i := j.offset; i < end; i++ {
		node := visible[i]
		indent := strings.Repeat("  ", node.Depth)

		var line string
		prefix := " "
		if len(node.Children) > 0 {
			if node.Expanded {
				prefix = "▼"
			} else {
				prefix = "▶"
			}
		}

		keyPart := ""
		if node.Key != "" {
			keyPart = keyStyle.Render(fmt.Sprintf("%q", node.Key)) + ": "
		}

		valuePart := j.renderValue(node, stringStyle, numberStyle, boolStyle, nullStyle)

		line = fmt.Sprintf("%s%s %s%s", indent, prefix, keyPart, valuePart)

		// Truncate long lines
		if len(line) > j.width-2 {
			line = line[:j.width-5] + "..."
		}

		if i == j.cursor {
			line = cursorStyle.Render(line)
		}

		lines = append(lines, line)
	}

	// Pad to full height
	for len(lines) < j.height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func (j *JSONViewer) renderValue(node *JSONNode, stringStyle, numberStyle, boolStyle, nullStyle lipgloss.Style) string {
	if len(node.Children) > 0 {
		count := len(node.Children)
		if node.IsArray {
			if node.Expanded {
				return fmt.Sprintf("[%d items]", count)
			}
			return fmt.Sprintf("[%d items...]", count)
		}
		if node.Expanded {
			return fmt.Sprintf("{%d keys}", count)
		}
		return fmt.Sprintf("{%d keys...}", count)
	}

	switch v := node.Value.(type) {
	case string:
		s := fmt.Sprintf("%q", v)
		if len(s) > 50 {
			s = s[:47] + "...\""
		}
		return stringStyle.Render(s)
	case float64:
		if v == float64(int(v)) {
			return numberStyle.Render(fmt.Sprintf("%d", int(v)))
		}
		return numberStyle.Render(fmt.Sprintf("%g", v))
	case bool:
		return boolStyle.Render(fmt.Sprintf("%t", v))
	case nil:
		return nullStyle.Render("null")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (j *JSONViewer) visibleNodes() []*JSONNode {
	if j.root == nil {
		return nil
	}
	var nodes []*JSONNode
	j.collectVisible(j.root, &nodes)
	return nodes
}

func (j *JSONViewer) collectVisible(node *JSONNode, nodes *[]*JSONNode) {
	*nodes = append(*nodes, node)
	if node.Expanded {
		for _, child := range node.Children {
			j.collectVisible(child, nodes)
		}
	}
}

func (j *JSONViewer) ensureVisible() {
	viewHeight := j.height - 2
	if j.cursor < j.offset {
		j.offset = j.cursor
	}
	if j.cursor >= j.offset+viewHeight {
		j.offset = j.cursor - viewHeight + 1
	}
}

func (j *JSONViewer) SetSize(width, height int) {
	j.width = width
	j.height = height
}

func (j *JSONViewer) Focused() bool {
	return j.focused
}

func (j *JSONViewer) SetFocused(focused bool) {
	j.focused = focused
}

func (j *JSONViewer) CanView(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".json"
}

func (j *JSONViewer) Load(path string) tea.Cmd {
	j.path = path
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		if err != nil {
			return JSONLoadedMsg{Path: path, Err: err}
		}

		var data any
		if err := json.Unmarshal(content, &data); err != nil {
			return JSONLoadedMsg{Path: path, Err: err}
		}

		root := buildTree("", data, 0)
		// Auto-expand root level
		root.Expanded = true

		return JSONLoadedMsg{Path: path, Root: root}
	}
}

func buildTree(key string, value any, depth int) *JSONNode {
	node := &JSONNode{
		Key:   key,
		Value: value,
		Depth: depth,
	}

	switch v := value.(type) {
	case map[string]any:
		node.Expanded = false
		for k, val := range v {
			child := buildTree(k, val, depth+1)
			node.Children = append(node.Children, child)
		}
	case []any:
		node.IsArray = true
		node.Expanded = false
		for i, val := range v {
			child := buildTree(fmt.Sprintf("[%d]", i), val, depth+1)
			node.Children = append(node.Children, child)
		}
	}

	return node
}

func (j *JSONViewer) centerText(text string) string {
	style := lipgloss.NewStyle().
		Width(j.width).
		Height(j.height).
		Align(lipgloss.Center, lipgloss.Center)
	return style.Render(text)
}

// JSONLoadedMsg is sent when JSON has been parsed
type JSONLoadedMsg struct {
	Path string
	Root *JSONNode
	Err  error
}
