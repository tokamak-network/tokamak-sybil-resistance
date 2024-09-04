package common

import (
	"encoding/binary"
	"fmt"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

const batchNumBytesLen = 8

// Batch is a struct that represents Hermez network batch
type Batch struct {
	BatchNum  BatchNum       `meddler:"batch_num"`
	EthTxHash ethCommon.Hash `meddler:"eth_tx_hash"`
	// Ethereum block in which the batch is forged
	EthBlockNum        int64                `meddler:"eth_block_num"`
	ForgerAddr         ethCommon.Address    `meddler:"forger_addr"`
	CollectedFees      map[TokenID]*big.Int `meddler:"fees_collected,json"`
	FeeIdxsCoordinator []Idx                `meddler:"fee_idxs_coordinator,json"`
	StateRoot          *big.Int             `meddler:"state_root,bigint"`
	NumAccounts        int                  `meddler:"num_accounts"`
	LastIdx            int64                `meddler:"last_idx"`
	ExitRoot           *big.Int             `meddler:"exit_root,bigint"`
	GasUsed            uint64               `meddler:"gas_used"`
	GasPrice           *big.Int             `meddler:"gas_price,bigint"`
	EtherPriceUSD      float64              `meddler:"ether_price_usd"`
	// ForgeL1TxsNum is optional, Only when the batch forges L1 txs. Identifier that corresponds
	// to the group of L1 txs forged in the current batch.
	ForgeL1TxsNum *int64   `meddler:"forge_l1_txs_num"`
	SlotNum       int64    `meddler:"slot_num"` // Slot in which the batch is forged
	TotalFeesUSD  *float64 `meddler:"total_fees_usd"`
}

type BatchNum int64

// BatchData contains the information of a Batch
type BatchData struct {
	L1Batch bool
	// L1UserTxs that were forged in the batch
	L1UserTxs        []L1Tx
	L1CoordinatorTxs []L1Tx
	L2Txs            []L2Tx
	CreatedAccounts  []Account
	UpdatedAccounts  []AccountUpdate
	ExitTree         []ExitInfo
	Batch            Batch
}

// AccountUpdate represents an account balance and/or nonce update after a
// processed batch
type AccountUpdate struct {
	EthBlockNum int64    `meddler:"eth_block_num"`
	BatchNum    BatchNum `meddler:"batch_num"`
	Idx         Idx      `meddler:"idx"`
	Nonce       Nonce    `meddler:"nonce"`
	Balance     *big.Int `meddler:"balance,bigint"`
}

// BatchNumFromBytes returns BatchNum from a []byte
func BatchNumFromBytes(b []byte) (BatchNum, error) {
	if len(b) != batchNumBytesLen {
		return 0,
			Wrap(fmt.Errorf("can not parse BatchNumFromBytes, bytes len %d, expected %d",
				len(b), batchNumBytesLen))
	}
	batchNum := binary.BigEndian.Uint64(b[:batchNumBytesLen])
	return BatchNum(batchNum), nil
}
