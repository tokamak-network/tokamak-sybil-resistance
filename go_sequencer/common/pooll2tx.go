package common

import (
	"math/big"
	"time"
	"tokamak-sybil-resistance/common/nonce"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// PoolL2TxState is a string that represents the status of a L2 transaction
type PoolL2TxState string

type PoolL2Tx struct {
	// Stored in DB: mandatory fields

	// TxID (12 bytes) for L2Tx is:
	// bytes:  |  1   |    6    |   5   |
	// values: | type | FromIdx | Nonce |
	TxID    TxID `meddler:"tx_id"`
	FromIdx Idx  `meddler:"from_idx"`
	ToIdx   Idx  `meddler:"to_idx,zeroisnull"`
	// AuxToIdx is only used internally at the StateDB to avoid repeated
	// computation when processing transactions (from Synchronizer,
	// TxSelector, BatchBuilder)
	AuxToIdx    Idx                   `meddler:"-"`
	ToEthAddr   ethCommon.Address     `meddler:"to_eth_addr,zeroisnull"`
	ToBJJ       babyjub.PublicKeyComp `meddler:"to_bjj,zeroisnull"`
	TokenID     TokenID               `meddler:"token_id"`
	Amount      *big.Int              `meddler:"amount,bigint"`
	Fee         FeeSelector           `meddler:"fee"`
	Nonce       nonce.Nonce           `meddler:"nonce"` // effective 40 bits used
	State       PoolL2TxState         `meddler:"state"`
	MaxNumBatch uint32                `meddler:"max_num_batch,zeroisnull"`
	// Info contains information about the status & State of the
	// transaction. As for example, if the Tx has not been selected in the
	// last batch due not enough Balance at the Sender account, this reason
	// would appear at this parameter.
	Info      string                `meddler:"info,zeroisnull"`
	ErrorCode int                   `meddler:"error_code,zeroisnull"`
	ErrorType string                `meddler:"error_type,zeroisnull"`
	Signature babyjub.SignatureComp `meddler:"signature"`         // tx signature
	Timestamp time.Time             `meddler:"timestamp,utctime"` // time when added to the tx pool
	// Stored in DB: optional fields, may be uninitialized
	AtomicGroupID     AtomicGroupID         `meddler:"atomic_group_id,zeroisnull"`
	RqFromIdx         Idx                   `meddler:"rq_from_idx,zeroisnull"`
	RqToIdx           Idx                   `meddler:"rq_to_idx,zeroisnull"`
	RqToEthAddr       ethCommon.Address     `meddler:"rq_to_eth_addr,zeroisnull"`
	RqToBJJ           babyjub.PublicKeyComp `meddler:"rq_to_bjj,zeroisnull"`
	RqTokenID         TokenID               `meddler:"rq_token_id,zeroisnull"`
	RqAmount          *big.Int              `meddler:"rq_amount,bigintnull"`
	RqFee             FeeSelector           `meddler:"rq_fee,zeroisnull"`
	RqNonce           nonce.Nonce           `meddler:"rq_nonce,zeroisnull"` // effective 48 bits used
	AbsoluteFee       float64               `meddler:"fee_usd,zeroisnull"`
	AbsoluteFeeUpdate time.Time             `meddler:"usd_update,utctimez"`
	Type              TxType                `meddler:"tx_type"`
	RqOffset          uint8                 `meddler:"rq_offset,zeroisnull"` // (max 3 bits)
	// Extra DB write fields (not included in JSON)
	ClientIP string `meddler:"client_ip"`
	// Extra metadata, may be uninitialized
	RqTxCompressedData []byte `meddler:"-"` // 253 bits, optional for atomic txs
	TokenSymbol        string `meddler:"-"` // Used for JSON marshaling the ToIdx
}
