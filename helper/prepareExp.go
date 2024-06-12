package helper

import (
	"blockDagger/rwset"
	"blockDagger/types"
	"context"
	"fmt"

	"github.com/ledgerwatch/erigon-lib/kv"
	originTypes "github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/params"
	"github.com/ledgerwatch/erigon/turbo/snapshotsync/freezeblocks"
)

func PrepareTransactions(ctx context.Context, dbTx kv.Tx, blkReader *freezeblocks.BlockReader, startNum, endNum, blockSize uint64, genRW bool) (txwsArray [][]*types.TransactionWrapper, startBlock *originTypes.Block, startHeader *originTypes.Header) {
	txws := make([]*types.TransactionWrapper, 0)
	totalTxs := 0
	// fetch transactions nad transfer to txws
	for i := startNum; i < endNum; i++ {
		// 产生预测环境
		block, header := GetBlockAndHeader(blkReader, ctx, dbTx, i)
		originblkCtx := GetOriginBlockContext(blkReader, block, dbTx, header)

		txs := block.Transactions()
		totalTxs += len(txs)
		for _, tx := range txs {
			if genRW {
				// 尽量提高预测准确率
				ibs := GetState(params.MainnetChainConfig, dbTx, i)
				fullstate := NewStateWithRwSets(ibs)
				rwset := generateRwSet(fullstate, tx, header, originblkCtx)
				if rwset == nil {
					continue
				}
				txws = append(txws, types.NewTransactionWrapper(tx, rwset, 0))
			} else {
				txws = append(txws, types.NewTransactionWrapper(tx, nil, 0))
			}
		}
	}

	fmt.Println("Total Number of Original Txs ", totalTxs, "Cannot Predict:", totalTxs-len(txws))

	// 设置txw的tid
	for i, txw := range txws {
		txw.Tid = i
	}

	// 将txws按照blockSize进行分组
	txwsArray = make([][]*types.TransactionWrapper, 0)
	for i := 0; i < len(txws); i += int(blockSize) {
		end := i + int(blockSize)
		if end > len(txws) {
			end = len(txws)
		}
		txwsArray = append(txwsArray, txws[i:end])
	}
	// 生成startBlock和startHeader
	startBlock, startHeader = GetBlockAndHeader(blkReader, ctx, dbTx, startNum)
	return
}

func GenerateAccessedBy(txws []*types.TransactionWrapper) (rwAccessedBy *rwset.RwAccessedBy) {
	rwAccessedBy = rwset.NewRwAccessedBy()
	for _, txw := range txws {
		rwAccessedBy.Add(txw.RwSet, txw.Tid)
	}
	return
}
