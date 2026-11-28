package themes

import "charm.land/lipgloss/v2"

// a minimal dark theme that blends some colors from bloodborne and tokyo night.

func Bloodborne() *Theme {
	return &Theme{
		Bg: lipgloss.Color("#24283b"),
		Fg: lipgloss.Color("#c0caf5"),

		SelectionHeadFg: lipgloss.Color("#ff0000"),
		SelectionHeadBg: lipgloss.Color("#00ff00"),
		SelectionTailFg: lipgloss.Color("#ffaaaa"),
		SelectionTailBg: lipgloss.Color("#aaaaff"),
	}
}
