package common

import (
	"fmt"
	"math/big"
)

// FeeFactorLsh60 is the feeFactor << 60
var FeeFactorLsh60 [256]*big.Int

// FeeSelector is used to select a percentage from the FeePlan.
type FeeSelector uint8

// RecommendedFee is the recommended fee to pay in USD per transaction set by
// the coordinator according to the tx type (if the tx requires to create an
// account and register, only register or he account already esists)
type RecommendedFee struct {
	ExistingAccount        float64 `json:"existingAccount"`
	CreatesAccount         float64 `json:"createAccount"`
	CreatesAccountInternal float64 `json:"createAccountInternal"`
}

// MaxFeePlan is the maximum value of the FeePlan
const MaxFeePlan = 256

// CalcFeeAmount calculates the fee amount in tokens from an amount and
// feeSelector (fee index).
func CalcFeeAmount(amount *big.Int, feeSel FeeSelector) (*big.Int, error) {
	feeAmount := new(big.Int).Mul(amount, FeeFactorLsh60[int(feeSel)])
	if feeSel < 192 { //nolint:gomnd
		feeAmount.Rsh(feeAmount, 60)
	}
	if feeAmount.BitLen() > 128 { //nolint:gomnd
		return nil, Wrap(fmt.Errorf("FeeAmount overflow (feeAmount doesn't fit in 128 bits)"))
	}
	return feeAmount, nil
}
