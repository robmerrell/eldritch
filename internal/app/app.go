package app

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/davecgh/go-spew/spew"
	"github.com/robmerrell/eldritch/internal/buffer"
	"github.com/robmerrell/eldritch/internal/components"
	"github.com/robmerrell/eldritch/internal/state"
	"github.com/robmerrell/eldritch/internal/themes"
)

type rootModel struct {
	theme            *themes.Theme
	currentInputMode state.InputMode

	// screen sizes
	screenWidth  int
	screenHeight int

	// ui components
	rootView *components.BufferView
}

func (m *rootModel) Init() tea.Cmd {
	return nil
}

func (m *rootModel) View() tea.View {
	mainView := tea.NewView(m.rootView.View().Content)
	mainView.AltScreen = true
	mainView.BackgroundColor = m.theme.Bg
	mainView.ForegroundColor = m.theme.Fg

	return mainView
}

func (m *rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// log out the messages
	log.Println(spew.Sdump(msg))
	log.Println("-----------")

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.screenWidth = msg.Width
		m.screenHeight = msg.Height
		m.rootView.Update(msg)

	// handle keypress events. We only handle state global state transitions in
	// this module. Everything else is delegated to components.
	case tea.KeyPressMsg:
		switch m.currentInputMode {
		case state.InputModeNormal:
			return m.handleNormalModeKey(msg)
		case state.InputModeInsert:
			return m.handleInsertModeKey(msg)
		}
	}

	return m, nil
}

func (m *rootModel) handleNormalModeKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		// quit for now
		return m, tea.Quit

	// case "alt-x", ":":
	// 	// enter command mode
	// 	return m, m.enterMode(state.InputModeCommand)

	case "i":
		// enter insert mode
		return m, m.enterMode(state.InputModeInsert)
	}

	// anything else send to the active buffer view
	// wrap the event in the current state before sending it
	_, rootCmd := m.rootView.Update(state.MsgModeKeyPress{Mode: m.currentInputMode, PressMsg: msg})
	return m, rootCmd
}

func (m *rootModel) handleInsertModeKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	// exit insert mode
	case "esc", "ctrl+g":
		return m, m.enterMode(state.InputModeNormal)
	}

	// TODO: any unhandled control keys should just return early

	// anything else send to the active buffer view
	// wrap the event in the current state before sending it
	_, rootCmd := m.rootView.Update(state.MsgModeKeyPress{Mode: m.currentInputMode, PressMsg: msg})
	return m, rootCmd
}

// enterMode switches the input mode and then returns a wrapped event to pass along
// to the child components.
func (m *rootModel) enterMode(mode state.InputMode) tea.Cmd {
	msg := state.MsgModeChanged{OldMode: m.currentInputMode, NewMode: mode}
	m.currentInputMode = mode

	return func() tea.Msg {
		return msg
	}
}

func Init(fileName *string) (*rootModel, error) {
	theme := themes.BatSquatch()

	// initial empty buffer
	var startBuffer *buffer.Buffer

	if fileName == nil {
		startBuffer = buffer.NewBuffer()
		startBuffer.SetContents("hello this is a really long line that should wrap because we hit the maximum width of the terminal that can display it. We also need to test how wrap characters get inserted\nworld\nthis\nis a buffer")
	} else {
		var err error
		startBuffer, err = buffer.NewBufferWithFile(*fileName)
		if err != nil {
			return nil, err
		}
	}

	return &rootModel{
		theme:            theme,
		currentInputMode: state.InputModeNormal,
		rootView:         components.NewBufferView(startBuffer, theme),
	}, nil
}
