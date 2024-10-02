/*
Package txprocessor is the module that takes the transactions from the input and
processes them, updating the Balances and Nonces of the Accounts in the StateDB.

It's a package used by 3 other different packages, and its behaviour will differ
depending on the Type of the StateDB of the TxProcessor:

- TypeSynchronizer:
  - The StateDB contains the full State MerkleTree, where the leafs are
    the accounts
  - Updates the StateDB and as output returns: ExitInfos, CreatedAccounts,
    CoordinatorIdxsMap, CollectedFees, UpdatedAccounts
  - Internally computes the ExitTree

- TypeTxSelector:
  - The StateDB contains only the Accounts, which are the equivalent to
    only the leafs of the State MerkleTree
  - Updates the Accounts from the StateDB

- TypeBatchBuilder:
  - The StateDB contains the full State MerkleTree, where the leafs are
    the accounts
  - Updates the StateDB. As output returns: ZKInputs, CoordinatorIdxsMap
  - Internally computes the ZKInputs

Packages dependency overview:

	Outputs: + ExitInfos              +                  +                       +
		 | CreatedAccounts        |                  |                       |
		 | CoordinatorIdxsMap     |                  |    ZKInputs           |
		 | CollectedFees          |                  |    CoordinatorIdxsMap |
		 | UpdatedAccounts        |                  |                       |
		 +------------------------+----------------+ +-----------------------+

		    +------------+           +----------+             +------------+
		    |Synchronizer|           |TxSelector|             |BatchBuilder|
		    +-----+------+           +-----+----+             +-----+------+
			  |                        |                        |
			  v                        v                        v
		     TxProcessor              TxProcessor              TxProcessor
			  +                        +                        +
			  |                        |                        |
		     +----+----+                   v                   +----+----+
		     |         |                StateDB                |         |
		     v         v                   +                   v         v
		  StateDB  ExitTree                |                StateDB  ExitTree
		     +                        +----+----+              +
		     |                        |         |              |
		+----+----+                   v         v         +----+----+
		|         |                 KVDB  AccountsDB      |         |
		v         v                                       v         v
	      KVDB   MerkleTree                                 KVDB   MerkleTree

The structure of the TxProcessor can be understand as:
  - StateDB: where the Rollup state is stored. It contains the Accounts &
    MerkleTree.
  - Config: parameters of the configuration of the circuit
  - ZKInputs: computed inputs for the circuit, depends on the Config parameters
  - ExitTree: only in the TypeSynchronizer & TypeBatchBuilder, contains
    the MerkleTree with the processed Exits of the Batch

The main exposed method of the TxProcessor is `ProcessTxs`, which as general
lines does:
  - if type==(Synchronizer || BatchBuilder), creates an ephemeral ExitTree
  - processes:
  - L1UserTxs --> for each tx calls ProcessL1Tx()
  - L1CoordinatorTxs --> for each tx calls ProcessL1Tx()
  - L2Txs --> for each tx calls ProcessL2Tx()
  - internally, it computes the Fees
  - each transaction processment includes:
  - updating the Account Balances (for sender & receiver, and in
    case that there is fee, updates the fee receiver account)
  - which includes updating the State MerkleTree (except
    for the type==TxSelector, which only updates the
    Accounts (leafs))
  - in case of Synchronizer & BatchBuilder, updates the ExitTree
    for the txs of type Exit (L1 & L2)
  - in case of BatchBuilder, computes the ZKInputs while processing the txs
  - if type==Synchronizer, once all the txs are processed, for each Exit
    it generates the ExitInfo data
*/
package txprocessor

import (
	"math/big"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/database/statedb"
)

// TxProcessor represents the TxProcessor object
type TxProcessor struct {
	state *statedb.StateDB
	zki   *common.ZKInputs
	// txIndex is the current transaction index in the ZKInputs generation (zki)
	txIndex int
	// AccumulatedFees contains the accumulated fees for each token (Coord
	// Idx) in the processed batch
	AccumulatedFees map[common.AccountIdx]*big.Int
	// updatedAccounts stores the last version of the account when it has
	// been created/updated by any of the processed transactions.
	updatedAccounts map[common.AccountIdx]*common.Account
	config          Config
}

// Config contains the TxProcessor configuration parameters
type Config struct {
	NLevels uint32
	// MaxFeeTx is the maximum number of coordinator accounts that can receive fees
	MaxFeeTx uint32
	MaxTx    uint32
	MaxL1Tx  uint32
	// ChainID of the blockchain
	ChainID uint16
}

type processedExit struct {
	exit    bool
	newExit bool
	idx     common.AccountIdx
	acc     common.Account
}

// ProcessTxOutput contains the output of the ProcessTxs method
type ProcessTxOutput struct {
	ZKInputs           *common.ZKInputs
	ExitInfos          []common.ExitInfo
	CreatedAccounts    []common.Account
	CoordinatorIdxsMap map[common.TokenID]common.AccountIdx
	CollectedFees      map[common.TokenID]*big.Int
	// UpdatedAccounts returns the current state of each account
	// created/updated by any of the processed transactions.
	UpdatedAccounts map[common.AccountIdx]*common.Account
}
