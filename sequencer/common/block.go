package common

import (
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// Block represents of an Ethereum block
type Block struct {
	Num        int64          `meddler:"eth_block_num"`
	Timestamp  time.Time      `meddler:"timestamp,utctime"`
	Hash       ethCommon.Hash `meddler:"hash"`
	ParentHash ethCommon.Hash `meddler:"-" json:"-"`
}

// BlockData contains the information of a Block
type BlockData struct {
	Block  Block
	Rollup RollupData
}

// RollupData contains information returned by the Rollup smart contract
type RollupData struct {
	// L1UserTxs that were submitted in the block
	L1UserTxs            []L1Tx
	Batches              []BatchData
	Withdrawals          []WithdrawInfo
	UpdateBucketWithdraw []BucketUpdate
	Vars                 *RollupVariables
	AddedTokens          []Token
}

// NewRollupData creates an empty RollupData with the slices initialized.
func NewRollupData() RollupData {
	return RollupData{
		L1UserTxs:   make([]L1Tx, 0),
		Batches:     make([]BatchData, 0),
		Withdrawals: make([]WithdrawInfo, 0),
		Vars:        nil,
	}
}
