package helper

import (
	multiversion "blockDagger/multiVersion"
	"blockDagger/types"
)

func transferTxToTask(txw types.TransactionWrapper, mvState *multiversion.GlobalVersionChain) *types.Task {
	task := types.NewTask(txw.Tid, txw.Tx.GetGas(), txw.Tx)
	for addr, readSet := range txw.RwSet.ReadSet {
		for hash := range readSet {
			// 先默认依赖snapshot，建图的时候再修改
			task.AddReadVersion(addr, hash, mvState.GetHeadVersion(addr, hash))
		}
	}
	for addr, writeSet := range txw.RwSet.WriteSet {
		for hash := range writeSet {
			// 新建对应的写version
			newVersion := multiversion.NewVersion(nil, txw.Tid, multiversion.Unscheduled)
			// 更新mvState
			mvState.InsertVersion(addr, hash, newVersion)
			// 写入task
			task.AddWriteVersion(addr, hash, newVersion)
		}
	}
	return task
}
