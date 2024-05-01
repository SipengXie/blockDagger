package greedy

import (
	"blockDagger/types"
	"container/heap"
)

// ID排序
type SortedTasks []*types.Task

func (s SortedTasks) Len() int {
	return len(s)
}

func (s SortedTasks) Less(i, j int) bool {
	return s[i].ID < s[j].ID
}

func (s SortedTasks) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type sumData struct {
	sum uint64
	id  int
}

// 做一个小根堆，用来做贪心的fallback
type sumArray []sumData

func (s sumArray) Len() int {
	return len(s)
}

func (s sumArray) Less(i, j int) bool {
	return s[i].sum < s[j].sum
}

func (s sumArray) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s *sumArray) Push(x interface{}) {
	*s = append(*s, x.(sumData))
}

func (s *sumArray) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[:n-1]
	return x
}

func Greedy(tasks []*types.Task, k int) ([]SortedTasks, uint64) {
	gbst := &GBST{} // 记录余下的task
	var sum uint64
	for _, t := range tasks {
		gbst.Add(t)
		sum += t.Cost
	}
	average := sum / uint64(k)

	// k 组结果
	result := make([]SortedTasks, k)
	gasSums := make(sumArray, k)
	var maxSum uint64

	for i := 0; i < k; i++ {
		// 先把最大的放进去
		largest, _ := gbst.Largest()
		if largest == nil {
			// 这意味着所有的task都已经分配完了，其实就是|task| < k的情况
			break
		}
		cur_group := SortedTasks{largest}
		cur_sum := largest.Cost

		gbst.Remove(largest)

		for {
			// 寻找一个最大的task，使得加入后cur_sum不超过average
			find := gbst.Search(average - cur_sum)
			if find == nil {
				break
			}

			cur_sum += find.Cost
			cur_group = append(cur_group, find)
			gbst.Remove(find)
		}
		result[i] = cur_group
		gasSums[i] = sumData{cur_sum, i}
		maxSum = max(maxSum, cur_sum)
	}

	// 可能余下了一些没有分配完的Txs
	heap.Init(&gasSums)
	nodes := gbst.Flatten()
	// nodes是按Gas从小到大排序的，我们要反过来遍历
	for i := len(nodes) - 1; i >= 0; i-- {
		t := nodes[i].task
		// 找一个最小的GasSum组丢进去,还需要获得gasSums对应的groupId
		targetGroup := heap.Pop(&gasSums).(sumData)
		result[targetGroup.id] = append(result[targetGroup.id], t)
		targetGroup.sum += t.Cost
		maxSum = max(maxSum, targetGroup.sum)
		heap.Push(&gasSums, targetGroup)
	}
	// fmt.Println("Average:", average)
	// fmt.Println(gasSums)
	// for i, group := range result {
	// 	sort.Sort(group)
	// 	fmt.Println("Group", i, ":", group)
	// }

	return result, maxSum
}
