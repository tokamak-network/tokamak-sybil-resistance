package coordinator

import (
	"math/big"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/database/l2db"
	"tokamak-sybil-resistance/synchronizer"

	"github.com/ethereum/go-ethereum/accounts"
)

// TxManager handles everything related to ethereum transactions:  It makes the
// call to forge, waits for transaction confirmation, and keeps checking them
// until a number of confirmed blocks have passed.
type TxManager struct {
	cfg Config
	// ethClient        eth.ClientInterface
	// etherscanService *etherscan.Service
	l2DB    *l2db.L2DB   // Used only to mark forged txs as forged in the L2DB
	coord   *Coordinator // Used only to send messages to stop the pipeline
	batchCh chan *BatchInfo
	chainID *big.Int
	account accounts.Account
	consts  common.SCConsts

	stats       synchronizer.Stats
	vars        common.SCVariables
	statsVarsCh chan statsVars

	discardPipelineCh chan int // int refers to the pipelineNum

	minPipelineNum int
	queue          Queue
	// lastSuccessBatch stores the last BatchNum that who's forge call was confirmed
	lastSuccessBatch common.BatchNum
	// lastPendingBatch common.BatchNum
	// accNonce is the account nonce in the last mined block (due to mined txs)
	accNonce uint64
	// accNextNonce is the nonce that we should use to send the next tx.
	// In some cases this will be a reused nonce of an already pending tx.
	accNextNonce uint64

	lastSentL1BatchBlockNum int64
}

// Queue of BatchInfos
type Queue struct {
	list []*BatchInfo
	// nonceByBatchNum map[common.BatchNum]uint64
	next int
}
