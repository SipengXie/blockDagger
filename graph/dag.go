package graph

import (
	"blockDagger/types"
)

// Vertex 表示图中的顶点
type Vertex struct {
	Task      *types.Task
	InDegree  uint // IN-DEGREE
	OutDegree uint // OUT-DEGREE

	// properties needed to schedule
	Rank_u uint64 // 算上自己的最大后续开销
	Rank_d uint64 // 不计算自己的最大前序开销
	CT     uint64 // 不算自己的最大后续开销
}

// UndirectedGraph 表示无向图
type Graph struct {
	Vertices     map[int]*Vertex          `json:"vertices"`     // 顶点集合
	AdjacencyMap map[int]map[int]struct{} `json:"adjacencyMap"` // 邻接边表
	ReverseMap   map[int]map[int]struct{} `json:"reverseMap"`   // 逆邻接边表

	CriticalPathLen uint64
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
		Task: task,
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
	g.Vertices[source].OutDegree++
	g.Vertices[destination].InDegree++
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

// 由于是DAG，我们可以不判断visited
func (g *Graph) Dfs(vid int) {
	for succid := range g.AdjacencyMap[vid] {
		succ := g.Vertices[succid]
		succ.Rank_d = max(succ.Rank_d, g.Vertices[vid].Rank_d+g.Vertices[vid].Task.Cost) // rank_d由pred更新，所以vid会更新succ的rank_d
		g.Dfs(succid)
	}
}

// 由于是DAG，我们可以不判断visited
func (g *Graph) RevDfs(vid int) {
	cur := g.Vertices[vid]
	for predid := range g.ReverseMap[vid] {
		pred := g.Vertices[predid]
		pred.Rank_u = max(pred.Rank_u, pred.Task.Cost+cur.Rank_u) //  rank_u由succ更新，所以vid会更新pred的rank_u
		pred.CT = max(pred.CT, cur.CT+cur.Task.Cost)
		g.RevDfs(predid)
	}
}

func (g *Graph) GenerateProperties() {
	g.CriticalPathLen = 0
	g.Dfs(-1)

	// 点数 = 任务数 + 2
	// -1 0 1 2 3
	// Vend 为 3 【即任务数】
	g.RevDfs(len(g.Vertices) - 2)

	for _, v := range g.Vertices {
		g.CriticalPathLen = max(g.CriticalPathLen, v.Rank_u+v.Rank_d)
	}
}
