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

func testEndOfDocumentOffset(t *testing.T) {
	buffer := NewBuffer()
	buffer.SetContents("hello\nworld")

	assert.Equal(t, 11, buffer.endOfDocumentOffset())
}

func TestAddSelection(t *testing.T) {
	buffer := NewBuffer()
	buffer.SetContents("hello")
	buffer.AddSelection(2)

	assert.Equal(t, 2, buffer.selections[1].Anchor)
	assert.Equal(t, 2, buffer.selections[1].Head)
}

func TestShiftSelectionsForward(t *testing.T) {
	buffer := NewBuffer()
	buffer.SetContents("hello")
	buffer.AddSelection(1) // at the e
	buffer.AddSelection(4) // at the o
	buffer.AddSelection(5) // at the end \n
	buffer.ShiftSelectionsForward(1, false)

	// primary
	assert.Equal(t, 0, buffer.selections[0].Anchor)
	assert.Equal(t, 1, buffer.selections[0].Head)

	// started at e
	assert.Equal(t, 1, buffer.selections[1].Anchor)
	assert.Equal(t, 2, buffer.selections[1].Head)

	// move from the o to the newline
	assert.Equal(t, 4, buffer.selections[2].Anchor)
	assert.Equal(t, 5, buffer.selections[2].Head)

	// we're at the end of the document, so don't move
	assert.Equal(t, 5, buffer.selections[3].Anchor)
	assert.Equal(t, 5, buffer.selections[3].Head)
}

func testShiftSelectionForwardMultiLine(t *testing.T) {
	buffer := NewBuffer()
	buffer.SetContents("hello\nsecond")

	// end of first line \n
	buffer.selections[0].Anchor = 5
	buffer.selections[0].Head = 5

	// end of the document
	buffer.AddSelection(12)

	buffer.ShiftSelectionsForward(3, true)

	assert.Equal(t, 8, buffer.selections[0].Anchor)
	assert.Equal(t, 8, buffer.selections[0].Head)

	assert.Equal(t, 12, buffer.selections[1].Anchor)
	assert.Equal(t, 12, buffer.selections[1].Head)
}

/*
func assertCollapsedSelection(t *testing.T, sel *Selection, x, y int) {
	t.Helper()

	if sel.AnchorX != x {
		t.Errorf("Anchor x got: %d, expected %d", sel.AnchorX, x)
	}

	if sel.AnchorY != y {
		t.Errorf("Anchor y got: %d, expected %d", sel.AnchorY, y)
	}

	if sel.HeadX != x {
		t.Errorf("Head x got: %d, expected %d", sel.HeadX, x)
	}

	if sel.HeadY != y {
		t.Errorf("Head y got: %d, expected %d", sel.HeadY, y)
	}
}
*/

/*
func TestInsertSelectionPosition(t *testing.T) {
	buffer := NewBuffer()

	buffer.Insert(strRune("h"))
	assertCollapsedSelection(t, buffer.selections[0], 1, 0)

	buffer.Insert(strRune("i"))
	assertCollapsedSelection(t, buffer.selections[0], 2, 0)
}
*/

/*
func TestInsertBacktrackOne(t *testing.T) {
	// add "eld" at the beginning of the buffer
	buffer := NewBuffer()
	buffer.Insert(strRune("e"))
	buffer.Insert(strRune("l"))
	buffer.Insert(strRune("d"))

	// insert - between the l and the d
	buffer.ShiftSelections(SelectionDirectionLeft, 1)
	buffer.Insert(strRune("-"))

	if got, want := string(buffer.contents[0].runes), "el-d"; got != want {
		t.Errorf("content=%s, want=%s", got, want)
	}
}

func TestInsertBacktrackToBeginning(t *testing.T) {
	// add "eld" at the beginning of the buffer
	buffer := NewBuffer()
	buffer.Insert(strRune("e"))
	buffer.Insert(strRune("l"))
	buffer.Insert(strRune("d"))

	// insert - between the l and the d
	buffer.ShiftSelections(SelectionDirectionLeft, 1)
	buffer.ShiftSelections(SelectionDirectionLeft, 1)
	buffer.ShiftSelections(SelectionDirectionLeft, 1)
	buffer.Insert([]rune("-")[0])

	if got, want := string(buffer.contents[0].runes), "-eld"; got != want {
		t.Errorf("content=%s, want=%s", got, want)
	}
}

func setupSelectionBuffer() *Buffer {
	buffer := NewBuffer()
	buffer.SetContents([]rune("hello\nworld\nthis\nis a buffer"))

	return buffer
}
*/

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
func TestShiftSelectionRight(t *testing.T) {
	// buffer := NewBuffer()
	// buffer.SetContents([]rune("hello\nworld\nthis\nis a buffer"))

	// buffer := setupSelectionBuffer()
	// line := buffer.contents[0]
	// lineCount := len(buffer.contents)

	// from 0 moves right
	// buffer.selections[0].SetCollapsed(0, 0)
	// buffer.ShiftSelections(SelectionDirectionRight, 1)
	// assertCollapsedSelection(t, buffer.selections[0], 1, 0)

	// // from the middle
	// buffer.selections[0].SetCollapsed(3, 0)
	// buffer.ShiftSelections(SelectionDirectionRight, 1)
	// assertCollapsedSelection(t, buffer.selections[0], 4, 0)

	// // after the last character moves the cursor past it to empty space
	// buffer.selections[0].SetCollapsed(line.length-1, 0)
	// buffer.ShiftSelections(SelectionDirectionRight, 1)
	// assertCollapsedSelection(t, buffer.selections[0], 5, 0)

	// // at end of line goes to the beginning of the next line
	// buffer.selections[0].SetCollapsed(line.length, 0)
	// buffer.ShiftSelections(SelectionDirectionRight, 1)
	// assertCollapsedSelection(t, buffer.selections[0], 0, 1)

	// // at the end of the last line doesn't move
	// line = buffer.contents[3]
	// buffer.selections[0].SetCollapsed(line.length, lineCount-1)
	// buffer.ShiftSelections(SelectionDirectionRight, 1)
	// assertCollapsedSelection(t, buffer.selections[0], line.length, lineCount-1)
}
*/

/*
func TestShiftSelectionLeft(t *testing.T) {
	buffer := setupSelectionBuffer()

	// from 0, 0 doesn't move
	buffer.selections[0].SetCollapsed(0, 0)
	buffer.ShiftSelections(SelectionDirectionLeft, 1)
	assertCollapsedSelection(t, buffer.selections[0], 0, 0)

	// from the middle
	buffer.selections[0].SetCollapsed(3, 0)
	buffer.ShiftSelections(SelectionDirectionLeft, 1)
	assertCollapsedSelection(t, buffer.selections[0], 2, 0)

	// at beginning of line goes to the end of the next line
	buffer.selections[0].SetCollapsed(0, 2)
	buffer.ShiftSelections(SelectionDirectionLeft, 1)
	newLine := buffer.contents[1]
	assertCollapsedSelection(t, buffer.selections[0], newLine.length, 1)
}
*/

// func TestShiftingSelectionsByCount(t *testing.T) {

// }

// func TestShiftingSelectionsPreserveAnchor(t *testing.T) {

// }

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
