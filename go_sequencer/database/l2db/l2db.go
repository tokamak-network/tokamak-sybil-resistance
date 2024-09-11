package l2db

import (
	"time"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/database"

	"github.com/jmoiron/sqlx"
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
