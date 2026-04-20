package state

import (
	tea "charm.land/bubbletea/v2"
)

// MsgModeKeyPress wraps a keypress message with a mode that that the
// keypress should be applied to.
type MsgModeKeyPress struct {
	Mode     InputMode
	PressMsg tea.KeyPressMsg
}

// MsgModeChanged is fired whenever we change modes.
type MsgModeChanged struct {
	OldMode InputMode
	NewMode InputMode
}
