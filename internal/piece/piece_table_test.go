package piece

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromSlice(t *testing.T) {
	pt := FromSlice([]rune("hello\nworld"))
	assert.Equal(t, bufferTypeOriginal, pt.pieces[0].buffer)
	assert.Equal(t, 0, pt.pieces[0].start)
	assert.Equal(t, 12, pt.pieces[0].length)
	assert.Equal(t, "hello\nworld\n", string(pt.original))

	// empty input
	pt = FromSlice([]rune{})
	assert.Equal(t, 0, pt.pieces[0].start)
	assert.Equal(t, 1, pt.pieces[0].length)
	assert.Equal(t, "\n", string(pt.original))

	// input with newline
	pt = FromSlice([]rune("a\n"))
	assert.Equal(t, 0, pt.pieces[0].start)
	assert.Equal(t, 2, pt.pieces[0].length)
	assert.Equal(t, "a\n", string(pt.original))
}

func TestInsert(t *testing.T) {
	pt := FromSlice([]rune("It's a dangerous business going out your door\n"))
	assert.Equal(t, []*piece{{bufferTypeOriginal, 0, 46}}, pt.pieces)
	assert.Equal(t, "It's a dangerous business going out your door\n", string(pt.original))

	// It's a dangerous business, Frodo, going out your door\n
	err := pt.Insert(25, []rune(", Frodo,"))
	assert.NoError(t, err)
	assert.Equal(t, []*piece{
		{bufferTypeOriginal, 0, 25},
		{bufferTypeAdd, 0, 8},
		{bufferTypeOriginal, 25, 21},
	}, pt.pieces)
	assert.Equal(t, ", Frodo,", string(pt.add))

	// It's a dangerous business, FrXodo, going out your door\n
	err = pt.Insert(29, []rune("X"))
	assert.NoError(t, err)
	assert.Equal(t, []*piece{
		{bufferTypeOriginal, 0, 25},
		{bufferTypeAdd, 0, 4},
		{bufferTypeAdd, 8, 1},
		{bufferTypeAdd, 4, 4},
		{bufferTypeOriginal, 25, 21},
	}, pt.pieces)
	assert.Equal(t, ", Frodo,X", string(pt.add))

	// Insert on a piece with 1 character. Caught a bug with this.
	pt = FromSlice([]rune("hello\n"))
	assert.NoError(t, pt.Insert(0, []rune("a")))
	assert.NoError(t, pt.Insert(1, []rune("b")))

	// insert at the beginning of a piece
	pt = FromSlice([]rune("hello\n"))
	assert.NoError(t, pt.Insert(2, []rune("00")))
	assert.NoError(t, pt.Insert(2, []rune("x")))
	assert.Equal(t, []*piece{
		{bufferTypeOriginal, 0, 2},
		{bufferTypeAdd, 2, 1},
		{bufferTypeAdd, 0, 2},
		{bufferTypeOriginal, 2, 4},
	}, pt.pieces)

	// insert at the end of a piece
	pt = FromSlice([]rune("hello\n"))
	assert.NoError(t, pt.Insert(2, []rune("00")))
	assert.NoError(t, pt.Insert(4, []rune("x")))
	assert.Equal(t, []*piece{
		{bufferTypeOriginal, 0, 2},
		{bufferTypeAdd, 0, 2},
		{bufferTypeAdd, 2, 1},
		{bufferTypeOriginal, 2, 4},
	}, pt.pieces)
}

func TestInsertDocBoundaries(t *testing.T) {
	// beginning of doc
	pt := FromSlice([]rune("\n"))
	err := pt.Insert(0, []rune("hello"))
	assert.NoError(t, err)
	assert.Equal(t, []*piece{
		{bufferTypeAdd, 0, 5},
		{bufferTypeOriginal, 0, 1},
	}, pt.pieces)
	assert.Equal(t, "hello", string(pt.add))

	// 1 char in
	pt = FromSlice([]rune("a\n"))
	err = pt.Insert(1, []rune("hello"))
	assert.NoError(t, err)
	assert.Equal(t, []*piece{
		{bufferTypeOriginal, 0, 1},
		{bufferTypeAdd, 0, 5},
		{bufferTypeOriginal, 1, 1},
	}, pt.pieces)
	assert.Equal(t, "hello", string(pt.add))

	// don't allow overwriting final \n
	pt = FromSlice([]rune("hello\n"))
	err = pt.Insert(6, []rune("x"))
}

func TestInsertInvalidOffsets(t *testing.T) {
	pt := FromSlice([]rune("\n"))

	err := pt.Insert(100, []rune("hello"))
	assert.Error(t, err)

	err = pt.Insert(-10, []rune("hello"))
	assert.Error(t, err)
}

func TestDelete(t *testing.T) {
	table := func() *PieceTable {
		pt := FromSlice([]rune(""))
		pt.Insert(0, []rune("Line 1\n"))
		pt.Insert(7, []rune("Line 2\n"))
		pt.Insert(14, []rune("Line 3\n"))
		pt.Insert(21, []rune("Line 4\n"))

		return pt
	}

	// trim the beginning of a piece
	pt := table()
	assert.NoError(t, pt.Delete(0, 3))
	fmt.Println(string(pt.Contents()))
	assert.Equal(t, []*piece{
		{bufferTypeAdd, 4, 3},
		{bufferTypeAdd, 7, 7},
		{bufferTypeAdd, 14, 7},
		{bufferTypeAdd, 21, 7},
		{bufferTypeOriginal, 0, 1},
	}, pt.pieces)

	// trim end of a piece
	pt = table()
	assert.NoError(t, pt.Delete(2, 6))
	fmt.Println(string(pt.Contents()))
	assert.Equal(t, []*piece{
		{bufferTypeAdd, 0, 2},
		{bufferTypeAdd, 7, 7},
		{bufferTypeAdd, 14, 7},
		{bufferTypeAdd, 21, 7},
		{bufferTypeOriginal, 0, 1},
	}, pt.pieces)

	// delete an entire piece
	pt = table()
	assert.NoError(t, pt.Delete(7, 13))
	fmt.Println(string(pt.Contents()))
	assert.Equal(t, []*piece{
		{bufferTypeAdd, 0, 7},
		{bufferTypeAdd, 14, 7},
		{bufferTypeAdd, 21, 7},
		{bufferTypeOriginal, 0, 1},
	}, pt.pieces)

	// delete the middle of a piece
	pt = table()
	assert.NoError(t, pt.Delete(1, 4))
	fmt.Println(string(pt.Contents()))
	assert.Equal(t, []*piece{
		{bufferTypeAdd, 0, 1},
		{bufferTypeAdd, 5, 2},
		{bufferTypeAdd, 7, 7},
		{bufferTypeAdd, 14, 7},
		{bufferTypeAdd, 21, 7},
		{bufferTypeOriginal, 0, 1},
	}, pt.pieces)

	// delete in multiple pieces
	pt = table()
	assert.NoError(t, pt.Delete(3, 17))
	fmt.Println(string(pt.Contents()))
	assert.Equal(t, []*piece{
		{bufferTypeAdd, 0, 3},
		{bufferTypeAdd, 18, 3},
		{bufferTypeAdd, 21, 7},
		{bufferTypeOriginal, 0, 1},
	}, pt.pieces)
}

func TestLines(t *testing.T) {
	pt := FromSlice([]rune("line 1\nline 2\nline 3\n"))
	assert.NoError(t, pt.Insert(20, []rune("\nline 4")))

	// middle of doc
	lines := slices.Collect(pt.Lines(1, 2))
	assert.Len(t, lines, 2)
	assert.Equal(t, "line 2\n", string(lines[0].Runes))
	assert.Equal(t, 7, lines[0].StartOffset)
	assert.Equal(t, "line 3\n", string(lines[1].Runes))
	assert.Equal(t, 14, lines[1].StartOffset)

	// beginning of doc
	lines = slices.Collect(pt.Lines(0, 2))
	assert.Len(t, lines, 3)
	assert.Equal(t, "line 1\n", string(lines[0].Runes))
	assert.Equal(t, 0, lines[0].StartOffset)
	assert.Equal(t, "line 2\n", string(lines[1].Runes))
	assert.Equal(t, 7, lines[1].StartOffset)
	assert.Equal(t, "line 3\n", string(lines[2].Runes))
	assert.Equal(t, 14, lines[2].StartOffset)

	// past end of doc
	lines = slices.Collect(pt.Lines(2, 100))
	assert.Len(t, lines, 2)
	assert.Equal(t, "line 3\n", string(lines[0].Runes))
	assert.Equal(t, 14, lines[0].StartOffset)
	assert.Equal(t, "line 4\n", string(lines[1].Runes))
	assert.Equal(t, 21, lines[1].StartOffset)

	// begin line is negative

	// start line is past the end
}

func TestContents(t *testing.T) {
	pt := FromSlice([]rune("It's a dangerous business going out your door\n"))
	assert.Equal(t, "It's a dangerous business going out your door\n", string(pt.Contents()))

	pt.Insert(25, []rune(", Frodo,"))
	assert.Equal(t, "It's a dangerous business, Frodo, going out your door\n", string(pt.Contents()))
}

func TestLen(t *testing.T) {
	pt := FromSlice([]rune("It's a dangerous business going out your door\n"))
	assert.Equal(t, 46, pt.Len())

	pt.Insert(25, []rune(", Frodo,"))
	assert.Equal(t, 54, pt.Len())
}

func TestLineCount(t *testing.T) {
	pt := FromSlice([]rune("line 1\nline 2\nline 3\n"))
	assert.Equal(t, 3, pt.LineCount())

	assert.NoError(t, pt.Insert(20, []rune("\nline 4")))
	assert.Equal(t, 4, pt.LineCount())
}

func TestCoordsToOffset(t *testing.T) {
	pt := FromSlice([]rune("line 1\nline 2\nline 3\n"))
	assert.NoError(t, pt.Insert(20, []rune("\nline 4")))

	assert.Equal(t, 17, pt.CoordsToOffset(Coords{2, 3}))
	assert.Equal(t, 23, pt.CoordsToOffset(Coords{3, 2}))
	assert.Equal(t, 2, pt.CoordsToOffset(Coords{0, 2}))

	assert.Panics(t, func() {
		pt.CoordsToOffset(Coords{-1, 0})
	})
	assert.Panics(t, func() {
		pt.CoordsToOffset(Coords{100, 100})
	})
}

func TestOffsetToCoords(t *testing.T) {
	pt := FromSlice([]rune("line 1\nline 2\nline 3\n"))
	assert.NoError(t, pt.Insert(20, []rune("\nline 4")))

	assert.Equal(t, Coords{2, 3}, pt.OffsetToCoords(17))
	assert.Equal(t, Coords{3, 2}, pt.OffsetToCoords(23))
	assert.Equal(t, Coords{0, 2}, pt.OffsetToCoords(2))
}
