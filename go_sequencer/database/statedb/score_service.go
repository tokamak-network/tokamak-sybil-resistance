package statedb

import (
	"errors"
	"tokamak-sybil-resistance/common"

	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db"
)

var (
	ErrAlreadyScored = errors.New("cannot add score as already added")
	// PrefixKeyScoreIdx is the key prefix for scoreIdx in the db
	PrefixKeyScoreIdx = []byte("s:")
)

// CreateScore creates a new Score in the StateDB for the given Idx.
// Idx would be account idx for which the score is.
// If StateDB.MT==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func (s *StateDB) CreateScore(idx common.AccountIdx, score *common.Score) (
	*merkletree.CircomProcessorProof, error) {
	cpp, err := AddScoreInTreeDB(s.db.DB(), s.ScoreTree, idx, score)
	if err != nil {
		return cpp, common.Wrap(err)
	}
	return cpp, nil
}

// AddScoreInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree. Creates a new Score in the StateDB for the given Idx. If
// StateDB.MT==nil, MerkleTree is no affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof
func AddScoreInTreeDB(sto db.Storage, mt *merkletree.MerkleTree, idx common.AccountIdx,
	score *common.Score) (*merkletree.CircomProcessorProof, error) {
	// store at the DB the key: idx, and value: leaf value
	txError := performTxScore(sto, idx, score, true)
	if txError != nil {
		return nil, txError
	}
	if mt != nil {
		return mt.AddAndGetCircomProof(idx.BigInt(), score.Score.BigIntFromScore())
	}
	return nil, nil
}

// MTGetScoreProof returns the CircomVerifierProof for a given accountIdx
func (s *StateDB) MTGetScoreProof(idx common.AccountIdx) (*merkletree.CircomVerifierProof, error) {
	if s.ScoreTree == nil {
		return nil, common.Wrap(ErrStateDBWithoutMT)
	}
	p, err := s.ScoreTree.GenerateSCVerifierProof(idx.BigInt(), s.ScoreTree.Root())
	if err != nil {
		return nil, common.Wrap(err)
	}
	return p, nil
}

// GetScore returns the score for the given Idx
func (s *StateDB) GetScore(idx common.AccountIdx) (*common.Score, error) {
	return GetScoreInTreeDB(s.db.DB(), idx)
}

// GetScoreInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  Returns the score for the given Idx
func GetScoreInTreeDB(sto db.Storage, idx common.AccountIdx) (*common.Score, error) {
	idxBytes, err := idx.Bytes()
	if err != nil {
		return nil, common.Wrap(err)
	}
	scoreBytes, err := sto.Get(append(PrefixKeyScoreIdx, idxBytes[:]...))
	if err != nil {
		return nil, common.Wrap(err)
	}
	var b [4]byte
	copy(b[:], scoreBytes)
	score := common.GetScoreFromBytes(b)
	return score, nil
}

// UpdateScore updates the Score in the StateDB for the given Idx.  If
// StateDB.mt==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func (s *StateDB) UpdateScore(idx common.AccountIdx, score *common.Score) (
	*merkletree.CircomProcessorProof, error) {
	return UpdateScoreInTreeDB(s.db.DB(), s.ScoreTree, idx, score)
}

// UpdateScoreInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  Updates the Score in the StateDB for the given Idx.  If
// StateDB.mt==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func UpdateScoreInTreeDB(sto db.Storage, mt *merkletree.MerkleTree, idx common.AccountIdx,
	score *common.Score) (*merkletree.CircomProcessorProof, error) {
	txError := performTxScore(sto, idx, score, false)
	if txError != nil {
		return nil, txError
	}
	if mt != nil {
		proof, err := mt.Update(idx.BigInt(), score.Score.BigIntFromScore())
		return proof, common.Wrap(err)
	}
	return nil, nil
}

func performTxScore(sto db.Storage, idx common.AccountIdx,
	score *common.Score, addCall bool) error {
	// store at the DB the key: idx and value: leaf value
	tx, err := sto.NewTx()
	if err != nil {
		return common.Wrap(err)
	}
	idxBytes, err := idx.Bytes()
	if err != nil {
		return common.Wrap(err)
	}
	scoreBytes, err := score.Score.Bytes()
	if err != nil {
		return common.Wrap(err)
	}
	if addCall {
		_, err = tx.Get(append(PrefixKeyScoreIdx, idxBytes[:]...))
		if err != db.ErrNotFound {
			return common.Wrap(ErrAlreadyScored)
		}
	}
	err = tx.Put(append(PrefixKeyAccIdx, idxBytes[:]...), scoreBytes[:])
	if err != nil {
		return common.Wrap(err)
	}

	if err := tx.Commit(); err != nil {
		return common.Wrap(err)
	}
	return nil
}
