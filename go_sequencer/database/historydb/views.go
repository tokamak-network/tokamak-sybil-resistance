package historydb

import (
	"math/big"
	"time"
	"tokamak-sybil-resistance/common"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// txWrite is an representatiion that merges common.L1Tx and common.L2Tx
// in order to perform inserts into tx table
// EffectiveAmount and EffectiveDepositAmount are not set since they have default values in the DB
type txWrite struct {
	// Generic
	IsL1             bool             `meddler:"is_l1"`
	TxID             common.TxID      `meddler:"id"`
	Type             common.TxType    `meddler:"type"`
	Position         int              `meddler:"position"`
	FromIdx          *common.Idx      `meddler:"from_idx"`
	EffectiveFromIdx *common.Idx      `meddler:"effective_from_idx"`
	ToIdx            common.Idx       `meddler:"to_idx"`
	Amount           *big.Int         `meddler:"amount,bigint"`
	AmountFloat      float64          `meddler:"amount_f"`
	TokenID          common.TokenID   `meddler:"token_id"`
	BatchNum         *common.BatchNum `meddler:"batch_num"`     // batchNum in which this tx was forged. If the tx is L2, this must be != 0
	EthBlockNum      int64            `meddler:"eth_block_num"` // Ethereum Block Number in which this L1Tx was added to the queue
	// L1
	ToForgeL1TxsNum    *int64                 `meddler:"to_forge_l1_txs_num"` // toForgeL1TxsNum in which the tx was forged / will be forged
	UserOrigin         *bool                  `meddler:"user_origin"`         // true if the tx was originated by a user, false if it was aoriginated by a coordinator. Note that this differ from the spec for implementation simplification purpposes
	FromEthAddr        *ethCommon.Address     `meddler:"from_eth_addr"`
	FromBJJ            *babyjub.PublicKeyComp `meddler:"from_bjj"`
	DepositAmount      *big.Int               `meddler:"deposit_amount,bigintnull"`
	DepositAmountFloat *float64               `meddler:"deposit_amount_f"`
	EthTxHash          *ethCommon.Hash        `meddler:"eth_tx_hash"`
	L1Fee              *big.Int               `meddler:"l1_fee,bigintnull"`
	// L2
	Fee   *common.FeeSelector `meddler:"fee"`
	Nonce *common.Nonce       `meddler:"nonce"`
}

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
