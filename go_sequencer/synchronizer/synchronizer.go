package synchronizer

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/database/historydb"
	"tokamak-sybil-resistance/database/l2db"
	"tokamak-sybil-resistance/database/statedb"
	"tokamak-sybil-resistance/eth"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

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
		Auction           struct {
			CurrentSlot common.Slot
			NextSlot    common.Slot
		}
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

// UpdateSync updates the synchronizer stats
func (s *StatsHolder) UpdateSync(lastBlock *common.Block, lastBatch *common.Batch,
	lastL1BatchBlock *int64, lastForgeL1TxsNum *int64) {
	now := time.Now()
	s.rw.Lock()
	s.Sync.LastBlock = *lastBlock
	if lastBatch != nil {
		s.Sync.LastBatch = *lastBatch
	}
	if lastL1BatchBlock != nil {
		s.Sync.LastL1BatchBlock = *lastL1BatchBlock
		s.Sync.LastForgeL1TxsNum = *lastForgeL1TxsNum
	}
	s.Sync.Updated = now
	s.rw.Unlock()
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
	consts           *common.SCConsts
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

// NewSynchronizer creates a new Synchronizer
func NewSynchronizer(ethClient eth.ClientInterface, historyDB *historydb.HistoryDB,
	l2DB *l2db.L2DB, stateDB *statedb.StateDB, cfg Config) (*Synchronizer, error) {
	rollupConstants, err := ethClient.RollupConstants()
	if err != nil {
		return nil, common.Wrap(fmt.Errorf("NewSynchronizer ethClient.RollupConstants(): %w",
			err))
	}

	consts := common.SCConsts{
		Rollup: *rollupConstants,
	}

	initVars, startBlockNum, err := getInitialVariables(ethClient, &consts)
	if err != nil {
		return nil, common.Wrap(err)
	}

	stats := NewStatsHolder(startBlockNum, cfg.StatsUpdateBlockNumDiffThreshold, cfg.StatsUpdateFrequencyDivider)
	s := &Synchronizer{
		EthClient:     ethClient,
		consts:        &consts,
		historyDB:     historyDB,
		l2DB:          l2DB,
		stateDB:       stateDB,
		cfg:           cfg,
		initVars:      *initVars,
		startBlockNum: startBlockNum,
		stats:         stats,
	}
	return s, s.init()
}

func (s *Synchronizer) Sync(ctx context.Context, lastSavedBlock *common.Block) (blockData *common.BlockData, discarded *int64, err error) {
	return nil, nil, nil
}

func getInitialVariables(ethClient eth.ClientInterface,
	consts *common.SCConsts) (*common.SCVariables, int64, error) {
	rollupInit, rollupInitBlock, err := ethClient.RollupEventInit(consts.Rollup.GenesisBlockNum) //TODO: Check this with hermuz code
	if err != nil {
		return nil, 0, common.Wrap(fmt.Errorf("RollupEventInit: %w", err))
	}
	rollupVars := rollupInit.RollupVariables()
	return &common.SCVariables{
		Rollup: *rollupVars,
	}, rollupInitBlock, nil
}

func (s *Synchronizer) init() error {
	// Update stats parameters so that they have valid values before the
	// first Sync call
	if err := s.stats.UpdateEth(s.EthClient); err != nil {
		return common.Wrap(err)
	}
	lastBlock := &common.Block{}
	lastSavedBlock, err := s.historyDB.GetLastBlock()
	// `s.historyDB.GetLastBlock()` will never return `sql.ErrNoRows`
	// because we always have the default block 0 in the DB
	if err != nil {
		return common.Wrap(err)
	}
	// If we only have the default block 0,
	// make sure that the stateDB is clean
	if lastSavedBlock.Num == 0 {
		if err := s.stateDB.Reset(0); err != nil {
			return common.Wrap(err)
		}
	} else {
		lastBlock = lastSavedBlock
	}

	if err := s.resetState(lastBlock); err != nil {
		s.resetStateFailed = true
		return common.Wrap(err)
	}
	s.resetStateFailed = false

	log.Info("Sync init block",
		"syncLastBlock", s.stats.Sync.LastBlock,
		"ethFirstBlockNum", s.stats.Eth.FirstBlockNum,
		"ethLastBlock", s.stats.Eth.LastBlock,
	)
	log.Info("Sync init batch",
		"syncLastBatch", s.stats.Sync.LastBatch.BatchNum,
		"ethLastBatch", s.stats.Eth.LastBatchNum,
	)
	return nil
}

// UpdateEth updates the ethereum stats, only if the previous stats expired
func (s *StatsHolder) UpdateEth(ethClient eth.ClientInterface) error {
	lastBlock, err := ethClient.EthBlockByNumber(context.TODO(), -1)
	if err != nil {
		return common.Wrap(fmt.Errorf("EthBlockByNumber: %w", err))
	}
	lastBatchNum, err := ethClient.RollupLastForgedBatch()
	if err != nil {
		return common.Wrap(fmt.Errorf("RollupLastForgedBatch: %w", err))
	}
	s.rw.Lock()
	s.Eth.LastBlock = *lastBlock
	s.Eth.LastBatchNum = lastBatchNum
	s.rw.Unlock()
	return nil
}

func (s *Synchronizer) resetState(block *common.Block) error {
	rollup, err := s.historyDB.GetSCVars()
	// If SCVars are not in the HistoryDB, this is probably the first run
	// of the Synchronizer: store the initial vars taken from config
	if common.Unwrap(err) == sql.ErrNoRows {
		vars := s.initVars
		log.Info("Setting initial SCVars in HistoryDB")
		if err = s.historyDB.SetInitialSCVars(&vars.Rollup); err != nil {
			return common.Wrap(fmt.Errorf("historyDB.SetInitialSCVars: %w", err))
		}
		s.vars.Rollup = *vars.Rollup.Copy()
		// Add initial boot coordinator to HistoryDB
		if err := s.historyDB.AddCoordinators([]common.Coordinator{{
			Forger:      ethCommon.HexToAddress(os.Getenv("BootCoordinator")),
			URL:         os.Getenv("BootCoordinatorURL"),
			EthBlockNum: s.initVars.Rollup.EthBlockNum, //TODO: Check this with Eth Block
		}}); err != nil {
			return common.Wrap(err)
		}
	} else if err != nil {
		return common.Wrap(err)
	} else {
		s.vars.Rollup = *rollup
	}

	batch, err := s.historyDB.GetLastBatch()
	if err != nil && common.Unwrap(err) != sql.ErrNoRows {
		return common.Wrap(fmt.Errorf("historyDB.GetLastBatchNum: %w", err))
	}
	if common.Unwrap(err) == sql.ErrNoRows {
		batch = &common.Batch{}
	}

	err = s.stateDB.Reset(batch.BatchNum)
	if err != nil {
		return common.Wrap(fmt.Errorf("stateDB.Reset: %w", err))
	}

	lastL1BatchBlockNum, err := s.historyDB.GetLastL1BatchBlockNum()
	if err != nil && common.Unwrap(err) != sql.ErrNoRows {
		return common.Wrap(fmt.Errorf("historyDB.GetLastL1BatchBlockNum: %w", err))
	}
	if common.Unwrap(err) == sql.ErrNoRows {
		lastL1BatchBlockNum = 0
	}

	lastForgeL1TxsNum, err := s.historyDB.GetLastL1TxsNum()
	if err != nil && common.Unwrap(err) != sql.ErrNoRows {
		return common.Wrap(fmt.Errorf("historyDB.GetLastL1BatchBlockNum: %w", err))
	}
	if common.Unwrap(err) == sql.ErrNoRows || lastForgeL1TxsNum == nil {
		n := int64(-1)
		lastForgeL1TxsNum = &n
	}

	s.stats.UpdateSync(block, batch, &lastL1BatchBlockNum, lastForgeL1TxsNum)
	return nil
}

// TODO: Update consts variable above to SConsts
// RollupConstants returns the RollupConstants read from the smart contract
func (s *Synchronizer) RollupConstants() *common.RollupConstants {
	return &s.consts.Rollup
}

// TODO: Need to check and Update type initialised above in Synchronizer struct RollupVariables -> to SCVariables
// SCVars returns a copy of the Smart Contract Variables
func (s *Synchronizer) SCVars() *common.SCVariables {
	return &common.SCVariables{
		Rollup: *s.vars.Rollup.Copy(),
	}
}