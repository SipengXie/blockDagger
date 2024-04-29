package schedule

import (
	"blockDagger/graph"
	"container/heap"
)

type ListScheduler struct {
	Graph      *graph.Graph
	NumWorkers int
}

func NewListScheduler(graph *graph.Graph, numWorkers int) *ListScheduler {
	return &ListScheduler{
		Graph:      graph,
		NumWorkers: numWorkers,
	}
}

func (s *ListScheduler) EFTScheduling() {

	// ------------------ task priority calculation ------------------
	var timespan uint64
	pq := make(PriorityTaskQueue, 0)
	tWrapMap := make(map[int]*TaskWrapper)

	for _, v := range s.Graph.Vertices {
		tWrap := &TaskWrapper{
			Task:     v.Task,
			Priority: v.Rank_u,
		}

		pq = append(pq, tWrap)
		tWrapMap[v.Task.ID] = tWrap

		timespan += v.Task.Cost
	}
	heap.Init(&pq)

	// ------------------ processor selection ------------------
	processors := make([]*Processor, s.NumWorkers)
	for i := 0; i < s.NumWorkers; i++ {
		processors[i] = NewProcessor(timespan)
	}

	for pq.Len() > 0 {
		tWrap := heap.Pop(&pq).(*TaskWrapper)

		var pid int = 0       // 记录选中哪一个processor
		var st, length uint64 // 记录选中该processor中哪一个slot

		var eft uint64 = MAXUINT64 // 记录最小EFT

		for id, p := range processors {
			tempSt, tempLength, tempEft := p.FindEFT(tWrap.EST, tWrap.Task.Cost)
			if tempEft < eft {
				pid = id
				st = tempSt
				length = tempLength
				eft = tempEft
			}
		}
		tWrap.EFT = eft
		processors[pid].AddTask(tWrap, st, length) // 添加任务

		// 更新后继任务的EST
		for succID := range s.Graph.AdjacencyMap[tWrap.Task.ID] {
			succTwrap := tWrapMap[succID]
			succTwrap.EST = max(succTwrap.EST, tWrap.EFT)
		}
	}
}
