package buffer

import (
	"bytes"
	"math/rand/v2"
)

// the two types of buffers for the piece tree.
type bufferType int

const (
	bufferTypeOriginal bufferType = iota
	bufferTypeAdd
)

// treap node for the tree
type node struct {
	// buffer the piece points to
	bufferType bufferType

	// node details
	priority     int
	start        int
	len          int
	newlineCount int

	// parent/child linking
	left   *node
	right  *node
	parent *node

	// subtree caches
	leftSubtreeLen          int
	leftSubtreeNewlineCount int
}

type PieceTree struct {
	// content buffers. Original holds the initial file state
	// and add is an append only buffer.
	originalBuffer     []byte
	originalLineStarts []int

	addBuffer     []byte
	addLineStarts []int

	// treap of the pieces
	root *node

	// caches
	// addLineStarts
	// lastInsertNode
	// lastInsertEnd
}

// NewPieceTree create a new piece tree with an initial content. Original buffer is filled
// with the contents.
func NewPieceTree(initial []byte) *PieceTree {
	// find the line starts. Always start at 0.
	originalLineStarts := []int{0}
	for i, byte := range initial {
		if byte == '\n' {
			originalLineStarts = append(originalLineStarts, i+1)
		}
	}
	rand.Int()

	return &PieceTree{
		originalBuffer:     initial,
		originalLineStarts: originalLineStarts,
		addBuffer:          []byte{},
		addLineStarts:      []int{0},
		root: &node{
			bufferType:              bufferTypeOriginal,
			start:                   0,
			len:                     len(initial),
			leftSubtreeLen:          0,
			newlineCount:            len(originalLineStarts) - 1,
			leftSubtreeNewlineCount: 0,
			priority:                rand.Int(),
		},
	}
}

// nodeContents returns the contents of a single node in the tree
func (p *PieceTree) nodeContents(node *node) []byte {
	if node.bufferType == bufferTypeOriginal {
		return p.originalBuffer[node.start : node.start+node.len]
	}

	return p.addBuffer[node.start : node.start+node.len]
}

// Contents returns the entire set of contents stored in the piece tree as a slice
// of bytes.
func (p *PieceTree) Contents() ([]byte, error) {
	var contents bytes.Buffer
	err := p.collectContents(p.root, &contents)

	return contents.Bytes(), err
}

// collectContents recurses through the tree and collect contents for all nodes.
func (p *PieceTree) collectContents(node *node, contents *bytes.Buffer) error {
	if node != nil {
		p.collectContents(node.left, contents)

		_, err := contents.Write(p.nodeContents(node))
		if err != nil {
			return err
		}

		p.collectContents(node.right, contents)
	}

	return nil
}
