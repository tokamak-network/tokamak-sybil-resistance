package statedb

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"tokamak-sybil-resistance/models"

	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db/pebble"
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

var (
	PrefixKeyAccHash    = []byte("accHash:")
	PrefixKeyLinkHash   = []byte("linkHash:")
	PrefixKeyAccountIdx = []byte("accIdx:")
	PrefixKeyLinkIdx    = []byte("linkIdx:")
)

// StateDB represents the state database with an integrated Merkle tree.
type StateDB struct {
	DB          *pebble.Storage
	AccountTree *merkletree.MerkleTree
	LinkTree    map[int]*merkletree.MerkleTree
}

// hashData computes the SHA-256 hash of the input data.
func HashData(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// initializeDB initializes and returns a Pebble DB instance.
func initializeDB(path string) (*pebble.Storage, error) {
	db, err := pebble.NewPebbleStorage(path, false)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// NewStateDB initializes a new StateDB.
func NewStateDB(dbPath string) (*StateDB, error) {
	db, err := initializeDB(dbPath)
	if err != nil {
		return nil, err
	}
	mt, _ := merkletree.NewMerkleTree(db, 14)
	fmt.Println(mt)
	return &StateDB{
		DB:          db,
		AccountTree: mt,
		LinkTree:    make(map[int]*merkletree.MerkleTree),
	}, nil
}

// Close closes the StateDB.
func (sdb *StateDB) Close() {
	sdb.DB.Close()
}

// performActions function for Account and Link are to test the db setup and
// it's mapping with merkel tree
func performActionsAccount(a *models.Account, s *StateDB, treeType enum) {
	proof, err := s.PutAccount(a)
	if err != nil {
		log.Fatalf("Failed to store key-value pair: %v", err)
	}
	fmt.Println(proof, "----------------------- Circom Processor Proof ---------------------")

	// Retrieve and print a value
	value, err := s.GetAccount(a.Idx)
	if err != nil {
		log.Fatalf("Failed to retrieve value: %v", err)
	}
	fmt.Printf("Retrieved account: %+v\n", value)

	// Get and print root hash for leaf
	root := s.GetMTRoot(a.Idx, treeType)
	fmt.Println(root, "MT root")
}

func performActionsLink(l *models.Link, s *StateDB, treeType enum) {
	proof, err := s.PutLink(l)
	if err != nil {
		log.Fatalf("Failed to store key-value pair: %v", err)
	}
	fmt.Println(proof, "----------------------- Circom Processor Proof ---------------------")
	// Retrieve and print a value
	value, err := s.GetLink(l.LinkIdx)
	if err != nil {
		log.Fatalf("Failed to retrieve value: %v", err)
	}
	fmt.Printf("Retrieved account: %+v\n", value)

	// Get and print root hash for leaf
	root := s.GetMTRoot(l.LinkIdx, treeType)
	fmt.Println(root, "MT root")
}

func printExamples(s *StateDB) {
	// Example accounts
	accountA := &models.Account{
		Idx:     10,
		EthAddr: "0xA",
		Sign:    true,
		Ay:      "ay_value",
		Balance: 10,
		Score:   1,
		Nonce:   0,
	}

	accountB := &models.Account{
		Idx:     20,
		EthAddr: "0xB",
		Sign:    true,
		Ay:      "ay_value",
		Balance: 10,
		Score:   1,
		Nonce:   0,
	}

	accountC := &models.Account{
		Idx:     30,
		EthAddr: "0xC",
		Sign:    true,
		Ay:      "ay_value",
		Balance: 10,
		Score:   1,
		Nonce:   0,
	}

	accountD := &models.Account{
		Idx:     40,
		EthAddr: "0xD",
		Sign:    true,
		Ay:      "ay_value",
		Balance: 10,
		Score:   1,
		Nonce:   0,
	}

	linkAB := &models.Link{
		LinkIdx: 1010,
		Value:   1,
	}

	linkAC := &models.Link{
		LinkIdx: 1030,
		Value:   1,
	}
	linkCD := &models.Link{
		LinkIdx: 3040,
		Value:   1,
	}
	linkCA := &models.Link{
		LinkIdx: 3010,
		Value:   1,
	}
	linkCB := &models.Link{
		LinkIdx: 3020,
		Value:   1,
	}
	// Add Account A
	performActionsAccount(accountA, s, Account)

	// Add Account B
	performActionsAccount(accountB, s, Account)

	//Add Account C
	performActionsAccount(accountC, s, Account)

	//Add Account D
	performActionsAccount(accountD, s, Account)

	//Add Link AB
	performActionsLink(linkAB, s, Link)

	performActionsLink(linkAC, s, Link)
	performActionsLink(linkCD, s, Link)
	performActionsLink(linkCA, s, Link)
	performActionsLink(linkCB, s, Link)

	// Print Merkle tree root
	// fmt.Printf("Merkle Account Tree Root: %s\n", s.AccountTree.Root.Hash)
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
