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

         // Write the contract address to a JSON file
        // string memory json = string(abi.encodePacked('{"2_elements": "', toString(addr2), '"}'));
        // vm.writeFile("broadcast/DeployPoseidon.s.sol/deployments.json", json);

        // Deploy Poseidon with 3 elements
        commands[2] = "3";  // For 3 elements
        bytes memory output3 = vm.ffi(commands);    // Call FFI to deploy Poseidon with 3 elements
        address addr3 = bytesToAddress(output3);
        console.log("Poseidon Contract with 3 elements deployed at:", addr3);

        // Deploy Poseidon with 4 elements
        commands[2] = "4";  // For 4 elements
        bytes memory output4 = vm.ffi(commands);    // Call FFI to deploy Poseidon with 4 elements
        address addr4 = bytesToAddress(output4);
        console.log("Poseidon Contract with 4 elements deployed at:", addr4);

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

    // Helper function to convert address to string
    // function toString(address addr) internal pure returns (string memory) {
    //     bytes32 value = bytes32(uint256(uint160(addr)));
    //     bytes memory alphabet = "0123456789abcdef";
    //     bytes memory str = new bytes(42);
    //     str[0] = '0';
    //     str[1] = 'x';
    //     for (uint256 i = 0; i < 20; i++) {
    //         str[2 + i * 2] = alphabet[uint8(value[i + 12] >> 4)];
    //         str[3 + i * 2] = alphabet[uint8(value[i + 12] & 0x0f)];
    //     }
    //     return string(str);
    // }
}
