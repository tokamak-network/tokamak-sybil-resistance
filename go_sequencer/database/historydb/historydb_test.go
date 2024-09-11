package historydb

import (
<<<<<<< HEAD
	"database/sql"
	"math/big"
=======
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
	"os"
	"testing"
	"time"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/database"
	"tokamak-sybil-resistance/test/til"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var historyDB *HistoryDB
var historyDBWithACC *HistoryDB

// Block0 represents Ethereum's genesis block,
// which is stored by default at HistoryDB
var Block0 common.Block = common.Block{
	Num: 0,
	Hash: ethCommon.Hash([32]byte{
		212, 229, 103, 64, 248, 118, 174, 248,
		192, 16, 184, 106, 64, 213, 245, 103,
		69, 161, 24, 208, 144, 106, 52, 230,
		154, 236, 140, 13, 177, 203, 143, 163,
	}), // 0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3
	Timestamp: time.Date(2015, time.July, 30, 3, 26, 13, 0, time.UTC), // 2015-07-30 03:26:13
}

// WipeDB redo all the migrations of the SQL DB (HistoryDB and L2DB),
// efectively recreating the original state
func WipeDB(db *sqlx.DB) {
	if err := database.MigrationsDown(db.DB, 0); err != nil {
		panic(err)
	}
	if err := database.MigrationsUp(db.DB); err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	// init DB
	db, err := database.InitTestSQLDB()
	if err != nil {
		panic(err)
	}
	historyDB = NewHistoryDB(db, db, nil)
	apiConnCon := database.NewAPIConnectionController(1, time.Second)
	historyDBWithACC = NewHistoryDB(db, db, apiConnCon)

	// Run tests
	result := m.Run()
	// Close DB
	if err := db.Close(); err != nil {
		log.Error("Error closing the history DB", err)
	}
	os.Exit(result)
}

func TestBlocks(t *testing.T) {
	var fromBlock, toBlock int64
	fromBlock = 0
	toBlock = 7
	// Reset DB
	WipeDB(historyDB.DB())
	// Generate blocks using til
	set1 := `
		Type: Blockchain
		// block 0 is stored as default in the DB
		// block 1 does not exist
		> block // blockNum=2
<<<<<<< HEAD
		> block // blockNum=3
=======
		> block // blockNum=3 
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
		> block // blockNum=4
		> block // blockNum=5
		> block // blockNum=6
	`
	tc := til.NewContext(uint16(0), 1)
	blocks, err := tc.GenerateBlocks(set1)
	require.NoError(t, err)
	// Save timestamp of a block with UTC and change it without UTC
	timestamp := time.Now().Add(time.Second * 13)
	blocks[fromBlock].Block.Timestamp = timestamp
	// Insert blocks into DB
	for i := 0; i < len(blocks); i++ {
		err := historyDB.AddBlock(&blocks[i].Block)
		assert.NoError(t, err)
	}
	// Add block 0 to the generated blocks
	blocks = append(
		[]common.BlockData{{Block: Block0}}, //nolint:gofmt
		blocks...,
	)
	// Get all blocks from DB
	fetchedBlocks, err := historyDB.getBlocks(fromBlock, toBlock)
	assert.Equal(t, len(blocks), len(fetchedBlocks))
	// Compare generated vs getted blocks
	assert.NoError(t, err)
	for i := range fetchedBlocks {
		assertEqualBlock(t, &blocks[i].Block, &fetchedBlocks[i])
	}
	// Compare saved timestamp vs getted
	nameZoneUTC, offsetUTC := timestamp.UTC().Zone()
	zoneFetchedBlock, offsetFetchedBlock := fetchedBlocks[fromBlock].Timestamp.Zone()
	assert.Equal(t, nameZoneUTC, zoneFetchedBlock)
	assert.Equal(t, offsetUTC, offsetFetchedBlock)
	// Get blocks from the DB one by one
	for i := int64(2); i < toBlock; i++ { // avoid block 0 for simplicity
		fetchedBlock, err := historyDB.GetBlock(i)
		assert.NoError(t, err)
		assertEqualBlock(t, &blocks[i-1].Block, fetchedBlock)
	}
	// Get last block
	lastBlock, err := historyDB.GetLastBlock()
	assert.NoError(t, err)
	assertEqualBlock(t, &blocks[len(blocks)-1].Block, lastBlock)
}

<<<<<<< HEAD
func TestBatches(t *testing.T) {
	// Reset DB
	WipeDB(historyDB.DB())
	// Generate batches using til (and blocks for foreign key)
	set := `
		Type: Blockchain
		
		CreateAccountDeposit A: 2000
		CreateAccountDeposit B: 1000
		> batchL1
		> batchL1
		CreateVouch A-B
		CreateVouch B-A
		> batch   // batchNum=2, L2 only batch, forges createVouches
		> block
		DeleteVouch A-B
		> batch   // batchNum=3, L2 only batch, forges deleteVouch
		DeleteVouch B-A
		> batch   // batchNum=4, L2 only batch, forges delteVouch
		> block
	`
	tc := til.NewContext(uint16(0), common.RollupConstMaxL1UserTx)
	tilCfgExtra := til.ConfigExtra{
		BootCoordAddr: ethCommon.HexToAddress("0xE39fEc6224708f0772D2A74fd3f9055A90E0A9f2"),
		CoordUser:     "A",
	}
	blocks, err := tc.GenerateBlocks(set)
	require.NoError(t, err)
	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	require.NoError(t, err)
	// Insert to DB
	batches := []common.Batch{}
	lastL1TxsNum := new(int64)
	lastL1BatchBlockNum := int64(0)
	for _, block := range blocks {
		// Insert block
		assert.NoError(t, historyDB.AddBlock(&block.Block))
		// Combine all generated batches into single array
		for _, batch := range block.Rollup.Batches {
			batch.Batch.GasPrice = big.NewInt(0)
			batches = append(batches, batch.Batch)
			forgeTxsNum := batch.Batch.ForgeL1TxsNum
			if forgeTxsNum != nil && (lastL1TxsNum == nil || *lastL1TxsNum < *forgeTxsNum) {
				*lastL1TxsNum = *forgeTxsNum
				lastL1BatchBlockNum = batch.Batch.EthBlockNum
			}
		}
	}
	// Insert batches
	assert.NoError(t, historyDB.AddBatches(batches))

	// Get batches from the DB
	fetchedBatches, err := historyDB.GetBatches(0, common.BatchNum(len(batches)+1))
	assert.NoError(t, err)
	assert.Equal(t, len(batches), len(fetchedBatches))
	for i, fetchedBatch := range fetchedBatches {
		assert.Equal(t, batches[i], fetchedBatch)
	}
	// Test GetLastBatchNum
	fetchedLastBatchNum, err := historyDB.GetLastBatchNum()
	assert.NoError(t, err)
	assert.Equal(t, batches[len(batches)-1].BatchNum, fetchedLastBatchNum)
	// Test GetLastBatch
	fetchedLastBatch, err := historyDB.GetLastBatch()
	assert.NoError(t, err)
	assert.Equal(t, &batches[len(batches)-1], fetchedLastBatch)
	// Test GetLastL1TxsNum
	fetchedLastL1TxsNum, err := historyDB.GetLastL1TxsNum()
	assert.NoError(t, err)
	assert.Equal(t, lastL1TxsNum, fetchedLastL1TxsNum)
	// Test GetLastL1BatchBlockNum
	fetchedLastL1BatchBlockNum, err := historyDB.GetLastL1BatchBlockNum()
	assert.NoError(t, err)
	assert.Equal(t, lastL1BatchBlockNum, fetchedLastL1BatchBlockNum)
	// Test GetBatch
	fetchedBatch, err := historyDB.GetBatch(1)
	require.NoError(t, err)
	assert.Equal(t, &batches[0], fetchedBatch)
	_, err = historyDB.GetBatch(common.BatchNum(len(batches) + 1))
	assert.Equal(t, sql.ErrNoRows, common.Unwrap(err))
}

=======
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
func assertEqualBlock(t *testing.T, expected *common.Block, actual *common.Block) {
	assert.Equal(t, expected.Num, actual.Num)
	assert.Equal(t, expected.Hash, actual.Hash)
	assert.Equal(t, expected.Timestamp.Unix(), actual.Timestamp.Unix())
}
