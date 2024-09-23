package eth

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/eth/contracts/tokamak"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
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

// TODO: Update interfaces and the functions
// RollupInterface is the inteface to to Rollup Smart Contract
type RollupInterface interface {
	//
	// Smart Contract Methods
	//

	// Public Functions

	// RollupForgeBatch(*RollupForgeBatchArgs, *bind.TransactOpts) (*types.Transaction, error)

	// RollupWithdrawMerkleProof(babyPubKey babyjub.PublicKeyComp, tokenID uint32, numExitRoot,
	// 	idx int64, amount *big.Int, siblings []*big.Int, instantWithdraw bool) (*types.Transaction,
	// 	error)
	// RollupWithdrawCircuit(proofA, proofC [2]*big.Int, proofB [2][2]*big.Int, tokenID uint32,
	// 	numExitRoot, idx int64, amount *big.Int, instantWithdraw bool) (*types.Transaction, error)

	// Governance Public Functions
	// RollupUpdateForgeL1L2BatchTimeout(newForgeL1L2BatchTimeout int64) (*types.Transaction, error)

	// Viewers
	RollupLastForgedBatch() (int64, error)

	//
	// Smart Contract Status
	//

	RollupConstants() (*common.RollupConstants, error)
	// RollupEventsByBlock(blockNum int64, blockHash *ethCommon.Hash) (*RollupEvents, error)
	// RollupForgeBatchArgs(ethCommon.Hash, uint16) (*RollupForgeBatchArgs, *ethCommon.Address, error)
	RollupEventInit(genesisBlockNum int64) (*RollupEventInitialize, int64, error)
}

//
// Implementation
//

// RollupClient is the implementation of the interface to the Rollup Smart Contract in ethereum.
type RollupClient struct {
	client      *EthereumClient
	chainID     *big.Int
	address     ethCommon.Address
	tokamak     *tokamak.Tokamak
	contractAbi abi.ABI
	opts        *bind.CallOpts
	consts      *common.RollupConstants
}

// RollupVariables returns the RollupVariables from the initialize event
func (ei *RollupEventInitialize) RollupVariables() *common.RollupVariables {
	return &common.RollupVariables{
		EthBlockNum:           0,
		ForgeL1L2BatchTimeout: int64(ei.ForgeL1L2BatchTimeout),
		Buckets:               []common.BucketParams{},
		SafeMode:              false,
	}
}

// TODO: Check and Update theses constants
var (
	// 	logHermezL1UserTxEvent = crypto.Keccak256Hash([]byte(
	// 		"L1UserTxEvent(uint32,uint8,bytes)"))
	// 	logHermezForgeBatch = crypto.Keccak256Hash([]byte(
	// 		"ForgeBatch(uint32,uint16)"))
	// 	logHermezUpdateForgeL1L2BatchTimeout = crypto.Keccak256Hash([]byte(
	// 		"UpdateForgeL1L2BatchTimeout(uint8)"))
	// 	logHermezWithdrawEvent = crypto.Keccak256Hash([]byte(
	// 		"WithdrawEvent(uint48,uint32,bool)"))
	// 	logHermezUpdateBucketWithdraw = crypto.Keccak256Hash([]byte(
	// 		"UpdateBucketWithdraw(uint8,uint256,uint256)"))
	// 	logHermezUpdateBucketsParameters = crypto.Keccak256Hash([]byte(
	// 		"UpdateBucketsParameters(uint256[])"))
	// 	logHermezSafeMode = crypto.Keccak256Hash([]byte(
	// 		"SafeMode()"))
	logHermezInitialize = crypto.Keccak256Hash([]byte(
		""))
)

// RollupEventInit returns the initialize event with its corresponding block number
func (c *RollupClient) RollupEventInit(genesisBlockNum int64) (*RollupEventInitialize, int64, error) {
	query := ethereum.FilterQuery{
		Addresses: []ethCommon.Address{
			c.address,
		},
		FromBlock: big.NewInt(max(0, genesisBlockNum-blocksPerDay)),
		ToBlock:   big.NewInt(genesisBlockNum),
		Topics:    [][]ethCommon.Hash{{logHermezInitialize}},
	}
	logs, err := c.client.client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, 0, common.Wrap(err)
	}
	if len(logs) != 1 {
		return nil, 0, common.Wrap(fmt.Errorf("no event of type InitializeHermezEvent found"))
	}
	vLog := logs[0]
	if vLog.Topics[0] != logHermezInitialize {
		return nil, 0, common.Wrap(fmt.Errorf("event is not InitializeHermezEvent"))
	}

	var rollupInit RollupEventInitialize
	if err := c.contractAbi.UnpackIntoInterface(&rollupInit, "InitializeHermezEvent",
		vLog.Data); err != nil {
		return nil, 0, common.Wrap(err)
	}
	return &rollupInit, int64(vLog.BlockNumber), common.Wrap(err)
}

// NewRollupClient creates a new RollupClient
func NewRollupClient(client *EthereumClient, address ethCommon.Address) (*RollupClient, error) {
	contractAbi, err := abi.JSON(strings.NewReader(string(tokamak.TokamakABI)))
	if err != nil {
		return nil, common.Wrap(err)
	}
	tokamak, err := tokamak.NewTokamak(address, client.Client())
	if err != nil {
		return nil, common.Wrap(err)
	}
	chainID, err := client.EthChainID()
	if err != nil {
		return nil, common.Wrap(err)
	}
	c := &RollupClient{
		client:      client,
		chainID:     chainID,
		address:     address,
		tokamak:     tokamak,
		contractAbi: contractAbi,
		opts:        newCallOpts(),
	}
	consts, err := c.RollupConstants()
	if err != nil {
		return nil, common.Wrap(fmt.Errorf("RollupConstants at %v: %w", address, err))
	}
	c.consts = consts
	// c.token, err = NewTokenClient(client, consts.TokenHEZ)
	// if err != nil {
	// 	return nil, common.Wrap(fmt.Errorf("new token client at %v: %w", address, err))
	// }
	return c, nil
}

// RollupConstants returns the Constants of the Rollup Smart Contract
func (c *RollupClient) RollupConstants() (rollupConstants *common.RollupConstants, err error) {
	rollupConstants = new(common.RollupConstants)
	if err := c.client.Call(func(ec *ethclient.Client) error {
		absoluteMaxL1L2BatchTimeout, err := c.tokamak.ABSOLUTEMAXL1L2BATCHTIMEOUT(c.opts)
		if err != nil {
			return common.Wrap(err)
		}
		rollupConstants.AbsoluteMaxL1L2BatchTimeout = int64(absoluteMaxL1L2BatchTimeout)
		// rollupConstants.TokenHEZ, err = c.tokamak.TokenHEZ(c.opts)
		if err != nil {
			return common.Wrap(err)
		}
		rollupVerifiersLength, err := c.tokamak.RollupVerifiersLength(c.opts)
		if err != nil {
			return common.Wrap(err)
		}
		for i := int64(0); i < rollupVerifiersLength.Int64(); i++ {
			var newRollupVerifier common.RollupVerifierStruct
			rollupVerifier, err := c.tokamak.RollupVerifiers(c.opts, big.NewInt(i))
			if err != nil {
				return common.Wrap(err)
			}
			newRollupVerifier.MaxTx = rollupVerifier.MaxTx.Int64()
			newRollupVerifier.NLevels = rollupVerifier.NLevels.Int64()
			rollupConstants.Verifiers = append(rollupConstants.Verifiers,
				newRollupVerifier)
		}
		// rollupConstants.HermezAuctionContract, err = c.hermez.HermezAuctionContract(c.opts)
		// if err != nil {
		// 	return common.Wrap(err)
		// }
		// rollupConstants.HermezGovernanceAddress, err = c.hermez.HermezGovernanceAddress(c.opts)
		// if err != nil {
		// 	return common.Wrap(err)
		// }
		// rollupConstants.WithdrawDelayerContract, err = c.hermez.WithdrawDelayerContract(c.opts)
		return common.Wrap(err)
	}); err != nil {
		return nil, common.Wrap(err)
	}
	return rollupConstants, nil
}

// RollupLastForgedBatch is the interface to call the smart contract function
func (c *RollupClient) RollupLastForgedBatch() (lastForgedBatch int64, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_lastForgedBatch, err := c.tokamak.LastForgedBatch(c.opts)
		lastForgedBatch = int64(_lastForgedBatch)
		return common.Wrap(err)
	}); err != nil {
		return 0, common.Wrap(err)
	}
	return lastForgedBatch, nil
}
