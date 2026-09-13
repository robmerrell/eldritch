package buffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Generates a realistic editing scenario where we have a 7 lines in a document "one\ntwo\n" etc.
// Nodes should be in order. Hopefully this can survive all the changes I'm making to the tree...
// The tree structure should look like:
//
//	    4
//	   / \
//	  2   6
//	 /\   /\
//	1 3  5 7
func pieceTreeFixture() *PieceTree {
	// buffer setup
	p := NewPieceTree([]byte("four\n"))
	p.addBuffer = append(p.addBuffer, []byte("one\n")...)
	p.addBuffer = append(p.addBuffer, []byte("two\n")...)
	p.addBuffer = append(p.addBuffer, []byte("three\n")...)
	p.addBuffer = append(p.addBuffer, []byte("five\n")...)
	p.addBuffer = append(p.addBuffer, []byte("six\n")...)
	p.addBuffer = append(p.addBuffer, []byte("seven\n")...)

	p.root.leftSubtreeLen = 14
	p.root.leftSubtreeNewlineCount = 3

	// line starts
	p.addLineStarts = append(p.addLineStarts, 4)
	p.addLineStarts = append(p.addLineStarts, 8)
	p.addLineStarts = append(p.addLineStarts, 14)
	p.addLineStarts = append(p.addLineStarts, 19)
	p.addLineStarts = append(p.addLineStarts, 23)
	p.addLineStarts = append(p.addLineStarts, 29)

	// one|two|three|five|six|seven| -- add buffer
	// one|two|three|four|five|six|seven| -- contents

	// two
	node2 := &node{
		bufferType:              bufferTypeAdd,
		start:                   4,
		len:                     4,
		leftSubtreeLen:          4,
		newlineCount:            1,
		leftSubtreeNewlineCount: 1,
		priority:                0,
		parent:                  p.root,
	}
	p.root.left = node2

	// one
	node1 := &node{
		bufferType:              bufferTypeAdd,
		start:                   0,
		len:                     4,
		leftSubtreeLen:          0,
		newlineCount:            1,
		leftSubtreeNewlineCount: 0,
		priority:                0,
		parent:                  node2,
	}
	node2.left = node1

	// three
	node3 := &node{
		bufferType:              bufferTypeAdd,
		start:                   8,
		len:                     6,
		leftSubtreeLen:          0,
		newlineCount:            1,
		leftSubtreeNewlineCount: 0,
		priority:                0,
		parent:                  node2,
	}
	node2.right = node3

	// six
	node6 := &node{
		bufferType:              bufferTypeAdd,
		start:                   19,
		len:                     4,
		leftSubtreeLen:          5,
		newlineCount:            1,
		leftSubtreeNewlineCount: 1,
		priority:                0,
		parent:                  p.root,
	}
	p.root.right = node6

	// five
	node5 := &node{
		bufferType:              bufferTypeAdd,
		start:                   14,
		len:                     5,
		leftSubtreeLen:          0,
		newlineCount:            1,
		leftSubtreeNewlineCount: 0,
		priority:                0,
		parent:                  node6,
	}
	node6.left = node5

	// seven
	node7 := &node{
		bufferType:              bufferTypeAdd,
		start:                   23,
		len:                     6,
		leftSubtreeLen:          0,
		newlineCount:            1,
		leftSubtreeNewlineCount: 0,
		priority:                0,
		parent:                  node6,
	}
	node6.right = node7

	return p
}

func TestNewPieceTreeBufferInit(t *testing.T) {
	p := NewPieceTree([]byte("one\ntwo\nthree\n"))

	assert.Equal(t, []byte("one\ntwo\nthree\n"), p.originalBuffer)

	// tree
	assert.Equal(t, []int{0, 4, 8, 14}, p.originalLineStarts)
	assert.Equal(t, []int{0}, p.addLineStarts)

	// node values
	assert.Equal(t, bufferTypeOriginal, p.root.bufferType)
	assert.Equal(t, 14, p.root.len)
	assert.Equal(t, 0, p.root.start)
	assert.Equal(t, 3, p.root.newlineCount)
}

func TestContentsReturnsFullTreeContent(t *testing.T) {
	p := pieceTreeFixture()
	contents, err := p.Contents()

	assert.NoError(t, err)
	assert.Equal(t, []byte("one\ntwo\nthree\nfour\nfive\nsix\nseven\n"), contents)
}
