package themes

import "charm.land/lipgloss/v2"

// a minimal blue theme. I can't remember where I got some of the colors from.

func BatSquatch() *Theme {
	return &Theme{
		Bg: lipgloss.Color("#24283b"),
		Fg: lipgloss.Color("#c0caf5"),

		ModelineBg: lipgloss.Color("#1f2335"),
		ModelineFg: lipgloss.Color("#ff0000"),
	}
}
