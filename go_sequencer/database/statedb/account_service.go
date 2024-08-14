package statedb

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"tokamak-sybil-resistance/models"

	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/poseidon"
	"github.com/iden3/go-merkletree"
)

func stringToByte(s string, numByte int) []byte {
	b := make([]byte, numByte)
	copy(b[:], s)
	return b
}

func BytesAccount(a *models.Account) [91]byte {
	var b [91]byte
	// Convert Idx into [2]byte
	binary.LittleEndian.PutUint16(b[0:2], uint16(a.Idx))

	// Convert EthAddr into [20]byte
	addBytes := stringToByte(a.EthAddr, 20)
	copy(b[2:22], addBytes)

	//Unpack BJJ
	// Conver pky into [32]byte and store pkSign
	pkSign, pkY := babyjub.UnpackSignY([32]byte(stringToByte(a.BJJ, 32)))
	if pkSign {
		b[22] = 1
	} else {
		b[22] = 0
	}

	copy(b[23:57], pkY.Bytes())

	// Convert balance into [24]byte
	binary.LittleEndian.PutUint64(b[57:79], uint64(a.Balance))

	// Convert score into [4]byte
	binary.LittleEndian.PutUint32(b[79:83], uint32(a.Score))

	// Convert nounce into [8]byte
	binary.LittleEndian.PutUint64(b[83:91], uint64(a.Nonce))

	return b
}

func AccountFromBytes(b [91]byte) models.Account {
	var a models.Account

	// Extract Idx from [0:2]
	a.Idx = int(binary.LittleEndian.Uint16(b[0:2]))

	// Extract EthAddr from [2:22]
	a.EthAddr = string(b[2:22])

	// Extract BJJ Sign and pkY from [22:57]
	pkSign := b[22] == 1
	pkY := new(big.Int).SetBytes(b[23:57])
	bjj := babyjub.PackSignY(pkSign, pkY)
	a.BJJ = string(bjj[:])
	// Extract Balance from [57:79]
	a.Balance = int(binary.LittleEndian.Uint64(b[57:79]))

	// Extract Score from [79:83]
	a.Score = int(binary.LittleEndian.Uint32(b[79:83]))

	// Extract Nonce from [83:91]
	a.Nonce = int(binary.LittleEndian.Uint64(b[83:91]))

	return a
}

// Calculates poseidonHash for zk-Snark rollup Proof
func PoseidonHashAccount(a *models.Account) (*big.Int, error) {
	bigInt := make([]*big.Int, 3)
	b := BytesAccount(a)

	bigInt[0] = new(big.Int).SetBytes(b[0:32])
	bigInt[1] = new(big.Int).SetBytes(b[32:64])
	bigInt[2] = new(big.Int).SetBytes(b[64:91])

	return poseidon.Hash(bigInt)
}

// Put stores an account in the database and updates the Merkle tree.
func (sdb *StateDB) PutAccount(a *models.Account) (*merkletree.CircomProcessorProof, error) {
	var idxBytes [2]byte

	apHash, _ := PoseidonHashAccount(a)
	fmt.Println(apHash, "---------------  Poseidon Hash Account ---------------")
	accountBytes := BytesAccount(a)

	binary.LittleEndian.PutUint16(idxBytes[:], uint16(a.Idx))

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
	var idxBytes [2]byte
	// Convert Idx into [2]byte
	binary.LittleEndian.PutUint16(idxBytes[0:2], uint16(idx))

	accountHashBytes, err := sdb.DB.Get(append(PrefixKeyAccountIdx, idxBytes[:]...))
	if err != nil {
		return nil, err
	}

	accountBytes, err := sdb.DB.Get(append(PrefixKeyAccHash, accountHashBytes...))
	if err != nil {
		return nil, err
	}

	account := AccountFromBytes([91]byte(accountBytes))

	return &account, nil
}
