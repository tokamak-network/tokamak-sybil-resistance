package common

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// Slot contains relevant information of a slot
type Slot struct {
	SlotNum          int64
	DefaultSlotBid   *big.Int
	StartBlock       int64
	EndBlock         int64
	ForgerCommitment bool
	// BatchesLen       int
	BidValue  *big.Int
	BootCoord bool
	// Bidder, Forger and URL correspond to the winner of the slot (which is
	// not always the highest bidder).  These are the values of the
	// coordinator that is able to forge exclusively before the deadline.
	Bidder ethCommon.Address
	Forger ethCommon.Address
	URL    string
}