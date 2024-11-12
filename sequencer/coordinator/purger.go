package coordinator

import (
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/database/l2db"
	"tokamak-sybil-resistance/database/statedb"
	"tokamak-sybil-resistance/log"
)

// PurgerCfg is the purger configuration
type PurgerCfg struct {
	// PurgeBatchDelay is the delay between batches to purge outdated
	// transactions. Outdated L2Txs are those that have been forged or
	// marked as invalid for longer than the SafetyPeriod and pending L2Txs
	// that have been in the pool for longer than TTL once there are
	// MaxTxs.
	PurgeBatchDelay int64
	// InvalidateBatchDelay is the delay between batches to mark invalid
	// transactions due to nonce lower than the account nonce.
	InvalidateBatchDelay int64
	// PurgeBlockDelay is the delay between blocks to purge outdated
	// transactions. Outdated L2Txs are those that have been forged or
	// marked as invalid for longer than the SafetyPeriod and pending L2Txs
	// that have been in the pool for longer than TTL once there are
	// MaxTxs.
	PurgeBlockDelay int64
	// InvalidateBlockDelay is the delay between blocks to mark invalid
	// transactions due to nonce lower than the account nonce.
	InvalidateBlockDelay int64
}

// Purger manages cleanup of transactions in the pool
type Purger struct {
	cfg                 PurgerCfg
	lastPurgeBlock      int64
	lastPurgeBatch      int64
	lastInvalidateBlock int64
	lastInvalidateBatch int64
}

// CanInvalidate returns true if it's a good time to invalidate according to
// the configuration
func (p *Purger) CanInvalidate(blockNum, batchNum int64) bool {
	if blockNum >= p.lastInvalidateBlock+p.cfg.InvalidateBlockDelay {
		return true
	}
	if batchNum >= p.lastInvalidateBatch+p.cfg.InvalidateBatchDelay {
		return true
	}
	return false
}

// CanPurge returns true if it's a good time to purge according to the
// configuration
func (p *Purger) CanPurge(blockNum, batchNum int64) bool {
	if blockNum >= p.lastPurgeBlock+p.cfg.PurgeBlockDelay {
		return true
	}
	if batchNum >= p.lastPurgeBatch+p.cfg.PurgeBatchDelay {
		return true
	}
	return false
}

// PurgeMaybe purges txs if it's a good time to do so
func (p *Purger) PurgeMaybe(l2DB *l2db.L2DB, blockNum, batchNum int64) (bool, error) {
	if !p.CanPurge(blockNum, batchNum) {
		return false, nil
	}
	p.lastPurgeBlock = blockNum
	p.lastPurgeBatch = batchNum
	log.Debugw("Purger: purging l2txs in pool", "block", blockNum, "batch", batchNum)
	err := l2DB.Purge(common.BatchNum(batchNum))
	return true, common.Wrap(err)
}

// InvalidateMaybe invalidates txs if it's a good time to do so
func (p *Purger) InvalidateMaybe(l2DB *l2db.L2DB, stateDB *statedb.LocalStateDB,
	blockNum, batchNum int64) (bool, error) {
	if !p.CanInvalidate(blockNum, batchNum) {
		return false, nil
	}
	p.lastInvalidateBlock = blockNum
	p.lastInvalidateBatch = batchNum
	log.Debugw("Purger: invalidating l2txs in pool", "block", blockNum, "batch", batchNum)

	//TODO: Update the functionality to handle old nouces for L2 txs check this needs to be for vouches or accounts
	// err := poolMarkInvalidOldNonces(l2DB, stateDB, common.BatchNum(batchNum))

	return true, nil
}

func idxsNonceFromPoolL2Txs(txs []common.PoolL2Tx) []common.IdxAccountNonce {
	idxNonceMap := map[common.AccountIdx]common.Nonce{}
	for _, tx := range txs {
		if nonce, ok := idxNonceMap[tx.FromIdx]; !ok {
			idxNonceMap[tx.FromIdx] = tx.Nonce
		} else if tx.Nonce > nonce {
			idxNonceMap[tx.FromIdx] = tx.Nonce
		}
	}
	idxsNonce := make([]common.IdxAccountNonce, 0, len(idxNonceMap))
	for idx, nonce := range idxNonceMap {
		idxsNonce = append(idxsNonce, common.IdxAccountNonce{Idx: idx, Nonce: nonce})
	}
	return idxsNonce
}
