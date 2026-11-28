package tui

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/davecgh/go-spew/spew"
	"github.com/robmerrell/eldritch/internal/buffer"
	"github.com/robmerrell/eldritch/internal/messages"
)

type rootModel struct {
	// screen sizes
	screenWidth  int
	screenHeight int

	// main container for editing content
	editContainer *EditContainer

	// global list of open buffers
	buffers []*buffer.Buffer
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

		// while this is the only thing being rendered it will just be the whole screen
		m.editContainer.Update(messages.UpdateSizeMsg{Width: msg.Width, Height: msg.Height})

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
	// The app either starts with a buffer loaded from a file or a new empty one
	var startBuffer *buffer.Buffer
	var err error

	if fileName == nil {
		startBuffer = buffer.NewBuffer()
	} else {
		startBuffer, err = buffer.NewBufferWithFile(*fileName)
		if err != nil {
			return nil, err
		}
	}

	return &rootModel{
		buffers:       []*buffer.Buffer{startBuffer},
		editContainer: NewEditContainer(startBuffer),
	}, nil
}
