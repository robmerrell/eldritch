package buffer

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand/v2"
	"sort"
)

var (
	ErrInvalidOffset = errors.New("Invalid offset")
	ErrNoContent     = errors.New("No Content Given")
	ErrInvalidRange  = errors.New("End Is LTE Start")
	ErrInvalidStart  = errors.New("Start is less than 0")
	ErrInvalidEnd    = errors.New("End is greater than document")
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

	// full content size
	len int

	// function to use to determine the priority. Handy for having
	// some tests that need a certain structure.
	priorityFunc func() int
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

	priorityFunc := rand.Int

	var root *node
	if len(initial) > 0 {
		root = &node{
			bufferType:              bufferTypeOriginal,
			start:                   0,
			len:                     len(initial),
			leftSubtreeLen:          0,
			newlineCount:            len(originalLineStarts) - 1,
			leftSubtreeNewlineCount: 0,
			priority:                priorityFunc(),
		}
	}

	return &PieceTree{
		originalBuffer:     initial,
		originalLineStarts: originalLineStarts,
		addBuffer:          []byte{},
		addLineStarts:      []int{0},
		len:                len(initial),
		priorityFunc:       priorityFunc,
		root:               root,
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
	if end <= start {
		return nil, ErrInvalidRange
	}

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

// newlinesBefore counts the newlines before the given offset.
func (p *PieceTree) newlinesBefore(node *node, offset int) int {
	if node == nil {
		return 0
	}

	// keep going left
	if offset < node.leftSubtreeLen {
		return p.newlinesBefore(node.left, offset)
	}

	// found on the node
	local := offset - node.leftSubtreeLen
	if local <= node.len {
		var lineStarts []int
		if node.bufferType == bufferTypeOriginal {
			lineStarts = p.originalLineStarts
		} else {
			lineStarts = p.addLineStarts
		}

		prefix := upperBounds(lineStarts, node.start+local) - upperBounds(lineStarts, node.start)
		return node.leftSubtreeNewlineCount + prefix
	}

	// go right
	return node.leftSubtreeNewlineCount +
		node.newlineCount +
		p.newlinesBefore(node.right, local-node.len)
}

// lineStartOffset returns the offset just after the newline in a subtree.
// The boolean is if it is a true 0 or the node didn't exist.
func (p *PieceTree) lineStartOffset(node *node, row int) (int, bool) {
	if row == 0 {
		return 0, true
	}
	if node == nil {
		return 0, false
	}

	// keep going left
	if row <= node.leftSubtreeNewlineCount {
		return p.lineStartOffset(node.left, row)
	}

	// found on the node
	row -= node.leftSubtreeNewlineCount
	if row <= node.newlineCount {
		var lineStarts []int
		if node.bufferType == bufferTypeOriginal {
			lineStarts = p.originalLineStarts
		} else {
			lineStarts = p.addLineStarts
		}

		index := upperBounds(lineStarts, node.start)
		return node.leftSubtreeLen + lineStarts[index+row-1] - node.start, true
	}

	// go right
	offset, ok := p.lineStartOffset(node.right, row-node.newlineCount)
	return node.leftSubtreeLen + node.len + offset, ok
}

// OffsetToCoords converts a document offset to coordinates.
func (p *PieceTree) OffsetToCoords(offset int) Coords {
	row := p.newlinesBefore(p.root, offset)
	lineStart, _ := p.lineStartOffset(p.root, row)

	// TODO: using an offset out of bounds just returns garbage I need some error checking
	return Coords{Line: row + 1, Col: offset - lineStart + 1}
}

// CoordsToOffset convert document coordinates to an offset.
func (p *PieceTree) CoordsToOffset(coords Coords) (int, error) {
	if coords.Line < 1 || coords.Col < 1 {
		return 0, fmt.Errorf("Invalid coordinates %d, %d", coords.Line, coords.Col)
	}

	// coords are 1 based and the tree is 0
	line := coords.Line - 1
	col := coords.Col - 1

	lineStart, ok := p.lineStartOffset(p.root, line)
	if !ok {
		return 0, fmt.Errorf("Invalid coordinates %d, %d", coords.Line, coords.Col)
	}

	lineEnd := p.len
	if nextLineStart, ok := p.lineStartOffset(p.root, line+1); ok {
		lineEnd = nextLineStart - 1
	}

	if lineStart+col > lineEnd {
		return 0, fmt.Errorf("Invalid coordinates %d, %d", coords.Line, coords.Col)
	}

	return lineStart + col, nil
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

// validate checks the tree structure against caches. This shouldn't be used from the application
// but used in tests.
func (p *PieceTree) validate() error {
	_, _, err := p.validateNode(p.root)
	return err
}

func (p *PieceTree) validateNode(node *node) (leftSubtreeLen, leftSubtreeNewlineCount int, err error) {
	if node != nil {
		// go left
		leftLen, leftCount, err := p.validateNode(node.left)
		if err != nil {
			return leftLen, leftCount, err
		}

		// go right
		rightLen, rightCount, err := p.validateNode(node.right)
		if err != nil {
			return rightLen, rightCount, err
		}

		// check the links
		if node.left != nil {
			// check left priority
			if node.left.priority > node.priority {
				return 0, 0, fmt.Errorf("Invalid priority (%d) and left (%d)", node.start, node.left.start)
			}

			// check left link
			if node.left.parent != node {
				// use the starts to identify the nodes
				return 0, 0, fmt.Errorf("Mismatched parent (%d) and left (%d)", node.start, node.left.start)
			}
		}
		if node.right != nil {
			// check right priority
			if node.right.priority > node.priority {
				return 0, 0, fmt.Errorf("Invalid priority (%d) and right (%d)", node.start, node.right.start)
			}

			// check right link
			if node.right.parent != node {
				// use the starts to identify the nodes
				return 0, 0, fmt.Errorf("Mismatched parent (%d) and right (%d)", node.start, node.right.start)
			}
		}

		// check caches
		if node.leftSubtreeLen != leftLen {
			return 0, 0, fmt.Errorf("Len cache mismatch on (%d) node %d != calculated %d", node.start, node.leftSubtreeLen, leftLen)
		}
		if node.leftSubtreeNewlineCount != leftCount {
			return 0, 0, fmt.Errorf("Len cache mismatch on (%d) node %d != calculated %d", node.start, node.leftSubtreeNewlineCount, leftLen)
		}

		return leftLen + node.len + rightLen, leftCount + node.newlineCount + rightCount, nil
	}

	return 0, 0, nil
}

// Insert inserts bytes into the piece tree. Nodes are split, created and balanced as needed.
func (p *PieceTree) Insert(offset int, contents []byte) error {
	if offset < 0 || (offset > p.len && p.root != nil) {
		return ErrInvalidOffset
	}

	if len(contents) < 1 {
		return ErrNoContent
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

	p.len += contentLen

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

	// tree is empty
	if p.root == nil {
		root := &node{
			bufferType:              bufferTypeAdd,
			start:                   0,
			len:                     contentLen,
			leftSubtreeLen:          0,
			leftSubtreeNewlineCount: 0,
			priority:                p.priorityFunc(),
		}
		root.newlineCount = p.nodeNewlineCount(root)

		p.root = root
		p.lastInsertNode = root
		p.lastInsertEnd = contentLen
		return nil
	}

	// end of the document
	if offset == p.len-contentLen {
		// find the node all the way to the right
		right := p.root
		for right.right != nil {
			right = right.right
		}

		newNode := &node{
			bufferType:   bufferTypeAdd,
			start:        addOffset,
			len:          contentLen,
			newlineCount: newlineCount,
			parent:       right,
			priority:     p.priorityFunc(),
		}
		right.right = newNode

		p.lastInsertNode = newNode
		p.lastInsertEnd = offset + contentLen
		p.balanceHeap(newNode)
		return nil
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
				priority:                p.priorityFunc(),
			}
			nodeLoc.node.left = newNode

			if newNode.left != nil {
				newNode.left.parent = newNode
			}

			p.updateCaches(nodeLoc.node, newNode, contentLen, newlineCount)
			p.lastInsertNode = newNode
			p.lastInsertEnd = offset + contentLen
			p.balanceHeap(newNode)
		} else {
			// middle of a piece - Create a new left piece, Create a new right, and modify current piece to point at new content

			// new sides
			newLeft := &node{
				bufferType:              nodeLoc.node.bufferType,
				start:                   nodeLoc.node.start,
				len:                     nodeLoc.localOffset,
				leftSubtreeLen:          nodeLoc.node.leftSubtreeLen,
				leftSubtreeNewlineCount: nodeLoc.node.leftSubtreeNewlineCount,
				parent:                  nodeLoc.node,
				left:                    nodeLoc.node.left,
				right:                   nil,
				priority:                p.priorityFunc(),
			}
			newLeft.newlineCount = p.nodeNewlineCount(newLeft)

			newRight := &node{
				bufferType:              nodeLoc.node.bufferType,
				start:                   nodeLoc.node.start + nodeLoc.localOffset,
				len:                     nodeLoc.node.len - nodeLoc.localOffset,
				leftSubtreeLen:          0,
				leftSubtreeNewlineCount: 0,
				parent:                  nodeLoc.node,
				left:                    nil,
				right:                   nodeLoc.node.right,
				priority:                p.priorityFunc(),
			}
			newRight.newlineCount = p.nodeNewlineCount(newRight)

			// modify current
			nodeLoc.node.bufferType = bufferTypeAdd
			nodeLoc.node.left = newLeft
			nodeLoc.node.right = newRight
			nodeLoc.node.start = addOffset
			nodeLoc.node.len = contentLen
			nodeLoc.node.leftSubtreeLen += newLeft.len
			nodeLoc.node.newlineCount = newlineCount
			nodeLoc.node.leftSubtreeNewlineCount += newLeft.newlineCount

			// modify the subtrees that moved around for the new nodes
			if newLeft.left != nil {
				newLeft.left.parent = newLeft
			}
			if newRight.right != nil {
				newRight.right.parent = newRight
			}

			p.updateCaches(nodeLoc.node.parent, nodeLoc.node, contentLen, newlineCount)
			p.lastInsertNode = nodeLoc.node
			p.lastInsertEnd = offset + contentLen

			p.balanceHeap(newLeft)
			p.balanceHeap(newRight)
		}

	} else {
		return errors.New("Out of bounds insert")
	}

	return nil
}

// Delete delets content by adjusting the tree and not touching the buffer. Like BoundedContents
// the offsets are [start, end).
func (p *PieceTree) Delete(start, end int) error {
	if end <= start {
		return ErrInvalidRange
	}

	if end > p.len {
		return ErrInvalidEnd
	}

	if start < 0 {
		return ErrInvalidStart
	}

	remaining := end - start
	p.len -= remaining

	// clear the last insert
	// TODO: Is this always necessary, are there cases where we can keep it?
	p.lastInsertEnd = 0
	p.lastInsertNode = nil

	for remaining > 0 {
		nodeLoc := p.nodeAtOffset(p.root, start)
		if nodeLoc == nil {
			return ErrInvalidOffset
		}

		bytesRemoved := 0

		// if the local offset is 0 the only valid operations are completely removing
		// the node or trimming from the beginning. If it is set > 0 the only operations
		// are splitting and trimming from the end.
		if nodeLoc.localOffset == 0 {
			// trim from the beginning
			if remaining < nodeLoc.node.len {
				oldNewlineCount := nodeLoc.node.newlineCount

				nodeLoc.node.start += remaining
				nodeLoc.node.len -= remaining
				nodeLoc.node.newlineCount = p.nodeNewlineCount(nodeLoc.node)

				p.updateCaches(nodeLoc.node.parent, nodeLoc.node, -remaining, nodeLoc.node.newlineCount-oldNewlineCount)

				// We should never have any remaining bytes to delete after this
				bytesRemoved = remaining
			} else {
				// remove the node
				bytesRemoved = nodeLoc.node.len

				// If we set the priority to -1 balanceHeap should do the work
				// for us. I don't think it's the most effecient way, but it's
				// effecient enough.
				nodeLoc.node.priority = -1
				p.balanceHeap(nodeLoc.node)
				p.updateCaches(nodeLoc.node.parent, nodeLoc.node, -nodeLoc.node.len, -nodeLoc.node.newlineCount)
				p.replaceChild(nodeLoc.node.parent, nodeLoc.node, nil)
			}
		} else {
			// trim end
			if nodeLoc.localOffset+remaining >= nodeLoc.node.len {
				oldNewlineCount := nodeLoc.node.newlineCount

				bytesRemoved = nodeLoc.node.len - nodeLoc.localOffset
				nodeLoc.node.len -= bytesRemoved
				nodeLoc.node.newlineCount = p.nodeNewlineCount(nodeLoc.node)

				p.updateCaches(nodeLoc.node.parent, nodeLoc.node, -bytesRemoved, nodeLoc.node.newlineCount-oldNewlineCount)
			} else {
				bytesRemoved = remaining
				oldNewlineCount := nodeLoc.node.newlineCount

				// right side
				tailStart := nodeLoc.node.start + nodeLoc.localOffset + remaining
				tailLen := nodeLoc.node.len - nodeLoc.localOffset - remaining

				newRight := &node{
					bufferType:              nodeLoc.node.bufferType,
					start:                   tailStart,
					len:                     tailLen,
					leftSubtreeLen:          0,
					leftSubtreeNewlineCount: 0,
					parent:                  nodeLoc.node,
					left:                    nil,
					right:                   nodeLoc.node.right,
					priority:                p.priorityFunc(),
				}
				newRight.newlineCount = p.nodeNewlineCount(newRight)

				if newRight.right != nil {
					newRight.right.parent = newRight
				}
				nodeLoc.node.right = newRight

				// shrink the existing node
				nodeLoc.node.len = nodeLoc.localOffset
				nodeLoc.node.newlineCount = p.nodeNewlineCount(nodeLoc.node)

				// update the ancestors
				newlineDelta := nodeLoc.node.newlineCount + newRight.newlineCount - oldNewlineCount
				p.updateCaches(nodeLoc.node.parent, nodeLoc.node, -bytesRemoved, newlineDelta)

				p.balanceHeap(newRight)
			}
		}

		remaining -= bytesRemoved
	}

	return nil
}

// Swap out child nodes and set the root if necessary
func (p *PieceTree) replaceChild(parent, oldChild, newChild *node) {
	if parent == nil {
		p.root = newChild
	} else if parent.left == oldChild {
		parent.left = newChild
	} else {
		parent.right = newChild
	}
	if newChild != nil {
		newChild.parent = parent
	}
}

// Compare the node with its parent and rotate if needed
func (p *PieceTree) rotateRight(node *node) {
	left := node.left
	grandChild := left.right

	node.left = grandChild
	if grandChild != nil {
		grandChild.parent = node
	}

	p.replaceChild(node.parent, node, left)
	left.right = node
	node.parent = left

	// update caches of the rotated node there's no size changes so do it inline
	node.leftSubtreeLen -= left.leftSubtreeLen + left.len
	node.leftSubtreeNewlineCount -= left.leftSubtreeNewlineCount + left.newlineCount
}

func (p *PieceTree) rotateLeft(node *node) {
	right := node.right
	grandChild := right.left

	node.right = grandChild
	if grandChild != nil {
		grandChild.parent = node
	}

	p.replaceChild(node.parent, node, right)
	right.left = node
	node.parent = right

	right.leftSubtreeLen += node.leftSubtreeLen + node.len
	right.leftSubtreeNewlineCount += node.leftSubtreeNewlineCount + node.newlineCount
}

// Calls rotate on the node up until it is in the proper place in the treap.
func (p *PieceTree) balanceHeap(node *node) {
	// check down first -- the node might sit above nodes that out rank it
	for {
		child := node.left

		if node.right != nil && (child == nil || node.right.priority > child.priority) {
			child = node.right
		}

		if child == nil || child.priority <= node.priority {
			break
		}

		if child == node.left {
			p.rotateRight(node)
		} else {
			p.rotateLeft(node)
		}
	}

	// check up
	for node.parent != nil && node.priority > node.parent.priority {
		if node == node.parent.left {
			p.rotateRight(node.parent)
		} else {
			p.rotateLeft(node.parent)
		}
	}
}
