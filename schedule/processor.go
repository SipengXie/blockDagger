package schedule

import (
	"blockDagger/core"
	"blockDagger/core/vm"
	"blockDagger/core/vm/evmtypes"
	"blockDagger/helper"
	"blockDagger/state"
	"container/heap"
	"context"
	"sync"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/params"
	"github.com/ledgerwatch/erigon/turbo/snapshotsync/freezeblocks"
)

type Processor struct {
	Tasks           ASTTaskQueue
	SlotsMaintainer *SlotsMaintainer
}

func NewProcessor(timespan uint64) *Processor {
	queue := make(ASTTaskQueue, 0)
	heap.Init(&queue)
	return &Processor{
		Tasks:           queue,
		SlotsMaintainer: NewSlotsMaintainer(timespan),
	}
}

// 返回找到的St与Length以及对应的Task EFT
func (p *Processor) FindEFT(EST, length uint64) (slotSt uint64, slotLength uint64, EFT uint64) {
	slotSt, slotLength = p.SlotsMaintainer.findSlot(EST, length)
	if slotSt == MAXUINT64 {
		EFT = MAXUINT64
	} else if slotSt <= EST {
		EFT = EST + length
	} else {
		EFT = slotSt + length
	}
	return
}

func (p *Processor) AddTask(tWrap *TaskWrapper, slotSt uint64, slotLength uint64) {
	heap.Push(&p.Tasks, tWrap)
	// slot包含AST
	if slotSt <= tWrap.AST {
		// 需要先更改再添加
		// 原先的slot 长度变为AST-slotSt
		p.SlotsMaintainer.modifySlot(slotSt, tWrap.AST-slotSt)
		// 新增的slot 从EFT开始 长度为Slot.ed-EFT=slotSt+slotLength-EFT的slot
		p.SlotsMaintainer.addSlot(tWrap.EFT, slotSt+slotLength-tWrap.EFT)
		return
	}

	// slot不包含AST
	// 原先的slot长度变为0
	// 新增的slot从EFT开始 长度为slotLength - length = slotSt + slotLength - EFT = slotLength - Task.Cost
	p.SlotsMaintainer.modifySlot(slotSt, 0)
	p.SlotsMaintainer.addSlot(tWrap.EFT, slotLength-tWrap.Task.Cost)
}

func (p *Processor) Execute(blockReader *freezeblocks.BlockReader, ctx context.Context, blk *types.Block, header *types.Header, db kv.RoDB, wg *sync.WaitGroup, errs map[int]error) {
	defer wg.Done()
	// 执行环境准备
	s := state.NewState()

	dbTx, err := db.BeginRo(ctx)
	if err != nil {
		panic(err)
	}
	blkCtx := helper.GetBlockContext(blockReader, blk, dbTx, header)

	evm := vm.NewEVM(blkCtx, evmtypes.TxContext{}, s, params.MainnetChainConfig, vm.Config{})
	// TODO: 可用一个map返回ret
	for p.Tasks.Len() > 0 {
		tWrap := heap.Pop(&p.Tasks).(*TaskWrapper)

		task := tWrap.Task
		if task.Tx == nil {
			continue
		}
		task.Wait()
		s.SetTaskContext(task)
		msg, err := task.Tx.AsMessage(*types.LatestSigner(params.MainnetChainConfig), blkCtx.BaseFee.ToBig(), evm.ChainRules())
		if err != nil {
			errs[task.ID] = err
			continue
		}
		msg.SetCheckNonce(false)
		txCtx := core.NewEVMTxContext(msg)
		evm.TxContext = txCtx
		res, err := core.ApplyMessage(evm, msg, new(core.GasPool).AddGas(blkCtx.GasLimit), true /* refunds */, false /* gasBailout */)
		if err != nil {
			s.TotallyAbort()
			errs[task.ID] = err
			continue
		} else if res.Err != nil {
			errs[task.ID] = res.Err
		}
		s.CommitLocalWrite()
	}
}
