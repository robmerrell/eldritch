package themes

import "image/color"

type Theme struct {
	// main colors
	Bg color.Color
	Fg color.Color

	// Modeline
	ModelineBg color.Color
	ModelineFg color.Color
}
