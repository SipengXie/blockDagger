package multiversion

import (
	"blockDagger/rwset"
	"sync"

	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon-lib/common"
	originEvmtypes "github.com/ledgerwatch/erigon/core/vm/evmtypes" // 这里不需要改动，你总要一个原汁原味的IBS
)

// 由于这个MVState可能被多个processor同时访问，我们保险起见使用Sync.Map
// 不用分开单个搞，就按照RwSet里的那几个Hash来搞
type GlobalVersionChain struct {
	ChainMap   sync.Map                       // ChainMap: version chain per record: addr -> hash -> *VersionChain
	innerState originEvmtypes.IntraBlockState // 落盘的innerState
}

func NewGlobalVersionChain(ibs originEvmtypes.IntraBlockState) *GlobalVersionChain {
	return &GlobalVersionChain{
		ChainMap:   sync.Map{},
		innerState: ibs,
	}
}

// ------------------ Insert Version ---------------------
// addr : 地址
// hash : BALANCE, NONCE, CODE, CODEHASH, ALIVE, SLOTS
func (mvs *GlobalVersionChain) InsertVersion(addr common.Address, hash common.Hash, version *Version) {
	cache, _ := mvs.ChainMap.LoadOrStore(addr, &sync.Map{})
	vc, _ := cache.(*sync.Map).LoadOrStore(hash, NewVersionChain())
	vc.(*VersionChain).InstallVersion(version)
}

// -------------------- Get Head Version --------------------
// addr : 地址
// hash : BALANCE, NONCE, CODE, CODEHASH, ALIVE, SLOTS
func (mvs *GlobalVersionChain) GetHeadVersion(addr common.Address, hash common.Hash) *Version {
	cache, _ := mvs.ChainMap.LoadOrStore(addr, &sync.Map{})
	vc, _ := cache.(*sync.Map).LoadOrStore(hash, NewVersionChain())
	return vc.(*VersionChain).Head
}

// TODO: Garbage Collection WIP
func (mvs *GlobalVersionChain) GarbageCollection() {}

func setVersion(v *Version, innerState originEvmtypes.IntraBlockState, addr common.Address, hash common.Hash) {
	switch hash {
	case rwset.BALANCE:
		v.Data = innerState.GetBalance(addr)
	case rwset.NONCE:
		v.Data = innerState.GetNonce(addr)
	case rwset.CODE:
		v.Data = innerState.GetCode(addr)
	case rwset.CODEHASH:
		v.Data = innerState.GetCodeHash(addr)
	case rwset.ALIVE:
		v.Data = !innerState.Selfdestruct(addr)
	default:
		ret := uint256.NewInt(0)
		innerState.GetState(addr, &hash, ret)
		v.Data = ret
	}
}

func (gvc *GlobalVersionChain) SetHeadVersion(addr common.Address, hash common.Hash) {
	v := gvc.GetHeadVersion(addr, hash)
	if v.Data != nil {
		// 这个判断是帮助以后的GC的，如果这个数据已经被预取过了，就不用再预取了
		// 有助于流水线
		return
	}
	setVersion(v, gvc.innerState, addr, hash)
}

// 这个预取如果ibs支持并发就可以并发
// 这个函数暂且不用，我们使用SetHeadVersion来预取
func (gvc *GlobalVersionChain) Prefetch(rwSets *rwset.RWSet) {
	for addr, hashMap := range rwSets.ReadSet {
		for hash := range hashMap {
			gvc.SetHeadVersion(addr, hash)
		}
	}
	for addr, hashMap := range rwSets.WriteSet {
		for hash := range hashMap {
			gvc.SetHeadVersion(addr, hash)
		}
	}
}
