package components

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/robmerrell/eldritch/internal/buffer"
	"github.com/robmerrell/eldritch/internal/themes"
)

type BufferView struct {
	buffer   *buffer.Buffer
	theme    *themes.Theme
	modeline *Modeline
	width    int
	height   int
}

func (b *BufferView) Init() tea.Cmd {
	return nil
}

func (b *BufferView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	_, cmd := b.modeline.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width = msg.Width
		b.height = msg.Height
	}

	return b, cmd
}

func (b *BufferView) View() tea.View {
	// calculate the number of chars in the view
	// area := b.width * b.height
	contentStyle := lipgloss.NewStyle().
		Foreground(b.theme.ModelineInputModeFg).
		Background(b.theme.ModelineInputModeBg).
		Width(b.width).
		Height(b.height - 1)

	var contents strings.Builder
	for lineRunes := range b.buffer.ContentsForRendering() {
		// line is greater than the size so it needs to wrap
		contents.WriteString(string(lineRunes))
	}

	layout := lipgloss.JoinVertical(
		lipgloss.Left,
		b.modeline.View().Content,
		contentStyle.Render(contents.String()))

	return tea.NewView(layout)
}

func NewBufferView(buffer *buffer.Buffer, theme *themes.Theme) *BufferView {
	modeline := NewModeline(theme)

	return &BufferView{
		buffer:   buffer,
		theme:    theme,
		modeline: modeline,
	}
}
