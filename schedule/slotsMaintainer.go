package schedule

import "fmt"

type SlotsMaintainer struct {
	Slots       *AVLBST
	TimeSpan    *SegTree
	largestTime uint64
}

// TODO:这里可以把Timespan改为MAXUINT64
// TODO:或者timespan<<1也可以
func NewSlotsMaintainer(timespan uint64) *SlotsMaintainer {
	bst := NewTree()
	bst.Add(0, timespan<<1)
	segt := NewSegTree(0, timespan<<1)
	segt.Modify(0, timespan<<1)
	return &SlotsMaintainer{
		Slots:       bst,
		TimeSpan:    segt,
		largestTime: timespan << 1,
	}
}

// 返回找到最小EFT的St与Length
func (sm *SlotsMaintainer) findSlot(EST, length uint64) (slotSt uint64, slotLength uint64) {
	// 首先在Slots里面找，看看有没有包含EST的Slot
	// 如果有，直接返回该slot，EFT=EST+length
	slot := sm.Slots.FindMaxLessThan(EST)
	if slot != nil && slot.st+slot.length >= EST+length {
		return slot.st, slot.length
	}

	// 在[EST, MAXENDING]中找到第一个大于等于length的Slot的st
	slotSt, slotLength = sm.TimeSpan.Query(EST, sm.largestTime, length)
	if slotSt == MAXUINT64 {
		fmt.Println("Error: No slot found")
	}
	return
}

// 增加一个slot
func (sm *SlotsMaintainer) addSlot(st, length uint64) {
	if length != 0 {
		sm.Slots.Add(st, length)
	}
	sm.TimeSpan.Modify(st, length)
}

// 修改原来的slot
func (sm *SlotsMaintainer) modifySlot(st, length uint64) {
	sm.Slots.Remove(st)
	sm.addSlot(st, length)
}
