package graph

import (
	"blockDagger/types"
)

const MAXINT = int(^uint(0) >> 1)

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

func (g *Graph) getTopo(rev bool) []int {
	mapDegree := make(map[int]uint)
	degreeZero := make([]int, 0)
	for id, v := range g.Vertices {
		if rev {
			mapDegree[id] = v.OutDegree
		} else {
			mapDegree[id] = v.InDegree
		}
		if mapDegree[id] == 0 {
			degreeZero = append(degreeZero, id)
		}
	}

	topo := make([]int, 0)
	for {
		newDegreeZero := make([]int, 0)
		for _, vid := range degreeZero {
			topo = append(topo, vid)
			edges := g.AdjacencyMap[vid]
			if rev {
				edges = g.ReverseMap[vid]
			}

			for succId := range edges {
				mapDegree[succId]--
				if mapDegree[succId] == 0 {
					newDegreeZero = append(newDegreeZero, succId)
				}
			}
		}
		degreeZero = newDegreeZero
		if len(degreeZero) == 0 {
			break
		}
	}
	return topo
}

// 在获得topo序的基础上，计算rank_d
func (g *Graph) calcRankD() {
	topo := g.getTopo(false)
	stid := topo[0]
	g.Vertices[stid].Rank_d = 0
	for i := 1; i < len(topo); i++ {
		vid := topo[i]
		curv := g.Vertices[vid]
		// getmaxPredcessor
		maxPred := uint64(0)
		for predid := range g.ReverseMap[vid] {
			pred := g.Vertices[predid]
			maxPred = max(maxPred, pred.Rank_d+pred.Task.Cost)
		}
		curv.Rank_d = maxPred
	}
}

// 在获得逆topo序的基础上，计算rank_u, CT
func (g *Graph) calcRankUCT() {
	topo := g.getTopo(true)
	edid := topo[0]
	g.Vertices[edid].CT = g.Vertices[edid].Task.Cost
	g.Vertices[edid].Rank_u = g.Vertices[edid].Task.Cost
	for i := 0; i < len(topo); i++ {
		vid := topo[i]
		cur := g.Vertices[vid]
		// getmaxSuccessor
		maxRanku := uint64(0)
		maxct := uint64(0)

		for succid := range g.AdjacencyMap[vid] {
			succ := g.Vertices[succid]
			maxRanku = max(maxRanku, succ.Rank_u)
			maxct = max(maxct, succ.CT+succ.Task.Cost)
		}
		cur.Rank_u = maxRanku + cur.Task.Cost
		cur.CT = maxct
	}
}

func (g *Graph) GenerateProperties() {
	g.CriticalPathLen = 0
	// 应该先得到topo序，然后再……
	g.calcRankD()
	g.calcRankUCT()

	for _, v := range g.Vertices {
		g.CriticalPathLen = max(g.CriticalPathLen, v.Rank_u+v.Rank_d)
	}
}
