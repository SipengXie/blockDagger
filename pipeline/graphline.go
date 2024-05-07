package pipeline

import (
	"blockDagger/helper"
	"sync"
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
	for input := range g.InputChan {
		if input.Flag == END {
			outMessage := &GraphMessage{
				Flag: END,
			}
			g.OutputChan <- outMessage
			close(g.OutputChan)
			g.Wg.Done()
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
