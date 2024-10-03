// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Script.sol";

contract DeployPoseidon is Script {
            string[]  commands;
    function run() external {
        // Start broadcasting the transaction
        vm.startBroadcast();

        // Deploy Poseidon with different numbers of elements
        
        commands[0] = "node";
        commands[1] = "deployPoseidon.js";  // Adjust this path if necessary
        commands[2] = "2";  // For 2 elements
        vm.ffi(commands);

        commands[2] = "3";  // For 3 elements
        vm.ffi(commands);

        commands[2] = "4";  // For 4 elements
        vm.ffi(commands);

        // Stop broadcasting
        vm.stopBroadcast();
    }
}
