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
		selections: []*Selection{NewSelection(0, 0)},
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

// endOfDocumentOffset calculates the offset that represents the end of the document
func (b *Buffer) endOfDocumentOffset() int {
	offset := 0

	for _, line := range b.contents {
		offset += line.length
	}

	return offset
}

// OffsetToLineNum returns the (0 based) line number that the offset is found in.
func (b *Buffer) OffsetToLineNum(offset int) int {
	acc := 0

	for i, line := range b.contents {
		if acc+line.length > offset {
			return i
		}

		acc += line.length
	}

	// fallback to last line
	return len(b.contents) - 1
}

// LocalLineOffset returns the offset of the current line the selection is on. Think column of the current line.
func (b *Buffer) LocalLineOffset(selection *Selection) int {
	acc := 0

	for _, line := range b.contents {
		if acc+line.length > selection.Head {
			// return i
		}

		acc += line.length
	}

	// fallback to beginning of line
	return 0
}

// AddSelection adds a selection to the buffer at the given offset for both head and anchor.
func (b *Buffer) AddSelection(offset int) {
	b.selections = append(b.selections, NewSelection(offset, offset))
}

// OffsetAttribute returns a single attribute for the given rune offset. I suspect this will go
// away once I need to render diagnostics.
func (b *Buffer) OffsetAttribute(lineIndex, offset int) string {
	contentOffset := offset
	for i := range lineIndex {
		contentOffset += b.contents[i].length
	}

	for _, selection := range b.selections {
		if contentOffset >= selection.Anchor && contentOffset <= selection.Head-1 {
			return "selection_tail"
		} else if contentOffset == selection.Head {
			return "selection_head"
		}
	}

	return "none"
}

// ShiftSelectionsForward shifts the selections "count" spaces forward. If collapsed is true then
// also move the anchor.
func (b *Buffer) ShiftSelectionsForward(count int, collapse bool) {
	for _, selection := range b.selections {
		selection.Head = min(selection.Head+count, b.endOfDocumentOffset())
		selection.PreferredLineOffset = b.LocalLineOffset(selection)

		if collapse {
			selection.Anchor = selection.Head
		}
	}
}

// ShiftSelectionsBackward shifts the selections "count" spaces backward. If collapsed is true then
// also move the anchor.
func (b *Buffer) ShiftSelectionsBackward(count int, collapse bool) {
	for _, selection := range b.selections {
		selection.Head = max(selection.Head-count, 0)

		if collapse {
			selection.Anchor = selection.Head
		}
	}
}

// ShiftSelectionsDown shifts the selections "count" spaces down. If collapsed is true then
// also move the anchor.
func (b *Buffer) ShiftSelectionsDown(count int, collapse bool) {
	for _, selection := range b.selections {
		// find the line number we need to jump to
		currentLineNum := b.OffsetToLineNum(selection.Head)
		lineNum := min(len(b.contents), currentLineNum+count)

		// find the offset within the line. If there is a preferred column on the selection use that
		// otherwise use the line end.
		lineOffset := min(selection.PreferredLineOffset, b.contents[lineNum].length)

		// find the offset in the buffer
		bufferOffset := 0
		for i := range lineNum {
			bufferOffset += b.contents[i].length
		}

		selection.Head = bufferOffset + lineOffset

		if collapse {
			selection.Anchor = selection.Head
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
