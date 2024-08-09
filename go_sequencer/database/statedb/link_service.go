package statedb

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"tokamak-sybil-resistance/models"

	"github.com/iden3/go-iden3-crypto/poseidon"
	"github.com/iden3/go-merkletree"
)

// Calculates poseidonHash for zk-Snark rollup Proof
func PoseidonHashLink(l *models.Link) (*big.Int, error) {
	var bigInts []*big.Int

	// Convert LinkIdx
	linkIdx := big.NewInt(int64(l.LinkIdx))
	bigInts = append(bigInts, linkIdx)

	// Convert Value
	value := big.NewInt(int64(l.Value))
	bigInts = append(bigInts, value)

	return poseidon.Hash(bigInts)
}

// Get Account Idx
func getAccountIdx(n int) int {
	numStr := strconv.Itoa(n)
	length := len(numStr)
	accountStr := numStr[:length/2]
	accountIdx, _ := strconv.Atoi(accountStr)
	return accountIdx
}

// PutLink stores a link in the database and updates the Link Merkle tree.
func (sdb *StateDB) PutLink(l *models.Link) (*merkletree.CircomProcessorProof, error) {
	linkBytes, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	idxBytes, err := json.Marshal(l.LinkIdx)
	if err != nil {
		return nil, err
	}
	linkHash, _ := PoseidonHashLink(l)
	fmt.Println(linkHash, "---------------  Poseidon Hash Account ---------------")

	tx, err := sdb.DB.NewTx()
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
	_, exists := sdb.LinkTree[getAccountIdx(l.LinkIdx)]
	if !exists {
		linkTree, _ := merkletree.NewMerkleTree(sdb.DB, 14)
		sdb.LinkTree[getAccountIdx(l.LinkIdx)] = linkTree
	}
	return sdb.LinkTree[getAccountIdx(l.LinkIdx)].AddAndGetCircomProof(BigInt(l.LinkIdx), linkHash)
}

// GetLink retrieves a link for a given linkIdx from the database.
func (sdb *StateDB) GetLink(linkIdx int) (*models.Link, error) {
	linkIdxBytes, err := json.Marshal(linkIdx)
	if err != nil {
		return nil, err
	}

	linkHashBytes, err := sdb.DB.Get(append(PrefixKeyLinkIdx, linkIdxBytes[:]...))
	if err != nil {
		return nil, err
	}

	linkBytes, err := sdb.DB.Get(append(PrefixKeyLinkHash, linkHashBytes...))
	if err != nil {
		return nil, err
	}

	var link models.Link
	err = json.Unmarshal(linkBytes, &link)
	if err != nil {
		return nil, err
	}

	return &link, nil
}
