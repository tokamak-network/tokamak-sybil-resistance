// SPDX-License-Identifier: MIT
pragma solidity ^0.8.23;

import {DevOpsTools} from "../../lib/foundry-devops/src/DevOpsTools.sol";
import "forge-std/Script.sol";
import {Sybil} from "../../src/sybil.sol";

contract CallFunctions is Script {
    function run() external {
        // Retrieve the deployed Sybil contract address
        address sybilAddress = DevOpsTools.get_most_recent_deployment(
            "Sybil",
            block.chainid
        );

        // Check if the Sybil contract address was retrieved successfully
        require(sybilAddress != address(0), "Sybil contract not found");

        Sybil sybilContract = Sybil(sybilAddress);

        vm.startBroadcast();

        // **Call `addL1Transaction` function**
        uint256 babyPubKey = 0x123456789abcdef; 
        uint48 fromIdx = 0; 
        uint40 loadAmountF = 100; 
        uint40 amountF = 0;
        uint48 toIdx = 0;

        uint256 loadAmount = uint256(loadAmountF) * 10 ** (18 - 8);

        sybilContract.addL1Transaction{value: loadAmount}(
            babyPubKey,
            fromIdx,
            loadAmountF,
            amountF,
            toIdx
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

        // **Call `setForgeL1L2BatchTimeout` function**
        sybilContract.setForgeL1L2BatchTimeout(120);

        console2.log("Called setForgeL1L2BatchTimeout successfully.");

        // **Call `withdrawMerkleProof` function**
        uint192 amount = 1 ether;
        uint256 withdrawBabyPubKey = 0x123456789abcdef; 
        uint32 numExitRoot = 1;
        uint48 idx = 0;
        uint256[] memory siblings;

        try sybilContract.withdrawMerkleProof(
            amount,
            withdrawBabyPubKey,
            numExitRoot,
            siblings,
            idx
        ) {
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
        console2.log("L1 Transaction Queue for index 1 length:", l1TxQueue.length);

        uint32 queueLength = sybilContract.getQueueLength();
        console2.log("Current queue length:", queueLength);

        vm.stopBroadcast();
    }
}
