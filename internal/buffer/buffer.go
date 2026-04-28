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
	"bufio"
	"errors"
	"log"
	"os"
	"slices"
	"strings"
)

// errors
var ErrNotFileBackedBuffer = errors.New("Not a file backed buffer")

// Default number of lines and runes per line to preallocate empty buffers
const DefaultLineCap = 256
const DefaultRuneCap = 256

// line holds line contents and the length. All lines end with a newline.
type line struct {
	runes  []rune
	length int
}

// Buffer is the backing structure for an editable document. Similar to Kakoune and Helix
// buffers can have multiple selections active at one time.
type Buffer struct {
	// just a slice of lines for now. We'll optimize this later.
	contents []line

	// the file backing the buffer, if backed by a file.
	backingFile *string

	// optional name of the buffer
	name *string

	// has the buffer changed
	dirty bool

	// selections are like cursors. Similar to Kakoune and Helix all cursors
	// are selections. Even selections of 1.
	selections       []*Selection
	primarySelection *Selection
}

// NewBuffer creates a new buffer with default options.
func NewBuffer() *Buffer {
	b := &Buffer{
		selections: []*Selection{NewSelection(0, 0, 0, 0)},
	}
	b.primarySelection = b.selections[0]
	b.Clear()

	return b
}

// NewBufferWithFile creates a new file backed buffer.
func NewBufferWithFile(filePath string) (*Buffer, error) {
	buffer := NewBuffer()
	err := buffer.LoadFile(filePath)

	return buffer, err
}

// Clear clears the buffer input by reallocating the content container.
func (b *Buffer) Clear() {
	b.contents = make([]line, 1, DefaultLineCap)
	b.contents[0] = newLine([]rune(""))
}

// AssignName gives the buffer a name
func (b *Buffer) AssignName(name string) {
	b.name = &name
}

// LogSelections is a temporary debug helper
func (b *Buffer) LogSelections() {
	for i, sel := range b.selections {
		log.Printf("%d -- row: %d, col: %d", i, sel.HeadRow, sel.HeadCol)
	}
}

// Insert inserts a rune at all selection positions. Characters are inserted before the selection.
func (b *Buffer) Insert(input rune) {
	/*
		for _, selection := range b.selections {
			// get the beginning of the selection regardless of anchor coords
			x, y := selection.Beginning()
			line := &b.contents[y]

			if line.length == x {
				// at the end of the line, so append
				line.runes = append(line.runes, input)
			} else if line.length > x {
				// Something already exists there, so insert it
				line.runes = slices.Insert(line.runes, int(x), input)
			}

			line.length = line.length + 1
			b.shiftSelection(selection, SelectionDirectionRight, 1)
		}
	*/
}

// SetContents replaces current contents with the given input.
func (b *Buffer) SetContents(contents string) {
	split := strings.SplitAfter(contents, "\n")
	b.contents = make([]line, len(split))

	for i, strLine := range split {
		b.contents[i] = newLine([]rune(strLine))
	}
}

// OffsetAttribute returns a single attribute for the given rune offset. I suspect this will go
// away once I need to render diagnostics.
func (b *Buffer) OffsetAttribute(row, col int) string {
	for _, selection := range b.selections {
		// head
		if selection.HeadRow == row && selection.HeadCol == col {
			return "selection_head"
		}

		// tail
		if min(selection.HeadRow, selection.AnchorRow) <= row &&
			row <= max(selection.HeadRow, selection.AnchorRow) &&
			min(selection.HeadCol, selection.AnchorCol) <= col &&
			col <= max(selection.HeadCol, selection.AnchorCol) {
			return "selection_tail"
		}
	}

	return "none"
}

// ShiftSelectionsForward shifts the selections "count" spaces forward. If collapsed is true then
// also move the anchor.
func (b *Buffer) ShiftSelectionsForward(count int, collapse bool) {
	for _, selection := range b.selections {
		// find the row the shift will move to and then move any leftover columns
		row := selection.HeadRow
		runesLeft := count
		for runesLeft > 0 && row < len(b.contents) {
			line := b.contents[row]
			if line.length > selection.HeadCol+count {
				selection.HeadCol += count
				break
			}
			runesLeft -= line.length

			// shift to the beginning of the next line
			row += 1
			selection.HeadCol = 0
		}

		selection.HeadRow = row
		selection.PreferredLineOffset = selection.HeadCol

		if collapse {
			selection.AnchorCol = selection.HeadCol
			selection.AnchorRow = selection.HeadRow
		}
	}
}

// ShiftSelectionsBackward shifts the selections "count" spaces backward. If collapsed is true then
// also move the anchor.
func (b *Buffer) ShiftSelectionsBackward(count int, collapse bool) {
	for _, selection := range b.selections {
		// find the row the shift will move to and then move any leftover columns
		row := selection.HeadRow
		runesLeft := count
		for runesLeft > 0 && row >= 0 {
			line := b.contents[row]
			if selection.HeadCol-count >= 0 {
				selection.HeadCol -= count
				break
			}
			runesLeft -= line.length

			// shift to the end of the previous line
			if row > 0 {
				row -= 1
				selection.HeadCol = b.contents[row].length - 1
			}
		}

		selection.HeadRow = row
		selection.PreferredLineOffset = selection.HeadCol

		if collapse {
			selection.AnchorCol = selection.HeadCol
			selection.AnchorRow = selection.HeadRow
		}
	}

}

// ShiftSelectionsDown shifts the selections "count" spaces down. If collapsed is true then
// also move the anchor.
func (b *Buffer) ShiftSelectionsDown(count int, collapse bool) {
	for _, selection := range b.selections {
		targetLine := min(len(b.contents), selection.HeadRow+count)

		selection.HeadRow = targetLine
		selection.HeadCol = min(selection.PreferredLineOffset, b.contents[targetLine].length-1)

		if collapse {
			selection.AnchorCol = selection.HeadCol
			selection.AnchorRow = selection.HeadRow
		}
	}
}

// ShiftSelectionsUp shifts the selections "count" spaces up. If collapsed is true then
// also move the anchor.
func (b *Buffer) ShiftSelectionsUp(count int, collapse bool) {
	for _, selection := range b.selections {
		targetLine := max(0, selection.HeadRow-count)

		selection.HeadRow = targetLine
		selection.HeadCol = min(selection.PreferredLineOffset, b.contents[targetLine].length-1)

		if collapse {
			selection.AnchorCol = selection.HeadCol
			selection.AnchorRow = selection.HeadRow
		}
	}
}

// ContentsForRendering returns a portion of the buffer suitable for rendering. This startLine is
// the first line to render and the maxLine is the last possible line to render. Possible because
// the viewport might be able to render 25 lines, but the buffer only has 5. So just the 5 are returned.
func (b *Buffer) ContentsForRendering(startLine, maxLine int) [][]rune {
	lineCount := len(b.contents)
	latestLine := min(maxLine, lineCount)

	lineContents := make([][]rune, latestLine-startLine)
	for i := startLine; i < latestLine; i++ {
		rns := slices.Clone(b.contents[i].runes)
		// render a \n with a space. TODO: This is dumb and needs to be redone.
		rns = append(rns[:len(rns)-1], ' ', '\n')
		lineContents[i-startLine] = rns
	}

	return lineContents
}

// Selections returns all selections active in the buffer
func (b *Buffer) Selections() []*Selection {
	return b.selections
}

// LoadFile loads a file into the buffer
func (b *Buffer) LoadFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	b.backingFile = &filePath
	b.contents = make([]line, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineRunes := []rune(scanner.Text())
		b.contents = append(b.contents, newLine(lineRunes))
	}

	return nil
}

// newLine creates an empty newline with the required line ending
func newLine(lineRunes []rune) line {
	// make sure it ends with a newline
	if len(lineRunes) == 0 || lineRunes[len(lineRunes)-1] != '\n' {
		lineRunes = append(lineRunes, []rune("\n")...)
	}

	return line{runes: lineRunes, length: len(lineRunes)}
}
