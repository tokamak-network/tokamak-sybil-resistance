package historydb

import (
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/database"

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

// GetLastBlock retrieve the block with the highest block number from the DB
func (hdb *HistoryDB) GetLastBlock() (*common.Block, error) {
	block := &common.Block{}
	err := meddler.QueryRow(
		hdb.dbRead, block, "SELECT * FROM block ORDER BY eth_block_num DESC LIMIT 1;",
	)
	return block, common.Wrap(err)
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

// GetBestBidCoordinator returns the forger address of the highest bidder in a slot by slotNum
func (hdb *HistoryDB) GetBidCoordinator(slotNum int64) (*common.BidCoordinator, error) {
	bidCoord := &common.BidCoordinator{}
	err := meddler.QueryRow(
		hdb.dbRead, bidCoord,
		`SELECT (
			SELECT default_slot_set_bid
			FROM auction_vars
			WHERE default_slot_set_bid_slot_num <= $1
			ORDER BY eth_block_num DESC LIMIT 1
		),
		bid.slot_num, bid.bid_value, bid.bidder_addr,
		coordinator.forger_addr, coordinator.url
		FROM bid
		INNER JOIN (
			SELECT bidder_addr, MAX(item_id) AS item_id FROM coordinator
			GROUP BY bidder_addr
		) c ON bid.bidder_addr = c.bidder_addr 
		INNER JOIN coordinator ON c.item_id = coordinator.item_id
		WHERE bid.slot_num = $1 ORDER BY bid.item_id DESC LIMIT 1;`,
		slotNum)

	return bidCoord, common.Wrap(err)
}
