package coordinator

import (
	"context"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/coordinator/prover"
	"tokamak-sybil-resistance/log"
)

// ProversPool contains the multiple prover clients
type ProversPool struct {
	pool chan prover.Client
}

// NewProversPool creates a new pool of provers.
func NewProversPool(maxServerProofs int) *ProversPool {
	return &ProversPool{
		pool: make(chan prover.Client, maxServerProofs),
	}
}

// Add a prover to the pool
func (p *ProversPool) Add(ctx context.Context, serverProof prover.Client) {
	select {
	case p.pool <- serverProof:
	case <-ctx.Done():
	}
}

// Get returns the next available prover
func (p *ProversPool) Get(ctx context.Context) (prover.Client, error) {
	select {
	case <-ctx.Done():
		log.Info("ServerProofPool.Get done")
		return nil, common.Wrap(common.ErrDone)
	case serverProof := <-p.pool:
		return serverProof, nil
	}
}
