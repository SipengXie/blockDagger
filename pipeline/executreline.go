package pipeline

import (
	"blockDagger/core/vm/evmtypes"
	"fmt"
	"sync"
)

type ExecuteLine struct {
	BlkCtx    evmtypes.BlockContext
	Wg        *sync.WaitGroup
	InputChan chan *ScheduleMessage
}

func NewExecuteLine(blkCtx evmtypes.BlockContext, wg *sync.WaitGroup, in chan *ScheduleMessage) *ExecuteLine {
	return &ExecuteLine{
		BlkCtx:    blkCtx,
		Wg:        wg,
		InputChan: in,
	}
}

func (e *ExecuteLine) Run() {
	for input := range e.InputChan {
		if input.Flag == END {
			e.Wg.Done()
			return
		}

		processors := input.Processors
		makespan := input.Makespan
		fmt.Println("makespan: ", makespan)

		var execwg sync.WaitGroup
		execwg.Add(len(processors))
		errMaps := make([]map[int]error, len(processors))
		for id, processor := range processors {
			errMaps[id] = make(map[int]error)
			go processor.Execute(e.BlkCtx, &execwg, errMaps[id])
		}
		execwg.Wait()
		for id, errMap := range errMaps {
			if len(errMap) != 0 {
				fmt.Println("Processor ", id, " has errors: ", len(errMap))
			}
		}
	}
}
