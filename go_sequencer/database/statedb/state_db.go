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

// Account represents an account with specified fields.
type Account struct {
	Idx     string
	EthAddr string
	Sign    bool
	Ay      string
	Balance int
	Score   int
	Nonce   int
}

// Link represents a link structure.
type Link struct {
	LinkIdx string
	Value   int
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
	DB          *pebble.DB
	AccountTree *MerkleTree
	LinkTree    map[string]*MerkleTree
}

// NewStateDB initializes a new StateDB.
func NewStateDB(dbPath string) (*StateDB, error) {
	db, err := initializeDB(dbPath)
	if err != nil {
		return nil, err
	}
	return &StateDB{
		DB:          db,
		AccountTree: &MerkleTree{},
		LinkTree:    make(map[string]*MerkleTree),
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

	err = sdb.DB.Set([]byte(account.Idx), accountBytes, nil)
	if err != nil {
		return err
	}

	leaf := &TreeNode{Hash: hashData(string(accountBytes))}
	if sdb.AccountTree.Root == nil {
		sdb.AccountTree.Root = leaf
	} else {
		updateMerkleTree(sdb.AccountTree, leaf)
	}

	return nil
}

// Get retrieves an account for a given idx from the database.
func (sdb *StateDB) GetAccount(idx string) (*Account, error) {
	value, closer, err := sdb.DB.Get([]byte(idx))
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

// PutLink stores a link in the database and updates the Link Merkle tree.
func (sdb *StateDB) PutLink(link *Link) error {
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

	leaf := &TreeNode{Hash: hashData(string(linkBytes))}
	updateMerkleTree(linkTree, leaf)

	return nil
}

// GetLink retrieves a link for a given linkIdx from the database.
func (sdb *StateDB) GetLink(linkIdx string) (*Link, error) {
	value, closer, err := sdb.DB.Get([]byte(linkIdx))
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	var link Link
	err = json.Unmarshal(value, &link)
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func performActionsAccount(a *Account, s *StateDB, treeType enum) {
	err := s.PutAccount(a)
	if err != nil {
		log.Fatalf("Failed to store key-value pair: %v", err)
	}

	// Retrieve and print a value
	value, err := s.GetAccount(a.Idx)
	if err != nil {
		log.Fatalf("Failed to retrieve value: %v", err)
	}
	fmt.Printf("Retrieved account: %+v\n", value)

	// Print Merkle path
	path, _ := GetMerkelTreePath(s, a.Idx, treeType)
	fmt.Printf("Merkle path: %v\n", path)

	// Get and print root hash for leaf
	GetRootHash(s, a.Idx, treeType)
}

func performActionsLink(l *Link, s *StateDB, treeType enum) {
	err := s.PutLink(l)
	if err != nil {
		log.Fatalf("Failed to store key-value pair: %v", err)
	}

	// Retrieve and print a value
	value, err := s.GetLink(l.LinkIdx)
	if err != nil {
		log.Fatalf("Failed to retrieve value: %v", err)
	}
	fmt.Printf("Retrieved account: %+v\n", value)

	// Print Merkle path
	path, _ := GetMerkelTreePath(s, l.LinkIdx, treeType)
	fmt.Printf("Merkle path: %v\n", path)

	// Get and print root hash for leaf
	GetRootHash(s, l.LinkIdx, treeType)
}

func printExamples(s *StateDB) {
	// Example accounts
	accountA := &Account{
		Idx:     "011",
		EthAddr: "0xA",
		Sign:    true,
		Ay:      "ay_value",
		Balance: 10,
		Score:   1,
		Nonce:   0,
	}

	accountB := &Account{
		Idx:     "001",
		EthAddr: "0xB",
		Sign:    true,
		Ay:      "ay_value",
		Balance: 10,
		Score:   1,
		Nonce:   0,
	}

	accountC := &Account{
		Idx:     "101",
		EthAddr: "0xC",
		Sign:    true,
		Ay:      "ay_value",
		Balance: 10,
		Score:   1,
		Nonce:   0,
	}

	accountD := &Account{
		Idx:     "111",
		EthAddr: "0xD",
		Sign:    true,
		Ay:      "ay_value",
		Balance: 10,
		Score:   1,
		Nonce:   0,
	}

	linkAB := &Link{
		LinkIdx: "011001",
		Value:   1,
	}

	linkAC := &Link{
		LinkIdx: "011101",
		Value:   1,
	}
	linkCD := &Link{
		LinkIdx: "101111",
		Value:   1,
	}
	linkCA := &Link{
		LinkIdx: "101011",
		Value:   1,
	}
	linkCB := &Link{
		LinkIdx: "101001",
		Value:   1,
	}
	// Add Account A
	performActionsAccount(accountA, s, AccountTree)

	// Add Account B
	performActionsAccount(accountB, s, AccountTree)

	//Add Account C
	performActionsAccount(accountC, s, AccountTree)

	//Add Account D
	performActionsAccount(accountD, s, AccountTree)

	//Add Link AB
	performActionsLink(linkAB, s, LinkTree)

	performActionsLink(linkAC, s, LinkTree)
	performActionsLink(linkCD, s, LinkTree)
	performActionsLink(linkCA, s, LinkTree)
	performActionsLink(linkCB, s, LinkTree)

	// Print Merkle tree root
	fmt.Printf("Merkle Account Tree Root: %s\n", s.AccountTree.Root.Hash)
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
