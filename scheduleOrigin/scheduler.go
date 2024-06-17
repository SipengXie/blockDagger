package scheduleOrigin

import (
	"blockDagger/graph"
	"container/heap"
	"sync"
)

type Method int

const MAXUINT64 = ^uint64(0) >> 1
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

func (s *Scheduler) taskPrioritize(m Method) (pq PriorityTaskQueue, tWrapMap map[int]*TaskWrapper) {
	tWrapMap = make(map[int]*TaskWrapper)

	for _, v := range s.Graph.Vertices {
		priority := uint64(0)
		switch m {
		case EFT:
			priority = v.Rank_u
		case CT:
			priority = v.CT
		}
		tWrap := &TaskWrapper{
			Task:     v.Task,
			Priority: priority,
			EST:      0,
			EFT:      0,
		}

		pq = append(pq, tWrap)
		tWrapMap[v.Task.ID] = tWrap
	}
	return
}

func (s *Scheduler) processorSelection(pq PriorityTaskQueue, tWrapMap map[int]*TaskWrapper) (processors []*Processor, makespan uint64) {
	heap.Init(&pq)
	processors = make([]*Processor, s.NumWorkers)
	for i := 0; i < s.NumWorkers; i++ {
		processors[i] = NewProcessor()
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
	makespan = tWrapMap[graph.MAXINT].EST
	return
}

func (s *Scheduler) selectBestProcessor(processors []*Processor, tWrap *TaskWrapper) {
	if tWrap.Task.ID == -1 || tWrap.Task.ID == graph.MAXINT {
		return
	}
	var pid int = 0                 // 记录选中哪一个processor
	var prev *TaskWrapperNode = nil // 记录插入位置的前一个节点
	var eft uint64 = MAXUINT64      // 记录最小EFT
	for id, p := range processors {
		tempPrev, tempEft := p.FindEFT(tWrap)
		if tempEft < eft {
			pid = id
			prev = tempPrev
			eft = tempEft
		}
	}
	tWrap.EFT = eft
	processors[pid].AddTask(tWrap, prev) // 添加任务
}

func (s *Scheduler) listSchedule(m Method, retProcessors *[]*Processor, retMakespan *uint64, wg *sync.WaitGroup) {
	defer wg.Done()
	// ------------------ task priority calculation ------------------
	pq, tWrapMap := s.taskPrioritize(m)

	// ------------------ processor selection ------------------
	processors, makespan := s.processorSelection(pq, tWrapMap)

	*retProcessors = append(*retProcessors, processors...)
	*retMakespan = makespan
}

func (s *Scheduler) pqSchedule(m Method, retProcessors *[]*Processor, retMakespan *uint64, wg *sync.WaitGroup) {
	defer wg.Done()
	tWrapMap := make(map[int]*TaskWrapper)
	var timespan uint64 = 0

	isCP := make(map[int]struct{})
	mapIndegree := make(map[int]uint)

	for _, v := range s.Graph.Vertices {
		var priority uint64
		switch m {
		case CPTL:
			priority = s.Graph.CriticalPathLen - v.Rank_d
		case CPOP:
			priority = v.Rank_d + v.Rank_u
		}

		if priority == s.Graph.CriticalPathLen {
			isCP[v.Task.ID] = struct{}{}
		}
		timespan += v.Task.Cost
		mapIndegree[v.Task.ID] = v.InDegree

		tWrap := &TaskWrapper{
			Task:     v.Task,
			Priority: priority,
			EST:      0,
			EFT:      0,
		}
		tWrapMap[v.Task.ID] = tWrap
	}

	processors := make([]*Processor, s.NumWorkers)
	for i := 0; i < s.NumWorkers; i++ {
		processors[i] = NewProcessor()
	}
	cpProcesser := processors[0]

	tEntry := tWrapMap[-1]
	pq := make(PriorityTaskQueue, 0)
	heap.Init(&pq)
	pq.Push(tEntry)

	for pq.Len() != 0 {
		tWrap := heap.Pop(&pq).(*TaskWrapper)
		if _, ok := isCP[tWrap.Task.ID]; ok && tWrap.Task.ID != -1 && tWrap.Task.ID != graph.MAXINT {
			prev, eft := cpProcesser.FindEFT(tWrap)
			tWrap.EFT = eft
			cpProcesser.AddTask(tWrap, prev)
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
	makespan := tWrapMap[graph.MAXINT].EST

	*retProcessors = append(*retProcessors, processors...)
	*retMakespan = makespan
}

func (s *Scheduler) Schedule() ([]*Processor, uint64) {
	retMakespans := make([]uint64, 4)
	retProcessors := make([][]*Processor, 4)
	for i := 0; i < 4; i++ {
		retProcessors[i] = make([]*Processor, 0)
	}
	var wg sync.WaitGroup
	wg.Add(4)
	go s.listSchedule(EFT, &retProcessors[0], &retMakespans[0], &wg)
	go s.pqSchedule(CPTL, &retProcessors[1], &retMakespans[1], &wg)
	go s.listSchedule(CT, &retProcessors[2], &retMakespans[2], &wg)
	go s.pqSchedule(CPOP, &retProcessors[3], &retMakespans[3], &wg)
	wg.Wait()

	makespan := retMakespans[0]
	processors := retProcessors[0]

	for i := 1; i < 4; i++ {
		if retMakespans[i] < makespan {
			makespan = retMakespans[i]
			processors = retProcessors[i]
		}
	}
	return processors, makespan
}
