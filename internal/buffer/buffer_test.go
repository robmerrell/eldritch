package buffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferWithBadFile(t *testing.T) {
	_, err := NewBufferWithFile("badfile")
	assert.NotNil(t, err)
}

func TestBufferWithFile(t *testing.T) {
	buffer, err := NewBufferWithFile("testdata/file.txt")
	assert.Nil(t, err)

	contents, err := buffer.contents.Contents()
	assert.NoError(t, err)

	assert.Equal(t, "testdata/file.txt", *buffer.backingFile)
	assert.Equal(t, []byte("line 1\nline 2\nline 3\n"), contents)
}
