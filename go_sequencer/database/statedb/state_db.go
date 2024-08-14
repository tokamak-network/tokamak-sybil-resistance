package statedb

import (
	"fmt"
	"log"
	"tokamak-sybil-resistance/models"

	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db/pebble"
)

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
	LinkTree    *merkletree.MerkleTree
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
	mtAccount, _ := merkletree.NewMerkleTree(db, 14)
	mtLink, _ := merkletree.NewMerkleTree(db, 14)
	return &StateDB{
		DB:          db,
		AccountTree: mtAccount,
		LinkTree:    mtLink,
	}, nil
}

// Close closes the StateDB.
func (sdb *StateDB) Close() {
	sdb.DB.Close()
}

// performActions function for Account and Link are to test the db setup and
// it's mapping with merkel tree
func performActionsAccount(a *models.Account, s *StateDB) {
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
	root := s.GetMTRoot(Account)
	fmt.Println(root, "MT root")
}

func performActionsLink(l *models.Link, s *StateDB) {
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
	root := s.GetMTRoot(Link)
	fmt.Println(root, "MT root")
}

func printExamples(s *StateDB) {
	// Example accounts
	accountA := &models.Account{
		Idx:     1,
		EthAddr: "0xA",
		BJJ:     "ay_value",
		Balance: 10,
		Score:   1,
		Nonce:   0,
	}

	accountB := &models.Account{
		Idx:     2,
		EthAddr: "0xB",
		BJJ:     "ay_value",
		Balance: 10,
		Score:   1,
		Nonce:   0,
	}

	accountC := &models.Account{
		Idx:     3,
		EthAddr: "0xC",
		BJJ:     "ay_value",
		Balance: 10,
		Score:   1,
		Nonce:   0,
	}

	accountD := &models.Account{
		Idx:     4,
		EthAddr: "0xD",
		BJJ:     "ay_value",
		Balance: 10,
		Score:   1,
		Nonce:   0,
	}

	linkAB := &models.Link{
		LinkIdx: 11,
		Value:   true,
	}

	linkAC := &models.Link{
		LinkIdx: 13,
		Value:   true,
	}
	linkCD := &models.Link{
		LinkIdx: 34,
		Value:   true,
	}
	linkCA := &models.Link{
		LinkIdx: 31,
		Value:   true,
	}
	linkCB := &models.Link{
		LinkIdx: 32,
		Value:   true,
	}
	// Add Account A
	performActionsAccount(accountA, s)

	// Add Account B
	performActionsAccount(accountB, s)

	//Add Account C
	performActionsAccount(accountC, s)

	//Add Account D
	performActionsAccount(accountD, s)

	//Add Link AB
	performActionsLink(linkAB, s)

	performActionsLink(linkAC, s)
	performActionsLink(linkCD, s)
	performActionsLink(linkCA, s)
	performActionsLink(linkCB, s)

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
