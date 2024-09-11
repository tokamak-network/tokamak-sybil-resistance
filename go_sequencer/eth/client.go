package eth

// ClientInterface is the eth Client interface used by hermez-node modules to
// interact with Ethereum Blockchain and smart contracts.
type ClientInterface interface {
	EthereumInterface
	RollupInterface
}

const (
	blocksPerDay = (3600 * 24) / 15 //nolint:gomnd
)
