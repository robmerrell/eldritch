package components

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/robmerrell/eldritch/internal/themes"
)

type Modeline struct {
	theme *themes.Theme
	width int
}

func (m *Modeline) Init() tea.Cmd {
	return nil
}

func (m *Modeline) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		log.Printf("msg w: %d", msg.Width)
		m.width = msg.Width

	}

	return m, nil
}

func (m *Modeline) View() tea.View {
	barStyle := lipgloss.NewStyle().
		Foreground(m.theme.ModelineFg).
		Background(m.theme.ModelineBg)

	winNumStyle := lipgloss.NewStyle().
		Inherit(barStyle).
		Foreground(m.theme.ModelineInputModeFg).
		Background(m.theme.ModelineInputModeBg).
		Padding(0, 2).
		Width(5)

	// ^ width: 3 + 2

	bufferNameStyle := lipgloss.NewStyle().
		Inherit(barStyle).
		Padding(0, 1).
		Width(m.width - 5 - 10 + 2)
		// Width(45)

	// ^ flex width

	offsetSyle := lipgloss.NewStyle().
		Inherit(barStyle).
		Padding(0, 1).
		Width(8).
		Align(lipgloss.Right)

	// width 10

	winNum := winNumStyle.Render("1")
	bufferName := bufferNameStyle.Render("components/modeline.go (main)")
	offset := offsetSyle.Render("1:0")

	content := lipgloss.JoinHorizontal(lipgloss.Top, winNum, bufferName, offset)
	return tea.NewView(content)
}

func NewModeline(theme *themes.Theme) *Modeline {
	return &Modeline{
		theme: theme,
	}
}
