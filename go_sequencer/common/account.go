package common

import (
	"encoding/binary"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// Account is a struct that gives information of the holdings of an address and
// a specific token. Is the data structure that generates the Value stored in
// the leaf of the MerkleTree
type Account struct {
	Idx      Idx                   `meddler:"idx"`
	TokenID  TokenID               `meddler:"token_id"`
	BatchNum BatchNum              `meddler:"batch_num"`
	BJJ      babyjub.PublicKeyComp `meddler:"bjj"`
	EthAddr  ethCommon.Address     `meddler:"eth_addr"`
	Nonce    Nonce           `meddler:"-"` // max of 40 bits used
	Balance  *big.Int              `meddler:"-"` // max of 192 bits used
}

// Idx represents the account Index in the MerkleTree
type Idx uint64

const (
	// NLeafElems is the number of elements for a leaf
	NLeafElems = 4

	// IdxBytesLen idx bytes
	IdxBytesLen = 6
	// maxIdxValue is the maximum value that Idx can have (48 bits:
	// maxIdxValue=2**48-1)
	maxIdxValue = 0xffffffffffff

	// UserThreshold determines the threshold from the User Idxs can be
	UserThreshold = 256
	// IdxUserThreshold is a Idx type value that determines the threshold
	// from the User Idxs can be
	IdxUserThreshold = Idx(UserThreshold)
)

// Bytes returns a byte array representing the Idx
func (idx Idx) Bytes() ([6]byte, error) {
	if idx > maxIdxValue {
		return [6]byte{}, Wrap(ErrIdxOverflow)
	}
	var idxBytes [8]byte
	binary.BigEndian.PutUint64(idxBytes[:], uint64(idx))
	var b [6]byte
	copy(b[:], idxBytes[2:])
	return b, nil
}