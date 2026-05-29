package buffer

// How selections and positioning work within the buffer:
//
// The selection exists on a point and inserting is done before the beginning point
// of the selection and appending done after the end point.
//
// imagine the line of text in a buffer and S representing the selection point on that line:
// hello world
//     S
//
// The selection x (0 indexed) is 4.
//
// If I insert the character "x" at this point we would end up with:
// hellxo world
//
// This also means that the selection can be further ahead on the line than the last character:
// hello world
//            S
//
// the letter "d" is at x 10, but the selection is at 11. Inserting "x" here would give
// hello worldx
//             S

import (
	"errors"
	"log"
	"os"
	"slices"

	"github.com/robmerrell/eldritch/internal/piece"
)

// errors
var ErrNotFileBackedBuffer = errors.New("Not a file backed buffer")

// Buffer is the backing structure for an editable document. Similar to Kakoune and Helix
// buffers can have multiple selections active at one time.
type Buffer struct {
	Contents *piece.PieceTable

	// the file backing the buffer, if backed by a file.
	backingFile *string

	// optional name of the buffer
	name *string

	// has the buffer changed
	dirty bool

	// selections are like cursors. Similar to Kakoune and Helix all cursors
	// are selections. Even selections of 1.
	selections []*Selection
}

// NewBuffer creates a new buffer with default options.
func NewBuffer() *Buffer {
	b := &Buffer{
		selections: []*Selection{NewSelection(0, 0, 0)},
	}
	b.Contents = piece.FromSlice([]rune(""))

	return b
}

// NewBufferWithFile creates a new file backed buffer.
func NewBufferWithFile(filePath string) (*Buffer, error) {
	buffer := NewBuffer()
	err := buffer.LoadFile(filePath)

	return buffer, err
}

// AssignName gives the buffer a name
func (b *Buffer) AssignName(name string) {
	b.name = &name
}

// LogSelections is a temporary debug helper
func (b *Buffer) LogSelections() {
	for i, sel := range b.selections {
		coords := b.Contents.OffsetToCoords(sel.HeadOffset)
		log.Printf("%d -- head: %d, anchor: %d, row: %d, col: %d", i, sel.HeadOffset, sel.AnchorOffset, coords.Row, coords.Col)
	}
}

// Insert inserts a rune at all selection positions. Characters are inserted before the selection.
func (b *Buffer) Insert(input []rune) {
	for _, selection := range b.selections {
		err := b.Contents.Insert(selection.HeadOffset, input)
		if err != nil {
			log.Println(err)
			continue
		}

		b.ShiftSelectionsForward(len(input), selection.IsCollapsed())
	}
}

// Delete delets the characters contained in each selection.
func (b *Buffer) Delete() {
	for _, selection := range b.selections {
		start := min(selection.HeadOffset, selection.AnchorOffset)
		end := max(selection.HeadOffset, selection.AnchorOffset)

		err := b.Contents.Delete(start, end)
		if err != nil {
			log.Println(err)
		}

		b.ShiftSelectionsBackward(end-start, true)
	}
}

// SetContents replaces current contents with the given input.
func (b *Buffer) SetContents(contents string) {
	b.Contents = piece.FromSlice([]rune(contents))
}

// OffsetAttribute returns a single attribute for the given rune offset. I dislike this and want
// it to go away, but it's helpful for now.
func (b *Buffer) OffsetAttribute(offset int) string {
	for _, selection := range b.selections {
		// head
		if selection.HeadOffset == offset {
			return "selection_head"
		}

		// tail
		start := min(selection.AnchorOffset, selection.HeadOffset)
		end := max(selection.AnchorOffset, selection.HeadOffset)
		if offset >= start && offset <= end {
			return "selection_tail"
		}
	}

	return "none"
}

// AddOpenSelection adds a new open selection at the given location. Returns a reference
// to the new selection.
/*
func (b *Buffer) AddOpenSelection(headRow, headCol, anchorRow, anchorCol int) *Selection {
	selection := NewSelection(headRow, headCol, anchorRow, anchorCol)
	b.selections = append(b.selections, selection)
	b.sortSelections()

	return selection
}

// AddCollapsedSelection adds a new collapsed selection at the given location. Returns a reference
// to the new selection.
func (b *Buffer) AddCollapsedSelection(row, col int) *Selection {
	selection := NewSelection(row, col, row, col)
	b.selections = append(b.selections, selection)
	b.sortSelections()

	return selection
}

// sortSelections sorts the selections from the top of the document to the end of the document. This
// should be done anytime a selection is added.
func (b *Buffer) sortSelections() {
	slices.SortFunc(b.selections, func(a, b *Selection) int {
		return cmp.Or(
			cmp.Compare(a.HeadRow, b.HeadRow),
			cmp.Compare(a.HeadCol, b.HeadCol),
		)
	})
}
*/

// ShiftSelectionsForward shifts the selections "count" spaces forward. If collapsed is true then
// also move the anchor.
func (b *Buffer) ShiftSelectionsForward(count int, collapse bool) {
	for _, selection := range b.selections {
		target := min(b.Contents.Len()-1, selection.HeadOffset+count)
		selection.HeadOffset = target

		coords := b.Contents.OffsetToCoords(target)
		selection.PreferredCol = coords.Col

		if collapse {
			selection.AnchorOffset = target
		}
	}
}

// ShiftSelectionsBackward shifts the selections "count" spaces backward. If collapsed is true then
// also move the anchor.
func (b *Buffer) ShiftSelectionsBackward(count int, collapse bool) {
	for _, selection := range b.selections {
		target := max(0, selection.HeadOffset-count)
		selection.HeadOffset = target

		coords := b.Contents.OffsetToCoords(target)
		selection.PreferredCol = coords.Col

		if collapse {
			selection.AnchorOffset = target
		}
	}
}

// ShiftSelectionsDown shifts the selections "count" spaces down. If collapsed is true then
// also move the anchor.
func (b *Buffer) ShiftSelectionsDown(count int, collapse bool) {
	for _, selection := range b.selections {
		coords := b.Contents.OffsetToCoords(selection.HeadOffset)
		targetRow := min(coords.Row+count, b.Contents.LineCount()-1)

		// get the furthest line we can jump to with the given count
		line := slices.Collect(b.Contents.Lines(targetRow, targetRow))[0]

		// get the length and find the right column to use
		targetCol := min(selection.PreferredCol, len(line.Runes)-1)

		// convert the new line coords with the pref column to an offset
		coords.Row = targetRow
		coords.Col = targetCol
		offset := b.Contents.CoordsToOffset(coords)

		selection.HeadOffset = offset
		if collapse {
			selection.AnchorOffset = offset
		}
	}
}

// ShiftSelectionsUp shifts the selections "count" spaces up. If collapsed is true then
// also move the anchor.
func (b *Buffer) ShiftSelectionsUp(count int, collapse bool) {
	for _, selection := range b.selections {
		coords := b.Contents.OffsetToCoords(selection.HeadOffset)
		targetRow := max(coords.Row-count, 0)

		// get the furthest line we can jump to with the given count
		line := slices.Collect(b.Contents.Lines(targetRow, targetRow))[0]

		// get the length and find the right column to use
		targetCol := min(selection.PreferredCol, len(line.Runes)-1)

		// convert the new line coords with the pref column to an offset
		coords.Row = targetRow
		coords.Col = targetCol
		offset := b.Contents.CoordsToOffset(coords)

		selection.HeadOffset = offset
		if collapse {
			selection.AnchorOffset = offset
		}
	}
}

// SelectLine anchors to the beginning of the line and moves the head to the end.
func (b *Buffer) SelectLine() {
	for _, selection := range b.selections {
		coords := b.Contents.OffsetToCoords(selection.HeadOffset)
		line := slices.Collect(b.Contents.Lines(coords.Row, coords.Row))[0]
		selectLength := len(line.Runes) - 1

		selection.AnchorOffset = b.Contents.CoordsToOffset(piece.Coords{Row: coords.Row, Col: 0})
		selection.HeadOffset = b.Contents.CoordsToOffset(piece.Coords{Row: coords.Row, Col: selectLength})
		selection.PreferredCol = selectLength
	}
}

// Selections returns all selections active in the buffer
func (b *Buffer) Selections() []*Selection {
	return b.selections
}

// LoadFile loads a file into the buffer
func (b *Buffer) LoadFile(filePath string) error {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	b.backingFile = &filePath
	b.Contents = piece.FromSlice([]rune(string(bytes)))
	return nil
}
