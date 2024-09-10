// SPDX-License-Identifier: MIT
pragma solidity 0.8.23;

import "forge-std/Test.sol";
import "../src/Sybil.sol";

contract SybilTest is Test {
    Sybil public sybil;

    function setUp() public {}

    function testInitialization() public {}

    function testForgeBatch() public {}

    function testSetForgeL1L2BatchTimeout() public {}
    function testFloat2Fix() public {}
    function testL1QueueAddTx() public {}

    function testGetStateRoot() external {}

    function testGetLastForgedBatch() external {}

    function testGetL1TransactionQueue() external {}

    function testGetQueueLength() external {}
}
