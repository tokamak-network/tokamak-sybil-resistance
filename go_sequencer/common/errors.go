package common

import (
	"errors"
)

// ErrNotInFF is used when the *big.Int does not fit inside the Finite Field
var ErrNotInFF = errors.New("BigInt not inside the Finite Field")

// ErrNumOverflow is used when a given value overflows the maximum capacity of the parameter
var ErrNumOverflow = errors.New("Value overflows the type")

// ErrIdxOverflow is used when a given nonce overflows the maximum capacity of the Idx (2**48-1)
var ErrIdxOverflow = errors.New("Idx overflow, max value: 2**48 -1")

// ErrBatchQueueEmpty is used when the coordinator.BatchQueue.Pop() is called and has no elements
var ErrBatchQueueEmpty = errors.New("BatchQueue empty")

// ErrScoreOverflow is used when a given score overflows the maximum capacity of the uint32 4,294,967,295
var ErrScoreOverflow = errors.New("Score overflow, max value uint32 - 4,294,967,295")

// ErrTODO is used when a function is not yet implemented
var ErrTODO = errors.New("TODO")

// ErrDone is used when a function returns earlier due to a cancelled context
var ErrDone = errors.New("done")

// IsErrDone returns true if the error or wrapped (with tracerr) error is ErrDone
func IsErrDone(err error) bool {
	return Unwrap(err) == ErrDone
}
