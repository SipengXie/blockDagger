package multiversion

// TODO: 我们需要实现Garbage Collection
type VersionChain struct {
	Head *Version
}

func NewVersionChain() *VersionChain {
	return &VersionChain{
		Head: NewVersion(nil, -1, Committed), // an dummy head which means its from the stateSnapshot
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
}

// TODO: WIP
func (vc *VersionChain) GarbageCollection() {}
