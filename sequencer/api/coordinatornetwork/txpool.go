package coordinatornetwork

import (
	"context"
	"tokamak-sybil-resistance/common"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// pubSubTxsPool represents a subscription to a single PubSub topic. Messages
// can be published to the topic with pubSubTxsPool.Publish, and received
// messages are pushed to the Messages channel.
type pubSubTxsPool struct {
	ctx     context.Context
	ps      *pubsub.PubSub
	topic   *pubsub.Topic
	sub     *pubsub.Subscription
	self    peer.ID
	handler func(common.PoolL2Tx) error
}
