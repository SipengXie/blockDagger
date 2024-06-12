package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

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
	Wg        *sync.WaitGroup
	DbTxs     []kv.Tx
	InputChan chan *ScheduleMessage
}

func NewExecuteLine(blockReader *freezeblocks.BlockReader, ctx context.Context, blk *types.Block, header *types.Header, dbTxs []kv.Tx, wg *sync.WaitGroup, in chan *ScheduleMessage) *ExecuteLine {
	return &ExecuteLine{
		BlkReader: blockReader,
		Ctx:       ctx,
		Blk:       blk,
		Header:    header,
		DbTxs:     dbTxs,
		Wg:        wg,
		InputChan: in,
	}
}

func (e *ExecuteLine) Run() {
	var elapsed int64
	for input := range e.InputChan {
		// fmt.Println("executeline")
		if input.Flag == END {
			e.Wg.Done()
			fmt.Println("Concurrent Execution Cost:", elapsed, "ms")
			return
		}

		processors := input.Processors
		for id, processor := range processors {
			processor.DbTx = e.DbTxs[id]
		}

		st := time.Now()
		var execwg sync.WaitGroup
		execwg.Add(len(processors))
		errMaps := make([]map[int]error, len(processors))
		for id, processor := range processors {
			errMaps[id] = make(map[int]error)
			go processor.Execute(id, e.BlkReader, e.Ctx, e.Blk, e.Header, &execwg, errMaps[id])
		}
		execwg.Wait()
		elapsed += time.Since(st).Milliseconds()
	}
}
