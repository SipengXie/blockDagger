package helper

import (
	dag "blockDagger/graph"
	multiversion "blockDagger/multiVersion"
	"blockDagger/rwset"
	"blockDagger/types"
	"context"
	"fmt"

	"github.com/ledgerwatch/erigon-lib/kv"
	originTypes "github.com/ledgerwatch/erigon/core/types"
	originEvmTypes "github.com/ledgerwatch/erigon/core/vm/evmtypes"
	"github.com/ledgerwatch/erigon/params"
	"github.com/ledgerwatch/erigon/turbo/snapshotsync/freezeblocks"
)

// 从blockNum开始搜集k个区块的txws
// 返回txws, rwAccessedBy, [block, header, ibs] -> blockNum开始
func collectTxws(ctx context.Context, dbTx kv.Tx, blkReader *freezeblocks.BlockReader, blockNum, k uint64) (txws []*types.TransactionWrapper, rwAccessedBy *rwset.RwAccessedBy, block *originTypes.Block, header *originTypes.Header, ibs originEvmTypes.IntraBlockState) {
	txs := make(originTypes.Transactions, 0)

	// fetch transactions
	for i := blockNum; i < blockNum+k; i++ {
		block, _ := GetBlockAndHeader(blkReader, ctx, dbTx, i)
		txs = append(txs, block.Transactions()...)
	}

	// generating execution environment
	block, header = GetBlockAndHeader(blkReader, ctx, dbTx, blockNum)
	originblkCtx := GetOriginBlockContext(blkReader, block, dbTx, header)

	txws = make([]*types.TransactionWrapper, 0)
	rwAccessedBy = rwset.NewRwAccessedBy()
	for i, tx := range txs {
		// 每个tx对应一个新的ibs
		ibs := GetState(params.MainnetChainConfig, dbTx, blockNum)
		fullstate := NewStateWithRwSets(ibs)
		rwset := generateRwSet(fullstate, tx, header, originblkCtx)
		if rwset == nil {
			continue
		}
		rwAccessedBy.Add(rwset, i)
		txws = append(txws, types.NewTransactionWrapper(tx, rwset, i))
	}
	ibs = GetState(params.MainnetChainConfig, dbTx, blockNum)
	fmt.Println("Transation count: ", len(txs), "Task count:", len(txws))
	return
}

// 从blockNum开始搜集k个区块的txws，并根据此进行建图以及gvc处理
func PrepareBlocks(ctx context.Context, dbTx kv.Tx, blkReader *freezeblocks.BlockReader, blockNum, k uint64) (map[int]*types.Task, *dag.Graph, *multiversion.GlobalVersionChain, *originTypes.Block, *originTypes.Header) {
	txws, rwAccessedBy, block, header, ibs := collectTxws(ctx, dbTx, blkReader, blockNum, k)
	tasks, graph, gvc := prepare(txws, rwAccessedBy, ibs)
	return tasks, graph, gvc, block, header
}

// 从blockNum开始，先拿到k个区块的txws，然后根据groupNum进行分组
func PrepareBlockGroups(ctx context.Context, dbTx kv.Tx, blkReader *freezeblocks.BlockReader, blockNum, k, groupNum uint64) (txwsGroup [][]*types.TransactionWrapper, gvc *multiversion.GlobalVersionChain, block *originTypes.Block, header *originTypes.Header) {
	txws, _, block, header, ibs := collectTxws(ctx, dbTx, blkReader, blockNum, k)
	gvc = multiversion.NewGlobalVersionChain(ibs)

	// 将txws按顺序分成groupnum个组
	txwsGroup = make([][]*types.TransactionWrapper, groupNum)
	eachGroupNum := len(txws) / int(groupNum)
	fmt.Println("eachGroupNum: ", eachGroupNum)
	for i := 0; i < int(groupNum); i++ {
		txwsGroup[i] = txws[i*eachGroupNum : (i+1)*eachGroupNum]
	}
	// 将余下的txws分配到最后一组
	if len(txws)%int(groupNum) != 0 {
		txwsGroup[groupNum-1] = append(txwsGroup[groupNum-1], txws[eachGroupNum*int(groupNum):]...)
	}
	return
}
