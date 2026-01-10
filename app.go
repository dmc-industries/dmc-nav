package main

import (
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Focus indicates which pane has keyboard focus
type Focus int

const (
	FocusNav Focus = iota
	FocusViewer
)

// Mode indicates the current application mode
type Mode int

const (
	ModeNav Mode = iota
	ModeViewer
	ModeEditor
)

// Pane is the interface that nav and viewer components implement
type Pane interface {
	tea.Model
	SetSize(width, height int)
	Focused() bool
	SetFocused(focused bool)
}

// App is the main application model that orchestrates panes
type App struct {
	width  int
	height int
	ready  bool

	focus Focus
	mode  Mode

	nav        Pane
	viewer     *ViewerRouter
	editor     *Editor
	editPath   string // path being edited
}

// NavPane ratio (left side width percentage)
const navPaneRatio = 0.25

func NewApp() *App {
	cwd, _ := os.Getwd()
	cwd, _ = filepath.EvalSymlinks(cwd) // resolve to real path

	nav := NewNavPane("/")
	nav.ExpandToPath(cwd)
	nav.PinTop() // keep root visible
	nav.SetFocused(true)

	return &App{
		focus:  FocusNav,
		mode:   ModeNav,
		nav:    nav,
		viewer: NewViewerRouter(),
		editor: NewEditor(),
	}
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// In editor mode, only editor handles keys (except ctrl+c for emergency exit)
		if a.mode == ModeEditor {
			if msg.String() == "ctrl+c" {
				return a, tea.Quit
			}
			var m tea.Model
			var cmd tea.Cmd
			m, cmd = a.editor.Update(msg)
			a.editor = m.(*Editor)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return a, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit

		case "tab":
			a.cycleFocus()
			return a, nil

		case "e":
			// Open editor for current file (if viewing a text file)
			if a.focus == FocusViewer && a.editPath != "" && isTextFile(a.editPath) {
				a.mode = ModeEditor
				a.editor.SetSize(a.width-int(float64(a.width)*navPaneRatio)-1, a.height)
				a.editor.SetFocused(true)
				a.viewer.SetFocused(false)
				cmd := a.editor.Open(a.editPath)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				return a, tea.Batch(cmds...)
			}
		}

		// Forward key events to focused pane
		cmd := a.updateFocusedPane(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.ready = true
		a.updatePaneSizes()

	case FileSelectedMsg:
		// Track path for potential editing
		a.editPath = msg.Path
		// Open file in viewer
		cmd := a.viewer.OpenFile(msg.Path)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case FileLoadedMsg:
		// Forward to viewer
		_, cmd := a.viewer.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case MarkdownLoadedMsg:
		// Forward to viewer
		_, cmd := a.viewer.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case JSONLoadedMsg:
		// Forward to viewer
		_, cmd := a.viewer.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case EditorOpenMsg:
		// Forward to editor
		_, cmd := a.editor.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case EditorSavedMsg:
		// Return to viewer mode and refresh
		a.mode = ModeViewer
		a.editor.SetFocused(false)
		a.viewer.SetFocused(true)
		a.focus = FocusViewer
		// Reload file in viewer to show changes
		if msg.Err == nil {
			cmd := a.viewer.OpenFile(msg.Path)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case EditorCancelledMsg:
		// Return to viewer mode without saving
		a.mode = ModeViewer
		a.editor.SetFocused(false)
		a.viewer.SetFocused(true)
		a.focus = FocusViewer
	}

	return a, tea.Batch(cmds...)
}

func (a *App) View() string {
	if !a.ready {
		return "Initializing..."
	}

	navWidth := int(float64(a.width) * navPaneRatio)
	rightWidth := a.width - navWidth - 1 // -1 for border

	navStyle := lipgloss.NewStyle().
		Width(navWidth).
		Height(a.height).
		AlignVertical(lipgloss.Top).
		BorderStyle(lipgloss.NormalBorder()).
		BorderRight(true)

	rightStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(a.height).
		AlignVertical(lipgloss.Top)

	if a.focus == FocusNav {
		navStyle = navStyle.BorderForeground(lipgloss.Color("62"))
	} else {
		rightStyle = rightStyle.BorderForeground(lipgloss.Color("62"))
	}

	// Show editor or viewer depending on mode
	var rightPane string
	if a.mode == ModeEditor {
		rightPane = a.editor.View()
	} else {
		rightPane = a.viewer.View()
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		navStyle.Render(a.nav.View()),
		rightStyle.Render(rightPane),
	)
}

func (a *App) cycleFocus() {
	if a.focus == FocusNav {
		a.focus = FocusViewer
	} else {
		a.focus = FocusNav
	}
	a.nav.SetFocused(a.focus == FocusNav)
	a.viewer.SetFocused(a.focus == FocusViewer)
}

func (a *App) updateFocusedPane(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	if a.focus == FocusNav {
		var m tea.Model
		m, cmd = a.nav.Update(msg)
		a.nav = m.(Pane)
	} else {
		var m tea.Model
		m, cmd = a.viewer.Update(msg)
		a.viewer = m.(*ViewerRouter)
	}
	return cmd
}

func (a *App) updatePaneSizes() {
	navWidth := int(float64(a.width) * navPaneRatio)
	rightWidth := a.width - navWidth - 1

	a.nav.SetSize(navWidth, a.height)
	a.viewer.SetSize(rightWidth, a.height)
	a.editor.SetSize(rightWidth, a.height)
}
