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
	st := time.Now()
	for input := range s.InputChan {
		// fmt.Println("scheduleline")
		if input.Flag == END {
			outMessage := &ScheduleMessage{
				Flag: END,
			}
			s.OutputChan <- outMessage
			close(s.OutputChan)
			s.Wg.Done()
			fmt.Println("Schedule Cost:", time.Since(st))
			return
		}

		scheduler := schedule.NewScheduler(input.Graph, s.NumWorker)
		processors, makespan := scheduler.Schedule()
		outMessage := &ScheduleMessage{
			Flag:       START,
			Processors: processors,
			Makespan:   makespan,
		}
		s.OutputChan <- outMessage
	}
}
