package test

import (
	"blockDagger/helper"
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

func TestPipeline(t *testing.T) {
	_, graphGroup, blkCtx := helper.PreparePipeline(18999989, 11, 2)
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

func TestSerial(t *testing.T) {
	serialTime := helper.SerialExecutionTime(18999999)
	fmt.Println("Serial Execution Time: ", serialTime)
}

func TestSerialMultipleBlocks(t *testing.T) {
	SerialExecutionKBlocksTime := helper.SerialExecutionKBlocks(18999989, 11)
	fmt.Println("Serial Execution Time: ", SerialExecutionKBlocksTime)
}
