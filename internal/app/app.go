package app

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

	// ui components
	modeline     *components.Modeline
	rootViewport *components.Viewport
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

	return m, nil
}

func (m rootModel) View() tea.View {
	layout := lipgloss.JoinVertical(
		lipgloss.Left,
		m.modeline.View().Content,
		m.rootViewport.View().Content)

	mainView := tea.NewView(layout)
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

	return rootModel{
		theme:             theme,
		currentInputState: InputStateNormal,
		rootViewport:      components.NewViewport(),
		modeline:          components.NewModeline(theme),
	}
}
