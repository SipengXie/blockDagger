package test

import (
	"blockDagger/core/vm"
	"blockDagger/helper"
	"blockDagger/pipeline"
	"blockDagger/schedule"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

// 就用这个测 测够50个区块就好
func TestParallel(t *testing.T) {
	ctx, _, graph, _, db, blkReader, blk, header := helper.PrepareforSingleBlock(18999999)

	scheduler := schedule.NewScheduler(graph, runtime.NumCPU())

	st := time.Now()
	processors, makespan := scheduler.Schedule()
	fmt.Println("makespan: ", makespan)
	fmt.Println("Schedule Cost: ", time.Since(st))
	var wg sync.WaitGroup
	wg.Add(len(processors))
	errMaps := make([]map[int]error, len(processors))
	st = time.Now()
	for id, processor := range processors {
		errMaps[id] = make(map[int]error)
		go processor.Execute(blkReader, ctx, blk, header, db, &wg, errMaps[id])
	}
	wg.Wait()
	fmt.Println("Parallel Execution Time: ", time.Since(st))
	systemAbortCnt := 0
	vmAbort := 0
	for _, errMap := range errMaps {
		for _, err := range errMap {
			if err == vm.ErrSystemAbort {
				systemAbortCnt++
			} else if err != nil {
				vmAbort++
			}

		}
	}
	fmt.Println("System Abort Count: ", systemAbortCnt)
	fmt.Println("VM Abort Count: ", vmAbort)
}

// 就用这个测 2、3、4、5、6、7、8、9、10区块聚合（5区块约1k）
func TestParallelMultipleBlocks(t *testing.T) {
	ctx, _, graph, _, db, blkReader, blk, header := helper.PrepareForKBlocks(18999990, 5)
	scheduler := schedule.NewScheduler(graph, runtime.NumCPU())
	st := time.Now()
	processors, makespan := scheduler.Schedule()
	fmt.Println("makespan: ", makespan)
	fmt.Println("Schedule Cost: ", time.Since(st))

	var wg sync.WaitGroup
	wg.Add(len(processors))
	errMaps := make([]map[int]error, len(processors))

	st = time.Now()
	for id, processor := range processors {
		errMaps[id] = make(map[int]error)
		go processor.Execute(blkReader, ctx, blk, header, db, &wg, errMaps[id])
	}
	wg.Wait()
	fmt.Println("Parallel Execution Time: ", time.Since(st))

	systemAbortCnt := 0
	vmAbort := 0
	for _, errMap := range errMaps {
		for _, err := range errMap {
			if err == vm.ErrSystemAbort {
				systemAbortCnt++
			} else if err != nil {
				vmAbort++
			}

		}
	}
	fmt.Println("System Abort Count: ", systemAbortCnt)
	fmt.Println("VM Abort Count: ", vmAbort)

}

// 就用这个测 2、3、4、5、6、7、8、9、10区块聚合（5区块约1k）
func TestPipelineSim(t *testing.T) {
	ctx, _, graphGroup, db, blkReader, blk, header := helper.PreparePipelineSim(18999940, 60, 12)
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
			go processor.Execute(blkReader, ctx, blk, header, db, &wg, errMaps[id])
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
	ctx, txwsGroup, db, gvc, blkReader, blk, header := helper.PreparePipeline(18999900, 100, 20)
	txwsMsgChan := make(chan *pipeline.TxwsMessage, len(txwsGroup)+2)
	taskMapsAndAccessedByChan := make(chan *pipeline.TaskMapsAndAccessedBy, len(txwsGroup)+2)
	graphMsgChan := make(chan *pipeline.GraphMessage, len(txwsGroup)+2)
	scheduleMsgChan := make(chan *pipeline.ScheduleMessage, len(txwsGroup)+2)

	//初始化四条流水线
	var wg sync.WaitGroup
	gvcLine := pipeline.NewGVCLine(gvc, &wg, txwsMsgChan, taskMapsAndAccessedByChan)
	graphLine := pipeline.NewGraphLine(&wg, taskMapsAndAccessedByChan, graphMsgChan)
	scheduleLine := pipeline.NewScheduleLine(runtime.NumCPU(), &wg, graphMsgChan, scheduleMsgChan)
	executeLine := pipeline.NewExecuteLine(blkReader, ctx, blk, header, db, &wg, scheduleMsgChan)

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
	st := time.Now()
	wg.Add(4)
	go executeLine.Run()
	go scheduleLine.Run()
	go graphLine.Run()
	go gvcLine.Run()
	wg.Wait()
	elapsed := time.Since(st)
	fmt.Println("Pipeline Execution Time: ", elapsed)
}

func TestSerial(t *testing.T) {
	serialTime := helper.SerialExecutionTime(18999999)
	fmt.Println("Serial Execution Time: ", serialTime)
}

func TestSerialMultipleBlocks(t *testing.T) {
	SerialExecutionKBlocksTime := helper.SerialExecutionKBlocks(18999800, 100)
	fmt.Println("Serial Execution Time: ", SerialExecutionKBlocksTime)
}
