package helper

import (
	"blockDagger/core/vm/evmtypes"
	dag "blockDagger/graph"
	multiversion "blockDagger/multiVersion"
	"blockDagger/rwset"
	"blockDagger/types"
	"fmt"

	originTypes "github.com/ledgerwatch/erigon/core/types"
	originEvmTypes "github.com/ledgerwatch/erigon/core/vm/evmtypes"
	"github.com/ledgerwatch/erigon/params"
)

// input: TransactionWarppers, rwAccessedBy
// output Tasks, Graph, gvc[预取过的]
func prepare(txws []*types.TransactionWrapper, rwAccessedBy *rwset.RwAccessedBy, ibs originEvmTypes.IntraBlockState) (map[int]*types.Task, *dag.Graph, *multiversion.GlobalVersionChain) {
	gVC := multiversion.NewGlobalVersionChain(ibs)
	taskMap := make(map[int]*types.Task)
	for _, txw := range txws {
		task := transferTxToTask(*txw, gVC)
		taskMap[task.ID] = task
	}
	// 在For循环结束后已经完成了GVC的生成
	graph := generateGraph(taskMap, rwAccessedBy)
	return taskMap, graph, gVC
}

// 这是一个中间态函数，后面也许会被删除
func prepareWithGVC(txws []*types.TransactionWrapper, rwAccessedBy *rwset.RwAccessedBy, gVC *multiversion.GlobalVersionChain) (map[int]*types.Task, *dag.Graph) {
	// 这里是GVC更新流水线的例子
	taskMap := make(map[int]*types.Task)
	for _, txw := range txws {
		task := transferTxToTask(*txw, gVC)
		taskMap[task.ID] = task
	}
	gVC.UpdateLastBlockTail()

	// 这里是建图流水线的例子
	graph := generateGraph(taskMap, rwAccessedBy)
	return taskMap, graph
}

// 返回带有RwSet的TransactionWrapper，RwAccessedBy，执行用的blockNum的BlockContext，以及从blockNum开始的IntraBlockState
// rwAccessedBy不一定都会用，比如Pipeline就不会用，而是使用generateAccessedBy
func prepareTxws(blockNum, k uint64) (txws []*types.TransactionWrapper, rwAccessedBy *rwset.RwAccessedBy, blkCtx evmtypes.BlockContext, ibs originEvmTypes.IntraBlockState) {
	ctx, dbTx, blkReader := PrepareEnv()
	txs := make(originTypes.Transactions, 0)

	// fetch transactions
	for i := blockNum; i < blockNum+k; i++ {
		block, _ := GetBlockAndHeader(blkReader, ctx, dbTx, i)
		txs = append(txs, block.Transactions()...)
	}

	// generating execution environment
	block, header := GetBlockAndHeader(blkReader, ctx, dbTx, blockNum)
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
	blkCtx = GetBlockContext(blkReader, block, dbTx, header)
	fmt.Println("Transation count: ", len(txs), "Task count:", len(txws))
	return
}

// for the test using
func generateAccessedBy(txws []*types.TransactionWrapper) (rwAccessedBy *rwset.RwAccessedBy) {
	rwAccessedBy = rwset.NewRwAccessedBy()
	for _, txw := range txws {
		rwAccessedBy.Add(txw.RwSet, txw.Tid)
	}
	return
}

func TransactionCounting(blockNum, k uint64) {
	ctx, dbTx, blkReader := PrepareEnv()
	txs := make(originTypes.Transactions, 0)

	// fetch transactions
	for i := blockNum; i < blockNum+k; i++ {
		block, _ := GetBlockAndHeader(blkReader, ctx, dbTx, i)
		txs = append(txs, block.Transactions()...)
	}

	sum := 0
	for _, tx := range txs {
		sum += tx.EncodingSize()
	}
	fmt.Println("Transaction Size in total: ", sum)
	fmt.Println("Transaction Size in average: ", sum/len(txs))
}
