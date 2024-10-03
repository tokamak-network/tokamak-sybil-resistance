// SPDX-License-Identifier: MIT
pragma solidity ^0.8.23;

import "forge-std/Script.sol";

contract DeployPoseidon is Script {
    // Define an event to log the contract address
    event PoseidonDeployed(address indexed contractAddress);

    function run() external {
        // Start broadcasting the transaction
        vm.startBroadcast();

        // Initialize array with 3 elements
        string[] memory commands = new string[](3); 

        commands[0] = "node";
        commands[1] = "./deployment/deployPoseidon.js";  // Adjust this path if necessary
        
        // Deploy Poseidon with 2 elements
        commands[2] = "2";  // For 2 elements
        bytes memory output2 = vm.ffi(commands);    // Call FFI to deploy Poseidon with 2 elements
        console2.log("FFI Output for 2 elements:", string(output2));

        // Deploy Poseidon with 3 elements
        commands[2] = "3";  // For 3 elements
       bytes memory output3 = vm.ffi(commands);    // Call FFI to deploy Poseidon with 3 elements
        console2.log("FFI Output for 3 elements:", string(output3));

        // Deploy Poseidon with 4 elements
        commands[2] = "4";  // For 4 elements
        bytes memory output4 = vm.ffi(commands);    // Call FFI to deploy Poseidon with 4 elements
        console2.log("FFI Output for 4 elements:", string(output4));

        // Stop broadcasting
        vm.stopBroadcast();

        console2.log("output:", parseAddressFromOutput(output2));
        emit PoseidonDeployed(parseAddressFromOutput(output2));
    }

    function parseAddressFromOutput(bytes memory output) internal pure returns (address) {
        // Remove "0x" prefix if present
        string memory outputString = string(output);
        bytes memory temp = bytes(outputString);
        
        uint160 addr = 0;
        for (uint i = 0; i < 20; i++) {
            addr <<= 8;
            addr += uint160(uint8(temp[i + 2])); // Skipping '0x'
        }
        return address(addr);
    }
}
