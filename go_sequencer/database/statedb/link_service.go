package statedb

import (
	"encoding/json"
	"tokamak-sybil-resistance/models"
)

// PutLink stores a link in the database and updates the Link Merkle tree.
func (sdb *StateDB) PutLink(link *models.Link) error {
	linkBytes, err := json.Marshal(link)
	if err != nil {
		return err
	}

	err = sdb.DB.Set([]byte(link.LinkIdx), linkBytes, nil)
	if err != nil {
		return err
	}

	linkTree, exists := sdb.LinkTree[link.LinkIdx[:len(link.LinkIdx)/2]]
	if !exists {
		linkTree = &MerkleTree{}
		sdb.LinkTree[link.LinkIdx[:len(link.LinkIdx)/2]] = linkTree
	}

	leaf := &TreeNode{Hash: HashData(string(linkBytes))}
	UpdateMerkleTree(linkTree, leaf)

	return nil
}

// GetLink retrieves a link for a given linkIdx from the database.
func (sdb *StateDB) GetLink(linkIdx string) (*models.Link, error) {
	value, closer, err := sdb.DB.Get([]byte(linkIdx))
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	var link models.Link
	err = json.Unmarshal(value, &link)
	if err != nil {
		return nil, err
	}

	return &link, nil
}
