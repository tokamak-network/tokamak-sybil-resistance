// SPDX-License-Identifier: MIT
pragma solidity 0.8.23;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "./interfaces/ISybil.sol";
import "./interfaces/IVerifierRollup.sol";
import "./sybilHelpers.sol";

contract Sybil is Initializable, OwnableUpgradeable, ISybil, SybilHelpers {
    uint8 public constant ABSOLUTE_MAX_L1L2BATCHTIMEOUT = 240;
    uint8 public forgeL1L2BatchTimeout;

    uint32 public lastForgedBatch;
    uint32 public nextL1ToForgeQueue;
    uint32 public nextL1FillingQueue;

    uint48 constant _RESERVED_IDX = 255;
    uint48 constant _EXIT_IDX = 1;
    uint48 constant _EXPLODE_IDX = 2;
    uint48 public lastIdx;

    uint64 public lastL1L2Batch;

    uint256 constant _LIMIT_LOAD_AMOUNT = (1 << 128);
    uint256 constant _LIMIT_L2TRANSFER_AMOUNT = (1 << 192);
    uint256 constant _L1_USER_TOTALBYTES = 74;
    uint256 constant _MAX_L1_TX = 128;

    mapping(uint32 => bytes) public mapL1TxQueue;
    mapping(uint32 => bytes32) public l1L2TxsDataHashMap;

    mapping(uint32 => uint256) public stateRootMap;
    mapping(uint32 => uint256) public vouchRootMap;
    mapping(uint32 => uint256) public scoreRootMap;
    mapping(uint32 => uint256) public exitRootsMap;

    // Mapping of exit nullifiers, only allowing each withdrawal to be made once
    // rootId => (Idx => true/false)
    mapping(uint32 => mapping(uint48 => bool)) public exitNullifierMap;

    struct VerifierRollup {
        VerifierRollupInterface verifierInterface;
        uint256 maxTxs; // maximum rollup transactions in a batch: L2-tx + L1-tx transactions
        uint256 nLevels; // number of levels of the circuit
    }

    // Verifiers array
    VerifierRollup[] public rollupVerifiers;

    event L1UserTxEvent(
        uint32 indexed queueIndex,
        uint8 indexed position,
        bytes l1UserTx
    );

    event ForgeBatch(uint32 indexed batchNum, uint16 l1UserTxsLen);
    event UpdateForgeL1L2BatchTimeout(uint8 newForgeL1L2BatchTimeout);
    event WithdrawEvent(uint48 indexed idx, uint32 indexed numExitRoot);
    event Initialize(uint8 forgeL1L2BatchTimeout);

    constructor(
        address[] memory verifiers,
        uint256[] memory maxTxs,
        uint256[] memory nLevels,
        uint8 _forgeL1L2BatchTimeout,
        address _poseidon2Elements,
        address _poseidon3Elements,
        address _poseidon4Elements
    ) {
        initialize(
            verifiers,
            maxTxs,
            nLevels,
            _forgeL1L2BatchTimeout,
            _poseidon2Elements,
            _poseidon3Elements,
            _poseidon4Elements
        );
    }
    /**
     * @notice Initializes the Sybil contract.
     * @param _forgeL1L2BatchTimeout Timeout value for batch creation in blocks.
     */
    function initialize(
        address[] memory verifiers,
        uint256[] memory maxTxs,
        uint256[] memory nLevels,
        uint8 _forgeL1L2BatchTimeout,
        address _poseidon2Elements,
        address _poseidon3Elements,
        address _poseidon4Elements
    ) public initializer {
        lastIdx = _RESERVED_IDX;
        nextL1FillingQueue = 1;

        _initializeVerifiers(verifiers, maxTxs, nLevels);

        _initializeHelpers(
            _poseidon2Elements,
            _poseidon3Elements,
            _poseidon4Elements
        );

        emit Initialize(_forgeL1L2BatchTimeout);
    }

    /**
     * @notice Adds an L1 transaction to the queue.
     * @param babyPubKey The public key for the account.
     * @param fromIdx The index of the sender in the queue.
     * @param loadAmountF The load amount in floating point.
     * @param amountF The transaction amount in floating point.
     * @param toIdx The index of the receiver in the queue.
     */
    function addL1Transaction(
        string memory babyPubKey,
        uint48 fromIdx,
        uint40 loadAmountF,
        uint40 amountF,
        uint48 toIdx
    ) external payable {
        uint256 loadAmount = _float2Fix(loadAmountF);
        if (loadAmount >= _LIMIT_LOAD_AMOUNT) {
            revert LoadAmountExceedsLimit();
        }
        if (loadAmount != msg.value) {
            revert LoadAmountDoesNotMatch();
        }

        uint256 amount = _float2Fix(amountF);
        if (amount >= _LIMIT_L2TRANSFER_AMOUNT) {
            revert AmountExceedsLimit();
        }

        if (fromIdx == 0 && toIdx == 0) {
            if (
                keccak256(abi.encodePacked(babyPubKey)) ==
                keccak256(abi.encodePacked("")) ||
                amount != 0
            ) {
                revert InvalidCreateAccountTransaction();
            }
        } else if (
            toIdx == 0 && fromIdx > _RESERVED_IDX && fromIdx <= lastIdx
        ) {
            if (
                keccak256(abi.encodePacked(babyPubKey)) !=
                keccak256(abi.encodePacked("")) ||
                amount != 0
            ) {
                revert InvalidDepositTransaction();
            }
        } else if (
            toIdx == _EXIT_IDX && fromIdx > _RESERVED_IDX && fromIdx <= lastIdx
        ) {
            if (
                keccak256(abi.encodePacked(babyPubKey)) !=
                keccak256(abi.encodePacked("")) ||
                loadAmount != 0
            ) {
                revert InvalidForceExitTransaction();
            }
        } else if (
            toIdx == _EXPLODE_IDX &&
            fromIdx > _RESERVED_IDX &&
            fromIdx <= lastIdx
        ) {
            if (
                keccak256(abi.encodePacked(babyPubKey)) !=
                keccak256(abi.encodePacked("")) ||
                amount != 0 ||
                loadAmount != 0
            ) {
                revert InvalidForceExplodeTransaction();
            }
        } else {
            revert InvalidTransactionParameters();
        }

        _l1QueueAddTx(
            msg.sender,
            babyPubKey,
            fromIdx,
            loadAmountF,
            amountF,
            toIdx
        );
    }

    /**
     * @notice Adds the transaction data to the L1 transaction queue.
     * @param ethAddress The Ethereum address of the sender.
     * @param babyPubKey The public key for the account.
     * @param fromIdx The index of the sender.
     * @param loadAmountF The load amount in floating point.
     * @param amountF The transaction amount in floating point.
     * @param toIdx The index of the receiver.
     */
    function _l1QueueAddTx(
        address ethAddress,
        string memory babyPubKey,
        uint48 fromIdx,
        uint40 loadAmountF,
        uint40 amountF,
        uint48 toIdx
    ) internal {
        bytes memory l1Tx = abi.encodePacked(
            ethAddress,
            babyPubKey,
            fromIdx,
            loadAmountF,
            amountF,
            toIdx
        );

        uint256 currentPosition = mapL1TxQueue[nextL1FillingQueue].length /
            _L1_USER_TOTALBYTES;

        mapL1TxQueue[nextL1FillingQueue] = bytes.concat(
            mapL1TxQueue[nextL1FillingQueue],
            l1Tx
        );

        emit L1UserTxEvent(nextL1FillingQueue, uint8(currentPosition), l1Tx);

        if (currentPosition + 1 >= _MAX_L1_TX) {
            nextL1FillingQueue++;
        }
    }

    /**
     * @notice Clears the current queue after batch processing.
     * @return l1UserTxsLen The number of user transactions in the batch.
     */
    function _clearQueue() internal returns (uint16) {
        uint16 l1UserTxsLen = uint16(
            mapL1TxQueue[nextL1ToForgeQueue].length / _L1_USER_TOTALBYTES
        );
        delete mapL1TxQueue[nextL1ToForgeQueue];
        nextL1ToForgeQueue++;
        return l1UserTxsLen;
    }

    /**
     * @notice Forges a new batch of transactions, updating state.
     * @param newLastIdx The last index of the batch.
     * @param newStRoot The new state root.
     * @param newVouchRoot The new vouch root.
     * @param newScoreRoot The new score root.
     * @param newExitRoot The new exit root.
     * @param l1Batch Whether this is an L1 batch or not.
     */
    function forgeBatch(
        uint48 newLastIdx,
        uint256 newStRoot,
        uint256 newVouchRoot,
        uint256 newScoreRoot,
        uint256 newExitRoot,
        uint8 verifierIdx,
        bool l1Batch,
        uint256[2] calldata proofA,
        uint256[2][2] calldata proofB,
        uint256[2] calldata proofC,
        uint256 input
    ) external virtual {
        lastForgedBatch++;
        lastIdx = newLastIdx;
        stateRootMap[lastForgedBatch] = newStRoot;
        vouchRootMap[lastForgedBatch] = newVouchRoot;
        scoreRootMap[lastForgedBatch] = newScoreRoot;
        exitRootsMap[lastForgedBatch] = newExitRoot;

        // verify proof
        if (
            !rollupVerifiers[verifierIdx].verifierInterface.verifyProof(
                proofA,
                proofB,
                proofC,
                [input]
            )
        ) {
            revert InvalidProof();
        }

        uint16 l1UserTxsLen;
        if (l1Batch) {
            lastL1L2Batch = uint64(block.number);
            l1UserTxsLen = _clearQueue();
        }

        emit ForgeBatch(lastForgedBatch, l1UserTxsLen);
    }

    /**
     * @dev Withdraw to retrieve the tokens from the exit tree to the owner account
     * Before this call an exit transaction must be done
     * @param amount Amount to retrieve
     * @param babyPubKey Public key babyjubjub represented as point: sign + (Ay)
     * @param numExitRoot Batch number where the exit transaction has been done
     * @param siblings Exit tree inclusion proof
     * @param idx Index of the exit tree account
     * Events: `WithdrawEvent`
     */
    function withdrawMerkleProof(
        uint192 amount,
        uint256 babyPubKey,
        uint32 numExitRoot,
        uint256[] calldata siblings,
        uint48 idx
    ) external {
        // Build 'key' and 'value' for exit tree
        uint256[4] memory arrayState = _buildTreeState(
            0,
            amount,
            babyPubKey,
            msg.sender
        );
        uint256 stateHash = _hash4Elements(arrayState);

        // Check exit tree nullifier
        if (exitNullifierMap[numExitRoot][idx]) {
            revert WithdrawAlreadyDone();
        }

        // Get exit root given its index depth
        uint256 exitRoot = exitRootsMap[numExitRoot];

        // Check sparse merkle tree proof
        if (!_smtVerifier(exitRoot, siblings, idx, stateHash)) {
            revert SmtProofInvalid();
        }

        // Set nullifier
        exitNullifierMap[numExitRoot][idx] = true;

        _withdrawFunds(amount);

        emit WithdrawEvent(idx, numExitRoot);
    }

    /**
     * @dev Withdraw the funds to the msg.sender if instant withdraw or to the withdraw delayer if delayed
     * @param amount Amount to retrieve
     */
    function _withdrawFunds(uint192 amount) internal {
        _safeTransfer(amount);
    }

    /**
     * @dev Transfer tokens or ether from the smart contract
     * @param value Quantity to transfer
     */
    function _safeTransfer(uint256 value) internal {
        /* solhint-disable avoid-low-level-calls */
        (bool success, ) = msg.sender.call{value: value}(new bytes(0));
        if (!success) {
            revert EthTransferFailed();
        }
    }

    //** SETTERS */

    /**
     * @notice Sets the L1/L2 batch timeout.
     * @param newTimeout The new timeout value in blocks.
     */
    function setForgeL1L2BatchTimeout(uint8 newTimeout) external onlyOwner {
        if (newTimeout > ABSOLUTE_MAX_L1L2BATCHTIMEOUT) {
            revert BatchTimeoutExceeded();
        }
        forgeL1L2BatchTimeout = newTimeout;
    }
    /**
     * @notice Sets the value of `lastForgedBatch`.
     * @param _lastForgedBatch The new value for `lastForgedBatch`.
     */
    function setLastForgedBatch(uint32 _lastForgedBatch) external onlyOwner {
        lastForgedBatch = _lastForgedBatch;
    }

    /**
     * @notice Sets the value of `nextL1ToForgeQueue`.
     * @param _nextL1ToForgeQueue The new value for `nextL1ToForgeQueue`.
     */
    function setNextL1ToForgeQueue(
        uint32 _nextL1ToForgeQueue
    ) external onlyOwner {
        nextL1ToForgeQueue = _nextL1ToForgeQueue;
    }

    /**
     * @notice Sets the value of `nextL1FillingQueue`.
     * @param _nextL1FillingQueue The new value for `nextL1FillingQueue`.
     */
    function setNextL1FillingQueue(
        uint32 _nextL1FillingQueue
    ) external onlyOwner {
        nextL1FillingQueue = _nextL1FillingQueue;
    }

    /**
     * @notice Sets the value of `lastIdx`.
     * @param _lastIdx The new value for `lastIdx`.
     */
    function setLastIdx(uint48 _lastIdx) external onlyOwner {
        lastIdx = _lastIdx;
    }

    /**
     * @notice Sets the value of `lastL1L2Batch`.
     * @param _lastL1L2Batch The new value for `lastL1L2Batch`.
     */
    function setLastL1L2Batch(uint64 _lastL1L2Batch) external onlyOwner {
        lastL1L2Batch = _lastL1L2Batch;
    }

    /**
     * @notice Sets the state root for a given batch.
     * @param batchNum The batch number to set the state root for.
     * @param newStateRoot The new state root value.
     */
    function setStateRoot(
        uint32 batchNum,
        uint256 newStateRoot
    ) external onlyOwner {
        stateRootMap[batchNum] = newStateRoot;
    }

    /**
     * @notice Sets the vouch root for a given batch.
     * @param batchNum The batch number to set the vouch root for.
     * @param newVouchRoot The new vouch root value.
     */
    function setVouchRoot(
        uint32 batchNum,
        uint256 newVouchRoot
    ) external onlyOwner {
        vouchRootMap[batchNum] = newVouchRoot;
    }

    /**
     * @notice Sets the score root for a given batch.
     * @param batchNum The batch number to set the score root for.
     * @param newScoreRoot The new score root value.
     */
    function setScoreRoot(
        uint32 batchNum,
        uint256 newScoreRoot
    ) external onlyOwner {
        scoreRootMap[batchNum] = newScoreRoot;
    }

    /**
     * @notice Sets the exit root for a given batch.
     * @param batchNum The batch number to set the exit root for.
     * @param newExitRoot The new exit root value.
     */
    function setExitRoot(
        uint32 batchNum,
        uint256 newExitRoot
    ) external onlyOwner {
        exitRootsMap[batchNum] = newExitRoot;
    }

    /**
     * @notice Sets the L1/L2 transactions data hash for a given batch.
     * @param batchNum The batch number to set the data hash for.
     * @param newDataHash The new data hash value.
     */
    function setL1L2TxsDataHash(
        uint32 batchNum,
        bytes32 newDataHash
    ) external onlyOwner {
        l1L2TxsDataHashMap[batchNum] = newDataHash;
    }

    //** GETTERS */

    /**
     * @notice Retrieves the state root for a given batch.
     * @param batchNum The batch number.
     * @return The state root of the batch.
     */
    function getStateRoot(uint32 batchNum) external view returns (uint256) {
        return stateRootMap[batchNum];
    }
    /**
     * @notice Gets the current L1/L2 batch timeout value.
     * @return The current timeout value in blocks.
     */
    function getForgeL1L2BatchTimeout() external view returns (uint8) {
        return forgeL1L2BatchTimeout;
    }
    /**
     * @notice Retrieves the last index used.
     * @return The last index value.
     */
    function getLastIdx() external view returns (uint48) {
        return lastIdx;
    }

    /**
     * @notice Retrieves the block number of the last L1/L2 batch.
     * @return The block number of the last L1/L2 batch.
     */
    function getLastL1L2Batch() external view returns (uint64) {
        return lastL1L2Batch;
    }

    /**
     * @notice Retrieves the next L1 queue index to be forged.
     * @return The next L1 queue index.
     */
    function getNextL1ToForgeQueue() external view returns (uint32) {
        return nextL1ToForgeQueue;
    }
    /**
     * @notice Retrieves the next L1 filling queue index.
     * @return The next L1 filling queue index.
     */
    function getNextL1FillingQueue() external view returns (uint32) {
        return nextL1FillingQueue;
    }

    /**
     * @notice Retrieves the vouch root for a given batch.
     * @param batchNum The batch number.
     * @return The vouch root of the batch.
     */
    function getVouchRoot(uint32 batchNum) external view returns (uint256) {
        return vouchRootMap[batchNum];
    }

    /**
     * @notice Retrieves the score root for a given batch.
     * @param batchNum The batch number.
     * @return The score root of the batch.
     */
    function getScoreRoot(uint32 batchNum) external view returns (uint256) {
        return scoreRootMap[batchNum];
    }

    /**
     * @notice Retrieves the exit root for a given batch.
     * @param batchNum The batch number.
     * @return The exit root of the batch.
     */
    function getExitRoot(uint32 batchNum) external view returns (uint256) {
        return exitRootsMap[batchNum];
    }

    /**
     * @notice Retrieves the L1/L2 transactions data hash for a given batch.
     * @param batchNum The batch number.
     * @return The data hash for the L1/L2 transactions.
     */
    function getL1L2TxsDataHash(
        uint32 batchNum
    ) external view returns (bytes32) {
        return l1L2TxsDataHashMap[batchNum];
    }

    /**
     * @notice Retrieves the last forged batch number.
     * @return The last forged batch number.
     */
    function getLastForgedBatch() external view returns (uint32) {
        return lastForgedBatch;
    }

    /**
     * @notice Retrieves the L1 transaction queue for a given index.
     * @param queueIndex The index of the queue.
     * @return The transaction queue in bytes format.
     */
    function getL1TransactionQueue(
        uint32 queueIndex
    ) external view returns (bytes memory) {
        return mapL1TxQueue[queueIndex];
    }

    /**
     * @notice Retrieves the current length of the queue.
     * @return The length of the queue.
     */
    function getQueueLength() external view returns (uint32) {
        return nextL1FillingQueue - nextL1ToForgeQueue;
    }

    /**
     * @notice Converts floating point values to fixed point for amounts.
     * @param floatVal The floating point value.
     * @return The fixed point equivalent of the floating value.
     */
    function _float2Fix(uint40 floatVal) internal pure returns (uint256) {
        return uint256(floatVal) * 10 ** (18 - 8);
    }

    /**
     * @dev Initialize verifiers
     * @param _verifiers verifiers address array
     * @param _maxTxs encoeded maxTx of the verifier
     * @param _nLevels encoeded nlevels of the verifier
     */
    function _initializeVerifiers(
        address[] memory _verifiers,
        uint256[] memory _maxTxs,
        uint256[] memory _nLevels
    ) internal {
        uint256 len = _verifiers.length;
        for (uint256 i = 0; i < len; ++i) {
            if (_verifiers[i] == address(0)) {
                revert InvalidVerifierAddress();
            }

            rollupVerifiers.push(
                VerifierRollup({
                    verifierInterface: VerifierRollupInterface(_verifiers[i]),
                    maxTxs: _maxTxs[i],
                    nLevels: _nLevels[i]
                })
            );
        }
    }
}
