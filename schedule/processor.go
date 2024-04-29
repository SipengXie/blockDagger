package schedule

import "blockDagger/types"

type Processor struct {
	Tasks           []*types.Task
	SlotsMaintainer *SlotsMaintainer
}

func NewProcessor(timespan uint64) *Processor {
	return &Processor{
		Tasks:           make([]*types.Task, 0),
		SlotsMaintainer: NewSlotsMaintainer(timespan),
	}
}

// 返回找到的St与Length以及对应的Task EFT
func (p *Processor) FindEFT(EST, length uint64) (slotSt uint64, slotLength uint64, EFT uint64) {
	slotSt, slotLength = p.SlotsMaintainer.findSlot(EST, length)
	if slotSt <= EST {
		EFT = EST + length
	} else {
		EFT = slotSt + length
	}
	return
}

func (p *Processor) AddTask(tWrap *TaskWrapper, slotSt uint64, slotLength uint64) {
	p.Tasks = append(p.Tasks, tWrap.Task)
	// slot包含EST
	if slotSt <= tWrap.EST {
		// 需要先更改再添加
		// 原先的slot 长度变为EST-slotSt
		p.SlotsMaintainer.modifySlot(slotSt, tWrap.EST-slotSt)
		// 新增的slot 从EFT开始 长度为Slot.ed-EFT=slotSt+slotLength-EFT的slot
		p.SlotsMaintainer.addSlot(tWrap.EFT, slotSt+slotLength-tWrap.EFT)
		return
	}

	// slot不包含EST
	// 原先的slot长度变为0
	// 新增的slot从EFT开始 长度为slotLength - length = slotSt + slotLength - EFT = slotLength - Task.Cost
	p.SlotsMaintainer.modifySlot(slotSt, 0)
	p.SlotsMaintainer.addSlot(tWrap.EFT, slotSt+slotLength-tWrap.EFT)
}
