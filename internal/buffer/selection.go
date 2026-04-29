package buffer

// Selection works similarly to how I imagine Kakoune and Helix selections work. Instead
// of the primitive text entry point being a cursor it is a selection. Every selection has
// an anchor point and a head point. Inserting is done at the beginning of the selection
// and appending done at the end.
//
// The anchor and head having the same coordinates is a valid state. This is called collapsed
// and this causes the selection to act more like a traditional cursor.
type Selection struct {
	AnchorCol int
	AnchorRow int

	HeadCol int
	HeadRow int

	// when moving horizontally set this so we can use it when moving vertically
	PreferredLineOffset int
}

// NewSelection creates a new selection at the given anchor and head.
func NewSelection(headRow, headCol, anchorRow, anchorCol int) *Selection {
	return &Selection{HeadRow: headRow, HeadCol: headCol, AnchorRow: anchorRow, AnchorCol: anchorCol}
}

// SwapPositions swaps the anchor and the head
func (s *Selection) SwapPositions() {
	s.AnchorRow, s.HeadRow = s.HeadRow, s.AnchorRow
	s.AnchorCol, s.HeadCol = s.HeadCol, s.AnchorCol
}

// IsCollapsed returns if the selection is collapsed or not.
func (s *Selection) IsCollapsed() bool {
	return s.AnchorRow == s.HeadRow && s.AnchorCol == s.HeadCol
}

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
		return col >= startCol
	} else if row == endRow {
		return col <= endCol
	}

	return row > startRow && row < endRow
}
