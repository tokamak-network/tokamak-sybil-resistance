package apitypes

import "tokamak-sybil-resistance/common"

type BigIntStr string

// CollectedFeesAPI is send common.batch.CollectedFee through the API
type CollectedFeesAPI map[common.TokenID]BigIntStr
