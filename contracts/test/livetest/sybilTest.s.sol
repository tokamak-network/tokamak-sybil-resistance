// SPDX-License-Identifier: MIT
pragma solidity ^0.8.23;

import {DevOpsTools} from "../../lib/foundry-devops/src/DevOpsTools.sol";
import "forge-std/Script.sol";
import {Sybil} from "../../src/sybil.sol";
import "../_helpers/constants.sol"; // Import TestHelpers
import "../_helpers/transactionTypes.sol"; // Import TransactionTypeHelper

contract CallFunctions is Script {
    error VerifierRollupStubNotDeployed();
    TransactionTypeHelper public transactionTypeHelper;
    TestHelpers public testHelpers; // Declare TestHelpers instance

    function run() external {
        // Instantiate the helper contracts
        transactionTypeHelper = new TransactionTypeHelper(); 
        testHelpers = new TestHelpers(); // Instantiate TestHelpers

        address verifier = DevOpsTools.get_most_recent_deployment(
            "VerifierRollupStub",
            block.chainid
        );

        // Properly declare and initialize the arrays
        address verifiers;  // Declare verifiers array
        uint256 maxTx;      // Declare maxTx array
        uint256 nLevels;    // Declare nLevels array

        // Set values for the arrays
        verifiers[0] = verifier;    // Set the verifier address in the array
        maxTx[0] = 100;             // Set maxTx value in the array
        nLevels[0] = 5;             // Set nLevels value in the array

        // Specify Poseidon contract addresses
        address poseidon2Elements = 0xb84B26659fBEe08f36A2af5EF73671d66DDf83db;
        address poseidon3Elements = 0xFc50367cf2bA87627f99EDD8703FF49252473AED;
        address poseidon4Elements = 0xF8AB2781AA06A1c3eF41Bd379Ec1681a70A148e0;

        vm.startBroadcast();
        // Deploy the Sybil contract
        Sybil sybilContract = new Sybil(
            verifiers,
            maxTx,
            nLevels,
            240,
            poseidon2Elements,
            poseidon3Elements,
            poseidon4Elements
        );

        vm.stopBroadcast();

        console2.log("Sybil contract deployed at:", address(sybilContract));

        // Check if the Sybil contract address was retrieved successfully
        require(
            address(sybilContract) != address(0),
            "Sybil contract not found"
        );

        vm.startBroadcast();

        // **Call `addL1Transaction` function**
        TransactionTypeHelper.TxParams memory txParams = transactionTypeHelper.validDeposit();
        uint256 loadAmount = testHelpers.toWei(txParams.loadAmountF); // Convert using TestHelpers

        // Now using txParams values to call the addL1Transaction function
        sybilContract.addL1Transaction{value: loadAmount}(
            txParams.babyPubKey,
            txParams.fromIdx,
            txParams.loadAmountF,
            txParams.amountF,
            txParams.toIdx
        );

        console2.log("Called addL1Transaction successfully.");

        // **Call `forgeBatch` function**
        uint48 newLastIdx = 256;
        uint256 newStRoot = 0xabc;
        uint256 newVouchRoot = 0;
        uint256 newScoreRoot = 0;
        uint256 newExitRoot = 0;
        uint8 verifierIdx = 0;
        bool l1Batch = true;
        uint256[2] memory proofA = [uint256(0), uint256(0)];
        uint256[2][2] memory proofB = [
            [uint256(0), uint256(0)],
            [uint256(0), uint256(0)]
        ];
        uint256[2] memory proofC = [uint256(0), uint256(0)];
        uint256 input = 0;

        sybilContract.forgeBatch(
            newLastIdx,
            newStRoot,
            newVouchRoot,
            newScoreRoot,
            newExitRoot,
            verifierIdx,
            l1Batch,
            proofA,
            proofB,
            proofC,
            input
        );

        console2.log("Called forgeBatch successfully.");

        // Add L1 transactions to queue
        sybilContract.addL1Transaction{value: loadAmount}(
            txParams.babyPubKey,
            txParams.fromIdx,
            txParams.loadAmountF,
            txParams.amountF,
            txParams.toIdx
        );

        // Forge a batch with L1 transactions
        sybilContract.forgeBatch(
            newLastIdx,
            newStRoot,
            newVouchRoot,
            newScoreRoot,
            newExitRoot,
            verifierIdx,
            true,
            proofA,
            proofB,
            proofC,
            input
        );

        console2.log(
            "Batch with mixed L1 and L2 transactions processed successfully."
        );

        // Set the timeout to the maximum allowed value (ABSOLUTE_MAX_L1L2BATCHTIMEOUT)
        sybilContract.setForgeL1L2BatchTimeout(
            sybilContract.ABSOLUTE_MAX_L1L2BATCHTIMEOUT()
        );

        console2.log("Batch timeout set to maximum successfully.");

        // Set the timeout to a lower value and verify change
        sybilContract.setForgeL1L2BatchTimeout(120);
        console2.log("Batch timeout set to 120 blocks successfully.");

        // **Call `withdrawMerkleProof` function**
        uint192 amount = testHelpers.ONE_ETHER;
        uint256 withdrawBabyPubKey = 0x123456789abcdef;
        uint32 numExitRoot = 1;
        uint48 idx = 0;
        uint256[] memory siblings;

        try
            sybilContract.withdrawMerkleProof(
                amount,
                withdrawBabyPubKey,
                numExitRoot,
                siblings,
                idx
            )
        {
            console2.log("Called withdrawMerkleProof successfully.");
        } catch Error(string memory reason) {
            console2.log("withdrawMerkleProof failed:", reason);
        } catch (bytes memory) {
            console2.log("withdrawMerkleProof failed with unknown error.");
        }

        // **Call getter functions**
        uint256 stateRoot = sybilContract.getStateRoot(1);
        console2.log("State root for batch 1:", stateRoot);

        uint32 lastForgedBatch = sybilContract.getLastForgedBatch();
        console2.log("Last forged batch:", lastForgedBatch);

        bytes memory l1TxQueue = sybilContract.getL1TransactionQueue(1);
        console2.log(
            "L1 Transaction Queue for index 1 length:",
            l1TxQueue.length
        );

        uint32 queueLength = sybilContract.getQueueLength();
        console2.log("Current queue length:", queueLength);

        vm.stopBroadcast();
    }
}
