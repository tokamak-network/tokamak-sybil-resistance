package statedb

import (
	"errors"
	"tokamak-sybil-resistance/common"

	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db"
)

var (
	ErrAlreadyVouched = errors.New("Can not Vouch because already vouched")
	// PrefixKeyVocIdx is the key prefix for vouchIdx in the db
	PrefixKeyVocIdx = []byte("v:")
)

// CreateVouch creates a new Vouch in the StateDB for the given Idx. If
// StateDB.MT==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func (s *StateDB) CreateVouch(idx common.VouchIdx, vouch *common.Vouch) (
	*merkletree.CircomProcessorProof, error) {
	cpp, err := CreateVouchInTreeDB(s.db.DB(), s.VouchTree, idx, vouch)
	if err != nil {
		return cpp, common.Wrap(err)
	}
	return cpp, nil
}

// CreateVouchInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree. Creates a new Vouch in the StateDB for the given Idx. If
// StateDB.MT==nil, MerkleTree is no affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof
func CreateVouchInTreeDB(sto db.Storage, mt *merkletree.MerkleTree, idx common.VouchIdx,
	vouch *common.Vouch) (*merkletree.CircomProcessorProof, error) {
	// store at the DB the key: idx, and value: leaf value
	tx, err := sto.NewTx()
	if err != nil {
		return nil, common.Wrap(err)
	}

	idxBytes, err := idx.Bytes()
	if err != nil {
		return nil, common.Wrap(err)
	}

	_, err = tx.Get(append(PrefixKeyVocIdx, idxBytes[:]...))
	if err != db.ErrNotFound {
		return nil, common.Wrap(ErrAlreadyVouched)
	}

	err = tx.Put(append(PrefixKeyVocIdx, idxBytes[:]...), vouch.BytesFromBool())
	if err != nil {
		return nil, common.Wrap(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, common.Wrap(err)
	}

	if mt != nil {
		return mt.AddAndGetCircomProof(idx.BigInt(), common.BigIntFromBool(vouch.Value))
	}

	return nil, nil
}

// MTGetVouchProof returns the CircomVerifierProof for a given vouchIdx
func (s *StateDB) MTGetVouchProof(idx common.VouchIdx) (*merkletree.CircomVerifierProof, error) {
	if s.VouchTree == nil {
		return nil, common.Wrap(ErrStateDBWithoutMT)
	}
	p, err := s.VouchTree.GenerateSCVerifierProof(idx.BigInt(), s.VouchTree.Root())
	if err != nil {
		return nil, common.Wrap(err)
	}
	return p, nil
}

// GetVouch returns the vouch for the given Idx
func (s *StateDB) GetVouch(idx common.VouchIdx) (*common.Vouch, error) {
	return GetVouchInTreeDB(s.db.DB(), idx)
}

// GetVouchInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  GetVouch returns the vouch for the given Idx
func GetVouchInTreeDB(sto db.Storage, idx common.VouchIdx) (*common.Vouch, error) {
	idxBytes, err := idx.Bytes()
	if err != nil {
		return nil, common.Wrap(err)
	}
	vocBytes, err := sto.Get(append(PrefixKeyVocIdx, idxBytes[:]...))
	if err != nil {
		return nil, common.Wrap(err)
	}
	var b [1]byte
	copy(b[:], vocBytes)
	vouch, err := common.VouchFromBytes(b)
	if err != nil {
		return nil, common.Wrap(err)
	}
	vouch.Idx = idx
	return vouch, nil
}

// UpdateVouch updates the Vouch in the StateDB for the given Idx.  If
// StateDB.mt==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func (s *StateDB) UpdateVouch(idx common.VouchIdx, vouch *common.Vouch) (
	*merkletree.CircomProcessorProof, error) {
	return UpdateVouchInTreeDB(s.db.DB(), s.VouchTree, idx, vouch)
}

// UpdateVouchInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  Updates the Vouch in the StateDB for the given Idx.  If
// StateDB.mt==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func UpdateVouchInTreeDB(sto db.Storage, mt *merkletree.MerkleTree, idx common.VouchIdx,
	vouch *common.Vouch) (*merkletree.CircomProcessorProof, error) {
	// store at the DB the key: idx and value: leaf value
	tx, err := sto.NewTx()
	if err != nil {
		return nil, common.Wrap(err)
	}
	idxBytes, err := idx.Bytes()
	if err != nil {
		return nil, common.Wrap(err)
	}
	err = tx.Put(append(PrefixKeyAccIdx, idxBytes[:]...), vouch.BytesFromBool())
	if err != nil {
		return nil, common.Wrap(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, common.Wrap(err)
	}

	if mt != nil {
		proof, err := mt.Update(idx.BigInt(), common.BigIntFromBool(vouch.Value))
		return proof, common.Wrap(err)
	}
	return nil, nil
}

// func BytesLink(l *models.Link) [5]byte {
// 	var b [5]byte

// 	// Convert linkIdx into [4]byte
// 	binary.LittleEndian.PutUint32(b[0:4], uint32(l.LinkIdx))

// 	if l.Value {
// 		b[4] = 1
// 	} else {
// 		b[4] = 0
// 	}

// 	return b
// }

// func LinkFromBytes(b [5]byte) models.Link {
// 	var l models.Link

// 	// Extract Idx from [0:4]
// 	l.LinkIdx = int(binary.LittleEndian.Uint32(b[0:4]))

// 	l.Value = b[4] == 1

// 	return l
// }

// // Calculates poseidonHash for zk-Snark rollup Proof
// func PoseidonHashLink(l *models.Link) (*big.Int, error) {
// 	bigInt := make([]*big.Int, 3)

// 	b := BytesLink(l)

// 	bigInt[0] = new(big.Int).SetBytes(b[0:2])
// 	bigInt[1] = new(big.Int).SetBytes(b[2:4])
// 	bigInt[2] = new(big.Int).SetBytes(b[4:4])

// 	return poseidon.Hash(bigInt)
// }

// // PutLink stores a link in the database and updates the Link Merkle tree.
// func (sdb *StateDB) PutLink(l *models.Link) (*merkletree.CircomProcessorProof, error) {
// 	var idxBytes [4]byte
// 	linkBytes := BytesLink(l)
// 	binary.LittleEndian.PutUint32(idxBytes[:], uint32(l.LinkIdx))
// 	linkHash, _ := PoseidonHashLink(l)
// 	fmt.Println(linkHash, "---------------  Poseidon Hash Account ---------------")

// 	tx, err := sdb.db.NewTx()
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = tx.Put(append(PrefixKeyLinkHash, linkHash.Bytes()...), linkBytes[:])
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = tx.Put(append(PrefixKeyLinkIdx, idxBytes[:]...), linkHash.Bytes())
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err := tx.Commit(); err != nil {
// 		return nil, err
// 	}
// 	return sdb.LinkTree.AddAndGetCircomProof(BigInt(l.LinkIdx), linkHash)
// }

// // GetLink retrieves a link for a given linkIdx from the database.
// func (sdb *StateDB) GetLink(linkIdx int) (*models.Link, error) {
// 	var linkIdxBytes [4]byte
// 	// Convert Idx into [2]byte
// 	binary.LittleEndian.PutUint32(linkIdxBytes[0:4], uint32(linkIdx))

// 	linkHashBytes, err := sdb.db.Get(append(PrefixKeyLinkIdx, linkIdxBytes[:]...))
// 	if err != nil {
// 		return nil, err
// 	}

// 	linkBytes, err := sdb.db.Get(append(PrefixKeyLinkHash, linkHashBytes...))
// 	if err != nil {
// 		return nil, err
// 	}

// 	link := LinkFromBytes([5]byte(linkBytes))

// 	return &link, nil
// }
