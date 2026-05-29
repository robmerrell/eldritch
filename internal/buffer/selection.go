package buffer

// Selection works similarly to how I imagine Kakoune and Helix selections work. Instead
// of the primitive text entry point being a cursor it is a selection. Every selection has
// an anchor point and a head point. Inserting is done at the beginning of the selection
// and appending done at the end.
//
// The anchor and head having the same coordinates is a valid state. This is called collapsed
// and this causes the selection to act more like a traditional cursor.
type Selection struct {
	AnchorOffset int
	HeadOffset   int

	// set the column when moving horizontally so we can preserve it when moving vertically.
	PreferredCol int
}

// NewSelection creates a new selection at the given anchor and head.
func NewSelection(headOffset, anchorOffset, preferredCol int) *Selection {
	return &Selection{HeadOffset: headOffset, AnchorOffset: anchorOffset, PreferredCol: preferredCol}
}

/*
// SetCollapsed sets a collapsed position
func (s *Selection) SetCollapsed(row, col int) {
	s.HeadRow = row
	s.HeadCol = col
	s.AnchorRow = row
	s.AnchorCol = col
	s.PreferredLineOffset = col
}
*/

// SwapPositions swaps the anchor and the head
func (s *Selection) SwapPositions() {
	s.AnchorOffset = s.HeadOffset
	s.PreferredCol = s.HeadOffset
}

// IsCollapsed returns if the selection is collapsed or not.
func (s *Selection) IsCollapsed() bool {
	return s.AnchorOffset == s.HeadOffset
}

// Collapse collapses anchor and head together.
func (s *Selection) Collapse() {
	s.AnchorOffset = s.HeadOffset
}

/*

// PointInSelections returns true if a point is between the anchor and head (inclusive)
func (s *Selection) PointSelected(row, col int) bool {
	startRow := min(s.HeadRow, s.AnchorRow)
	startCol := min(s.HeadCol, s.AnchorCol)
	endRow := max(s.HeadRow, s.AnchorRow)
	endCol := max(s.HeadCol, s.AnchorCol)

	// if we're on a start or end row we need to check the column, otherwise use full row
	if row == startRow && row == endRow {
		return col >= startCol && col <= endCol
	} else if row == startRow {
		// smallest point using row and col
		if startRow == s.HeadRow {
			return col >= s.HeadCol
		} else {
			return col >= s.AnchorCol
		}
	} else if row == endRow {
		// larget point using row and col
		if endRow == s.HeadRow {
			return col <= s.HeadCol
		} else {
			return col <= s.AnchorCol
		}
	}

	return row > startRow && row < endRow
}
*/
