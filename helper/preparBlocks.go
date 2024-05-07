package helper

import (
	"blockDagger/core/vm/evmtypes"
	dag "blockDagger/graph"
	multiversion "blockDagger/multiVersion"
	"blockDagger/types"
)

func PrepareforSingleBlock(blockNum uint64) (map[int]*types.Task, *dag.Graph, *multiversion.GlobalVersionChain, evmtypes.BlockContext) {
	txws, rwAccessedBy, blkCtx, ibs := prepareTxws(blockNum, 1)
	tasks, graph, gvc := prepare(txws, rwAccessedBy, ibs)

	return tasks, graph, gvc, blkCtx
}

func PrepareForKBlocks(blockNum, k uint64) (map[int]*types.Task, *dag.Graph, *multiversion.GlobalVersionChain, evmtypes.BlockContext) {
	txws, rwAccessedBy, blkCtx, ibs := prepareTxws(blockNum, k)
	tasks, graph, gvc := prepare(txws, rwAccessedBy, ibs)
	return tasks, graph, gvc, blkCtx
}

// 从blockNum开始k个block的Tx，分为groupNum个组
func PreparePipelineSim(blockNum, k, groupNum uint64) (taskMapGroup []map[int]*types.Task, graphGroup []*dag.Graph, blkCtx evmtypes.BlockContext) {
	txws, _, blkCtx, ibs := prepareTxws(blockNum, k)
	gvc := multiversion.NewGlobalVersionChain(ibs)
	// 将txws按顺序分成groupnum个组
	txwsGroup := make([][]*types.TransactionWrapper, groupNum)
	eachGroupNum := len(txws) / int(groupNum)
	for i := 0; i < int(groupNum); i++ {
		txwsGroup[i] = txws[i*eachGroupNum : (i+1)*eachGroupNum]
	}
	// 将余下的txws分配到最后一组
	if len(txws)%int(groupNum) != 0 {
		txwsGroup[groupNum-1] = append(txwsGroup[groupNum-1], txws[eachGroupNum*int(groupNum):]...)
	}
	taskMapGroup = make([]map[int]*types.Task, groupNum)
	graphGroup = make([]*dag.Graph, groupNum)
	for i := range txwsGroup {
		taskMapGroup[i], graphGroup[i] = prepareWithGVC(txwsGroup[i], gvc)
	}
	return
}

// 从blockNum开始k个block的Tx，分为groupNum个组
func PreparePipeline(blockNum, k, groupNum uint64) (txwsGroup [][]*types.TransactionWrapper, blkCtx evmtypes.BlockContext, gvc *multiversion.GlobalVersionChain) {
	txws, _, blkCtx, ibs := prepareTxws(blockNum, k)
	gvc = multiversion.NewGlobalVersionChain(ibs)
	// 将txws按顺序分成groupnum个组
	txwsGroup = make([][]*types.TransactionWrapper, groupNum)
	eachGroupNum := len(txws) / int(groupNum)
	for i := 0; i < int(groupNum); i++ {
		txwsGroup[i] = txws[i*eachGroupNum : (i+1)*eachGroupNum]
	}
	// 将余下的txws分配到最后一组
	if len(txws)%int(groupNum) != 0 {
		txwsGroup[groupNum-1] = append(txwsGroup[groupNum-1], txws[eachGroupNum*int(groupNum):]...)
	}
	return
}
