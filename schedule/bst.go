package schedule

// 维护st
type BstNode struct {
	st     uint64
	length uint64

	lson, rson *BstNode
	height     int
}

func (n *BstNode) getHeight() int {
	if n == nil {
		return 0
	}
	return n.height
}

func (n *BstNode) recalculateHeight() {
	n.height = 1 + max(n.lson.getHeight(), n.rson.getHeight())
}

// Checks if BstNode is balanced and rebalance
func (n *BstNode) rebalanceTree() *BstNode {
	if n == nil {
		return n
	}
	n.recalculateHeight()

	// check balance factor and rotatelson if rson-heavy and rotaterson if lson-heavy
	balanceFactor := n.lson.getHeight() - n.rson.getHeight()
	if balanceFactor == -2 {
		// check if child is lson-heavy and rotaterson first
		if n.rson.lson.getHeight() > n.rson.rson.getHeight() {
			n.rson = n.rson.rotaterson()
		}
		return n.rotatelson()
	} else if balanceFactor == 2 {
		// check if child is rson-heavy and rotatelson first
		if n.lson.rson.getHeight() > n.lson.lson.getHeight() {
			n.lson = n.lson.rotatelson()
		}
		return n.rotaterson()
	}
	return n
}

// Rotate BstNodes lson to balance BstNode
func (n *BstNode) rotatelson() *BstNode {
	newRoot := n.rson
	n.rson = newRoot.lson
	newRoot.lson = n
	n.recalculateHeight()
	newRoot.recalculateHeight()
	return newRoot
}

// Rotate BstNodes rson to balance BstNode
func (n *BstNode) rotaterson() *BstNode {
	newRoot := n.lson
	n.lson = newRoot.rson
	newRoot.rson = n
	n.recalculateHeight()
	newRoot.recalculateHeight()
	return newRoot
}

func (n *BstNode) add(st, length uint64) *BstNode {
	if n == nil {
		return &BstNode{st: st, length: length, height: 1}
	}
	if st < n.st {
		n.lson = n.lson.add(st, length)
	} else {
		n.rson = n.rson.add(st, length)
	}
	return n.rebalanceTree()
}

func (n *BstNode) remove(st uint64) *BstNode {
	if n == nil {
		return nil
	}
	if st < n.st {
		n.lson = n.lson.remove(st)
	} else if st > n.st {
		n.rson = n.rson.remove(st)
	} else {
		if n.lson != nil && n.rson != nil {
			minBstNode := n.rson.findMin()
			n.st = minBstNode.st
			n.rson = n.rson.remove(minBstNode.st)
		} else if n.lson != nil {
			n = n.lson
		} else if n.rson != nil {
			n = n.rson
		} else {
			n = nil
			return n
		}
	}
	return n.rebalanceTree()
}

func (n *BstNode) findMin() *BstNode {
	if n.lson == nil {
		return n
	}
	return n.lson.findMin()
}

// 寻找小于等于st的最大节点
func (n *BstNode) findMaxLessThan(st uint64) *BstNode {
	if n == nil {
		return nil
	}
	if n.st == st {
		return n
	} else if n.st < st {
		if n.rson == nil {
			return n
		}
		// n是一个可能的答案，我们还要去右子树找
		// 若右子树找不到答案，则返回n
		tmp := n.rson.findMaxLessThan(st)
		if tmp == nil {
			return n
		}
		return tmp
	} else {
		// n.payload > st
		// lson == nil 会返回nil
		return n.lson.findMaxLessThan(st)
	}
}

type AVLBST struct {
	root *BstNode
}

func NewTree() *AVLBST {
	return &AVLBST{root: nil}
}

func (t *AVLBST) Add(st, length uint64) {
	t.root = t.root.add(st, length)
}

func (t *AVLBST) Remove(st uint64) {
	t.root = t.root.remove(st)
}

func (t *AVLBST) FindMaxLessThan(st uint64) *BstNode {
	return t.root.findMaxLessThan(st)
}
