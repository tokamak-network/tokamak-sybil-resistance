package historydb

import (
	"database/sql"
	"math/big"
	"os"
	"strconv"
	"testing"
	"time"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/database"
	"tokamak-sybil-resistance/log"
	"tokamak-sybil-resistance/test"
	"tokamak-sybil-resistance/test/til"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var historyDB *HistoryDB
var historyDBWithACC *HistoryDB
var debug bool

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

func TestMain(m *testing.M) {
	// init DB
	db, err := database.InitTestSQLDB()
	if err != nil {
		panic(err)
	}
	historyDB = NewHistoryDB(db, db, nil)
	apiConnCon := database.NewAPIConnectionController(1, time.Second)
	historyDBWithACC = NewHistoryDB(db, db, apiConnCon)
	debug, err = strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		debug = false
	}

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
	test.WipeDB(historyDB.DB())
	// Generate blocks using til
	set1 := `
		Type: Blockchain
		// block 0 is stored as default in the DB
		// block 1 does not exist
		> block // blockNum=2
		> block // blockNum=3
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

func TestBatches(t *testing.T) {
	// Reset DB
	test.WipeDB(historyDB.DB())
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
			if forgeTxsNum != nil && *lastL1TxsNum < *forgeTxsNum {
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

func TestAccounts(t *testing.T) {
	const fromBlock int64 = 1
	const toBlock int64 = 5

	// Prepare blocks in the DB
	blocks := setTestBlocks(fromBlock, toBlock)

	// Generate fake batches
	const nBatches = 10
	batches := test.GenBatches(nBatches, blocks)
	err := historyDB.AddBatches(batches)
	assert.NoError(t, err)

	// Generate fake accounts
	const nAccounts = 3
	accs := test.GenAccounts(nAccounts, 0, nil, nil, batches)
	err = historyDB.AddAccounts(accs)
	assert.NoError(t, err)

	// Fetch accounts
	fetchedAccs, err := historyDB.GetAllAccounts()
	assert.NoError(t, err)

	// Compare fetched accounts vs generated accounts
	for i, acc := range fetchedAccs {
		accs[i].Balance = nil
		assert.Equal(t, accs[i], acc)
	}

	// Test AccountBalances
	accUpdates := make([]common.AccountUpdate, len(accs))
	for i, acc := range accs {
		accUpdates[i] = common.AccountUpdate{
			EthBlockNum: batches[acc.BatchNum-1].EthBlockNum,
			BatchNum:    acc.BatchNum,
			Idx:         acc.Idx,
			Nonce:       common.Nonce(i),
			Balance:     big.NewInt(int64(i)),
		}
	}
	err = historyDB.AddAccountUpdates(accUpdates)
	require.NoError(t, err)
	fetchedAccBalances, err := historyDB.GetAllAccountUpdates()
	require.NoError(t, err)
	assert.Equal(t, accUpdates, fetchedAccBalances)
}

func TestTxs(t *testing.T) {
	// Reset DB
	test.WipeDB(historyDB.DB())

	set := `
	Type: Blockchain
	
	CreateAccountDeposit A: 1
	CreateAccountDeposit B: 2
	> batchL1
	> batchL1
	> block  // block 1
	
	Deposit B: 10
	Exit A: 10
	> batch
	> block  // block 2
	
	ForceExit A: 5
	> batchL1
	> batchL1
	> block	// block 3

	CreateAccountDeposit D: 10
	> batchL1
	> block  // block 4

	CreateAccountDeposit E: 10
	> batchL1
	> batchL1
	> block	// block 5

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

	// Sanity check
	require.Equal(t, 5, len(blocks)) // total number of blocks is 5

	assert.Equal(t, 2, len(blocks[0].Rollup.L1UserTxs))                  // block 1 contains 2 L1UserTxs
	assert.Equal(t, 2, len(blocks[0].Rollup.Batches))                    // block 1 contains 2 Batches
	assert.Equal(t, 2, len(blocks[0].Rollup.Batches[1].CreatedAccounts)) // block 1, batch 2 contains 2 CreatedAccounts

	require.Equal(t, 1, len(blocks[1].Rollup.L1UserTxs))        // block 2 contains 1 L1UserTxs
	require.Equal(t, 1, len(blocks[1].Rollup.Batches))          // block 2 contains 1 Batch
	require.Equal(t, 1, len(blocks[1].Rollup.Batches[0].L2Txs)) // block 2, batch 1 contains 1 L2Tx

	require.Equal(t, 2, len(blocks[2].Rollup.Batches))          // block 3 contains 2 Batches
	require.Equal(t, 1, len(blocks[2].Rollup.L1UserTxs))        // block 3 contains 1 L1UserTxs
	require.Equal(t, 0, len(blocks[2].Rollup.Batches[1].L2Txs)) // block 2, batch 2 contains 0 L2Tx

	require.Equal(t, 1, len(blocks[3].Rollup.Batches))          // block 4 contains 1 Batch
	require.Equal(t, 1, len(blocks[3].Rollup.L1UserTxs))        // block 4 contains 1 L1UserTxs
	require.Equal(t, 0, len(blocks[2].Rollup.Batches[0].L2Txs)) // block 5, batch 2 contains 0 L2Tx

	require.Equal(t, 2, len(blocks[4].Rollup.Batches))          // block 5 contains 2 Batches
	require.Equal(t, 1, len(blocks[4].Rollup.L1UserTxs))        // block 5 contains 1 L1UserTxs
	require.Equal(t, 0, len(blocks[4].Rollup.Batches[0].L2Txs)) // block 5, batch 1 contains 0 L2Tx

	var null *common.BatchNum = nil
	var txID common.TxID

	// Insert blocks into DB
	for i := range blocks {
		if i == len(blocks)-1 {
			blocks[i].Block.Timestamp = time.Now()
			dbL1Txs, err := historyDB.GetAllL1UserTxs()
			assert.NoError(t, err)
			// Check batch_num is nil before forging
			assert.Equal(t, null, dbL1Txs[len(dbL1Txs)-1].BatchNum)
		}
		err = historyDB.AddBlockSCData(&blocks[i])
		assert.NoError(t, err)
	}

	// Check blocks
	dbBlocks, err := historyDB.GetAllBlocks()
	assert.NoError(t, err)
	assert.Equal(t, len(blocks)+1, len(dbBlocks))

	// Check batches
	batches, err := historyDB.GetAllBatches()
	assert.NoError(t, err)
	assert.Equal(t, 8, len(batches))

	// Check L1 Transactions
	dbL1Txs, err := historyDB.GetAllL1UserTxs()
	assert.NoError(t, err)
	assert.Equal(t, 6, len(dbL1Txs))

	// Tx Type
	// assert.Equal(t, common.TxTypeCreateAccountDeposit, dbL1Txs[0].Type)
	// assert.Equal(t, common.TxTypeCreateAccountDeposit, dbL1Txs[1].Type)
	// assert.Equal(t, common.TxTypeCreateAccountDepositTransfer, dbL1Txs[2].Type)
	// assert.Equal(t, common.TxTypeDeposit, dbL1Txs[3].Type)
	// assert.Equal(t, common.TxTypeDeposit, dbL1Txs[4].Type)
	// assert.Equal(t, common.TxTypeDepositTransfer, dbL1Txs[5].Type)
	// assert.Equal(t, common.TxTypeForceTransfer, dbL1Txs[6].Type)
	// assert.Equal(t, common.TxTypeForceExit, dbL1Txs[7].Type)
	// assert.Equal(t, common.TxTypeCreateAccountDeposit, dbL1Txs[8].Type)
	// assert.Equal(t, common.TxTypeCreateAccountDeposit, dbL1Txs[9].Type)

	// Tx Type
	assert.Equal(t, common.TxTypeCreateAccountDeposit, dbL1Txs[0].Type)
	assert.Equal(t, common.TxTypeCreateAccountDeposit, dbL1Txs[1].Type)
	assert.Equal(t, common.TxTypeDeposit, dbL1Txs[2].Type)
	assert.Equal(t, common.TxTypeForceExit, dbL1Txs[3].Type)
	assert.Equal(t, common.TxTypeCreateAccountDeposit, dbL1Txs[4].Type)
	assert.Equal(t, common.TxTypeCreateAccountDeposit, dbL1Txs[5].Type)

	// Tx ID
	assert.Equal(t, "0x00e979da4b80d60a17ce56fa19278c6f3a7e1b43359fb8a8ea46d0264de7d653ab", dbL1Txs[0].TxID.String())
	assert.Equal(t, "0x00af9bf96eb60f2d618519402a2f6b07057a034fa2baefd379fe8e1c969f1c5cf4", dbL1Txs[1].TxID.String())
	assert.Equal(t, "0x00a256ee191905243320ea830840fd666a73c7b4e6f89ce4bd47ddf998dfee627a", dbL1Txs[2].TxID.String())
	assert.Equal(t, "0x007f5383186254f364bfe82ef3ccff4b1bf532bfb1d424fe3858e492b61b0262fe", dbL1Txs[3].TxID.String())
	assert.Equal(t, "0x00930696d03ae0a1e6150b6ccb88043cb539a4e06a7f8baf213029ce9a0600197e", dbL1Txs[4].TxID.String())
	assert.Equal(t, "0x00c33f316240f8d33a973db2d0e901e4ac1c96de30b185fcc6b63dac4d0e147bd4", dbL1Txs[5].TxID.String())

	// Tx From IDx
	assert.Equal(t, common.AccountIdx(0), dbL1Txs[0].FromIdx)
	assert.Equal(t, common.AccountIdx(0), dbL1Txs[1].FromIdx)
	assert.NotEqual(t, common.AccountIdx(0), dbL1Txs[2].FromIdx)
	assert.NotEqual(t, common.AccountIdx(0), dbL1Txs[3].FromIdx)
	assert.Equal(t, common.AccountIdx(0), dbL1Txs[4].FromIdx)
	assert.Equal(t, common.AccountIdx(0), dbL1Txs[5].FromIdx)

	// Batch Number
	var bn = common.BatchNum(2)

	assert.Equal(t, &bn, dbL1Txs[0].BatchNum)
	assert.Equal(t, &bn, dbL1Txs[1].BatchNum)

	bn = common.BatchNum(5)
	assert.Equal(t, bn, *dbL1Txs[3].BatchNum)

	bn = common.BatchNum(7)
	assert.Equal(t, &bn, dbL1Txs[3].BatchNum)
	assert.Equal(t, &bn, dbL1Txs[4].BatchNum)
	assert.Equal(t, &bn, dbL1Txs[5].BatchNum)
	return

	bn = common.BatchNum(8)
	assert.Equal(t, &bn, dbL1Txs[6].BatchNum)
	assert.Equal(t, &bn, dbL1Txs[7].BatchNum)

	bn = common.BatchNum(10)
	assert.Equal(t, &bn, dbL1Txs[8].BatchNum)

	bn = common.BatchNum(11)
	assert.Equal(t, &bn, dbL1Txs[9].BatchNum)

	// eth_block_num
	assert.Equal(t, int64(2), dbL1Txs[0].EthBlockNum)
	assert.Equal(t, int64(2), dbL1Txs[1].EthBlockNum)
	assert.Equal(t, int64(3), dbL1Txs[2].EthBlockNum)
	assert.Equal(t, int64(4), dbL1Txs[3].EthBlockNum)
	assert.Equal(t, int64(4), dbL1Txs[4].EthBlockNum)
	assert.Equal(t, int64(5), dbL1Txs[5].EthBlockNum)
	assert.Equal(t, int64(6), dbL1Txs[6].EthBlockNum)
	assert.Equal(t, int64(6), dbL1Txs[7].EthBlockNum)
	assert.Equal(t, int64(7), dbL1Txs[8].EthBlockNum)
	assert.Equal(t, int64(8), dbL1Txs[9].EthBlockNum)

	// User Origin
	assert.Equal(t, true, dbL1Txs[0].UserOrigin)
	assert.Equal(t, true, dbL1Txs[1].UserOrigin)
	assert.Equal(t, true, dbL1Txs[2].UserOrigin)
	assert.Equal(t, true, dbL1Txs[3].UserOrigin)
	assert.Equal(t, true, dbL1Txs[4].UserOrigin)
	assert.Equal(t, true, dbL1Txs[5].UserOrigin)
	assert.Equal(t, true, dbL1Txs[6].UserOrigin)
	assert.Equal(t, true, dbL1Txs[7].UserOrigin)
	assert.Equal(t, true, dbL1Txs[8].UserOrigin)
	assert.Equal(t, true, dbL1Txs[9].UserOrigin)

	// Deposit Amount
	assert.Equal(t, big.NewInt(10), dbL1Txs[0].DepositAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[1].DepositAmount)
	assert.Equal(t, big.NewInt(20), dbL1Txs[2].DepositAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[3].DepositAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[4].DepositAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[5].DepositAmount)
	assert.Equal(t, big.NewInt(0), dbL1Txs[6].DepositAmount)
	assert.Equal(t, big.NewInt(0), dbL1Txs[7].DepositAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[8].DepositAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[9].DepositAmount)

	// Check saved txID's batch_num is not nil
	assert.Equal(t, txID, dbL1Txs[len(dbL1Txs)-2].TxID)
	assert.NotEqual(t, null, dbL1Txs[len(dbL1Txs)-2].BatchNum)

	// Check Coordinator TXs
	coordTxs, err := historyDB.GetAllL1CoordinatorTxs()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(coordTxs))
	assert.Equal(t, common.TxTypeCreateAccountDeposit, coordTxs[0].Type)
	assert.Equal(t, false, coordTxs[0].UserOrigin)

	// Check L2 TXs
	dbL2Txs, err := historyDB.GetAllL2Txs()
	assert.NoError(t, err)
	assert.Equal(t, 4, len(dbL2Txs))

	// Tx Type
	assert.Equal(t, common.TxTypeTransfer, dbL2Txs[0].Type)
	assert.Equal(t, common.TxTypeTransfer, dbL2Txs[1].Type)
	assert.Equal(t, common.TxTypeTransfer, dbL2Txs[2].Type)
	assert.Equal(t, common.TxTypeExit, dbL2Txs[3].Type)

	// Tx ID
	assert.Equal(t, "0x024e555248100b69a8aabf6d31719b9fe8a60dcc6c3407904a93c8d2d9ade18ee5", dbL2Txs[0].TxID.String())
	assert.Equal(t, "0x021ae87ca34d50ff35d98dfc0d7c95f2bf2e4ffeebb82ea71f43a8b0dfa5d36d89", dbL2Txs[1].TxID.String())
	assert.Equal(t, "0x024abce7f3f2382dc520ed557593f11dea1ee197e55b60402e664facc27aa19774", dbL2Txs[2].TxID.String())
	assert.Equal(t, "0x02f921ad9e7a6e59606570fe12a7dde0e36014197de0363b9b45e5097d6f2b1dd0", dbL2Txs[3].TxID.String())

	// Tx From and To IDx
	assert.Equal(t, dbL2Txs[0].ToIdx, dbL2Txs[2].FromIdx)
	assert.Equal(t, dbL2Txs[1].ToIdx, dbL2Txs[0].FromIdx)
	assert.Equal(t, dbL2Txs[2].ToIdx, dbL2Txs[1].FromIdx)

	// Batch Number
	assert.Equal(t, common.BatchNum(5), dbL2Txs[0].BatchNum)
	assert.Equal(t, common.BatchNum(5), dbL2Txs[1].BatchNum)
	assert.Equal(t, common.BatchNum(5), dbL2Txs[2].BatchNum)
	assert.Equal(t, common.BatchNum(5), dbL2Txs[3].BatchNum)

	// eth_block_num
	assert.Equal(t, int64(4), dbL2Txs[0].EthBlockNum)
	assert.Equal(t, int64(4), dbL2Txs[1].EthBlockNum)
	assert.Equal(t, int64(4), dbL2Txs[2].EthBlockNum)

	// Amount
	assert.Equal(t, big.NewInt(10), dbL2Txs[0].Amount)
	assert.Equal(t, big.NewInt(10), dbL2Txs[1].Amount)
	assert.Equal(t, big.NewInt(10), dbL2Txs[2].Amount)
	assert.Equal(t, big.NewInt(10), dbL2Txs[3].Amount)
}

func assertEqualBlock(t *testing.T, expected *common.Block, actual *common.Block) {
	assert.Equal(t, expected.Num, actual.Num)
	assert.Equal(t, expected.Hash, actual.Hash)
	assert.Equal(t, expected.Timestamp.Unix(), actual.Timestamp.Unix())
}

// setTestBlocks WARNING: this will delete the blocks and recreate them
func setTestBlocks(from, to int64) []common.Block {
	test.WipeDB(historyDB.DB())
	blocks := test.GenBlocks(from, to)
	if err := historyDB.AddBlocks(blocks); err != nil {
		panic(err)
	}
	return blocks
}
