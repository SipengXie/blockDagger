package types

import (
	multiversion "blockDagger/multiVersion"

	"github.com/ledgerwatch/erigon-lib/common"
	erigonTypes "github.com/ledgerwatch/erigon/core/types"
)

type Task struct {
	ID            int
	Cost          uint64
	Tx            erigonTypes.Transaction
	ReadVersions  map[common.Address]map[common.Hash]*multiversion.Version
	WriteVersions map[common.Address]map[common.Hash]*multiversion.Version

	Priority uint64
}

func NewTask(id int, cost uint64, tx erigonTypes.Transaction) *Task {
	return &Task{
		ID:            id,
		Cost:          cost,
		Tx:            tx,
		ReadVersions:  make(map[common.Address]map[common.Hash]*multiversion.Version),
		WriteVersions: make(map[common.Address]map[common.Hash]*multiversion.Version),
	}
}

func (t *Task) AddReadVersion(addr common.Address, hash common.Hash, version *multiversion.Version) {
	if _, ok := t.ReadVersions[addr]; !ok {
		t.ReadVersions[addr] = make(map[common.Hash]*multiversion.Version)
	}
	t.ReadVersions[addr][hash] = version
}

func (t *Task) AddWriteVersion(addr common.Address, hash common.Hash, version *multiversion.Version) {
	if _, ok := t.WriteVersions[addr]; !ok {
		t.WriteVersions[addr] = make(map[common.Hash]*multiversion.Version)
	}
	t.WriteVersions[addr][hash] = version
}
