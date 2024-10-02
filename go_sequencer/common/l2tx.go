package common

import (
	"fmt"
	"math/big"

	ethCrypto "github.com/ethereum/go-ethereum/crypto"
)

// L2Tx is a struct that represents an already forged L2 tx
type L2Tx struct {
	// Stored in DB: mandatory fields
	TxID     TxID       `meddler:"id"`
	BatchNum BatchNum   `meddler:"batch_num"` // batchNum in which this tx was forged.
	Position int        `meddler:"position"`
	FromIdx  AccountIdx `meddler:"from_idx"`
	ToIdx    AccountIdx `meddler:"to_idx"`
	Nonce    Nonce      `meddler:"nonce"`
	Type     TxType     `meddler:"type"`
	Amount   *big.Int   `meddler:"amount,bigint"`
	// EthBlockNum in which this L2Tx was added to the queue
	EthBlockNum int64 `meddler:"eth_block_num"`
}

// NewL2Tx returns the given L2Tx with the TxId & Type parameters calculated
// from the L2Tx values
func NewL2Tx(tx *L2Tx) (*L2Tx, error) {
	txTypeOld := tx.Type
	if err := tx.SetType(); err != nil {
		return nil, Wrap(err)
	}
	// If original Type doesn't match the correct one, return error
	if txTypeOld != "" && txTypeOld != tx.Type {
		return nil, Wrap(fmt.Errorf("L2Tx.Type: %s, should be: %s",
			tx.Type, txTypeOld))
	}

	txIDOld := tx.TxID
	if err := tx.SetID(); err != nil {
		return nil, Wrap(err)
	}
	// If original TxID doesn't match the correct one, return error
	if txIDOld != (TxID{}) && txIDOld != tx.TxID {
		return tx, Wrap(fmt.Errorf("L2Tx.TxID: %s, should be: %s",
			tx.TxID.String(), txIDOld.String()))
	}

	return tx, nil
}

// SetType sets the type of the transaction.  Uses (FromIdx, Nonce).
func (tx *L2Tx) SetType() error {
	if tx.ToIdx == AccountIdx(1) {
		tx.Type = TxTypeExit
	} else if tx.ToIdx < IdxUserThreshold {
		return Wrap(fmt.Errorf(
			"cannot determine type of L2Tx, invalid ToIdx value: %d", tx.ToIdx))
	}
	return nil
}

// SetID sets the ID of the transaction
func (tx *L2Tx) SetID() error {
	txID, err := tx.CalculateTxID()
	if err != nil {
		return err
	}
	tx.TxID = txID
	return nil
}

// CalculateTxID returns the TxID of the transaction. This method is used to
// set the TxID for L2Tx and for PoolL2Tx.
func (tx L2Tx) CalculateTxID() ([TxIDLen]byte, error) {
	var txID TxID
	var b []byte
	// FromIdx
	fromIdxBytes, err := tx.FromIdx.Bytes()
	if err != nil {
		return txID, Wrap(err)
	}
	b = append(b, fromIdxBytes[:]...)
	nonceBytes, err := tx.Nonce.Bytes()
	if err != nil {
		return txID, Wrap(err)
	}
	b = append(b, nonceBytes[:]...)

	// calculate hash
	h := ethCrypto.Keccak256Hash(b).Bytes()

	txID[0] = TxIDPrefixL2Tx
	copy(txID[1:], h)
	return txID, nil
}
