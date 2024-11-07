/*
Package coordinatornetwork implements a comunication layer among coordinators
in order to share information such as transactions in the pool and create account authorizations.

To do so the pubsub gossip protocol is used.
This code is currently heavily based on this example: https://github.com/libp2p/go-libp2p/blob/master/examples/pubsub
*/

package coordinatornetwork

import (
	"context"

	"github.com/libp2p/go-libp2p-core/host"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

// CoordinatorNetwork it's a p2p communication layer that enables coordinators to exchange information
// in benefit of the network and them selfs. The main goal is to share L2 data (common.PoolL2Tx and common.AccountCreationAuth)
type CoordinatorNetwork struct {
	self                host.Host
	dht                 *dht.IpfsDHT
	ctx                 context.Context
	discovery           *discovery.RoutingDiscovery
	txsPool             pubSubTxsPool
	discoveryServiceTag string
}
