package common

// AtomicGroup represents a set of atomic transactions
type AtomicGroup struct {
	ID  AtomicGroupID `json:"atomicGroupId"`
	Txs []PoolL2Tx    `json:"transactions"`
}

// AtomicGroupIDLen is the length of a Hermez network atomic group
const AtomicGroupIDLen = 32

// AtomicGroupID is the identifier of a Hermez network atomic group
type AtomicGroupID [AtomicGroupIDLen]byte

// EmptyAtomicGroupID represents an empty Hermez network atomic group identifier
var EmptyAtomicGroupID = AtomicGroupID([32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
