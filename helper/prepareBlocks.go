package helper

import (
	dag "blockDagger/graph"
	multiversion "blockDagger/multiVersion"
	"blockDagger/types"
	"context"
	"fmt"

	"github.com/ledgerwatch/erigon-lib/kv"
	originTypes "github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/turbo/snapshotsync/freezeblocks"
)

func PrepareforSingleBlock(blockNum uint64) (context.Context, map[int]*types.Task, *dag.Graph, *multiversion.GlobalVersionChain, kv.RoDB, *freezeblocks.BlockReader, *originTypes.Block, *originTypes.Header) {
	ctx, txws, rwAccessedBy, db, ibs, blkReader, blk, header := prepareTxws(blockNum, 1)
	tasks, graph, gvc := prepare(txws, rwAccessedBy, ibs)

	return ctx, tasks, graph, gvc, db, blkReader, blk, header
}

func PrepareForKBlocks(blockNum, k uint64) (context.Context, map[int]*types.Task, *dag.Graph, *multiversion.GlobalVersionChain, kv.RoDB, *freezeblocks.BlockReader, *originTypes.Block, *originTypes.Header) {
	ctx, txws, rwAccessedBy, db, ibs, blkReader, blk, header := prepareTxws(blockNum, k)
	tasks, graph, gvc := prepare(txws, rwAccessedBy, ibs)
	return ctx, tasks, graph, gvc, db, blkReader, blk, header
}

// 从blockNum开始k个block的Tx，分为groupNum个组
func PreparePipelineSim(blockNum, k, groupNum uint64) (ctx context.Context, taskMapGroup []map[int]*types.Task, graphGroup []*dag.Graph, db kv.RoDB, blkReader *freezeblocks.BlockReader, blk *originTypes.Block, header *originTypes.Header) {
	ctx, txws, _, db, ibs, blkReader, blk, header := prepareTxws(blockNum, k)
	gvc := multiversion.NewGlobalVersionChain(ibs)
	// 将txws按顺序分成groupnum个组
	txwsGroup := make([][]*types.TransactionWrapper, groupNum)
	eachGroupNum := len(txws) / int(groupNum)
	fmt.Println("eachGroupNum: ", eachGroupNum)
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
func PreparePipeline(blockNum, k, groupNum uint64) (ctx context.Context, txwsGroup [][]*types.TransactionWrapper, db kv.RoDB, gvc *multiversion.GlobalVersionChain, blkReader *freezeblocks.BlockReader, blk *originTypes.Block, header *originTypes.Header) {
	ctx, txws, _, db, ibs, blkReader, blk, header := prepareTxws(blockNum, k)
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
