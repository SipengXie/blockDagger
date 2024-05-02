package state

import (
	"blockDagger/rwset"

	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon-lib/common"
)

type localWrite struct {
	storage map[common.Address]map[common.Hash]interface{}
}

func newLocalWrite() *localWrite {
	return &localWrite{
		storage: make(map[common.Address]map[common.Hash]interface{}),
	}
}

// --------------------- Getters ------------------------------

func (lw *localWrite) getBalance(addr common.Address) (*uint256.Int, bool) {
	if _, ok := lw.storage[addr]; !ok {
		return nil, false
	}
	if val, ok := lw.storage[addr][rwset.BALANCE]; ok {
		return val.(*uint256.Int), true
	}
	return nil, false
}

func (lw *localWrite) getNonce(addr common.Address) (uint64, bool) {
	if _, ok := lw.storage[addr]; !ok {
		return 0, false
	}
	if val, ok := lw.storage[addr][rwset.NONCE]; ok {
		return val.(uint64), true
	}
	return 0, false
}

func (lw *localWrite) getCode(addr common.Address) ([]byte, bool) {
	if _, ok := lw.storage[addr]; !ok {
		return nil, false
	}
	if val, ok := lw.storage[addr][rwset.CODE]; ok {
		return val.([]byte), true
	}
	return nil, false
}

func (lw *localWrite) getCodeHash(addr common.Address) (common.Hash, bool) {
	if _, ok := lw.storage[addr]; !ok {
		return common.Hash{}, false
	}
	if val, ok := lw.storage[addr][rwset.CODEHASH]; ok {
		return val.(common.Hash), true
	}
	return common.Hash{}, false
}

func (lw *localWrite) getAlive(addr common.Address) (bool, bool) {
	if _, ok := lw.storage[addr]; !ok {
		return false, false
	}
	if val, ok := lw.storage[addr][rwset.ALIVE]; ok {
		return val.(bool), true
	}
	return false, false
}

func (lw *localWrite) getSlot(addr common.Address, hash common.Hash) (*uint256.Int, bool) {
	if _, ok := lw.storage[addr]; !ok {
		return nil, false
	}
	if val, ok := lw.storage[addr][hash]; ok {
		return val.(*uint256.Int), true
	}
	return nil, false
}

// // A functional process

// func (lw *localWrite) exist(addr common.Address) bool {
// 	_, ok := lw.storage[addr]
// 	return ok
// }

// ------------------------- Setters ------------------------------

func (lw *localWrite) setBalance(addr common.Address, balance *uint256.Int) {
	if _, ok := lw.storage[addr]; !ok {
		lw.storage[addr] = make(map[common.Hash]interface{})
	}
	lw.storage[addr][rwset.BALANCE] = balance
}

func (lw *localWrite) setNonce(addr common.Address, nonce uint64) {
	if _, ok := lw.storage[addr]; !ok {
		lw.storage[addr] = make(map[common.Hash]interface{})
	}
	lw.storage[addr][rwset.NONCE] = nonce
}

func (lw *localWrite) setCode(addr common.Address, code []byte) {
	if _, ok := lw.storage[addr]; !ok {
		lw.storage[addr] = make(map[common.Hash]interface{})
	}
	lw.storage[addr][rwset.CODE] = code
}

func (lw *localWrite) setCodeHash(addr common.Address, codeHash common.Hash) {
	if _, ok := lw.storage[addr]; !ok {
		lw.storage[addr] = make(map[common.Hash]interface{})
	}
	lw.storage[addr][rwset.CODEHASH] = codeHash
}

func (lw *localWrite) setAlive(addr common.Address, alive bool) {
	if _, ok := lw.storage[addr]; !ok {
		lw.storage[addr] = make(map[common.Hash]interface{})
	}
	lw.storage[addr][rwset.ALIVE] = alive
}

func (lw *localWrite) setSlot(addr common.Address, hash common.Hash, slot *uint256.Int) {
	if _, ok := lw.storage[addr]; !ok {
		lw.storage[addr] = make(map[common.Hash]interface{})
	}
	lw.storage[addr][hash] = slot
}

// A functional procedure

func (lw *localWrite) createAccount(addr common.Address) {
	lw.setBalance(addr, uint256.NewInt(0))
	lw.setNonce(addr, 0)
	lw.setCode(addr, []byte{})
	lw.setCodeHash(addr, common.Hash{})
	lw.setAlive(addr, true)
}

func (lw *localWrite) contains(addr common.Address, hash common.Hash) bool {
	if _, ok := lw.storage[addr]; !ok {
		return false
	}
	_, ok := lw.storage[addr][hash]
	return ok
}
