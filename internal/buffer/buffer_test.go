package buffer

import (
	"testing"
)

func strRune(input string) rune {
	return []rune(input)[0]
}

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

func TestInsertSelectionPosition(t *testing.T) {
	buffer := NewBuffer()

	buffer.Insert(strRune("h"))
	assertCollapsedSelection(t, buffer.selections[0], 1, 0)

	buffer.Insert(strRune("i"))
	assertCollapsedSelection(t, buffer.selections[0], 2, 0)
}

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

func TestShiftSelectionRight(t *testing.T) {
	buffer := setupSelectionBuffer()
	line := buffer.contents[0]
	lineCount := len(buffer.contents)

	// from 0 moves right
	buffer.selections[0].SetCollapsed(0, 0)
	buffer.ShiftSelections(SelectionDirectionRight, 1)
	assertCollapsedSelection(t, buffer.selections[0], 1, 0)

	// from the middle
	buffer.selections[0].SetCollapsed(3, 0)
	buffer.ShiftSelections(SelectionDirectionRight, 1)
	assertCollapsedSelection(t, buffer.selections[0], 4, 0)

	// after the last character moves the cursor past it to empty space
	buffer.selections[0].SetCollapsed(line.length-1, 0)
	buffer.ShiftSelections(SelectionDirectionRight, 1)
	assertCollapsedSelection(t, buffer.selections[0], 5, 0)

	// at end of line goes to the beginning of the next line
	buffer.selections[0].SetCollapsed(line.length, 0)
	buffer.ShiftSelections(SelectionDirectionRight, 1)
	assertCollapsedSelection(t, buffer.selections[0], 0, 1)

	// at the end of the last line doesn't move
	line = buffer.contents[3]
	buffer.selections[0].SetCollapsed(line.length, lineCount-1)
	buffer.ShiftSelections(SelectionDirectionRight, 1)
	assertCollapsedSelection(t, buffer.selections[0], line.length, lineCount-1)
}

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

// func TestShiftingSelectionsByCount(t *testing.T) {

// }

// func TestShiftingSelectionsPreserveAnchor(t *testing.T) {

// }

/*
func TestBufferContentsForRendering(t *testing.T) {
	buffer, err := NewBufferWithFile("testdata/render.txt")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assertLines := func(start, height, width int, assertion string) {
		content := []string{}
		for line := range buffer.ContentsForRendering(start, height, width) {
			lineStr := fmt.Sprintf("%d %s", line.RenderedRows, line.LineContents)
			content = append(content, lineStr)
		}

		strContent := strings.Join(content, "\n")
		if strContent != assertion {
			t.Errorf("got:\n%s \nwanted:\n%s", strContent, assertion)
		}
	}

	// starting at 0
	assertLines(0, 3, 100,
		"1 This is a file used to test the renderer.\n"+
			"1 We want to\n"+
			"1 make sure that it can")

	// starting at a non zero line (like we scrolled)
	assertLines(1, 2, 100,
		"1 We want to\n"+
			"1 make sure that it can")

	// more lines to display than there is content
	assertLines(0, 100, 100,
		"1 This is a file used to test the renderer.\n"+
			"1 We want to\n"+
			"1 make sure that it can\n"+
			"1 handle showing partial content, wrapping lines, etc.\n"+
			"1 ")

	// Test last line
	assertLines(3, 100, 100,
		"1 handle showing partial content, wrapping lines, etc.\n"+
			"1 ")

	// Test line wrap
	assertLines(0, 2, 15,
		"4 This is a file \n"+
			"used to test th\n"+
			"e renderer.\n"+
			"1 We want to")

}
*/

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
