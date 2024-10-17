package statedb

import (
	"math/big"
)

type enum string

const (
	Account enum = "Account"
	Vouch   enum = "Vouch"
	Score   enum = "Score"
)

type TreeNodeHash interface {
}

// BigInt returns a *big.Int representing the Idx
func BigInt(idx int) *big.Int {
	return big.NewInt(int64(idx))
}

// GetMTRoot returns the root of the Merkle Tree Account
func (s *StateDB) GetMTRootAccount(treeType enum) *big.Int {
	return s.AccountTree.Root().BigInt()
}

// GetMTRoot returns the root of the Merkle Tree Vouch
func (s *StateDB) GetMTRootVouch(treeType enum) *big.Int {
	return s.VouchTree.Root().BigInt()
}
