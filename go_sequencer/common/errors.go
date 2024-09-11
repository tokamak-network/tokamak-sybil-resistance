package common

import "errors"

// ErrIdxOverflow is used when a given nonce overflows the maximum capacity of the Idx (2**48-1)
var ErrIdxOverflow = errors.New("Idx overflow, max value: 2**48 -1")

// ErrNonceOverflow is used when a given nonce overflows the maximum capacity of the Nonce (2**40-1)
var ErrNonceOverflow = errors.New("Nonce overflow, max value: 2**40 -1")