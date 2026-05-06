package buffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetContents(t *testing.T) {
	buffer := NewBuffer()
	buffer.SetContents("hello\nworld\nthis\nis a buffer")

	assert.Equal(t, "hello\n", string(buffer.contents[0].runes))
	assert.Equal(t, "world\n", string(buffer.contents[1].runes))
	assert.Equal(t, "this\n", string(buffer.contents[2].runes))
	assert.Equal(t, "is a buffer\n", string(buffer.contents[3].runes))
}

func TestAddOpenSelection(t *testing.T) {
	buffer := NewBuffer()
	buffer.SetContents("hello")
	buffer.AddOpenSelection(0, 2, 0, 3)

	assert.Equal(t, 0, buffer.selections[1].HeadRow)
	assert.Equal(t, 2, buffer.selections[1].HeadCol)
	assert.Equal(t, 0, buffer.selections[1].AnchorRow)
	assert.Equal(t, 3, buffer.selections[1].AnchorCol)
}

func TestSortSelections(t *testing.T) {
	buffer := NewBuffer()
	buffer.SetContents("hello\nsecond\nthird\nfourth")
	buffer.selections[0].SetCollapsed(1, 3)
	buffer.AddCollapsedSelection(3, 5)
	buffer.AddCollapsedSelection(3, 4)
	buffer.AddCollapsedSelection(0, 2)

	assert.Equal(t, 0, buffer.selections[0].HeadRow)
	assert.Equal(t, 2, buffer.selections[0].HeadCol)

	assert.Equal(t, 1, buffer.selections[1].HeadRow)
	assert.Equal(t, 3, buffer.selections[1].HeadCol)

	assert.Equal(t, 3, buffer.selections[2].HeadRow)
	assert.Equal(t, 4, buffer.selections[2].HeadCol)

	assert.Equal(t, 3, buffer.selections[3].HeadRow)
	assert.Equal(t, 5, buffer.selections[3].HeadCol)
}

func TestAddCollapsedSelection(t *testing.T) {
	buffer := NewBuffer()
	buffer.SetContents("hello")
	buffer.AddCollapsedSelection(0, 2)

	assert.Equal(t, 0, buffer.selections[1].HeadRow)
	assert.Equal(t, 2, buffer.selections[1].HeadCol)
	assert.Equal(t, 0, buffer.selections[1].AnchorRow)
	assert.Equal(t, 2, buffer.selections[1].AnchorCol)
}

func TestShiftSelectionsForward(t *testing.T) {
	buffer := NewBuffer()
	buffer.SetContents("hello")
	buffer.AddCollapsedSelection(0, 1) // at the e
	buffer.AddCollapsedSelection(0, 4) // at the o
	buffer.AddCollapsedSelection(0, 5) // at the end \n
	buffer.ShiftSelectionsForward(1, false)

	// primary
	assert.Equal(t, 0, buffer.selections[0].AnchorCol)
	assert.Equal(t, 1, buffer.selections[0].HeadCol)

	// started at e
	assert.Equal(t, 1, buffer.selections[1].AnchorCol)
	assert.Equal(t, 2, buffer.selections[1].HeadCol)

	// move from the o to the newline
	assert.Equal(t, 4, buffer.selections[2].AnchorCol)
	assert.Equal(t, 5, buffer.selections[2].HeadCol)

	// we're at the end of the document, so don't move
	assert.Equal(t, 5, buffer.selections[3].AnchorCol)
	assert.Equal(t, 5, buffer.selections[3].HeadCol)
}

func TestShiftSelectionsForwardMultiLine(t *testing.T) {
	buffer := NewBuffer()
	buffer.SetContents("hello\nsecond")

	// end of first line \n
	buffer.selections[0].AnchorCol = 5
	buffer.selections[0].HeadCol = 5

	// end of the document
	buffer.AddCollapsedSelection(1, 6)

	buffer.ShiftSelectionsForward(3, true)

	assert.Equal(t, 1, buffer.selections[0].AnchorRow)
	assert.Equal(t, 2, buffer.selections[0].AnchorCol)
	assert.Equal(t, 1, buffer.selections[0].HeadRow)
	assert.Equal(t, 2, buffer.selections[0].HeadCol)

	assert.Equal(t, 1, buffer.selections[1].AnchorRow)
	assert.Equal(t, 6, buffer.selections[1].AnchorCol)
	assert.Equal(t, 1, buffer.selections[1].HeadRow)
	assert.Equal(t, 6, buffer.selections[1].HeadCol)
}

func TestShiftSelectionsBackward(t *testing.T) {
	buffer := NewBuffer()
	buffer.SetContents("hello")
	buffer.AddCollapsedSelection(0, 4) // at the o
	buffer.AddCollapsedSelection(0, 5) // at the end \n
	buffer.ShiftSelectionsBackward(1, false)

	// primary beginning of document, don't move
	assert.Equal(t, 0, buffer.selections[0].AnchorCol)
	assert.Equal(t, 0, buffer.selections[0].HeadCol)

	// started at o
	assert.Equal(t, 4, buffer.selections[1].AnchorCol)
	assert.Equal(t, 3, buffer.selections[1].HeadCol)

	// move from the newline to o
	assert.Equal(t, 5, buffer.selections[2].AnchorCol)
	assert.Equal(t, 4, buffer.selections[2].HeadCol)
}

func TestShiftSelectionsBackwardMultiLine(t *testing.T) {
	buffer := NewBuffer()
	buffer.SetContents("hello\nsecond")

	// beginning of the second line
	buffer.selections[0].AnchorRow = 1
	buffer.selections[0].HeadRow = 1
	buffer.selections[0].AnchorCol = 0
	buffer.selections[0].HeadCol = 0

	buffer.ShiftSelectionsBackward(3, true)

	assert.Equal(t, 0, buffer.selections[0].AnchorRow)
	assert.Equal(t, 3, buffer.selections[0].AnchorCol)
	assert.Equal(t, 0, buffer.selections[0].HeadRow)
	assert.Equal(t, 3, buffer.selections[0].HeadCol)
}

func TestShiftSelectionUp(t *testing.T) {
	buffer := NewBuffer()
	buffer.SetContents("hello\nsecond\nthird\nfourth")

	// from 0 doesn't go up
	buffer.selections[0].SetCollapsed(0, 3)
	buffer.ShiftSelectionsUp(1, true)
	assert.Equal(t, 0, buffer.selections[0].HeadRow)
	assert.Equal(t, 3, buffer.selections[0].HeadCol)

	// from 2 goes up
	buffer.selections[0].SetCollapsed(2, 3)
	buffer.ShiftSelectionsUp(1, true)
	assert.Equal(t, 1, buffer.selections[0].HeadRow)
	assert.Equal(t, 3, buffer.selections[0].HeadCol)

	// move up to end of line if preferred column is too big
	buffer.selections[0].SetCollapsed(1, 6)
	buffer.ShiftSelectionsUp(1, true)
	assert.Equal(t, 0, buffer.selections[0].HeadRow)
	assert.Equal(t, 5, buffer.selections[0].HeadCol)

	// move up multiple lines
	buffer.selections[0].SetCollapsed(3, 2)
	buffer.ShiftSelectionsUp(2, true)
	assert.Equal(t, 1, buffer.selections[0].HeadRow)
	assert.Equal(t, 2, buffer.selections[0].HeadCol)
}

func TestShiftSelectionDown(t *testing.T) {
	buffer := NewBuffer()
	buffer.SetContents("hello\nsecond\nthird\nfourth")

	// from 0 goes down
	buffer.selections[0].SetCollapsed(0, 3)
	buffer.ShiftSelectionsDown(1, true)
	assert.Equal(t, 1, buffer.selections[0].HeadRow)
	assert.Equal(t, 3, buffer.selections[0].HeadCol)

	// move down, but preferred column is bigger than next line
	buffer.selections[0].SetCollapsed(1, 6)
	buffer.ShiftSelectionsDown(1, true)
	assert.Equal(t, 2, buffer.selections[0].HeadRow)
	assert.Equal(t, 5, buffer.selections[0].HeadCol)

	// from last line don't go down
	buffer.selections[0].SetCollapsed(3, 1)
	buffer.ShiftSelectionsDown(1, true)
	assert.Equal(t, 3, buffer.selections[0].HeadRow)
	assert.Equal(t, 1, buffer.selections[0].HeadCol)

	// move down multiple lines
	buffer.selections[0].SetCollapsed(0, 3)
	buffer.ShiftSelectionsDown(2, true)
	assert.Equal(t, 2, buffer.selections[0].HeadRow)
	assert.Equal(t, 3, buffer.selections[0].HeadCol)
}

func TestSelectLine(t *testing.T) {
	buffer := NewBuffer()
	buffer.SetContents("hello\nsecond\nthird\nfourth")

	buffer.selections[0].SetCollapsed(1, 2)
	buffer.SelectLine()

	assert.Equal(t, 1, buffer.selections[0].AnchorRow)
	assert.Equal(t, 0, buffer.selections[0].AnchorCol)
	assert.Equal(t, 1, buffer.selections[0].HeadRow)
	assert.Equal(t, 6, buffer.selections[0].HeadCol)
}

func TestInsert(t *testing.T) {
	buffer := NewBuffer()
	buffer.SetContents("this is a test\n")
	buffer.selections[0].SetCollapsed(0, 0)
	buffer.AddCollapsedSelection(0, 2)
	buffer.AddCollapsedSelection(0, 10)

	buffer.Insert('x')

	assert.Equal(t, "xthxis is a xtest\n", string(buffer.contents[0].runes))
}

func TestInsertNewLine(t *testing.T) {
	buffer := NewBuffer()
	buffer.SetContents("hello\nworld\n")
	buffer.selections[0].SetCollapsed(0, 2)
	buffer.AddCollapsedSelection(1, 2)

	buffer.InsertNewLine()

	assert.Equal(t, "he\n", string(buffer.contents[0].runes))
	assert.Equal(t, "llo\n", string(buffer.contents[1].runes))
	assert.Equal(t, "wo\n", string(buffer.contents[2].runes))
	assert.Equal(t, "rld\n", string(buffer.contents[3].runes))
}

func TestBufferWithBadFile(t *testing.T) {
	_, err := NewBufferWithFile("badfile")
	assert.NotNil(t, err)
}

func TestBufferWithFile(t *testing.T) {
	buffer, err := NewBufferWithFile("testdata/file.txt")
	assert.Nil(t, err)

	assert.Equal(t, "testdata/file.txt", *buffer.backingFile)
	assert.Equal(t, 3, len(buffer.contents))
	assert.Equal(t, "line 1\n", string(buffer.contents[0].runes))
	assert.Equal(t, "line 2\n", string(buffer.contents[1].runes))
	assert.Equal(t, "line 3\n", string(buffer.contents[2].runes))
}

func TestClear(t *testing.T) {
	buffer := &Buffer{}
	buffer.Clear()
	assert.Equal(t, DefaultLineCap, cap(buffer.contents))
}
