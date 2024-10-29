// Package common zk.go contains all the common data structures used at the
// hermez-node, zk.go contains the zkSnark inputs used to generate the proof
package common

import (
	"math/big"

	"github.com/iden3/go-merkletree"
)

// ZKMetadata contains ZKInputs metadata that is not used directly in the
// ZKInputs result, but to calculate values for Hash check
type ZKMetadataCommon struct {
	// Circuit parameters
	// absolute maximum of L1 or L2 transactions allowed
	MaxLevels uint32
	// merkle tree depth
	NLevels uint32
	// absolute maximum of L1 transaction allowed
	MaxL1Tx uint32
	// total txs allowed
	MaxTx uint32

	L1TxsData             [][]byte
	L1TxsDataAvailability [][]byte
	L2TxsData             [][]byte
	ChainID               uint16

	NewStateRootRaw *merkletree.Hash
	NewExitRootRaw  *merkletree.Hash
}

type ZKMetadataAccount struct {
	ZKMetadataCommon
	NewLastIdxRaw AccountIdx
}

type ZKMetadataVouch struct {
	ZKMetadataCommon
	NewLastIdxRaw VouchIdx
}

// ZKInputs represents the inputs that will be used to generate the zkSNARK
// proof
type ZKInputs struct {
	MetadataAccount ZKMetadataAccount `json:"-"`
	MetadataVouch   ZKMetadataVouch   `json:"-"`

	//
	// General
	//

	// CurrentNumBatch is the current batch number processed
	CurrentNumBatch *big.Int `json:"currentNumBatch"` // uint32
	// inputs for final `hashGlobalInputs`
	// OldLastIdx is the last index assigned to an account
	OldLastIdxAccount *big.Int `json:"oldLastIdxAccount"` // uint64 (max nLevels bits)
	// OldStateRoot is the current state merkle tree root
	OldStateRootAccount *big.Int `json:"oldStateRoot"` // Hash

	//Vouches
	OldLastIdxVouch *big.Int `json:"oldLastIdxVouch"` // uint64 (max nLevels bits)
	// OldStateRoot is the current state merkle tree root
	OldStateRootVouch *big.Int `json:"oldStateRootVouch"` // Hash

	// GlobalChainID is the blockchain ID (0 for Ethereum mainnet). This
	// value can be get from the smart contract.
	GlobalChainID *big.Int `json:"globalChainID"` // uint16
	//
	// Txs (L1&L2)
	//

	// transaction L1-L2
	// TxCompressedData
	TxCompressedData []*big.Int `json:"txCompressedData"` // big.Int (max 251 bits), len: [maxTx]
	// TxCompressedDataV2, only used in L2Txs, in L1Txs is set to 0
	TxCompressedDataV2 []*big.Int `json:"txCompressedDataV2"` // big.Int (max 193 bits), len: [maxTx]
	// MaxNumBatch is the maximum allowed batch number when the transaction
	// can be processed
	MaxNumBatch []*big.Int `json:"maxNumBatch"` // big.Int (max 32 bits), len: [maxTx]

	// FromIdx
	FromIdx []*big.Int `json:"fromIdx"` // uint64 (max nLevels bits), len: [maxTx]
	// AuxFromAccountIdx is the AccountIdx of the new created account which is
	// consequence of a L1CreateAccountTx
	AuxFromAccountIdx []*big.Int `json:"auxFromIdx"` // uint64 (max nLevels bits), len: [maxTx]

	// ToIdx
	ToIdx []*big.Int `json:"toIdx"` // uint64 (max nLevels bits), len: [maxTx]
	// AuxToIdx is the Idx of the Tx that has 'toIdx==0', is the
	// coordinator who will find which Idx corresponds to the 'toBJJAy' or
	// 'toEthAddr'
	AuxToIdx []*big.Int `json:"auxToIdx"` // uint64 (max nLevels bits), len: [maxTx]
	// ToBJJAy
	ToBJJAy []*big.Int `json:"toBjjAy"` // big.Int, len: [maxTx]
	// ToEthAddr
	ToEthAddr []*big.Int `json:"toEthAddr"` // ethCommon.Address, len: [maxTx]
	// AmountF encoded as float40
	AmountF []*big.Int `json:"amountF"` // uint40 len: [maxTx]

	// OnChain determines if is L1 (1/true) or L2 (0/false)
	OnChain []*big.Int `json:"onChain"` // bool, len: [maxTx]

	//
	// Txs/L1Txs
	//
	// NewAccount boolean (0/1) flag set 'true' when L1 tx creates a new
	// account (fromIdx==0)
	NewAccount []*big.Int `json:"newAccount"` // bool, len: [maxTx]
	// DepositAmountF encoded as float40
	DepositAmountF []*big.Int `json:"loadAmountF"` // uint40, len: [maxTx]
	// FromEthAddr
	FromEthAddr []*big.Int `json:"fromEthAddr"` // ethCommon.Address, len: [maxTx]
	// FromBJJCompressed boolean encoded where each value is a *big.Int
	FromBJJCompressed [][256]*big.Int `json:"fromBjjCompressed"` // bool array, len: [maxTx][256]

	//
	// Txs/L2Txs
	//
	// transaction L2 signature
	// S
	S []*big.Int `json:"s"` // big.Int, len: [maxTx]
	// R8x
	R8x []*big.Int `json:"r8x"` // big.Int, len: [maxTx]
	// R8y
	R8y []*big.Int `json:"r8y"` // big.Int, len: [maxTx]

	//
	// State MerkleTree Leafs transitions
	//

	// state 1, value of the sender (from) account leaf. The values at the
	// moment pre-smtprocessor of the update (before updating the Sender
	// leaf).
	// TokenID1  []*big.Int   `json:"tokenID1"`  // uint32, len: [maxTx]
	NonceA1    []*big.Int   `json:"nonce1"`    // uint64 (max 40 bits), len: [maxTx]
	SignA1     []*big.Int   `json:"sign1"`     // bool, len: [maxTx]
	AyA1       []*big.Int   `json:"ay1"`       // big.Int, len: [maxTx]
	BalanceA1  []*big.Int   `json:"balance1"`  // big.Int (max 192 bits), len: [maxTx]
	EthAddrA1  []*big.Int   `json:"ethAddr1"`  // ethCommon.Address, len: [maxTx]
	SiblingsA1 [][]*big.Int `json:"siblings1"` // big.Int, len: [maxTx][nLevels + 1]
	// Required for inserts and deletes, values of the CircomProcessorProof
	// (smt insert proof)
	IsOld0_A1  []*big.Int `json:"isOld0_1"`  // bool, len: [maxTx]
	OldKeyA1   []*big.Int `json:"oldKey1"`   // uint64 (max 40 bits), len: [maxTx]
	OldValueA1 []*big.Int `json:"oldValue1"` // Hash, len: [maxTx]

	// state 2, value of the receiver (to) account leaf. The values at the
	// moment pre-smtprocessor of the update (before updating the Receiver
	// leaf).
	// If Tx is an Exit (tx.ToIdx=1), state 2 is used for the Exit Merkle
	// Proof of the Exit MerkleTree.
	// TokenID2  []*big.Int   `json:"tokenID2"`  // uint32, len: [maxTx]
	Nonce2    []*big.Int   `json:"nonce2"`    // uint64 (max 40 bits), len: [maxTx]
	Sign2     []*big.Int   `json:"sign2"`     // bool, len: [maxTx]
	Ay2       []*big.Int   `json:"ay2"`       // big.Int, len: [maxTx]
	Balance2  []*big.Int   `json:"balance2"`  // big.Int (max 192 bits), len: [maxTx]
	EthAddr2  []*big.Int   `json:"ethAddr2"`  // ethCommon.Address, len: [maxTx]
	Siblings2 [][]*big.Int `json:"siblings2"` // big.Int, len: [maxTx][nLevels + 1]
	// NewExit determines if an exit transaction has to create a new leaf
	// in the exit tree. If already exists an exit leaf of an account in
	// the ExitTree, there is no 'new leaf' creation and 'NewExit' for that
	// tx is 0 (if is an 'insert' in the tree, NewExit=1, if is an 'update'
	// of an existing leaf, NewExit=0).
	NewExit []*big.Int `json:"newExit"` // bool, len: [maxTx]
	// Required for inserts and deletes, values of the CircomProcessorProof
	// (smt insert proof)
	IsOld0_2  []*big.Int `json:"isOld0_2"`  // bool, len: [maxTx]
	OldKey2   []*big.Int `json:"oldKey2"`   // uint64 (max 40 bits), len: [maxTx]
	OldValue2 []*big.Int `json:"oldValue2"` // Hash, len: [maxTx]

	//
	// Intermediate States
	//

	// Intermediate States to parallelize witness computation
	// Note: the Intermediate States (IS) of the last transaction does not
	// exist. Meaning that transaction 3 (4th) will fill the parameters
	// FromIdx[3] and ISOnChain[3], but last transaction (maxTx-1) will fill
	// FromIdx[maxTx-1] but will not fill ISOnChain. That's why IS have
	// length of maxTx-1, while the other parameters have length of maxTx.
	// Last transaction does not need intermediate state since its output
	// will not be used.

	// decode-tx
	// ISOnChain indicates if tx is L1 (true (1)) or L2 (false (0))
	ISOnChain []*big.Int `json:"imOnChain"` // bool, len: [maxTx - 1]
	// ISOutIdx current index account for each Tx
	// Contains the index of the created account in case that the tx is of
	// account creation type.
	ISOutIdxAccount []*big.Int `json:"imOutIdx"` // uint64 (max nLevels bits), len: [maxTx - 1]
	// rollup-tx
	// ISStateRoot root at the moment of the Tx (once processed), the state
	// root value once the Tx is processed into the state tree
	ISStateRootAccount []*big.Int `json:"imStateRoot"` // Hash, len: [maxTx - 1]
	// ISExitTree root at the moment (once processed) of the Tx the value
	// once the Tx is processed into the exit tree
	ISExitRoot []*big.Int `json:"imExitRoot"` // Hash, len: [maxTx - 1]
}

// func bigIntsToStrings(v interface{}) interface{} {
// 	switch c := v.(type) {
// 	case *big.Int:
// 		return c.String()
// 	case []*big.Int:
// 		r := make([]interface{}, len(c))
// 		for i := range c {
// 			r[i] = bigIntsToStrings(c[i])
// 		}
// 		return r
// 	case [256]*big.Int:
// 		r := make([]interface{}, len(c))
// 		for i := range c {
// 			r[i] = bigIntsToStrings(c[i])
// 		}
// 		return r
// 	case [][]*big.Int:
// 		r := make([]interface{}, len(c))
// 		for i := range c {
// 			r[i] = bigIntsToStrings(c[i])
// 		}
// 		return r
// 	case [][256]*big.Int:
// 		r := make([]interface{}, len(c))
// 		for i := range c {
// 			r[i] = bigIntsToStrings(c[i])
// 		}
// 		return r
// 	case map[string]interface{}:
// 		// avoid printing a warning when there is a struct type
// 	default:
// 		log.Warnf("bigIntsToStrings unexpected type: %T\n", v)
// 	}
// 	return nil
// }

// // MarshalJSON implements the json marshaler for ZKInputs
// func (z ZKInputs) MarshalJSON() ([]byte, error) {
// 	var m map[string]interface{}
// 	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
// 		TagName: "json",
// 		Result:  &m,
// 	})
// 	if err != nil {
// 		return nil, tracerr.Wrap(err)
// 	}
// 	err = dec.Decode(z)
// 	if err != nil {
// 		return nil, tracerr.Wrap(err)
// 	}

// 	for k, v := range m {
// 		m[k] = bigIntsToStrings(v)
// 	}
// 	return json.Marshal(m)
// }

// NewZKInputs returns a pointer to an initialized struct of ZKInputs
func NewZKInputs(chainID uint16, maxTx, maxL1Tx, maxFeeIdxs, nLevels uint32,
	currentNumBatch *big.Int) *ZKInputs {
	zki := &ZKInputs{}
	zki.MetadataAccount.MaxLevels = uint32(48) //nolint:gomnd
	zki.MetadataAccount.NLevels = nLevels
	zki.MetadataAccount.MaxL1Tx = maxL1Tx
	zki.MetadataAccount.MaxTx = maxTx
	zki.MetadataAccount.ChainID = chainID

	// General
	zki.CurrentNumBatch = currentNumBatch
	zki.OldLastIdxAccount = big.NewInt(0)
	zki.OldStateRootAccount = big.NewInt(0)

	zki.OldLastIdxVouch = big.NewInt(0)
	zki.OldStateRootVouch = big.NewInt(0)

	zki.GlobalChainID = big.NewInt(int64(chainID))

	// Txs
	zki.TxCompressedData = newSlice(maxTx)
	zki.TxCompressedDataV2 = newSlice(maxTx)
	zki.MaxNumBatch = newSlice(maxTx)
	zki.FromIdx = newSlice(maxTx)
	zki.AuxFromAccountIdx = newSlice(maxTx)
	zki.ToIdx = newSlice(maxTx)
	zki.AuxToIdx = newSlice(maxTx)
	zki.ToBJJAy = newSlice(maxTx)
	zki.ToEthAddr = newSlice(maxTx)
	zki.AmountF = newSlice(maxTx)
	zki.OnChain = newSlice(maxTx)
	zki.NewAccount = newSlice(maxTx)

	// L1
	zki.DepositAmountF = newSlice(maxTx)
	zki.FromEthAddr = newSlice(maxTx)
	zki.FromBJJCompressed = make([][256]*big.Int, maxTx)
	for i := 0; i < len(zki.FromBJJCompressed); i++ {
		// zki.FromBJJCompressed[i] = newSlice(256)
		for j := 0; j < 256; j++ {
			zki.FromBJJCompressed[i][j] = big.NewInt(0)
		}
	}

	// L2
	zki.S = newSlice(maxTx)
	zki.R8x = newSlice(maxTx)
	zki.R8y = newSlice(maxTx)

	// State MerkleTree Leafs transitions
	// zki.TokenID1 = newSlice(maxTx)
	zki.NonceA1 = newSlice(maxTx)
	zki.SignA1 = newSlice(maxTx)
	zki.AyA1 = newSlice(maxTx)
	zki.BalanceA1 = newSlice(maxTx)
	zki.EthAddrA1 = newSlice(maxTx)
	zki.SiblingsA1 = make([][]*big.Int, maxTx)
	for i := 0; i < len(zki.SiblingsA1); i++ {
		zki.SiblingsA1[i] = newSlice(nLevels + 1)
	}
	zki.IsOld0_A1 = newSlice(maxTx)
	zki.OldKeyA1 = newSlice(maxTx)
	zki.OldValueA1 = newSlice(maxTx)

	// zki.TokenID2 = newSlice(maxTx)
	zki.Nonce2 = newSlice(maxTx)
	zki.Sign2 = newSlice(maxTx)
	zki.Ay2 = newSlice(maxTx)
	zki.Balance2 = newSlice(maxTx)
	zki.EthAddr2 = newSlice(maxTx)
	zki.Siblings2 = make([][]*big.Int, maxTx)
	for i := 0; i < len(zki.Siblings2); i++ {
		zki.Siblings2[i] = newSlice(nLevels + 1)
	}
	zki.NewExit = newSlice(maxTx)
	zki.IsOld0_2 = newSlice(maxTx)
	zki.OldKey2 = newSlice(maxTx)
	zki.OldValue2 = newSlice(maxTx)

	// Intermediate States
	zki.ISOnChain = newSlice(maxTx - 1)
	zki.ISOutIdxAccount = newSlice(maxTx - 1)
	zki.ISStateRootAccount = newSlice(maxTx - 1)
	zki.ISExitRoot = newSlice(maxTx - 1)

	return zki
}

// newSlice returns a []*big.Int slice of length n with values initialized at
// 0.
// Is used to initialize all *big.Ints of the ZKInputs data structure, so when
// the transactions are processed and the ZKInputs filled, there is no need to
// set all the elements, and if a transaction does not use a parameter, can be
// leaved as it is in the ZKInputs, as will be 0, so later when using the
// ZKInputs to generate the zkSnark proof there is no 'nil'/'null' values.
func newSlice(n uint32) []*big.Int {
	s := make([]*big.Int, n)
	for i := 0; i < len(s); i++ {
		s[i] = big.NewInt(0)
	}
	return s
}

// // HashGlobalData returns the HashGlobalData
// func (z ZKInputs) HashGlobalData() (*big.Int, error) {
// 	b, err := z.ToHashGlobalData()
// 	if err != nil {
// 		return nil, tracerr.Wrap(err)
// 	}

// 	h := sha256.New()
// 	_, err = h.Write(b)
// 	if err != nil {
// 		return nil, tracerr.Wrap(err)
// 	}

// 	r := new(big.Int).SetBytes(h.Sum(nil))
// 	v := r.Mod(r, cryptoConstants.Q)

// 	return v, nil
// }

// // ToHashGlobalData returns the data to be hashed in the method HashGlobalData
// func (z ZKInputs) ToHashGlobalData() ([]byte, error) {
// 	var b []byte
// 	bytesMaxLevels := int(z.Metadata.MaxLevels / 8) //nolint:gomnd
// 	bytesNLevels := int(z.Metadata.NLevels / 8)     //nolint:gomnd

// 	// [MAX_NLEVELS bits] oldLastIdx
// 	oldLastIdx := make([]byte, bytesMaxLevels)
// 	oldLastIdxBytes := z.OldLastIdx.Bytes()
// 	copy(oldLastIdx[len(oldLastIdx)-len(oldLastIdxBytes):], oldLastIdxBytes)
// 	b = append(b, oldLastIdx...)

// 	// [MAX_NLEVELS bits] newLastIdx
// 	newLastIdx := make([]byte, bytesMaxLevels)
// 	newLastIdxBytes, err := z.Metadata.NewLastIdxRaw.Bytes()
// 	if err != nil {
// 		return nil, tracerr.Wrap(err)
// 	}
// 	copy(newLastIdx, newLastIdxBytes[len(newLastIdxBytes)-bytesMaxLevels:])
// 	b = append(b, newLastIdx...)

// 	// [256 bits] oldStRoot
// 	oldStateRoot := make([]byte, 32)
// 	copy(oldStateRoot, z.OldStateRoot.Bytes())
// 	b = append(b, oldStateRoot...)

// 	// [256 bits] newStateRoot
// 	newStateRoot := make([]byte, 32)
// 	copy(newStateRoot, z.Metadata.NewStateRootRaw.Bytes())
// 	b = append(b, newStateRoot...)

// 	// [256 bits] newExitRoot
// 	newExitRoot := make([]byte, 32)
// 	copy(newExitRoot, z.Metadata.NewExitRootRaw.Bytes())
// 	b = append(b, newExitRoot...)

// 	// [MAX_L1_TX * (2 * MAX_NLEVELS + 528) bits] L1TxsData
// 	l1TxDataLen := (2*z.Metadata.MaxLevels + 528) //nolint:gomnd
// 	l1TxsDataLen := (z.Metadata.MaxL1Tx * l1TxDataLen)
// 	l1TxsData := make([]byte, l1TxsDataLen/8) //nolint:gomnd
// 	for i := 0; i < len(z.Metadata.L1TxsData); i++ {
// 		dataLen := int(l1TxDataLen) / 8 //nolint:gomnd
// 		pos0 := i * dataLen
// 		pos1 := i*dataLen + dataLen
// 		copy(l1TxsData[pos0:pos1], z.Metadata.L1TxsData[i])
// 	}
// 	b = append(b, l1TxsData...)

// 	var l1TxsDataAvailability []byte
// 	for i := 0; i < len(z.Metadata.L1TxsDataAvailability); i++ {
// 		l1TxsDataAvailability = append(l1TxsDataAvailability, z.Metadata.L1TxsDataAvailability[i]...)
// 	}
// 	b = append(b, l1TxsDataAvailability...)

// 	// [MAX_TX*(2*NLevels + 48) bits] L2TxsData
// 	var l2TxsData []byte
// 	l2TxDataLen := 2*z.Metadata.NLevels + 48 //nolint:gomnd
// 	l2TxsDataLen := (z.Metadata.MaxTx * l2TxDataLen)
// 	expectedL2TxsDataLen := l2TxsDataLen / 8 //nolint:gomnd
// 	for i := 0; i < len(z.Metadata.L2TxsData); i++ {
// 		l2TxsData = append(l2TxsData, z.Metadata.L2TxsData[i]...)
// 	}
// 	if len(l2TxsData) > int(expectedL2TxsDataLen) {
// 		return nil, tracerr.Wrap(fmt.Errorf("len(l2TxsData): %d, expected: %d",
// 			len(l2TxsData), expectedL2TxsDataLen))
// 	}

// 	b = append(b, l2TxsData...)
// 	l2TxsPadding := make([]byte,
// 		(int(z.Metadata.MaxTx)-len(z.Metadata.L1TxsDataAvailability)-
// 			len(z.Metadata.L2TxsData))*int(l2TxDataLen)/8) //nolint:gomnd
// 	b = append(b, l2TxsPadding...)

// 	// [NLevels * MAX_TOKENS_FEE bits] feeTxsData
// 	for i := 0; i < len(z.FeeIdxs); i++ {
// 		feeIdx := make([]byte, bytesNLevels) //nolint:gomnd
// 		feeIdxBytes := z.FeeIdxs[i].Bytes()
// 		copy(feeIdx[len(feeIdx)-len(feeIdxBytes):], feeIdxBytes[:])
// 		b = append(b, feeIdx...)
// 	}

// 	// [16 bits] chainID
// 	var chainID [2]byte
// 	binary.BigEndian.PutUint16(chainID[:], z.Metadata.ChainID)
// 	b = append(b, chainID[:]...)

// 	// [32 bits] currentNumBatch
// 	currNumBatchBytes := z.CurrentNumBatch.Bytes()
// 	var currNumBatch [4]byte
// 	copy(currNumBatch[4-len(currNumBatchBytes):], currNumBatchBytes)
// 	b = append(b, currNumBatch[:]...)

// 	return b, nil
// }
