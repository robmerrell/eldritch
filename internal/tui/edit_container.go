package tui

import tea "charm.land/bubbletea/v2"

// EditContainer holds the main "panel" for our editable content. Splits, BufferViews, etc.
// It isn't strictly necessary now, but I have a feeling when I start on the bottom panel
// I'll be happy that I have this even though it doesn't do much but act as a predictable
// node in the application component tree.
type EditContainer struct {
	// The renderable child can either be splits (nested allowed) or just a single
	// BufferView. I might narrow down this interface in the future, but right now
	// this can technically render any tea model.
	content tea.Model
}

func (e *EditContainer) Init() tea.Cmd {
	return nil
}

func (e *EditContainer) View() tea.View {
	return e.content.View()
}

func (e *EditContainer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return e, nil
}

func NewEditContainer() *EditContainer {
	return &EditContainer{
		content: NewBufferView(),
	}
}
