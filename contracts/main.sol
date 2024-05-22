// SPDX-License-Identifier: GPL-3.0

pragma solidity >=0.8.2 <0.9.0;

contract Sybil { 
    // Last account index created inside the rollup
    uint48 public lastIdx;

    // Last batch forged
    uint32 public lastForgedBatch;

    // Each batch forged will have a correlated 'state root'
    mapping(uint32 => uint256) public stateRootMap;

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




    //////////////
    // Coordinator operations
    /////////////

    /**
     * @dev Forge a new batch providing the L2 Transactions, L1Corrdinator transactions and the proof.
     * If the proof is succesfully verified, update the current state, adding a new state and exit root.
     * In order to optimize the gas consumption the parameters `encodedL1CoordinatorTx`, `l1L2TxsData` and `feeIdxCoordinator`
     * are read directly from the calldata using assembly with the instruction `calldatacopy`
     * @param newLastIdx New total rollup accounts
     * @param newStRoot New state root
     * @param newExitRoot New exit root
     * @param encodedL1CoordinatorTx Encoded L1-coordinator transactions
     * @param l1L2TxsData Encoded l2 data
     * @param verifierIdx Verifier index
     * @param l1Batch Indicates if this batch will be L2 or L1-L2
     * @param proofA zk-snark input
     * @param proofB zk-snark input
     * @param proofC zk-snark input
     * Events: `ForgeBatch`
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
    ) external virtual {
        // Assure data availability from regular ethereum nodes
        // We include this line because it's easier to track the transaction data, as it will never be in an internal TX.
        // In general this makes no sense, as callling this function from another smart contract will have to pay the calldata twice.
        // But forcing, it avoids having to check.
        require(
            msg.sender == tx.origin,
            "Hermez::forgeBatch: INTENAL_TX_NOT_ALLOWED"
        );

        // ask the auction if this coordinator is allow to forge
        require(
            hermezAuctionContract.canForge(msg.sender, block.number) == true,
            "Hermez::forgeBatch: AUCTION_DENIED"
        );

        if (!l1Batch) {
            require(
                block.number < (lastL1L2Batch + forgeL1L2BatchTimeout), // No overflow since forgeL1L2BatchTimeout is an uint8
                "Hermez::forgeBatch: L1L2BATCH_REQUIRED"
            );
        }

        // calculate input
        uint256 input = _constructCircuitInput(
            newLastIdx,
            newStRoot,
            newExitRoot,
            l1Batch,
            verifierIdx
        );

        // verify proof
        require(
            rollupVerifiers[verifierIdx].verifierInterface.verifyProof(
                proofA,
                proofB,
                proofC,
                [input]
            ),
            "Hermez::forgeBatch: INVALID_PROOF"
        );

        // update state
        lastForgedBatch++;
        lastIdx = newLastIdx;
        stateRootMap[lastForgedBatch] = newStRoot;
        exitRootsMap[lastForgedBatch] = newExitRoot;
        l1L2TxsDataHashMap[lastForgedBatch] = sha256(l1L2TxsData);

        uint16 l1UserTxsLen;
        if (l1Batch) {
            // restart the timeout
            lastL1L2Batch = uint64(block.number);
            // clear current queue
            l1UserTxsLen = _clearQueue();
        }

        // auction must be aware that a batch is being forged
        hermezAuctionContract.forge(msg.sender);

        emit ForgeBatch(lastForgedBatch, l1UserTxsLen);
    }




}
