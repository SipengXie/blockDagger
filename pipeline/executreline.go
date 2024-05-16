package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/turbo/snapshotsync/freezeblocks"
)

// blockReader *freezeblocks.BlockReader, ctx context.Context, blk *types.Block, header *types.Header, db kv.RoDB
type ExecuteLine struct {
	BlkReader *freezeblocks.BlockReader
	Ctx       context.Context
	Blk       *types.Block
	Header    *types.Header
	Db        kv.RoDB
	Wg        *sync.WaitGroup
	InputChan chan *ScheduleMessage
}

func NewExecuteLine(blockReader *freezeblocks.BlockReader, ctx context.Context, blk *types.Block, header *types.Header, db kv.RoDB, wg *sync.WaitGroup, in chan *ScheduleMessage) *ExecuteLine {
	return &ExecuteLine{
		BlkReader: blockReader,
		Ctx:       ctx,
		Blk:       blk,
		Header:    header,
		Db:        db,
		Wg:        wg,
		InputChan: in,
	}
}

func (e *ExecuteLine) Run() {
	for input := range e.InputChan {
		if input.Flag == END {
			e.Wg.Done()
			return
		}

		processors := input.Processors
		makespan := input.Makespan
		fmt.Println("makespan: ", makespan)

		var execwg sync.WaitGroup
		execwg.Add(len(processors))
		errMaps := make([]map[int]error, len(processors))
		for id, processor := range processors {
			errMaps[id] = make(map[int]error)
			go processor.Execute(e.BlkReader, e.Ctx, e.Blk, e.Header, e.Db, &execwg, errMaps[id])
		}
		execwg.Wait()
		for id, errMap := range errMaps {
			if len(errMap) != 0 {
				fmt.Println("Processor ", id, " has errors: ", len(errMap))
			}
		}
	}
}
