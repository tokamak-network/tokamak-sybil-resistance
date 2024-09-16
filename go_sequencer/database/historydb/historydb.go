package historydb

import (
	"math"
	"math/big"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/database"
	"tokamak-sybil-resistance/log"

	"github.com/jmoiron/sqlx"
	"github.com/russross/meddler"
)

// HistoryDB persist the historic of the rollup
type HistoryDB struct {
	dbRead     *sqlx.DB
	dbWrite    *sqlx.DB
	apiConnCon *database.APIConnectionController
}

// NewHistoryDB initialize the DB
func NewHistoryDB(dbRead, dbWrite *sqlx.DB, apiConnCon *database.APIConnectionController) *HistoryDB {
	return &HistoryDB{
		dbRead:     dbRead,
		dbWrite:    dbWrite,
		apiConnCon: apiConnCon,
	}
}

// DB returns a pointer to the L2DB.db. This method should be used only for
// internal testing purposes.
func (hdb *HistoryDB) DB() *sqlx.DB {
	return hdb.dbWrite
}

// AddBlock insert a block into the DB
func (hdb *HistoryDB) AddBlock(block *common.Block) error { return hdb.addBlock(hdb.dbWrite, block) }
func (hdb *HistoryDB) addBlock(d meddler.DB, block *common.Block) error {
	return common.Wrap(meddler.Insert(d, "block", block))
}

// GetBlock retrieve a block from the DB, given a block number
func (hdb *HistoryDB) GetBlock(blockNum int64) (*common.Block, error) {
	block := &common.Block{}
	err := meddler.QueryRow(
		hdb.dbRead, block,
		"SELECT * FROM block WHERE eth_block_num = $1;", blockNum,
	)
	return block, common.Wrap(err)
}

// GetLastBlock retrieve the block with the highest block number from the DB
func (hdb *HistoryDB) GetLastBlock() (*common.Block, error) {
	block := &common.Block{}
	err := meddler.QueryRow(
		hdb.dbRead, block, "SELECT * FROM block ORDER BY eth_block_num DESC LIMIT 1;",
	)
	return block, common.Wrap(err)
}

// getBlocks retrieve blocks from the DB, given a range of block numbers defined by from and to
func (hdb *HistoryDB) getBlocks(from, to int64) ([]common.Block, error) {
	var blocks []*common.Block
	err := meddler.QueryAll(
		hdb.dbRead, &blocks,
		"SELECT * FROM block WHERE $1 <= eth_block_num AND eth_block_num < $2 ORDER BY eth_block_num;",
		from, to,
	)
	return database.SlicePtrsToSlice(blocks).([]common.Block), common.Wrap(err)
}

// GetSCVars returns the rollup, auction and wdelayer smart contracts variables at their last update.
func (hdb *HistoryDB) GetSCVars() (*common.RollupVariables, error) {
	var rollup common.RollupVariables
	if err := meddler.QueryRow(hdb.dbRead, &rollup,
		"SELECT * FROM rollup_vars ORDER BY eth_block_num DESC LIMIT 1;"); err != nil {
		return nil, common.Wrap(err)
	}
	return &rollup, nil
}

// SetInitialSCVars sets the initial state of rollup, auction, wdelayer smart
// contract variables.  This initial state is stored linked to block 0, which
// always exist in the DB and is used to store initialization data that always
// exist in the smart contracts.
func (hdb *HistoryDB) SetInitialSCVars(rollup *common.RollupVariables) error {
	txn, err := hdb.dbWrite.Beginx()
	if err != nil {
		return common.Wrap(err)
	}
	defer func() {
		if err != nil {
			database.Rollback(txn)
		}
	}()
	// Force EthBlockNum to be 0 because it's the block used to link data
	// that belongs to the creation of the smart contracts
	rollup.EthBlockNum = 0
	if err := hdb.setRollupVars(txn, rollup); err != nil {
		return common.Wrap(err)
	}
	return common.Wrap(txn.Commit())
}

func (hdb *HistoryDB) setRollupVars(d meddler.DB, rollup *common.RollupVariables) error {
	return common.Wrap(meddler.Insert(d, "rollup_vars", rollup))
}

// AddCoordinators insert Coordinators into the DB
func (hdb *HistoryDB) AddCoordinators(coordinators []common.Coordinator) error {
	return common.Wrap(hdb.addCoordinators(hdb.dbWrite, coordinators))
}
func (hdb *HistoryDB) addCoordinators(d meddler.DB, coordinators []common.Coordinator) error {
	if len(coordinators) == 0 {
		return nil
	}
	return common.Wrap(database.BulkInsert(
		d,
		"INSERT INTO coordinator (bidder_addr, forger_addr, eth_block_num, url) VALUES %s;",
		coordinators,
	))
}

// GetLastBatch returns the last forged batch
func (hdb *HistoryDB) GetLastBatch() (*common.Batch, error) {
	var batch common.Batch
	err := meddler.QueryRow(
		hdb.dbRead, &batch, `SELECT batch.batch_num, batch.eth_block_num, batch.forger_addr,
		batch.fees_collected, batch.fee_idxs_coordinator, batch.state_root,
		batch.num_accounts, batch.last_idx, batch.exit_root, batch.forge_l1_txs_num,
		batch.slot_num, batch.total_fees_usd, batch.gas_price, batch.gas_used, batch.ether_price_usd
		FROM batch ORDER BY batch_num DESC LIMIT 1;`,
	)
	return &batch, common.Wrap(err)
}

// GetLastL1BatchBlockNum returns the blockNum of the latest forged l1Batch
func (hdb *HistoryDB) GetLastL1BatchBlockNum() (int64, error) {
	row := hdb.dbRead.QueryRow(`SELECT eth_block_num FROM batch
		WHERE forge_l1_txs_num IS NOT NULL
		ORDER BY batch_num DESC LIMIT 1;`)
	var blockNum int64
	return blockNum, common.Wrap(row.Scan(&blockNum))
}

// GetLastL1TxsNum returns the greatest ForgeL1TxsNum in the DB from forged
// batches.  If there's no batch in the DB (nil, nil) is returned.
func (hdb *HistoryDB) GetLastL1TxsNum() (*int64, error) {
	row := hdb.dbRead.QueryRow("SELECT MAX(forge_l1_txs_num) FROM batch;")
	lastL1TxsNum := new(int64)
	return lastL1TxsNum, common.Wrap(row.Scan(&lastL1TxsNum))
}

// AddBatch insert a Batch into the DB
func (hdb *HistoryDB) AddBatch(batch *common.Batch) error { return hdb.addBatch(hdb.dbWrite, batch) }
func (hdb *HistoryDB) addBatch(d meddler.DB, batch *common.Batch) error {
	// Calculate total collected fees in USD
	// Get IDs of collected tokens for fees
	tokenIDs := []common.TokenID{}
	for id := range batch.CollectedFees {
		tokenIDs = append(tokenIDs, id)
	}
	// Get USD value of the tokens
	type tokenPrice struct {
		ID       common.TokenID `meddler:"token_id"`
		USD      *float64       `meddler:"usd"`
		Decimals int            `meddler:"decimals"`
	}
	var tokenPrices []*tokenPrice
	if len(tokenIDs) > 0 {
		query, args, err := sqlx.In(
			"SELECT token_id, usd, decimals FROM token WHERE token_id IN (?);",
			tokenIDs,
		)
		if err != nil {
			return common.Wrap(err)
		}
		query = hdb.dbWrite.Rebind(query)
		if err := meddler.QueryAll(
			hdb.dbWrite, &tokenPrices, query, args...,
		); err != nil {
			return common.Wrap(err)
		}
	}
	// Calculate total collected
	var total float64
	for _, tokenPrice := range tokenPrices {
		if tokenPrice.USD == nil {
			continue
		}
		f := new(big.Float).SetInt(batch.CollectedFees[tokenPrice.ID])
		amount, _ := f.Float64()
		total += *tokenPrice.USD * (amount / math.Pow(10, float64(tokenPrice.Decimals))) //nolint decimals have to be ^10
	}
	batch.TotalFeesUSD = &total
	// Check current ether price and insert it into batch table
	var ether TokenWithUSD
	err := meddler.QueryRow(
		hdb.dbRead, &ether,
		"SELECT * FROM token WHERE symbol = 'ETH';",
	)
	if err != nil {
		log.Warn("error getting ether price from db: ", err)
		batch.EtherPriceUSD = 0
	} else if ether.USD == nil {
		batch.EtherPriceUSD = 0
	} else {
		batch.EtherPriceUSD = *ether.USD
	}
	if batch.GasPrice == nil {
		batch.GasPrice = big.NewInt(0)
	}
	// Insert to DB
	return common.Wrap(meddler.Insert(d, "batch", batch))
}

// AddBatches insert Bids into the DB
func (hdb *HistoryDB) AddBatches(batches []common.Batch) error {
	return common.Wrap(hdb.addBatches(hdb.dbWrite, batches))
}
func (hdb *HistoryDB) addBatches(d meddler.DB, batches []common.Batch) error {
	for i := 0; i < len(batches); i++ {
		if err := hdb.addBatch(d, &batches[i]); err != nil {
			return common.Wrap(err)
		}
	}
	return nil
}

// GetBatches retrieve batches from the DB, given a range of batch numbers defined by from and to
func (hdb *HistoryDB) GetBatches(from, to common.BatchNum) ([]common.Batch, error) {
	var batches []*common.Batch
	err := meddler.QueryAll(
		hdb.dbRead, &batches,
		`SELECT batch_num, eth_block_num, forger_addr, fees_collected, fee_idxs_coordinator, 
		state_root, num_accounts, last_idx, exit_root, forge_l1_txs_num, slot_num, total_fees_usd, gas_price, gas_used, ether_price_usd 
		FROM batch WHERE $1 <= batch_num AND batch_num < $2 ORDER BY batch_num;`,
		from, to,
	)
	return database.SlicePtrsToSlice(batches).([]common.Batch), common.Wrap(err)
}

// GetLastBatchNum returns the BatchNum of the latest forged batch
func (hdb *HistoryDB) GetLastBatchNum() (common.BatchNum, error) {
	row := hdb.dbRead.QueryRow("SELECT batch_num FROM batch ORDER BY batch_num DESC LIMIT 1;")
	var batchNum common.BatchNum
	return batchNum, common.Wrap(row.Scan(&batchNum))
}

// GetBatch returns the batch with the given batchNum
func (hdb *HistoryDB) GetBatch(batchNum common.BatchNum) (*common.Batch, error) {
	var batch common.Batch
	err := meddler.QueryRow(
		hdb.dbRead, &batch, `SELECT batch.batch_num, batch.eth_block_num, batch.forger_addr,
		batch.fees_collected, batch.fee_idxs_coordinator, batch.state_root,
		batch.num_accounts, batch.last_idx, batch.exit_root, batch.forge_l1_txs_num,
		batch.slot_num, batch.total_fees_usd, batch.gas_price, batch.gas_used, batch.ether_price_usd
		FROM batch WHERE batch_num = $1;`,
		batchNum,
	)
	return &batch, common.Wrap(err)
}
