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
	"encoding/json"
	"fmt"
	"iter"
	"os"
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

// WriteJSON writes the piece table out to JSON. Helpful for snapshotting things for writing
// tests and debugging.
func (pt *PieceTable) writeJSON(path string) error {
	type pieceOut struct {
		Buffer string `json:"buffer"`
		Start  int    `json:"start"`
		Length int    `json:"length"`
	}
	out := struct {
		Original string     `json:"original"`
		Add      string     `json:"add"`
		Pieces   []pieceOut `json:"pieces"`
	}{Original: string(pt.original), Add: string(pt.add)}

	for _, p := range pt.pieces {
		buf := "original"
		if p.buffer == bufferTypeAdd {
			buf = "add"
		}
		out.Pieces = append(out.Pieces, pieceOut{buf, p.start, p.length})
	}

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
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

type Line struct {
	Runes       []rune
	StartOffset int
}

// Lines is a generator that produces lines (index is 0 based).
func (p *PieceTable) Lines(lineStart, lineEnd int) iter.Seq[*Line] {
	return func(yield func(*Line) bool) {
		lineCount := 0
		offset := 0

		line := make([]rune, 0, defaultLineSize)
		for _, piece := range p.pieces {
			for _, rn := range p.pieceContents(piece) {
				offset += 1

				if lineCount >= lineStart && lineCount <= lineEnd {
					line = append(line, rn)
				}

				if rn == '\n' {
					if lineCount >= lineStart && lineCount <= lineEnd {
						if !yield(&Line{Runes: line, StartOffset: offset - len(line)}) {
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

// LineCount returns the number of lines in the document.
func (p *PieceTable) LineCount() int {
	lineCount := 0

	for _, piece := range p.pieces {
		for _, rn := range p.pieceContents(piece) {
			if rn == '\n' {
				lineCount += 1
			}
		}
	}

	return lineCount
}

// Coords is used to represent x,y coordinate pairs in the contents.
type Coords struct {
	Row int
	Col int
}

// CoordsToOffset converts coordinates to the offset in the document. Crashes on invalid coordinates. I
// think this function should fail very noisly since the knock-on effect could be really bad.
func (p *PieceTable) CoordsToOffset(coords Coords) int {
	// should only ever be one line
	line := slices.Collect(p.Lines(coords.Row, coords.Row))[0]
	return line.StartOffset + coords.Col
}

// OffsetToCoords converts an offset to coordinates.
func (p *PieceTable) OffsetToCoords(offset int) Coords {
	lineCount := 0
	offsetAcc := 0 // rolling offset count for comparison
	col := 0

	for _, piece := range p.pieces {
		for _, rn := range p.pieceContents(piece) {
			if offsetAcc == offset {
				return Coords{Row: lineCount, Col: col}
			}

			if rn == '\n' {
				lineCount += 1
				col = 0
			} else {
				col += 1
			}

			offsetAcc += 1
		}
	}

	return Coords{Row: lineCount, Col: col}
}

// Insert inserts a new slice of runes into the piece table. This is a good comment.
func (p *PieceTable) Insert(offset int, input []rune) error {
	if offset < 0 {
		return fmt.Errorf("Invalid negative offset: %d", offset)
	}

	addOffset := len(p.add)
	p.add = append(p.add, input...)

	// inserting at the beginning of the document is a special case. Just prepend the new node
	if offset == 0 {
		newPiece := &piece{buffer: bufferTypeAdd, start: addOffset, length: len(input)}
		p.pieces = append([]*piece{newPiece}, p.pieces...)
		return nil
	}

	// find the piece that needs to be split
	pieceIndex, pieceOffset, err := p.pieceIndex(offset)
	if err != nil {
		return err
	}

	// new pieces
	splitPiece := p.pieces[pieceIndex]
	beforePiece := &piece{buffer: splitPiece.buffer, start: splitPiece.start, length: pieceOffset}
	newPiece := &piece{buffer: bufferTypeAdd, start: addOffset, length: len(input)}
	afterPiece := &piece{buffer: splitPiece.buffer, start: splitPiece.start + pieceOffset, length: splitPiece.length - pieceOffset}

	splicePieces := []*piece{beforePiece, newPiece}
	if afterPiece.length > 0 {
		splicePieces = append(splicePieces, afterPiece)
	}

	p.pieces = slices.Replace(p.pieces, pieceIndex, pieceIndex+1, splicePieces...)
	p.writeJSON("/tmp/out.json")
	return nil
}

// Delete deletes from the pieace table between start and end offsets inclusive.
func (p *PieceTable) Delete(start, end int) error {
	// collect pieces for deleting
	acc := 0

	// track what indecies to delete we start with -1 because we use it as a check to make sure we
	// have set the deleteFrom on the first pass through.
	deleteFrom := -1
	deleteTo := -1

	// the piece that needs to be added and its offset after we're done ranging
	var pieceToAdd *piece
	addOffset := -1

	// I need access to the piece struct inside of the loop. Gross.
	for i, bufferPiece := range p.pieces {
		nextAcc := acc + bufferPiece.length

		// determine if the piece is affected by the deletion.
		deleteInPiece := rangesOverlap(acc, nextAcc-1, start, end)
		deleteFirstRune := acc >= start && acc <= end
		deleteLastRune := nextAcc-1 >= start && nextAcc-1 <= end

		if deleteInPiece {
			// splitting pieces and deleting pieces can't happen in the same operation

			// split the piece
			// piece1*piece2*piece3
			//         ^ ^
			// piece 2 needs to be split into 2
			if !deleteFirstRune && !deleteLastRune {
				// modify the piece to the right and create a new piece to the left
				pieceToAdd = &piece{buffer: bufferPiece.buffer, start: bufferPiece.start, length: start - acc}
				addOffset = i

				diff := end - acc + 1
				bufferPiece.start += diff
				bufferPiece.length -= diff
			} else {
				// delete the entire piece
				// piece1*piece2*piece3
				//     ^          ^
				// all of piece 2 would need to be deleted
				if deleteFirstRune && deleteLastRune {
					if deleteFrom == -1 {
						deleteFrom = i
					}

					deleteTo = i
				}

				// trim from the end of the piece
				// piece1*piece2*piece3
				//         ^   ^
				// the end of piece 2 would need to be deleted
				if !deleteFirstRune && deleteLastRune {
					bufferPiece.length = start - acc
				}

				// trim from the beginning of the piece
				// piece1*piece2*piece3
				//        ^  ^
				// the beginning of piece 2 would need to be deleted
				if deleteFirstRune && !deleteLastRune {
					diff := end - acc + 1
					bufferPiece.start += diff
					bufferPiece.length -= diff
				}
			}
		}

		acc = nextAcc

		// finished with piece ranges
		if end < acc {
			break
		}
	}

	// remove the deleted pieces
	if deleteFrom != -1 {
		p.pieces = append(p.pieces[:deleteFrom], p.pieces[deleteTo+1:]...)
	}

	// insert the split piece
	if pieceToAdd != nil {
		p.pieces = slices.Insert(p.pieces, addOffset, pieceToAdd)
	}

	p.writeJSON("/tmp/out.json")
	return nil
}

// pieceIndex takes an offset and return the index in the pieces slice that include that offset.
// The content offset within that piece is also returned. If an offset greater than total content
// is given return an error.
func (p *PieceTable) pieceIndex(offset int) (int, int, error) {
	acc := 0

	for i, piece := range p.pieces {
		if offset > acc && offset <= acc+piece.length {
			return i, offset - acc, nil
		}

		acc += piece.length
	}

	return 0, 0, fmt.Errorf("Invalid offset for insertion: %d", offset)
}

// rangesOverlap tests if there is any overlap between the two given ranges.
func rangesOverlap(x1, x2, y1, y2 int) bool {
	return max(x1, y1) <= min(x2, y2)
}
