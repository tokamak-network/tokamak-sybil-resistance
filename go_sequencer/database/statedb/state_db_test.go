package statedb

import (
	"encoding/hex"
	"math/big"
	"math/rand"
	"os"
	"testing"
	"time"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/log"

	ethCrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var deleteme []string

func init() {
	log.Init("debug", []string{"stdout"})
}
func TestMain(m *testing.M) {
	exitVal := 0
	exitVal = m.Run()
	for _, dir := range deleteme {
		if err := os.RemoveAll(dir); err != nil {
			panic(err)
		}
	}
	os.Exit(exitVal)
}

func newAccount(t *testing.T, i int) *common.Account {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:],
		[]byte("0001020304050607080900010203040506070809000102030405060708090001"))
	require.NoError(t, err)
	pk := sk.Public()

	key, err := ethCrypto.GenerateKey()
	require.NoError(t, err)
	address := ethCrypto.PubkeyToAddress(key.PublicKey)

	return &common.Account{
		Idx:     common.AccountIdx(256 + i),
		Nonce:   common.Nonce(i),
		Balance: big.NewInt(1000),
		BJJ:     pk.Compress(),
		EthAddr: address,
	}
}

func newVouch(i int) *common.Vouch {
	r := rand.New(rand.NewSource(int64(time.Now().UnixNano())))
	v := r.Intn(2) == 1
	return &common.Vouch{
		Idx:   common.VouchIdx(256257 + i),
		Value: v,
	}
}

func newScore(i int) *common.Score {
	return &common.Score{
		Idx:   common.AccountIdx(256 + i),
		Value: uint64(1 + i),
	}
}

func TestAccountInStateDB(t *testing.T) {
	dir, err := os.MkdirTemp("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	sdb, err := NewStateDB(Config{Path: dir, Keep: 128, Type: TypeSynchronizer, NLevels: 0})
	require.NoError(t, err)

	// create test accounts
	var accounts []*common.Account
	for i := 0; i < 4; i++ {
		accounts = append(accounts, newAccount(t, i))
	}

	// get non-existing account, expecting an error
	unexistingAccount := common.AccountIdx(1)
	_, err = sdb.GetAccount(unexistingAccount)
	assert.NotNil(t, err)
	assert.Equal(t, db.ErrNotFound, common.Unwrap(err))

	// add test accounts
	for i := 0; i < len(accounts); i++ {
		_, err = sdb.CreateAccount(accounts[i].Idx, accounts[i])
		require.NoError(t, err)
	}

	for i := 0; i < len(accounts); i++ {
		existingAccount := accounts[i].Idx
		accGetted, err := sdb.GetAccount(existingAccount)
		require.NoError(t, err)
		assert.Equal(t, accounts[i], accGetted)
	}

	// try already existing idx and get error
	existingAccount := common.AccountIdx(256)
	_, err = sdb.GetAccount(existingAccount) // check that exist
	require.NoError(t, err)
	_, err = sdb.CreateAccount(common.AccountIdx(256), accounts[1]) // check that can not be created twice
	assert.NotNil(t, err)
	assert.Equal(t, ErrAccountAlreadyExists, common.Unwrap(err))

	_, err = sdb.MTGetAccountProof(common.AccountIdx(256))
	require.NoError(t, err)

	// update accounts
	for i := 0; i < len(accounts); i++ {
		accounts[i].Nonce = accounts[i].Nonce + 1
		existingAccount = accounts[i].Idx
		_, err = sdb.UpdateAccount(existingAccount, accounts[i])
		require.NoError(t, err)
	}

	sdb.Close()
}

func TestVouchInStateDB(t *testing.T) {
	dir, err := os.MkdirTemp("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	sdb, err := NewStateDB(Config{Path: dir, Keep: 128, Type: TypeSynchronizer, NLevels: 32})
	require.NoError(t, err)

	// create test vouches
	var vouches []*common.Vouch
	for i := 0; i < 4; i++ {
		vouches = append(vouches, newVouch(i))
	}

	// get non-existing vouch, expecting an error
	unexistingVouch := common.VouchIdx(001001)
	_, err = sdb.GetVouch(unexistingVouch)
	assert.NotNil(t, err)
	assert.Equal(t, db.ErrNotFound, common.Unwrap(err))

	// add test vouches
	for i := 0; i < len(vouches); i++ {
		_, err = sdb.CreateVouch(vouches[i].Idx, vouches[i])
		require.NoError(t, err)
	}

	for i := 0; i < len(vouches); i++ {
		existingVouch := vouches[i].Idx
		vocGetted, err := sdb.GetVouch(existingVouch)
		require.NoError(t, err)
		assert.Equal(t, vouches[i], vocGetted)
	}

	// try already existing idx and get error
	existingVouch := common.VouchIdx(256257)
	_, err = sdb.GetVouch(existingVouch) // check that exist
	require.NoError(t, err)
	_, err = sdb.CreateVouch(common.VouchIdx(256257), vouches[1]) // check that can not be created twice
	assert.NotNil(t, err)
	assert.Equal(t, ErrAlreadyVouched, common.Unwrap(err))

	_, err = sdb.MTGetVouchProof(common.VouchIdx(256257))
	require.NoError(t, err)

	// update vouches
	for i := 0; i < len(vouches); i++ {
		vouches[i].Value = !vouches[i].Value
		existingVouch = vouches[i].Idx
		_, err = sdb.UpdateVouch(existingVouch, vouches[i])
		require.NoError(t, err)
	}

	sdb.Close()
}

func TestScoreInStateDB(t *testing.T) {
	dir, err := os.MkdirTemp("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	sdb, err := NewStateDB(Config{Path: dir, Keep: 128, Type: TypeSynchronizer, NLevels: 32})
	require.NoError(t, err)

	// create test scores
	var scores []*common.Score
	for i := 0; i < 4; i++ {
		scores = append(scores, newScore(i))
	}

	// get non-existing score, expecting an error
	unexistingScore := common.AccountIdx(1)
	_, err = sdb.GetScore(unexistingScore)
	assert.NotNil(t, err)
	assert.Equal(t, db.ErrNotFound, common.Unwrap(err))

	// add test scores
	for i := 0; i < len(scores); i++ {
		_, err = sdb.CreateScore(scores[i].Idx, scores[i])
		require.NoError(t, err)
	}

	for i := 0; i < len(scores); i++ {
		existingScore := scores[i].Idx
		scoGetted, err := sdb.GetScore(existingScore)
		require.NoError(t, err)
		assert.Equal(t, scores[i], scoGetted)
	}

	// try already existing idx and get error
	existingScore := common.AccountIdx(257)
	_, err = sdb.GetScore(existingScore) // check that exist
	require.NoError(t, err)
	_, err = sdb.CreateScore(common.AccountIdx(257), scores[1]) // check that can not be created twice
	assert.NotNil(t, err)
	assert.Equal(t, ErrScoreAlreadyExists, common.Unwrap(err))

	_, err = sdb.MTGetAccountProof(common.AccountIdx(257))
	require.NoError(t, err)

	// update scores
	for i := 0; i < len(scores); i++ {
		existingScore = scores[i].Idx
		_, err = sdb.UpdateScore(existingScore, scores[i])
		require.NoError(t, err)
	}

	sdb.Close()
}

// // performActions function for Account and Link are to test the db setup and
// // it's mapping with merkel tree
// func performActionsAccount(a *models.Account, s *StateDB) {
// 	proof, err := s.PutAccount(a)
// 	if err != nil {
// 		log.Fatalf("Failed to store key-value pair: %v", err)
// 	}
// 	fmt.Println(proof, "----------------------- Circom Processor Proof ---------------------")

// 	// Retrieve and print a value
// 	value, err := s.GetAccount(a.Idx)
// 	if err != nil {
// 		log.Fatalf("Failed to retrieve value: %v", err)
// 	}
// 	fmt.Printf("Retrieved account: %+v\n", value)

// 	// Get and print root hash for leaf
// 	root := s.GetMTRoot(Account)
// 	fmt.Println(root, "MT root")
// }

// func performActionsLink(l *models.Link, s *StateDB) {
// 	proof, err := s.PutLink(l)
// 	if err != nil {
// 		log.Fatalf("Failed to store key-value pair: %v", err)
// 	}
// 	fmt.Println(proof, "----------------------- Circom Processor Proof ---------------------")
// 	// Retrieve and print a value
// 	value, err := s.GetLink(l.LinkIdx)
// 	if err != nil {
// 		log.Fatalf("Failed to retrieve value: %v", err)
// 	}
// 	fmt.Printf("Retrieved account: %+v\n", value)

// 	// Get and print root hash for leaf
// 	root := s.GetMTRoot(Link)
// 	fmt.Println(root, "MT root")
// }

// func printExamples(s *StateDB) {
// 	// Example accounts
// 	accountA := &models.Account{
// 		Idx:     1,
// 		EthAddr: "0xA",
// 		BJJ:     "ay_value",
// 		Balance: 10,
// 		Score:   1,
// 		Nonce:   0,
// 	}

// 	accountB := &models.Account{
// 		Idx:     2,
// 		EthAddr: "0xB",
// 		BJJ:     "ay_value",
// 		Balance: 10,
// 		Score:   1,
// 		Nonce:   0,
// 	}

// 	accountC := &models.Account{
// 		Idx:     3,
// 		EthAddr: "0xC",
// 		BJJ:     "ay_value",
// 		Balance: 10,
// 		Score:   1,
// 		Nonce:   0,
// 	}

// 	accountD := &models.Account{
// 		Idx:     4,
// 		EthAddr: "0xD",
// 		BJJ:     "ay_value",
// 		Balance: 10,
// 		Score:   1,
// 		Nonce:   0,
// 	}

// 	linkAB := &models.Link{
// 		LinkIdx: 11,
// 		Value:   true,
// 	}

// 	linkAC := &models.Link{
// 		LinkIdx: 13,
// 		Value:   true,
// 	}
// 	linkCD := &models.Link{
// 		LinkIdx: 34,
// 		Value:   true,
// 	}
// 	linkCA := &models.Link{
// 		LinkIdx: 31,
// 		Value:   true,
// 	}
// 	linkCB := &models.Link{
// 		LinkIdx: 32,
// 		Value:   true,
// 	}
// 	// Add Account A
// 	performActionsAccount(accountA, s)

// 	// Add Account B
// 	performActionsAccount(accountB, s)

// 	//Add Account C
// 	performActionsAccount(accountC, s)

// 	//Add Account D
// 	performActionsAccount(accountD, s)

// 	//Add Link AB
// 	performActionsLink(linkAB, s)

// 	performActionsLink(linkAC, s)
// 	performActionsLink(linkCD, s)
// 	performActionsLink(linkCA, s)
// 	performActionsLink(linkCB, s)

// 	// Print Merkle tree root
// 	// fmt.Printf("Merkle Account Tree Root: %s\n", s.AccountTree.Root.Hash)
// }

// func TestInitNewStateDB(t *testing.T) {
// 	dir, err := ioutil.TempDir("", "tmpdb")

// 	// Initialize the StateDB
// 	var stateDB *StateDB
// 	stateDB, err = NewStateDB(Config{Path: dir, Keep: 128, Type: TypeTxSelector, NLevels: 0})
// 	if err != nil {
// 		log.Fatalf("Failed to initialize StateDB: %v", err)
// 	}
// 	defer stateDB.Close()
// 	printExamples(stateDB)
// }
