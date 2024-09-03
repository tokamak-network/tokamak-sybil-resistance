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
    uint8 public constant ABSOLUTE_MAX_L1L2BATCHTIMEOUT = 240;
    uint48 constant _RESERVED_IDX = 255;
    uint256 constant _LIMIT_LOAD_AMOUNT = (1 << 128);
    uint256 constant _LIMIT_L2TRANSFER_AMOUNT = (1 << 192);

    // Struct definition: DO WE NEED THIS?
    struct VerifierRollup {
        address verifierInterface;
        uint256 maxTx;
        uint256 nLevels;
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
    bytes32 public merkleProof; // Placeholder for Merkle proof

    // Mappings
    mapping(uint32 => uint256) public stateRootMap;
    mapping(uint32 => uint256) public exitRootsMap;
    mapping(uint32 => bytes32) public l1L2TxsDataHashMap;
    mapping(uint32 => mapping(uint48 => bool)) public exitNullifierMap;
    mapping(uint32 => bytes) public mapL1TxQueue;

    address[] public tokenList;
    mapping(address => uint256) public tokenMap;

    // Events (as per interface)
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
    ) external override initializer {
        __Ownable_init();

        // Initialization logic
        withdrawVerifier = _withdrawVerifier;
        tokenHEZ = _tokenHEZ;
        forgeL1L2BatchTimeout = _forgeL1L2BatchTimeout;
        feeAddToken = _feeAddToken;

        lastIdx = _RESERVED_IDX;
        nextL1FillingQueue = 1;
        tokenList.push(address(0));

        emit InitializeSybilVerifierEvent(_forgeL1L2BatchTimeout, _feeAddToken, _withdrawalDelay);
    }

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
    ) external payable override {
        // Validate transaction type
        if (transactionType != DEPOSIT && transactionType != WITHDRAWAL && transactionType != TRANSFER) {
            revert InvalidTransactionType(transactionType);
        }

        // Validate token registration
        if (tokenID >= tokenList.length) {
            revert TokenNotRegistered(tokenID);
        }

        // Convert floating point to fixed point
        uint256 loadAmount = _float2Fix(loadAmountF);
        if (loadAmount >= _LIMIT_LOAD_AMOUNT) {
            revert InvalidLoadAmount(loadAmount, _LIMIT_LOAD_AMOUNT);
        }

        // Handle ETH vs ERC20 transactions
        if (loadAmount > 0) {
            if (tokenID == 0) { // ETH
                if (loadAmount != msg.value) {
                    revert InvalidLoadAmount(msg.value, loadAmount);
                }
            } else { // ERC20
                if (msg.value != 0) {
                    revert InvalidLoadAmount(msg.value, 0);
                }
                if (permit.length != 0) {
                    _permit(tokenList[tokenID], loadAmount, permit);
                }
                _safeTransferFrom(tokenList[tokenID], msg.sender, address(this), loadAmount);
            }
        }

        // Add the transaction to the queue
        _addL1Transaction(msg.sender, babyPubKey, fromIdx, loadAmountF, amountF, tokenID, toIdx);

        emit L1TransactionAdded(msg.sender, fromIdx, toIdx, tokenID, amountF, transactionType);
    }

    function _addL1Transaction(
        address ethAddress,
        uint256 babyPubKey,
        uint48 fromIdx,
        uint40 loadAmountF,
        uint40 amountF,
        uint32 tokenID,
        uint48 toIdx
    ) internal {
        // Encode the transaction data
        bytes memory l1Tx = abi.encodePacked(
            ethAddress,
            babyPubKey,
            fromIdx,
            loadAmountF,
            amountF,
            tokenID,
            toIdx
        );

        // Get the current position in the queue
        uint256 currentPosition = mapL1TxQueue[nextL1FillingQueue].length;

        // Append the transaction to the queue
        mapL1TxQueue[nextL1FillingQueue] = bytes.concat(mapL1TxQueue[nextL1FillingQueue], l1Tx);

        // If the queue exceeds the maximum transactions, move to the next queue
        if ((currentPosition + l1Tx.length) >= _LIMIT_L2TRANSFER_AMOUNT) {
            nextL1FillingQueue++;
        }
    }

    // Clears the queue after processing
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

        if (!l1Batch && block.number >= (lastL1L2Batch + forgeL1L2BatchTimeout)) {
            revert BatchTimeoutExceeded();
        }

        uint256 input = _constructCircuitInput(newLastIdx, newStRoot, newExitRoot, l1Batch, verifierIdx);

        if (!_verifyProof(rollupVerifiers[verifierIdx].verifierInterface, proofA, proofB, proofC, input)) {
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

        emit BatchForged(lastForgedBatch, newLastIdx, newStRoot, newExitRoot, l1Batch ? 1 : 0); // 1: L1, 0: L2
    }

    // Governance functions
    function setForgeL1L2BatchTimeout(uint8 newTimeout) external override onlyOwner {
        require(newTimeout <= ABSOLUTE_MAX_L1L2BATCHTIMEOUT, "SybilVerifier::setForgeL1L2BatchTimeout: MAX_TIMEOUT_EXCEEDED");
        forgeL1L2BatchTimeout = newTimeout;
    }

    function setFeeAddToken(uint256 newFee) external override onlyOwner {
        feeAddToken = newFee;
    }

    // Merkle proof functions
    function setMerkleProof(bytes32 proof) external override onlyOwner {
        merkleProof = proof;
    }

    function getMerkleProof() external view override returns (bytes32) {
        return merkleProof;
    }

    // Getter functions
    function getStateRoot(uint32 batchNum) external view override returns (uint256) {
        return stateRootMap[batchNum];
    }

    function getLastForgedBatch() external view override returns (uint32) {
        return lastForgedBatch;
    }

    function getUniquenessScore(address account) external view override returns (uint256) {
        return accountData[account].uniquenessScore;
    }

    // L1 Transaction Queue functions
    function getL1TransactionQueue(uint32 queueIndex) external view override returns (bytes memory) {
        return mapL1TxQueue[queueIndex];
    }

    function getQueueLength() external view override returns (uint32) {
        return nextL1FillingQueue - nextL1ToForgeQueue;
    }

    // Placeholder for floating point to fixed point conversion
    function _float2Fix(uint40 floatVal) internal pure returns (uint256) {
        return uint256(floatVal) * 10**(18 - 8);
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
    function _permit(address token, uint256 _amount, bytes calldata _permitData) internal {
        // Placeholder logic
    }
}
