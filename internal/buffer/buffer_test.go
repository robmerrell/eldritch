package buffer

import (
	"testing"
)

func strRune(input string) rune {
	return []rune(input)[0]
}

func assertCollapsedSelection(t *testing.T, sel *Selection, x, y uint) {
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
	buffer.SetContents([]rune("hello\nworld\nthis is\na buffer"))

	return buffer
}

func TestShiftSelectionUp(t *testing.T) {
	// from 0 removed

	// from 1 goes to 0
}

func TestShiftSelectionDown(t *testing.T) {
	// from 0 goes down

	// from 1 (end) removed
}

func TestShiftSelectionRight(t *testing.T) {
	buffer := setupSelectionBuffer()

	/*
		// from 0 moves right
		buffer.selections[0].SetCollapsed(0, 0)
		buffer.ShiftSelections(SelectionDirectionRight, 1)
		assertCollapsedSelection(t, buffer.selections[0], 1, 0)

		// from the middle
		buffer.selections[0].SetCollapsed(3, 0)
		buffer.ShiftSelections(SelectionDirectionRight, 1)
		assertCollapsedSelection(t, buffer.selections[0], 4, 0)

		// after the last character moves the cursor past it to empty space
		buffer.selections[0].SetCollapsed(5, 0)
		buffer.ShiftSelections(SelectionDirectionRight, 1)
		assertCollapsedSelection(t, buffer.selections[0], 6, 0)
	*/

	// at end of line goes to the beginning of the next line
	buffer.selections[0].SetCollapsed(5, 0)
	buffer.ShiftSelections(SelectionDirectionRight, 1)
	assertCollapsedSelection(t, buffer.selections[0], 0, 1)

	// at the end of the last line doesn't move
}

func TestShiftSelectionLeft(t *testing.T) {
	// buffer := setupSelectionBuffer()

	// from 0 previous line

	// from the middle

	// at the last character

	// at end of line
}

// func TestBufferWithBadFile(t *testing.T) {
// 	_, err := NewBufferWithFile("badfile")
// 	if err == nil {
// 		t.Fatalf("Expected error, got none")
// 	}
// }

// func TestBufferWithFile(t *testing.T) {
// 	buffer, err := NewBufferWithFile("testdata/file.txt")
// 	if err != nil {
// 		t.Fatalf("Unexpected error: %v", err)
// 	}

// 	for i, a := range buffer.lines {
// 		fmt.Printf("%d %s\n", i, a)
// 	}

// 	// backing file
// 	if got, want := (*buffer.backingFile), "testdata/file.txt"; got != want {
// 		t.Errorf("file=%s, want=%s", got, want)
// 	}

// 	// line length
// 	if got, want := len(buffer.lines), 3; got != want {
// 		t.Errorf("length=%d, want=%d", got, want)
// 	}

// 	// file contents
// 	if got, want := (buffer.lines[0]), "line 1"; got != want {
// 		t.Errorf("line=%s, want=%s", got, want)
// 	}
// 	if got, want := (buffer.lines[1]), "line 2"; got != want {
// 		t.Errorf("line=%s, want=%s", got, want)
// 	}
// 	if got, want := (buffer.lines[2]), "line 3"; got != want {
// 		t.Errorf("line=%s, want=%s", got, want)
// 	}
// }

// func TestClear(t *testing.T) {
// 	buffer := &Buffer{}
// 	buffer.Clear()

// 	if got, want := len(buffer.lines), DefaultLineSize; got != want {
// 		t.Fatalf("length=%d, want=%d", got, want)
// 	}
// }

// func TestSaveNotFileBacked(t *testing.T) {
// 	buffer := NewBuffer()

// 	if err := buffer.Save(); !errors.Is(err, ErrNotFileBackedBuffer) {
// 		t.Fatalf("error=%v, want=not buffer backed", err)
// 	}
// }

// func TestSave(t *testing.T) {
// 	tempDir := t.TempDir()
// 	tempFile := filepath.Join(tempDir, "input.txt")
// 	os.WriteFile(tempFile, []byte{}, 0666)

// 	buffer, err := NewBufferWithFile(tempFile)
// 	if err != nil {
// 		t.Fatalf("Unexpected error: %v", err)
// 	}

// 	if err := buffer.Save(); err != nil {
// 		t.Fatalf("Unexpected error: %v", err)
// 	}

// 	t.Error("ok")

// 	contents, _ := os.ReadFile(tempFile)
// 	fmt.Println(string(contents))
// }
