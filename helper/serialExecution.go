package helper

import (
	"time"

	originCore "github.com/ledgerwatch/erigon/core"
	originTypes "github.com/ledgerwatch/erigon/core/types"
	originVm "github.com/ledgerwatch/erigon/core/vm"
	originEvmTypes "github.com/ledgerwatch/erigon/core/vm/evmtypes"
	"github.com/ledgerwatch/erigon/params"
)

func SerialExecutionTime(blockNum uint64) time.Duration {
	ctx, dbTx, blkReader := PrepareEnv()

	block, header := GetBlockAndHeader(blkReader, ctx, dbTx, blockNum)
	originblkCtx := GetOriginBlockContext(blkReader, block, dbTx, header)
	ibs := GetState(params.MainnetChainConfig, dbTx, blockNum)
	txs := block.Transactions()

	evm := originVm.NewEVM(originblkCtx, originEvmTypes.TxContext{}, ibs, params.MainnetChainConfig, originVm.Config{})

	st := time.Now()
	for _, tx := range txs {
		msg, _ := tx.AsMessage(*originTypes.LatestSigner(params.MainnetChainConfig), header.BaseFee, evm.ChainRules())

		// Skip the nonce check!
		msg.SetCheckNonce(false)
		txCtx := originCore.NewEVMTxContext(msg)
		evm.TxContext = txCtx

		originCore.ApplyMessage(evm, msg, new(originCore.GasPool).AddGas(header.GasLimit), true /* refunds */, false /* gasBailout */)
	}

	return time.Since(st)
}

func SerialExecutionKBlocks(blockNum, k uint64) time.Duration {
	ctx, dbTx, blkReader := PrepareEnv()

	txs := make(originTypes.Transactions, 0)

	// fetch transactions
	for i := blockNum; i < blockNum+k; i++ {
		block, _ := GetBlockAndHeader(blkReader, ctx, dbTx, i)
		txs = append(txs, block.Transactions()...)
	}

	block, header := GetBlockAndHeader(blkReader, ctx, dbTx, blockNum)
	originblkCtx := GetOriginBlockContext(blkReader, block, dbTx, header)
	ibs := GetState(params.MainnetChainConfig, dbTx, blockNum)

	evm := originVm.NewEVM(originblkCtx, originEvmTypes.TxContext{}, ibs, params.MainnetChainConfig, originVm.Config{})

	st := time.Now()
	for _, tx := range txs {
		msg, _ := tx.AsMessage(*originTypes.LatestSigner(params.MainnetChainConfig), header.BaseFee, evm.ChainRules())

		// Skip the nonce check!
		msg.SetCheckNonce(false)
		txCtx := originCore.NewEVMTxContext(msg)
		evm.TxContext = txCtx

		originCore.ApplyMessage(evm, msg, new(originCore.GasPool).AddGas(header.GasLimit), true /* refunds */, false /* gasBailout */)
	}

	return time.Since(st)
}
