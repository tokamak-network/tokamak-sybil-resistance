package prover

import (
	"context"
	"math/big"
	"time"
	"tokamak-sybil-resistance/common"

	"github.com/dghubble/sling"
)

// Proof TBD this type will be received from the proof server
type Proof struct {
	PiA      [3]*big.Int    `json:"pi_a"`
	PiB      [3][2]*big.Int `json:"pi_b"`
	PiC      [3]*big.Int    `json:"pi_c"`
	Protocol string         `json:"protocol"`
}

type bigInt big.Int

// PublicInputs are the public inputs of the proof
type PublicInputs []*big.Int

// Client is the interface to a ServerProof that calculates zk proofs
type Client interface {
	// Non-blocking
	CalculateProof(ctx context.Context, zkInputs *common.ZKInputs) error
	// Blocking.  Returns the Proof and Public Data (public inputs)
	GetProof(ctx context.Context) (*Proof, []*big.Int, error)
	// Non-Blocking
	Cancel(ctx context.Context) error
	// Blocking
	WaitReady(ctx context.Context) error
}

// StatusCode is the status string of the ProofServer
type StatusCode string

const (
	// StatusCodeAborted means prover is ready to take new proof. Previous
	// proof was aborted.
	StatusCodeAborted StatusCode = "aborted"
	// StatusCodeBusy means prover is busy computing proof.
	StatusCodeBusy StatusCode = "busy"
	// StatusCodeFailed means prover is ready to take new proof. Previous
	// proof failed
	StatusCodeFailed StatusCode = "failed"
	// StatusCodeSuccess means prover is ready to take new proof. Previous
	// proof succeeded
	StatusCodeSuccess StatusCode = "success"
	// StatusCodeUnverified means prover is ready to take new proof.
	// Previous proof was unverified
	StatusCodeUnverified StatusCode = "unverified"
	// StatusCodeUninitialized means prover is not initialized
	StatusCodeUninitialized StatusCode = "uninitialized"
	// StatusCodeUndefined means prover is in an undefined state. Most
	// likely is booting up. Keep trying
	StatusCodeUndefined StatusCode = "undefined"
	// StatusCodeInitializing means prover is initializing and not ready yet
	StatusCodeInitializing StatusCode = "initializing"
	// StatusCodeReady means prover initialized and ready to do first proof
	StatusCodeReady StatusCode = "ready"
)

// Status is the return struct for the status API endpoint
type Status struct {
	Status  StatusCode `json:"status"`
	Proof   string     `json:"proof"`
	PubData string     `json:"pubData"`
}

// ErrorServer is the return struct for an API error
type ErrorServer struct {
	Status  StatusCode `json:"status"`
	Message string     `json:"msg"`
}

type apiMethod string

const (
	// GET is an HTTP GET
	GET apiMethod = "GET"
	// POST is an HTTP POST with maybe JSON body
	POST apiMethod = "POST"
)

// ProofServerClient contains the data related to a ProofServerClient
type ProofServerClient struct {
	URL          string
	client       *sling.Sling
	pollInterval time.Duration
}

// MockClient is a mock ServerProof to be used in tests.  It doesn't calculate anything
type MockClient struct {
	counter int64
	Delay   time.Duration
}
