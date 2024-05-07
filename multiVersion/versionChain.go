package multiversion

// TODO: 我们需要实现Garbage Collection
type VersionChain struct {
	Head          *Version
	Tail          *Version // 辅助产生LastBlockTail
	LastBlockTail *Version // 用于跨Block Schedule使用
}

func NewVersionChain() *VersionChain {
	head := NewVersion(nil, -1, Committed)
	return &VersionChain{
		Head:          head, // an dummy head which means its from the stateSnapshot
		Tail:          head,
		LastBlockTail: head,
	}
}

func (vc *VersionChain) InstallVersion(iv *Version) {
	cur_v := vc.Head
	for {
		if cur_v == nil {
			break
		}
		cur_v = cur_v.InsertOrNext(iv)
	}
	if iv.Next == nil {
		vc.Tail = iv
	}
}

func (vc *VersionChain) UpdateLastBlockTail() {
	vc.LastBlockTail = vc.Tail
}

// 找到最后一个commit的版本，并设置为Head(data, -1, committed)，同时返回这个版本用于落盘
// Find the last committed version
func (vc *VersionChain) GarbageCollection() *Version {
	// 从Tail开始，找到最后一个committed的版本
	cur_v := vc.Tail
	for {
		if cur_v == nil {
			break
		}
		if cur_v.Status == Committed {
			break
		}
		cur_v = cur_v.Prev
	}
	// 生成新的Head
	newHead := NewVersion(nil, -1, Committed)
	if cur_v != nil {
		newHead.Data = cur_v.Data
	}
	vc.Head = newHead
	vc.Tail = newHead
	vc.LastBlockTail = newHead
	// 余下的版本全部删除，依靠golang自己的GC机制吧
	return newHead
}
