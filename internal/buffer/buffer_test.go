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

/*
func TestShiftSelectionUp(t *testing.T) {
	buffer := setupSelectionBuffer()

	// from 0 doesn't go up
	buffer.selections[0].SetCollapsed(3, 0)
	buffer.ShiftSelections(SelectionDirectionUp, 1)
	assertCollapsedSelection(t, buffer.selections[0], 3, 0)

	// from 1 goes up
	buffer.selections[0].SetCollapsed(3, 1)
	buffer.ShiftSelections(SelectionDirectionUp, 1)
	assertCollapsedSelection(t, buffer.selections[0], 3, 0)

	// move up at end of line to a shorter line
	line := buffer.contents[3]
	buffer.selections[0].SetCollapsed(line.length, 3)
	buffer.ShiftSelections(SelectionDirectionUp, 1)
	assertCollapsedSelection(t, buffer.selections[0], 4, 2)
}

func TestShiftSelectionDown(t *testing.T) {
	buffer := setupSelectionBuffer()
	lineCount := len(buffer.contents)

	// from 0 goes down
	buffer.selections[0].SetCollapsed(0, 0)
	buffer.ShiftSelections(SelectionDirectionDown, 1)
	assertCollapsedSelection(t, buffer.selections[0], 0, 1)

	// move down at end of line, but move to shorter line
	line := buffer.contents[1]
	buffer.selections[0].SetCollapsed(line.length, 1)
	buffer.ShiftSelections(SelectionDirectionDown, 1)
	assertCollapsedSelection(t, buffer.selections[0], 4, 2)

	// move down, but don't move X if not at the end of the line
	buffer.selections[0].SetCollapsed(0, 1)
	buffer.ShiftSelections(SelectionDirectionDown, 1)
	assertCollapsedSelection(t, buffer.selections[0], 0, 2)

	// from last line don't go down
	buffer.selections[0].SetCollapsed(0, lineCount-1)
	buffer.ShiftSelections(SelectionDirectionDown, 1)
	assertCollapsedSelection(t, buffer.selections[0], 0, lineCount-1)
}
*/

/*
func TestBufferWithBadFile(t *testing.T) {
	_, err := NewBufferWithFile("badfile")
	if err == nil {
		t.Fatalf("Expected error, got none")
	}
}

func TestBufferWithFile(t *testing.T) {
	buffer, err := NewBufferWithFile("testdata/file.txt")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// backing file
	if got, want := (*buffer.backingFile), "testdata/file.txt"; got != want {
		t.Errorf("file=%s, want=%s", got, want)
	}

	// line length
	if got, want := len(buffer.contents), 3; got != want {
		t.Errorf("length=%d, want=%d", got, want)
	}

	// file contents
	if got, want := string(buffer.contents[0].runes), "line 1"; got != want {
		t.Errorf("line=%s, want=%s", got, want)
	}
	if got, want := string(buffer.contents[1].runes), "line 2"; got != want {
		t.Errorf("line=%s, want=%s", got, want)
	}
	if got, want := string(buffer.contents[2].runes), "line 3"; got != want {
		t.Errorf("line=%s, want=%s", got, want)
	}
}

func TestClear(t *testing.T) {
	buffer := &Buffer{}
	buffer.Clear()

	if got, want := cap(buffer.contents), DefaultLineCap; got != want {
		t.Fatalf("cap=%d, want=%d", got, want)
	}
}
*/
