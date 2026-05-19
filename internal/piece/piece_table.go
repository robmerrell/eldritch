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
	"fmt"
	"iter"
	"slices"
)

// defaultAddBufferSize is the default capacity for the add buffer. In most editing scenarios
// I can't image the add buffer increasing significantly, but there may be some opportunity for
// optimization here.
const defaultAddBufferSize = 4096

// defaultPieceTableSize is the default capacity for the pieces list. Same idea as the default
// add buffer size
const defaultPieceSize = 512

// defaultLineSize is the default size we use for lines when creating line slices.
const defaultLineSize = 256

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

// Contents traverses all pieces and generates a full document.
func (p *PieceTable) Contents() []rune {
	contents := []rune{}

	for _, piece := range p.pieces {
		contents = append(contents, p.pieceContents(piece)...)
	}

	return contents
}

// pieceContents returns the contents that a piece points at.
func (p *PieceTable) pieceContents(piece *piece) []rune {
	if piece.buffer == bufferTypeOriginal {
		return p.original[piece.start : piece.start+piece.length]
	}

	return p.add[piece.start : piece.start+piece.length]
}

// Lines is a generator that produces lines (index is 0 based).
func (p *PieceTable) Lines(lineStart, lineEnd int) iter.Seq[[]rune] {
	return func(yield func([]rune) bool) {
		lineCount := 0

		line := make([]rune, 0, defaultLineSize)
		for _, piece := range p.pieces {
			for _, rn := range p.pieceContents(piece) {
				if lineCount >= lineStart && lineCount <= lineEnd {
					line = append(line, rn)
				}

				if rn == '\n' {
					if lineCount >= lineStart && lineCount <= lineEnd {
						if !yield(line) {
							return
						}
						line = make([]rune, 0, defaultLineSize)
					}

					lineCount += 1
				}
			}
		}
	}
}

// Len returns the number of runes in the content
func (p *PieceTable) Len() int {
	count := 0

	for _, piece := range p.pieces {
		count += piece.length
	}

	return count
}

// func (p *PieceTable) LineToOffset(line int) int {

// }

// func (p *PieceTable) OffsetToLine(offset int) int {

// }

// Insert inserts a new slice of runes into the piece table. This is a good comment.
func (p *PieceTable) Insert(offset int, input []rune) error {
	if offset < 0 {
		return fmt.Errorf("Invalid offset for insertion: %d", offset)
	}

	addOffset := len(p.add)
	p.add = append(p.add, input...)

	// inserting at the beginning of the document is a special case. Just prepend the new node
	if offset == 0 {
		newPiece := &piece{buffer: bufferTypeAdd, start: addOffset, length: len(input)}
		p.pieces = append([]*piece{newPiece}, p.pieces...)
		return nil
	}

	// find the piece that we want that needs to be split
	pieceIndex, pieceOffset, err := p.pieceIndex(offset)
	if err != nil {
		return err
	}

	// new pieces
	splitPiece := p.pieces[pieceIndex]
	beforePiece := &piece{buffer: splitPiece.buffer, start: splitPiece.start, length: pieceOffset}
	newPiece := &piece{buffer: bufferTypeAdd, start: addOffset, length: len(input)}
	afterPiece := &piece{buffer: splitPiece.buffer, start: pieceOffset, length: splitPiece.length - pieceOffset}

	// update
	p.pieces = slices.Replace(p.pieces, pieceIndex, pieceIndex+1, beforePiece, newPiece, afterPiece)
	return nil
}

// pieceIndex takes an offset and return the index in the pieces slice that include that offset.
// The content offset within that piece is also returned. If an offset greater than total content
// is given return an error.
func (p *PieceTable) pieceIndex(offset int) (int, int, error) {
	acc := 0

	for i, piece := range p.pieces {
		if offset > acc && offset < acc+piece.length {
			return i, offset - acc, nil
		}

		acc += piece.length
	}

	return 0, 0, fmt.Errorf("Invalid offset for insertion: %d", offset)
}
