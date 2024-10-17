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
	Score    ScoreType  `meddler:"score"` //30bits
}

type ScoreType uint32

// MaxScoreValue represents the maximum value for ScoreType.
const MaxScoreValue = ^ScoreType(0)

// Bytes converts the ScoreType to a [4]byte array.
func (s ScoreType) Bytes() ([4]byte, error) {
	if s > MaxScoreValue {
		return [4]byte{}, ErrScoreOverflow
	}
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], uint32(s))
	return b, nil
}

// FromBytes returns Score from a [4]byte
func GetScoreFromBytes(b [4]byte) *Score {
	var scoreBytes [4]byte
	copy(scoreBytes[:], b[:])
	score := binary.BigEndian.Uint32(scoreBytes[:])
	s := Score{Score: ScoreType(score)}
	return &s
}

// BigIntFromBool returns bool from *big.Int
func (s ScoreType) BigIntFromScore() *big.Int {
	bigInt := new(big.Int)
	bigInt.SetUint64(uint64(s))
	return bigInt
}
