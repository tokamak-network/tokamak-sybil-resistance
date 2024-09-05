/*
Package stateapiupdater is responsible for generating and storing the object response of the GET /state endpoint exposed through the api package.
This object is extensively defined at the OpenAPI spec located at api/swagger.yml.

Deployment considerations: in a setup where multiple processes are used (dedicated api process, separated coord / sync, ...), only one process should care
of using this package.
*/
package stateapiupdater

import (
	"sync"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/database/historydb"
)

// Updater is an utility object to facilitate updating the StateAPI
type Updater struct {
	hdb           *historydb.HistoryDB
	state         historydb.StateAPI
	config        historydb.NodeConfig
	vars          common.SCVariablesPtr
	consts        historydb.Constants
	rw            sync.RWMutex
	rfp           *RecommendedFeePolicy
	maxTxPerBatch int64
}

// RecommendedFeePolicy describes how the recommended fee is calculated
type RecommendedFeePolicy struct {
	PolicyType      RecommendedFeePolicyType `validate:"required" env:"HEZNODE_RECOMMENDEDFEEPOLICY_POLICYTYPE"`
	StaticValue     float64                  `env:"HEZNODE_RECOMMENDEDFEEPOLICY_STATICVALUE"`
	BreakThreshold  int                      `env:"HEZNODE_RECOMMENDEDFEEPOLICY_BREAKTHRESHOLD"`
	NumLastBatchAvg int                      `env:"HEZNODE_RECOMMENDEDFEEPOLICY_NUMLASTBATCHAVG"`
}

// RecommendedFeePolicyType describes the different available recommended fee strategies
type RecommendedFeePolicyType string

const (
	// RecommendedFeePolicyTypeStatic always give the same StaticValue as recommended fee
	RecommendedFeePolicyTypeStatic RecommendedFeePolicyType = "Static"
	// RecommendedFeePolicyTypeAvgLastHour set the recommended fee using the average fee of the last hour
	RecommendedFeePolicyTypeAvgLastHour RecommendedFeePolicyType = "AvgLastHour"
	// RecommendedFeePolicyTypeDynamicFee set the recommended fee taking in account the gas used in L1,
	// the gasPrice and the ether price in the last batches
	RecommendedFeePolicyTypeDynamicFee RecommendedFeePolicyType = "DynamicFee"
)
