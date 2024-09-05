package coordinator

import "tokamak-sybil-resistance/coordinator/prover"

// ProversPool contains the multiple prover clients
type ProversPool struct {
	pool chan prover.Client
}
