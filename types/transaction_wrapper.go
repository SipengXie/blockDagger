package types

import (
	"blockDagger/rwset"

	erigonTypes "github.com/ledgerwatch/erigon/core/types"
)

type TransactionWrapper struct {
	Tx    erigonTypes.Transaction
	RwSet *rwset.RWSet
	Tid   int
}

func NewTransactionWrapper(tx erigonTypes.Transaction, rwSet *rwset.RWSet, tid int) *TransactionWrapper {
	return &TransactionWrapper{
		Tx:    tx,
		RwSet: rwSet,
		Tid:   tid,
	}
}
