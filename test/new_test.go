package test

import (
	"blockDagger/core/vm"
	"blockDagger/helper"
	multiversion "blockDagger/multiVersion"
	"blockDagger/pipeline"
	"blockDagger/schedule"
	"blockDagger/types"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ledgerwatch/erigon-lib/kv"
	originCore "github.com/ledgerwatch/erigon/core"
	originTypes "github.com/ledgerwatch/erigon/core/types"
	originVm "github.com/ledgerwatch/erigon/core/vm"
	originEvmTypes "github.com/ledgerwatch/erigon/core/vm/evmtypes"
	"github.com/ledgerwatch/erigon/params"
	"github.com/panjf2000/ants/v2"
)

var blockSize []int = []int{200, 400, 600, 800, 1000, 1200, 1400, 1600, 1800, 2000}
var processorNum []int = []int{2, 4, 8, 16, 32, 64} // serial单独测
const blockCount = 500                              // 运行blockCount个区块

func TestSerialOriginal(t *testing.T) {
	ctx, blkReader, db := helper.PrepareEnv()
	dbTx, err := db.BeginRo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	endNum := uint64(19000000)
	startNum := endNum - blockCount

	fatalAborts := 0
	vmAborts := 0

	for i := startNum; i < 19000000; i++ {
		fmt.Println("=============================================")
		fmt.Println("BlockNumber:", i)
		block, header := helper.GetBlockAndHeader(blkReader, ctx, dbTx, i)
		originblkCtx := helper.GetOriginBlockContext(blkReader, block, dbTx, header)

		txs := block.Transactions()
		ibs := helper.GetState(params.MainnetChainConfig, dbTx, i)
		evm := originVm.NewEVM(originblkCtx, originEvmTypes.TxContext{}, ibs, params.MainnetChainConfig, originVm.Config{})
		st := time.Now()
		for _, tx := range txs {
			msg, _ := tx.AsMessage(*originTypes.LatestSigner(params.MainnetChainConfig), header.BaseFee, evm.ChainRules())
			// Skip the nonce check!
			msg.SetCheckNonce(false)
			txCtx := originCore.NewEVMTxContext(msg)
			evm.TxContext = txCtx
			res, err := originCore.ApplyMessage(evm, msg, new(originCore.GasPool).AddGas(header.GasLimit), true /* refunds */, false /* gasBailout */)
			if err != nil {
				fatalAborts++
			} else if res.Err != nil {
				vmAborts++
			}
		}
		fmt.Println("Serial Execution Cost:", time.Since(st))
	}
	fmt.Println("=============================================")
	fmt.Println("Fatal Abort:", fatalAborts)
	fmt.Println("VM Abort:", vmAborts)
}

// 该函数测试了在MegaBlock情况下的Serial执行情况
// 我们还需要测试一下原生block的Serial执行情况
func TestSerial(t *testing.T) {
	ctx, blkReader, db := helper.PrepareEnv()
	dbTx, err := db.BeginRo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	endNum := uint64(19000000)
	startNum := endNum - blockCount

	for _, size := range blockSize {
		fmt.Println("=============================================")
		fmt.Println("Block Size: ", size)
		// 准备执行环境，然后串行执行
		txwsArray, startBlock, startHeader := helper.PrepareTransactions(ctx, dbTx, blkReader, startNum, endNum, uint64(size), false /*no need rwset*/)
		originCtx := helper.GetOriginBlockContext(blkReader, startBlock, dbTx, startHeader)
		ibs := helper.GetState(params.MainnetChainConfig, dbTx, startNum)
		evm := originVm.NewEVM(originCtx, originEvmTypes.TxContext{}, ibs, params.MainnetChainConfig, originVm.Config{})

		fmt.Println("Block Number:", len(txwsArray))
		fmt.Println("Total Number of Txs:", (len(txwsArray)-1)*size+len(txwsArray[len(txwsArray)-1]))
		vmAbort, fatalAbort := 0, 0
		for i, txws := range txwsArray {
			st := time.Now()
			for _, txw := range txws {
				msg, _ := txw.Tx.AsMessage(*originTypes.LatestSigner(params.MainnetChainConfig), startHeader.BaseFee, evm.ChainRules())
				// Skip the nonce check!
				msg.SetCheckNonce(false)
				txCtx := originCore.NewEVMTxContext(msg)
				evm.TxContext = txCtx

				res, err := originCore.ApplyMessage(evm, msg, new(originCore.GasPool).AddGas(msg.Gas()), true /* refunds */, false /* gasBailout */)
				if err != nil {
					fatalAbort++
				} else if res.Err != nil {
					vmAbort++
				}
			}
			fmt.Println("Block Number:", i, "Execution Cost:", time.Since(st))
		}
		fmt.Println("Fatal Abort:", fatalAbort)
		fmt.Println("VM Abort:", vmAbort)
	}
	fmt.Println("=============================================")
}

// 该函数测试在MegaBlock情况下的并行执行情况
// TODO:对于DBTX的优化只在NoPipeline做了，Pipeline那边还要做
func TestNoPipeline(t *testing.T) {
	ctx, blkReader, db := helper.PrepareEnv()
	dbTx, err := db.BeginRo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	endNum := uint64(19000000)
	startNum := endNum - blockCount
	for _, size := range blockSize {
		fmt.Println("=============================================")
		fmt.Println("Block Size: ", size)
		// 准备分组交易
		txwsArray, startBlock, startHeader := helper.PrepareTransactions(ctx, dbTx, blkReader, startNum, endNum, uint64(size), true /*need rwset*/)

		fmt.Println("Block Number:", len(txwsArray))
		fmt.Println("Total Number of Txs:", (len(txwsArray)-1)*size+len(txwsArray[len(txwsArray)-1]))

		for _, threadNum := range processorNum {
			fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
			fmt.Println("Processor Number: ", threadNum)

			// 为每一个processor准备dbtx
			dbTxs := make([]kv.Tx, threadNum)
			for i := 0; i < threadNum; i++ {
				dbTxs[i], err = db.BeginRo(ctx)
				if err != nil {
					t.Fatal(err)
				}
			}

			// 为执行准备环境
			ibs := helper.GetState(params.MainnetChainConfig, dbTx, startNum)
			gvc := multiversion.NewGlobalVersionChain(ibs)

			pool, _ := ants.NewPool(threadNum)
			for i, txws := range txwsArray {
				fmt.Println("---------------------------------------------")
				fmt.Println("Block Number:", i)
				// Pre-Processing
				st := time.Now()
				rwAccessedBy := helper.GenerateAccessedBy(txws)
				taskMap := make(map[int]*types.Task)
				for _, txw := range txws {
					task := helper.TransferTxToTask(*txw, gvc)
					taskMap[task.ID] = task
				}
				gvc.UpdateLastBlockTail()
				fmt.Println("Pre-Processing Cost:", time.Since(st))

				// Graph Generation
				st = time.Now()
				graph := helper.GenerateGraph(taskMap, rwAccessedBy)
				fmt.Println("Graph Generation Cost:", time.Since(st))

				// Parallel Schedule
				st = time.Now()
				scheduler := schedule.NewScheduler(graph, threadNum)
				processors, makespan := scheduler.Schedule()
				fmt.Println("Parallel Schedule Cost:", time.Since(st))

				for id, processor := range processors {
					processor.DbTx = dbTxs[id]
				}
				// Concurrent Execution
				st = time.Now()
				var wg sync.WaitGroup
				wg.Add(len(processors))
				errMaps := make([]map[int]error, len(processors))
				for id, processor := range processors {
					errMaps[id] = make(map[int]error)
					// err := pool.Submit(func() {
					// 	processor.Execute(blkReader, ctx, startBlock, startHeader, db, &wg, errMaps[id], out)
					// })
					// if err != nil {
					// 	fmt.Println("Error Submitting Task")
					// 	wg.Done()
					// }
					go processor.Execute(id, blkReader, ctx, startBlock, startHeader, &wg, errMaps[id])
				}
				wg.Wait()
				fmt.Println("Concurrent Execution Cost:", time.Since(st))

				// Schedule Result
				fmt.Println("Critical Path Length:", graph.CriticalPathLen)
				fmt.Println("Makespan:", makespan)

				// Error Analysis
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
			}
			pool.Release()
		}
	}
	fmt.Println("=============================================")
}

// 该函数测试在MegaBlock情况下Pipeline的并行执行情况
func TestPipeline(t *testing.T) {
	ctx, blkReader, db := helper.PrepareEnv()
	dbTx, err := db.BeginRo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	endNum := uint64(19000000)
	startNum := endNum - blockCount

	for _, size := range blockSize {
		fmt.Println("=============================================")
		fmt.Println("Block Size: ", size)
		// 准备分组交易
		txwsArray, startBlock, startHeader := helper.PrepareTransactions(ctx, dbTx, blkReader, startNum, endNum, uint64(size), true /*need rwset*/)

		fmt.Println("Block Number:", len(txwsArray))
		fmt.Println("Total Number of Txs:", (len(txwsArray)-1)*size+len(txwsArray[len(txwsArray)-1]))

		for _, threadNum := range processorNum {
			fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
			fmt.Println("Processor Number: ", threadNum)

			// 为每一个processor准备dbtx
			dbTxs := make([]kv.Tx, threadNum)
			for i := 0; i < threadNum; i++ {
				dbTxs[i], err = db.BeginRo(ctx)
				if err != nil {
					t.Fatal(err)
				}
			}

			// 为执行准备环境
			ibs := helper.GetState(params.MainnetChainConfig, dbTx, startNum)
			gvc := multiversion.NewGlobalVersionChain(ibs)

			// 准备消息通道
			txwsMsgChan := make(chan *pipeline.TxwsMessage, len(txwsArray)+2)
			taskMapsAndAccessedByChan := make(chan *pipeline.TaskMapsAndAccessedBy, len(txwsArray)+2)
			graphMsgChan := make(chan *pipeline.GraphMessage, len(txwsArray)+2)
			scheduleMsgChan := make(chan *pipeline.ScheduleMessage, len(txwsArray)+2)

			var wg sync.WaitGroup
			gvcLine := pipeline.NewGVCLine(gvc, &wg, txwsMsgChan, taskMapsAndAccessedByChan)
			graphLine := pipeline.NewGraphLine(&wg, taskMapsAndAccessedByChan, graphMsgChan)
			scheduleLine := pipeline.NewScheduleLine(threadNum, &wg, graphMsgChan, scheduleMsgChan)
			executeLine := pipeline.NewExecuteLine(blkReader, ctx, startBlock, startHeader, dbTxs, &wg, scheduleMsgChan)

			//向第一条流水线填充交易
			for _, txws := range txwsArray {
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
		}
	}
	fmt.Println("=============================================")
}

func TestSize(t *testing.T) {
	ctx, blkReader, db := helper.PrepareEnv()
	dbTx, err := db.BeginRo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	endNum := uint64(19000000)
	intervals := []uint64{100, 200, 300, 400, 500}

	// 我们将测试100, 200, 300, 400 ,500个区块中交易平均长度以及RwSet平均大小
	for _, interval := range intervals {
		startNum := endNum - interval

		// 准备交易
		txwsArray, _, _ := helper.PrepareTransactions(ctx, dbTx, blkReader, startNum, endNum, 10000 /*用大一点的组容量尽量快点返回*/, true /*need rwset*/)
		totalTx := (len(txwsArray)-1)*10000 + len(txwsArray[len(txwsArray)-1])

		// 计算平均交易长度和RwSet大小
		totalTxLen, totalRwSetLen := 0, 0
		for _, txws := range txwsArray {
			for _, txw := range txws {
				totalTxLen += txw.Tx.EncodingSize()
				if txw.RwSet != nil {
					for _, r := range txw.RwSet.ReadSet {
						totalRwSetLen += 20
						totalRwSetLen += 32 * len(r)
					}
					for _, w := range txw.RwSet.WriteSet {
						totalRwSetLen += 20
						totalRwSetLen += 32 * len(w)
					}
				}
			}
		}
		fmt.Println("Interval:", interval)
		fmt.Println("Average Transaction Length:", float64(totalTxLen)/float64(totalTx))
		fmt.Println("Average RwSet Length:", float64(totalRwSetLen)/float64(totalTx))
	}
}

func TestOriginSchedule(t *testing.T) {
	ctx, blkReader, db := helper.PrepareEnv()
	dbTx, err := db.BeginRo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	endNum := uint64(19000000)
	startNum := endNum - blockCount
	for _, size := range blockSize {
		fmt.Println("=============================================")
		fmt.Println("Block Size: ", size)
		// 准备分组交易
		txwsArray, _, _ := helper.PrepareTransactions(ctx, dbTx, blkReader, startNum, endNum, uint64(size), true /*need rwset*/)

		for _, threadNum := range processorNum {
			fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
			fmt.Println("Processor Number: ", threadNum)

			// 为执行准备环境
			ibs := helper.GetState(params.MainnetChainConfig, dbTx, startNum)
			gvc := multiversion.NewGlobalVersionChain(ibs)

			for i, txws := range txwsArray {
				fmt.Println("---------------------------------------------")
				fmt.Println("Block Number:", i)

				// Pre-Processing
				rwAccessedBy := helper.GenerateAccessedBy(txws)
				taskMap := make(map[int]*types.Task)
				for _, txw := range txws {
					task := helper.TransferTxToTask(*txw, gvc)
					taskMap[task.ID] = task
				}
				gvc.UpdateLastBlockTail()

				// Graph Generation
				graph := helper.GenerateGraph(taskMap, rwAccessedBy)

				// Parallel Schedule
				st := time.Now()
				// scheduler := scheduleOrigin.NewScheduler(graph, threadNum)
				scheduler := schedule.NewScheduler(graph, threadNum)
				_, makespan := scheduler.Schedule()
				fmt.Println("Original Parallel Schedule Cost:", time.Since(st))

				// Schedule Result
				fmt.Println("Critical Path Length:", graph.CriticalPathLen)
				fmt.Println("Makespan:", makespan)
			}
		}
	}
	fmt.Println("=============================================")
}
