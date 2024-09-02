package eth

import (
	"math/big"
	"tokamak-sybil-resistance/common"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// RollupForgeBatchArgs are the arguments to the ForgeBatch function in the Rollup Smart Contract
type RollupForgeBatchArgs struct {
	NewLastIdx            int64
	NewStRoot             *big.Int
	NewExitRoot           *big.Int
	L1UserTxs             []common.L1Tx
	L1CoordinatorTxs      []common.L1Tx
	L1CoordinatorTxsAuths [][]byte // Authorization for accountCreations for each L1CoordinatorTx
	L2TxsData             []common.L2Tx
	FeeIdxCoordinator     []common.Idx
	// Circuit selector
	VerifierIdx uint8
	L1Batch     bool
	ProofA      [2]*big.Int
	ProofB      [2][2]*big.Int
	ProofC      [2]*big.Int
}

// RollupEventInitialize is the InitializeHermezEvent event of the
// Smart Contract
type RollupEventInitialize struct {
	ForgeL1L2BatchTimeout uint8
	FeeAddToken           *big.Int
	WithdrawalDelay       uint64
}

// RollupEventL1UserTx is an event of the Rollup Smart Contract
type RollupEventL1UserTx struct {
	// ToForgeL1TxsNum int64 // QueueIndex       *big.Int
	// Position        int   // TransactionIndex *big.Int
	L1UserTx common.L1Tx
}

// RollupEventAddToken is an event of the Rollup Smart Contract
type RollupEventAddToken struct {
	TokenAddress ethCommon.Address
	TokenID      uint32
}

// RollupEventForgeBatch is an event of the Rollup Smart Contract
type RollupEventForgeBatch struct {
	BatchNum int64
	// Sender    ethCommon.Address
	EthTxHash    ethCommon.Hash
	L1UserTxsLen uint16
	GasUsed      uint64
	GasPrice     *big.Int
}

// RollupEventUpdateForgeL1L2BatchTimeout is an event of the Rollup Smart Contract
type RollupEventUpdateForgeL1L2BatchTimeout struct {
	NewForgeL1L2BatchTimeout int64
}

// RollupEventUpdateFeeAddToken is an event of the Rollup Smart Contract
type RollupEventUpdateFeeAddToken struct {
	NewFeeAddToken *big.Int
}

// RollupEventWithdraw is an event of the Rollup Smart Contract
type RollupEventWithdraw struct {
	Idx             uint64
	NumExitRoot     uint64
	InstantWithdraw bool
	TxHash          ethCommon.Hash // Hash of the transaction that generated this event
}

// RollupEventUpdateBucketWithdraw is an event of the Rollup Smart Contract
type RollupEventUpdateBucketWithdraw struct {
	NumBucket   int
	BlockStamp  int64 // blockNum
	Withdrawals *big.Int
}

// RollupEventUpdateWithdrawalDelay is an event of the Rollup Smart Contract
type RollupEventUpdateWithdrawalDelay struct {
	NewWithdrawalDelay uint64
}

// RollupUpdateBucketsParameters are the bucket parameters used in an update
type RollupUpdateBucketsParameters struct {
	CeilUSD         *big.Int
	BlockStamp      *big.Int
	Withdrawals     *big.Int
	RateBlocks      *big.Int
	RateWithdrawals *big.Int
	MaxWithdrawals  *big.Int
}

// RollupEventUpdateBucketsParameters is an event of the Rollup Smart Contract
type RollupEventUpdateBucketsParameters struct {
	ArrayBuckets []RollupUpdateBucketsParameters
	SafeMode     bool
}

// RollupEventUpdateTokenExchange is an event of the Rollup Smart Contract
type RollupEventUpdateTokenExchange struct {
	AddressArray []ethCommon.Address
	ValueArray   []uint64
}

// RollupEventSafeMode is an event of the Rollup Smart Contract
type RollupEventSafeMode struct{}

// RollupEvents is the list of events in a block of the Rollup Smart Contract
type RollupEvents struct {
	L1UserTx                    []RollupEventL1UserTx
	AddToken                    []RollupEventAddToken
	ForgeBatch                  []RollupEventForgeBatch
	UpdateForgeL1L2BatchTimeout []RollupEventUpdateForgeL1L2BatchTimeout
	UpdateFeeAddToken           []RollupEventUpdateFeeAddToken
	Withdraw                    []RollupEventWithdraw
	UpdateWithdrawalDelay       []RollupEventUpdateWithdrawalDelay
	UpdateBucketWithdraw        []RollupEventUpdateBucketWithdraw
	UpdateBucketsParameters     []RollupEventUpdateBucketsParameters
	UpdateTokenExchange         []RollupEventUpdateTokenExchange
	SafeMode                    []RollupEventSafeMode
}

// RollupInterface is the inteface to to Rollup Smart Contract
type RollupInterface interface {
	//
	// Smart Contract Methods
	//

	// Public Functions

	RollupForgeBatch(*RollupForgeBatchArgs, *bind.TransactOpts) (*types.Transaction, error)
	RollupAddToken(tokenAddress ethCommon.Address, feeAddToken,
		deadline *big.Int) (*types.Transaction, error)

	RollupWithdrawMerkleProof(babyPubKey babyjub.PublicKeyComp, tokenID uint32, numExitRoot,
		idx int64, amount *big.Int, siblings []*big.Int, instantWithdraw bool) (*types.Transaction,
		error)
	RollupWithdrawCircuit(proofA, proofC [2]*big.Int, proofB [2][2]*big.Int, tokenID uint32,
		numExitRoot, idx int64, amount *big.Int, instantWithdraw bool) (*types.Transaction, error)

	RollupL1UserTxERC20ETH(fromBJJ babyjub.PublicKeyComp, fromIdx int64, depositAmount *big.Int,
		amount *big.Int, tokenID uint32, toIdx int64) (*types.Transaction, error)
	RollupL1UserTxERC20Permit(fromBJJ babyjub.PublicKeyComp, fromIdx int64,
		depositAmount *big.Int, amount *big.Int, tokenID uint32, toIdx int64,
		deadline *big.Int) (tx *types.Transaction, err error)

	// Governance Public Functions
	RollupUpdateForgeL1L2BatchTimeout(newForgeL1L2BatchTimeout int64) (*types.Transaction, error)
	RollupUpdateFeeAddToken(newFeeAddToken *big.Int) (*types.Transaction, error)

	// Viewers
	RollupRegisterTokensCount() (*big.Int, error)
	RollupLastForgedBatch() (int64, error)

	//
	// Smart Contract Status
	//

	RollupConstants() (*common.RollupConstants, error)
	RollupEventsByBlock(blockNum int64, blockHash *ethCommon.Hash) (*RollupEvents, error)
	RollupForgeBatchArgs(ethCommon.Hash, uint16) (*RollupForgeBatchArgs, *ethCommon.Address, error)
	RollupEventInit(genesisBlockNum int64) (*RollupEventInitialize, int64, error)
}