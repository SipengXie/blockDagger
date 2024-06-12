package pipeline

import (
	"blockDagger/schedule"
	"fmt"
	"sync"
	"time"
)

type ScheduleLine struct {
	NumWorker  int
	Wg         *sync.WaitGroup
	InputChan  chan *GraphMessage
	OutputChan chan *ScheduleMessage
}

func NewScheduleLine(numWorker int, wg *sync.WaitGroup, in chan *GraphMessage, out chan *ScheduleMessage) *ScheduleLine {
	return &ScheduleLine{
		NumWorker:  numWorker,
		Wg:         wg,
		InputChan:  in,
		OutputChan: out,
	}
}

func (s *ScheduleLine) Run() {
	var elapsed int64
	for input := range s.InputChan {
		// fmt.Println("scheduleline")
		if input.Flag == END {
			outMessage := &ScheduleMessage{
				Flag: END,
			}
			s.OutputChan <- outMessage
			close(s.OutputChan)
			s.Wg.Done()
			fmt.Println("Parallel Schedule Cost:", elapsed, "ms")
			return
		}

		scheduler := schedule.NewScheduler(input.Graph, s.NumWorker)
		st := time.Now()
		processors, makespan := scheduler.Schedule()
		elapsed += time.Since(st).Milliseconds()
		outMessage := &ScheduleMessage{
			Flag:       START,
			Processors: processors,
			Makespan:   makespan,
		}
		s.OutputChan <- outMessage
	}
}
