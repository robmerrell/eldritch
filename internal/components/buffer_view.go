package components

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/robmerrell/eldritch/internal/buffer"
	"github.com/robmerrell/eldritch/internal/themes"
)

type BufferView struct {
	buffer   *buffer.Buffer
	theme    *themes.Theme
	modeline *Modeline
}

func (b *BufferView) Init() tea.Cmd {
	return nil
}

func (b *BufferView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return b.modeline.Update(msg)
}

func (b *BufferView) View() tea.View {

	// calculate the number of chars in a full screen

	contents := b.buffer.ContentsForRendering()

	layout := lipgloss.JoinVertical(
		lipgloss.Left,
		b.modeline.View().Content,
		contents)

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
