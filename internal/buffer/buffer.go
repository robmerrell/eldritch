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
	"os"
)

// errors
var ErrNotFileBackedBuffer = errors.New("Not a file backed buffer")

type Coords struct {
	Line int
	Col  int
}

// Buffer is the backing structure for an editable document.
type Buffer struct {
	contents *PieceTree

	// the file backing the buffer, if backed by a file.
	backingFile *string

	// optional name of the buffer
	name *string

	// has the buffer changed
	dirty bool
}

// NewBuffer creates a new buffer with default options.
func NewBuffer() *Buffer {
	b := &Buffer{
		contents: NewPieceTree([]byte("\n")),
	}

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

// LoadFile loads a file into the buffer
func (b *Buffer) LoadFile(filePath string) error {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	b.backingFile = &filePath
	b.contents = NewPieceTree(bytes)
	return nil
}
