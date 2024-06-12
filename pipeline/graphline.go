package pipeline

import (
	"blockDagger/helper"
	"fmt"
	"sync"
	"time"
)

type GraphLine struct {
	Wg         *sync.WaitGroup
	InputChan  chan *TaskMapsAndAccessedBy
	OutputChan chan *GraphMessage
}

func NewGraphLine(wg *sync.WaitGroup, in chan *TaskMapsAndAccessedBy, out chan *GraphMessage) *GraphLine {
	return &GraphLine{
		Wg:         wg,
		InputChan:  in,
		OutputChan: out,
	}
}

func (g *GraphLine) Run() {
	var elapsed int64
	for input := range g.InputChan {
		if input.Flag == END {
			outMessage := &GraphMessage{
				Flag: END,
			}
			g.OutputChan <- outMessage
			close(g.OutputChan)
			g.Wg.Done()
			fmt.Println("Graph Generation Cost:", elapsed, "ms")
			return
		}

		st := time.Now()
		graph := helper.GenerateGraph(input.TaskMap, input.RwAccessedBy)
		elapsed += time.Since(st).Milliseconds()

		outMessage := &GraphMessage{
			Flag:  START,
			Graph: graph,
		}
		g.OutputChan <- outMessage
	}
}
