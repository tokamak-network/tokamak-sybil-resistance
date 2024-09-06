// SPDX-License-Identifier: MIT
pragma solidity 0.8.23;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "./interfaces/ISybil.sol";

contract SybilVerifier is Initializable, OwnableUpgradeable, ISybil {
    using SafeERC20 for IERC20;

    // Constants
    uint48 constant _RESERVED_IDX = 255;
    uint256 constant _LIMIT_LOAD_AMOUNT = (1 << 128);
    uint256 constant _LIMIT_L2TRANSFER_AMOUNT = (1 << 192);
    uint8 public constant ABSOLUTE_MAX_L1L2BATCHTIMEOUT = 240;

    // Struct definition
    struct VerifierRollup {
        address verifierInterface;
        uint256 maxTx;
        uint256 nLevels;
    }

    // Struct for account data
    struct AccountData {
        uint256 uniquenessScore;
        uint48 accountIndex;
    }

    // State variables
    VerifierRollup[] public rollupVerifiers;
    address public withdrawVerifier;
    uint48 public lastIdx;
    uint32 public lastForgedBatch;
    uint64 public lastL1L2Batch;
    uint32 public nextL1ToForgeQueue;
    uint32 public nextL1FillingQueue;
    uint8 public forgeL1L2BatchTimeout;
    uint256 public feeAddToken;
    address public tokenHEZ;
    bytes32 public merkleProof;

    // Mappings for various state roots and data
    mapping(uint32 => uint256) public stateRootMap;
    mapping(uint32 => uint256) public exitRootsMap;
    mapping(uint32 => bytes32) public l1L2TxsDataHashMap;
    mapping(uint32 => mapping(uint48 => bool)) public exitNullifierMap;
    mapping(uint32 => bytes) public mapL1TxQueue;
    mapping(address => AccountData) public accountData;

    address[] public tokenList;
    mapping(address => uint256) public tokenMap;

    /**
     * @dev Initializes the Sybil Verifier contract with necessary parameters and settings.
     * @param _verifiers Array of verifier addresses.
     * @param _verifiersParams Array of verifier parameters (e.g., maxTx, nLevels).
     * @param _withdrawVerifier Address of the withdrawal verifier.
     * @param _tokenHEZ Address of the token to be used for fees.
     * @param _forgeL1L2BatchTimeout The timeout for L1/L2 batch forging.
     * @param _feeAddToken Fee required to add a token.
     * @param _poseidon2Elements Address for Poseidon 2 elements.
     * @param _poseidon3Elements Address for Poseidon 3 elements.
     * @param _poseidon4Elements Address for Poseidon 4 elements.
     * @param _sybilGovernanceAddress Governance contract address.
     * @param _withdrawalDelay Delay for withdrawals.
     * @param _withdrawDelayerContract Contract handling withdrawal delays.
     */
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
    ) external override initializer {
        __Ownable_init(msg.sender);

        withdrawVerifier = _withdrawVerifier;
        tokenHEZ = _tokenHEZ;
        forgeL1L2BatchTimeout = _forgeL1L2BatchTimeout;
        feeAddToken = _feeAddToken;

        lastIdx = _RESERVED_IDX;
        nextL1FillingQueue = 1;
        tokenList.push(address(0));

        emit InitializeSybilVerifierEvent(
            _forgeL1L2BatchTimeout,
            _feeAddToken,
            _withdrawalDelay
        );
    }

    /**
     * @dev Adds an L1 transaction to the queue for processing.
     * @param babyPubKey The public key for the account.
     * @param fromIdx The index of the sender's account.
     * @param loadAmountF The load amount for the transaction.
     * @param amountF The amount for the transaction.
     * @param toIdx The index of the recipient's account.
     *
     * @notice Different transactions are validated based on the values of `fromIdx` and `toIdx`.
     *         The following are possible transactions:
     *         - Create Account: fromIdx = 0, toIdx = 0, loadAmountF = 0, amountF = 0, babyPubKey != 0
     *         - Create Account Deposit: fromIdx = 0, toIdx = 0, loadAmountF > 0, babyPubKey != 0
     *         - Deposit: fromIdx >= _RESERVED_IDX, toIdx = 0, babyPubKey = 0
     *         - ForceExit: fromIdx >= _RESERVED_IDX, toIdx = 1, amountF > 0
     *         - ForceExplode: fromIdx >= _RESERVED_IDX, toIdx = 2
     */
    function addL1Transaction(
        uint256 babyPubKey,
        uint48 fromIdx,
        uint40 loadAmountF,
        uint40 amountF,
        uint48 toIdx
    ) external payable {
        // Check loadAmount
        uint256 loadAmount = _float2Fix(loadAmountF);
        require(
            loadAmount < _LIMIT_LOAD_AMOUNT,
            "Sybil::addL1Transaction: LOADAMOUNT_EXCEED_LIMIT"
        );

        // Validate transaction type based on fromIdx and toIdx
        if (fromIdx == 0 && toIdx == 0) {
            // CreateAccount or CreateAccountDeposit
            if (babyPubKey == 0 || amountF != 0) {
                revert InvalidCreateAccountTransaction();
            }
            // If loadAmount is non-zero, it is a CreateAccountDeposit
            if (loadAmount > 0 && loadAmount != msg.value) {
                revert InvalidCreateAccountDepositTransaction();
            }
        } else if (fromIdx >= _RESERVED_IDX && toIdx == 0) {
            // Deposit transaction
            if (babyPubKey != 0 || amountF != 0) {
                revert InvalidDepositTransaction();
            }
        } else if (fromIdx >= _RESERVED_IDX && toIdx == 1) {
            // ForceExit transaction
            if (amountF == 0 || babyPubKey != 0 || loadAmount != 0) {
                revert InvalidForceExitTransaction();
            }
        } else if (fromIdx >= _RESERVED_IDX && toIdx == 2) {
            // ForceExplode transaction
            if (babyPubKey != 0 || amountF != 0 || loadAmount != 0) {
                revert InvalidForceExplodeTransaction();
            }
        } else {
            // Invalid transaction parameters
            revert("Invalid transaction parameters");
        }

        // Perform L1 User Transaction
        _addL1Transaction(
            msg.sender,
            babyPubKey,
            fromIdx,
            loadAmountF,
            amountF,
            toIdx
        );

        // Emit event with the appropriate interpretation of `toIdx`
        emit L1TransactionAdded(msg.sender, fromIdx, toIdx, amountF);
    }

    /**
     * @dev Internal function to append an L1 transaction to the queue.
     * @param ethAddress The sender's address.
     * @param babyPubKey The public key for the account.
     * @param fromIdx The index of the sender's account.
     * @param loadAmountF The load amount for the transaction.
     * @param amountF The amount for the transaction.
     * @param toIdx The index of the recipient's account.
     */
    function _addL1Transaction(
        address ethAddress,
        uint256 babyPubKey,
        uint48 fromIdx,
        uint40 loadAmountF,
        uint40 amountF,
        uint48 toIdx
    ) internal {
        // Encode the transaction data
        bytes memory l1Tx = abi.encodePacked(
            ethAddress,
            babyPubKey,
            fromIdx,
            loadAmountF,
            amountF,
            toIdx
        );

        // Get the current position in the queue
        uint256 currentPosition = mapL1TxQueue[nextL1FillingQueue].length;

        // Append the transaction to the queue
        mapL1TxQueue[nextL1FillingQueue] = bytes.concat(
            mapL1TxQueue[nextL1FillingQueue],
            l1Tx
        );

        // If the queue exceeds the maximum transactions, move to the next queue
        if ((currentPosition + l1Tx.length) >= _LIMIT_L2TRANSFER_AMOUNT) {
            nextL1FillingQueue++;
        }
    }

    /**
     * @dev Clears the queue after processing transactions.
     * @return Number of transactions cleared from the queue.
     */
    function _clearQueue() internal returns (uint16) {
        uint16 l1UserTxsLen = uint16(mapL1TxQueue[nextL1ToForgeQueue].length);
        delete mapL1TxQueue[nextL1ToForgeQueue];
        nextL1ToForgeQueue++;
        if (nextL1ToForgeQueue == nextL1FillingQueue) {
            nextL1FillingQueue++;
        }

        emit QueueCleared(nextL1ToForgeQueue, l1UserTxsLen);
        return l1UserTxsLen;
    }

    /**
     * @dev Forges a batch of transactions, processing L1 and/or L2 batches.
     * @param newLastIdx The index of the last processed account.
     * @param newStRoot The new state root.
     * @param newExitRoot The new exit root.
     * @param encodedL1CoordinatorTx Encoded data for L1 coordinator transaction.
     * @param l1L2TxsData Encoded L1 and L2 transaction data.
     * @param feeIdxCoordinator Coordinator index for fees.
     * @param verifierIdx Index of the verifier to use.
     * @param l1Batch Boolean flag indicating if this is an L1 batch.
     * @param proofA Proof part A for zk-SNARK verification.
     * @param proofB Proof part B for zk-SNARK verification.
     * @param proofC Proof part C for zk-SNARK verification.
     */
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
    ) external override {
        if (msg.sender != tx.origin) {
            revert InternalTxNotAllowed();
        }

        if (
            !l1Batch && block.number >= (lastL1L2Batch + forgeL1L2BatchTimeout)
        ) {
            revert BatchTimeoutExceeded();
        }

        uint256 input = _constructCircuitInput(
            newLastIdx,
            newStRoot,
            newExitRoot,
            l1Batch,
            verifierIdx
        );

        if (
            !_verifyProof(
                rollupVerifiers[verifierIdx].verifierInterface,
                proofA,
                proofB,
                proofC,
                input
            )
        ) {
            revert InvalidProof();
        }

        lastForgedBatch++;
        lastIdx = newLastIdx;
        stateRootMap[lastForgedBatch] = newStRoot;
        exitRootsMap[lastForgedBatch] = newExitRoot;
        l1L2TxsDataHashMap[lastForgedBatch] = sha256(l1L2TxsData);

        uint16 l1UserTxsLen;
        if (l1Batch) {
            lastL1L2Batch = uint64(block.number);
            l1UserTxsLen = _clearQueue();
        }

        emit BatchForged(
            lastForgedBatch,
            newLastIdx,
            newStRoot,
            newExitRoot,
            l1Batch ? 1 : 0
        ); // 1: L1, 0: L2
    }

    /**
     * @dev Sets the L1/L2 batch timeout.
     * @param newTimeout New timeout value for the batches.
     */
    function setForgeL1L2BatchTimeout(
        uint8 newTimeout
    ) external override onlyOwner {
        require(
            newTimeout <= ABSOLUTE_MAX_L1L2BATCHTIMEOUT,
            "SybilVerifier::setForgeL1L2BatchTimeout: MAX_TIMEOUT_EXCEEDED"
        );
        forgeL1L2BatchTimeout = newTimeout;
    }

    /**
     * @dev Sets the fee for adding a new token.
     * @param newFee The new fee to be charged for adding a token.
     */
    function setFeeAddToken(uint256 newFee) external override onlyOwner {
        feeAddToken = newFee;
    }

    /**
     * @dev Sets the Merkle proof for the verifier.
     * @param proof The Merkle proof.
     */
    function setMerkleProof(bytes32 proof) external override onlyOwner {
        merkleProof = proof;
    }

    /**
     * @dev Returns the Merkle proof.
     * @return The Merkle proof.
     */
    function getMerkleProof() external view override returns (bytes32) {
        return merkleProof;
    }

    // Getter functions
    function getStateRoot(
        uint32 batchNum
    ) external view override returns (uint256) {
        return stateRootMap[batchNum];
    }

    function getLastForgedBatch() external view override returns (uint32) {
        return lastForgedBatch;
    }

    function getUniquenessScore(
        address account
    ) external view override returns (uint256) {
        return accountData[account].uniquenessScore;
    }

    function getL1TransactionQueue(
        uint32 queueIndex
    ) external view override returns (bytes memory) {
        return mapL1TxQueue[queueIndex];
    }

    function getQueueLength() external view override returns (uint32) {
        return nextL1FillingQueue - nextL1ToForgeQueue;
    }

    // Placeholder for floating point to fixed point conversion
    function _float2Fix(uint40 floatVal) internal pure returns (uint256) {
        return uint256(floatVal) * 10 ** (18 - 8);
    }

    // Placeholder for proof verification logic
    function _verifyProof(
        address verifier,
        uint256[2] calldata proofA,
        uint256[2][2] calldata proofB,
        uint256[2] calldata proofC,
        uint256 input
    ) internal view returns (bool) {
        return true; // Placeholder logic
    }

    // Placeholder for constructing circuit input
    function _constructCircuitInput(
        uint48 newLastIdx,
        uint256 newStRoot,
        uint256 newExitRoot,
        bool l1Batch,
        uint8 verifierIdx
    ) internal view returns (uint256) {
        return 0; // Placeholder logic
    }

    // Placeholder for ERC20 permit functionality
    function _permit(
        address token,
        uint256 _amount,
        bytes calldata _permitData
    ) internal {
        // Placeholder logic
    }
}
