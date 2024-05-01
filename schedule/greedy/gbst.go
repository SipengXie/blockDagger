package greedy

import (
	"blockDagger/types"
	"errors"
)

var (
	ErrEmptyTree = errors.New("empty tree")
)

// GBST structure. Public methods are Add, Remove, Update, Search, Flatten.
type GBST struct {
	root *GBSTNode
}

func (t *GBST) Add(p *types.Task) {
	t.root = t.root.add(p)
}

func (t *GBST) Remove(p *types.Task) {
	t.root = t.root.remove(p)
}

func (t *GBST) Search(targetGas uint64) *types.Task {
	return t.root.search(targetGas)
}

func (t *GBST) Largest() (*types.Task, error) {
	if t.root == nil {
		return nil, ErrEmptyTree
	}
	return t.root.findLargest().task, nil // might get error if root is nil
}

func (t *GBST) Flatten() []*GBSTNode {
	nodes := make([]*GBSTNode, 0)
	if t.root == nil {
		return nodes
	}
	t.root.displayNodesInOrder(&nodes)
	return nodes
}

// GBSTNode structure
type GBSTNode struct {
	task *types.Task

	// height counts nodes (not edges)
	height int
	left   *GBSTNode
	right  *GBSTNode
}

// 正常比大小，Gas一样比Tid
func my_cmp(a, b *types.Task) int {
	if a.Cost < b.Cost {
		return -1
	}
	if a.Cost > b.Cost {
		return 1
	}
	if a.ID < b.ID {
		return -1
	}
	if a.ID > b.ID {
		return 1
	}
	return 0
}

// Adds a new node
func (n *GBSTNode) add(p *types.Task) *GBSTNode {
	if n == nil {
		return &GBSTNode{p, 1, nil, nil}
	}
	res := my_cmp(p, n.task)
	if res < 0 {
		n.left = n.left.add(p)
	} else {
		n.right = n.right.add(p)
		// add中不会有res = 0的情况
	}
	return n.rebalanceTree()
}

// Removes a node
func (n *GBSTNode) remove(p *types.Task) *GBSTNode {
	if n == nil {
		return nil
	}
	res := my_cmp(p, n.task)
	if res < 0 {
		n.left = n.left.remove(p)
	} else if res > 0 {
		n.right = n.right.remove(p)
	} else {
		if n.left != nil && n.right != nil {
			// node to delete found with both children;
			// replace values with smallest node of the right sub-tree
			rightMinNode := n.right.findSmallest()
			n.task = rightMinNode.task
			// delete smallest node that we replaced
			n.right = n.right.remove(rightMinNode.task)
		} else if n.left != nil {
			// node only has left child
			n = n.left
		} else if n.right != nil {
			// node only has right child
			n = n.right
		} else {
			// node has no children
			n = nil
			return n
		}
	}
	return n.rebalanceTree()
}

// 寻找小于等于targetGas的最大的节点
func (n *GBSTNode) search(targetGas uint64) *types.Task {
	if n == nil {
		return nil
	}
	if n.task.Cost == targetGas {
		return n.task
	}
	// 当前小于Target，向右走
	if n.task.Cost < targetGas {
		if n.right == nil {
			return n.task
		}
		// n是一个可能的答案，我们还要去右子树找
		// 若右子树找不到答案，则返回n
		tmp := n.right.search(targetGas)
		if tmp == nil {
			return n.task
		}
		return tmp
	} else {
		// 当前大于Target，向左走
		return n.left.search(targetGas)
	}
}

func (n *GBSTNode) displayNodesInOrder(nodes *[]*GBSTNode) {
	if n.left != nil {
		n.left.displayNodesInOrder(nodes)
	}
	(*nodes) = append((*nodes), n)
	if n.right != nil {
		n.right.displayNodesInOrder(nodes)
	}
}

func (n *GBSTNode) getHeight() int {
	if n == nil {
		return 0
	}
	return n.height
}

func (n *GBSTNode) recalculateHeight() {
	n.height = 1 + max(n.left.getHeight(), n.right.getHeight())
}

// Checks if node is balanced and rebalance
func (n *GBSTNode) rebalanceTree() *GBSTNode {
	if n == nil {
		return n
	}
	n.recalculateHeight()

	// check balance factor and rotateLeft if right-heavy and rotateRight if left-heavy
	balanceFactor := n.left.getHeight() - n.right.getHeight()
	if balanceFactor == -2 {
		// check if child is left-heavy and rotateRight first
		if n.right.left.getHeight() > n.right.right.getHeight() {
			n.right = n.right.rotateRight()
		}
		return n.rotateLeft()
	} else if balanceFactor == 2 {
		// check if child is right-heavy and rotateLeft first
		if n.left.right.getHeight() > n.left.left.getHeight() {
			n.left = n.left.rotateLeft()
		}
		return n.rotateRight()
	}
	return n
}

// Rotate nodes left to balance node
func (n *GBSTNode) rotateLeft() *GBSTNode {
	newRoot := n.right
	n.right = newRoot.left
	newRoot.left = n
	n.recalculateHeight()
	newRoot.recalculateHeight()
	return newRoot
}

// Rotate nodes right to balance node
func (n *GBSTNode) rotateRight() *GBSTNode {
	newRoot := n.left
	n.left = newRoot.right
	newRoot.right = n
	n.recalculateHeight()
	newRoot.recalculateHeight()
	return newRoot
}

// Finds the smallest child (based on the key) for the current node
func (n *GBSTNode) findSmallest() *GBSTNode {
	if n.left != nil {
		return n.left.findSmallest()
	} else {
		return n
	}
}

// Finds the largest child (based on the key) for the current node
func (n *GBSTNode) findLargest() *GBSTNode {
	if n.right != nil {
		return n.right.findLargest()
	} else {
		return n
	}
}
