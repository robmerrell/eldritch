package buffer

import (
	"testing"
)

func TestInsertBacktrackOne(t *testing.T) {
	// add "eld" at the beginning of the buffer
	buffer := NewBuffer()
	buffer.Insert([]rune("e")[0])
	buffer.Insert([]rune("l")[0])
	buffer.Insert([]rune("d")[0])

	// insert - between the l and the d
	buffer.ShiftSelections(SelectionDirectionLeft, 1)
	buffer.Insert([]rune("-")[0])

	if got, want := string(buffer.contents[0].runes), "el-d"; got != want {
		t.Errorf("content=%s, want=%s", got, want)
	}
}

func TestInsertBacktrackToBeginning(t *testing.T) {
	// add "eld" at the beginning of the buffer
	buffer := NewBuffer()
	buffer.Insert([]rune("e")[0])
	buffer.Insert([]rune("l")[0])
	buffer.Insert([]rune("d")[0])

	// insert - between the l and the d
	buffer.ShiftSelections(SelectionDirectionLeft, 1)
	buffer.ShiftSelections(SelectionDirectionLeft, 1)
	buffer.ShiftSelections(SelectionDirectionLeft, 1)
	buffer.Insert([]rune("-")[0])

	if got, want := string(buffer.contents[0].runes), "-eld"; got != want {
		t.Errorf("content=%s, want=%s", got, want)
	}
}

func TestShiftSelection(t *testing.T) {
	buffer := NewBuffer()
}

// func TestInsertAtEndOfLine(t *testing.T) {
// 	input := []rune("s")[0]

// 	buffer := NewBuffer()
// 	buffer.Insert(input)

// 	if got, want := buffer.contents[0].runes[0], input; got != want {
// 		t.Errorf("input=%s, want=%s", string(got), string(want))
// 	}
// }

// func TestInsertMiddleOfString(t *testing.T) {
// 	// add "hi" at the beginnign of the buffer
// 	buffer := NewBuffer()
// 	buffer.Insert([]rune("h")[0])
// 	buffer.Insert([]rune("i")[0])

// 	// insert between the h and the i
// 	buffer.MoveCursorsInDirection(CursorLeft, 1)
// 	buffer.Insert([]rune("-")[0])

// 	if got, want := string(buffer.contents[0].runes), "h-i"; got != want {
// 		t.Errorf("content=%s, want=%s", got, want)
// 	}
// }

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
