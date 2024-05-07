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

// -------------------- Get LastBlockTail Version --------------------
// addr : 地址
// hash : BALANCE, NONCE, CODE, CODEHASH, ALIVE, SLOTS
func (mvs *GlobalVersionChain) GetLastBlockTailVersion(addr common.Address, hash common.Hash) *Version {
	cache, _ := mvs.ChainMap.LoadOrStore(addr, &sync.Map{})
	vc, _ := cache.(*sync.Map).LoadOrStore(hash, NewVersionChain())
	return vc.(*VersionChain).LastBlockTail
}

// TODO:我们不需要立刻实现它，在这里我们可以跳过
func writeBack(v *Version, innerState originEvmtypes.IntraBlockState, addr common.Address, hash common.Hash) {
	// switch hash {
	// case rwset.BALANCE:
	// 	innerState.SetBalance(addr, v.Data.(*uint256.Int))
	// case rwset.NONCE:
	// 	innerState.SetNonce(addr, v.Data.(uint64))
	// case rwset.CODE:
	// 	innerState.SetCode(addr, v.Data.([]byte))
	// case rwset.CODEHASH:
	// 	// innerState.SetCodeHash(addr, v.Data.(common.Hash))
	// case rwset.ALIVE:
	// 	if !v.Data.(bool) {
	// 		innerState.Selfdestruct(addr)
	// 	}
	// default:
	// 	innerState.SetState(addr, &hash, *v.Data.(*uint256.Int))
	// }
}

// 针对每一条VersionChain进行gc并落盘
func (mvs *GlobalVersionChain) GarbageCollection() {
	mvs.ChainMap.Range(func(key, value interface{}) bool {
		addr := key.(common.Address)
		cache := value.(*sync.Map)
		cache.Range(func(key, value interface{}) bool {
			hash := key.(common.Hash)
			vc := value.(*VersionChain)
			newhead := vc.GarbageCollection()
			writeBack(newhead, mvs.innerState, addr, hash)
			return true
		})
		return true
	})
}

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

func (gvc *GlobalVersionChain) DoPrefetch(addr common.Address, hash common.Hash) {
	v := gvc.GetLastBlockTailVersion(addr, hash)
	if v.Data != nil || v.Tid != -1 {
		// 如果v.Tid != -1，则代表被依赖的版本是上一个区块产生的，不需要预取【即必然有前序区块预取过了这条VC的-1版本】
		// 这个判断是帮助以后的GC的，如果这个数据已经被预取过了，就不用再预取了
		// 有助于流水线
		return
	}
	setVersion(v, gvc.innerState, addr, hash)
}

// 在每一次建图结束后调用一次
func (gvc *GlobalVersionChain) UpdateLastBlockTail() {
	gvc.ChainMap.Range(func(key, value interface{}) bool {
		// addr := key.(common.Address)
		cache := value.(*sync.Map)
		cache.Range(func(key, value interface{}) bool {
			// hash := key.(common.Hash)
			vc := value.(*VersionChain)
			vc.UpdateLastBlockTail()
			return true
		})
		return true
	})
}
