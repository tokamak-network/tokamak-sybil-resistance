package database

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	"tokamak-sybil-resistance/common"

	"github.com/jmoiron/sqlx"
	"github.com/russross/meddler"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

var log *zap.SugaredLogger

// APIConnectionController is used to limit the SQL open connections used by the API
type APIConnectionController struct {
	smphr   *semaphore.Weighted
	timeout time.Duration
}

// NewAPIConnectionController initialize APIConnectionController
func NewAPIConnectionController(maxConnections int, timeout time.Duration) *APIConnectionController {
	return &APIConnectionController{
		smphr:   semaphore.NewWeighted(int64(maxConnections)),
		timeout: timeout,
	}
}

// Rollback an sql transaction, and log the error if it's not nil
func Rollback(txn *sqlx.Tx) {
	if err := txn.Rollback(); err != nil {
		log.Errorw("Rollback", "err", err)
	}
}

// BulkInsert performs a bulk insert with a single statement into the specified table.  Example:
// `db.BulkInsert(myDB, "INSERT INTO block (eth_block_num, timestamp, hash) VALUES %s", blocks[:])`
// Note that all the columns must be specified in the query, and they must be
// in the same order as in the table.
// Note that the fields in the structs need to be defined in the same order as
// in the table columns.
func BulkInsert(db meddler.DB, q string, args interface{}) error {
	arrayValue := reflect.ValueOf(args)
	arrayLen := arrayValue.Len()
	valueStrings := make([]string, 0, arrayLen)
	var arglist = make([]interface{}, 0)
	for i := 0; i < arrayLen; i++ {
		arg := arrayValue.Index(i).Addr().Interface()
		elemArglist, err := meddler.Default.Values(arg, true)
		if err != nil {
			return common.Wrap(err)
		}
		arglist = append(arglist, elemArglist...)
		value := "("
		for j := 0; j < len(elemArglist); j++ {
			value += fmt.Sprintf("$%d, ", i*len(elemArglist)+j+1)
		}
		value = value[:len(value)-2] + ")"
		valueStrings = append(valueStrings, value)
	}
	stmt := fmt.Sprintf(q, strings.Join(valueStrings, ","))
	_, err := db.Exec(stmt, arglist...)
	return common.Wrap(err)
}
