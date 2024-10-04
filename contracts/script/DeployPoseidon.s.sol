// SPDX-License-Identifier: MIT
pragma solidity ^0.8.23;

import "forge-std/Script.sol";

contract DeployPoseidon is Script {

    function run() external {
        vm.startBroadcast();

        string[] memory commands = new string[](3); 

        commands[0] = "node";
        commands[1] = "./deployment/deployPoseidon.js"; 
        
        // Deploy Poseidon with 2 elements
        commands[2] = "2";  // For 2 elements
        bytes memory output2 = vm.ffi(commands);    // Call FFI to deploy Poseidon with 2 elements
        address addr2 = bytesToAddress(output2);
        console.log("Poseidon Contract with 2 elements deployed at:", addr2);

        // Deploy Poseidon with 3 elements
        commands[2] = "3";  // For 3 elements
       bytes memory output3 = vm.ffi(commands);    // Call FFI to deploy Poseidon with 3 elements
        address addr3 = bytesToAddress(output3);
        console.log("Poseidon Contract with 2 elements deployed at:", addr3);

        // Deploy Poseidon with 4 elements
        commands[2] = "4";  // For 4 elements
        bytes memory output4 = vm.ffi(commands);    // Call FFI to deploy Poseidon with 4 elements
        address addr4 = bytesToAddress(output4);
        console.log("Poseidon Contract with 2 elements deployed at:", addr4);

        // Stop broadcasting
        vm.stopBroadcast();
    }

    // Helper function to convert bytes to an Ethereum address
    function bytesToAddress(bytes memory b) internal pure returns (address) {
        require(b.length >= 20, "Bytes array too short to be an address");
        address addr;
        assembly {
            addr := mload(add(b, 20))  // Load 20 bytes (address length)
        }
        return addr;
    }
}
