// SPDX-License-Identifier: MIT
pragma solidity 0.8.23;

contract TestHelpers {
    // Constant for a zero address
    address public constant ZERO_ADDRESS = address(0);

    // Constant for zero amount
    uint256 public constant ZERO_AMOUNT = 0;

    // Constant for 1 ether in wei (1e18)
    uint256 internal constant ONE_ETHER = 1e18;

    // Ether value
    function toEther(uint256 amount) public pure returns (uint256) {
        return amount * 1e18;
    }

    // Utility to return a zero address
    function getZeroAddress() public pure returns (address) {
        return address(0);
    }

    function toWei(uint256 ethAmount) public pure returns (uint256) {
        return ethAmount * 1e18;
    }

    function isValidAddress(address addr) public pure returns (bool) {
        return addr != address(0);
    }

    function exceedsLimit(
        uint256 amount,
        uint256 limit
    ) public pure returns (bool) {
        return amount > limit;
    }

    // Utility to return zero amount
    function getZeroAmount() public pure returns (uint256) {
        return 0;
    }

    // Utility to convert amount to 1e18 (1 Ether)
    function getOneEther() public pure returns (uint256) {
        return 1e18;
    }
    // Convert GWei to Ether
    function gweiToEther(uint256 gweiAmount) public pure returns (uint256) {
        return gweiAmount * 1e9;
    }

    // Convert Ether to GWei
    function etherToGwei(uint256 etherAmount) public pure returns (uint256) {
        return etherAmount / 1e9;
    }


    // Convert a custom floating point number to fixed point
    function float2Fix(uint256 floatVal) public pure returns (uint256) {
        return floatVal * 1e18;
    }
}
