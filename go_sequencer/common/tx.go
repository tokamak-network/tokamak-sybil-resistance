package common

import "encoding/hex"

type TxID [TxIDLen]byte

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

// String returns a string hexadecimal representation of the TxID
func (txid TxID) String() string {
	return "0x" + hex.EncodeToString(txid[:])
}