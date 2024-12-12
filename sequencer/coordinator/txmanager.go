package coordinator

import (
	"context"
	"fmt"
	"math/big"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/eth"
	"tokamak-sybil-resistance/etherscan"
	"tokamak-sybil-resistance/log"
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
	// l2DB    *l2db.L2DB   // Used only to mark forged txs as forged in the L2DB
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

// NewQueue returns a new queue
func NewQueue() Queue {
	return Queue{
		list: make([]*BatchInfo, 0),
		// nonceByBatchNum: make(map[common.BatchNum]uint64),
		next: 0,
	}
}

// NewTxManager creates a new TxManager
func NewTxManager(
	ctx context.Context,
	cfg *Config,
	ethClient eth.ClientInterface,
	// l2DB *l2db.L2DB,
	coord *Coordinator,
	scConsts *common.SCConsts,
	initSCVars *common.SCVariables,
	etherscanService *etherscan.Service,
) (
	*TxManager, error) {
	chainID, err := ethClient.EthChainID()
	if err != nil {
		return nil, common.Wrap(err)
	}
	address, err := ethClient.EthAddress()
	if err != nil {
		return nil, common.Wrap(err)
	}
	accNonce, err := ethClient.EthNonceAt(ctx, *address, nil)
	if err != nil {
		return nil, common.Wrap(fmt.Errorf("failed to get nonce: %w", err))
	}
	log.Infow("TxManager started", "nonce", accNonce)
	return &TxManager{
		cfg: *cfg,
		// ethClient:         ethClient,
		// etherscanService:  etherscanService,
		// l2DB:              l2DB,
		coord:             coord,
		batchCh:           make(chan *BatchInfo, queueLen),
		statsVarsCh:       make(chan statsVars, queueLen),
		discardPipelineCh: make(chan int, queueLen),
		account: accounts.Account{
			Address: *address,
		},
		chainID: chainID,
		consts:  *scConsts,

		vars: *initSCVars,

		minPipelineNum: 0,
		queue:          NewQueue(),
		accNonce:       accNonce,
		accNextNonce:   accNonce,
	}, nil
}

// SetSyncStatsVars is a thread safe method to sets the synchronizer Stats
func (t *TxManager) SetSyncStatsVars(ctx context.Context, stats *synchronizer.Stats,
	vars *common.SCVariablesPtr) {
	select {
	case t.statsVarsCh <- statsVars{Stats: *stats, Vars: *vars}:
	case <-ctx.Done():
	}
}

// Run the TxManager
func (t *TxManager) Run(ctx context.Context) {
	// TODO: implement
}

// AddBatch is a thread safe method to pass a new batch TxManager to be sent to
// the smart contract via the forge call
func (t *TxManager) AddBatch(ctx context.Context, batchInfo *BatchInfo) {
	select {
	case t.batchCh <- batchInfo:
	case <-ctx.Done():
	}
}
