package evmtypes

import (
	"math/big"

	"github.com/holiman/uint256"

	"github.com/ledgerwatch/erigon-lib/chain"
	"github.com/ledgerwatch/erigon-lib/common"
	types2 "github.com/ledgerwatch/erigon-lib/types"

	"github.com/ledgerwatch/erigon/core/types"
)

// BlockContext provides the EVM with auxiliary information. Once provided
// it shouldn't be modified.
type BlockContext struct {
	// CanTransfer returns whether the account contains
	// sufficient ether to transfer the value
	CanTransfer CanTransferFunc
	// Transfer transfers ether from one account to the other
	Transfer TransferFunc
	// GetHash returns the hash corresponding to n
	GetHash GetHashFunc

	// Block information
	Coinbase      common.Address // Provides information for COINBASE
	GasLimit      uint64         // Provides information for GASLIMIT
	MaxGasLimit   bool           // Use GasLimit override for 2^256-1 (to be compatible with OpenEthereum's trace_call)
	BlockNumber   uint64         // Provides information for NUMBER
	Time          uint64         // Provides information for TIME
	Difficulty    *big.Int       // Provides information for DIFFICULTY
	BaseFee       *uint256.Int   // Provides information for BASEFEE
	PrevRanDao    *common.Hash   // Provides information for PREVRANDAO
	ExcessBlobGas *uint64        // Provides information for handling data blobs
}

// TxContext provides the EVM with information about a transaction.
// All fields can change between transactions.
type TxContext struct {
	// Message information
	TxHash     common.Hash
	Origin     common.Address // Provides information for ORIGIN
	GasPrice   *uint256.Int   // Provides information for GASPRICE
	BlobHashes []common.Hash  // Provides versioned blob hashes for BLOBHASH
}

type (
	// CanTransferFunc is the signature of a transfer guard function
	CanTransferFunc func(IntraBlockState, common.Address, *uint256.Int) (bool, bool)
	// TransferFunc is the signature of a transfer function
	TransferFunc func(IntraBlockState, common.Address, common.Address, *uint256.Int, bool)
	// GetHashFunc returns the nth block hash in the blockchain
	// and is used by the BLOCKHASH EVM op code.
	GetHashFunc func(uint64) common.Hash
)

// IntraBlockState is an EVM database for full state querying.
// TODO:
type IntraBlockState interface {
	CreateAccount(common.Address, bool) bool

	SubBalance(common.Address, *uint256.Int) bool
	AddBalance(common.Address, *uint256.Int) bool
	GetBalance(common.Address) (*uint256.Int, bool)

	GetNonce(common.Address) (uint64, bool)
	SetNonce(common.Address, uint64) bool

	GetCodeHash(common.Address) (common.Hash, bool)
	GetCode(common.Address) ([]byte, bool)
	SetCode(common.Address, []byte) bool
	GetCodeSize(common.Address) (int, bool)

	AddRefund(uint64)
	SubRefund(uint64)
	GetRefund() uint64

	GetCommittedState(common.Address, *common.Hash, *uint256.Int) bool
	GetState(address common.Address, slot *common.Hash, outValue *uint256.Int) bool
	SetState(common.Address, *common.Hash, uint256.Int) bool

	GetTransientState(addr common.Address, key common.Hash) uint256.Int
	SetTransientState(addr common.Address, key common.Hash, value uint256.Int)

	Selfdestruct(common.Address) (bool, bool)
	HasSelfdestructed(common.Address) (bool, bool)
	Selfdestruct6780(common.Address) bool

	// Exist reports whether the given account exists in state.
	// Notably this should also return true for suicided accounts.
	Exist(common.Address) (bool, bool)
	// Empty returns whether the given account is empty. Empty
	// is defined according to EIP161 (balance = nonce = code = 0).
	Empty(common.Address) (bool, bool)

	Prepare(rules *chain.Rules, sender, coinbase common.Address, dest *common.Address,
		precompiles []common.Address, txAccesses types2.AccessList)

	AddressInAccessList(addr common.Address) bool
	// AddAddressToAccessList adds the given address to the access list. This operation is safe to perform
	// even if the feature/fork is not active yet
	AddAddressToAccessList(addr common.Address) (addrMod bool)
	// AddSlotToAccessList adds the given (address,slot) to the access list. This operation is safe to perform
	// even if the feature/fork is not active yet
	AddSlotToAccessList(addr common.Address, slot common.Hash) (addrMod, slotMod bool)

	RevertToSnapshot()
	Snapshot()

	AddLog(*types.Log)
}
