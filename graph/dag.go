package graph

import (
	"blockDagger/types"
)

// Vertex 表示图中的顶点
type Vertex struct {
	Task   *types.Task
	Degree uint // IN-DEGREE
}

// UndirectedGraph 表示无向图
type Graph struct {
	Vertices     map[int]*Vertex          `json:"vertices"`     // 顶点集合
	AdjacencyMap map[int]map[int]struct{} `json:"adjacencyMap"` // 邻接边表
	ReverseMap   map[int]map[int]struct{} `json:"reverseMap"`   // 逆邻接边表
}

func NewGraph() *Graph {
	return &Graph{
		Vertices:     make(map[int]*Vertex),
		AdjacencyMap: make(map[int]map[int]struct{}),
		ReverseMap:   make(map[int]map[int]struct{}),
	}
}

func (g *Graph) AddVertex(task *types.Task) {
	id := task.ID
	_, exist := g.Vertices[id]
	if exist {
		return
	}
	v := &Vertex{
		Task:   task,
		Degree: 0,
	}
	g.Vertices[id] = v
	g.AdjacencyMap[id] = make(map[int]struct{})
	g.ReverseMap[id] = make(map[int]struct{})
}

func (g *Graph) AddEdge(source, destination int) {
	if g.HasEdge(source, destination) {
		return
	}
	g.AdjacencyMap[source][destination] = struct{}{}
	g.ReverseMap[destination][source] = struct{}{}
	g.Vertices[destination].Degree++
}

func (g *Graph) HasEdge(source, destination int) bool {
	_, ok := g.Vertices[source]
	if !ok {
		return false
	}

	_, ok = g.Vertices[destination]
	if !ok {
		return false
	}

	_, ok = g.AdjacencyMap[source][destination]
	return ok
}

func (g *Graph) GetTopo() [][]int {
	ans := make([][]int, 0)
	degreeZero := make([]int, 0)
	for id, v := range g.Vertices {
		if v.Degree == 0 {
			degreeZero = append(degreeZero, id)
		}
	}
	ans = append(ans, degreeZero)
	for {
		newDegreeZero := make([]int, 0)
		for _, vid := range degreeZero {
			for neighborid := range g.AdjacencyMap[vid] {
				g.Vertices[neighborid].Degree--
				if g.Vertices[neighborid].Degree == 0 {
					newDegreeZero = append(newDegreeZero, neighborid)
				}
			}
		}
		degreeZero = newDegreeZero
		if len(degreeZero) == 0 {
			break
		} else {
			ans = append(ans, degreeZero)
		}
	}
	return ans
}

// 由于是DAG，我们可以不判断visited
func (g *Graph) Dfs(vid int) {
	// fmt.Println(vid)
	for neighborid := range g.AdjacencyMap[vid] {
		g.Dfs(neighborid)
	}
}

// 由于是DAG，我们可以不判断visited
func (g *Graph) RevDfs(vid int) {
	for neighborid := range g.ReverseMap[vid] {
		g.RevDfs(neighborid)
	}
}
