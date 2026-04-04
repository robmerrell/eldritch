package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/robmerrell/eldritch/internal/buffer"
	"github.com/robmerrell/eldritch/internal/components"
	"github.com/robmerrell/eldritch/internal/themes"
)

type InputState int

const (
	// Similar to normal mode in Kakoune
	InputStateNormal InputState = iota

	// Insert characters
	InputStateInsert
)

type rootModel struct {
	theme             *themes.Theme
	currentInputState InputState

	// screen sizes
	screenWidth  int
	screenHeight int

	// ui components
	rootView *components.BufferView
}

func (m rootModel) Init() tea.Cmd {
	return nil
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.screenWidth = msg.Width
		m.screenHeight = msg.Height
		m.rootView.Update(msg)

	// handle keypress events
	case tea.KeyPressMsg:
		switch m.currentInputState {
		case InputStateNormal:
			return m.handleNormalStateKey(msg.String())
		case InputStateInsert:
			return m.handleInsertStateKey(msg.String())
		}
	}

	return m, nil
}

func (m rootModel) View() tea.View {
	mainView := tea.NewView(m.rootView.View().Content)
	mainView.AltScreen = true
	mainView.BackgroundColor = m.theme.Bg
	mainView.ForegroundColor = m.theme.Fg

	return mainView
}

func (m rootModel) handleNormalStateKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	// quit for now
	case "ctrl+c":
		return m, tea.Quit

	// insert mode
	case "i":
		m.currentInputState = InputStateInsert
		return m, nil
	}

	return m, nil
}

func (m rootModel) handleInsertStateKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	// exit insert state with esc or ctrl-g
	case "esc", "ctrl+g":
		m.currentInputState = InputStateNormal

		// insert rune keys
	}

	return m, nil
}

func Init() rootModel {
	theme := themes.BatSquatch()

	// initial empty buffer
	buffer := buffer.NewBuffer()

	return rootModel{
		theme:             theme,
		currentInputState: InputStateNormal,
		rootView:          components.NewBufferView(buffer, theme),
	}
}
