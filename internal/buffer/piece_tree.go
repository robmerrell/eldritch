package buffer

import (
	"bytes"
	"errors"
	"math/rand/v2"
	"sort"
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

	// caches to help writing to the last node
	lastInsertNode *node
	lastInsertEnd  int
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

// Contents returns the entire set of contents stored in the piece tree as a slice of bytes.
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

// BoundedContents returns the contents between the offsets [start, end)
func (p *PieceTree) BoundedContents(start, end int) ([]byte, error) {
	var contents bytes.Buffer
	err := p.collectBoundedContents(p.root, &contents, start, end, 0)

	return contents.Bytes(), err
}

// collectBoundedContents recurses through the tree and collect contents for nodes between start and end.
func (p *PieceTree) collectBoundedContents(node *node, contents *bytes.Buffer, start, end, base int) error {
	if node != nil {
		contentStart := base + node.leftSubtreeLen
		contentEnd := contentStart + node.len

		// keep going left
		if start < contentStart {
			p.collectBoundedContents(node.left, contents, start, end, base)
		}

		// append data and trim if needed
		lo := max(start, contentStart)
		hi := min(end, contentEnd)
		if lo < hi {
			buf := p.nodeContents(node)

			_, err := contents.Write(buf[lo-contentStart : hi-contentStart])
			if err != nil {
				return err
			}
		}

		// go right
		if end > contentEnd {
			p.collectBoundedContents(node.right, contents, start, end, contentEnd)
		}
	}

	return nil
}

// nodeLocation store a node and a local offset inside of that node
type nodeLocation struct {
	node        *node
	localOffset int
}

// nodeAtOffset returns a nodeLocation for the node that countains the offset of bytes.
func (p *PieceTree) nodeAtOffset(node *node, offset int) *nodeLocation {
	if node != nil {
		// keep going left
		if offset < node.leftSubtreeLen {
			return p.nodeAtOffset(node.left, offset)
		}

		// found on the node
		local := offset - node.leftSubtreeLen
		if local < node.len {
			return &nodeLocation{node: node, localOffset: local}
		}

		// go right
		return p.nodeAtOffset(node.right, offset-node.leftSubtreeLen-node.len)
	}

	return &nodeLocation{node: nil, localOffset: 0}
}

// Returns the upperBounds of an int list above the bounds
func upperBounds(cont []int, bounds int) int {
	return sort.Search(len(cont), func(i int) bool { return cont[i] > bounds })
}

// nodeNewlineCount returns the number of newlines in a piece using the buffer line starts so that
// the buffer content doesn't need to be read.
func (p *PieceTree) nodeNewlineCount(node *node) int {
	var lineStarts []int
	if node.bufferType == bufferTypeOriginal {
		lineStarts = p.originalLineStarts
	} else {
		lineStarts = p.addLineStarts
	}

	end := node.start + node.len
	return upperBounds(lineStarts, end) - upperBounds(lineStarts, node.start)
}

// updateCaches walks upt the tree and updates all size caches
func (p *PieceTree) updateCaches(node *node, fromNode *node, delta int, newlineDelta int) {
	if node != nil {
		if fromNode == node.left {
			node.leftSubtreeLen += delta
			node.leftSubtreeNewlineCount += newlineDelta
		}

		p.updateCaches(node.parent, node, delta, newlineDelta)
	}
}

// Insert inserts bytes into the piece tree. Nodes are split, created and balanced as needed.
func (p *PieceTree) Insert(offset int, contents []byte) error {
	if len(contents) < 1 {
		return errors.New("Missing content for insert")
	}

	contentLen := len(contents)
	addOffset := len(p.addBuffer)
	p.addBuffer = append(p.addBuffer, contents...)

	// add the line starts
	newlineCount := 0
	for i, byteVal := range contents {
		if byteVal == '\n' {
			newlineCount += 1
			p.addLineStarts = append(p.addLineStarts, addOffset+i+1)
		}
	}

	// try to grow the last written node
	if p.lastInsertNode != nil {
		node := p.lastInsertNode
		if node.bufferType == bufferTypeAdd && (node.start+node.len == addOffset) && offset == p.lastInsertEnd {
			node.len += contentLen
			node.newlineCount += newlineCount
			p.updateCaches(node.parent, node, contentLen, newlineCount)
			p.lastInsertEnd += contentLen
			return nil
		}
	}

	nodeLoc := p.nodeAtOffset(p.root, offset)
	if nodeLoc != nil {
		if nodeLoc.localOffset == 0 {
			// create a new piece and place on left side of parent and relink left child further down
			newNode := &node{
				bufferType:              bufferTypeAdd,
				start:                   addOffset,
				len:                     contentLen,
				leftSubtreeLen:          nodeLoc.node.leftSubtreeLen,
				newlineCount:            newlineCount,
				leftSubtreeNewlineCount: nodeLoc.node.leftSubtreeNewlineCount,
				parent:                  nodeLoc.node,
				left:                    nodeLoc.node.left,
				priority:                0,
			}
			nodeLoc.node.left = newNode

			if newNode.left != nil {
				newNode.left.parent = newNode
			}

			p.updateCaches(nodeLoc.node, newNode, contentLen, newlineCount)
			p.lastInsertNode = newNode
			p.lastInsertEnd = offset + contentLen
		}

	} else {
		return errors.New("Out of bounds insert")
	}

	return nil
}
