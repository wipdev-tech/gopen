// Package fzf contains types and logic for the interactive fuzzy finder part
// of Gopen.
package fzf

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	l "github.com/charmbracelet/lipgloss"
	"github.com/wipdev-tech/gopen/internal/config"
)

var styles = struct {
	selected l.Style
	rest     l.Style
	cursor   l.Style
	window   l.Style
}{
	rest:   l.NewStyle().Faint(true),
	cursor: l.NewStyle().Blink(true),
	window: l.NewStyle().PaddingLeft(1).PaddingRight(1).Border(l.RoundedBorder()),
	selected: l.NewStyle().
		Foreground(l.Color("255")).
		Background(l.Color("56")),
}

// Model implements the tea.Model interface to be used as the model part of the
// bubbletea program and includes fields that hold the program state.
//
// Note that the fields `Config` and `Selected` are exported because the are
// used by the main package.
type Model struct {
	Config      config.C
	Selected    string
	searchStr   string
	selectedIdx int
	helpShown   bool
	done        bool
}

// Init is one of the tea.Model interface methods but not used by the fuzzy
// finder.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update is one of the tea.Model interface methods. It triggers updates to the
// model and its state on keypresses.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.done = true
			return m, tea.Quit

		case "ctrl+w":
			m.searchStr = ""

		case "up", "ctrl+p":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}

		case "down", "ctrl+n":
			if m.selectedIdx < 9 && m.selectedIdx < len(m.Config.DirAliases)-1 {
				m.selectedIdx++
			}

		case "enter":
			m.done = true
			return m, tea.Quit

		case "backspace":
			if len(m.searchStr) >= 1 {
				m.searchStr = m.searchStr[:len(m.searchStr)-1]
			}

		case "?":
			m.helpShown = !m.helpShown

		default:
			if len(msg.String()) == 1 {
				m.searchStr += msg.String()
				m.Selected = m.searchStr
			}
		}
	}

	return m, nil
}

// View is one of the tea.Model interface methods. It includes the rendering logic.
func (m Model) View() string {
	s := fmt.Sprintf("Which project do you want to open?\n> %s", m.searchStr)
	if !m.done {
		s += styles.cursor.Render("█")
	}
	s += "\n\n"

	maxLenAlias := 0
	maxLenPath := 0
	for _, a := range m.Config.DirAliases {
		if len(a.Alias) > maxLenAlias {
			maxLenAlias = len(a.Alias)
		}
		if len(a.Path) > maxLenPath {
			maxLenPath = len(a.Path)
		}
	}

	fmtStr := fmt.Sprintf("  %%-%ds  %%-%ds ", maxLenAlias, maxLenPath+1)
	for i, a := range m.Config.DirAliases {
		if i == m.selectedIdx {
			s += styles.selected.Render(fmt.Sprintf(fmtStr, a.Alias, a.Path))
			s += "\n"
			continue
		}

		s += styles.rest.Render(fmt.Sprintf(fmtStr, a.Alias, a.Path))
		s += "\n"

		if i >= 9 {
			break
		}
	}

	if m.helpShown {
		s += "\n?         hide key bindings"
		s += "\nctrl+n/↓  move selection down"
		s += "\nctrl+p/↑  move selection up"
		s += "\nctrl+w    clear search string"
		s += "\nctrl+c    quit"
	} else {
		s += "\n?         show key bindings"
		s += "\nctrl+c    quit"
	}
	return styles.window.Render(s) + "\n"
}

func initialModel(configPath string) Model {
	cfg, err := config.Read(configPath)
	if err != nil {
		panic(err)
	}
	return Model{
		Config: cfg,
	}
}

// StartFzf is the entry point for the fuzzy finder which spawns the bubbletea
// program.
func StartFzf(configPath string) *tea.Program {
	return tea.NewProgram(initialModel(configPath))
}
