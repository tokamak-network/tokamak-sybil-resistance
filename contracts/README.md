## Foundry

**Foundry is a blazing fast, portable and modular toolkit for Ethereum application development written in Rust.**

Foundry consists of:

-   **Forge**: Ethereum testing framework (like Truffle, Hardhat and DappTools).
-   **Cast**: Swiss army knife for interacting with EVM smart contracts, sending transactions and getting chain data.
-   **Anvil**: Local Ethereum node, akin to Ganache, Hardhat Network.
-   **Chisel**: Fast, utilitarian, and verbose solidity REPL.

## Documentation

https://book.getfoundry.sh/

## Usage

### Build

```shell
$ forge build
```

### Test

```shell
$ forge test
```

### Format

```shell
$ forge fmt
```

### Gas Snapshots

```shell
$ forge snapshot
```

### Anvil

```shell
$ anvil
```

### Deploy

```shell
$ forge script script/Counter.s.sol:CounterScript --rpc-url <your_rpc_url> --private-key <your_private_key>
```

### Cast

```shell
$ cast <subcommand>
```

### Help

```shell
$ forge --help
$ anvil --help
$ cast --help
```

# Poseidon Contracts (Thanos)
Poseidon2Elements deployed at: 0xb84B26659fBEe08f36A2af5EF73671d66DDf83db
Poseidon3Elements deployed at: 0xFc50367cf2bA87627f99EDD8703FF49252473AED
Poseidon4Elements deployed at: 0xF8AB2781AA06A1c3eF41Bd379Ec1681a70A148e0

# deploy Poseidon
forge script script/DeployPoseidon.s.sol --rpc-url https://rpc.thanos-sepolia.tokamak.network --private-key 1f56dbaabada0d6fce528c32e73890e633553f905cf1d939dfb05b553ff2db4d --broadcast

# deploy Sybil
forge script script/DeployVerifier.s.sol --rpc-url https://rpc.thanos-sepolia.tokamak.network --private-key <your_private_key> --broadcast


forge script script/DeploySybil.s.sol --rpc-url https://rpc.thanos-sepolia.tokamak.network --private-key <your_private_key> --broadcast
