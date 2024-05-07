package pipeline

import (
	dag "blockDagger/graph"
	"blockDagger/rwset"
	"blockDagger/schedule"
	"blockDagger/types"
)

type FLAG int

const (
	START FLAG = iota
	END
)

// 主线程到GVCline的信号
type TxwsMessage struct {
	Flag FLAG
	Txws []*types.TransactionWrapper
}

// GVCLine到GraphLine的信号
type TaskMapsAndAccessedBy struct {
	Flag         FLAG
	TaskMap      map[int]*types.Task
	RwAccessedBy *rwset.RwAccessedBy
}

// GraphLine到ScheduleLine的信号
type GraphMessage struct {
	Flag  FLAG
	Graph *dag.Graph
}

// Scheduleline到ExecuteLine信号
type ScheduleMessage struct {
	Flag       FLAG
	Processors []*schedule.Processor
	Makespan   uint64
}
