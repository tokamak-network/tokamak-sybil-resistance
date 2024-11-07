package etherscan

import (
	"context"
	"net/http"
	"time"

	"github.com/dghubble/sling"
)

const (
	defaultMaxIdleConns    = 10
	defaultIdleConnTimeout = 2 * time.Second
)

type etherscanResponse struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Result  GasPriceEtherscan `json:"result"`
}

// GasPriceEtherscan definition
type GasPriceEtherscan struct {
	LastBlock       string `json:"LastBlock"`
	SafeGasPrice    string `json:"SafeGasPrice"`
	ProposeGasPrice string `json:"ProposeGasPrice"`
	FastGasPrice    string `json:"FastGasPrice"`
}

// Service definition
type Service struct {
	clientEtherscan *sling.Sling
	apiKey          string
}

// Client is the interface to a ServerProof that calculates zk proofs
type Client interface {
	// Blocking.  Returns the gas price.
	GetGasPrice(ctx context.Context) (*GasPriceEtherscan, error)
}

// NewEtherscanService is the constructor that creates an etherscanService
func NewEtherscanService(etherscanURL string, apikey string) (*Service, error) {
	// Init
	tr := &http.Transport{
		MaxIdleConns:       defaultMaxIdleConns,
		IdleConnTimeout:    defaultIdleConnTimeout,
		DisableCompression: true,
	}
	httpClient := &http.Client{Transport: tr}
	return &Service{
		clientEtherscan: sling.New().Base(etherscanURL).Client(httpClient),
		apiKey:          apikey,
	}, nil
}
