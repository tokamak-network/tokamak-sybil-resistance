package common

import (
	"encoding/binary"
	"math/big"
)

// Score is a struct that gives an information about score
// of each accounts.
type Score struct {
	Idx      AccountIdx `meddler:"idx"`
	BatchNum BatchNum   `meddler:"batch_num"`
	Value    uint32     `meddler:"score"`
}

const (
	// maxScore is the maximum value that Score can have (30 bits:
	// maxScoreValue=2**30-1)
	maxScoreValue = 0xffffffff
)

// Bytes returns a byte array representing the score
func (s *Score) Bytes() ([4]byte, error) {
	if s.Value > maxScoreValue {
		return [4]byte{}, Wrap(ErrScoreOverflow)
	}
	var scoreBytes [8]byte
	binary.BigEndian.PutUint64(scoreBytes[:], uint64(s.Value))
	var b [4]byte
	copy(b[:], scoreBytes[4:])
	return b, nil
}

// ScoreFromBytes returns score from a byte array
func ScoreFromBytes(b [4]byte) (*Score, error) {
	var scoreBytes [4]byte
	copy(scoreBytes[:], b[:])
	score := binary.BigEndian.Uint32(scoreBytes[:])
	s := Score{
		Value: score,
	}
	return &s, nil
}

// BigInt returns a *big.Int representing the score
func (s *Score) BigInt() *big.Int {
	return big.NewInt(int64(s.Value))
}
