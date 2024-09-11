/*
Package stateapiupdater is responsible for generating and storing the object response of the GET /state endpoint exposed through the api package.
This object is extensively defined at the OpenAPI spec located at api/swagger.yml.

Deployment considerations: in a setup where multiple processes are used (dedicated api process, separated coord / sync, ...), only one process should care
of using this package.
*/
package stateapiupdater

import (
	"fmt"
	"sync"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/database/historydb"
	"tokamak-sybil-resistance/log"

	"github.com/hermeznetwork/tracerr"
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

func (rfp *RecommendedFeePolicy) valid() bool {
	switch rfp.PolicyType {
	case RecommendedFeePolicyTypeStatic:
		if rfp.StaticValue == 0 {
			log.Warn("RecommendedFee is set to 0 USD, and the policy is static")
		}
		return true
	case RecommendedFeePolicyTypeAvgLastHour:
		return true
	case RecommendedFeePolicyTypeDynamicFee:
		return true
	default:
		return false
	}
}

// SetSCVars sets the smart contract vars (ony updates those that are not nil)
func (u *Updater) SetSCVars(vars *common.SCVariablesPtr) {
	u.rw.Lock()
	defer u.rw.Unlock()
	if vars.Rollup != nil {
		u.vars.Rollup = vars.Rollup
		rollupVars := historydb.NewRollupVariablesAPI(u.vars.Rollup)
		u.state.Rollup = *rollupVars
	}
	// if vars.Auction != nil {
	// 	u.vars.Auction = vars.Auction
	// 	auctionVars := historydb.NewAuctionVariablesAPI(u.vars.Auction)
	// 	u.state.Auction = *auctionVars
	// }
	// if vars.WDelayer != nil {
	// 	u.vars.WDelayer = vars.WDelayer
	// 	u.state.WithdrawalDelayer = *u.vars.WDelayer
	// }
}

// NewUpdater creates a new Updater
func NewUpdater(hdb *historydb.HistoryDB, config *historydb.NodeConfig, vars *common.SCVariables,
	consts *historydb.Constants, rfp *RecommendedFeePolicy, maxTxPerBatch int64) (*Updater, error) {
	if ok := rfp.valid(); !ok {
		return nil, tracerr.Wrap(fmt.Errorf("Invalid recommended fee policy: %v", rfp.PolicyType))
	}
	u := Updater{
		hdb:    hdb,
		config: *config,
		consts: *consts,
		state: historydb.StateAPI{
			NodePublicInfo: historydb.NodePublicInfo{
				ForgeDelay: config.ForgeDelay,
			},
		},
		rfp:           rfp,
		maxTxPerBatch: maxTxPerBatch,
	}
	u.SetSCVars(vars.AsPtr())
	return &u, nil
}
