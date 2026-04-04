package themes

import "charm.land/lipgloss/v2"

// a minimal dark theme. I can't remember where I got some of the colors from. Tokyo night, catpuccin?

func BatSquatch() *Theme {
	return &Theme{
		Bg: lipgloss.Color("#24283b"),
		Fg: lipgloss.Color("#c0caf5"),

		ModelineBg:          lipgloss.Color("#1f2335"),
		ModelineFg:          lipgloss.Color("#c0caf5"),
		ModelineInputModeBg: lipgloss.Color("#181b29"),
		ModelineInputModeFg: lipgloss.Color("#c0caf5"),
	}
}
