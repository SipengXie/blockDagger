package schedule

import (
	"blockDagger/core"
	"blockDagger/core/vm"
	"blockDagger/core/vm/evmtypes"
	"blockDagger/state"
	"container/heap"

	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/params"
)

type Processor struct {
	Tasks           ESTTaskQueue
	SlotsMaintainer *SlotsMaintainer
}

func NewProcessor(timespan uint64) *Processor {
	queue := make(ESTTaskQueue, 0)
	heap.Init(&queue)
	return &Processor{
		Tasks:           queue,
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
	heap.Push(&p.Tasks, tWrap)
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
	p.SlotsMaintainer.addSlot(tWrap.EFT, slotLength-tWrap.Task.Cost)
}

func (p *Processor) Execute(blkCtx evmtypes.BlockContext) map[int]error {

	// 执行环境准备
	s := state.NewState()
	evm := vm.NewEVM(blkCtx, evmtypes.TxContext{}, s, params.MainnetChainConfig, vm.Config{})
	errs := make(map[int]error)

	// TODO: 可用一个map返回ret
	for p.Tasks.Len() > 0 {
		tWrap := heap.Pop(&p.Tasks).(*TaskWrapper)

		task := tWrap.Task
		task.Wait()

		msg, err := task.Tx.AsMessage(*types.LatestSigner(params.MainnetChainConfig), blkCtx.BaseFee.ToBig(), evm.ChainRules())
		if err != nil {
			errs[task.ID] = err
			continue
		}

		res, err := core.ApplyMessage(evm, msg, new(core.GasPool).AddGas(blkCtx.GasLimit), true /* refunds */, false /* gasBailout */)
		if err != nil {
			s.TotallyAbort()
			errs[task.ID] = err
		} else if res.Err != nil {
			errs[task.ID] = res.Err
		}
		s.CommitLocalWrite()
	}
	return errs
}
