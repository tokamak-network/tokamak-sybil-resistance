package synchronizer

import (
	"context"
	"sync"
	"time"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/database/historydb"
	"tokamak-sybil-resistance/database/l2db"
	"tokamak-sybil-resistance/database/statedb"
	"tokamak-sybil-resistance/eth"
)

// Stats of the synchronizer
type Stats struct {
	Eth struct {
		UpdateBlockNumDiffThreshold uint16
		UpdateFrequencyDivider      uint16
		FirstBlockNum               int64
		LastBlock                   common.Block
		LastBatchNum                int64
	}
	Sync struct {
		Updated   time.Time
		LastBlock common.Block
		LastBatch common.Batch
		// LastL1BatchBlock is the last ethereum block in which an
		// l1Batch was forged
		LastL1BatchBlock  int64
		LastForgeL1TxsNum int64
	}
}

// StatsHolder stores stats and that allows reading and writing them
// concurrently
type StatsHolder struct {
	Stats
	rw sync.RWMutex
}

// NewStatsHolder creates a new StatsHolder
func NewStatsHolder(firstBlockNum int64, updateBlockNumDiffThreshold uint16, updateFrequencyDivider uint16) *StatsHolder {
	stats := Stats{}
	stats.Eth.UpdateBlockNumDiffThreshold = updateBlockNumDiffThreshold
	stats.Eth.UpdateFrequencyDivider = updateFrequencyDivider
	stats.Eth.FirstBlockNum = firstBlockNum
	stats.Sync.LastForgeL1TxsNum = -1
	return &StatsHolder{Stats: stats}
}

// Config is the Synchronizer configuration
type Config struct {
	StatsUpdateBlockNumDiffThreshold uint16
	StatsUpdateFrequencyDivider      uint16
	ChainID                          uint16
}

// Synchronizer implements the Synchronizer type
type Synchronizer struct {
	EthClient        eth.ClientInterface
	consts           common.SCConsts
	historyDB        *historydb.HistoryDB
	l2DB             *l2db.L2DB
	stateDB          *statedb.StateDB
	cfg              Config
	initVars         common.SCVariables
	startBlockNum    int64
	vars             common.SCVariables
	stats            *StatsHolder
	resetStateFailed bool
}

func (s *Synchronizer) Sync (ctx context.Context, lastSavedBlock *common.Block) (blockData *common.BlockData, discarded *int64, err error) {
	return nil, nil, nil
}