package helper

import (
	"blockDagger/graph"
	"blockDagger/rwset"
	"blockDagger/types"
)

// This generateGraph only used in the real environment
func generateGraph(taskMap map[int]*types.Task, rwAccessedBy *rwset.RwAccessedBy) *graph.Graph {
	Graph := graph.NewGraph()
	readBy := rwAccessedBy.ReadBy
	writeBy := rwAccessedBy.WriteBy

	// 先添加所有的点,task顺序就是他的id顺序
	for _, task := range taskMap {
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
						// 因为多个区块中TxID和TaskID不一定一一对应，我们需要将TaskArray转为TaskMap
						taskMap[rTx].AddReadVersion(addr, hash, taskMap[wTx].WriteVersions[addr][hash])
					}
				}
			}
		}

	}
	return Graph
}
