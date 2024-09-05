package historydb

import (
	"time"
	"tokamak-sybil-resistance/common"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// NodeConfig contains the node config exposed in the API
type NodeConfig struct {
	MaxPoolTxs uint32
	MinFeeUSD  float64
	MaxFeeUSD  float64
	ForgeDelay float64
}

// NodePublicInfo is the configuration and metrics of the node that is exposed via API
type NodePublicInfo struct {
	// ForgeDelay in seconds
	ForgeDelay float64 `json:"forgeDelay"`
	// PoolLoad amount of transactions in the pool
	PoolLoad int64 `json:"poolLoad"`
}

// Period represents a time period in ethereum
type Period struct {
	SlotNum       int64     `json:"slotNum"`
	FromBlock     int64     `json:"fromBlock"`
	ToBlock       int64     `json:"toBlock"`
	FromTimestamp time.Time `json:"fromTimestamp"`
	ToTimestamp   time.Time `json:"toTimestamp"`
}

// NextForgerAPI represents the next forger exposed via the API
type NextForgerAPI struct {
	Coordinator CoordinatorAPI `json:"coordinator"`
	Period      Period         `json:"period"`
}

// NetworkAPI is the network state exposed via the API
type NetworkAPI struct {
	LastEthBlock  int64           `json:"lastEthereumBlock"`
	LastSyncBlock int64           `json:"lastSynchedBlock"`
	LastBatch     *BatchAPI       `json:"lastBatch"`
	CurrentSlot   int64           `json:"currentSlot"`
	NextForgers   []NextForgerAPI `json:"nextForgers"`
	PendingL1Txs  int             `json:"pendingL1Transactions"`
}

// StateAPI is an object representing the node and network state exposed via the API
type StateAPI struct {
	NodePublicInfo NodePublicInfo        `json:"node"`
	Network        NetworkAPI            `json:"network"`
	Metrics        MetricsAPI            `json:"metrics"`
	Rollup         RollupVariablesAPI    `json:"rollup"`
	RecommendedFee common.RecommendedFee `json:"recommendedFee"`
}

// Constants contains network constants
type Constants struct {
	common.SCConsts
	ChainID       uint16
	HermezAddress ethCommon.Address
}
