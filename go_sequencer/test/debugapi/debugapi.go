package debugapi

import (
	"tokamak-sybil-resistance/database/statedb"
	"tokamak-sybil-resistance/synchronizer"
)

// DebugAPI is an http API with debugging endpoints
type DebugAPI struct {
	addr    string
	stateDB *statedb.StateDB // synchronizer statedb
	sync    *synchronizer.Synchronizer
}
