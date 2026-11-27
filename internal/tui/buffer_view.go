package tui

import tea "charm.land/bubbletea/v2"

// BufferView is our visual interface into a buffer. All interactions on a buffer must go
// through a BufferView. BufferViews are also responsible for maintaining a list of active
// selections on a buffer. Multiple BufferViews can all point at the same buffer.
type BufferView struct{}

func (b *BufferView) Init() tea.Cmd {
	return nil
}

func (b *BufferView) View() tea.View {
	return tea.NewView("Buffer View Content")
}

func (b *BufferView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return b, nil
}

func NewBufferView() *BufferView {
	return &BufferView{}
}
