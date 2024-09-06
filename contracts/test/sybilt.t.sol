// SPDX-License-Identifier: MIT
pragma solidity 0.8.23;

import "forge-std/Test.sol"; 
import "../src/sybil.sol"; 

contract SybilVerifierTest is Test {
    SybilVerifier public sybilVerifier;

    address[] verifiers;
    uint256[] verifiersParams;
    address withdrawVerifier = address(0x123); // Mock withdraw verifier
    address tokenHEZ = address(0x456); // Mock HEZ token
    uint8 forgeL1L2BatchTimeout = 60;
    uint256 feeAddToken = 1e8;
    address poseidon2Elements = address(0x789); // Mock Poseidon elements
    address poseidon3Elements = address(0xabc); // Mock Poseidon elements
    address poseidon4Elements = address(0xdef); // Mock Poseidon elements
    address sybilGovernanceAddress = address(0x987); // Mock governance address
    uint64 withdrawalDelay = 100;
    address withdrawDelayerContract = address(0x654); // Mock withdraw delayer contract

    function setUp() public {
        // Initialize the contract
        sybilVerifier = new SybilVerifier();
        sybilVerifier.initializeSybilVerifier(
            verifiers,
            verifiersParams,
            withdrawVerifier,
            tokenHEZ,
            forgeL1L2BatchTimeout,
            feeAddToken,
            poseidon2Elements,
            poseidon3Elements,
            poseidon4Elements,
            sybilGovernanceAddress,
            withdrawalDelay,
            withdrawDelayerContract
        );
    }

    function testInitializeSybilVerifier() public {
        // Test that contract is initialized properly
        assertEq(sybilVerifier.withdrawVerifier(), withdrawVerifier);
        assertEq(sybilVerifier.tokenHEZ(), tokenHEZ);
        assertEq(sybilVerifier.forgeL1L2BatchTimeout(), forgeL1L2BatchTimeout);
        assertEq(sybilVerifier.feeAddToken(), feeAddToken);
    }

    function testAddL1Transaction_CreateAccount() public {
        uint256 babyPubKey = 123456;
        uint48 fromIdx = 0;
uint40 loadAmountF = 0; 

        uint40 amountF = 0;
        uint48 toIdx = 0;

        // Expect revert if babyPubKey is 0
        vm.expectRevert();
        sybilVerifier.addL1Transaction(0, fromIdx, loadAmountF, amountF, toIdx);

        // Should succeed when babyPubKey is not 0 and loadAmountF, amountF are 0
        sybilVerifier.addL1Transaction{value: 0}(babyPubKey, fromIdx, loadAmountF, amountF, toIdx);
    }

    function testAddL1Transaction_CreateAccountDeposit() public {
        uint256 babyPubKey = 123456;
        uint48 fromIdx = 0;
        uint40 loadAmountF =  1e8; 
        uint40 amountF = 0;
        uint48 toIdx = 0;

        // Expect revert if loadAmount does not match msg.value
        vm.expectRevert();
        sybilVerifier.addL1Transaction{value: 0.5 ether}(babyPubKey, fromIdx, loadAmountF, amountF, toIdx);
        vm.expectRevert();
        sybilVerifier.addL1Transaction{value: 1e8}(babyPubKey, fromIdx, loadAmountF, amountF, toIdx);
    }

    function testAddL1Transaction_Deposit() public {
        uint48 fromIdx = 256; // Reserved index
        uint40 loadAmountF = 1e8;
        uint40 amountF = 0;
        uint48 toIdx = 0;
        uint256 babyPubKey = 0;

        // Should succeed for deposit
        sybilVerifier.addL1Transaction{value: 1e8}(babyPubKey, fromIdx, loadAmountF, amountF, toIdx);
    }

    function testAddL1Transaction_ForceExit() public {
        uint48 fromIdx = 256; // Reserved index
        uint40 loadAmountF = 0;
        uint40 amountF = 1e8;
        uint48 toIdx = 1;
        uint256 babyPubKey = 0;

        // Expect revert if amountF is 0 for ForceExit
        vm.expectRevert();
        sybilVerifier.addL1Transaction(babyPubKey, fromIdx, loadAmountF, 0, toIdx);

        // Should succeed for ForceExit
        sybilVerifier.addL1Transaction(babyPubKey, fromIdx, loadAmountF, amountF, toIdx);
    }

    function testAddL1Transaction_ForceExplode() public {
        uint48 fromIdx = 256; // Reserved index
        uint40 loadAmountF = 0;
        uint40 amountF = 0;
        uint48 toIdx = 2;
        uint256 babyPubKey = 0;

        // Should succeed for ForceExplode
        sybilVerifier.addL1Transaction(babyPubKey, fromIdx, loadAmountF, amountF, toIdx);
    }

    function testForgeBatch() public {
        uint48 newLastIdx = 300;
        uint256 newStRoot = 123456789;
        uint256 newExitRoot = 987654321;
        bytes memory encodedL1CoordinatorTx = "";
        bytes memory l1L2TxsData = "";
        bytes memory feeIdxCoordinator = "";
        uint8 verifierIdx = 0;
        bool l1Batch = true;

        uint256[2] memory proofA;
        uint256[2][2] memory proofB;
        uint256[2] memory proofC;
        //TODO change this for testing..
        //internal transaction not allowed
        vm.expectRevert();

        // Simulate batch forging
        sybilVerifier.forgeBatch(
            newLastIdx,
            newStRoot,
            newExitRoot,
            encodedL1CoordinatorTx,
            l1L2TxsData,
            feeIdxCoordinator,
            verifierIdx,
            l1Batch,
            proofA,
            proofB,
            proofC
        );

        // Check that batch is forged correctly
        // assertEq(sybilVerifier.getStateRoot(sybilVerifier.getLastForgedBatch()), newStRoot);

    }
}
