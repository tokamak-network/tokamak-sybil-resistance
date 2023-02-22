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
  mapping(uint256 => bytes32) public filledSubtrees;
  mapping(uint256 => bytes32) public roots;
  uint32 public constant ROOT_HISTORY_SIZE = 30;
  uint32 public currentRootIndex = 0;
  uint32 public nextIndex = 0;

  constructor(uint32 _levels, IHasher _hasher) {
    require(_levels > 0, "_levels should be greater than zero");
    require(_levels < 32, "_levels should be less than 32");
    levels = _levels;
    hasher = _hasher;

    for (uint32 i = 0; i < _levels; i++) {
       filledSubtrees[i] = zeros(i);
    }
    roots[0] = zeros(_levels - 1);
   }

  function hashLeftRight(
    IHasher _hasher,
    bytes32 _left,
    bytes32 _right
  ) public pure returns (bytes32) {
    require(uint256(_left) < FIELD_SIZE, "_left should be inside the field");
    require(uint256(_right) < FIELD_SIZE, "_right should be inside the field");
    uint256 R = uint256(_left);
    uint256 C = 0;
    (R, C) = _hasher.MiMCSponge(R, C);
    R = addmod(R, uint256(_right), FIELD_SIZE);
    (R, C) = _hasher.MiMCSponge(R, C);
    return bytes32(R);
  }
 
 
  function _insert(bytes32 _leaf) internal returns (uint32 index) {
	uint32 _nextIndex = nextIndex;
	require(
	    _nextIndex != uint32(2) ** levels,
	    "Merkle tree is full. No more leaves can be added"
	);
	uint32 currentIndex = _nextIndex;
	bytes32 currentLevelHash = _leaf;
	bytes32 left;
	bytes32 right;

	for (uint32 i = 0; i < levels; i++) {
	    if (currentIndex % 2 == 0) {
		left = currentLevelHash;
		right = zeros(i);
		filledSubtrees[i] = currentLevelHash;
	    } else {
		left = filledSubtrees[i];
		right = currentLevelHash;
	    }
	    currentLevelHash = hashLeftRight(uint256(left), uint256(right));
	    currentIndex /= 2;
  }
	
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
