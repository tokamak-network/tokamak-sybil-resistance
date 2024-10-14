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
        address;
        uint256;
        uint256;

        // Set values for the arrays
        verifiers[0] = verifier;
        maxTx[0] = 100;
        nLevels[0] = 5;

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
    }
}
