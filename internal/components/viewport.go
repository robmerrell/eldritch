package components

import (
	tea "charm.land/bubbletea/v2"
	"github.com/robmerrell/eldritch/internal/buffer"
)

type Viewport struct {
	buffer *buffer.Buffer
}

func (v *Viewport) Init() tea.Cmd {
	return nil
}

func (v *Viewport) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return v, nil
}

func (v *Viewport) View() tea.View {
	return tea.NewView(v.buffer.ContentsForRendering())
}

func NewViewport(buffer *buffer.Buffer) *Viewport {
	return &Viewport{
		buffer: buffer,
	}
}
