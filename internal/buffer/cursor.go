package buffer

type SelectionDirection int

const (
	SelectionDirectionUp SelectionDirection = iota
	SelectionDirectionRight
	SelectionDirectionDown
	SelectionDirectionLeft
)

type Cursor struct {
	X int
	Y int
}

type Selection struct {
	// start of the selection
	AnchorX uint
	AnchorY uint

	// moving end of the selection. Can be thought of as the cursor.
	HeadX uint
	HeadY uint
}

func NewSelection(x, y uint) *Selection {
	return &Selection{
		AnchorX: x,
		AnchorY: y,
		HeadX:   x,
		HeadY:   y,
	}
}

func (s *Selection) Beginning() (uint, uint) {
	// the most common case is anchor and head being the same coords so check for it first
	if (s.AnchorY == s.HeadY) && (s.AnchorX == s.HeadX) {
		return s.HeadX, s.HeadY
	}

	if s.AnchorY < s.HeadY {
		return s.AnchorX, s.AnchorY
	} else if s.AnchorY == s.HeadY {
		// we're on the same line
		if s.AnchorX < s.HeadX {
			return s.AnchorX, s.AnchorY
		} else {
			return s.HeadX, s.HeadY
		}
	} else {
		return s.HeadX, s.HeadY
	}
}

func NewCursor(x, y int) *Cursor {
	return &Cursor{
		X: x,
		Y: y,
	}
}
