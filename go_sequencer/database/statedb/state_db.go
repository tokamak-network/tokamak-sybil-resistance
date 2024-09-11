package statedb

import (
	"errors"
	"log"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/database/kvdb"

	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db/pebble"
)

const (
	// TypeSynchronizer defines a StateDB used by the Synchronizer, that
	// generates the ExitTree when processing the txs
	TypeSynchronizer = "synchronizer"
	// TypeTxSelector defines a StateDB used by the TxSelector, without
	// computing ExitTree neither the ZKInputs
	TypeTxSelector = "txselector"
	// TypeBatchBuilder defines a StateDB used by the BatchBuilder, that
	// generates the ExitTree and the ZKInput when processing the txs
	TypeBatchBuilder = "batchbuilder"
	// MaxNLevels is the maximum value of NLevels for the merkle tree,
	// which comes from the fact that AccountIdx has 48 bits.
	MaxNLevels = 48
)

var (
	PrefixKeyAccountIdx = []byte("accIdx:")
	PrefixKeyLinkHash   = []byte("linkHash:")
	PrefixKeyLinkIdx    = []byte("linkIdx:")
)

// Config of the StateDB
type Config struct {
	// Path where the checkpoints will be stored
	Path string
	// Keep is the number of old checkpoints to keep.  If 0, all
	// checkpoints are kept.
	Keep int
	// NoLast skips having an opened DB with a checkpoint to the last
	// batchNum for thread-safe reads.
	NoLast bool
	// Type of StateDB (
	Type string
	// NLevels is the number of merkle tree levels in case the Type uses a
	// merkle tree.  If the Type doesn't use a merkle tree, NLevels should
	// be 0.
	NLevels int
	// At every checkpoint, check that there are no gaps between the
	// checkpoints
	noGapsCheck bool
}

var (
	// ErrStateDBWithoutMT is used when a method that requires a MerkleTree
	// is called in a StateDB that does not have a MerkleTree defined
	ErrStateDBWithoutMT = errors.New(
		"Can not call method to use MerkleTree in a StateDB without MerkleTree")

	// ErrAccountAlreadyExists is used when CreateAccount is called and the
	// Account already exists
	ErrAccountAlreadyExists = errors.New("Can not CreateAccount because Account already exists")

	// ErrIdxNotFound is used when trying to get the Idx from EthAddr or
	// EthAddr&ToBJJ
	ErrIdxNotFound = errors.New("Idx can not be found")
	// ErrGetIdxNoCase is used when trying to get the Idx from EthAddr &
	// BJJ with not compatible combination
	ErrGetIdxNoCase = errors.New(
		"Can not get Idx due unexpected combination of ethereum Address & BabyJubJub PublicKey")

	// PrefixKeyIdx is the key prefix for idx in the db
	PrefixKeyIdx = []byte("i:")
	// PrefixKeyAccHash is the key prefix for account hash in the db
	PrefixKeyAccHash = []byte("h:")
	// PrefixKeyMT is the key prefix for merkle tree in the db
	PrefixKeyMT = []byte("m:")
	// PrefixKeyAddr is the key prefix for address in the db
	PrefixKeyAddr = []byte("a:")
	// PrefixKeyAddrBJJ is the key prefix for address-babyjubjub in the db
	PrefixKeyAddrBJJ = []byte("ab:")
)

// StateDB represents the state database with an integrated Merkle tree.
type StateDB struct {
	cfg         Config
	DB          *kvdb.KVDB
	AccountTree *merkletree.MerkleTree
	LinkTree    *merkletree.MerkleTree
}

// LocalStateDB represents the local StateDB which allows to make copies from
// the synchronizer StateDB, and is used by the tx-selector and the
// batch-builder. LocalStateDB is an in-memory storage.
type LocalStateDB struct {
	*StateDB
	synchronizerStateDB *StateDB
}

// initializeDB initializes and returns a Pebble DB instance.
func initializeDB(path string) (*pebble.Storage, error) {
	db, err := pebble.NewPebbleStorage(path, false)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// NewStateDB initializes a new StateDB.
func NewStateDB(cfg Config) (*StateDB, error) {
	var kv *kvdb.KVDB
	var err error

	kv, err = kvdb.NewKVDB(kvdb.Config{Path: cfg.Path, Keep: cfg.Keep,
		NoGapsCheck: cfg.noGapsCheck, NoLast: cfg.NoLast})
	if err != nil {
		return nil, common.Wrap(err)
	}

	mtAccount, _ := merkletree.NewMerkleTree(kv.StorageWithPrefix(PrefixKeyMT), 14)
	mtLink, _ := merkletree.NewMerkleTree(kv.StorageWithPrefix(PrefixKeyMT), 14)
	return &StateDB{
		DB:          kv,
		AccountTree: mtAccount,
		LinkTree:    mtLink,
	}, nil
}

// Close closes the StateDB.
func (sdb *StateDB) Close() {
	sdb.DB.Close()
}

// NewLocalStateDB returns a new LocalStateDB connected to the given
// synchronizerDB.  Checkpoints older than the value defined by `keep` will be
// deleted.
func NewLocalStateDB(cfg Config, synchronizerDB *StateDB) (*LocalStateDB, error) {
	cfg.noGapsCheck = true
	cfg.NoLast = true
	s, err := NewStateDB(cfg)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &LocalStateDB{
		s,
		synchronizerDB,
	}, nil
}

// Reset resets the StateDB to the checkpoint at the given batchNum. Reset
// does not delete the checkpoints between old current and the new current,
// those checkpoints will remain in the storage, and eventually will be
// deleted when MakeCheckpoint overwrites them.
func (s *StateDB) Reset(batchNum common.BatchNum) error {
	log.Fatalf("Making StateDB Reset", "batch", batchNum, "type", s.cfg.Type)
	if err := s.DB.Reset(batchNum); err != nil {
		return common.Wrap(err)
	}
	if s.AccountTree != nil {
		// open the MT for the current s.db
		accountTree, err := merkletree.NewMerkleTree(s.DB.StorageWithPrefix(PrefixKeyMT), s.AccountTree.MaxLevels())
		if err != nil {
			return common.Wrap(err)
		}
		s.AccountTree = accountTree
	}
	if s.LinkTree != nil {
		// open the MT for the current s.db
		linkTree, err := merkletree.NewMerkleTree(s.DB.StorageWithPrefix(PrefixKeyMT), s.LinkTree.MaxLevels())
		if err != nil {
			return common.Wrap(err)
		}
		s.AccountTree = linkTree
	}
	return nil
}
