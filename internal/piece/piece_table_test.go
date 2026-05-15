package piece

import (
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

func TestInsertPieces(t *testing.T) {
	pt := FromSlice([]rune("It's a dangerous business going out your door\n"))
	assert.Equal(t, []*piece{{bufferTypeOriginal, 0, 46}}, pt.pieces)

	// It's a dangerous business, Frodo, going out your door\n
	pt.Insert(25, []rune(", Frodo,"))
	assert.Equal(t, []*piece{
		{bufferTypeOriginal, 0, 25},
		{bufferTypeAdd, 0, 8},
		{bufferTypeOriginal, 25, 21},
	}, pt.pieces)

	// It's a dangerous business, Froodo, going out your door\n
	pt.Insert(29, []rune("o"))
	assert.Equal(t, []*piece{
		{bufferTypeOriginal, 0, 25},
		{bufferTypeAdd, 0, 4},
		{bufferTypeAdd, 8, 1},
		{bufferTypeAdd, 4, 4},
		{bufferTypeOriginal, 25, 21},
	}, pt.pieces)
}
