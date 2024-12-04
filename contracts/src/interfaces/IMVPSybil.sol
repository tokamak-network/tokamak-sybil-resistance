// SPDX-License-Identifier: MIT
pragma solidity 0.8.23;

interface IMVPSybil {
    error InvalidVerifierAddress();
    error InvalidCreateAccountTransaction();
    error InvalidDepositTransaction();
    error InvalidForceExitTransaction();
    error InvalidForceExplodeTransaction();
    error InternalTxNotAllowed();
    error BatchTimeoutExceeded();
    error LoadAmountExceedsLimit();
    error LoadAmountDoesNotMatch();
    error AmountExceedsLimit();
    error InvalidTransactionParameters();
    error WithdrawAlreadyDone();
    error SmtProofInvalid();
    error EthTransferFailed();
    error InvalidProof();



    // Initialization function
    function initialize(
        address[] memory verifiers,
        uint256[] memory maxTxs,
        uint256[] memory nLevels,
        uint8 _forgeL1L2BatchTimeout, 
        address _poseidon2Elements,
        address _poseidon3Elements,
        address _poseidon4Elements
    ) external;

    // L1 Transaction functions
    function _addTx(
        address ethAddress,
        uint48 fromIdx,
        uint40 loadAmountF,
        uint40 amountF,
        uint48 toIdx
    ) external;

    // Batch forging function
    function forgeBatch(
        uint48 newLastIdx,
        uint256 newStRoot,
        uint256 newVouchRoot,
        uint256 newScoreRoot,
        uint256 newExitRoot,
        uint8 verifierIdx,
        uint256[2] calldata proofA,
        uint256[2][2] calldata proofB,
        uint256[2] calldata proofC,
        uint256 input
    ) external;

    // Getter functions
    function getStateRoot(uint32 batchNum) external view returns (uint256);
    function getLastForgedBatch() external view returns (uint32);

    // L1 Transaction Queue functions
    function getL1TransactionQueue(
        uint32 queueIndex
    ) external view returns (bytes memory);
    function getQueueLength() external view returns (uint32);

    // Creating the Account
    function createAccount() external payable;

    // Deposting Fucnction
    function deposit(uint48 fromIdx, uint40 loadAmountF, uint40 amountF) external payable;

    // 
    function exit(uint48 fromIdx, uint40 loadAmountF, uint40 amountF) external payable;

    function explode(uint48 fromIdx, uint40 loadAmountF, uint40 amountF) external payable;

    function setForgeL1BatchTimeout(uint8 newTimeout) external pure;


}
