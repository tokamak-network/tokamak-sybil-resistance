package historydb

import (
	"tokamak-sybil-resistance/database"

	"github.com/jmoiron/sqlx"
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