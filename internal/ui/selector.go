package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ksred/ccswitch/internal/git"
)

type SessionSelector struct {
	sessions []git.SessionInfo
	cursor   int
	selected int
	quit     bool
}

func NewSessionSelector(sessions []git.SessionInfo) *SessionSelector {
	return &SessionSelector{
		sessions: sessions,
		selected: -1,
	}
}

func (s *SessionSelector) Init() tea.Cmd {
	return nil
}

func (s *SessionSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"))):
			s.quit = true
			return s, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if s.cursor > 0 {
				s.cursor--
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if s.cursor < len(s.sessions)-1 {
				s.cursor++
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter", " "))):
			s.selected = s.cursor
			return s, tea.Quit
		}
	}
	return s, nil
}

func (s *SessionSelector) View() string {
	if s.quit || s.selected >= 0 {
		return ""
	}

	var b strings.Builder
	
	b.WriteString(TitleStyle.Render("📂 Select session to switch to:"))
	b.WriteString("\n\n")

	for i, session := range s.sessions {
		cursor := "  "
		if s.cursor == i {
			cursor = "→ "
		}
		
		sessionLine := fmt.Sprintf("%s%s (%s)", cursor, session.Name, session.Branch)
		if s.cursor == i {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("202")).Bold(true).Render(sessionLine))
		} else {
			b.WriteString(sessionLine)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("↑/↓/j/k: navigate • enter/space: select • q/esc: quit"))

	return b.String()
}

func (s *SessionSelector) GetSelected() *git.SessionInfo {
	if s.selected >= 0 && s.selected < len(s.sessions) {
		return &s.sessions[s.selected]
	}
	return nil
}

func (s *SessionSelector) IsQuit() bool {
	return s.quit
}