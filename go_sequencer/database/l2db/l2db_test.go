package l2db

import (
	"os"
	"testing"
	"time"

	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/database"
	"tokamak-sybil-resistance/database/historydb"
	"tokamak-sybil-resistance/log"
	"tokamak-sybil-resistance/test/til"

	ethCommon "github.com/ethereum/go-ethereum/common"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var decimals = uint64(3)
var l2DB *L2DB
var l2DBWithACC *L2DB
var historyDB *historydb.HistoryDB
var tc *til.Context

var accs map[common.Idx]common.Account

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
	l2DB = NewL2DB(db, db, 10, 1000, 0.0, 1000.0, 24*time.Hour, nil)
	apiConnCon := database.NewAPIConnectionController(1, time.Second)
	l2DBWithACC = NewL2DB(db, db, 10, 1000, 0.0, 1000.0, 24*time.Hour, apiConnCon)
	WipeDB(l2DB.DB())
	historyDB = historydb.NewHistoryDB(db, db, nil)
	// Run tests
	result := m.Run()
	// Close DB
	if err := db.Close(); err != nil {
		log.Error("Error closing the history DB:", err)
	}
	os.Exit(result)
}

func prepareHistoryDB(historyDB *historydb.HistoryDB) error {
	// Reset DB
	WipeDB(l2DB.DB())
	// Generate pool txs using til
	setBlockchain := `
			Type: Blockchain

			CreateAccountDeposit A: 20000
			CreateAccountDeposit B: 10000
			> batchL1 
			> batchL1
			> block
			> block
			`

	tc = til.NewContext(uint16(0), common.RollupConstMaxL1UserTx)
	tilCfgExtra := til.ConfigExtra{
		BootCoordAddr: ethCommon.HexToAddress("0xE39fEc6224708f0772D2A74fd3f9055A90E0A9f2"),
		CoordUser:     "A",
	}
	blocks, err := tc.GenerateBlocks(setBlockchain)
	if err != nil {
		return common.Wrap(err)
	}

	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	if err != nil {
		return common.Wrap(err)
	}

	accs = make(map[common.Idx]common.Account)

	// Add all blocks except for the last one
	for i := range blocks[:len(blocks)-1] {
		if err := historyDB.AddBlockSCData(&blocks[i]); err != nil {
			return common.Wrap(err)
		}
		for _, batch := range blocks[i].Rollup.Batches {
			for _, account := range batch.CreatedAccounts {
				accs[account.Idx] = account
			}
		}
	}
	return nil
}

func generatePoolL2Txs() ([]common.PoolL2Tx, error) {
	// Fee = 126 corresponds to ~10%
	setPool := `
			Type: PoolL2
			PoolCreateVouch A-B
			PoolDeleteVouch A-B
			PoolCreateVouch B-A
			PoolDeleteVouch B-A

			// PoolExit A: 5000
			// PoolExit B: 3000
		`
	poolL2Txs, err := tc.GeneratePoolL2Txs(setPool)
	if err != nil {
		return nil, common.Wrap(err)
	}
	return poolL2Txs, nil
}

func TestAddTxTest(t *testing.T) {
	err := prepareHistoryDB(historyDB)
	if err != nil {
		log.Error("Error prepare historyDB", err)
	}
	poolL2Txs, err := generatePoolL2Txs()
	require.NoError(t, err)
	for i := range poolL2Txs {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		require.NoError(t, err)
		fetchedTx, err := l2DB.GetTx(poolL2Txs[i].TxID)
		require.NoError(t, err)
		assertTx(t, &poolL2Txs[i], fetchedTx)
		nameZone, offset := fetchedTx.Timestamp.Zone()
		assert.Equal(t, "UTC", nameZone)
		assert.Equal(t, 0, offset)
	}

	// test, that we can update already existing tx
	tx := &poolL2Txs[1]
	fetchedTx, err := l2DB.GetTx(tx.TxID)
	require.NoError(t, err)
	assert.Equal(t, fetchedTx.ToIdx, tx.ToIdx)
	tx.ToIdx = common.Idx(1)
	err = l2DBWithACC.UpdateTxAPI(tx)
	require.NoError(t, err)
	fetchedTx, err = l2DB.GetTx(tx.TxID)
	require.NoError(t, err)
	assert.Equal(t, fetchedTx.ToIdx, common.Idx(1))
}

func assertTx(t *testing.T, expected, actual *common.PoolL2Tx) {
	// Check that timestamp has been set within the last 3 seconds
	assert.Less(t, time.Now().UTC().Unix()-3, actual.Timestamp.Unix())
	assert.GreaterOrEqual(t, time.Now().UTC().Unix(), actual.Timestamp.Unix())
	expected.Timestamp = actual.Timestamp

	assert.Equal(t, expected, actual)
}
