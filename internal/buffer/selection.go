package buffer

type SelectionDirection int

const (
	SelectionDirectionUp SelectionDirection = iota
	SelectionDirectionRight
	SelectionDirectionDown
	SelectionDirectionLeft
)

// Selection works similarly to how I imagine Kakoune and Helix selections work. Instead
// of the primitive text entry point being a cursor it is a selection. Every selection has
// an anchor point and a head point. Inserting is done at the beginning of the selection
// and appending done at the end.
//
// The anchor and head having the same coordinates is a valid state. This is called collapsed
// and this causes the selection to act more like a traditional cursor.
type Selection struct {
	// start of the selection
	AnchorX int
	AnchorY int

	// moving end of the selection. Can be thought of as the cursor.
	HeadX int
	HeadY int
}

// NewSelection creates a new selection with both anchor and head pointing at the same location.
func NewSelection(x, y int) *Selection {
	return &Selection{
		AnchorX: x,
		AnchorY: y,
		HeadX:   x,
		HeadY:   y,
	}
}

// Beginning returns the anchor or head location with the position closest to the top left
// of the document.
func (s *Selection) Beginning() (int, int) {
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

// SetAnchor sets the anchor location
func (s *Selection) SetAnchor(x, y int) {
	s.AnchorX = x
	s.AnchorY = y
}

// SetHead sets the head location
func (s *Selection) SetHead(x, y int) {
	s.HeadX = x
	s.HeadY = y
}

// SetCollapsed sets both the anchor and head locations to the same point
func (s *Selection) SetCollapsed(x, y int) {
	s.SetAnchor(x, y)
	s.SetHead(x, y)
}
