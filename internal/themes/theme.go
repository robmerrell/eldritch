package themes

import "image/color"

type Theme struct {
	// main colors
	Bg color.Color
	Fg color.Color

	// selection colors
	SelectionHeadFg color.Color
	SelectionHeadBg color.Color
	SelectionTailFg color.Color
	SelectionTailBg color.Color

	// Modeline
	ModelineBg          color.Color
	ModelineFg          color.Color
	ModelineInputModeBg color.Color
	ModelineInputModeFg color.Color
}
