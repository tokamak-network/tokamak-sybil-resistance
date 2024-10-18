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
