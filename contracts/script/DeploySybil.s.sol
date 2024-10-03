// SPDX-License-Identifier: MIT
pragma solidity ^0.8.23;

import {DevOpsTools} from "lib/foundry-devops/src/DevOpsTools.sol";
import "forge-std/Script.sol";
import {
    Sybil
} from "../src/sybil.sol";
import {
    VerifierRollupStub
} from "../src/stub/VerifierRollupStub.sol"; // Import the VerifierRollupStub

contract MyScript is Script {
    function run() external {
    address verfier = DevOpsTools.get_most_recent_deployment(
            "VerifierRollupStub",
            block.chainid
        );
        // Declare arrays for verifiers, maxTxs, and nLevels
        address[] memory verifiers = new address[](1);
        uint256[] memory maxTx = new uint256[](1);
        uint256[] memory nLevels = new uint256[](1);

        // Set values for the arrays
        verifiers[0] = address(verifier); // Use the deployed verifier's address
        maxTx[0] = 100; // Set an example maxTxs value
        nLevels[0] = 5; // Set an example nLevels value

        // Specify Poseidon contract addresses
        address poseidon2Elements = 0xb84B26659fBEe08f36A2af5EF73671d66DDf83db; // Replace with actual Poseidon 2 elements contract address
        address poseidon3Elements = 0xFc50367cf2bA87627f99EDD8703FF49252473AED; // Replace with actual Poseidon 3 elements contract address
        address poseidon4Elements = 0xF8AB2781AA06A1c3eF41Bd379Ec1681a70A148e0; // Replace with actual Poseidon 4 elements contract address

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

        console.log("VerifierRollupStub deployed at:", address(verifier));
        console.log("Sybil contract deployed at:", address(sybilContract));
    }
}