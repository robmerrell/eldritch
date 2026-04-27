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
	// anchor and head points as rune offsets in the document
	Anchor int
	Head   int

	// when moving horizontally set this so we can use it when moving vertically
	PreferredLineOffset int
}

// NewSelection creates a new selection at the given anchor and head.
func NewSelection(anchor, head int) *Selection {
	return &Selection{Anchor: anchor, Head: head}
}

// SwapPositions swaps the anchor and the head
func (s *Selection) SwapPositions() {
	s.Anchor, s.Head = s.Head, s.Anchor
}

// IsCollapsed returns if the selection is collapsed or not.
func (s *Selection) IsCollapsed() bool {
	return s.Anchor == s.Head
}

// Beginning returns the anchor or head location with the position closest to the top left
// of the document.
// func (s *Selection) Beginning() (int, int) {
// 	// the most common case is anchor and head being the same coords so check for it first
// 	if (s.AnchorY == s.HeadY) && (s.AnchorX == s.HeadX) {
// 		return s.HeadX, s.HeadY
// 	}

// 	if s.AnchorY < s.HeadY {
// 		return s.AnchorX, s.AnchorY
// 	} else if s.AnchorY == s.HeadY {
// 		// we're on the same line
// 		if s.AnchorX < s.HeadX {
// 			return s.AnchorX, s.AnchorY
// 		} else {
// 			return s.HeadX, s.HeadY
// 		}
// 	} else {
// 		return s.HeadX, s.HeadY
// 	}
// }

// SetAnchor sets the anchor location
// func (s *Selection) SetAnchor(x, y int) {
// 	s.AnchorX = x
// 	s.AnchorY = y
// }

// // SetHead sets the head location
// func (s *Selection) SetHead(x, y int) {
// 	s.HeadX = x
// 	s.HeadY = y
// }

// SetCollapsed sets both the anchor and head locations to the same point
// func (s *Selection) SetCollapsed(x, y int) {
// 	s.SetAnchor(x, y)
// 	s.SetHead(x, y)
// }
