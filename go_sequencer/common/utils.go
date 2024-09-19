package common

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// EthAddrToBigInt returns a *big.Int from a given ethereum common.Address.
func EthAddrToBigInt(a ethCommon.Address) *big.Int {
	return new(big.Int).SetBytes(a.Bytes())
}
