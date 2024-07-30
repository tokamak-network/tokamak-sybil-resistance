package statedb

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cockroachdb/pebble"
)

// TreeNode represents a node in the Merkle tree.
type TreeNode struct {
	Hash  string
	Left  *TreeNode
	Right *TreeNode
}

// MerkleTree represents a Merkle tree.
type MerkleTree struct {
	Root *TreeNode
}

// Account represents an account with a link subtree.
type Account struct {
	Address  string
	Deposit  int
	Nonce    int
	Score    int
	LinkRoot string // Root hash of the Links Subtree
}

// Link represents a link structure.
type Link struct {
	TargetID string
	Stake    int
}

// hashData computes the SHA-256 hash of the input data.
func hashData(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// initializeDB initializes and returns a Pebble DB instance.
func initializeDB(path string) (*pebble.DB, error) {
	db, err := pebble.Open(path, &pebble.Options{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// closeDB closes the Pebble DB instance.
func closeDB(db *pebble.DB) error {
	return db.Close()
}

// StateDB represents the state database with an integrated Merkle tree.
type StateDB struct {
	DB   *pebble.DB
	Tree *MerkleTree
}

// NewStateDB initializes a new StateDB.
func NewStateDB(dbPath string) (*StateDB, error) {
	db, err := initializeDB(dbPath)
	if err != nil {
		return nil, err
	}
	return &StateDB{
		DB:   db,
		Tree: &MerkleTree{},
	}, nil
}

// Close closes the StateDB.
func (sdb *StateDB) Close() error {
	return closeDB(sdb.DB)
}

// Put stores an account in the database and updates the Merkle tree.
func (sdb *StateDB) PutAccount(account *Account) error {
	accountBytes, err := json.Marshal(account)
	if err != nil {
		return err
	}

	err = sdb.DB.Set([]byte(account.Address), accountBytes, nil)
	if err != nil {
		return err
	}

	leaf := &TreeNode{Hash: hashData(string(accountBytes))}
	if sdb.Tree.Root == nil {
		sdb.Tree.Root = leaf
	} else {
		updateMerkleTree(sdb.Tree, leaf)
	}

	return nil
}

// Get retrieves an account for a given address from the database.
func (sdb *StateDB) GetAccount(address string) (*Account, error) {
	value, closer, err := sdb.DB.Get([]byte(address))
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	var account Account
	err = json.Unmarshal(value, &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// UpdateLink updates the link tree for an account and updates the account's link root in the database.
func (sdb *StateDB) UpdateLink(address string, link *Link) error {
	account, err := sdb.GetAccount(address)
	if err != nil {
		return err
	}

	// Deserialize the existing link tree if it exists
	var linkTree MerkleTree
	if account.LinkRoot != "" {
		linkTree.Root = &TreeNode{Hash: account.LinkRoot}
	}

	linkBytes, err := json.Marshal(link)
	if err != nil {
		return err
	}

	linkLeaf := &TreeNode{Hash: hashData(string(linkBytes))}
	updateMerkleTree(&linkTree, linkLeaf)

	account.LinkRoot = linkTree.Root.Hash

	return sdb.PutAccount(account)
}

// GetMerklePath retrieves the Merkle path for a given key-value pair.
func (sdb *StateDB) GetMerklePath(key, value string) ([]string, error) {
	targetHash := hashData(key + value)
	path, found := FindPathToRoot(sdb.Tree.Root, targetHash)
	if !found {
		return nil, fmt.Errorf("path not found for key: %s", key)
	}
	return path, nil
}

func performActionsAccount(a *Account, s *StateDB) {
	err := s.PutAccount(a)
	if err != nil {
		log.Fatalf("Failed to store key-value pair: %v", err)
	}

	// Retrieve and print a value
	value, err := s.GetAccount(a.Address)
	if err != nil {
		log.Fatalf("Failed to retrieve value: %v", err)
	}
	fmt.Printf("Retrieved account: %+v\n", value)

	// Print Merkle path
	path, _ := GetMerkelTreePath(s, a.Address)
	fmt.Printf("Merkle path: %v\n", path)

	// Get and print root hash for leaf
	GetRootHash(s, a.Address)

}

func printExamples(s *StateDB) {
	// Example accounts
	accountA := &Account{
		Address: "account_0xA",
		Deposit: 10,
		Nonce:   1,
		Score:   0,
	}

	accountB := &Account{
		Address: "account_0xB",
		Deposit: 5,
		Nonce:   2,
		Score:   0,
	}

	accountC := &Account{
		Address: "account_0xC",
		Deposit: 5,
		Nonce:   2,
		Score:   0,
	}

	accountD := &Account{
		Address: "account_0xD",
		Deposit: 5,
		Nonce:   2,
		Score:   0,
	}
	// Add Account A
	performActionsAccount(accountA, s)

	// Add Account B
	performActionsAccount(accountB, s)

	//Add Account C
	performActionsAccount(accountC, s)

	//Add Account D
	performActionsAccount(accountD, s)

	// Print Merkle tree root
	fmt.Printf("Merkle Tree Root: %s\n", s.Tree.Root.Hash)
}

func InitNewStateDB() *StateDB {
	// Initialize the StateDB
	stateDB, err := NewStateDB("stateDB")
	if err != nil {
		log.Fatalf("Failed to initialize StateDB: %v", err)
	}
	defer stateDB.Close()
	printExamples(stateDB)
	return stateDB
}
