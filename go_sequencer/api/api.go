package api

import (
	"tokamak-sybil-resistance/api/coordinatornetwork"
	"tokamak-sybil-resistance/database/historydb"
	"tokamak-sybil-resistance/database/l2db"
	"tokamak-sybil-resistance/database/statedb"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/go-playground/validator"
)

// API serves HTTP requests to allow external interaction with the Hermez node
type API struct {
	historyDB     *historydb.HistoryDB
	config        *configAPI
	l2DB          *l2db.L2DB
	stateDB       *statedb.StateDB
	hermezAddress ethCommon.Address
	validate      *validator.Validate
	coordnet      *coordinatornetwork.CoordinatorNetwork
}
