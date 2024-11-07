package common

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
)

type Coordinator struct {
	// Bidder is the address of the bidder
	Bidder ethCommon.Address `meddler:"bidder_addr"`
	// Forger is the address of the forger
	Forger ethCommon.Address `meddler:"forger_addr"`
	// EthBlockNum is the block in which the coordinator was registered
	EthBlockNum int64 `meddler:"eth_block_num"`
	// URL of the coordinators API
	URL string `meddler:"url"`
}
