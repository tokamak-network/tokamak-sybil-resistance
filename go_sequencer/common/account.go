package common

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"tokamak-sybil-resistance/common/nonce"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/poseidon"
	cryptoUtils "github.com/iden3/go-iden3-crypto/utils"
)

// Account is a struct that gives information of the holdings of an address.
// Is the data structure that generates the Value stored in
// the leaf of the MerkleTree
type Account struct {
	Idx      Idx                   `meddler:"idx"`
	BatchNum BatchNum              `meddler:"batch_num"`
	BJJ      babyjub.PublicKeyComp `meddler:"bjj"`
	Sign     Sign                  `meddler:"sign"`
	Ay       *big.Int              `meddler:"ay"` // max of 253 bits used
	EthAddr  ethCommon.Address     `meddler:"eth_addr"`
	Nonce    Nonce                 `meddler:"-"` // max of 40 bits used
	Balance  *big.Int              `meddler:"-"` // max of 192 bits used
}

// Idx represents the account Index in the MerkleTree
type Idx uint64

// Sign represents the 1 bit of baby jubjub
type Sign bool

// Ay represents the ay of baby jubjub of 253 bits
type Ay *big.Int

const (
	// NLeafElems is the number of elements for a leaf
	NLeafElems = 4

	// maxBalanceBytes is the maximum bytes that can use the
	// Account.Balance *big.Int
	maxBalanceBytes = 24

	// IdxBytesLen idx bytes
	IdxBytesLen = 6

	// maxIdxValue is the maximum value that Idx can have (48 bits:
	// maxIdxValue=2**48-1)
	maxIdxValue = 0xffffffffffff

	// UserThreshold determines the threshold from the User Idxs can be
	UserThreshold = 256
	// IdxUserThreshold is a Idx type value that determines the threshold
	// from the User Idxs can be
	IdxUserThreshold = Idx(UserThreshold)
)

var (
	// FFAddr is used to check if an ethereum address is 0xff..ff
	FFAddr = ethCommon.HexToAddress("0xffffffffffffffffffffffffffffffffffffffff")
	// EmptyAddr is used to check if an ethereum address is 0
	EmptyAddr = ethCommon.HexToAddress("0x0000000000000000000000000000000000000000")
)

// Bytes returns a byte array representing the Idx
func (idx Idx) Bytes() ([6]byte, error) {
	if idx > maxIdxValue {
		return [6]byte{}, Wrap(ErrIdxOverflow)
	}
	var idxBytes [8]byte
	binary.BigEndian.PutUint64(idxBytes[:], uint64(idx))
	var b [6]byte
	copy(b[:], idxBytes[2:])
	return b, nil
}

// IdxFromBytes returns Idx from a byte array
func IdxFromBytes(b []byte) (Idx, error) {
	if len(b) != IdxBytesLen {
		return 0, Wrap(fmt.Errorf("can not parse Idx, bytes len %d, expected %d",
			len(b), IdxBytesLen))
	}
	var idxBytes [8]byte
	copy(idxBytes[2:], b[:])
	idx := binary.BigEndian.Uint64(idxBytes[:])
	return Idx(idx), nil
}

// BigInt returns a *big.Int representing the Idx
func (idx Idx) BigInt() *big.Int {
	return big.NewInt(int64(idx))
}

// Bytes returns the bytes representing the Account, in a way that each BigInt
// is represented by 32 bytes, in spite of the BigInt could be represented in
// less bytes (due a small big.Int), so in this way each BigInt is always 32
// bytes and can be automatically parsed from a byte array.
func (a *Account) Bytes() ([32 * NLeafElems]byte, error) {
	var b [32 * NLeafElems]byte

	if a.Nonce > nonce.MaxNonceValue {
		return b, Wrap(fmt.Errorf("%s Nonce", ErrNumOverflow))
	}
	if len(a.Balance.Bytes()) > maxBalanceBytes {
		return b, Wrap(fmt.Errorf("%s Balance", ErrNumOverflow))
	}

	nonceBytes, err := a.Nonce.Bytes()
	if err != nil {
		return b, Wrap(err)
	}

	copy(b[23:28], nonceBytes[:])

	pkSign, pkY := babyjub.UnpackSignY(a.BJJ)
	if pkSign {
		b[22] = 1
	}
	balanceBytes := a.Balance.Bytes()
	copy(b[64-len(balanceBytes):64], balanceBytes)
	// Check if there is possibility of finite field overflow
	ayBytes := pkY.Bytes()
	if len(ayBytes) == 32 { //nolint:gomnd
		ayBytes[0] = ayBytes[0] & 0x3f //nolint:gomnd
		pkY = big.NewInt(0).SetBytes(ayBytes)
	}
	finiteFieldMod, ok := big.NewInt(0).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10) //nolint:gomnd
	if !ok {
		return b, errors.New("error setting bjj finite field")
	}
	pkY = pkY.Mod(pkY, finiteFieldMod)
	ayBytes = pkY.Bytes()
	copy(b[96-len(ayBytes):96], ayBytes)
	copy(b[108:128], a.EthAddr.Bytes())
	return b, nil
}

// HashValue returns the value of the Account, which is the Poseidon hash of its
// *big.Int representation
func (a *Account) HashValue() (*big.Int, error) {
	bi, err := a.BigInts()
	if err != nil {
		return nil, Wrap(err)
	}
	return poseidon.Hash(bi[:])
}

// BigInts returns the [5]*big.Int, where each *big.Int is inside the Finite Field
func (a *Account) BigInts() ([NLeafElems]*big.Int, error) {
	e := [NLeafElems]*big.Int{}

	b, err := a.Bytes()
	if err != nil {
		return e, Wrap(err)
	}

	e[0] = new(big.Int).SetBytes(b[0:32])
	e[1] = new(big.Int).SetBytes(b[32:64])
	e[2] = new(big.Int).SetBytes(b[64:96])
	e[3] = new(big.Int).SetBytes(b[96:128])

	return e, nil
}

// AccountFromBytes returns a Account from a byte array
func AccountFromBytes(b [32 * NLeafElems]byte) (*Account, error) {
	// tokenID, err := TokenIDFromBytes(b[28:32])
	// if err != nil {
	// 	return nil, Wrap(err)
	// }
	var nonceBytes5 [5]byte
	copy(nonceBytes5[:], b[23:28])
	nonce := FromBytes(nonceBytes5)
	sign := b[22] == 1

	balance := new(big.Int).SetBytes(b[40:64])
	// Balance is max of 192 bits (24 bytes)
	if !bytes.Equal(b[32:40], []byte{0, 0, 0, 0, 0, 0, 0, 0}) {
		return nil, Wrap(fmt.Errorf("%s Balance", ErrNumOverflow))
	}
	ay := new(big.Int).SetBytes(b[64:96])
	publicKeyComp := babyjub.PackSignY(sign, ay)
	ethAddr := ethCommon.BytesToAddress(b[108:128])

	if !cryptoUtils.CheckBigIntInField(balance) {
		return nil, Wrap(ErrNotInFF)
	}
	if !cryptoUtils.CheckBigIntInField(ay) {
		return nil, Wrap(ErrNotInFF)
	}

	a := Account{
		Nonce:   nonce,
		Balance: balance,
		BJJ:     publicKeyComp,
		EthAddr: ethAddr,
	}
	return &a, nil
}
