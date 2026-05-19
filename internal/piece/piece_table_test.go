package piece

import (
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

	// insert at the beginning of a piece

	// insert at the end of a piece
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

func TestLines(t *testing.T) {
	pt := FromSlice([]rune("line 1\nline 2\nline 3\n"))
	assert.NoError(t, pt.Insert(20, []rune("\nline 4")))

	// middle of doc
	lines := slices.Collect(pt.Lines(1, 2))
	assert.Len(t, lines, 2)
	assert.Equal(t, "line 2\n", string(lines[0]))
	assert.Equal(t, "line 3\n", string(lines[1]))

	// beginning of doc
	lines = slices.Collect(pt.Lines(0, 2))
	assert.Len(t, lines, 3)
	assert.Equal(t, "line 1\n", string(lines[0]))
	assert.Equal(t, "line 2\n", string(lines[1]))
	assert.Equal(t, "line 3\n", string(lines[2]))

	// past end of doc
	lines = slices.Collect(pt.Lines(2, 100))
	assert.Len(t, lines, 2)
	assert.Equal(t, "line 3\n", string(lines[0]))
	assert.Equal(t, "line 4\n", string(lines[1]))
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
