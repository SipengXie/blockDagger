package state

import (
	multiversion "blockDagger/multiVersion"
	"blockDagger/rwset"
	"blockDagger/types"

	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon-lib/chain"
	"github.com/ledgerwatch/erigon-lib/common"
	erigonTypes "github.com/ledgerwatch/erigon-lib/types"
	coreTypes "github.com/ledgerwatch/erigon/core/types"
)

type TaskContext struct {
	ID            int
	ReadVersions  map[common.Address]map[common.Hash]*multiversion.Version
	WriteVersions map[common.Address]map[common.Hash]*multiversion.Version
}

func NewTaskContext(task *types.Task) *TaskContext {
	return &TaskContext{
		ID:            task.ID,
		ReadVersions:  task.ReadVersions,
		WriteVersions: task.WriteVersions,
	}
}

type State struct {
	localWrite *localWrite
	snapshot   *localWrite
	taskCtx    *TaskContext
}

func NewState() *State {
	return &State{
		localWrite: newLocalWrite(),
		snapshot:   newLocalWrite(),
	}
}

func (s *State) SetTaskContext(task *types.Task) {
	s.taskCtx = NewTaskContext(task)
}

// --------------- 读写都不能超过Task的ReadVersions和WriteVersions ----------------

func (s *State) isValidRead(addr common.Address, hash common.Hash) bool {
	if (hash == common.Hash{}) {
		// 一个用于判断这个addr在不在ReadVersions里的实现
		_, ok := s.taskCtx.ReadVersions[addr]
		return ok
	}

	if _, ok := s.taskCtx.ReadVersions[addr]; ok {
		if _, ok := s.taskCtx.ReadVersions[addr][hash]; ok {
			return true
		}
	}
	return false
}

func (s *State) isValidWrite(addr common.Address, hash common.Hash) bool {
	if (hash == common.Hash{}) {
		// 一个用于判断这个addr在不在WriteVersions里的实现
		_, ok := s.taskCtx.ReadVersions[addr]
		return ok
	}

	if _, ok := s.taskCtx.WriteVersions[addr]; ok {
		if _, ok := s.taskCtx.WriteVersions[addr][hash]; ok {
			return true
		}
	}
	return false
}

// ---------------- Getters ----------------

func (s *State) GetBalance(addr common.Address) (*uint256.Int, bool) {
	isValid := s.isValidRead(addr, rwset.BALANCE)
	if !isValid {
		return nil, false
	}
	balance, exist := s.localWrite.getBalance(addr)
	if exist {
		return balance, true
	}
	version := (s.taskCtx.ReadVersions[addr][rwset.BALANCE]).GetVisible()
	balance, ok := version.Data.(*uint256.Int)
	if !ok {
		// 这意味预取不出来且valid
		// 这意味着这个balance应该是task自己写的
		// 但还没被task自己写过，我认为这是不会出现的情况
		return nil, false
	}
	return balance, true
}

func (s *State) GetNonce(addr common.Address) (uint64, bool) {
	isValid := s.isValidRead(addr, rwset.NONCE)
	if !isValid {
		return 0, false
	}
	nonce, exist := s.localWrite.getNonce(addr)
	if exist {
		return nonce, true
	}
	version := (s.taskCtx.ReadVersions[addr][rwset.NONCE]).GetVisible()
	nonce, ok := version.Data.(uint64)
	if !ok {
		// 这意味预取不出来且valid
		// 这意味着这个nonce应该是task自己写的
		// 但还没被task自己写过，我认为这是不会出现的情况
		return 0, false
	}
	return nonce, true
}

func (s *State) GetCodeHash(addr common.Address) (common.Hash, bool) {
	isValid := s.isValidRead(addr, rwset.CODEHASH)
	if !isValid {
		return common.Hash{}, false
	}
	codeHash, exist := s.localWrite.getCodeHash(addr)
	if exist {
		return codeHash, true
	}
	version := (s.taskCtx.ReadVersions[addr][rwset.CODEHASH]).GetVisible()
	codeHash, ok := version.Data.(common.Hash)
	if !ok {
		// 这意味预取不出来且valid
		// 这意味着这个codeHash应该是task自己写的
		// 但还没被task自己写过，我认为这是不会出现的情况
		return common.Hash{}, false
	}
	return codeHash, true
}

func (s *State) GetCode(addr common.Address) ([]byte, bool) {
	isValid := s.isValidRead(addr, rwset.CODE)
	if !isValid {
		return nil, false
	}
	code, exist := s.localWrite.getCode(addr)
	if exist {
		return code, true
	}
	version := (s.taskCtx.ReadVersions[addr][rwset.CODE]).GetVisible()
	code, ok := version.Data.([]byte)
	if !ok {
		// 这意味预取不出来且valid
		// 这意味着这个code应该是task自己写的
		// 但还没被task自己写过，我认为这是不会出现的情况
		return nil, false
	}
	return code, true
}

func (s *State) GetCodeSize(addr common.Address) (int, bool) {
	code, valid := s.GetCode(addr)
	if !valid {
		return 0, false
	}
	return len(code), true
}

func (s *State) GetState(addr common.Address, hash *common.Hash, ret *uint256.Int) bool {
	isValid := s.isValidRead(addr, *hash)
	if !isValid {
		return false
	}
	state, exist := s.localWrite.getSlot(addr, *hash)
	if exist {
		ret.Set(state)
		return true
	}

	version := (s.taskCtx.ReadVersions[addr][*hash]).GetVisible()
	state, ok := version.Data.(*uint256.Int)
	if !ok {
		// 这意味预取不出来且valid
		// 这意味着这个state应该是task自己写的
		// 但还没被task自己写过，我认为这是不会出现的情况
		return false
	}
	ret.Set(state)
	return true
}

func (s *State) HasSelfdestructed(addr common.Address) (bool, bool) {
	isValid := s.isValidRead(addr, rwset.ALIVE)
	if !isValid {
		// 默认值是"死"的
		return true, false
	}
	isAlive, exist := s.localWrite.getAlive(addr)
	if exist {
		return !isAlive, true
	}
	version := (s.taskCtx.ReadVersions[addr][rwset.ALIVE]).GetVisible()
	isAlive, ok := version.Data.(bool)
	if !ok {
		// 这意味预取不出来且valid
		// 这意味着这个alive应该是task自己写的
		// 但还没被task自己写过，我认为这是不会出现的情况
		return true, false
	}
	return !isAlive, true
}

func (s *State) Exist(addr common.Address) (bool, bool) {
	isValid := s.isValidRead(addr, common.Hash{})
	if !isValid {
		// 默认是不存在
		return false, false
	}
	// 存在与否应该看看预取出来的结果是否表示该地址存在
	// 然而根据原生State的操作，如果我们预取了不存在地址，它会新建一个Account，从而返回一个确切的值
	// 有两个解决方式：
	// (1) 在预取时增加Exist判断，如果不存在就不预取
	// (2) 在这里直接返回true
	// 根据我对core/vm/evm.go的理解，若不Exist，只会多进行一次CreateAccount
	// 虽然逻辑上 (1) 更完整，但此处我决定使用 (2) 来简化逻辑
	return true, true
}

func (s *State) Empty(addr common.Address) (bool, bool) {
	exist, valid := s.Exist(addr)
	if !valid {
		// 默认是空
		return true, false
	}
	if !exist {
		return true, true
	}
	balance, valid := s.GetBalance(addr)
	if !valid {
		return true, false
	}
	if balance.Sign() != 0 {
		return false, true
	}
	nonce, valid := s.GetNonce(addr)
	if !valid {
		return true, false
	}
	if nonce != 0 {
		return false, true
	}
	codesize, valid := s.GetCodeSize(addr)
	if !valid {
		return true, false

	}
	return codesize == 0, true
}

func (s *State) GetCommittedState(addr common.Address, hash *common.Hash, ret *uint256.Int) bool {
	// TODO: Implement
	valid := s.GetState(addr, hash, ret)
	return valid
}

func (s *State) GetTransientState(addr common.Address, key common.Hash) uint256.Int {
	// TODO: Implement
	return *uint256.NewInt(0)
}

func (s *State) GetRefund() uint64 {
	// TODO: Implement
	return 0
}

// --------------- Setters ----------------
func (s *State) CreateAccount(addr common.Address, _ bool) bool {
	valid := s.isValidWrite(addr, common.Hash{})
	if !valid {
		return false
	}
	s.localWrite.createAccount(addr)
	return true
}

func (s *State) SetBalance(addr common.Address, balance *uint256.Int) bool {
	valid := s.isValidWrite(addr, rwset.BALANCE)
	if !valid {
		return false
	}
	s.localWrite.setBalance(addr, balance)
	return true
}

func (s *State) SetNonce(addr common.Address, nonce uint64) bool {
	valid := s.isValidWrite(addr, rwset.NONCE)
	if !valid {
		return false
	}
	s.localWrite.setNonce(addr, nonce)
	return true
}

func (s *State) SetCode(addr common.Address, code []byte) bool {
	valid := s.isValidWrite(addr, rwset.CODE)
	if !valid {
		return false
	}
	s.localWrite.setCode(addr, code)
	return true
}

func (s *State) SetCodeHash(addr common.Address, codeHash common.Hash) bool {
	valid := s.isValidWrite(addr, rwset.CODEHASH)
	if !valid {
		return false
	}
	s.localWrite.setCodeHash(addr, codeHash)
	return true
}

func (s *State) SetState(addr common.Address, hash *common.Hash, state uint256.Int) bool {
	valid := s.isValidWrite(addr, *hash)
	if !valid {
		return false
	}
	s.localWrite.setSlot(addr, *hash, &state)
	return true
}

func (s *State) SubBalance(addr common.Address, value *uint256.Int) bool {
	balance, valid := s.GetBalance(addr)
	if !valid {
		return false
	}
	balance.Sub(balance, value)
	valid = s.SetBalance(addr, balance)
	return valid
}

func (s *State) AddBalance(addr common.Address, value *uint256.Int) bool {
	balance, valid := s.GetBalance(addr)
	if !valid {
		return false
	}
	balance.Add(balance, value)
	valid = s.SetBalance(addr, balance)
	return valid
}

func (s *State) Selfdestruct(addr common.Address) (bool, bool) {
	valid := s.isValidWrite(addr, rwset.ALIVE)
	if !valid {
		return false, false
	}
	valid = s.SetBalance(addr, uint256.NewInt(0))
	if !valid {
		return false, false
	}
	s.localWrite.setAlive(addr, false)
	return true, true
}

func (s *State) Selfdestruct6780(addr common.Address) bool {
	_, valid := s.Selfdestruct(addr)
	return valid
}

// 这个Snapshot只会记录进入第一次进入VM之前的Localwrite
// 若发生VM Level的Revert，会恢复到进入VM之前的Localwrite
// 进入vm之前的localwrite只会修改balance用来提前买断Gas
// snapshot只会从newLocalWrite()开始记录
func (s *State) Snapshot() {
	for addr, data := range s.localWrite.storage {
		if _, ok := s.snapshot.storage[addr]; !ok {
			s.snapshot.storage[addr] = make(map[common.Hash]interface{})
		}
		for hash, value := range data {
			s.snapshot.storage[addr][hash] = value
		}
	}
}

// simplify
func (s *State) RevertToSnapshot() {
	// 需要清空当前localWrite
	s.localWrite = newLocalWrite()
	for addr, data := range s.snapshot.storage {
		if _, ok := s.localWrite.storage[addr]; !ok {
			s.localWrite.storage[addr] = make(map[common.Hash]interface{})
		}
		for hash, value := range data {
			s.localWrite.storage[addr][hash] = value
		}
	}
}

// ignore
func (s *State) AddRefund(_ uint64) {
	// TODO: Implement
}

// ignore
func (s *State) SubRefund(_ uint64) {
	// TODO: Implement
}

// ignore
func (s *State) SetTransientState(addr common.Address, key common.Hash, value uint256.Int) {
	// TODO: Implement
}

// AddAddressToAccessList adds the given address to the access list. This operation is safe to perform
// even if the feature/fork is not active yet
// ignore
func (s *State) AddAddressToAccessList(addr common.Address) bool {
	return false
	// TODO: Implement
}

// AddSlotToAccessList adds the given (address,slot) to the access list. This operation is safe to perform
// even if the feature/fork is not active yet
// ignore
func (s *State) AddSlotToAccessList(addr common.Address, slot common.Hash) (bool, bool) {
	return false, false
	// TODO: Implement
}

// ignore
func (s *State) Prepare(rules *chain.Rules, sender common.Address, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses erigonTypes.AccessList) {
	// TODO: Implement
}

// ignore
func (s *State) AddLog(*coreTypes.Log) {
	// TODO: Implement
}

// ignore
func (s *State) AddPreimage(_ common.Hash, _ []byte) {
	// TODO: Implement
}

// ignore
func (s *State) AddressInAccessList(addr common.Address) bool {
	// TODO: Implement
	return true
}

// ignore
func (s *State) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	// TODO: Implement
	return true, true
}

// 将所有writeVersion都设为ignore
func (s *State) TotallyAbort() {
	for _, data := range s.taskCtx.WriteVersions {
		for _, version := range data {
			version.Status = multiversion.Ignore
		}
	}
}

// 将所有在localWrite中的writeVersion都设为committed
// 不在localWrite中的writeVersion都设为Ignore
func (s *State) CommitLocalWrite() {
	for addr, data := range s.taskCtx.WriteVersions {
		for hash, version := range data {
			if s.localWrite.contains(addr, hash) {
				// 这里version应该加锁
				version.Status = multiversion.Committed
				version.Data = s.localWrite.storage[addr][hash]
			} else {
				version.Status = multiversion.Ignore
			}
		}
	}
}
