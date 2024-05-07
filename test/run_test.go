package test

import (
	"blockDagger/helper"
	"blockDagger/pipeline"
	"blockDagger/schedule"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestParallel(t *testing.T) {
	_, graph, _, blkCtx := helper.PrepareforSingleBlock(18999999)
	scheduler := schedule.NewScheduler(graph, runtime.NumCPU())

	processors, makespan := scheduler.Schedule()
	fmt.Println("makespan: ", makespan)
	var wg sync.WaitGroup
	wg.Add(len(processors))
	errMaps := make([]map[int]error, len(processors))
	st := time.Now()
	for id, processor := range processors {
		errMaps[id] = make(map[int]error)
		go processor.Execute(blkCtx, &wg, errMaps[id])
	}
	wg.Wait()
	elapsed := time.Since(st)
	for id, errMap := range errMaps {
		if len(errMap) != 0 {
			fmt.Println("Processor ", id, " has errors: ", errMap)
		}
	}
	fmt.Println("Parallel Execution Time: ", elapsed)
}

func TestParallelMultipleBlocks(t *testing.T) {
	_, graph, _, blkCtx := helper.PrepareForKBlocks(18999989, 11)
	scheduler := schedule.NewScheduler(graph, runtime.NumCPU())
	processors, makespan := scheduler.Schedule()
	fmt.Println("makespan: ", makespan)
	var wg sync.WaitGroup
	wg.Add(len(processors))
	errMaps := make([]map[int]error, len(processors))
	st := time.Now()
	for id, processor := range processors {
		errMaps[id] = make(map[int]error)
		go processor.Execute(blkCtx, &wg, errMaps[id])
	}
	wg.Wait()
	elapsed := time.Since(st)

	for id, errMap := range errMaps {
		if len(errMap) != 0 {
			fmt.Println("Processor ", id, " has errors: ", len(errMap))
		}
	}
	fmt.Println("Parallel Execution Time: ", elapsed)
}

func TestPipelineSim(t *testing.T) {
	_, graphGroup, blkCtx := helper.PreparePipelineSim(18999989, 11, 2)
	// 这里是Schduler流水线的例子
	schedulers := make([]*schedule.Scheduler, len(graphGroup))
	processorsGroup, makespanGroup := make([][]*schedule.Processor, len(graphGroup)), make([]uint64, len(graphGroup))
	for i := range graphGroup {
		schedulers[i] = schedule.NewScheduler(graphGroup[i], runtime.NumCPU())
		processorsGroup[i], makespanGroup[i] = schedulers[i].Schedule()
	}

	// 这里是执行流水线的例子
	for i := range processorsGroup {
		processors := processorsGroup[i]
		makespan := makespanGroup[i]
		fmt.Println("megaBlock ", i, " makespan: ", makespan)
		var wg sync.WaitGroup
		wg.Add(len(processors))
		errMaps := make([]map[int]error, len(processors))
		for id, processor := range processors {
			errMaps[id] = make(map[int]error)
			go processor.Execute(blkCtx, &wg, errMaps[id])
		}
		wg.Wait()
		for id, errMap := range errMaps {
			if len(errMap) != 0 {
				fmt.Println("megaBlock ", i, "Processor ", id, " has errors: ", len(errMap))
			}
		}
	}
}

func TestPipeline(t *testing.T) {
	//初始化执行环境与Channel
	txwsGroup, blkCtx, gvc := helper.PreparePipeline(18999989, 11, 2)
	txwsMsgChan := make(chan *pipeline.TxwsMessage, len(txwsGroup)+2)
	taskMapsAndAccessedByChan := make(chan *pipeline.TaskMapsAndAccessedBy, len(txwsGroup)+2)
	graphMsgChan := make(chan *pipeline.GraphMessage, len(txwsGroup)+2)
	scheduleMsgChan := make(chan *pipeline.ScheduleMessage, len(txwsGroup)+2)

	//初始化四条流水线
	var wg sync.WaitGroup
	gvcLine := pipeline.NewGVCLine(gvc, &wg, txwsMsgChan, taskMapsAndAccessedByChan)
	graphLine := pipeline.NewGraphLine(&wg, taskMapsAndAccessedByChan, graphMsgChan)
	scheduleLine := pipeline.NewScheduleLine(runtime.NumCPU(), &wg, graphMsgChan, scheduleMsgChan)
	executeLine := pipeline.NewExecuteLine(blkCtx, &wg, scheduleMsgChan)

	//向第一条流水线填充交易
	for _, txws := range txwsGroup {
		txwsMsgChan <- &pipeline.TxwsMessage{
			Flag: pipeline.START,
			Txws: txws,
		}
	}
	txwsMsgChan <- &pipeline.TxwsMessage{
		Flag: pipeline.END,
	}
	close(txwsMsgChan)

	//启动四条流水线
	wg.Add(4)
	go gvcLine.Run()
	go graphLine.Run()
	go scheduleLine.Run()
	go executeLine.Run()
	wg.Wait()
}

func TestSerial(t *testing.T) {
	serialTime := helper.SerialExecutionTime(18999999)
	fmt.Println("Serial Execution Time: ", serialTime)
}

func TestSerialMultipleBlocks(t *testing.T) {
	SerialExecutionKBlocksTime := helper.SerialExecutionKBlocks(18999989, 11)
	fmt.Println("Serial Execution Time: ", SerialExecutionKBlocksTime)
}
