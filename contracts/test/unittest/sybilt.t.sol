// SPDX-License-Identifier: MIT
pragma solidity 0.8.23;

import "forge-std/Test.sol";
import "../../src/Sybil.sol";
import  "../_helpers/constants.sol";


contract SybilTest is Test {
    Sybil public sybil;

    function setUp() public {
        // Deploy a new instance of the Sybil contract and initialize it
        sybil = new Sybil();
        sybil.initialize(120); // Initialize with a timeout of 120 blocks
    }

    // Test for getStateRoot
    function testGetStateRoot() external {
        uint32 batchNum = 1;
        uint256[2] memory proofA = [uint(0),uint(0)];
        uint256[2][2] memory proofB = [[uint(0), uint(0)], [uint(0), uint(0)]];
        uint256[2] memory proofC = [uint(0), uint(0)];

        // Simulate setting a state root in the contract
        vm.prank(address(this));
        sybil.forgeBatch(
            256, 
            0xabc, 
            0, 
            0, 
            0, 
            0, 
            false, 
            proofA,
            proofB,
            proofC
        );

        // Retrieve the state root for batch 1 and assert it's correct
        uint256 stateRoot = sybil.getStateRoot(batchNum);
        assertEq(stateRoot, 0xabc, "State root should match the one set during forgeBatch");
    }

    // Test for getLastForgedBatch
    function testGetLastForgedBatch() external {
        // Initially the last forged batch should be 0
        uint32 lastForged = sybil.getLastForgedBatch();
        assertEq(lastForged, 0, "Initially, the last forged batch should be 0");

        uint256[2] memory proofA = [uint(0),uint(0)];
        uint256[2][2] memory proofB = [[uint(0), uint(0)], [uint(0), uint(0)]];
        uint256[2] memory proofC = [uint(0), uint(0)];

        // After forging a batch, the last forged batch should increment
        vm.prank(address(this));
        sybil.forgeBatch(
            256, 
            0xabc, 
            0, 
            0, 
            0, 
            0, 
            false, 
            proofA,
            proofB,
            proofC
        );

        lastForged = sybil.getLastForgedBatch();
        assertEq(lastForged, 1, "Last forged batch should increment after forgeBatch call");
    }

function testGetL1TransactionQueue() external {
    // Set up a valid Deposit transaction
    uint256 babyPubKey = 2; // Deposit transactions should have babyPubKey = 2
    uint48 fromIdx = 0; // Ensure fromIdx is > _RESERVED_IDX and <= lastIdx
    uint40 loadAmountF = 100;
    uint40 amountF = 0; // Deposit transactions should have amountF = 0
    uint48 toIdx = 0; // Deposit transactions should have toIdx = 0

    // Compute the correct value to be sent with the transaction
    uint256 loadAmount = (loadAmountF) * 10 ** (18 - 8);

    // Add a transaction to the queue
    vm.prank(address(this));
    sybil.addL1Transaction{value: loadAmount}(babyPubKey, fromIdx, loadAmountF, amountF, toIdx);

    // Retrieve the queue data and check the encoding
    bytes memory txData = sybil.getL1TransactionQueue(1);
    bytes memory expectedTxData = abi.encodePacked(
        address(this), babyPubKey, fromIdx, loadAmountF, amountF, toIdx
    );

    assertEq(txData, expectedTxData, "Transaction data in queue should match expected encoding");
}

function testGetQueueLength() external {
    // Initially the queue length should be 1
    uint32 queueLength = sybil.getQueueLength();
    assertEq(queueLength, 1, "Initially, the queue length should be 1 (due to initialization)");

    uint256 babyPubKey = 2; // Deposit transactions should have babyPubKey = 2
    uint48 fromIdx = 0; // Ensure fromIdx is > _RESERVED_IDX and <= lastIdx
    uint40 loadAmountF = 100;
    uint40 amountF = 0; // Deposit transactions should have amountF = 0
    uint48 toIdx = 0; // Deposit transactions should have toIdx = 0

    // Compute the correct value to be sent with the transaction
    uint256 loadAmount = (loadAmountF) * 10 ** (18 - 8);

    vm.prank(address(this));
    sybil.addL1Transaction{value: loadAmount}(babyPubKey, fromIdx, loadAmountF, amountF, toIdx);

    queueLength = sybil.getQueueLength();
    assertEq(queueLength, 1, "Queue length should still be 1 after adding a transaction");
}
    function testClearQueue() public {
        uint256 babyPubKey = 2; // Deposit transactions should have babyPubKey = 2
        uint48 fromIdx = 0; // Ensure fromIdx is > _RESERVED_IDX and <= lastIdx
        uint40 loadAmountF = 100;
        uint40 amountF = 0; // Deposit transactions should have amountF = 0
        uint48 toIdx = 0; // Deposit transactions should have toIdx = 0

        uint256 loadAmount = (loadAmountF) * 10 ** (18 - 8);

        // Add some transactions to the queue
        vm.prank(address(this));
        sybil.addL1Transaction{value: loadAmount}(babyPubKey, fromIdx, loadAmountF, amountF, toIdx);
        sybil.addL1Transaction{value: loadAmount}(babyPubKey, fromIdx, loadAmountF, amountF, toIdx);

        uint256[2] memory proofA = [uint(0),uint(0)];
        uint256[2][2] memory proofB = [[uint(0), uint(0)], [uint(0), uint(0)]];
        uint256[2] memory proofC = [uint(0), uint(0)];

        // Forge a batch and clear the queue
        vm.prank(address(this));
        sybil.forgeBatch(
            256, 
            0xabc, 
            0, 
            0, 
            0, 
            0, 
            true, 
            proofA,
            proofB,
            proofC
        );

        // Verify queue length has been cleared
        uint32 queueAfter = sybil.getQueueLength();
        assertEq(queueAfter, 1, "Queue should be cleared after forgeBatch");
    }
}
