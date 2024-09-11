package common

import (
	"encoding/binary"
)

const (
	// MaxNonceValue is the maximum value that the Account.Nonce can have
	// (40 bits: MaxNonceValue=2**40-1)
	MaxNonceValue = 0xffffffffff
)

// Nonce represents the nonce value in a uint64, which has the method Bytes
// that returns a byte array of length 5 (40 bits).
type Nonce uint64

// Bytes returns a byte array of length 5 representing the Nonce
func (n Nonce) Bytes() ([5]byte, error) {
	if n > MaxNonceValue {
		return [5]byte{}, Wrap(ErrNonceOverflow)
	}
	var nonceBytes [8]byte
	binary.BigEndian.PutUint64(nonceBytes[:], uint64(n))
	var b [5]byte
	copy(b[:], nonceBytes[3:])
	return b, nil
}