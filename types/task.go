package types

import (
	multiversion "blockDagger/multiVersion"
	"blockDagger/rwset"

	"github.com/ledgerwatch/erigon-lib/common"
	erigonTypes "github.com/ledgerwatch/erigon/core/types"
)

type Task struct {
	ID            int
	Cost          uint64
	Tx            erigonTypes.Transaction
	ReadVersions  map[common.Address]map[common.Hash]*multiversion.Version
	WriteVersions map[common.Address]map[common.Hash]*multiversion.Version
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

func (t *Task) Wait() {
	for _, versions := range t.ReadVersions {
		for _, version := range versions {
			for version.Status == multiversion.Pending {
				// fmt.Println(t.ID, "waiting for read version", version.Tid)
				continue
			}
		}
	}
}

func (t *Task) OutputReadVersion() string {
	tempStruct := TempSturct{
		Map: make(map[common.Address]map[string]interface{}),
	}

	for addr, versions := range t.ReadVersions {
		for hash, version := range versions {
			if _, ok := tempStruct.Map[addr]; !ok {
				tempStruct.Map[addr] = make(map[string]interface{})
			}
			tempStruct.Map[addr][rwset.DecodeHash(hash)] = version.GetVisible().Tid
		}
	}

	return tempStruct.ToString()
}
