package multiversion

import (
	"sync"

	"github.com/ledgerwatch/erigon-lib/common"
)

// 由于这个MVState可能被多个核心同时访问，我们保险起见使用Sync.Map
// 不用分开单个搞，就按照RwSet里的那几个Hash来搞
type GlobalVersionChain struct {
	ChainMap sync.Map // ChainMap: version chain per record: addr -> hash -> *VersionChain
}

func NewGlobalVersionChain() *GlobalVersionChain {
	return &GlobalVersionChain{
		ChainMap: sync.Map{},
	}
}

// ------------------ Insert Version ---------------------
// addr : 地址
// hash : BALANCE, NONCE, CODE, CODEHASH, ALIVE, SLOTS
func (mvs *GlobalVersionChain) InsertVersion(addr common.Address, hash common.Hash, version *Version) {
	cache, _ := mvs.ChainMap.LoadOrStore(addr, sync.Map{})
	vc, _ := cache.(*sync.Map).LoadOrStore(hash, NewVersionChain())
	vc.(*VersionChain).InstallVersion(version)
}

// -------------------- Get Head Version --------------------
// addr : 地址
// hash : BALANCE, NONCE, CODE, CODEHASH, ALIVE, SLOTS
func (mvs *GlobalVersionChain) GetHeadVersion(addr common.Address, hash common.Hash) *Version {
	cache, _ := mvs.ChainMap.LoadOrStore(addr, sync.Map{})
	vc, _ := cache.(*sync.Map).LoadOrStore(hash, NewVersionChain())
	return vc.(*VersionChain).Head
}

// TODO: Garbage Collection WIP
func (mvs *GlobalVersionChain) GarbageCollection() {}
