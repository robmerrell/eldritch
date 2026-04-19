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
	"iter"
	"os"
	"slices"
	"strings"
)

// errors
var ErrNotFileBackedBuffer = errors.New("Not a file backed buffer")

// Default number of lines and runes per line to preallocate empty buffers
const DefaultLineCap = 256
const DefaultRuneCap = 256

// line holds line contents and the length
type line struct {
	runes  []rune
	length uint
}

// RenderableLine stores a line ready for rendering. This includes already split out into multiple
// display lines for word wrapping.
type RenderableLine struct {
	LineContents string
	RenderedRows int
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
	b.contents[0].runes = make([]rune, 0, DefaultRuneCap)
}

// AssignName gives the buffer a name
func (b *Buffer) AssignName(name string) {
	b.name = &name
}

// Insert inserts a rune at all selection positions. Characters are inserted before the selection.
func (b *Buffer) Insert(input rune) {
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
}

// SetContents replaces current contents with the given input.
func (b *Buffer) SetContents(contents []rune) {
	b.Clear()

	for _, rn := range contents {
		if rn == '\n' {
			newLine := line{runes: make([]rune, 0, DefaultRuneCap)}
			b.contents = append(b.contents, newLine)
			b.ShiftSelections(SelectionDirectionDown, 1)
		} else {
			b.Insert(rn)
		}
	}
}

// ShiftSelections shifts all selections in a direction
func (b *Buffer) ShiftSelections(direction SelectionDirection, count uint) {
	for _, selection := range b.selections {
		b.shiftSelection(selection, direction, count)
	}
}

// shiftSelection shifts a specified selection in a direction
func (b *Buffer) shiftSelection(selection *Selection, direction SelectionDirection, count uint) {
	switch direction {
	case SelectionDirectionUp:
		// if on the first line don't move
		if selection.HeadY > 0 {
			// if previous line is shorter than current move to end of previous line
			lineLength := b.contents[selection.HeadY].length
			prevLineLength := b.contents[selection.HeadY-1].length
			if prevLineLength < lineLength {
				selection.HeadX = prevLineLength
				selection.AnchorX = prevLineLength
			}

			selection.AnchorY = selection.AnchorY - count
			selection.HeadY = selection.HeadY - count
		}

	case SelectionDirectionRight:
		line := b.contents[selection.HeadY]

		// selection at the end of the line
		if selection.HeadX > line.length-1 {
			// if not the last line wrap around to the next
			if selection.HeadY < uint(len(b.contents)-1) {
				selection.SetCollapsed(0, selection.HeadY+1)
			}

			// if last line don't move
			return
		}

		selection.AnchorX = selection.AnchorX + count
		selection.HeadX = selection.HeadX + count

	case SelectionDirectionDown:
		// if on the last line don't move
		if selection.HeadY < uint(len(b.contents)-1) {
			// if next line is shorter than current move to end of next line if cursor is
			// past the position of the next line's end.
			lineLength := b.contents[selection.HeadY].length
			nextLineLength := b.contents[selection.HeadY+1].length
			if nextLineLength < lineLength && selection.HeadX > nextLineLength {
				selection.HeadX = nextLineLength
				selection.AnchorX = nextLineLength
			}

			selection.AnchorY = selection.AnchorY + count
			selection.HeadY = selection.HeadY + count
		}

	case SelectionDirectionLeft:
		// selection at the beginning of the line
		if selection.HeadX == 0 {
			// if not the first line wrap around to the next
			if selection.HeadY > 0 {
				prevLine := b.contents[selection.HeadY-1]
				selection.SetCollapsed(prevLine.length, selection.HeadY-1)
			}

			// if first line don't move
			return
		}

		selection.AnchorX = selection.AnchorX - count
		selection.HeadX = selection.HeadX - count
	}
}

// ContentsForRendering is an iterator that yields a Renderable line for each line to be rendered to the
// terminal. Line wrapping is done here.
func (b *Buffer) ContentsForRendering(startLine, contentHeight, contentWidth int) iter.Seq[*RenderableLine] {
	// the latest line is the possible last line accoring to contentHeight, but not guaranteed to be
	// displayed because some lines might wrap. Clamp it to the last content line if necessary.
	latestLine := min(startLine+contentHeight, len(b.contents))

	return func(yield func(*RenderableLine) bool) {
		for i := startLine; i < latestLine; i++ {
			renderableLine := &RenderableLine{RenderedRows: 1}

			// if the line length is greater than the width then wrap
			if b.contents[i].length > uint(contentWidth) {
				wrappedLine := make([]string, 0, 2)
				for chunk := range slices.Chunk(b.contents[i].runes, contentWidth) {
					wrappedLine = append(wrappedLine, string(chunk))
					renderableLine.RenderedRows += 1
				}

				renderableLine.LineContents = strings.Join(wrappedLine, "\n")
			} else {
				renderableLine.LineContents = string(b.contents[i].runes)
			}

			if !yield(renderableLine) {
				return
			}
		}
	}
}

// LoadFile loads a file into the buffer
func (b *Buffer) LoadFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	b.contents = make([]line, 0)
	b.backingFile = &filePath

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineRunes := []rune(scanner.Text())
		line := line{runes: lineRunes, length: uint(len(lineRunes))}
		b.contents = append(b.contents, line)
	}

	return nil
}
