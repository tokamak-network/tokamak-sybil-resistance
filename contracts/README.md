## Foundry

**Foundry is a blazing fast, portable and modular toolkit for Ethereum application development written in Rust.**

Foundry consists of:

- **Forge**: Ethereum testing framework (like Truffle, Hardhat and DappTools).
- **Cast**: Swiss army knife for interacting with EVM smart contracts, sending transactions and getting chain data.
- **Anvil**: Local Ethereum node, akin to Ganache, Hardhat Network.
- **Chisel**: Fast, utilitarian, and verbose solidity REPL.

## Documentation

https://book.getfoundry.sh/

## Usage

### Node setuup

```shell
npm i
```

### Build

```shell
forge build
```

### Test

```shell
forge test
```

### Format

```shell
forge fmt
```

### Gas Snapshots

```shell
forge snapshot
```

### Anvil

```shell
anvil
```

### Deploy

```shell
forge script script/Counter.s.sol:CounterScript --rpc-url <your_rpc_url> --private-key <your_private_key>
```

### Cast

```shell
$ cast <subcommand>
```

### Help

```shell
forge --help
anvil --help
cast --help
```

# Poseidon Contracts (Thanos)

Poseidon2Elements deployed at: 0x31c3EBCa9c9eFAeE59FD30A968BCA0634F42Ed95
Poseidon3Elements deployed at: 0x82c5d2d227b5C6f69A978cfA7025654517e82351
Poseidon4Elements deployed at: 0xfFe18609E5641527191408BfC5776129037794f2

# Verifier Contract (Thanos)
0x734fDa76b5BfaB6793E2bb69F20834aCDda510C7

# Sybil Contract (Thanos)
0x3b5008b4bB4489ddD36627d41E55D351D3fAa5dE

# Deploy Poseidon

```shell
forge script script/DeployPoseidon.s.sol --broadcast --ffi
```

# Deploy Sybil

```shell
forge script script/DeployVerifier.s.sol --rpc-url https://rpc.thanos-sepolia.tokamak.network --private-key <your_private_key> --broadcast --legacy
```

```shell
forge script script/DeploySybil.s.sol --rpc-url https://rpc.thanos-sepolia.tokamak.network --private-key <your_private_key> --broadcast --legacy
```

# Verify Contract

# VerifierRollupStub

```shell
forge create --rpc-url https://rpc.thanos-sepolia.tokamak.network --legacy --private-key <your_private_key> src/stub/VerifierRollupStub.sol:VerifierRollupStub --verify --verifier blockscout --verifier-url "https://explorer.thanos-sepolia.tokamak.network/api?module=contract&action=verify_via_sourcify&addressHash=0x734fDa76b5BfaB6793E2bb69F20834aCDda510C7"
```

# Sybil
```shell
forge create --rpc-url https://rpc.thanos-sepolia.tokamak.network --legacy --private-key <your_private_key> src/sybil.sol:Sybil --verify --verifier blockscout --verifier-url "https://explorer.thanos-sepolia.tokamak.network/api?module=contract&action=verify_via_sourcify&addressHash=0x3b5008b4bB4489ddD36627d41E55D351D3fAa5dE"
```

# Run Queue Simulation Test

```shell
node simulations/SybilQueue.test.js
```
