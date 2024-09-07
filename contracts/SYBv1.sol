// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.23;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "./interfaces/ISybil.sol";

contract Sybil is Initializable, OwnableUpgradeable, ISybil {

    // Constants
    uint48 constant _RESERVED_IDX = 255;
    uint48 constant _EXIT_IDX = 1;
    uint48 constant _EXPLODE_IDX = 2;
    uint256 constant _LIMIT_LOAD_AMOUNT = (1 << 128);
    uint256 constant _LIMIT_L2TRANSFER_AMOUNT = (1 << 192);
    uint256 constant _L1_USER_TOTALBYTES = 74;
    uint256 constant _MAX_L1_TX = 128;

    // 74 = [20 bytes]fromEthAddr + [32 bytes]fromBjj-compressed + [6 bytes]fromIdx +[5 bytes]loadAmountFloat40 + [5 bytes]amountFloat40 + [6 bytes] toIdx
    // _MAX_L1_TX = Maximum L1 txns allowed to be queued in a batch. Hermez also has _MAX_L1_USER_TX, since L1txns = L1usertxns + L1Coordinatortxns, but we dont have coordinator txns

    //State variables
    uint48 public lastIdx;
    uint32 public lastForgedBatch;
    uint32 public nextL1ToForgeQueue;
    uint32 public nextL1FillingQueue;
    uint64 public lastL1L2Batch;
    uint8 public forgeL1L2BatchTimeout;

    // lastIdx = Last account index created inside the rollup
    // lastL1L2Batch = Ethereum block where the last L1-L2-batch was forged
    // forgeL1L2BatchTimeout = Max ethereum blocks after the last L1-L2-batch, when exceeds the timeout only L1-L2-batch are allowed


    // Mappings for various state roots and queue. Each batch forged will have a correlated 'state root', 'vouch root', 'score root' and 'exit root' and a 'l1L2TxDataHash'
    mapping(uint32 => uint256) public stateRootMap;
    mapping(uint32 => uint256) public vouchRootMap;
    mapping(uint32 => uint256) public scoreRootMap;
    mapping(uint32 => uint256) public exitRootsMap;
    mapping(uint32 => bytes32) public l1L2TxsDataHashMap;
    mapping(uint32 => bytes) public mapL1TxQueue;


    // Event emitted when a L1-user transaction is called and added to the nextL1FillingQueue queue
    event L1UserTxEvent(
        uint32 indexed queueIndex,
        uint8 indexed position, // Position inside the queue where the TX resides
        bytes l1UserTx
    );

    // Event emitted every time a batch is forged
    event ForgeBatch(uint32 indexed batchNum, uint16 l1UserTxsLen);

    // Event emitted when the governance update the `forgeL1L2BatchTimeout`
    event UpdateForgeL1L2BatchTimeout(uint8 newForgeL1L2BatchTimeout);

    // Event emitted when a withdrawal is done
    event WithdrawEvent(
        uint48 indexed idx,
        uint32 indexed numExitRoot,
        bool indexed instantWithdraw
    );

    // Event emitted when the contract is initialized
    event InitializeHermezEvent(
        uint8 forgeL1L2BatchTimeout,
    );


    /**
     * @dev Initializer function (equivalent to the constructor). Since we use
     * upgradeable smartcontracts the state vars have to be initialized here.
     */
    function initializeSybil(
        uint8 _forgeL1L2BatchTimeout,
    ) external initializer {
        // set default state variables
        lastIdx = _RESERVED_IDX;
        // lastL1L2Batch = 0 --> first batch forced to be L1Batch
        // nextL1ToForgeQueue = 0 --> First queue will be forged
        nextL1FillingQueue = 1;
        // stateRootMap[0] = 0 --> genesis batch will have root = 0
        emit InitializeHermezEvent(
            _forgeL1L2BatchTimeout
        );
    }

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
/*
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

*/




    function addL1Transaction(
        uint256 babyPubKey,
        uint48 fromIdx,
        uint40 loadAmountF,
        uint40 amountF,
        uint48 toIdx,
    ) external payable {
        uint256 loadAmount = _float2Fix(loadAmountF);
        require(
            loadAmount < _LIMIT_LOAD_AMOUNT,
            "Hermez::addL1Transaction: LOADAMOUNT_EXCEED_LIMIT"
        );
        require(
            loadAmount == msg.value,
            "Hermez::addL1Transaction: LOADAMOUNT_ETH_DOES_NOT_MATCH"
        );

        uint256 amount = _float2Fix(amountF);
        require(
            amount < _LIMIT_L2TRANSFER_AMOUNT,
            "Hermez::_addL1Transaction: AMOUNT_EXCEED_LIMIT"
        );

        if (fromIdx == 0 && toIdx == 0) {                 //is it safer to bracket each condition?
            // CreateAccount or CreateAccountDeposit
            if (babyPubKey == 0 || amount != 0) {
                revert InvalidCreateAccountTransaction();
            }
        } else if (toIdx == 0 && fromIdx > _RESERVED_IDX && fromIdx <= lastIdx) {
            // Deposit transaction
            if (babyPubKey != 0 || amount != 0) {
                revert InvalidDepositTransaction();
            }
        } else if (toIdx == _EXIT_IDX && fromIdx > _RESERVED_IDX && fromIdx <= lastIdx) {
            // ForceExit transaction
            if (babyPubKey != 0 || loadAmount != 0) {
                revert InvalidForceExitTransaction();
            }
        } else if (toIdx == _EXPLODE_IDX && fromIdx > _RESERVED_IDX && fromIdx <= lastIdx) {
            // ForceExplode transaction
            if (babyPubKey != 0 || amount != 0 || loadAmount != 0) {
                revert InvalidForceExplodeTransaction();
            }
        } else {
            // Invalid transaction parameters
            revert("Invalid transaction parameters");
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


    function _l1QueueAddTx(
        address ethAddress,
        uint256 babyPubKey,
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

        // Append the transaction to the queue
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
     * @dev Clear the current queue, and update the `nextL1ToForgeQueue` and `nextL1FillingQueue` if needed
     */
    function _clearQueue() internal returns (uint16) {
        uint16 l1UserTxsLen = uint16(
            mapL1TxQueue[nextL1ToForgeQueue].length / _L1_USER_TOTALBYTES
        );
        delete mapL1TxQueue[nextL1ToForgeQueue];
        nextL1ToForgeQueue++;
        if (nextL1ToForgeQueue == nextL1FillingQueue) {
            nextL1FillingQueue++;
        }
        //emit QueueCleared(nextL1ToForgeQueue, l1UserTxsLen);
        // do we need an event here?
        return l1UserTxsLen;
    }



    function forgeBatch(
        uint48 newLastIdx,
        uint256 newStRoot,
        uint256 newVouchRoot,
        uint256 newScoreRoot,
        uint256 newExitRoot,
        //bytes calldata encodedL1CoordinatorTx,
        //bytes calldata l1L2TxsData,
        //uint8 verifierIdx,
        bool l1Batch,
        //uint256[2] calldata proofA,
        //uint256[2][2] calldata proofB,
        //uint256[2] calldata proofC
    ) external virtual {
        if (msg.sender != tx.origin) {
            revert InternalTxNotAllowed();
        }

        if (
            !l1Batch && block.number >= (lastL1L2Batch + forgeL1L2BatchTimeout)
        ) {
            revert BatchTimeoutExceeded();
        }

        // update state
        lastForgedBatch++;
        lastIdx = newLastIdx;
        stateRootMap[lastForgedBatch] = newStRoot;
        vouchRootMap[lastForgedBatch] = newVouchRoot;
        scoreRootMap[lastForgedBatch] = newScoreRoot;
        exitRootsMap[lastForgedBatch] = newExitRoot;
        l1L2TxsDataHashMap[lastForgedBatch] = sha256(l1L2TxsData);

        uint16 l1UserTxsLen;
        if (l1Batch) {
            lastL1L2Batch = uint64(block.number);
            l1UserTxsLen = _clearQueue();
        }

        emit ForgeBatch(lastForgedBatch, l1UserTxsLen);
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


    // Getter functions
    function getStateRoot(
        uint32 batchNum
    ) external view override returns (uint256) {
        return stateRootMap[batchNum];
    }

    function getLastForgedBatch() external view override returns (uint32) {
        return lastForgedBatch;
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

}
