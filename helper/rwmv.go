package helper

import (
	multiversion "blockDagger/multiVersion"
	"blockDagger/types"
)

func transferTxToTask(txw types.TransactionWrapper, gVC *multiversion.GlobalVersionChain) *types.Task {
	task := types.NewTask(txw.Tid, txw.Tx.GetGas(), txw.Tx)
	for addr, readSet := range txw.RwSet.ReadSet {
		for hash := range readSet {
			// 先默认依赖snapshot，建图的时候再修改
			task.AddReadVersion(addr, hash, gVC.GetHeadVersion(addr, hash))
			// 顺带预取了
			gVC.SetHeadVersion(addr, hash)
		}
	}
	for addr, writeSet := range txw.RwSet.WriteSet {
		for hash := range writeSet {
			// 新建对应的写version
			newVersion := multiversion.NewVersion(nil, txw.Tid, multiversion.Pending)
			// 更新gVC
			gVC.InsertVersion(addr, hash, newVersion)
			// 写入task
			task.AddWriteVersion(addr, hash, newVersion)
			// 顺带预取了
			gVC.SetHeadVersion(addr, hash)
		}
	}
	return task
}
