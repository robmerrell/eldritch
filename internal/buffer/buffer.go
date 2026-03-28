package buffer

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// errors
var ErrNotFileBackedBuffer = errors.New("Not a file backed buffer")

// Default number of lines and runes per line to preallocate empty buffers
const DefaultLineCap = 256
const DefaultRuneCap = 256

type line struct {
	runes []rune
}

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
	selections []*Selection

	// a cursor will always be at index 0. This is the main cursor.
	// any cursor beyond the first is a transient cursor.
	// A collection of cursor selections for the
	cursors []*Cursor
}

// NewBuffer creates a new buffer with default options.
func NewBuffer() *Buffer {
	b := &Buffer{
		cursors:    []*Cursor{NewCursor(0, 0)},
		selections: []*Selection{NewSelection(0, 0)},
	}
	b.Clear()

	return b
}

// NewBufferWithFile creates a new file backed buffer.
// func NewBufferWithFile(filePath string) (*Buffer, error) {
// 	buffer := NewBuffer()
// 	err := buffer.LoadFile(filePath)

// 	return buffer, err
// }

// Clear clears the buffer input by reallocating the content container.
func (b *Buffer) Clear() {
	b.contents = make([]line, 1, DefaultLineCap)
	b.contents[0].runes = make([]rune, 0, DefaultRuneCap)
}

// AssignName gives the buffer a name
func (b *Buffer) AssignName(name string) {
	b.name = &name
}

// PrintDbg prints out a line and cursor position to use before I'm actually rendering anything.
func (b *Buffer) PrintDbg(lineNum uint) {
	line := b.contents[lineNum]

	fmt.Println(string(line.runes))

	// draw any selections on the line
	for _, selection := range b.selections {
		selectionOut := make([]string, len(line.runes)+1)
		for i := range selectionOut {
			selectionOut[i] = " "
		}

		if selection.AnchorY == lineNum && selection.HeadY == lineNum {
			selectionOut[selection.AnchorX] = "A"

			if selectionOut[selection.HeadX] != " " {
				selectionOut[selection.HeadX] = "B"
			} else {
				selectionOut[selection.HeadX] = "H"
			}
		}

		fmt.Println(strings.Join(selectionOut, ""))
	}
}

// Insert inserts a rune at all cursor positions. Characters are inserted before the selection.
func (b *Buffer) Insert(input rune) {
	for _, selection := range b.selections {
		// get the beginning of the selection regardless of anchor coords
		x, y := selection.Beginning()
		line := &b.contents[y]
		lineLen := uint(len(line.runes))

		if lineLen == x {
			// at the end of the line, so append
			line.runes = append(line.runes, input)
		} else if lineLen > x {
			// Something already exists there, so insert it
			line.runes = slices.Insert(line.runes, int(x), input)
		}

		b.shiftSelection(selection, SelectionDirectionRight, 1)
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
	case SelectionDirectionRight:
		selection.AnchorX = selection.AnchorX + count
		selection.HeadX = selection.HeadX + count

	case SelectionDirectionLeft:
		selection.AnchorX = selection.AnchorX - count
		selection.HeadX = selection.HeadX - count
	}
}

// LoadFile loads a file into the buffer
// func (b *Buffer) LoadFile(filePath string) error {
// 	contents, err := os.ReadFile(filePath)
// 	if err != nil {
// 		return err
// 	}

// 	b.backingFile = &filePath
// 	b.lines = strings.Split(string(contents), "\n")

// 	// drop the last empty newline for our buffer represenation
// 	// because newlines are automatically added.
// 	if b.lines[len(b.lines)-1] == "" {
// 		b.lines = b.lines[:len(b.lines)-1]
// 	}

// 	return nil
// }

// Save writes the buffer to disk
// func (b *Buffer) Save() error {
// 	if b.backingFile == nil {
// 		return ErrNotFileBackedBuffer
// 	}

// 	// TODO: do the permissions here overwrite existing or is it only for new files?
// 	err := os.WriteFile(*b.backingFile, b.asBytes(), 0666)
// 	return err
// }

// asBytes convert the buffer contents to bytes
// func (b *Buffer) asBytes() []byte {
// 	return []byte(strings.Join(b.lines, "\n"))
// }
