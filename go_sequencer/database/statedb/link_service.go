package statedb

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"tokamak-sybil-resistance/models"

	"github.com/iden3/go-iden3-crypto/poseidon"
	"github.com/iden3/go-merkletree"
)

func BytesLink(l *models.Link) [5]byte {
	var b [5]byte

	// Convert linkIdx into [4]byte
	binary.LittleEndian.PutUint32(b[0:4], uint32(l.LinkIdx))

	if l.Value {
		b[4] = 1
	} else {
		b[4] = 0
	}

	return b
}

func LinkFromBytes(b [5]byte) models.Link {
	var l models.Link

	// Extract Idx from [0:4]
	l.LinkIdx = int(binary.LittleEndian.Uint32(b[0:4]))

	l.Value = b[4] == 1

	return l
}

// Calculates poseidonHash for zk-Snark rollup Proof
func PoseidonHashLink(l *models.Link) (*big.Int, error) {
	bigInt := make([]*big.Int, 3)

	b := BytesLink(l)

	bigInt[0] = new(big.Int).SetBytes(b[0:2])
	bigInt[1] = new(big.Int).SetBytes(b[2:4])
	bigInt[2] = new(big.Int).SetBytes(b[4:4])

	return poseidon.Hash(bigInt)
}

// PutLink stores a link in the database and updates the Link Merkle tree.
func (sdb *StateDB) PutLink(l *models.Link) (*merkletree.CircomProcessorProof, error) {
	var idxBytes [4]byte
	linkBytes := BytesLink(l)
	binary.LittleEndian.PutUint32(idxBytes[:], uint32(l.LinkIdx))
	linkHash, _ := PoseidonHashLink(l)
	fmt.Println(linkHash, "---------------  Poseidon Hash Account ---------------")

	tx, err := sdb.db.NewTx()
	if err != nil {
		return nil, err
	}

	err = tx.Put(append(PrefixKeyLinkHash, linkHash.Bytes()...), linkBytes[:])
	if err != nil {
		return nil, err
	}
	err = tx.Put(append(PrefixKeyLinkIdx, idxBytes[:]...), linkHash.Bytes())
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return sdb.LinkTree.AddAndGetCircomProof(BigInt(l.LinkIdx), linkHash)
}

// GetLink retrieves a link for a given linkIdx from the database.
func (sdb *StateDB) GetLink(linkIdx int) (*models.Link, error) {
	var linkIdxBytes [4]byte
	// Convert Idx into [2]byte
	binary.LittleEndian.PutUint32(linkIdxBytes[0:4], uint32(linkIdx))

	linkHashBytes, err := sdb.db.Get(append(PrefixKeyLinkIdx, linkIdxBytes[:]...))
	if err != nil {
		return nil, err
	}

	linkBytes, err := sdb.db.Get(append(PrefixKeyLinkHash, linkHashBytes...))
	if err != nil {
		return nil, err
	}

	link := LinkFromBytes([5]byte(linkBytes))

	return &link, nil
}
