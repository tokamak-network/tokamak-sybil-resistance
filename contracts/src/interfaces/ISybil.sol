// SPDX-License-Identifier: MIT
pragma solidity 0.8.23;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";

interface ISybil {
    // L1 Transaction types
    uint8 constant DEPOSIT = 1;
    uint8 constant WITHDRAWAL = 2;
    uint8 constant TRANSFER = 3;

    // Events
    event L1TransactionAdded(
        address indexed sender,
        uint48 indexed fromIdx,
        uint48 toIdx,
        uint32 tokenID,
        uint40 amount,
        uint8 transactionType
    );

    event QueueCleared(uint32 queueIndex, uint16 numTransactionsCleared);
    event BatchForged(
        uint32 indexed batchNum,
        uint48 newLastIdx,
        uint256 newStateRoot,
        uint256 newExitRoot,
        uint8 batchType
    );

    event InitializeSybilVerifierEvent(
        uint8 forgeL1L2BatchTimeout,
        uint256 feeAddToken,
        uint64 withdrawalDelay
    );

    // Initialization function
    function initializeSybilVerifier(
        address[] memory _verifiers,
        uint256[] memory _verifiersParams,
        address _withdrawVerifier,
        address _tokenHEZ,
        uint8 _forgeL1L2BatchTimeout,
        uint256 _feeAddToken,
        address _poseidon2Elements,
        address _poseidon3Elements,
        address _poseidon4Elements,
        address _sybilGovernanceAddress,
        uint64 _withdrawalDelay,
        address _withdrawDelayerContract
    ) external;

    // L1 Transaction functions
    function addL1Transaction(
        uint256 babyPubKey,
        uint48 fromIdx,
        uint40 loadAmountF,
        uint40 amountF,
        uint32 tokenID,
        uint48 toIdx,
        bytes calldata permit,
        uint8 transactionType
    ) external payable;

    function forgeBatch(
        uint48 newLastIdx,
        uint256 newStRoot,
        uint256 newExitRoot,
        bytes calldata encodedL1CoordinatorTx,
        bytes calldata l1L2TxsData,
        bytes calldata feeIdxCoordinator,
        uint8 verifierIdx,
        bool l1Batch,
        uint256[2] calldata proofA,
        uint256[2][2] calldata proofB,
        uint256[2] calldata proofC
    ) external;

    // Governance functions
    function setForgeL1L2BatchTimeout(uint8 newTimeout) external;
    function setFeeAddToken(uint256 newFee) external;

    // Merkle proof functions
    function setMerkleProof(bytes32 proof) external;
    function getMerkleProof() external view returns (bytes32);

    // Getter functions
    function getStateRoot(uint32 batchNum) external view returns (uint256);
    function getLastForgedBatch() external view returns (uint32);
    function getUniquenessScore(address account) external view returns (uint256);

    // L1 Transaction Queue functions
    function getL1TransactionQueue(uint32 queueIndex) external view returns (bytes memory);
    function getQueueLength() external view returns (uint32);
}
