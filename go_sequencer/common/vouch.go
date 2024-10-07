package common

import (
	"encoding/binary"
	"fmt"
	"math/big"
)

// Vouch is a struct that gives an information about vouch
// between accounts. Each Idx is represented by fromIdx and toIdx
// of each accounts.
type Vouch struct {
	Idx      VouchIdx `meddler:"idx"`
	BatchNum BatchNum `meddler:"batch_num"`
	Value    bool     `meddler:"value"`
}

type VouchIdx uint64

const (
	// maxVouchIdxValue is the maximum value that VouchIdx can have
	maxVouchIdxValue = 0xffffffffffff
)

// Bytes returns a byte array representation the vouchIdx
func (idx VouchIdx) Bytes() ([2 * NLevelsAsBytes]byte, error) {
	if idx > maxVouchIdxValue {
		return [2 * NLevelsAsBytes]byte{}, Wrap(ErrIdxOverflow)
	}
	var idxBytes [8]byte
	binary.BigEndian.PutUint64(idxBytes[:], uint64(idx))
	var b [2 * NLevelsAsBytes]byte
	copy(b[:], idxBytes[8-2*NLevelsAsBytes:])
	return b, nil
}

func VouchIdxFromBytes(b []byte) (VouchIdx, error) {
	if len(b) != IdxBytesLen {
		return 0, Wrap(fmt.Errorf("can not parse Idx, bytes len %d, expected %d",
			len(b), IdxBytesLen))
	}
	var idxBytes [8]byte
	copy(idxBytes[8-2*NLevelsAsBytes:], b[:])
	idx := binary.BigEndian.Uint64(idxBytes[:])
	return VouchIdx(idx), nil
}

// BigInt returns a *big.Int representing the Idx
func (idx VouchIdx) BigInt() *big.Int {
	return big.NewInt(int64(idx))
}

// BytesFromBool returns []byte representing the vouch's value
func (b *Vouch) BytesFromBool() []byte {
	if b.Value {
		return []byte{1}
	} else {
		return []byte{0}
	}
}

// BigIntFromBool returns bool from *big.Int
func BigIntFromBool(b bool) *big.Int {
	if b {
		return big.NewInt(1)
	} else {
		return big.NewInt(0)
	}
}

// VouchFromBytes returns a vouch from [1]byte
func VouchFromBytes(b [1]byte) (*Vouch, error) {
	var value bool
	if b[0] == 1 {
		value = true
	} else {
		value = false
	}
	v := Vouch{
		Value: value,
	}
	return &v, nil
}
