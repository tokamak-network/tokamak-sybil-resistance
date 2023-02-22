pragma solidity ^0.8.13;

contract L1Contract {
  struct Node {
		uint32 id;       // id of node in merkle tree
		address owner;   // eth address that own node
		uint deposit;    // amount of TON deposited into node
		uint score;      // uniqueness score for node  
		bytes32 link_tree_hash;  // root hash for subtree of links made by this node
  }

  mapping (Node => mapping (Node => uint256)) stakes;
  struct PendingTx {
		uint8 optype;
		bytes20 txDataHash;
		uint64 expiryBlock;
  }

  mapping (uint64 => PendingTx) internal transactionQueue;

  uint64 public firstUnprocessedTx;

  uint64 public totalUnprocessedTxs;
  
  mapping(bytes32 => bool) public nullifierHashes;
  
  function ProveBlock(bytes scores, bytes proofData) external {
		if (totalUnprocessedTxs < N) {
				//reject
		}

		Verifier.checkProof(scores, proofData) {
				//check proof using verifier contract
		}		
  }

  function forgeBatch {
  }

  function addL1Transaction {
  }

  function withdrawMerkleProof{
  }
}
