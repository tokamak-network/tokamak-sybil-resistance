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

// PoolL2Tx returns the data structure of PoolL2Tx with the parameters of a
// L2Tx filled
func (tx L2Tx) PoolL2Tx() *PoolL2Tx {
	return &PoolL2Tx{
		TxID:    tx.TxID,
		FromIdx: tx.FromIdx,
		ToIdx:   tx.ToIdx,
		Amount:  tx.Amount,
		Nonce:   tx.Nonce,
		Type:    tx.Type,
	}
}

// L2TxsToPoolL2Txs returns an array of []*PoolL2Tx from an array of []*L2Tx,
// where the PoolL2Tx only have the parameters of a L2Tx filled.
func L2TxsToPoolL2Txs(txs []L2Tx) []PoolL2Tx {
	var r []PoolL2Tx
	for _, tx := range txs {
		r = append(r, *tx.PoolL2Tx())
	}
	return r
}

// L2TxFromBytesDataAvailability decodes a L2Tx from []byte (Data Availability)
func L2TxFromBytesDataAvailability(b []byte, nLevels int) (*L2Tx, error) {
	idxLen := nLevels / 8 //nolint:gomnd
	tx := &L2Tx{}
	var err error

	var paddedFromIdxBytes [3]byte
	copy(paddedFromIdxBytes[3-idxLen:], b[0:idxLen])
	tx.FromIdx, err = AccountIdxFromBytes(paddedFromIdxBytes[:])
	if err != nil {
		return nil, Wrap(err)
	}

	var paddedToIdxBytes [3]byte
	copy(paddedToIdxBytes[3-idxLen:3], b[idxLen:idxLen*2])
	tx.ToIdx, err = AccountIdxFromBytes(paddedToIdxBytes[:])
	if err != nil {
		return nil, Wrap(err)
	}

	tx.Amount, err = Float40FromBytes(b[idxLen*2 : idxLen*2+Float40BytesLength]).BigInt()
	if err != nil {
		return nil, Wrap(err)
	}
	// tx.Fee = FeeSelector(b[idxLen*2+Float40BytesLength])
	return tx, nil
}
