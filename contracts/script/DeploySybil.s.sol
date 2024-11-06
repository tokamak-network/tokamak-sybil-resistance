// SPDX-License-Identifier: MIT
pragma solidity ^0.8.23;

import {DevOpsTools} from "lib/foundry-devops/src/DevOpsTools.sol";
import "forge-std/Script.sol";
import {Sybil} from "../src/sybil.sol";
import {VerifierRollupStub} from "../src/stub/VerifierRollupStub.sol";

contract FunctionScript is Script {
    error VerifierRollupStubNotDeployed();

    function run() external {
        address verifier = DevOpsTools.get_most_recent_deployment(
            "VerifierRollupStub",
            block.chainid
        );

        // Declare arrays for verifiers, maxTxs, and nLevels
        address[] memory verifiers = new address[](1);
        uint256[] memory maxTx = new uint256[](1);
        uint256[] memory nLevels = new uint256[](1);

        // Set values for the arrays
        verifiers[0] = verifier;
        maxTx[0] = 100;
        nLevels[0] = 5;

        // Specify Poseidon contract addresses
        address poseidon2Elements = 0x31c3EBCa9c9eFAeE59FD30A968BCA0634F42Ed95;
        address poseidon3Elements = 0x82c5d2d227b5C6f69A978cfA7025654517e82351;
        address poseidon4Elements = 0xfFe18609E5641527191408BfC5776129037794f2;

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
    }
}
