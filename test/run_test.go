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

	// processors, makespan := scheduler.ListSchedule(schedule.EFT)
	// fmt.Println("makespan: ", makespan)

	// processors, makespan := scheduler.ListSchedule(schedule.CPTL)
	// fmt.Println("makespan: ", makespan)

	// processors, makespan := scheduler.ListSchedule(schedule.CT)
	// fmt.Println("makespan: ", makespan)

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

func TestSerial(t *testing.T) {
	serialTime := helper.SerialExecutionTime(18999999)
	fmt.Println("Serial Execution Time: ", serialTime)
}
