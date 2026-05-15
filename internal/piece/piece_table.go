// Package piece is a piece table implementation for buffer contents.
//
// For this first pass I'm using a slice for the piece table. I strongly suspect
// this will be enough to get my feet under me and using a linked list in practice
// isn't going to do much for making this any faster. We'll see though.
//
// A future optimization path is to replace the slice with red-black tree that can cache
// some line counts and total character counts per branch.
package piece

import (
	"slices"
)

// defaultAddBufferSize is the default capacity for the add buffer. In most editing scenarios
// I can't image the add buffer increasing significantly, but there may be some opportunity for
// optimization here.
const defaultAddBufferSize = 4096

// defaultPieceTableSize is the default capacity for the pieces list. Same idea as the default
// add buffer size
const defaultPieceSize = 512

// the two types of buffers for the piece table.
type bufferType int

const (
	bufferTypeOriginal bufferType = iota
	bufferTypeAdd
)

// piece stores the individual edit pieces in the table.
type piece struct {
	buffer bufferType
	start  int
	length int
}

// PieceTable holds the edits and contents of a document.
type PieceTable struct {
	original []rune
	add      []rune
	pieces   []*piece
}

// FromSlice create a new piece table using the given input as the original buffer. It ensures that
// the buffer ends in a newline.
func FromSlice(input []rune) *PieceTable {
	// ensure that we always have a newline present
	if len(input) == 0 || input[len(input)-1] != '\n' {
		input = append(input, []rune("\n")...)
	}

	pieces := make([]*piece, 1, defaultPieceSize)
	pieces[0] = &piece{buffer: bufferTypeOriginal, start: 0, length: len(input)}

	return &PieceTable{
		original: input,
		add:      make([]rune, 0, defaultAddBufferSize),
		pieces:   pieces,
	}
}

// Insert inserts a new slice of runes into the piece table. This is a good comment.
func (p *PieceTable) Insert(offset int, input []rune) {
	addOffset := len(p.add)
	p.add = append(p.add, input...)

	// find the piece that we want that needs to be split
	pieceIndex, pieceOffset := p.pieceIndex(offset)
	splitPiece := p.pieces[pieceIndex]

	// new pieces
	beforePiece := &piece{buffer: splitPiece.buffer, start: splitPiece.start, length: pieceOffset}
	newPiece := &piece{buffer: bufferTypeAdd, start: addOffset, length: len(input)}
	afterPiece := &piece{buffer: splitPiece.buffer, start: pieceOffset, length: splitPiece.length - pieceOffset}

	// update
	p.pieces = slices.Replace(p.pieces, pieceIndex, pieceIndex+1, beforePiece, newPiece, afterPiece)
}

// pieceIndex takes an offset and returns the index in the pieces slice that includes that offset.
func (p *PieceTable) pieceIndex(offset int) (int, int) {
	acc := 0

	for i, piece := range p.pieces {
		if offset > acc && offset < acc+piece.length {
			return i, offset - acc
		}

		acc += piece.length
	}

	// this shouldn't be reached.
	return -1, -1
}
