/*
Package l2db is responsible for storing and retrieving the data received by the coordinator through the api.
Note that this data will be different for each coordinator in the network, as this represents the L2 information.

The data managed by this package is fundamentally PoolL2Tx and AccountCreationAuth. All this data come from
the API sent by clients and is used by the txselector to decide which transactions are selected to forge a batch.

Some of the database tooling used in this package such as meddler and migration tools is explained in the db package.

This package is spitted in different files following these ideas:
- l2db.go: constructor and functions used by packages other than the api.
- apiqueries.go: functions used by the API, the queries implemented in this functions use a semaphore
to restrict the maximum concurrent connections to the database.
- views.go: structs used to retrieve/store data from/to the database. When possible, the common structs are used, however
most of the time there is no 1:1 relation between the struct fields and the tables of the schema, especially when joining tables.
In some cases, some of the structs defined in this file also include custom Marshallers to easily match the expected api formats.
*/
package l2db

import (
	"errors"
	"fmt"
	"math/big"
	"time"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/database"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/russross/meddler"

	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/jmoiron/sqlx"
)

var (
	errPoolFull = fmt.Errorf("the pool is at full capacity. More transactions are not accepted currently")
)

// L2DB stores L2 txs and authorization registers received by the coordinator and keeps them until they are no longer relevant
// due to them being forged or invalid after a safety period
type L2DB struct {
	dbRead       *sqlx.DB
	dbWrite      *sqlx.DB
	safetyPeriod common.BatchNum
	ttl          time.Duration
	maxTxs       uint32 // limit of txs that are accepted in the pool
	minFeeUSD    float64
	maxFeeUSD    float64
	apiConnCon   *database.APIConnectionController
}

// NewL2DB creates a L2DB.
// To create it, it's needed db connection, safety period expressed in batches,
// maxTxs that the DB should have and TTL (time to live) for pending txs.
func NewL2DB(
	dbRead, dbWrite *sqlx.DB,
	safetyPeriod common.BatchNum,
	maxTxs uint32,
	minFeeUSD float64,
	maxFeeUSD float64,
	TTL time.Duration,
	apiConnCon *database.APIConnectionController,
) *L2DB {
	return &L2DB{
		dbRead:       dbRead,
		dbWrite:      dbWrite,
		safetyPeriod: safetyPeriod,
		ttl:          TTL,
		maxTxs:       maxTxs,
		minFeeUSD:    minFeeUSD,
		maxFeeUSD:    maxFeeUSD,
		apiConnCon:   apiConnCon,
	}
}

// DB returns a pointer to the L2DB.db. This method should be used only for
// internal testing purposes.
func (l2db *L2DB) DB() *sqlx.DB {
	return l2db.dbWrite
}

// MinFeeUSD returns the minimum fee in USD that is required to accept txs into
// the pool
func (l2db *L2DB) MinFeeUSD() float64 {
	return l2db.minFeeUSD
}

// AddAccountCreationAuth inserts an account creation authorization into the DB
func (l2db *L2DB) AddAccountCreationAuth(auth *common.AccountCreationAuth) error {
	_, err := l2db.dbWrite.Exec(
		`INSERT INTO account_creation_auth (eth_addr, bjj, signature)
		VALUES ($1, $2, $3);`,
		auth.EthAddr, auth.BJJ, auth.Signature,
	)
	return common.Wrap(err)
}

// AddTxTest inserts a tx into the L2DB, without security checks. This is useful for test purposes,
func (l2db *L2DB) AddTxTest(tx *common.PoolL2Tx) error {
	// Add tx without checking if pool is full
	return common.Wrap(
		l2db.addTxs([]common.PoolL2Tx{*tx}, false),
	)
}

// Insert PoolL2Tx transactions into the pool. If checkPoolIsFull is set to true the insert will
// fail if the pool is fool and errPoolFull will be returned
func (l2db *L2DB) addTxs(txs []common.PoolL2Tx, checkPoolIsFull bool) error {
	// Set the columns that will be affected by the insert on the table
	const queryInsertPart = `INSERT INTO tx_pool (
		tx_id, from_idx, to_idx, to_eth_addr, to_bjj, token_id,
		amount, fee, nonce, state, info, signature, rq_from_idx, 
		rq_to_idx, rq_to_eth_addr, rq_to_bjj, rq_token_id, rq_amount, rq_fee, rq_nonce, 
		tx_type, amount_f, client_ip, rq_offset, atomic_group_id, max_num_batch
	)`
	var (
		queryVarsPart string
		queryVars     []interface{}
	)
	for i := range txs {
		// Format extra DB fields and nullables
		var (
			toEthAddr *ethCommon.Address
			toBJJ     *babyjub.PublicKeyComp
			// Info (always nil)
			info *string
			// Rq fields, nil unless tx.RqFromIdx != 0
			rqFromIdx     *common.Idx
			rqToIdx       *common.Idx
			rqToEthAddr   *ethCommon.Address
			rqToBJJ       *babyjub.PublicKeyComp
			rqTokenID     *common.TokenID
			rqAmount      *string
			rqFee         *common.FeeSelector
			rqNonce       *common.Nonce
			rqOffset      *uint8
			atomicGroupID *common.AtomicGroupID
			maxNumBatch   *uint32
		)
		// AmountFloat
		f := new(big.Float).SetInt((*big.Int)(txs[i].Amount))
		amountF, _ := f.Float64()
		// ToEthAddr
		if txs[i].ToEthAddr != common.EmptyAddr {
			toEthAddr = &txs[i].ToEthAddr
		}
		// ToBJJ
		if txs[i].ToBJJ != common.EmptyBJJComp {
			toBJJ = &txs[i].ToBJJ
		}
		// MAxNumBatch
		if txs[i].MaxNumBatch != 0 {
			maxNumBatch = &txs[i].MaxNumBatch
		}
		// Rq fields
		if txs[i].RqFromIdx != 0 {
			// RqFromIdx
			rqFromIdx = &txs[i].RqFromIdx
			// RqToIdx
			if txs[i].RqToIdx != 0 {
				rqToIdx = &txs[i].RqToIdx
			}
			// RqToEthAddr
			if txs[i].RqToEthAddr != common.EmptyAddr {
				rqToEthAddr = &txs[i].RqToEthAddr
			}
			// RqToBJJ
			if txs[i].RqToBJJ != common.EmptyBJJComp {
				rqToBJJ = &txs[i].RqToBJJ
			}
			// RqTokenID
			rqTokenID = &txs[i].RqTokenID
			// RqAmount
			if txs[i].RqAmount != nil {
				rqAmountStr := txs[i].RqAmount.String()
				rqAmount = &rqAmountStr
			}
			// RqFee
			rqFee = &txs[i].RqFee
			// RqNonce
			rqNonce = &txs[i].RqNonce
			// RqOffset
			rqOffset = &txs[i].RqOffset
			// AtomicGroupID
			atomicGroupID = &txs[i].AtomicGroupID
		}
		// Each ? match one of the columns to be inserted as defined in queryInsertPart
		const queryVarsPartPerTx = `(?::BYTEA, ?::BIGINT, ?::BIGINT, ?::BYTEA, ?::BYTEA, ?::INT, 
		?::NUMERIC, ?::SMALLINT, ?::BIGINT, ?::CHAR(4), ?::VARCHAR, ?::BYTEA, ?::BIGINT,
		?::BIGINT, ?::BYTEA, ?::BYTEA, ?::INT, ?::NUMERIC, ?::SMALLINT, ?::BIGINT,
		?::VARCHAR(40), ?::NUMERIC, ?::VARCHAR, ?::SMALLINT, ?::BYTEA, ?::BIGINT)`
		if i == 0 {
			queryVarsPart += queryVarsPartPerTx
		} else {
			// Add coma before next tx values.
			queryVarsPart += ", " + queryVarsPartPerTx
		}
		// Add values that will replace the ?
		// caution: hardcoded tokenID and amount as 0
		queryVars = append(queryVars,
			txs[i].TxID, txs[i].FromIdx, txs[i].ToIdx, toEthAddr, toBJJ, txs[i].TokenID,
			txs[i].Amount.String(), txs[i].Fee, txs[i].Nonce, txs[i].State, info, txs[i].Signature, rqFromIdx,
			rqToIdx, rqToEthAddr, rqToBJJ, rqTokenID, rqAmount, rqFee, rqNonce,
			txs[i].Type, amountF, txs[i].ClientIP, rqOffset, atomicGroupID, maxNumBatch,
		)
	}
	// Query begins with the insert statement
	query := queryInsertPart
	if checkPoolIsFull {
		// This query creates a temporary table containing the values to insert
		// that will only get selected if the pool is not full
		query += " SELECT * FROM ( VALUES " + queryVarsPart + " ) as tmp " + // Temporary table with the values of the txs
			" WHERE (SELECT COUNT (*) FROM tx_pool WHERE state = ? AND NOT external_delete) < ?;" // Check if the pool is full
		queryVars = append(queryVars, common.PoolL2TxStatePending, l2db.maxTxs)
	} else {
		query += " VALUES " + queryVarsPart + ";"
	}
	// Replace "?, ?, ... ?" ==> "$1, $2, ..., $(len(queryVars))"
	query = l2db.dbRead.Rebind(query)
	// Execute query
	res, err := l2db.dbWrite.Exec(query, queryVars...)
	if err == nil && checkPoolIsFull {
		if rowsAffected, err := res.RowsAffected(); err != nil || rowsAffected == 0 {
			// If the query didn't affect any row, and there is no error in the query
			// it's safe to assume that the WERE clause wasn't true, and so the pool is full
			return common.Wrap(errPoolFull)
		}
	}
	return common.Wrap(err)
}

// selectPoolTxCommon select part of queries to get common.PoolL2Tx
const selectPoolTxCommon = `SELECT  tx_pool.tx_id, from_idx, to_idx, tx_pool.to_eth_addr, 
tx_pool.to_bjj, tx_pool.token_id, tx_pool.amount, tx_pool.fee, tx_pool.nonce, 
tx_pool.state, tx_pool.info, tx_pool.signature, tx_pool.timestamp, rq_from_idx, 
rq_to_idx, tx_pool.rq_to_eth_addr, tx_pool.rq_to_bjj, tx_pool.rq_token_id, tx_pool.rq_amount, 
tx_pool.rq_fee, tx_pool.rq_nonce, tx_pool.tx_type, tx_pool.rq_offset, tx_pool.atomic_group_id, tx_pool.max_num_batch, 
(fee_percentage(tx_pool.fee::NUMERIC) * token.usd * tx_pool.amount_f) /
	(10.0 ^ token.decimals::NUMERIC) AS fee_usd, token.usd_update
FROM tx_pool INNER JOIN token ON tx_pool.token_id = token.token_id `

// GetTx  return the specified Tx in common.PoolL2Tx format
func (l2db *L2DB) GetTx(txID common.TxID) (*common.PoolL2Tx, error) {
	tx := new(common.PoolL2Tx)
	return tx, common.Wrap(meddler.QueryRow(
		l2db.dbRead, tx,
		selectPoolTxCommon+"WHERE tx_id = $1;",
		txID,
	))
}

// Update PoolL2Tx transaction in the pool
func (l2db *L2DB) updateTx(tx common.PoolL2Tx) error {
	const queryUpdate = `UPDATE tx_pool SET to_idx = ?, to_eth_addr = ?, to_bjj = ?, max_num_batch = ?, 
	signature = ?, client_ip = ?, tx_type = ? WHERE tx_id = ? AND tx_pool.atomic_group_id IS NULL;`

	if tx.ToIdx == 0 && tx.ToEthAddr == common.EmptyAddr && tx.ToBJJ == common.EmptyBJJComp && tx.MaxNumBatch == 0 {
		return common.Wrap(errors.New("nothing to update"))
	}

	queryVars := []interface{}{tx.ToIdx, tx.ToEthAddr, tx.ToBJJ, tx.MaxNumBatch, tx.Signature, tx.ClientIP, tx.Type, tx.TxID}

	query, args, err := sqlx.In(queryUpdate, queryVars...)
	if err != nil {
		return common.Wrap(err)
	}

	query = l2db.dbWrite.Rebind(query)
	_, err = l2db.dbWrite.Exec(query, args...)
	return common.Wrap(err)
}
