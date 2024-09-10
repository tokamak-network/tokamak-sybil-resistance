// SPDX-License-Identifier: MIT
pragma solidity 0.8.23;

interface ISybil {
    // Custom Errors
    error InvalidCreateAccountTransaction();
    error InvalidDepositTransaction();
    error InvalidForceExitTransaction();
    error InvalidForceExplodeTransaction();
    error InternalTxNotAllowed();
    error BatchTimeoutExceeded();

    // Initialization function
    function initialize(uint8 _forgeL1L2BatchTimeout) external;

    // L1 Transaction functions
    function addL1Transaction(
        uint256 babyPubKey,
        uint48 fromIdx,
        uint40 loadAmountF,
        uint40 amountF,
        uint48 toIdx
    ) external payable;

    // Batch forging function
    function forgeBatch(
        uint48 newLastIdx,
        uint256 newStRoot,
        uint256 newVouchRoot,
        uint256 newScoreRoot,
        uint256 newExitRoot,
        bool l1Batch
    ) external;

    // Governance function
    function setForgeL1L2BatchTimeout(uint8 newTimeout) external;

    // Getter functions
    function getStateRoot(uint32 batchNum) external view returns (uint256);
    function getLastForgedBatch() external view returns (uint32);

    // L1 Transaction Queue functions
    function getL1TransactionQueue(
        uint32 queueIndex
    ) external view returns (bytes memory);
    function getQueueLength() external view returns (uint32);
}
