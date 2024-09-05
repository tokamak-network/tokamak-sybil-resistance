package api

import (
	"math/big"
	"tokamak-sybil-resistance/common"
)

type rollupConstants struct {
	PublicConstants         common.RollupConstants `json:"publicConstants"`
	MaxFeeIdxCoordinator    int                    `json:"maxFeeIdxCoordinator"`
	ReservedIdx             int                    `json:"reservedIdx"`
	ExitIdx                 int                    `json:"exitIdx"`
	LimitDepositAmount      *big.Int               `json:"limitDepositAmount"`
	LimitL2TransferAmount   *big.Int               `json:"limitL2TransferAmount"`
	LimitTokens             int                    `json:"limitTokens"`
	L1CoordinatorTotalBytes int                    `json:"l1CoordinatorTotalBytes"`
	L1UserTotalBytes        int                    `json:"l1UserTotalBytes"`
	MaxL1UserTx             int                    `json:"maxL1UserTx"`
	MaxL1Tx                 int                    `json:"maxL1Tx"`
	InputSHAConstantBytes   int                    `json:"inputSHAConstantBytes"`
	NumBuckets              int                    `json:"numBuckets"`
	// MaxWithdrawalDelay      int                    `json:"maxWithdrawalDelay"`
	ExchangeMultiplier int `json:"exchangeMultiplier"`
}

type configAPI struct {
	ChainID         uint16          `json:"chainId"`
	RollupConstants rollupConstants `json:"hermez"`
}
