package helper

import (
	dag "blockDagger/graph"
	"blockDagger/rwset"
	"blockDagger/types"
)

// This GenerateGraph only used in the real environment
func GenerateGraph(taskMap map[int]*types.Task, rwAccessedBy *rwset.RwAccessedBy) *dag.Graph {
	graph := dag.NewGraph()
	readBy := rwAccessedBy.ReadBy
	writeBy := rwAccessedBy.WriteBy
	for _, task := range taskMap {
		graph.AddVertex(task)
	}

	for addr, wAccess := range writeBy {
		for hash := range wAccess {
			wTxs := writeBy.TxIds(addr, hash)
			rTxs := readBy.TxIds(addr, hash)
			// 由于有多版本，所以不存在写写冲突
			for _, rTx := range rTxs {
				for _, wTx := range wTxs {
					if wTx >= rTx {
						break
					}
					// 构建数据依赖
					graph.AddEdge(wTx, rTx)
					// 还应该据此建立多版本的依赖, 后读依赖先写
					// 因为多个区块中TxID和TaskID不一定一一对应，我们需要将TaskArray转为TaskMap
					taskMap[rTx].AddReadVersion(addr, hash, taskMap[wTx].WriteVersions[addr][hash])
				}

			}
		}

	}

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

	return graph
}
