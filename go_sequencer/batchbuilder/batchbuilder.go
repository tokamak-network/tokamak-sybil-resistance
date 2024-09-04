package batchbuilder

import (
	"tokamak-sybil-resistance/database/statedb"
	"tokamak-sybil-resistance/txprocessor"
)

// ConfigCircuit contains the circuit configuration
type ConfigCircuit struct {
	TxsMax       uint64
	L1TxsMax     uint64
	SMTLevelsMax uint64
}

// BatchBuilder implements the batch builder type, which contains the
// functionalities
type BatchBuilder struct {
	localStateDB *statedb.LocalStateDB
}

// ConfigBatch contains the batch configuration
type ConfigBatch struct {
	TxProcessorConfig txprocessor.Config
}
