package components

import (
	tea "charm.land/bubbletea/v2"
	"github.com/robmerrell/eldritch/internal/buffer"
)

type BufferView struct {
	buffer *buffer.Buffer
}

func (b *BufferView) Init() tea.Cmd {
	return nil
}

func (b *BufferView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return b, nil
}

func (b *BufferView) View() tea.View {
	return tea.NewView(b.buffer.ContentsForRendering())
}

func NewBufferView(buffer *buffer.Buffer) *BufferView {
	return &BufferView{
		buffer: buffer,
	}
}
