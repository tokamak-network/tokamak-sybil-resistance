package statedb

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/models"

	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/poseidon"
	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db"
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

// CreateAccount creates a new Account in the StateDB for the given Idx.  If
// StateDB.MT==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func (s *StateDB) CreateAccount(idx common.Idx, account *common.Account) (
	*merkletree.CircomProcessorProof, error) {
	cpp, err := CreateAccountInTreeDB(s.db.DB(), s.AccountTree, idx, account)
	if err != nil {
		return cpp, common.Wrap(err)
	}
	// store idx by EthAddr & BJJ
	err = s.setIdxByEthAddrBJJ(idx, account.EthAddr, account.BJJ)
	return cpp, common.Wrap(err)
}

// CreateAccountInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  Creates a new Account in the StateDB for the given Idx.  If
// StateDB.MT==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func CreateAccountInTreeDB(sto db.Storage, mt *merkletree.MerkleTree, idx common.Idx,
	account *common.Account) (*merkletree.CircomProcessorProof, error) {
	// store at the DB the key: v, and value: leaf.Bytes()
	v, err := account.HashValue()
	if err != nil {
		return nil, common.Wrap(err)
	}
	accountBytes, err := account.Bytes()
	if err != nil {
		return nil, common.Wrap(err)
	}

	// store the Leaf value
	tx, err := sto.NewTx()
	if err != nil {
		return nil, common.Wrap(err)
	}

	idxBytes, err := idx.Bytes()
	if err != nil {
		return nil, common.Wrap(err)
	}
	_, err = tx.Get(append(PrefixKeyIdx, idxBytes[:]...))
	if common.Unwrap(err) != db.ErrNotFound {
		return nil, common.Wrap(ErrAccountAlreadyExists)
	}

	err = tx.Put(append(PrefixKeyAccHash, v.Bytes()...), accountBytes[:])
	if err != nil {
		return nil, common.Wrap(err)
	}
	err = tx.Put(append(PrefixKeyIdx, idxBytes[:]...), v.Bytes())
	if err != nil {
		return nil, common.Wrap(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, common.Wrap(err)
	}

	if mt != nil {
		return mt.AddAndGetCircomProof(idx.BigInt(), v)
	}

	return nil, nil
}

// MTGetProof returns the CircomVerifierProof for a given Idx
func (s *StateDB) MTGetProof(idx common.Idx) (*merkletree.CircomVerifierProof, error) {
	if s.AccountTree == nil {
		return nil, common.Wrap(ErrStateDBWithoutMT)
	}
	p, err := s.AccountTree.GenerateSCVerifierProof(idx.BigInt(), s.AccountTree.Root())
	if err != nil {
		return nil, common.Wrap(err)
	}
	return p, nil
}

// Put stores an account in the database and updates the Merkle tree.
func (sdb *StateDB) PutAccount(a *models.Account) (*merkletree.CircomProcessorProof, error) {
	var idxBytes [2]byte

	apHash, _ := PoseidonHashAccount(a)
	fmt.Println(apHash, "---------------  Poseidon Hash Account ---------------")
	accountBytes := BytesAccount(a)

	binary.LittleEndian.PutUint16(idxBytes[:], uint16(a.Idx))

	tx, err := sdb.db.NewTx()
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

// GetAccount returns the account for the given Idx
func (s *StateDB) GetAccount(idx common.Idx) (*common.Account, error) {
	return GetAccountInTreeDB(s.db.DB(), idx)
}

// GetAccountInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  GetAccount returns the account for the given Idx
func GetAccountInTreeDB(sto db.Storage, idx common.Idx) (*common.Account, error) {
	idxBytes, err := idx.Bytes()
	if err != nil {
		return nil, common.Wrap(err)
	}
	vBytes, err := sto.Get(append(PrefixKeyIdx, idxBytes[:]...))
	if err != nil {
		return nil, common.Wrap(err)
	}
	accBytes, err := sto.Get(append(PrefixKeyAccHash, vBytes...))
	if err != nil {
		return nil, common.Wrap(err)
	}
	var b [32 * common.NLeafElems]byte
	copy(b[:], accBytes)
	account, err := common.AccountFromBytes(b)
	if err != nil {
		return nil, common.Wrap(err)
	}
	account.Idx = idx
	return account, nil
}

// UpdateAccount updates the Account in the StateDB for the given Idx.  If
// StateDB.mt==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func (s *StateDB) UpdateAccount(idx common.Idx, account *common.Account) (
	*merkletree.CircomProcessorProof, error) {
	return UpdateAccountInTreeDB(s.db.DB(), s.AccountTree, idx, account)
}

// UpdateAccountInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  Updates the Account in the StateDB for the given Idx.  If
// StateDB.mt==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func UpdateAccountInTreeDB(sto db.Storage, mt *merkletree.MerkleTree, idx common.Idx,
	account *common.Account) (*merkletree.CircomProcessorProof, error) {
	// store at the DB the key: v, and value: account.Bytes()
	v, err := account.HashValue()
	if err != nil {
		return nil, common.Wrap(err)
	}
	accountBytes, err := account.Bytes()
	if err != nil {
		return nil, common.Wrap(err)
	}

	tx, err := sto.NewTx()
	if err != nil {
		return nil, common.Wrap(err)
	}
	err = tx.Put(append(PrefixKeyAccHash, v.Bytes()...), accountBytes[:])
	if err != nil {
		return nil, common.Wrap(err)
	}
	idxBytes, err := idx.Bytes()
	if err != nil {
		return nil, common.Wrap(err)
	}
	err = tx.Put(append(PrefixKeyIdx, idxBytes[:]...), v.Bytes())
	if err != nil {
		return nil, common.Wrap(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, common.Wrap(err)
	}

	if mt != nil {
		proof, err := mt.Update(idx.BigInt(), v)
		return proof, common.Wrap(err)
	}
	return nil, nil
}
