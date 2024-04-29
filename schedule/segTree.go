package schedule

const MAXUINT64 = ^uint64(0)

// 区间查询，单点修改的线段树
type SegTreeNode struct {
	L, R       uint64       // 主管的区间
	seg_max    uint64       // 区间最大值, 存储slot.length
	lson, rson *SegTreeNode // 左右儿子
}

func NewNode(L, R uint64) *SegTreeNode {
	return &SegTreeNode{L: L, R: R}
}

type SegTree struct {
	root *SegTreeNode
}

func NewSegTree(L, R uint64) *SegTree {
	root := NewNode(L, R)
	return &SegTree{root: root}
}

func (t *SegTree) Modify(x, val uint64) {
	t.modify(t.root, x, val)
}

func (t *SegTree) modify(cur *SegTreeNode, x, val uint64) {
	if cur.L == cur.R {
		cur.seg_max = val
		return
	}
	mid := (cur.L + cur.R) >> 1
	if x <= mid {
		if cur.lson == nil {
			cur.lson = NewNode(cur.L, mid)
		}
		t.modify(cur.lson, x, val)
	} else {
		if cur.rson == nil {
			cur.rson = NewNode(mid+1, cur.R)
		}
		t.modify(cur.rson, x, val)
	}
	cur.seg_max = max(cur.lson.seg_max, cur.rson.seg_max)
}

// 寻找[L,R]中第一个大于等于threshold的位置
// 一般的用法是在[EST, MAXENDING]中找到第一个大于等于taskLength的Slot的st和对应的length
func (t *SegTree) Query(L, R, threshold uint64) (uint64, uint64) {
	return t.query(t.root, L, R, threshold)
}

// 正常逻辑下不会出现MAXUINT64，因为我们的区间是整个timespan
func (t *SegTree) query(cur *SegTreeNode, L, R, threshold uint64) (uint64, uint64) {
	if cur.seg_max < threshold {
		// 快速返回
		return MAXUINT64, 0
	}

	mid := (cur.L + cur.R) >> 1
	if cur.L == L && cur.R == R {
		if cur.seg_max >= threshold {
			if L == R {
				return L, cur.seg_max
			}

			if cur.lson != nil {
				return t.query(cur.lson, L, mid, threshold)
			}
			if cur.rson != nil {
				return t.query(cur.rson, mid+1, R, threshold)
			}
		}
		return MAXUINT64, 0
	}
	// 你要查询我左边区间的最大值
	if R <= mid {
		if cur.lson != nil {
			return t.query(cur.lson, L, R, threshold)
		}
		// 如果我左边区间都没有东西，对不起你找不到
		return MAXUINT64, 0
	}
	// 你要查询我右边区间的最大值
	if L > mid {
		if cur.rson != nil {
			return t.query(cur.rson, L, R, threshold)
		}
		// 如果我右边区间都没有东西，对不起你找不到
		return MAXUINT64, 0
	}

	// 跨区了
	var ans uint64 = MAXUINT64
	var ansLength uint64 = 0
	if cur.lson != nil {
		ans, ansLength = t.query(cur.lson, L, mid, threshold)
	}
	if ans != MAXUINT64 {
		// 左边找到了
		return ans, ansLength
	}
	// 左边没找到，去右边找
	if cur.rson != nil {
		return t.query(cur.rson, mid+1, R, threshold)
	}
	return MAXUINT64, 0
}
