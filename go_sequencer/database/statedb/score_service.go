package statedb

import (
	"errors"
	"tokamak-sybil-resistance/common"

	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db"
)

var (
	// ErrScoreAlreadyExists is used when CreateScore is called and the
	// Score already exists
	ErrScoreAlreadyExists = errors.New("cannot CreateScore becase Score already exists")
	// PrefixKeyScoIdx is the key prefix for accountIdx in ScoreTree
	PrefixKeyScoIdx = []byte("s:")
)

// CreateScore creates a new Score in the StateDB for the given Idx. If
// StateDB.MT==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func (s *StateDB) CreateScore(idx common.AccountIdx, score *common.Score) (
	*merkletree.CircomProcessorProof, error) {
	cpp, err := CreateScoreInTreeDB(s.db.DB(), s.ScoreTree, idx, score)
	if err != nil {
		return cpp, common.Wrap(err)
	}
	return cpp, nil
}

// CreateScoreInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree. Creates a new Score in the StateDB for the given Idx. If
// StateDB.MT==nil, MerkleTree is no affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof
func CreateScoreInTreeDB(sto db.Storage, mt *merkletree.MerkleTree, idx common.AccountIdx,
	score *common.Score) (*merkletree.CircomProcessorProof, error) {
	// store at the DB the key: idx, and value: leaf value
	tx, err := sto.NewTx()
	if err != nil {
		return nil, common.Wrap(err)
	}

	idxBytes, err := idx.Bytes()
	if err != nil {
		return nil, common.Wrap(err)
	}

	_, err = tx.Get(append(PrefixKeyScoIdx, idxBytes[:]...))
	if err != db.ErrNotFound {
		return nil, common.Wrap(ErrScoreAlreadyExists)
	}

	bytesFromScore, err := score.Bytes()
	if err != nil {
		return nil, common.Wrap(err)
	}

	err = tx.Put(append(PrefixKeyScoIdx, idxBytes[:]...), bytesFromScore[:])
	if err != nil {
		return nil, common.Wrap(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, common.Wrap(err)
	}

	if mt != nil {
		return mt.AddAndGetCircomProof(idx.BigInt(), score.BigInt())
	}

	return nil, nil
}

// GetScore returns the score for the given Idx
func (s *StateDB) GetScore(idx common.AccountIdx) (*common.Score, error) {
	return GetScoreInTreeDB(s.db.DB(), idx)
}

// GetScoreInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  GetScore returns the score for the given Idx
func GetScoreInTreeDB(sto db.Storage, idx common.AccountIdx) (*common.Score, error) {
	idxBytes, err := idx.Bytes()
	if err != nil {
		return nil, common.Wrap(err)
	}
	scoBytes, err := sto.Get(append(PrefixKeyScoIdx, idxBytes[:]...))
	if err != nil {
		return nil, common.Wrap(err)
	}
	var b [4]byte
	copy(b[:], scoBytes)
	score, err := common.ScoreFromBytes(b)
	if err != nil {
		return nil, common.Wrap(err)
	}
	score.Idx = idx
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
	// store at the DB the key: idx and value: leaf value
	tx, err := sto.NewTx()
	if err != nil {
		return nil, common.Wrap(err)
	}
	idxBytes, err := idx.Bytes()
	if err != nil {
		return nil, common.Wrap(err)
	}
	bytesFromScore, err := score.Bytes()
	if err != nil {
		return nil, common.Wrap(err)
	}

	err = tx.Put(append(PrefixKeyScoIdx, idxBytes[:]...), bytesFromScore[:])
	if err != nil {
		return nil, common.Wrap(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, common.Wrap(err)
	}

	if mt != nil {
		proof, err := mt.Update(idx.BigInt(), score.BigInt())
		return proof, common.Wrap(err)
	}
	return nil, nil
}
