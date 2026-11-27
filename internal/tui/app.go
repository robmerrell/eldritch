package tui

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/davecgh/go-spew/spew"
)

type rootModel struct {
	// screen sizes
	screenWidth  int
	screenHeight int

	// main container for editing content
	editContainer *EditContainer
}

func (m *rootModel) Init() tea.Cmd {
	return nil
}

func (m *rootModel) View() tea.View {
	view := tea.NewView(m.editContainer.View().Content)
	view.AltScreen = true
	return view
}

func (m *rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// log out the messages
	log.Println(spew.Sdump(msg))
	log.Println("-----------")

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.screenWidth = msg.Width
		m.screenHeight = msg.Height
		// m.rootView.Update(msg)

	// handle keypress events. We only handle state global state transitions in
	// this module. Everything else is delegated to components.
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			// quit for now
			return m, tea.Quit
		}
	}

	return m, nil
}

func InitApp(fileName *string) (*rootModel, error) {
	return &rootModel{
		editContainer: NewEditContainer(),
	}, nil
}
