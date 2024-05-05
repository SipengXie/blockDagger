package test

import (
	"blockDagger/graph"
	"blockDagger/schedule"
	"blockDagger/types"
	"fmt"
	"testing"
)

// 包括建图、测试两部分
// 这是真实任务图，没加entry和end
func generateTestGraph() *graph.Graph {
	graph := graph.NewGraph()
	taskList := make([]*types.Task, 10)

	taskList[0] = types.NewTask(0, 13, nil)
	taskList[1] = types.NewTask(1, 17, nil)
	taskList[2] = types.NewTask(2, 14, nil)
	taskList[3] = types.NewTask(3, 13, nil)
	taskList[4] = types.NewTask(4, 12, nil)
	taskList[5] = types.NewTask(5, 13, nil)
	taskList[6] = types.NewTask(6, 11, nil)
	taskList[7] = types.NewTask(7, 10, nil)
	taskList[8] = types.NewTask(8, 17, nil)
	taskList[9] = types.NewTask(9, 15, nil)

	for _, task := range taskList {
		graph.AddVertex(task)
	}

	graph.AddEdge(0, 1)
	graph.AddEdge(0, 2)
	graph.AddEdge(0, 3)
	graph.AddEdge(0, 4)
	graph.AddEdge(0, 5)

	graph.AddEdge(1, 7)
	graph.AddEdge(1, 8)

	graph.AddEdge(2, 6)

	graph.AddEdge(3, 7)
	graph.AddEdge(3, 8)

	graph.AddEdge(4, 8)

	graph.AddEdge(5, 7)

	graph.AddEdge(6, 9)

	graph.AddEdge(7, 9)

	graph.AddEdge(8, 9)

	return graph
}

func TestEFT(t *testing.T) {
	graph := generateTestGraph()
	// 增加entry与end
	taskEntry := types.NewTask(-1, 0, nil)
	taskEnd := types.NewTask(10, 0, nil)

	graph.AddVertex(taskEntry)
	graph.AddVertex(taskEnd)

	for id, v := range graph.Vertices {
		if id == -1 || id == 10 {
			continue
		}
		if v.InDegree == 0 {
			graph.AddEdge(-1, id)
		} else if v.OutDegree == 0 {
			graph.AddEdge(id, 10)
		}
	}

	graph.GenerateProperties()
	listScheduler := schedule.NewScheduler(graph, 4)
	processors, makespan := listScheduler.ListSchedule(schedule.EFT)
	for id, p := range processors {
		fmt.Printf("Processor %d: ", id)
		for _, task := range p.Tasks {
			fmt.Printf("%d ", task.Task.ID)
		}
		fmt.Println()
	}
	fmt.Println("makespan: ", makespan)

	processors, makespan = listScheduler.ListSchedule(schedule.CPTL)
	for id, p := range processors {
		fmt.Printf("Processor %d: ", id)
		for _, task := range p.Tasks {
			fmt.Printf("%d ", task.Task.ID)
		}
		fmt.Println()
	}
	fmt.Println("makespan: ", makespan)

	processors, makespan = listScheduler.ListSchedule(schedule.CT)
	for id, p := range processors {
		fmt.Printf("Processor %d: ", id)
		for _, task := range p.Tasks {
			fmt.Printf("%d ", task.Task.ID)
		}
		fmt.Println()
	}
	fmt.Println("makespan: ", makespan)

	processors, makespan = listScheduler.CPOPSchedule()
	for id, p := range processors {
		fmt.Printf("Processor %d: ", id)
		for _, task := range p.Tasks {
			fmt.Printf("%d ", task.Task.ID)
		}
		fmt.Println()
	}
	fmt.Println("makespan: ", makespan)

	// processors, makespan = listScheduler.TopoSchedule()
	// for id, p := range processors {
	// 	fmt.Printf("Processor %d: ", id)
	// 	for _, task := range p.Tasks {
	// 		fmt.Printf("%d ", task.ID)
	// 	}
	// 	fmt.Println()
	// }
	// fmt.Println("makespan: ", makespan)
}
