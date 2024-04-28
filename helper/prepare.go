package helper

import (
	"blockDagger/graph"
	multiversion "blockDagger/multiVersion"
	"blockDagger/rwset"
	"blockDagger/types"
)

// input: TransactionWarppers, rwAccessedBy
// output Tasks, Graph, mvState
func Prepare(txws []*types.TransactionWrapper, rwAccessedBy *rwset.RwAccessedBy) ([]*types.Task, *graph.Graph, *multiversion.MultiVersionState) {
	mvState := multiversion.NewMultiVersionState()
	taskList := make([]*types.Task, len(txws))
	for i, txw := range txws {
		taskList[i] = transferTxToTask(*txw, mvState)
	}
	graph := generateGraph(taskList, rwAccessedBy)
	return taskList, graph, mvState
}
