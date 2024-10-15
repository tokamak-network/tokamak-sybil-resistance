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

// GetMTRoot returns the root of the Merkle Tree
func (s *StateDB) GetMTRoot(treeType enum) *big.Int {
	var root *big.Int
	if treeType == Account {
		root = s.AccountTree.Root().BigInt()
	} else if treeType == Vouch {
		root = s.VouchTree.Root().BigInt()
	} else {
		root = s.ScoreTree.Root().BigInt()
	}
	return root
}
