package l2db

import (
	"tokamak-sybil-resistance/common"
)

// UpdateTxAPI Update PoolL2Tx regular transactions in the pool.
func (l2db *L2DB) UpdateTxAPI(tx *common.PoolL2Tx) error {
	cancel, err := l2db.apiConnCon.Acquire()
	defer cancel()
	if err != nil {
		return common.Wrap(err)
	}
	defer l2db.apiConnCon.Release()
	return common.Wrap(l2db.updateTx(*tx))
}
