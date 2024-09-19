package common

import (
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

const (
	// AccountCreationAuthMsg is the message that is signed to authorize a
	// Hermez account creation
	AccountCreationAuthMsg = "Account creation"
	// EIP712Version is the used version of the EIP-712
	EIP712Version = "1"
	// EIP712Provider defines the Provider for the EIP-712
	EIP712Provider = "Hermez Network"
)

var (
	// EmptyEthSignature is an ethereum signature of all zeroes
	EmptyEthSignature = make([]byte, 65)
)

// AccountCreationAuth authorizations sent by users to the L2DB, to be used for
// account creations when necessary
type AccountCreationAuth struct {
	EthAddr   ethCommon.Address     `meddler:"eth_addr"`
	BJJ       babyjub.PublicKeyComp `meddler:"bjj"`
	Signature []byte                `meddler:"signature"`
	Timestamp time.Time             `meddler:"timestamp,utctime"`
}
