// SPDX-License-Identifier: AGPL-3.0

pragma solidity 0.6.12;

contract Hermez is InstantWithdrawManager {
    // First 256 indexes reserved, first user index will be the 256
    uint48 constant _RESERVED_IDX = 255;

    // IDX 1 is reserved for exits
    uint48 constant _EXIT_IDX = 1;

    // IDX 2 is reserved for exits
    uint48 constant _EXPLODE_IDX = 2;

    // [20 bytes] fromEthAddr + [32 bytes] fromBjj-compressed + [6 bytes] fromIdx +
    // [5 bytes] loadAmountFloat40 + [5 bytes] amountFloat40 + [6 bytes] toIdx
    uint256 constant _L1_USER_TOTALBYTES = 74;

    // Maximum L1 transactions allowed to be queued in a batch
    uint256 constant _MAX_L1_TX = 128;

    // Modulus zkSNARK
    uint256 constant _RFIELD = 21888242871839275222246405745257275088548364400416034343698204186575808495617;

    // [6 bytes] lastIdx + [6 bytes] newLastIdx  + [32 bytes] stateRoot  + [32 bytes] newStRoot  + [32 bytes] newExitRoot +
    // [_MAX_L1_TX * _L1_USER_TOTALBYTES bytes] l1TxsData + totall1L2TxsDataLength + feeIdxCoordinatorLength + [2 bytes] chainID + [4 bytes] batchNum =
    // 18546 bytes + totall1L2TxsDataLength + feeIdxCoordinatorLength
    uint256 constant _INPUT_SHA_CONSTANT_BYTES = 20082;

    uint8 public constant ABSOLUTE_MAX_L1L2BATCHTIMEOUT = 240;

    // This ethereum address is used internally for rollup accounts that don't have ethereum address, only Babyjubjub
    // This non-ethereum accounts can be created by the coordinator and allow users to have a rollup
    // account without needing an ethereum address
    address constant _ETH_ADDRESS_INTERNAL_ONLY = address(
        0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF
    );

    // Verifiers array
    VerifierRollup[] public rollupVerifiers;

    // Withdraw verifier interface
    VerifierWithdrawInterface public withdrawVerifier;

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

    // Each batch forged will have a correlated 'l1L2TxDataHash'
    mapping(uint32 => bytes32) public l1L2TxsDataHashMap;

    // Mapping of exit nullifiers, only allowing each withdrawal to be made once
    // rootId => (Idx => true/false)
    mapping(uint32 => mapping(uint48 => bool)) public exitNullifierMap;

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

    // Event emitted when a withdrawal is done
    event WithdrawEvent(
        uint48 indexed idx,
        uint32 indexed numExitRoot,
        bool indexed instantWithdraw
    );

    // Event emitted when the contract is initialized
    event InitializeHermezEvent(
        uint8 forgeL1L2BatchTimeout,
        uint256 feeAddToken,
        uint64 withdrawalDelay
    );

    function updateVerifiers() external {
        require(
            msg.sender == address(0xb6D3f1056c015962fA66A4020E50522B58292D1E),
            "Hermez::updateVerifiers ONLY_DEPLOYER"
        );
        require(
            rollupVerifiers[0].maxTx == 344, // Old verifier 344 tx
            "Hermez::updateVerifiers VERIFIERS_ALREADY_UPDATED"
        );
        rollupVerifiers[0] = VerifierRollup({
            verifierInterface: VerifierRollupInterface(
                address(0x3DAa0B2a994b1BC60dB9e312aD0a8d87a1Bb16D2) // New verifier 400 tx
            ),
            maxTx: 400,
            nLevels: 32
        });

        rollupVerifiers[1] = VerifierRollup({
            verifierInterface: VerifierRollupInterface(
                address(0x1DC4b451DFcD0e848881eDE8c7A99978F00b1342) // New verifier 2048 tx
            ),
            maxTx: 2048,
            nLevels: 32
        });

        withdrawVerifier = VerifierWithdrawInterface(
            0x4464A1E499cf5443541da6728871af1D5C4920ca
        );
        emit hermezV2();
    }

    /**
     * @dev Initializer function (equivalent to the constructor). Since we use
     * upgradeable smartcontracts the state vars have to be initialized here.
     */
    function initializeHermez(
        address[] memory _verifiers,
        uint256[] memory _verifiersParams,
        address _withdrawVerifier,
        uint8 _forgeL1L2BatchTimeout,
        address _poseidon2Elements,
        address _poseidon3Elements,
        address _poseidon4Elements,
    ) external initializer {
        require(
            _hermezAuctionContract != address(0) &&
                _withdrawDelayerContract != address(0),
            "Hermez::initializeHermez ADDRESS_0_NOT_VALID"
        );

        // set state variables
        _initializeVerifiers(_verifiers, _verifiersParams);
        withdrawVerifier = VerifierWithdrawInterface(_withdrawVerifier);
        forgeL1L2BatchTimeout = _forgeL1L2BatchTimeout;

        // set default state variables
        lastIdx = _RESERVED_IDX;
        // lastL1L2Batch = 0 --> first batch forced to be L1Batch
        // nextL1ToForgeQueue = 0 --> First queue will be forged
        nextL1FillingQueue = 1;
        // stateRootMap[0] = 0 --> genesis batch will have root = 0

        // initialize libs
        _initializeHelpers(
            _poseidon2Elements,
            _poseidon3Elements,
            _poseidon4Elements
        );
        _initializeWithdraw(
            _hermezGovernanceAddress
        );
        emit InitializeHermezEvent(
            _forgeL1L2BatchTimeout,
            _withdrawalDelay
        );
    }

 function addL1Transaction(
        uint256 babyPubKey,
        uint48 fromIdx,
        uint40 loadAmountF,
        uint40 amountF,
        uint48 toIdx,
        bytes calldata permit
    ) external payable {
        uint256 loadAmount = _float2Fix(loadAmountF);
        require(
            loadAmount == msg.value,
            "Hermez::addL1Transaction: LOADAMOUNT_ETH_DOES_NOT_MATCH"
        );

        // perform L1 User Tx
        _addL1Transaction(
            msg.sender,
            babyPubKey,
            fromIdx,
            loadAmountF,
            amountF,
            toIdx
        );
    }


    function _addL1Transaction(
        address ethAddress,
        uint256 babyPubKey,
        uint48 fromIdx,
        uint40 loadAmountF,
        uint40 amountF,
        uint48 toIdx
    ) internal {
        uint256 amount = _float2Fix(amountF);

        // toIdx can be: 0, _EXIT_IDX , _EXPLODE_IDX or (toIdx > _RESERVED_IDX)
        if (toIdx == 0) {
            require(
                (amount == 0),
                "Hermez::_addL1Transaction: AMOUNT_MUST_BE_0_IF_NOT_TRANSFER"
            );
        } else {
            if ((toIdx == _EXIT_IDX)) { //WHY IS THIS DOUBLE BRACKETTED??
                require(
                    (loadAmountF == 0),
                    "Hermez::_addL1Transaction: LOADAMOUNT_MUST_BE_0_IF_EXIT"
                );
            } else if ((toIdx == _EXPLODE_IDX)) {
                require(
                    (loadAmountF == 0),
                    "Hermez::_addL1Transaction: LOADAMOUNT_MUST_BE_0_IF_EXPLODE"
                )
            } else {
                require(
                    ((toIdx > _RESERVED_IDX) && (toIdx <= lastIdx)),
                    "Hermez::_addL1Transaction: INVALID_TOIDX"
                );
            }
        }

        // fromIdx can be: 0 if create account or (fromIdx > _RESERVED_IDX)
        if (fromIdx == 0) {
            require(
                babyPubKey != 0,
                "Hermez::_addL1Transaction: INVALID_CREATE_ACCOUNT_WITH_NO_BABYJUB"
            );
        } else {
            require(
                (fromIdx > _RESERVED_IDX) && (fromIdx <= lastIdx),
                "Hermez::_addL1Transaction: INVALID_FROMIDX"
            );
            require(
                babyPubKey == 0,
                "Hermez::_addL1Transaction: BABYJUB_MUST_BE_0_IF_NOT_CREATE_ACCOUNT"
            );
        }

        _l1QueueAddTx(
            ethAddress,
            babyPubKey,
            fromIdx,
            loadAmountF,
            amountF,
            toIdx
        );
    }

    /**
     * @dev Add L1-user-tx, add it to the correspoding queue
     * l1Tx L1-user-tx encoded in bytes as follows: [20 bytes] fromEthAddr || [32 bytes] fromBjj-compressed || [4 bytes] fromIdx ||
     * [5 bytes] loadAmountFloat40 || [5 bytes] amountFloat40 || [4 bytes] tokenId || [4 bytes] toIdx
     * @param ethAddress Ethereum address of the rollup account
     * @param babyPubKey Public key babyjubjub represented as point: sign + (Ay)
     * @param fromIdx Index account of the sender account
     * @param loadAmountF Amount from L1 to L2
     * @param amountF  Amount transfered between L2 accounts
     * @param tokenID  Token identifier
     * @param toIdx Index leaf of recipient account
     * Events: `L1UserTxEvent`
     */
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
        if (currentPosition + 1 >= _MAX_L1_USER_TX) {
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
