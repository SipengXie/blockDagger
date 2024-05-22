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
	st := time.Now()
	for input := range g.InputChan {
		// fmt.Println("graphline")
		if input.Flag == END {
			outMessage := &GraphMessage{
				Flag: END,
			}
			g.OutputChan <- outMessage
			close(g.OutputChan)
			g.Wg.Done()
			fmt.Println("Graph Cost:", time.Since(st))
			return
		}

		graph := helper.GenerateGraph(input.TaskMap, input.RwAccessedBy)
		outMessage := &GraphMessage{
			Flag:  START,
			Graph: graph,
		}
		g.OutputChan <- outMessage
	}
}
