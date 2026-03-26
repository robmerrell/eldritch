package app

import (
	tea "charm.land/bubbletea/v2"
)

type InputState int

const (
	// Similar to normal mode in Kakoune
	InputStateNormal InputState = iota

	// Insert characters
	InputStateInsert
)

type rootModel struct {
	currentInputState InputState
}

func (m rootModel) Init() tea.Cmd {
	return nil
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// handle keypress events
	case tea.KeyPressMsg:
		switch m.currentInputState {
		case InputStateNormal:
			return m.handleNormalStateKey(msg.String())
		case InputStateInsert:
			return m.handleInsertStateKey(msg.String())
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m rootModel) View() tea.View {
	stateString := "u"
	switch m.currentInputState {
	case InputStateNormal:
		stateString = "n"
	case InputStateInsert:
		stateString = "i"
	}

	return tea.NewView(stateString)
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
	return rootModel{
		currentInputState: InputStateNormal,
	}
}
