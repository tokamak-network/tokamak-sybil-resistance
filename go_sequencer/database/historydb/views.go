package historydb

import (
	"math/big"
	"time"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/common/apitypes"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// MetricsAPI define metrics of the network
type MetricsAPI struct {
	TransactionsPerBatch   float64 `json:"transactionsPerBatch"`
	BatchFrequency         float64 `json:"batchFrequency"`
	TransactionsPerSecond  float64 `json:"transactionsPerSecond"`
	TokenAccounts          int64   `json:"tokenAccounts"`
	Wallets                int64   `json:"wallets"`
	AvgTransactionFee      float64 `json:"avgTransactionFee"`
	EstimatedTimeToForgeL1 float64 `json:"estimatedTimeToForgeL1" meddler:"estimated_time_to_forge_l1"`
}

// BucketParamsAPI are the parameter variables of each Bucket of Rollup Smart
// Contract
type BucketParamsAPI struct {
	CeilUSD         *apitypes.BigIntStr `json:"ceilUSD"`
	BlockStamp      *apitypes.BigIntStr `json:"blockStamp"`
	Withdrawals     *apitypes.BigIntStr `json:"withdrawals"`
	RateBlocks      *apitypes.BigIntStr `json:"rateBlocks"`
	RateWithdrawals *apitypes.BigIntStr `json:"rateWithdrawals"`
	MaxWithdrawals  *apitypes.BigIntStr `json:"maxWithdrawals"`
}

// RollupVariablesAPI are the variables of the Rollup Smart Contract
type RollupVariablesAPI struct {
	EthBlockNum           int64               `json:"ethereumBlockNum" meddler:"eth_block_num"`
	FeeAddToken           *apitypes.BigIntStr `json:"feeAddToken" meddler:"fee_add_token" validate:"required"`
	ForgeL1L2BatchTimeout int64               `json:"forgeL1L2BatchTimeout" meddler:"forge_l1_timeout" validate:"required"`
	WithdrawalDelay       uint64              `json:"withdrawalDelay" meddler:"withdrawal_delay" validate:"required"`
	Buckets               []BucketParamsAPI   `json:"buckets" meddler:"buckets,json"`
	SafeMode              bool                `json:"safeMode" meddler:"safe_mode"`
}

// CoordinatorAPI is a representation of a coordinator with additional information
// required by the API
type CoordinatorAPI struct {
	ItemID      uint64            `json:"itemId" meddler:"item_id"`
	Bidder      ethCommon.Address `json:"bidderAddr" meddler:"bidder_addr"`
	Forger      ethCommon.Address `json:"forgerAddr" meddler:"forger_addr"`
	EthBlockNum int64             `json:"ethereumBlock" meddler:"eth_block_num"`
	URL         string            `json:"URL" meddler:"url"`
	TotalItems  uint64            `json:"-" meddler:"total_items"`
	FirstItem   uint64            `json:"-" meddler:"first_item"`
	LastItem    uint64            `json:"-" meddler:"last_item"`
}

// BatchAPI is a representation of a batch with additional information
// required by the API, and extracted by joining block table
type BatchAPI struct {
	ItemID           uint64                      `json:"itemId" meddler:"item_id"`
	BatchNum         common.BatchNum             `json:"batchNum" meddler:"batch_num"`
	EthereumTxHash   ethCommon.Hash              `json:"ethereumTxHash" meddler:"eth_tx_hash"`
	EthBlockNum      int64                       `json:"ethereumBlockNum" meddler:"eth_block_num"`
	EthBlockHash     ethCommon.Hash              `json:"ethereumBlockHash" meddler:"hash"`
	Timestamp        time.Time                   `json:"timestamp" meddler:"timestamp,utctime"`
	ForgerAddr       ethCommon.Address           `json:"forgerAddr" meddler:"forger_addr"`
	CollectedFeesDB  map[common.TokenID]*big.Int `json:"-" meddler:"fees_collected,json"`
	CollectedFeesAPI apitypes.CollectedFeesAPI   `json:"collectedFees" meddler:"-"`
	TotalFeesUSD     *float64                    `json:"historicTotalCollectedFeesUSD" meddler:"total_fees_usd"`
	StateRoot        apitypes.BigIntStr          `json:"stateRoot" meddler:"state_root"`
	NumAccounts      int                         `json:"numAccounts" meddler:"num_accounts"`
	ExitRoot         apitypes.BigIntStr          `json:"exitRoot" meddler:"exit_root"`
	ForgeL1TxsNum    *int64                      `json:"forgeL1TransactionsNum" meddler:"forge_l1_txs_num"`
	SlotNum          int64                       `json:"slotNum" meddler:"slot_num"`
	ForgedTxs        int                         `json:"forgedTransactions" meddler:"forged_txs"`
	TotalItems       uint64                      `json:"-" meddler:"total_items"`
	FirstItem        uint64                      `json:"-" meddler:"first_item"`
	LastItem         uint64                      `json:"-" meddler:"last_item"`
}
