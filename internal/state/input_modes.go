package state

type InputMode int

const (
	// Similar to normal mode in Kakoune
	InputModeNormal InputMode = iota

	// Insert characters
	InputModeInsert

	// Run editor commands
	InputModeCommand
)
