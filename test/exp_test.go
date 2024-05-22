package test

import (
	"blockDagger/core/vm"
	"blockDagger/helper"
	"blockDagger/pipeline"
	"blockDagger/schedule"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	originCore "github.com/ledgerwatch/erigon/core"
	originTypes "github.com/ledgerwatch/erigon/core/types"
	originVm "github.com/ledgerwatch/erigon/core/vm"
	originEvmTypes "github.com/ledgerwatch/erigon/core/vm/evmtypes"
	"github.com/ledgerwatch/erigon/params"
)

func TestSerialExp(t *testing.T) {
	ctx, dbTx, blkReader, _ := helper.PrepareEnv()
	blockNum := uint64(18999950) // 走50个区块

	for collectSize := uint64(1); collectSize <= 10; collectSize++ {
		fmt.Println("Collect Size: ", collectSize)
		fmt.Println("=============================================")
		// 我们的i是每一个megaBlock的起始位置
		for i := blockNum; i < 19000000; i += collectSize {
			fmt.Println("MegaBlock Start From:", i)

			// 收集collecteSize个区块
			txs := make(originTypes.Transactions, 0)
			for j := i; j < min(i+collectSize, 19000000); j++ {
				block, _ := helper.GetBlockAndHeader(blkReader, ctx, dbTx, j)
				txs = append(txs, block.Transactions()...)
			}

			block, header := helper.GetBlockAndHeader(blkReader, ctx, dbTx, i)
			originblkCtx := helper.GetOriginBlockContext(blkReader, block, dbTx, header)
			ibs := helper.GetState(params.MainnetChainConfig, dbTx, i)

			evm := originVm.NewEVM(originblkCtx, originEvmTypes.TxContext{}, ibs, params.MainnetChainConfig, originVm.Config{})

			st := time.Now()
			for _, tx := range txs {
				msg, _ := tx.AsMessage(*originTypes.LatestSigner(params.MainnetChainConfig), header.BaseFee, evm.ChainRules())

				// Skip the nonce check!
				msg.SetCheckNonce(false)
				txCtx := originCore.NewEVMTxContext(msg)
				evm.TxContext = txCtx

				originCore.ApplyMessage(evm, msg, new(originCore.GasPool).AddGas(header.GasLimit), true /* refunds */, false /* gasBailout */)
			}
			fmt.Println("Serial Execution Cost:", time.Since(st))
			fmt.Println("=============================================")
		}
	}
}

func TestExp(t *testing.T) {
	ctx, dbTx, blkReader, db := helper.PrepareEnv()
	workerNum := min(64, runtime.NumCPU())
	blockNum := uint64(18999950) // 走50个区块

	for collectSize := uint64(10); collectSize <= 10; collectSize++ {
		fmt.Println("Collect Size: ", collectSize)
		fmt.Println("=============================================")
		for i := blockNum; i < 19000000; i += collectSize {
			fmt.Println("MegaBlock Start From:", i)
			// 这里有计时的输出
			_, graph, _, block, header := helper.PrepareBlocks(ctx, dbTx, blkReader, i, min(collectSize, 19000000-i))

			scheduler := schedule.NewScheduler(graph, workerNum)

			st := time.Now()
			processors, makespan := scheduler.Schedule()
			fmt.Println("Schedule Cost:", time.Since(st))
			fmt.Println("makespan: ", makespan)

			var wg sync.WaitGroup
			wg.Add(len(processors))
			errMaps := make([]map[int]error, len(processors))
			st = time.Now()
			for id, processor := range processors {
				errMaps[id] = make(map[int]error)
				go processor.Execute(blkReader, ctx, block, header, db, &wg, errMaps[id])
			}
			wg.Wait()
			fmt.Println("Parallel Execution Time: ", time.Since(st))

			systemAbortCnt := 0
			vmAbort := 0
			for _, errMap := range errMaps {
				for _, err := range errMap {
					if err == vm.ErrSystemAbort {
						systemAbortCnt++
					} else if err != nil {
						vmAbort++
					}

				}
			}
			fmt.Println("System Abort Count: ", systemAbortCnt)
			fmt.Println("VM Abort Count: ", vmAbort)
			fmt.Println("=============================================")
		}
	}
}

func TestPipelineExp(t *testing.T) {
	ctx, dbTx, blkReader, db := helper.PrepareEnv()
	workerNum := min(48, runtime.NumCPU())
	blockNum := uint64(18999950) // 走50个区块
	groupNums := []uint64{25, 17, 13, 10, 9, 8, 6, 5}
	fmt.Println("=============================================")
	for _, groupNum := range groupNums {
		fmt.Println("Group Num: ", groupNum)
		txwsGroup, gvc, block, header := helper.PrepareBlockGroups(ctx, dbTx, blkReader, blockNum, 50, groupNum)
		txwsMsgChan := make(chan *pipeline.TxwsMessage, len(txwsGroup)+2)
		taskMapsAndAccessedByChan := make(chan *pipeline.TaskMapsAndAccessedBy, len(txwsGroup)+2)
		graphMsgChan := make(chan *pipeline.GraphMessage, len(txwsGroup)+2)
		scheduleMsgChan := make(chan *pipeline.ScheduleMessage, len(txwsGroup)+2)

		//初始化四条流水线
		var wg sync.WaitGroup
		gvcLine := pipeline.NewGVCLine(gvc, &wg, txwsMsgChan, taskMapsAndAccessedByChan)
		graphLine := pipeline.NewGraphLine(&wg, taskMapsAndAccessedByChan, graphMsgChan)
		scheduleLine := pipeline.NewScheduleLine(workerNum, &wg, graphMsgChan, scheduleMsgChan)
		executeLine := pipeline.NewExecuteLine(blkReader, ctx, block, header, db, &wg, scheduleMsgChan)

		//向第一条流水线填充交易
		for _, txws := range txwsGroup {
			txwsMsgChan <- &pipeline.TxwsMessage{
				Flag: pipeline.START,
				Txws: txws,
			}
		}
		txwsMsgChan <- &pipeline.TxwsMessage{
			Flag: pipeline.END,
		}
		close(txwsMsgChan)

		//启动四条流水线
		st := time.Now()
		wg.Add(4)
		go executeLine.Run()
		go scheduleLine.Run()
		go graphLine.Run()
		go gvcLine.Run()
		wg.Wait()
		fmt.Println("Pipeline Execution Time: ", time.Since(st))
		fmt.Println("=============================================")
	}

}
