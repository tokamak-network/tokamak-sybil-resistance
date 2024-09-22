package common

import (
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

type TxID [TxIDLen]byte

// Scan implements Scanner for database/sql.
func (txid *TxID) Scan(src interface{}) error {
	srcB, ok := src.([]byte)
	if !ok {
		return Wrap(fmt.Errorf("can't scan %T into TxID", src))
	}
	if len(srcB) != TxIDLen {
		return Wrap(fmt.Errorf("can't scan []byte of len %d into TxID, need %d",
			len(srcB), TxIDLen))
	}
	copy(txid[:], srcB)
	return nil
}

// Value implements valuer for database/sql.
func (txid TxID) Value() (driver.Value, error) {
	return txid[:], nil
}

// String returns a string hexadecimal representation of the TxID
func (txid TxID) String() string {
	return "0x" + hex.EncodeToString(txid[:])
}

// NewTxIDFromString returns a string hexadecimal representation of the TxID
func NewTxIDFromString(idStr string) (TxID, error) {
	txid := TxID{}
	idStr = strings.TrimPrefix(idStr, "0x")
	decoded, err := hex.DecodeString(idStr)
	if err != nil {
		return TxID{}, Wrap(err)
	}
	if len(decoded) != TxIDLen {
		return txid, Wrap(errors.New("Invalid idStr"))
	}
	copy(txid[:], decoded)
	return txid, nil
}

// MarshalText marshals a TxID
func (txid TxID) MarshalText() ([]byte, error) {
	return []byte(txid.String()), nil
}

// UnmarshalText unmarshalls a TxID
func (txid *TxID) UnmarshalText(data []byte) error {
	idStr := string(data)
	id, err := NewTxIDFromString(idStr)
	if err != nil {
		return Wrap(err)
	}
	*txid = id
	return nil
}

type TxType string

const (
	// TxTypeExit represents L2->L1 token transfer.  A leaf for this account appears in the exit
	// tree of the block
	TxTypeExit TxType = "Exit"
	// TxTypeTransfer represents L2->L2 token transfer
	TxTypeTransfer TxType = "Transfer"
	// TxTypeDeposit represents L1->L2 transfer
	TxTypeDeposit TxType = "Deposit"
	// TxTypeCreateAccountDeposit represents creation of a new leaf in the state tree
	// (newAcconut) + L1->L2 transfer
	TxTypeCreateAccountDeposit TxType = "CreateAccountDeposit"
	// TxTypeCreateAccountDepositTransfer represents L1->L2 transfer + L2->L2 transfer
	TxTypeCreateAccountDepositTransfer TxType = "CreateAccountDepositTransfer"
	// TxTypeDepositTransfer TBD
	TxTypeDepositTransfer TxType = "DepositTransfer"
	// TxTypeForceTransfer TBD
	TxTypeForceTransfer TxType = "ForceTransfer"
	// TxTypeForceExit TBD
	TxTypeForceExit TxType = "ForceExit"
	// TxTypeTransferToEthAddr TBD
	TxTypeTransferToEthAddr TxType = "TransferToEthAddr"
	// TxTypeTransferToBJJ TBD
	TxTypeTransferToBJJ TxType = "TransferToBJJ"
	// TxTypeCreateVouch
	TxTypeCreateVouch TxType = "CreateVouch"
	// TxTypeDeleteVouch
	TxTypeDeleteVouch TxType = "DeleteVouch"
)

const (
	// TxIDPrefixL1UserTx is the prefix that determines that the TxID is for
	// a L1UserTx
	//nolinter:gomnd
	TxIDPrefixL1UserTx = byte(0)

	// TxIDPrefixL1CoordTx is the prefix that determines that the TxID is
	// for a L1CoordinatorTx
	//nolinter:gomnd
	TxIDPrefixL1CoordTx = byte(1)

	// TxIDPrefixL2Tx is the prefix that determines that the TxID is for a
	// L2Tx (or PoolL2Tx)
	//nolinter:gomnd
	TxIDPrefixL2Tx = byte(2)

	// TxIDLen is the length of the TxID byte array
	TxIDLen = 33
)

var (
	// SignatureConstantBytes contains the SignatureConstant in byte array
	// format, which is equivalent to 3322668559 as uint32 in byte array in
	// big endian representation.
	SignatureConstantBytes = []byte{198, 11, 230, 15}

	// EmptyTxID is used to check if a TxID is 0
	EmptyTxID = TxID([TxIDLen]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
)
