package coordinator

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
