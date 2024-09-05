/*
Package node does the initialization of all the required objects to either run
as a synchronizer or as a coordinator.

The Node contains several goroutines that run in the background or that
periodically perform tasks.  One of this goroutines periodically calls the
`Synchronizer.Sync` function, allowing the synchronization of one block at a
time.  After every call to `Synchronizer.Sync`, the Node sends a message to the
Coordinator to notify it about the new synced block (and associated state) or
reorg (and resetted state) in case one happens.

Other goroutines perform tasks such as: updating the token prices, update
metrics stored in the historyDB, update recommended fee stored in the
historyDB, run the http API server, run the debug http API server, etc.
*/
package node

import (
	"context"
	"sync"
	"time"
	"tokamak-sybil-resistance/api"
	"tokamak-sybil-resistance/api/stateapiupdater"
	"tokamak-sybil-resistance/config"
	"tokamak-sybil-resistance/coordinator"
	"tokamak-sybil-resistance/database/historydb"
	"tokamak-sybil-resistance/synchronizer"
	"tokamak-sybil-resistance/test/debugapi"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Mode sets the working mode of the node (synchronizer or coordinator)
type Mode string

const (
	// ModeCoordinator defines the mode of the HermezNode as Coordinator, which
	// means that the node is set to forge (which also will be synchronizing with
	// the L1 blockchain state)
	ModeCoordinator Mode = "coordinator"

	// ModeSynchronizer defines the mode of the HermezNode as Synchronizer, which
	// means that the node is set to only synchronize with the L1 blockchain state
	// and will not forge
	ModeSynchronizer Mode = "synchronizer"
)

// Node is the Hermez Node
type Node struct {
	nodeAPI         *NodeAPI
	stateAPIUpdater *stateapiupdater.Updater
	debugAPI        *debugapi.DebugAPI
	// Coordinator
	coord *coordinator.Coordinator

	// Synchronizer
	sync *synchronizer.Synchronizer

	// General
	cfg          *config.Node
	mode         Mode
	sqlConnRead  *sqlx.DB
	sqlConnWrite *sqlx.DB
	historyDB    *historydb.HistoryDB
	ctx          context.Context
	wg           sync.WaitGroup
	cancel       context.CancelFunc
}

// APIServer is a server that only runs the API
type APIServer struct {
	nodeAPI *NodeAPI
	mode    Mode
	ctx     context.Context
	wg      sync.WaitGroup
	cancel  context.CancelFunc
}

// NodeAPI holds the node http API
type NodeAPI struct { //nolint:golint
	api                                     *api.API
	engine                                  *gin.Engine
	addr                                    string
	coordinatorNetwork                      bool
	coordinatorNetworkFindMorePeersInterval time.Duration
	readtimeout                             time.Duration
	writetimeout                            time.Duration
}
