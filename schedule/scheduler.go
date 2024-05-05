package schedule

import (
	"blockDagger/graph"
	"container/heap"
)

type Method int

const (
	EFT Method = iota
	CPTL
	CT
	CPOP
)

type Scheduler struct {
	Graph      *graph.Graph
	NumWorkers int
}

func NewScheduler(graph *graph.Graph, numWorkers int) *Scheduler {
	return &Scheduler{
		Graph:      graph,
		NumWorkers: numWorkers,
	}
}

func (s *Scheduler) taskPrioritize(m Method) (timespan uint64, pq PriorityTaskQueue, tWrapMap map[int]*TaskWrapper) {
	tWrapMap = make(map[int]*TaskWrapper)

	for _, v := range s.Graph.Vertices {
		priority := uint64(0)
		switch m {
		case EFT:
			priority = v.Rank_u
		case CPTL:
			priority = s.Graph.CriticalPathLen - v.Rank_d
		case CT:
			priority = v.CT
		}
		tWrap := &TaskWrapper{
			Task:     v.Task,
			Priority: priority,
		}

		pq = append(pq, tWrap)
		tWrapMap[v.Task.ID] = tWrap

		timespan += v.Task.Cost
	}
	return
}

func (s *Scheduler) processorSelection(timespan uint64, pq PriorityTaskQueue, tWrapMap map[int]*TaskWrapper) (processors []*Processor, makespan uint64) {
	heap.Init(&pq)
	processors = make([]*Processor, s.NumWorkers)
	for i := 0; i < s.NumWorkers; i++ {
		processors[i] = NewProcessor(timespan)
	}

	for pq.Len() > 0 {
		tWrap := heap.Pop(&pq).(*TaskWrapper)
		s.selectBestProcessor(processors, tWrap)
		// 更新后继任务的EST
		for succID := range s.Graph.AdjacencyMap[tWrap.Task.ID] {
			succTwrap := tWrapMap[succID]
			succTwrap.EST = max(succTwrap.EST, tWrap.EFT)
		}
	}
	makespan = tWrapMap[len(tWrapMap)-2].EFT
	return
}

func (s *Scheduler) selectBestProcessor(processors []*Processor, tWrap *TaskWrapper) {
	var pid int = 0       // 记录选中哪一个processor
	var st, length uint64 // 记录选中该processor中哪一个slot

	var eft uint64 = MAXUINT64 // 记录最小EFT
	for id, p := range processors {
		// TODO:选择的时候如果多个EFT一样，我们尽量选择不产生新Slot的processor
		// 在我们的实现中好像没区别，都要做一个modify和add
		tempSt, tempLength, tempEft := p.FindEFT(tWrap.EST, tWrap.Task.Cost)
		// !!注意在CT的情况下，由于processor是同质的，所以CT没有区别，最终还是比较EFT
		if tempEft < eft {
			pid = id
			st = tempSt
			length = tempLength
			eft = tempEft
		}
		// 下面代码实现了这个TODO，但先不急
		// else if tempEft == eft {
		// 	if tempSt >= tWrap.EST {
		// 		// 这个slot不会产生新的slot
		// 		pid = id
		// 		st = tempSt
		// 		length = tempLength
		// 		eft = tempEft
		// 	}
		// }
	}
	tWrap.EFT = eft
	processors[pid].AddTask(tWrap, st, length) // 添加任务
}

func (s *Scheduler) ListSchedule(m Method) ([]*Processor, uint64) {
	// ------------------ task priority calculation ------------------
	timespan, pq, tWrapMap := s.taskPrioritize(m)

	// ------------------ processor selection ------------------
	return s.processorSelection(timespan, pq, tWrapMap)
}

func (s *Scheduler) CPOPSchedule() ([]*Processor, uint64) {

	tWrapMap := make(map[int]*TaskWrapper)
	var timespan uint64 = 0

	isCP := make(map[int]struct{})
	mapIndegree := make(map[int]uint)

	for _, v := range s.Graph.Vertices {
		priority := v.Rank_d + v.Rank_u
		if priority == s.Graph.CriticalPathLen {
			isCP[v.Task.ID] = struct{}{}
		}
		timespan += v.Task.Cost
		mapIndegree[v.Task.ID] = v.InDegree

		tWrap := &TaskWrapper{
			Task:     v.Task,
			Priority: priority,
		}
		tWrapMap[v.Task.ID] = tWrap
	}

	processors := make([]*Processor, s.NumWorkers)
	for i := 0; i < s.NumWorkers; i++ {
		processors[i] = NewProcessor(timespan)
	}
	cpProcesser := processors[0]

	tEntry := tWrapMap[-1]
	pq := make(PriorityTaskQueue, 0)
	heap.Init(&pq)
	pq.Push(tEntry)

	for pq.Len() != 0 {
		tWrap := heap.Pop(&pq).(*TaskWrapper)
		if _, ok := isCP[tWrap.Task.ID]; ok {
			st, length, eft := cpProcesser.FindEFT(tWrap.EST, tWrap.Task.Cost)
			tWrap.EFT = eft
			cpProcesser.AddTask(tWrap, st, length)
		} else {
			s.selectBestProcessor(processors, tWrap)
		}
		// 更新后继任务的EST以及入度
		for succID := range s.Graph.AdjacencyMap[tWrap.Task.ID] {
			succTwrap := tWrapMap[succID]
			succTwrap.EST = max(succTwrap.EST, tWrap.EFT)
			mapIndegree[succID]--
			if mapIndegree[succID] == 0 {
				pq.Push(succTwrap)
			}
		}
	}
	makespan := tWrapMap[len(tWrapMap)-2].EFT
	return processors, makespan
}

// Deprecated
// func (s *Scheduler) TopoSchedule() ([]*Processor, uint64) {
// 	// Topo Sort, get levels
// 	mapIndegree := make(map[int]uint)
// 	degreeZero := make([]int, 0)
// 	for _, v := range s.Graph.Vertices {
// 		mapIndegree[v.Task.ID] = v.InDegree
// 		if v.InDegree == 0 {
// 			degreeZero = append(degreeZero, v.Task.ID)
// 		}
// 	}
// 	levels := make([][]int, 0)
// 	levels = append(levels, degreeZero)

// 	for {
// 		newDegreeZero := make([]int, 0)
// 		for _, vid := range degreeZero {
// 			for succId := range s.Graph.AdjacencyMap[vid] {
// 				mapIndegree[succId]--
// 				if mapIndegree[succId] == 0 {
// 					newDegreeZero = append(newDegreeZero, succId)
// 				}
// 			}
// 		}
// 		degreeZero = newDegreeZero
// 		if len(degreeZero) == 0 {
// 			break
// 		} else {
// 			levels = append(levels, degreeZero)
// 		}
// 	}

// 	// converse Int in levels into types.Task
// 	levelsTask := make([][]*types.Task, 0)
// 	for _, level := range levels {
// 		temp := make([]*types.Task, 0)
// 		for _, id := range level {
// 			temp = append(temp, s.Graph.Vertices[id].Task)
// 		}
// 		levelsTask = append(levelsTask, temp)
// 	}

// 	// A greedy algorithm to schedule tasks in levels
// 	processors := make([]*Processor, s.NumWorkers)
// 	for i := 0; i < s.NumWorkers; i++ {
// 		processors[i] = NewProcessor(s.Graph.CriticalPathLen)
// 	}

// 	// 顺便计算makespan
// 	var makespan uint64 = 0
// 	for _, level := range levelsTask {
// 		// TODO: 如果|level| < workers, 我们可以不走贪心
// 		groups, maxSum := greedy.Greedy(level, s.NumWorkers)
// 		makespan += maxSum
// 		for i, group := range groups {
// 			// TODO: 这里需要更改
// 			// processors[i].Tasks = append(processors[i].Tasks, group...)
// 		}
// 	}

// 	return processors, makespan
// }
