package statedb

import (
	"encoding/json"
	"tokamak-sybil-resistance/models"
)

// Put stores an account in the database and updates the Merkle tree.
func (sdb *StateDB) PutAccount(account *models.Account) error {
	accountBytes, err := json.Marshal(account)
	if err != nil {
		return err
	}

	err = sdb.DB.Set([]byte(account.Idx), accountBytes, nil)
	if err != nil {
		return err
	}

	leaf := &TreeNode{Hash: HashData(string(accountBytes))}
	if sdb.AccountTree.Root == nil {
		sdb.AccountTree.Root = leaf
	} else {
		UpdateMerkleTree(sdb.AccountTree, leaf)
	}

	return nil
}

// Get retrieves an account for a given idx from the database.
func (sdb *StateDB) GetAccount(idx string) (*models.Account, error) {
	value, closer, err := sdb.DB.Get([]byte(idx))
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	var account models.Account
	err = json.Unmarshal(value, &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}
