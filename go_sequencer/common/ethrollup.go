package common

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// BucketParams are the parameter variables of each Bucket of Rollup Smart
// Contract
type BucketParams struct {
	CeilUSD         *big.Int
	BlockStamp      *big.Int
	Withdrawals     *big.Int
	RateBlocks      *big.Int
	RateWithdrawals *big.Int
	MaxWithdrawals  *big.Int
}

// BucketUpdate are the bucket updates (tracking the withdrawals value changes)
// in Rollup Smart Contract
type BucketUpdate struct {
	EthBlockNum int64    `meddler:"eth_block_num"`
	NumBucket   int      `meddler:"num_bucket"`
	BlockStamp  int64    `meddler:"block_stamp"`
	Withdrawals *big.Int `meddler:"withdrawals,bigint"`
}

// TokenExchange are the exchange value for tokens registered in the Rollup
// Smart Contract
type TokenExchange struct {
	EthBlockNum int64             `json:"ethereumBlockNum" meddler:"eth_block_num"`
	Address     ethCommon.Address `json:"address" meddler:"eth_addr"`
	ValueUSD    int64             `json:"valueUSD" meddler:"value_usd"`
}

// RollupVariables are the variables of the Rollup Smart Contract
//nolint:lll
type RollupVariables struct {
	EthBlockNum           int64          `meddler:"eth_block_num"`
	FeeAddToken           *big.Int       `meddler:"fee_add_token,bigint" validate:"required"`
	ForgeL1L2BatchTimeout int64          `meddler:"forge_l1_timeout" validate:"required"`
	Buckets               []BucketParams `meddler:"buckets,json"`
	SafeMode              bool           `meddler:"safe_mode"`
}

// RollupVerifierStruct is the information about verifiers of the Rollup Smart Contract
type RollupVerifierStruct struct {
	MaxTx   int64 `json:"maxTx"`
	NLevels int64 `json:"nlevels"`
}

// RollupConstants are the constants of the Rollup Smart Contract
type RollupConstants struct {
	AbsoluteMaxL1L2BatchTimeout int64                  `json:"absoluteMaxL1L2BatchTimeout"`
	TokenHEZ                    ethCommon.Address      `json:"tokenHEZ"`
	Verifiers                   []RollupVerifierStruct `json:"verifiers"`
	HermezAuctionContract       ethCommon.Address      `json:"hermezAuctionContract"`
	HermezGovernanceAddress     ethCommon.Address      `json:"hermezGovernanceAddress"`
}