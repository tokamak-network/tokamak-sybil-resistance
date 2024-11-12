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
	"github.com/hermeznetwork/tracerr"
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
	// Reset DB
	test.WipeDB(historyDB.DB())

	var fromBlock, toBlock int64
	fromBlock = 0
	toBlock = 7
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
	// Reset DB
	test.WipeDB(historyDB.DB())

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
	
	CreateAccountDeposit A: 1 	// L1Tx 1
	CreateAccountDeposit B: 2	// L1Tx 2
	> batchL1					// batch 1
	> batchL1					// batch 2

	CreateVouch A-B		// L2Tx 1
	CreateVouch B-A		// L2Tx 2
	> batch				// batch 3
	> block  			// block 1
	
	Deposit B: 10	// L1Tx 3
	Exit A: 10		// L2Tx 3
	> batch			// batch 4
	> block  		// block 2
	
	ForceExit A: 5		// L1Tx 4
	> batchL1			// batch 5
	> batchL1			// batch 6
	> block				// block 3

	CreateAccountDeposit C: 10		// L1Tx 5
	> batchL1						// batch 7
	> block  						// block 4
	
	CreateAccountDeposit D: 10		// L1Tx 6
	> batchL1						// batch 8
	> batchL1						// batch 9
	> block							// block 5
	
	CreateVouch A-C		// L2Tx 4
	CreateVouch C-D		// L2Tx 5
	DeleteVouch B-A		// L2Tx 6
	> batch				// batch 10
	> block 			// block 6

	// CreateAccountDeposit E: 10		// L1Tx 7
	// > batchL1						// batch 11
	// > batchL1						// batch 12
	// > block							// block 7
`

	tc := til.NewContext(uint16(0), common.RollupConstMaxL1UserTx)
	tilCfgExtra := til.ConfigExtra{
		BootCoordAddr: ethCommon.HexToAddress("0xE39fEc6224708f0772D2A74fd3f9055A90E0A9f2"),
		CoordUser:     "A",
	}
	blocks, err := tc.GenerateBlocks(set)
	assert.NoError(t, err)

	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	assert.NoError(t, err)

	// Sanity check
	assert.Equal(t, 6, len(blocks)) // total number of blocks is 6

	assert.Equal(t, 2, len(blocks[0].Rollup.L1UserTxs))                  // block 1 contains 2 L1UserTxs
	assert.Equal(t, 3, len(blocks[0].Rollup.Batches))                    // block 1 contains 2 Batches
	assert.Equal(t, 2, len(blocks[0].Rollup.Batches[1].CreatedAccounts)) // block 1, batch 2 contains 2 CreatedAccounts
	assert.Equal(t, 2, len(blocks[0].Rollup.Batches[2].L2Txs))           // block 1, batch 1 contains 2 L2Txs

	assert.Equal(t, 1, len(blocks[1].Rollup.L1UserTxs))        // block 2 contains 1 L1UserTxs
	assert.Equal(t, 1, len(blocks[1].Rollup.Batches))          // block 2 contains 1 Batch
	assert.Equal(t, 1, len(blocks[1].Rollup.Batches[0].L2Txs)) // block 2, batch 1 contains 1 L2Tx

	assert.Equal(t, 2, len(blocks[2].Rollup.Batches))          // block 3 contains 2 Batches
	assert.Equal(t, 1, len(blocks[2].Rollup.L1UserTxs))        // block 3 contains 1 L1UserTxs
	assert.Equal(t, 0, len(blocks[2].Rollup.Batches[1].L2Txs)) // block 2, batch 2 contains 0 L2Tx

	assert.Equal(t, 1, len(blocks[3].Rollup.Batches))          // block 4 contains 1 Batch
	assert.Equal(t, 1, len(blocks[3].Rollup.L1UserTxs))        // block 4 contains 1 L1UserTxs
	assert.Equal(t, 0, len(blocks[3].Rollup.Batches[0].L2Txs)) // block 4, batch 2 contains 0 L2Tx

	assert.Equal(t, 2, len(blocks[4].Rollup.Batches))          // block 5 contains 2 Batches
	assert.Equal(t, 1, len(blocks[4].Rollup.L1UserTxs))        // block 5 contains 1 L1UserTxs
	assert.Equal(t, 0, len(blocks[4].Rollup.Batches[0].L2Txs)) // block 5, batch 1 contains 0 L2Tx

	assert.Equal(t, 1, len(blocks[5].Rollup.Batches))          // block 6 contains 1 Batch
	assert.Equal(t, 0, len(blocks[5].Rollup.L1UserTxs))        // block 6 contains 0 L1UserTxs
	assert.Equal(t, 3, len(blocks[5].Rollup.Batches[0].L2Txs)) // block 6, batch 1 contains 3 L2Tx

	// assert.Equal(t, 2, len(blocks[6].Rollup.Batches))          // block 7 contains 1 Batch
	// assert.Equal(t, 1, len(blocks[6].Rollup.L1UserTxs))        // block 7 contains 0 L1UserTxs
	// assert.Equal(t, 0, len(blocks[6].Rollup.Batches[0].L2Txs)) // block 7, batch 1 contains 3 L2Tx

	var null *common.BatchNum = nil

	// Insert blocks into DB
	for i := range blocks {
		// if i == len(blocks)-1 {
		// 	blocks[i].Block.Timestamp = time.Now()
		// 	dbL1Txs, err := historyDB.GetAllL1UserTxs()
		// 	assert.NoError(t, err)
		// 	// Check batch_num is nil before forging
		// 	assert.Equal(t, null, dbL1Txs[len(dbL1Txs)-1].BatchNum)
		// }
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
	assert.Equal(t, 10, len(batches))

	// Check L1 Transactions
	dbL1Txs, err := historyDB.GetAllL1UserTxs()
	assert.NoError(t, err)
	assert.Equal(t, 6, len(dbL1Txs))

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

	bn = common.BatchNum(6)
	assert.Equal(t, bn, *dbL1Txs[2].BatchNum)
	assert.Equal(t, bn, *dbL1Txs[3].BatchNum)

	bn = common.BatchNum(8)
	assert.Equal(t, &bn, dbL1Txs[4].BatchNum)

	bn = common.BatchNum(9)
	assert.Equal(t, &bn, dbL1Txs[5].BatchNum)

	// eth_block_num
	assert.Equal(t, int64(2), dbL1Txs[0].EthBlockNum)
	assert.Equal(t, int64(2), dbL1Txs[1].EthBlockNum)
	assert.Equal(t, int64(3), dbL1Txs[2].EthBlockNum)
	assert.Equal(t, int64(4), dbL1Txs[3].EthBlockNum)
	assert.Equal(t, int64(5), dbL1Txs[4].EthBlockNum)
	assert.Equal(t, int64(6), dbL1Txs[5].EthBlockNum)

	// User Origin
	assert.Equal(t, true, dbL1Txs[0].UserOrigin)
	assert.Equal(t, true, dbL1Txs[1].UserOrigin)
	assert.Equal(t, true, dbL1Txs[2].UserOrigin)
	assert.Equal(t, true, dbL1Txs[3].UserOrigin)
	assert.Equal(t, true, dbL1Txs[4].UserOrigin)
	assert.Equal(t, true, dbL1Txs[5].UserOrigin)

	// Deposit Amount
	assert.Equal(t, big.NewInt(1), dbL1Txs[0].DepositAmount)
	assert.Equal(t, big.NewInt(2), dbL1Txs[1].DepositAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[2].DepositAmount)
	assert.Equal(t, big.NewInt(0), dbL1Txs[3].DepositAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[4].DepositAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[5].DepositAmount)

	// Check saved txID's batch_num is not nil
	assert.NotEqual(t, null, dbL1Txs[len(dbL1Txs)-2].BatchNum)

	// Check L2 TXs
	dbL2Txs, err := historyDB.GetAllL2Txs()
	assert.NoError(t, err)
	assert.Equal(t, 6, len(dbL2Txs))

	// Tx Type
	assert.Equal(t, common.TxTypeCreateVouch, dbL2Txs[0].Type)
	assert.Equal(t, common.TxTypeCreateVouch, dbL2Txs[1].Type)
	assert.Equal(t, common.TxTypeExit, dbL2Txs[2].Type)
	assert.Equal(t, common.TxTypeCreateVouch, dbL2Txs[3].Type)
	assert.Equal(t, common.TxTypeCreateVouch, dbL2Txs[4].Type)
	assert.Equal(t, common.TxTypeDeleteVouch, dbL2Txs[5].Type)

	// Tx ID
	assert.Equal(t, "0x0222d2c1f4190752ad0d273024267197c5c65e1069dfff1baca7302e3fbca3c523", dbL2Txs[0].TxID.String())
	assert.Equal(t, "0x0297e5d629ad9740f54172116adcde0751068486987ffd904c0c46f3df8d9c81fb", dbL2Txs[1].TxID.String())
	assert.Equal(t, "0x021c11dcf29666db7add866b4969e7e1cbbb9d22debe27709310004cd56d969957", dbL2Txs[2].TxID.String())
	assert.Equal(t, "0x0217cc446dfb442ed9f0387b01c0e67166a5dbdc86fca8c6fef7df0fa2e2216c4a", dbL2Txs[3].TxID.String())
	assert.Equal(t, "0x02cbd4e1312607ee5074dd9897b57e84f9d8730f00afdf34a78739a697baecd7d5", dbL2Txs[4].TxID.String())
	assert.Equal(t, "0x028c138f5d828978c2c578280a70dc09353944312d5597e9972838146e8c539ec6", dbL2Txs[5].TxID.String())

	// Batch Number
	assert.Equal(t, common.BatchNum(3), dbL2Txs[0].BatchNum)
	assert.Equal(t, common.BatchNum(3), dbL2Txs[1].BatchNum)
	assert.Equal(t, common.BatchNum(4), dbL2Txs[2].BatchNum)
	assert.Equal(t, common.BatchNum(10), dbL2Txs[3].BatchNum)
	assert.Equal(t, common.BatchNum(10), dbL2Txs[4].BatchNum)
	assert.Equal(t, common.BatchNum(10), dbL2Txs[5].BatchNum)

	// eth_block_num
	assert.Equal(t, int64(2), dbL2Txs[0].EthBlockNum)
	assert.Equal(t, int64(2), dbL2Txs[1].EthBlockNum)
	assert.Equal(t, int64(3), dbL2Txs[2].EthBlockNum)
	assert.Equal(t, int64(7), dbL2Txs[3].EthBlockNum)
	assert.Equal(t, int64(7), dbL2Txs[4].EthBlockNum)
	assert.Equal(t, int64(7), dbL2Txs[5].EthBlockNum)

	// Amount
	assert.Equal(t, big.NewInt(10), dbL2Txs[2].Amount)
}

func TestExitTree(t *testing.T) {
	// Reset DB
	test.WipeDB(historyDB.DB())

	nBatches := 17

	blocks := setTestBlocks(1, 10)
	batches := test.GenBatches(nBatches, blocks)
	err := historyDB.AddBatches(batches)
	assert.NoError(t, err)

	const nAccounts = 3
	accs := test.GenAccounts(nAccounts, 0, nil, nil, batches)
	assert.NoError(t, historyDB.AddAccounts(accs))

	exitTree := test.GenExitTree(nBatches, batches, accs, blocks)
	err = historyDB.AddExitTree(exitTree)
	assert.NoError(t, err)
}

func TestGetUnforgedL1UserTxs(t *testing.T) {
	// Reset DB
	test.WipeDB(historyDB.DB())

	set := `
		Type: Blockchain

		CreateAccountDeposit A: 20
		CreateAccountDeposit A: 20
		CreateAccountDeposit B: 5
		CreateAccountDeposit C: 5
		CreateAccountDeposit D: 5
		> block

		> batchL1
		> block

		CreateAccountDeposit E: 5
		CreateAccountDeposit F: 5
		> block

	`
	tc := til.NewContext(uint16(0), 128)
	blocks, err := tc.GenerateBlocks(set)
	require.NoError(t, err)
	// Sanity check
	require.Equal(t, 3, len(blocks))
	require.Equal(t, 5, len(blocks[0].Rollup.L1UserTxs))

	for i := range blocks {
		err = historyDB.AddBlockSCData(&blocks[i])
		require.NoError(t, err)
	}

	l1UserTxs, err := historyDB.GetUnforgedL1UserFutureTxs(0)
	require.NoError(t, err)
	assert.Equal(t, 7, len(l1UserTxs))

	l1UserTxs, err = historyDB.GetUnforgedL1UserTxs(1)
	require.NoError(t, err)
	assert.Equal(t, 5, len(l1UserTxs))
	assert.Equal(t, blocks[0].Rollup.L1UserTxs, l1UserTxs)

	l1UserTxs, err = historyDB.GetUnforgedL1UserFutureTxs(1)
	require.NoError(t, err)
	assert.Equal(t, 2, len(l1UserTxs))

	count, err := historyDB.GetUnforgedL1UserTxsCount()
	require.NoError(t, err)
	assert.Equal(t, 7, count)

	l1UserTxs, err = historyDB.GetUnforgedL1UserTxs(2)
	require.NoError(t, err)
	assert.Equal(t, 2, len(l1UserTxs))

	l1UserTxs, err = historyDB.GetUnforgedL1UserFutureTxs(2)
	require.NoError(t, err)
	assert.Equal(t, 0, len(l1UserTxs))

	// No l1UserTxs for this toForgeL1TxsNum
	l1UserTxs, err = historyDB.GetUnforgedL1UserTxs(3)
	require.NoError(t, err)
	assert.Equal(t, 0, len(l1UserTxs))
}

func exampleInitSCVars() *common.RollupVariables { // *common.AuctionVariables,
	// *common.WDelayerVariables,

	rollup := &common.RollupVariables{
		EthBlockNum: 0,
		// FeeAddToken:           big.NewInt(10),
		ForgeL1L2BatchTimeout: 12,
		// WithdrawalDelay:       13,
		Buckets:  []common.BucketParams{},
		SafeMode: false,
	}
	// auction := &common.AuctionVariables{
	// 	EthBlockNum:        0,
	// 	DonationAddress:    ethCommon.BigToAddress(big.NewInt(2)),
	// 	BootCoordinator:    ethCommon.BigToAddress(big.NewInt(3)),
	// 	BootCoordinatorURL: "https://boot.coord.com",
	// 	DefaultSlotSetBid: [6]*big.Int{
	// 		big.NewInt(1), big.NewInt(2), big.NewInt(3),
	// 		big.NewInt(4), big.NewInt(5), big.NewInt(6),
	// 	},
	// 	DefaultSlotSetBidSlotNum: 0,
	// 	ClosedAuctionSlots:       2,
	// 	OpenAuctionSlots:         4320,
	// 	AllocationRatio:          [3]uint16{10, 11, 12},
	// 	Outbidding:               1000,
	// 	SlotDeadline:             20,
	// }
	// wDelayer := &common.WDelayerVariables{
	// 	EthBlockNum:                0,
	// 	HermezGovernanceAddress:    ethCommon.BigToAddress(big.NewInt(2)),
	// 	EmergencyCouncilAddress:    ethCommon.BigToAddress(big.NewInt(3)),
	// 	WithdrawalDelay:            13,
	// 	EmergencyModeStartingBlock: 14,
	// 	EmergencyMode:              false,
	// }
	return rollup //, auction, wDelayer
}

func TestSetInitialSCVars(t *testing.T) {
	// Reset DB
	test.WipeDB(historyDB.DB())

	_, err := historyDB.GetSCVars()
	assert.Equal(t, sql.ErrNoRows, tracerr.Unwrap(err))
	rollup := exampleInitSCVars()
	err = historyDB.SetInitialSCVars(rollup)
	require.NoError(t, err)
	dbRollup, err := historyDB.GetSCVars()
	require.NoError(t, err)
	require.Equal(t, rollup, dbRollup)
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
