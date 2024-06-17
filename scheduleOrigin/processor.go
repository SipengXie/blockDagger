package scheduleOrigin

import "fmt"

// 这里我们将实现原始insertion-based policy中的processor
// 主要区别在于processor里包含一个task链表
// 然后FindEFT(task)是通过遍历链表来实现寻找可以容下task的地方

type TaskWrapperNode struct {
	TaskWrapper
	Next *TaskWrapperNode
}

type Processor struct {
	head *TaskWrapperNode
}

func NewProcessor() *Processor {
	var Head = &TaskWrapperNode{
		TaskWrapper: TaskWrapper{
			Task:     nil,
			Priority: 0,
			EST:      0,
			EFT:      0,
		},
		Next: nil,
	}
	return &Processor{
		head: Head,
	}
}

// 找到一个合适的位置插入task
func (p *Processor) FindEFT(tWrap *TaskWrapper) (prev *TaskWrapperNode, EFT uint64) {
	cur := p.head
	prev = nil
	for cur.Next != nil {
		if cur.Next.EST >= tWrap.Task.Cost+max(cur.EFT, tWrap.EST) {
			prev = cur
			break
		}
		cur = cur.Next
	}

	// 走到链尾了
	if cur.Next == nil {
		prev = cur
	}

	return prev, max(prev.EFT, tWrap.EST) + tWrap.Task.Cost
}

func (p *Processor) AddTask(tWrap *TaskWrapper, prev *TaskWrapperNode) {
	if prev == nil {
		panic("prev should not be nil")
	}
	// 一定 prev != nil
	originNext := prev.Next
	prev.Next = &TaskWrapperNode{
		TaskWrapper: *tWrap,
		Next:        originNext,
	}

}

func (p *Processor) Print() {
	cur := p.head.Next
	for cur != nil {
		fmt.Printf("%d ", cur.TaskWrapper.Task.ID)
		cur = cur.Next
	}
}
