package helper

import (
	dag "blockDagger/graph"
	multiversion "blockDagger/multiVersion"
	"blockDagger/rwset"
	"blockDagger/types"
	"context"
	"fmt"
	"time"

	"github.com/ledgerwatch/erigon-lib/kv"
	originTypes "github.com/ledgerwatch/erigon/core/types"
	originEvmTypes "github.com/ledgerwatch/erigon/core/vm/evmtypes"
	"github.com/ledgerwatch/erigon/params"
	"github.com/ledgerwatch/erigon/turbo/snapshotsync/freezeblocks"
)

// input: TransactionWarppers, rwAccessedBy
// output Tasks, Graph, gvc[预取过的]
func prepare(txws []*types.TransactionWrapper, rwAccessedBy *rwset.RwAccessedBy, ibs originEvmTypes.IntraBlockState) (map[int]*types.Task, *dag.Graph, *multiversion.GlobalVersionChain) {
	gVC := multiversion.NewGlobalVersionChain(ibs)
	taskMap := make(map[int]*types.Task)
	st := time.Now()
	for _, txw := range txws {
		task := TransferTxToTask(*txw, gVC)
		taskMap[task.ID] = task
	}
	gVC.UpdateLastBlockTail()
	fmt.Println("Pre-Processing time: ", time.Since(st))
	// 在For循环结束后已经完成了GVC的生成
	st = time.Now()
	graph := GenerateGraph(taskMap, rwAccessedBy)
	fmt.Println("Graph generation time: ", time.Since(st))
	fmt.Println("Critical Path Length: ", graph.CriticalPathLen)
	return taskMap, graph, gVC
}

// 这是一个中间态函数，后面也许会被删除
func prepareWithGVC(txws []*types.TransactionWrapper, gVC *multiversion.GlobalVersionChain) (map[int]*types.Task, *dag.Graph) {
	// 这里是GVC更新流水线的例子
	rwAccessedBy := GenerateAccessedBy(txws)
	taskMap := make(map[int]*types.Task)
	st := time.Now()
	for _, txw := range txws {
		task := TransferTxToTask(*txw, gVC)
		taskMap[task.ID] = task
	}
	gVC.UpdateLastBlockTail()
	fmt.Println("Pre-Processing time: ", time.Since(st))

	// 这里是建图流水线的例子
	st = time.Now()
	graph := GenerateGraph(taskMap, rwAccessedBy)
	fmt.Println("Graph generation time: ", time.Since(st))
	fmt.Println("Critical Path Length: ", graph.CriticalPathLen)
	return taskMap, graph
}

// 返回带有RwSet的把k个区块攒在一起的TransactionWrapper
// rwAccessedBy不一定都会用，比如Pipeline就不会用，而是使用generateAccessedBy
func prepareTxws(blockNum, k uint64) (ctx context.Context, txws []*types.TransactionWrapper, rwAccessedBy *rwset.RwAccessedBy, db kv.RoDB, ibs originEvmTypes.IntraBlockState, blkReader *freezeblocks.BlockReader, block *originTypes.Block, header *originTypes.Header) {
	ctx, dbTx, blkReader, db := PrepareEnv()
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

// for the test using
func GenerateAccessedBy(txws []*types.TransactionWrapper) (rwAccessedBy *rwset.RwAccessedBy) {
	rwAccessedBy = rwset.NewRwAccessedBy()
	for _, txw := range txws {
		rwAccessedBy.Add(txw.RwSet, txw.Tid)
	}
	return
}

func TransactionCounting(blockNum, k uint64) {
	_, txws, _, _, _, _, _, _ := prepareTxws(blockNum, k)

	sum := 0
	sumRw := 0
	for _, txw := range txws {
		sum += txw.Tx.EncodingSize()
		for _, s := range txw.RwSet.ReadSet {
			sumRw += 20
			sumRw += 32 * len(s)
		}
	}
	fmt.Println("Transaction count: ", len(txws))
	fmt.Println("Transaction Size in total: ", sum)
	fmt.Println("RwSet Size in total: ", sumRw)
	fmt.Println("Transaction Size in average: ", sum/len(txws))
	fmt.Println("RwSet Size in average: ", sumRw/len(txws))
}
