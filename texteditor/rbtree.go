package texteditor

import (
	"fmt"
	"strings"

	"golang.org/x/exp/constraints"
)

type colour byte

const (
	black colour = iota
	red
)

// 3 5 22 1 7 53 6 200 18 4 2 2 5 9

func TestAndPrint() {
	tree := NewRBTree[int]()

	tree.insert(3)

	tree.insert(5)

	tree.insert(22)

	tree.insert(1)

	tree.insert(7)

	tree.insert(53)

	tree.insert(6)

	tree.insert(200)

	tree.insert(18)

	tree.insert(4)

	tree.insert(2)

	tree.delete(7)

	fmt.Print(tree.traversePreOrder())
}

func createRBNode(index int) *RBNode {
	return &RBNode{
		left:   nil,
		right:  nil,
		colour: black,
		index:  index,
	}
}

type RBTree struct {
	root      *RBNode
	sentinel  *RBNode
	buffers   []Buffer
	lineCount int64
	length    int64
	eol       [2]rune
	crEnabled bool
}

type pos struct {
	x int64
	y int64
}

type Buffer struct {
	Buffer     [][16000]rune
	LineStarts []int
}

type RBNode struct {
	key    int
	index  int
	parent *RBNode
	left   *RBNode
	right  *RBNode
	colour colour

	start       pos
	end         pos
	length      int
	nlCount     int
	lengthLeft  int
	nlCountLeft int
}

type PieceTable struct {
	buffer   []rune
	rootNode *RBNode
}

func NewRBTree[T constraints.Ordered]() *RBTree {
	leaf := &RBNode{
		left:   nil,
		right:  nil,
		colour: black,
	}
	return &RBTree{
		root:     leaf,
		sentinel: leaf,
	}
}

func (tree *RBTree) leftMost(x *RBNode) *RBNode {
	for x.left != tree.sentinel {
		x = x.left
	}

	return x
}

func (tree *RBTree) rightMost(x *RBNode) *RBNode {
	for x.right != tree.sentinel {
		x = x.right
	}

	return x
}

func (tree *RBTree) fixNodeCounts(x *RBNode) {
	if x == tree.root {
		return
	}

	for x != tree.root && x == x.parent.right {
		x = x.parent
	}

	if x == tree.root {
		return
	}

	x = x.parent

	delta := tree.calculateLength(x.left) - x.lengthLeft
	nlDelta := tree.calculateNL(x.left) - x.nlCountLeft

	if delta == 0 && nlDelta == 0 {
		return
	}

	for x != tree.root {
		if x == x.parent.left {
			x.parent.lengthLeft += delta
			x.parent.nlCountLeft += nlDelta
		}

		x = x.parent
	}
}

func (tree *RBTree) calculateLength(x *RBNode) int {
	if x == tree.sentinel {
		return 0
	}

	return x.lengthLeft + x.length + tree.calculateNL(x.right)
}

func (tree *RBTree) calculateNL(x *RBNode) int {
	if x == tree.sentinel {
		return 0
	}

	return x.nlCountLeft + x.nlCount + tree.calculateLength(x.right)
}

func (tree *RBTree) traversePreOrder() string {
	if tree.root == tree.sentinel {
		return ""
	}

	var col string = "black"
	if tree.root.colour == red {
		col = "red"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s %v\n", col, tree.root.length))

	lineRight := "└──"
	lineLeft := "└──"
	if tree.root.left != tree.sentinel {
		lineRight = "├──"
	}

	tree.traverseNodes(&sb, "", lineRight, tree.root.right, tree.root.left != tree.sentinel)
	tree.traverseNodes(&sb, "", lineLeft, tree.root.left, false)

	return sb.String()
}

func (tree *RBTree) transplant(detach, substitute *RBNode) {
	if detach.parent == tree.sentinel {
		tree.root = substitute
	} else if detach == detach.parent.left {
		detach.parent.left = substitute
	} else {
		detach.parent.right = substitute
	}

	substitute.parent = detach.parent
}

func (tree RBTree) minimum(node *RBNode) *RBNode {
	minNode := node
	for minNode.left != tree.sentinel {
		minNode = minNode.left
	}

	return minNode
}

func (tree *RBTree) delete(key int) {
	node := tree.search(key)

	if node == tree.sentinel {
		return
	}

	var child *RBNode
	deletedNode := node
	originalCol := node.colour

	if node.left == tree.sentinel {
		child = node.right
		tree.transplant(node, node.right)
	} else if node.right == tree.sentinel {
		child = node.left
		tree.transplant(node, node.left)
	} else {
		minNode := tree.minimum(node.right)
		originalCol = minNode.colour
		child = deletedNode.right

		if minNode.parent == node {
			child.parent = minNode
		} else {
			tree.transplant(minNode, minNode.right)
			minNode.right = node.right
			minNode.right.parent = minNode
		}

		tree.transplant(node, minNode)
		minNode.left = node.left
		minNode.left.parent = minNode
		minNode.colour = node.colour
	}

	if originalCol == black {
		tree.deleteFixup(child)
	}
}

func (tree *RBTree) deleteFixup(node *RBNode) {
	for node != tree.root && node.colour == black {
		if node == node.parent.left {
			sibling := node.parent.right

			if sibling.colour == red {
				sibling.colour = black
				node.parent.colour = red
				tree.leftRotate(node.parent)
				sibling = node.parent.right
			}

			if sibling.left.colour == black && sibling.right.colour == black {
				sibling.colour = red
				node = node.parent
			} else {

				if sibling.right.colour == black {
					sibling.left.colour = black
					sibling.colour = red
					tree.rightRotate(sibling)
					sibling = node.parent.right
				}

				sibling.colour = node.parent.colour
				node.parent.colour = black
				sibling.right.colour = black
				tree.leftRotate(node.parent)
				node = tree.root

			}
		} else {
			sibling := node.parent.left

			if sibling.colour == red {
				sibling.colour = black
				node.parent.colour = red
				tree.rightRotate(node.parent)
				sibling = node.parent.left
			}

			if sibling.right.colour == black && node.left.colour == black {
				sibling.colour = red
				node = node.parent
			} else {

				if sibling.left.colour == black {
					sibling.right.colour = black
					sibling.colour = red
					tree.leftRotate(sibling)
					sibling = node.parent.left
				}

				sibling.colour = node.parent.colour
				node.parent.colour = black
				sibling.left.colour = black
				tree.rightRotate(node.parent)
				node = tree.root

			}

		}
	}

	node.colour = black
}

func (tree *RBTree) search(key int) *RBNode {
	node := tree.root
	for node != tree.sentinel && node.key != key {
		if key < node.key {
			node = node.left
		} else {
			node = node.right
		}
	}

	return node
}

func (tree *RBTree) traverseNodes(sb *strings.Builder, padding, pointer string, node *RBNode, hasLeftSibling bool) {
	if node == tree.sentinel {
		return
	}

	var col string = "black"
	if node.colour == red {
		col = "red"
	}

	side := "(r)"
	if node.parent.left == node {
		side = "(l)"
	}

	sb.WriteString(fmt.Sprintf("%s%s%s %s %v\n", side, padding, pointer, col, node.key))

	if hasLeftSibling {
		padding += "│  "
	} else {
		padding += "   "
	}

	lineRight := "└──"
	lineLeft := "└──"
	if node.left != tree.sentinel {
		lineRight = "├──"
	}

	tree.traverseNodes(sb, padding, lineRight, node.right, node.left != tree.sentinel)
	tree.traverseNodes(sb, padding, lineLeft, node.left, false)
}

func (tree *RBTree) insert(key int) {
	node := &RBNode{
		key:    key,
		left:   tree.sentinel,
		right:  tree.sentinel,
		parent: tree.sentinel,
	}
	y := tree.sentinel
	x := tree.root

	for x != tree.sentinel {
		y = x
		if node.key < x.key {
			x = x.left
		} else {
			x = x.right
		}
	}

	if y == tree.sentinel {
		tree.root = node
	} else if node.length < y.length {
		y.left = node
		node.parent = y
	} else {
		y.right = node
		node.parent = y
	}

	node.left = tree.sentinel
	node.right = tree.sentinel
	node.colour = red

	tree.correctInsertion(node)
}

func (tree *RBTree) correctInsertion(z *RBNode) {
	parent := z.parent

	for parent != tree.sentinel && parent.colour == red {
		grandParent := parent.parent

		if parent == grandParent.left {
			fmt.Println("Left")
			uncle := grandParent.right
			if uncle != tree.sentinel && uncle.colour == red {
				parent.colour = black
				uncle.colour = black
				grandParent.colour = red
				z = grandParent
			} else {
				if z == parent.right {
					z = parent
					tree.leftRotate(z)
				}

				parent.colour = black
				grandParent.colour = red
				tree.rightRotate(grandParent)
			}
		} else {
			fmt.Println("Right")
			uncle := grandParent.left
			if uncle != tree.sentinel && uncle.colour == red {
				fmt.Println("uncle red")
				parent.colour = black
				uncle.colour = black
				grandParent.colour = red
			} else {
				if z == parent.left {
					z = parent
					tree.rightRotate(z)
				}

				parent.colour = black
				grandParent.colour = red
				tree.leftRotate(grandParent)
			}
		}

		parent = z.parent
	}

	tree.root.colour = black
}

func (tree *RBTree) leftRotate(x *RBNode) {
	fmt.Println("Left rotation.")
	// child node being rotated upwards
	y := x.right
	// the new right hand side for the node moving downwards is the left hand side of the initial child node
	x.right = y.left

	// change child nodes parent to point to its new parent
	if y.left != tree.sentinel {
		y.left.parent = x
	}

	// node rotated up is assigned the former parent of x
	y.parent = x.parent
	// if the parent node is nil the node rotated up becomes the root
	if x.parent == tree.sentinel {
		tree.root = y
		// else if x was on the left of it's parent then set the parents left node to point at the new child
	} else if x == x.parent.left {
		x.parent.left = y
		// else put it on the right hand side
	} else {
		x.parent.right = y
	}

	// complete the rotation by assigning the node rotated down to the node rotated up's left side
	y.left = x
	// set the parent
	x.parent = y
}

func (tree *RBTree) rightRotate(x *RBNode) {
	fmt.Println("Right rotation.")
	y := x.left
	x.left = y.right

	if y.right != tree.sentinel {
		y.right.parent = x
	}

	y.parent = x.parent
	if x.parent == tree.sentinel {
		tree.root = y
	} else if x == x.parent.left {
		x.parent.left = y
	} else {
		x.parent.right = y
	}

	y.right = x
	x.parent = y
}
