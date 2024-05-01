package schedule

import "blockDagger/types"

// 因为后面要并发处理，所以需要将Task包装一下
type TaskWrapper struct {
	Task     *types.Task
	Priority uint64
	EST      uint64
	EFT      uint64
}

type PriorityTaskQueue []*TaskWrapper

func (pq PriorityTaskQueue) Len() int { return len(pq) }

func (pq PriorityTaskQueue) Less(i, j int) bool {
	if pq[i].Priority == pq[j].Priority {
		return pq[i].Task.ID < pq[j].Task.ID
	}
	return pq[i].Priority > pq[j].Priority
}

func (pq PriorityTaskQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityTaskQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*TaskWrapper))
}

func (pq *PriorityTaskQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	x := old[n-1]
	*pq = old[0 : n-1]
	return x
}
