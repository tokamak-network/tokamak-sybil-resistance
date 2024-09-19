package common

import (
	"encoding/binary"
	"errors"
	"math/big"
)

const (
	// maxFloat40Value is the maximum value that the Float40 can have
	// (40 bits: maxFloat40Value=2**40-1)
	maxFloat40Value = 0xffffffffff
	// Float40BytesLength defines the length of the Float40 values
	// represented as byte arrays
	Float40BytesLength = 5
)

var (
	// ErrFloat40Overflow is used when a given Float40 overflows the
	// maximum capacity of the Float40 (2**40-1)
	ErrFloat40Overflow = errors.New("Float40 overflow, max value: 2**40 -1")
	// ErrFloat40E31 is used when the e > 31 when trying to convert a
	// *big.Int to Float40
	ErrFloat40E31 = errors.New("Float40 error, e > 31")
	// ErrFloat40NotEnoughPrecission is used when the given *big.Int can
	// not be represented as Float40 due not enough precission
	ErrFloat40NotEnoughPrecission = errors.New("Float40 error, not enough precission")

	thres = big.NewInt(0x08_00_00_00_00)
)

// Float40 represents a float in a 64 bit format
type Float40 uint64

// Bytes return a byte array of length 5 with the Float40 value encoded in
// BigEndian
func (f40 Float40) Bytes() ([]byte, error) {
	if f40 > maxFloat40Value {
		return []byte{}, Wrap(ErrFloat40Overflow)
	}

	var f40Bytes [8]byte
	binary.BigEndian.PutUint64(f40Bytes[:], uint64(f40))
	var b [5]byte
	copy(b[:], f40Bytes[3:])
	return b[:], nil
}

// NewFloat40 encodes a *big.Int integer as a Float40, returning error in case
// of loss during the encoding.
func NewFloat40(f *big.Int) (Float40, error) {
	m := f
	e := big.NewInt(0)
	zero := big.NewInt(0)
	ten := big.NewInt(10)
	for new(big.Int).Mod(m, ten).Cmp(zero) == 0 && m.Cmp(thres) >= 0 {
		m = new(big.Int).Div(m, ten)
		e = new(big.Int).Add(e, big.NewInt(1))
	}
	if e.Int64() > 31 {
		return 0, Wrap(ErrFloat40E31)
	}
	if m.Cmp(thres) >= 0 {
		return 0, Wrap(ErrFloat40NotEnoughPrecission)
	}
	r := new(big.Int).Add(m,
		new(big.Int).Mul(e, thres))
	return Float40(r.Uint64()), nil
}
