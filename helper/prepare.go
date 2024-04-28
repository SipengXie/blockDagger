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

	return taskList, graph, mvState
}
