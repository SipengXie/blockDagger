package helper

import (
	"blockDagger/graph"
	"blockDagger/rwset"
	"blockDagger/types"
)

func generateGraph(taskArray []*types.Task, rwAccessedBy *rwset.RwAccessedBy) *graph.Graph {
	Graph := graph.NewGraph()
	readBy := rwAccessedBy.ReadBy
	writeBy := rwAccessedBy.WriteBy

	// 先添加所有的点,task顺序就是他的id顺序
	for _, task := range taskArray {
		Graph.AddVertex(task)
	}

	for addr, wAccess := range writeBy {
		for hash := range wAccess {
			wTxs := writeBy.TxIds(addr, hash)
			rTxs := readBy.TxIds(addr, hash)
			// 由于有多版本，所以不存在写写冲突
			for _, rTx := range rTxs {
				for _, wTx := range wTxs {
					if rTx > wTx {
						// 构建数据依赖
						Graph.AddEdge(wTx, rTx)
						// 还应该据此建立多版本的依赖, 后读依赖先写
						taskArray[rTx].AddReadVersion(addr, hash, taskArray[wTx].WriteVersions[addr][hash])
					}
				}
			}
		}

	}
	return Graph
}
