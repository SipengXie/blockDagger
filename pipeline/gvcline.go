package pipeline

import (
	"blockDagger/helper"
	multiversion "blockDagger/multiVersion"
	"blockDagger/types"
	"fmt"
	"sync"
	"time"
)

type GVCLine struct {
	Gvc        *multiversion.GlobalVersionChain
	Wg         *sync.WaitGroup
	InputChan  chan *TxwsMessage
	OutputChan chan *TaskMapsAndAccessedBy
}

func NewGVCLine(gvc *multiversion.GlobalVersionChain, wg *sync.WaitGroup, in chan *TxwsMessage, out chan *TaskMapsAndAccessedBy) *GVCLine {
	return &GVCLine{
		Gvc:        gvc,
		Wg:         wg,
		InputChan:  in,
		OutputChan: out,
	}
}

func (g *GVCLine) Run() {
	var elapsed int64
	for input := range g.InputChan {
		// 如果是END信号，那么就结束
		// fmt.Println("gvcline")
		if input.Flag == END {
			outMessage := &TaskMapsAndAccessedBy{
				Flag: END,
			}
			g.OutputChan <- outMessage
			close(g.OutputChan) // 通知下一个Line结束循环
			g.Wg.Done()
			fmt.Println("Pre-Processing Cost:", elapsed, "ms")
			return
		}

		//否则队Gvc进行更新并把建图所需要的信息传递给GraphLine
		st := time.Now()
		txws := input.Txws
		rwAccessedBy := helper.GenerateAccessedBy(txws)
		taskMap := make(map[int]*types.Task)
		for _, txw := range txws {
			task := helper.TransferTxToTask(*txw, g.Gvc)
			taskMap[task.ID] = task
		}
		g.Gvc.UpdateLastBlockTail()
		since := time.Since(st)
		elapsed += since.Milliseconds()

		outMessage := &TaskMapsAndAccessedBy{
			Flag:         START,
			TaskMap:      taskMap,
			RwAccessedBy: rwAccessedBy,
		}
		g.OutputChan <- outMessage
	}
}
