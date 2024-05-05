package helper

import (
	"blockDagger/rwset"

	originCore "github.com/ledgerwatch/erigon/core"
	"github.com/ledgerwatch/erigon/core/types"
	originVm "github.com/ledgerwatch/erigon/core/vm"
	originEvmTypes "github.com/ledgerwatch/erigon/core/vm/evmtypes"
	"github.com/ledgerwatch/erigon/params"
)

func generateRwSet(fulldb *StateWithRwSets, tx types.Transaction, header *types.Header, blkCtx originEvmTypes.BlockContext) *rwset.RWSet {
	rwSet := rwset.NewRWSet()
	fulldb.SetRWSet(rwSet)
	evm := originVm.NewEVM(blkCtx, originEvmTypes.TxContext{}, fulldb, params.MainnetChainConfig, originVm.Config{})
	msg, err := tx.AsMessage(*types.LatestSigner(params.MainnetChainConfig), header.BaseFee, evm.ChainRules())
	if err != nil {
		return nil
	}
	// Skip the nonce check!
	msg.SetCheckNonce(false)
	txCtx := originCore.NewEVMTxContext(msg)
	evm.TxContext = txCtx

	_, err = originCore.ApplyMessage(evm, msg, new(originCore.GasPool).AddGas(header.GasLimit), false /* refunds */, false /* gasBailout */)

	if err != nil {
		return nil
	}
	return rwSet
}
