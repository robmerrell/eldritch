package buffer

import (
	"errors"
	"os"
	"slices"
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
	p.len = 34
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

func TestBoundedContentsReturnsBetweenOffsets(t *testing.T) {
	p := pieceTreeFixture()

	// end before start
	contents, err := p.BoundedContents(20, 2)
	assert.ErrorIs(t, err, ErrInvalidRange)

	// multiple pieces
	contents, err = p.BoundedContents(6, 20)
	assert.NoError(t, err)
	assert.Equal(t, []byte("o\nthree\nfour\nf"), contents)

	// full range
	contents, err = p.BoundedContents(0, 2000)
	assert.NoError(t, err)
	assert.Equal(t, []byte("one\ntwo\nthree\nfour\nfive\nsix\nseven\n"), contents)

	// single piece
	contents, err = p.BoundedContents(0, 2)
	assert.NoError(t, err)
	assert.Equal(t, []byte("on"), contents)

	// exactly on boundaries
	contents, err = p.BoundedContents(0, 4)
	assert.NoError(t, err)
	assert.Equal(t, []byte("one\n"), contents)
}

func TestNodeAtOffset(t *testing.T) {
	p := pieceTreeFixture()

	// out of bounds
	nodeLoc := p.nodeAtOffset(p.root, 1000)
	assert.Nil(t, nodeLoc.node)
	assert.Equal(t, 0, nodeLoc.localOffset)

	// one
	nodeLoc = p.nodeAtOffset(p.root, 1)
	assert.Equal(t, p.root.left.left, nodeLoc.node)
	assert.Equal(t, 1, nodeLoc.localOffset)

	// two
	nodeLoc = p.nodeAtOffset(p.root, 4)
	assert.Equal(t, p.root.left, nodeLoc.node)
	assert.Equal(t, 0, nodeLoc.localOffset)

	// three
	nodeLoc = p.nodeAtOffset(p.root, 13)
	assert.Equal(t, p.root.left.right, nodeLoc.node)
	assert.Equal(t, 5, nodeLoc.localOffset)

	// four
	nodeLoc = p.nodeAtOffset(p.root, 15)
	assert.Equal(t, p.root, nodeLoc.node)
	assert.Equal(t, 1, nodeLoc.localOffset)

	// five
	nodeLoc = p.nodeAtOffset(p.root, 21)
	assert.Equal(t, p.root.right.left, nodeLoc.node)
	assert.Equal(t, 2, nodeLoc.localOffset)

	// six
	nodeLoc = p.nodeAtOffset(p.root, 25)
	assert.Equal(t, p.root.right, nodeLoc.node)
	assert.Equal(t, 1, nodeLoc.localOffset)

	// seven
	nodeLoc = p.nodeAtOffset(p.root, 31)
	assert.Equal(t, p.root.right.right, nodeLoc.node)
	assert.Equal(t, 3, nodeLoc.localOffset)
}

func TestNodeNewlineCount(t *testing.T) {
	p := pieceTreeFixture()

	// every piece in the fixture holds exactly one line
	assert.Equal(t, 1, p.nodeNewlineCount(p.root.left.left))
	assert.Equal(t, 1, p.nodeNewlineCount(p.root.left))
	assert.Equal(t, 1, p.nodeNewlineCount(p.root.left.right))
	assert.Equal(t, 1, p.nodeNewlineCount(p.root))

	// ending on the newline "one\n"
	assert.Equal(t, 1, p.nodeNewlineCount(&node{bufferType: bufferTypeAdd, start: 0, len: 4}))

	// starting past a newline "two"
	assert.Equal(t, 0, p.nodeNewlineCount(&node{bufferType: bufferTypeAdd, start: 4, len: 3}))
}

func TestUpdateCaches(t *testing.T) {
	p := pieceTreeFixture()

	// left side traversal causes update (update node "1")
	node := p.root.left.left
	p.updateCaches(node.parent, node, 3, 2)
	assert.Equal(t, 7, p.root.left.leftSubtreeLen)
	assert.Equal(t, 17, p.root.leftSubtreeLen)
	assert.Equal(t, 3, p.root.left.leftSubtreeNewlineCount)
	assert.Equal(t, 5, p.root.leftSubtreeNewlineCount)

	// right does not until a left side is used again (update node "3")
	node = p.root.left.right
	p.updateCaches(node.parent, node, 3, 2)
	assert.Equal(t, 7, p.root.left.leftSubtreeLen)
	assert.Equal(t, 20, p.root.leftSubtreeLen)
	assert.Equal(t, 3, p.root.left.leftSubtreeNewlineCount)
	assert.Equal(t, 7, p.root.leftSubtreeNewlineCount)

	// subtree of right updates (update node "5")
	node = p.root.right.left
	p.updateCaches(node.parent, node, 5, 3)
	assert.Equal(t, 10, p.root.right.leftSubtreeLen)
	assert.Equal(t, 20, p.root.leftSubtreeLen)
	assert.Equal(t, 4, p.root.right.leftSubtreeNewlineCount)
	assert.Equal(t, 7, p.root.leftSubtreeNewlineCount)

	// all rights do not update
	node = p.root.right.right
	p.updateCaches(node.parent, node, 5, 3)
	assert.Equal(t, 10, p.root.right.leftSubtreeLen)
	assert.Equal(t, 20, p.root.leftSubtreeLen)
	assert.Equal(t, 4, p.root.right.leftSubtreeNewlineCount)
	assert.Equal(t, 7, p.root.leftSubtreeNewlineCount)
}

func TestInsertAtBeginningOfPiece(t *testing.T) {
	p := pieceTreeFixture()

	// split nodes - adds "2" to the beginning of "two"
	err := p.Insert(4, []byte("2"))
	assert.NoError(t, err)

	contents, err := p.Contents()
	assert.NoError(t, err)
	assert.Equal(t, []byte("one\n2two\nthree\nfour\nfive\nsix\nseven\n"), contents)

	// test the node setup
	newNode := p.root.left
	assert.Equal(t, p.root, newNode.parent)
	assert.Equal(t, 5, newNode.leftSubtreeLen)
	assert.Equal(t, 15, newNode.parent.leftSubtreeLen)
}

func TestInsertInMiddleOfPiece(t *testing.T) {
	p := pieceTreeFixture()

	// splits nodes adds "4" to the middle of "four"
	err := p.Insert(16, []byte("4"))
	assert.NoError(t, err)

	contents, err := p.Contents()
	assert.NoError(t, err)
	assert.Equal(t, []byte("one\ntwo\nthree\nfo4ur\nfive\nsix\nseven\n"), contents)

	// updated node
	updated := p.root
	assert.Equal(t, 1, updated.len)
	assert.Equal(t, 16, updated.leftSubtreeLen)

	// new left
	assert.Equal(t, 2, updated.left.len)
	assert.Equal(t, 14, updated.left.leftSubtreeLen)

	// new right
	assert.Equal(t, 3, updated.right.len)
	assert.Equal(t, 0, updated.right.leftSubtreeLen)
}

func TestInsertAtEndOfPiece(t *testing.T) {
	p := pieceTreeFixture()

	// splits nodes adds "1" to the end of "one"
	err := p.Insert(3, []byte("1"))
	assert.NoError(t, err)

	contents, err := p.Contents()
	assert.NoError(t, err)
	assert.Equal(t, []byte("one1\ntwo\nthree\nfour\nfive\nsix\nseven\n"), contents)

	// parent
	parent := p.root.left.left
	assert.Equal(t, 1, parent.len)
	assert.Equal(t, 3, parent.leftSubtreeLen)

	// left
	assert.Equal(t, 3, parent.left.len)
	assert.Equal(t, 0, parent.left.leftSubtreeLen)
}

func TestInsertContinueWriting(t *testing.T) {
	p := pieceTreeFixture()

	// splits nodes, adds "1" to the end of "one"
	err := p.Insert(3, []byte("1"))
	assert.NoError(t, err)

	// keeps inserting onto it
	err = p.Insert(4, []byte("2345"))
	assert.NoError(t, err)

	contents, err := p.Contents()
	assert.NoError(t, err)
	assert.Equal(t, []byte("one12345\ntwo\nthree\nfour\nfive\nsix\nseven\n"), contents)

	// nodes
	parent := p.root.left.left
	assert.Equal(t, 5, parent.len)
	assert.Equal(t, 9, parent.parent.leftSubtreeLen)
}

func TestEnd(t *testing.T) {
	p := NewPieceTree([]byte("hello"))
	err := p.Insert(5, []byte(", world"))
	assert.NoError(t, err)

	contents, err := p.Contents()
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello, world"), contents)
}

func FuzzInsert(f *testing.F) {
	f.Fuzz(func(t *testing.T, offset int, contents []byte) {
		p := NewPieceTree([]byte("hello"))
		ref := []byte("hello")

		err := p.Insert(offset, contents)
		if !errors.Is(err, ErrInvalidOffset) && !errors.Is(err, ErrNoContent) {
			assert.NoError(t, err)
		}

		if err != nil {
			return
		}

		ref = slices.Insert(ref, offset, contents...)

		treeContents, err := p.Contents()
		assert.NoError(t, err)
		assert.Equal(t, ref, treeContents)
	})
}

func FuzzReadBoundedContents(f *testing.F) {
	content, err := os.ReadFile("testdata/dracula.txt")
	assert.NoError(f, err)
	contentLen := uint16(len(content))

	f.Fuzz(func(t *testing.T, start, end uint16) {
		// clamp start and end to the end of the document
		if start > contentLen {
			start = contentLen
		}
		if end > contentLen {
			end = contentLen
		}

		if end <= start {
			return
		}

		p := NewPieceTree(content)
		contents, err := p.BoundedContents(int(start), int(end))
		assert.NoError(t, err)
		assert.Equal(t, content[start:end], contents)
	})
}
