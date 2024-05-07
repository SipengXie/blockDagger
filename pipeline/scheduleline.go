package pipeline

import (
	"blockDagger/schedule"
	"sync"
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
	for input := range s.InputChan {
		if input.Flag == END {
			s.Wg.Done()
			outMessage := &ScheduleMessage{
				Flag: END,
			}
			s.OutputChan <- outMessage
			close(s.OutputChan)
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
