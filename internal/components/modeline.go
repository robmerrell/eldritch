package components

import (
	tea "charm.land/bubbletea/v2"
	"github.com/robmerrell/eldritch/internal/themes"
)

type Modeline struct {
	theme *themes.Theme
}

func (m Modeline) Init() tea.Cmd {
	return nil
}

func (m Modeline) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Modeline) View() tea.View {
	return tea.NewView("modeline")
}

func NewModeline(theme *themes.Theme) *Modeline {
	return &Modeline{
		theme: theme,
	}
}
