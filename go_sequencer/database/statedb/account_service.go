package statedb

import (
	"encoding/json"
	"fmt"
	"math/big"
	"tokamak-sybil-resistance/models"

	"github.com/iden3/go-iden3-crypto/poseidon"
	"github.com/iden3/go-merkletree"
)

// Calculates poseidonHash for zk-Snark rollup Proof
func PoseidonHashAccount(a *models.Account) (*big.Int, error) {
	var bigInts []*big.Int

	// Convert Idx
	idx := big.NewInt(int64(a.Idx))
	bigInts = append(bigInts, idx)

	// Convert EthAddr
	if ethAddr, ok := new(big.Int).SetString(a.EthAddr, 16); ok {
		bigInts = append(bigInts, ethAddr)
	}

	// Convert Sign
	sign := big.NewInt(0)
	if a.Sign {
		sign = big.NewInt(1)
	}
	bigInts = append(bigInts, sign)

	// Convert Ay
	if ay, ok := new(big.Int).SetString(a.Ay, 16); ok {
		bigInts = append(bigInts, ay)
	}

	// Convert Balance
	balance := big.NewInt(int64(a.Balance))
	bigInts = append(bigInts, balance)

	// Convert Score
	score := big.NewInt(int64(a.Score))
	bigInts = append(bigInts, score)

	// Convert Nonce
	nonce := big.NewInt(int64(a.Nonce))
	bigInts = append(bigInts, nonce)

	return poseidon.Hash(bigInts)
}

// Put stores an account in the database and updates the Merkle tree.
func (sdb *StateDB) PutAccount(a *models.Account) (*merkletree.CircomProcessorProof, error) {
	apHash, _ := PoseidonHashAccount(a)
	fmt.Println(apHash, "---------------  Poseidon Hash Account ---------------")
	accountBytes, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	idxBytes, err := json.Marshal(a.Idx)
	if err != nil {
		return nil, err
	}
	tx, err := sdb.DB.NewTx()
	if err != nil {
		return nil, err
	}

	err = tx.Put(append(PrefixKeyAccHash, apHash.Bytes()...), accountBytes[:])
	if err != nil {
		return nil, err
	}
	err = tx.Put(append(PrefixKeyAccountIdx, idxBytes[:]...), apHash.Bytes())
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Update the Merkle tree and return a CircomProcessorProof if the Merkle tree is not nil
	if sdb.AccountTree != nil {
		return sdb.AccountTree.AddAndGetCircomProof(BigInt(a.Idx), apHash)
	}
	return nil, nil
}

// Get retrieves an account for a given idx from the database.
func (sdb *StateDB) GetAccount(idx int) (*models.Account, error) {
	idxBytes, err := json.Marshal(idx)
	if err != nil {
		return nil, err
	}

	accountHashBytes, err := sdb.DB.Get(append(PrefixKeyAccountIdx, idxBytes[:]...))
	if err != nil {
		return nil, err
	}

	accountBytes, err := sdb.DB.Get(append(PrefixKeyAccHash, accountHashBytes...))
	if err != nil {
		return nil, err
	}

	var account models.Account
	err = json.Unmarshal(accountBytes, &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}
