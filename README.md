# Tokamak Sybil Resistence
## Overview
The goal is to create a register for Ethereuem addresses, where each address on the register is assigned a ‘score’ in a sybil-resistent way. This means that although someone can create an unlimited number of Ethereum addresses and add them to the register, the total ‘score’ of the addresses will still be limited. These scores can then be used as a measure of uniqueness. Applications that require protection against sybil attacks can use this register. 

The project uses a mechanism similar to a web-of-trust. Adding an address to the register requires a deposit (which can be retrieved by removing address from the register). Then any two registered address can vouch for each other by both ‘staking’ some of their deposit on each other. This gives the data of a weighted graph, where the vertices are the addresses and the edges are the pairs of addressses that have vouched for each other (the weight of an edge is the amount that was staked by the pair on each other).

From this weighted graph, a score is calculated for each address. The score is a measure of how well connected a particular address is and is robust against manipulation and collusion. The algorithm to compute the score is computationally expensive so it is not possible to run the scoring algorithm in the L1 contract. Instead, the contract requires the updated scores to be submitted to it along with a proof that the score updates are correct. The contract will verify this proof and then update the state. To incentivize people to compute the scores for the contract, it will award a small amount of TON for a correct proof. 

Applications can integrate with this project to obtain sybil-resistence, for example for an airdrop or for a voting application. Users can also use this register in a privacy preserving way by creating a proof that they own an address in the register with a score above a certain threshold.

## Scoring algorithm
The way the scoring algorithm works is by starting with a set (denoted by a calligraphic A) of subsets of the nodes (in practice, these will be the subsets that are smaller than a certain threshold). Then to compute the score for a node, we look at each of these subsets that contains the node and compute the sum of the weights of the links that are leaving this subset and divide it by the number of elements in the subset. The score is then defined to be minimum of these values. More precisely, the scoring function **_f_** is defined by:![eq1](https://github.com/tokamak-network/proof-of-uniqueness/blob/main/imgs/img1.png?raw=true) 

where 

![graph](https://github.com/tokamak-network/proof-of-uniqueness/blob/main/imgs/img3.png?raw=true)

and

![graph](https://github.com/tokamak-network/proof-of-uniqueness/blob/main/imgs/img2.png?raw=true)


### Example:

![graph](https://github.com/tokamak-network/proof-of-uniqueness/blob/main/imgs/graph-example.png?raw=true)


In this example, let us suppose the set of subsets we are using is all subsets with three or less nodes. Then, to compute the score for node 5, we look at all subsets with three or less elements, that contain node 5:

```{5},{0,5},{1,5},{2,5},{3,5},{4,5},{5,6},{0,1,5},{0,2,5},{0,3,5},{0,4,5},{0,6,5},{1,2,5},{1,3,5},{1,4,5},{1,5,6},{2,3,5},{2,4,5},{2,5,6},{3,4,5},{3,5,6},{4,5,6}```

Then for each of these subsets, we compute the sum of the weights of the links leaving the subset. For example, for the subset ```{1,3,5}``` the sum of the links leaving it is ```7+1+1+1+2=12```. Then we divide this by the number of elements in the subset we get ```12/3=4```. If we do this for each subset we get a list of values, and then if we take the minimum of this list of values we get the score for node 5.

### Scoring circuit
This algorithm is implemented in the file [scoring_algorithm.circom](https://github.com/tokamak-network/proof-of-uniqueness/blob/main/scoring_algorithm.circom). It is implemented as a circom circuit so that the scores can be computed off chain and the proof sent to convince the L1 contract that the scores have been computed correctly.


## Usage
To test this code first install circom and snarkjs: https://docs.circom.io/getting-started/installation/.

Our circuit can use both the Groth16 and Plonk zk-SNARK schemes. Groth16 has a smaller proof size compared to Plonk, but it requires a specific setup for each individual circuit. On the other hand, Plonk has a larger proof size but only requires a universal setup that can be used across different circuits without needing a new setup for each one.

Since we use Circom to write our circuits, the constraints are expressed in R1CS format. However, Plonk uses its own constraint format, which results in an increased number of constraints when converting the same circuit into Plonk’s format.

The following example will use a pre-prepared universal setup file. By downloading the setup file provided by [the snarkjs repo](https://github.com/iden3/snarkjs?tab=readme-ov-file#7-prepare-phase-2), you can skip the setup process. In our case, we need a setup with a power of at least 15. Note that as the power increases, the file size grows exponentially, so be cautious when downloading larger setup files.

Our example uses a universal setup file provided by snarkjs with a power of 15(`powersOfTau28_hez_final_15.ptau`).
If you want to start the setup by own, refer to [this section](https://github.com/iden3/snarkjs?tab=readme-ov-file#1-start-a-new-powers-of-tau-ceremony) in snarkjs.

The overall flow is as follows: **circuit compile -> setup -> proof -> verify**


### 1. Cicuit compile

```bash
cd circuits
circom scoring_algorithm.circom --r1cs --wasm
```
The circom command takes one input (the circuit to compile, in our case circuit.circom) and three options:
- r1cs: generates circuit.r1cs (the r1cs constraint system of the circuit in binary format).
- wasm: generates circuit.wasm (the wasm code to generate the witness).

Next, you need to create the input for this circuit. A test input is currently provided: `test_input1.json`.

Once the input is ready, you can use the Javascript/wasm program generated in the previous step to create a witness (values of all the wires) for our input:
```bash
scoring_algorithm_js$ node generate_witness.js scoring_algorithm.wasm ../test_input1.json ../witness.wtns
```
You can verify that the witness was generated correctly with the following command:
```bash
snarkjs wtns check scoring_algorithm.r1cs witness.wtns
```

### 2. Setup

#### Plonk
```bash
snarkjs plonk setup scoring_algorithm.r1cs powersOfTau28_hez_final_15.ptau circuit_final.zkey
```


#### groth16
As mentioned earlier, Groth16 requires an additional setup step for each individual circuit. If you want to proceed with Groth16, follow the snarkjs [groth16 setup documentation](https://github.com/iden3/snarkjs?tab=readme-ov-file#groth16).


### 3. Proof

#### Export the verification key
```bash
snarkjs zkey export verificationkey circuit_final.zkey verification_key.json
```

### Create the proof

#### Plonk
```bash
snarkjs plonk prove circuit_final.zkey witness.wtns proof.json public.json
```


#### Groth16
```bash
snarkjs groth16 prove circuit_final.zkey witness.wtns proof.json public.json
```
Note: The circuit_final.zkey for Groth16 is different from that for Plonk and requires the additional Groth16 setup mentioned above.

### 4. Verify

#### Plonk
```bash
snarkjs plonk verify verification_key.json public.json proof.json
```


#### Groth16
```bash
snarkjs groth16 verify verification_key.json public.json proof.json
```



### Extra: Turn the verifier into a smart contract
```bash
snarkjs zkey export solidityverifier circuit_final.zkey verifier.sol
```




## Contract design

The contract will keep a record of the id, owner, deposit and score for each node, as well as how much each node has staked on each other node. This data will be organized into a Merkle tree with a subtree for each leaf. For this tree will use the MiMC hash function.


![graph](https://github.com/tokamak-network/proof-of-uniqueness/blob/main/imgs/img4.jpg?raw=true)

Each leaf in the main accounts tree stores the ETH address associated with the account, how much TON that account has deposited into the rollup, the “uniqueness” score for that account, and the root hash of a subtree of link data. The link subtree stores the amount that has been staked by that account on a particular other account in the system, (using the account’s id in the account tree). For example if the first account has staked 0.5 TON on the account with id 3, then the third leaf of the subtree for the first leaf of the account tree will store the value 0.5. 

The amounts staked on a particular account are used in the scoring algorithm to compute it’s uniqueness score. To prevent this from being manipulatable, we need to ensure that the total amount that an account can stake on other accounts is limited by how much TON they have deposited into the rollup. Hence we make the rule that a valid instance of the state tree must satisfy

![graph](https://github.com/tokamak-network/proof-of-uniqueness/blob/main/imgs/img5.png?raw=true)

for each leaf of the account tree.

The contract will accept transactions that update this tree, and store them in a queue. Then anyone can create a batch of these transactions, compute the update of the state tree and provide a proof of correctness. At any time, there is only one possibility for the next block, namely the next N txns in the queue (if the size of the queue is less than N, then the contract wont accept a proposed block). This means we dont need any leader selection process, its just the first valid proof from any user which creates the next block. 

## Ownership proofs
Given the current state root, a user can make a proof of ownership of a leaf with score above a certain threshold. A different proof circuit is needed depending on the threshold. Also, any contract that wishes to intergrate with this system and check ownership proofs would need its own verifier contract and also a mapping to record nullifiers to ensure each user can only prove once.
