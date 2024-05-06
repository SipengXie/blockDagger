package helper

import (
	"blockDagger/core/vm/evmtypes"
	dag "blockDagger/graph"
	multiversion "blockDagger/multiVersion"
	"blockDagger/rwset"
	"blockDagger/types"
	"fmt"
	"time"

	originCore "github.com/ledgerwatch/erigon/core"
	originTypes "github.com/ledgerwatch/erigon/core/types"
	originVm "github.com/ledgerwatch/erigon/core/vm"
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
	graph := generateGraph(taskMap, rwAccessedBy)

	taskEntry := types.NewTask(-1, 0, nil)
	taskEnd := types.NewTask(dag.MAXINT, 0, nil)

	graph.AddVertex(taskEntry)
	graph.AddVertex(taskEnd)

	for id, v := range graph.Vertices {
		if id == -1 || id == dag.MAXINT {
			continue
		}
		if v.InDegree == 0 {
			graph.AddEdge(-1, id)
		}
		if v.OutDegree == 0 {
			graph.AddEdge(id, dag.MAXINT)
		}
	}
	graph.GenerateProperties()
	return taskMap, graph, gVC
}

func PrepareforSingleBlock(blockNum uint64) (map[int]*types.Task, *dag.Graph, *multiversion.GlobalVersionChain, evmtypes.BlockContext) {
	ctx, dbTx, blkReader := PrepareEnv()

	block, header := GetBlockAndHeader(blkReader, ctx, dbTx, blockNum)
	originblkCtx := GetOriginBlockContext(blkReader, block, dbTx, header)
	blkCtx := GetBlockContext(blkReader, block, dbTx, header)
	txs := block.Transactions()

	txsWrapper := make([]*types.TransactionWrapper, 0)
	rwAccessedBy := rwset.NewRwAccessedBy()
	for i, tx := range txs {
		// 每个tx对应一个新的ibs
		ibs := GetState(params.MainnetChainConfig, dbTx, blockNum)
		fullstate := NewStateWithRwSets(ibs)
		rwset := generateRwSet(fullstate, tx, header, originblkCtx)
		if rwset == nil {
			continue
		}
		rwAccessedBy.Add(rwset, i)
		txsWrapper = append(txsWrapper, types.NewTransactionWrapper(tx, rwset, i))
	}

	//需要一个初始ibs
	ibs := GetState(params.MainnetChainConfig, dbTx, blockNum)
	tasks, graph, gvc := prepare(txsWrapper, rwAccessedBy, ibs)

	return tasks, graph, gvc, blkCtx
}

func PrepareForKBlocks(blockNum, k uint64) (map[int]*types.Task, *dag.Graph, *multiversion.GlobalVersionChain, evmtypes.BlockContext) {
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
	blkCtx := GetBlockContext(blkReader, block, dbTx, header)

	txsWrapper := make([]*types.TransactionWrapper, 0)
	rwAccessedBy := rwset.NewRwAccessedBy()
	for i, tx := range txs {
		// 每个tx对应一个新的ibs
		ibs := GetState(params.MainnetChainConfig, dbTx, blockNum)
		fullstate := NewStateWithRwSets(ibs)
		rwset := generateRwSet(fullstate, tx, header, originblkCtx)
		if rwset == nil {
			continue
		}
		rwAccessedBy.Add(rwset, i)
		txsWrapper = append(txsWrapper, types.NewTransactionWrapper(tx, rwset, i))
	}

	//需要一个初始ibs
	ibs := GetState(params.MainnetChainConfig, dbTx, blockNum)
	tasks, graph, gvc := prepare(txsWrapper, rwAccessedBy, ibs)
	fmt.Println("Transation count: ", len(txs), "Task count:", len(tasks))
	return tasks, graph, gvc, blkCtx
}

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
