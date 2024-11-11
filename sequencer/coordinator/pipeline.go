package coordinator

import (
	"context"
	"sync"
	"time"
	"tokamak-sybil-resistance/batchbuilder"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/coordinator/prover"
	"tokamak-sybil-resistance/database/historydb"
	"tokamak-sybil-resistance/database/l2db"
	"tokamak-sybil-resistance/synchronizer"
	"tokamak-sybil-resistance/txselector"
)

type statsVars struct {
	Stats synchronizer.Stats
	Vars  common.SCVariablesPtr
}

type state struct {
	batchNum                     common.BatchNum
	lastScheduledL1BatchBlockNum int64
	lastForgeL1TxsNum            int64
	lastSlotForged               int64
}

// Pipeline manages the forging of batches with parallel server proofs
type Pipeline struct {
	num    int
	cfg    Config
	consts common.SCConsts

	// state
	state         state
	started       bool
	rw            sync.RWMutex
	errAtBatchNum common.BatchNum
	lastForgeTime time.Time

	proversPool           *ProversPool
	provers               []prover.Client
	coord                 *Coordinator
	txManager             *TxManager
	historyDB             *historydb.HistoryDB
	l2DB                  *l2db.L2DB
	txSelector            *txselector.TxSelector
	batchBuilder          *batchbuilder.BatchBuilder
	mutexL2DBUpdateDelete *sync.Mutex
	purger                *Purger

	stats       synchronizer.Stats
	vars        common.SCVariables
	statsVarsCh chan statsVars

	ctx    context.Context
	wg     sync.WaitGroup
	cancel context.CancelFunc
}
