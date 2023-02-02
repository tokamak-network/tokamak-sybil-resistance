This project is to create a zk-rollup for proof-of-uniqueness. This is a system where users can register an Ethereum address in a Sybil resistant way, i.e. one person can’t register many Ethereum addresses. This register can be used for many things such as doing an airdrop that can’t be cheated and even things like voting systems. With this register, users can also do things like make a proof that they own an Ethereum address that is on the register without revealing which one, which allows things like private voting.


The system allows account to deposit TON and 'stake' part of that TON on different accounts. This then results in the data of a weighted graph. Next an algorithm is run on this weighted graph to compute a score for each account. This algorithm is computationally expensive so is encoded as a circuit and run off chain and then a proof of correctness is sent to the contract. 

The code for this is in the file scoring_
