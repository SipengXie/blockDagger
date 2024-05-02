package helper

import (
	"blockDagger/graph"
	multiversion "blockDagger/multiVersion"
	"blockDagger/rwset"
	"blockDagger/types"

	"github.com/ledgerwatch/erigon/core/vm/evmtypes"
)

// input: TransactionWarppers, rwAccessedBy
// output Tasks, Graph, gvc[预取过的]
func Prepare(txws []*types.TransactionWrapper, rwAccessedBy *rwset.RwAccessedBy, ibs evmtypes.IntraBlockState) ([]*types.Task, *graph.Graph, *multiversion.GlobalVersionChain) {
	gVC := multiversion.NewGlobalVersionChain(ibs)
	taskList := make([]*types.Task, len(txws))
	for i, txw := range txws {
		taskList[i] = transferTxToTask(*txw, gVC)
	}
	graph := generateGraph(taskList, rwAccessedBy)

	taskEntry := types.NewTask(-1, 0, nil)
	taskEnd := types.NewTask(len(txws), 0, nil)

	graph.AddVertex(taskEntry)
	graph.AddVertex(taskEnd)

	for id, v := range graph.Vertices {
		if id == -1 || id == len(txws) {
			continue
		}
		if v.InDegree == 0 {
			graph.AddEdge(-1, id)
		} else if v.OutDegree == 0 {
			graph.AddEdge(id, len(txws))
		}
	}

	return taskList, graph, gVC
}
