package common

import (
	"encoding/binary"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// Token is a struct that represents an Ethereum token that is supported in Hermez network
type Token struct {
	TokenID TokenID `json:"id" meddler:"token_id"`
	// EthBlockNum indicates the Ethereum block number in which this token was registered
	EthBlockNum int64             `json:"ethereumBlockNum" meddler:"eth_block_num"`
	EthAddr     ethCommon.Address `json:"ethereumAddress" meddler:"eth_addr"`
	Name        string            `json:"name" meddler:"name"`
	Symbol      string            `json:"symbol" meddler:"symbol"`
	Decimals    uint64            `json:"decimals" meddler:"decimals"`
}

// TokenID is the unique identifier of the token, as set in the smart contract
type TokenID int32

// Bytes returns a byte array of length 4 representing the TokenID
func (t TokenID) Bytes() []byte {
	var tokenIDBytes [4]byte
	binary.BigEndian.PutUint32(tokenIDBytes[:], uint32(t))
	return tokenIDBytes[:]
}
