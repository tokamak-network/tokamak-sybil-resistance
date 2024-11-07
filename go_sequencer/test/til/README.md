# Til (Test instructions language)
Language to define sets of instructions to simulate Tokamak transactions (L1 & L2) with real data.

## Syntax
### Global
Set type definition

- Blockchain: generate the transactions that would come from the Tokamak smart contract on the blockchain.
	
```
Type: Blockchain
```

- PoolL2: generate the transactions that would come from the Pool of L2Txs
```
Type: PoolL2
```

### Blockchain set of instructions
Available instructions:
```go
Type: Blockchain

// deposit on the account of the user A, of an amount of 50 units
CreateAccountDeposit A: 50

// deposit at the account A, of 6 units
Deposit A: 6

// exit from the account A of 5 units. Transaction will be a L2Tx
Exit A: 5

// force-exit from the account A of 5 units.
// Transaction will be L1UserTx of ForceExit type
ForceExit A: 5

// create vouch from account A to B. Transaction will be a L2Tx
CreateVouch A-B

// delete vouch from account A to B. Transaction will be a L2Tx
DeleteVouch A-B

// advance one batch, forging without L1UserTxs, only can contain L2Txs and
// L1CoordinatorTxs
> batch

// advance one batch, forging with L1UserTxs (and L2Txs and L1CoordinatorTxs)
> batchL1

// advance an ethereum block
> block
```

### PoolL2 set of instructions
Available instructions:
```go
Type: PoolL2

// exit from the account A of 3 units
PoolExit A: 3

// create vouch from account A to B
CreateVouch A-B

// delete vouch from account A to B
DeleteVouch A-B
```

## Usage
```go
// create a new til.Context
tc := til.NewContext(eth.RollupConstMaxL1UserTx)

// generate Blockchain blocks data from the common.SetBlockcahin0 instructions set
blocks, err = tc.GenerateBlocks(common.SetBlockchainMinimumFlow0)
assert.Nil(t, err)

// generate PoolL2 transactions data from the common.SetPool0 instructions set
poolL2Txs, err = tc.GeneratePoolL2Txs(common.SetPoolL2MinimumFlow0)
assert.Nil(t, err)
```

Where `blocks` will contain:
```go
// BatchData contains the information of a Batch
type BatchData struct {
    L2Txs                   []common.L2Tx
    CreatedAccounts         []common.Account
}

// BlockData contains the information of a Block
type BlockData struct {
    L1UserTxs           []common.L1Tx
    Batches             []BatchData
}
```

## Tests
```bash
cp .env.example .env
brew install go-task # for running various tasks, especially tests
task test-til
```