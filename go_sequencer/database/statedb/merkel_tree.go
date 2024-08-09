package statedb

import (
	"math/big"
)

type enum string

const (
	Account enum = "Account"
	Link    enum = "Link"
)

type TreeNodeHash interface {
}

// BigInt returns a *big.Int representing the Idx
func BigInt(idx int) *big.Int {
	return big.NewInt(int64(idx))
}

// GetMTRoot returns the root of the Merkle Tree
func (s *StateDB) GetMTRoot(idx int, treeType enum) *big.Int {
	var root *big.Int
	if treeType == Account {
		root = s.AccountTree.Root().BigInt()
	} else {
		root = s.LinkTree[getAccountIdx(idx)].Root().BigInt()
	}
	return root
}
