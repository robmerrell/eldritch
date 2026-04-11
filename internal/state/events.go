package state

import (
	tea "charm.land/bubbletea/v2"
)

type MsgModeKeyPress struct {
	Mode     InputMode
	PressMsg tea.Msg
}

// type EventModeChanged struct {
// 	oldMode InputMode
// 	newMode InputMode
// }
