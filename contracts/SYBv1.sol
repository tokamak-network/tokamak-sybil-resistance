// SPDX-License-Identifier: AGPL-3.0

pragma solidity 0.8.23;

contract Hermez is InstantWithdrawManager {

    // First 256 indexes reserved, first user index will be the 256
    uint48 constant _RESERVED_IDX = 255;

    // IDX 1 is reserved for exits
    uint48 constant _EXIT_IDX = 1;

    // IDX 2 is reserved for exits
    uint48 constant _EXPLODE_IDX = 1;

    // Max load amount allowed (loadAmount: L1 --> L2)
    uint256 constant _LIMIT_LOAD_AMOUNT = (1 << 128);

    // Max amount allowed (amount L2 --> L2)
    uint256 constant _LIMIT_L2TRANSFER_AMOUNT = (1 << 192);

    // [20 bytes] fromEthAddr + [32 bytes] fromBjj-compressed + [6 bytes] fromIdx +
    // [5 bytes] loadAmountFloat40 + [5 bytes] amountFloat40 + [6 bytes] toIdx
    uint256 constant _L1_USER_TOTALBYTES = 74;

    // Maximum L1 transactions allowed to be queued in a batch
    // Hermez also has _MAX_L1_USER_TX, since L1txns = L1usertxns + L1Coordinatortxns, but we dont have coordinator txns
    uint256 constant _MAX_L1_TX = 128;

    // Last account index created inside the rollup
    uint48 public lastIdx;

    // Last batch forged
    uint32 public lastForgedBatch;

    // Each batch forged will have a correlated 'state root'
    mapping(uint32 => uint256) public stateRootMap;

    // Each batch forged will have a correlated 'vouch root'
    mapping(uint32 => uint256) public vouchRootMap;

    // Each batch forged will have a correlated 'score root'
    mapping(uint32 => uint256) public scoreRootMap;

    // Each batch forged will have a correlated 'exit tree' represented by the exit root
    mapping(uint32 => uint256) public exitRootsMap;

   
    // Map of queues of L1-user-tx transactions, the transactions are stored in bytes32 sequentially
    // The coordinator is forced to forge the next queue in the next L1-L2-batch
    mapping(uint32 => bytes) public mapL1TxQueue;

    // Ethereum block where the last L1-L2-batch was forged
    uint64 public lastL1L2Batch;

    // Queue index that will be forged in the next L1-L2-batch
    uint32 public nextL1ToForgeQueue;

    // Queue index wich will be filled with the following L1-User-Tx
    uint32 public nextL1FillingQueue;

    // Max ethereum blocks after the last L1-L2-batch, when exceeds the timeout only L1-L2-batch are allowed
    uint8 public forgeL1L2BatchTimeout;

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
    function initializeHermez(
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
        // Assure data availability from regular ethereum nodes
        // We include this line because it's easier to track the transaction data, as it will never be in an internal TX.
        // In general this makes no sense, as callling this function from another smart contract will have to pay the calldata twice.
        // But forcing, it avoids having to check.
        require(
            msg.sender == tx.origin,
            "Hermez::forgeBatch: INTENAL_TX_NOT_ALLOWED"
        );

        if (!l1Batch) {
            require(
                block.number < (lastL1L2Batch + forgeL1L2BatchTimeout), // No overflow since forgeL1L2BatchTimeout is an uint8
                "Hermez::forgeBatch: L1L2BATCH_REQUIRED"
            );
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
            // restart the timeout
            lastL1L2Batch = uint64(block.number);
            // clear current queue
            l1UserTxsLen = _clearQueue();
        }

        emit ForgeBatch(lastForgedBatch, l1UserTxsLen);
    }


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
        return l1UserTxsLen;
    }



}
