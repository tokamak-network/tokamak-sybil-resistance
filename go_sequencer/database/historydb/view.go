package historydb

import (
	"time"
	"tokamak-sybil-resistance/common"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// TokenWithUSD add USD info to common.Token
type TokenWithUSD struct {
	ItemID      uint64            `json:"itemId" meddler:"item_id"`
	TokenID     common.TokenID    `json:"id" meddler:"token_id"`
	EthBlockNum int64             `json:"ethereumBlockNum" meddler:"eth_block_num"` // Ethereum block number in which this token was registered
	EthAddr     ethCommon.Address `json:"ethereumAddress" meddler:"eth_addr"`
	Name        string            `json:"name" meddler:"name"`
	Symbol      string            `json:"symbol" meddler:"symbol"`
	Decimals    uint64            `json:"decimals" meddler:"decimals"`
	USD         *float64          `json:"USD" meddler:"usd"`
	USDUpdate   *time.Time        `json:"fiatUpdate" meddler:"usd_update,utctime"`
	TotalItems  uint64            `json:"-" meddler:"total_items"`
	FirstItem   uint64            `json:"-" meddler:"first_item"`
	LastItem    uint64            `json:"-" meddler:"last_item"`
}
