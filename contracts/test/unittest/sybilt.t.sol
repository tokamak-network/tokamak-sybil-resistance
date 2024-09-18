// SPDX-License-Identifier: MIT
pragma solidity 0.8 .23;

import "forge-std/Test.sol";
import "../../src/Sybil.sol";
import "../_helpers/constants.sol";
import "../_helpers/transactionTypes.sol";
import "../../src/sybilHelpers.sol";
import "../../src/stub/VerifierRollupStub.sol";

contract SybilTest is Test, TestHelpers, TransactionTypeHelper {
    Sybil public sybil;

    function setUp() public {
        // Deploy Poseidon contracts
        PoseidonUnit2 mockPoseidon2 = new PoseidonUnit2();
        PoseidonUnit3 mockPoseidon3 = new PoseidonUnit3();
        PoseidonUnit4 mockPoseidon4 = new PoseidonUnit4();
        emit log_address(address(mockPoseidon2));
        emit log_address(address(mockPoseidon3));
        emit log_address(address(mockPoseidon4));

        // Deploy verifier stub
        VerifierRollupStub verifierStub = new VerifierRollupStub(); 

        address[] memory verifiers = new address[](1);
        uint256[] memory maxTx = new uint256[](1);
        uint256[] memory nLevels = new uint256[](1);

        verifiers[0] = address(verifierStub);
        maxTx[0] = uint(256);
        nLevels[0] = uint(1);

        // Initialize the Sybil contract with mock Poseidon addresses
        sybil = new Sybil(
            verifiers, 
            maxTx, 
            nLevels, 
            120, 
            address(mockPoseidon2), 
            address(mockPoseidon3), 
            address(mockPoseidon4)
        );
    }

    // Forge batch tests
    function testGetStateRoot() external {
        uint32 batchNum = 1;
        uint256 input = uint(1);
        uint256[2] memory proofA = [uint(0),uint(0)];
        uint256[2][2] memory proofB = [[uint(0), uint(0)], [uint(0), uint(0)]];
        uint256[2] memory proofC = [uint(0), uint(0)];

        vm.prank(address(this));
        sybil.forgeBatch(
            256, 
            0xabc, 
            0, 
            0, 
            0, 
            0, 
            false, 
            proofA,
            proofB,
            proofC,
            input
        );

        uint256 stateRoot = sybil.getStateRoot(batchNum);
        assertEq(stateRoot, 0xabc);
    }

    function testGetLastForgedBatch() external {
        uint32 lastForged = sybil.getLastForgedBatch();
        assertEq(lastForged, 0);

        uint256[2] memory proofA = [uint(0),uint(0)];
        uint256[2][2] memory proofB = [[uint(0), uint(0)], [uint(0), uint(0)]];
        uint256[2] memory proofC = [uint(0), uint(0)];
        uint256 input = uint(1);

        vm.prank(address(this));
        sybil.forgeBatch(
            256, 
            0xabc, 
            0, 
            0, 
            0, 
            0, 
            false, 
            proofA,
            proofB,
            proofC,
            input
        );

        lastForged = sybil.getLastForgedBatch();
        assertEq(lastForged, 1);
    }

    // L1 transactions tests
    function testGetL1TransactionQueue() external {
        uint256 babyPubKey = 2;
        uint48 fromIdx = 0;
        uint40 loadAmountF = 100;
        uint40 amountF = 0;
        uint48 toIdx = 0;

        uint256 loadAmount = (loadAmountF) * 10 ** (18 - 8);

        vm.prank(address(this));
        sybil.addL1Transaction {
            value: loadAmount
        }(babyPubKey, fromIdx, loadAmountF, amountF, toIdx);

        bytes memory txData = sybil.getL1TransactionQueue(1);
        bytes memory expectedTxData = abi.encodePacked(address(this), babyPubKey, fromIdx, loadAmountF, amountF, toIdx);
        assertEq(txData, expectedTxData);
    }

    function testGetQueueLength() external {
        uint32 queueLength = sybil.getQueueLength();
        assertEq(queueLength, 1);

        TxParams memory params = valid();
        uint256 loadAmount = (params.loadAmountF) * 10 ** (18 - 8);

        vm.prank(address(this));
        sybil.addL1Transaction {
            value: loadAmount
        }(params.babyPubKey, params.fromIdx, params.loadAmountF, params.amountF, params.toIdx);

        queueLength = sybil.getQueueLength();
        assertEq(queueLength, 1);
    }

    function testClearQueue() public {
        TxParams memory params = valid();
        uint256 loadAmount = (params.loadAmountF) * 10 ** (18 - 8);

        vm.prank(address(this));
        sybil.addL1Transaction {
            value: loadAmount
        }(params.babyPubKey, params.fromIdx, params.loadAmountF, params.amountF, params.toIdx);
        sybil.addL1Transaction {
            value: loadAmount
        }(params.babyPubKey, params.fromIdx, params.loadAmountF, params.amountF, params.toIdx);

        uint256[2] memory proofA = [uint(0),uint(0)];
        uint256[2][2] memory proofB = [[uint(0), uint(0)], [uint(0), uint(0)]];
        uint256[2] memory proofC = [uint(0), uint(0)];
        uint256 input = uint(1);

        vm.prank(address(this));
        sybil.forgeBatch(
            256, 
            0xabc, 
            0, 
            0, 
            0, 
            0, 
            false, 
            proofA,
            proofB,
            proofC,
            input
        );

        uint32 queueAfter = sybil.getQueueLength();
        assertEq(queueAfter, 1);
    }

    // Events tests
    function testForgeBatchEventEmission() public {
        vm.expectEmit(true, true, true, true);
        emit Sybil.ForgeBatch(1, 0);

        uint256[2] memory proofA = [uint(0),uint(0)];
        uint256[2][2] memory proofB = [[uint(0), uint(0)], [uint(0), uint(0)]];
        uint256[2] memory proofC = [uint(0), uint(0)];
        uint256 input = uint(1);

        vm.prank(address(this));
        sybil.forgeBatch(
            256, 
            0xabc, 
            0, 
            0, 
            0, 
            0, 
            true, 
            proofA,
            proofB,
            proofC,
            input
        );
    }

    function testL1UserTxEventEmission() public {
        TxParams memory params = valid();
        uint256 loadAmount = (params.loadAmountF) * 10 ** (18 - 8);

        vm.expectEmit(true, true, true, true);
        emit Sybil.L1UserTxEvent(1, 0, abi.encodePacked(address(this), params.babyPubKey, params.fromIdx, params.loadAmountF, params.amountF, params.toIdx));

        vm.prank(address(this));
        sybil.addL1Transaction {
            value: loadAmount
        }(params.babyPubKey, params.fromIdx, params.loadAmountF, params.amountF, params.toIdx);
    }

    function testInitializeEventEmission() public {
        PoseidonUnit2 mockPoseidon2 = new PoseidonUnit2();
        PoseidonUnit3 mockPoseidon3 = new PoseidonUnit3();
        PoseidonUnit4 mockPoseidon4 = new PoseidonUnit4();

        // Deploy verifier stub
        VerifierRollupStub verifierStub = new VerifierRollupStub(); 
        
        address[] memory verifiers = new address[](1);
        uint256[] memory maxTx = new uint256[](1);
        uint256[] memory nLevels = new uint256[](1);

        verifiers[0] = address(verifierStub);
        maxTx[0] = uint(256);
        nLevels[0] = uint(1);

        Sybil newSybil = new Sybil(
            verifiers, 
            maxTx, 
            nLevels, 
            120, 
            address(mockPoseidon2), 
            address(mockPoseidon3), 
            address(mockPoseidon4)
        );
        emit Sybil.Initialize(120);
    }

    // CreateAccount transactions tests
    function testCreateAccountTransaction() public {
        TxParams memory params = validCreateAccount();
        uint256 loadAmount = (params.loadAmountF) * 10 ** (18 - 8);

        vm.prank(address(this));
        sybil.addL1Transaction {
            value: loadAmount
        }(params.babyPubKey, params.fromIdx, params.loadAmountF, params.amountF, params.toIdx);
    }

    function testInvalidCreateAccountTransaction() public {
        TxParams memory params = invalidCreateAccount();
        uint256 loadAmount = (params.loadAmountF) * 10 ** (18 - 8);

        vm.expectRevert(ISybil.InvalidCreateAccountTransaction.selector);
        vm.prank(address(this));
        sybil.addL1Transaction {
            value: loadAmount
        }(params.babyPubKey, params.fromIdx, params.loadAmountF, params.amountF, params.toIdx);
    }

    // Deposit transactions tests
    function testDepositTransaction() public {
        TxParams memory params = validDeposit();
        uint256 loadAmount = (params.loadAmountF) * 10 ** (18 - 8);
        uint48 initialLastIdx = 256;

        uint256[2] memory proofA = [uint(0),uint(0)];
        uint256[2][2] memory proofB = [[uint(0), uint(0)], [uint(0), uint(0)]];
        uint256[2] memory proofC = [uint(0), uint(0)];
        uint256 input = uint(1);

        vm.prank(address(this));
        sybil.forgeBatch(
            initialLastIdx, 
            0xabc, 
            0, 
            0, 
            0, 
            0, 
            false, 
            proofA,
            proofB,
            proofC,
            input
        );

        vm.prank(address(this));
        sybil.addL1Transaction {
            value: loadAmount
        }(params.babyPubKey, params.fromIdx, params.loadAmountF, params.amountF, params.toIdx);
    }

    function testInvalidDepositTransaction() public {
        TxParams memory params = invalidDeposit();
        uint256 loadAmount = (params.loadAmountF) * 10 ** (18 - 8);
        uint48 initialLastIdx = 256;

        uint256[2] memory proofA = [uint(0),uint(0)];
        uint256[2][2] memory proofB = [[uint(0), uint(0)], [uint(0), uint(0)]];
        uint256[2] memory proofC = [uint(0), uint(0)];
        uint256 input = uint(1);

        vm.prank(address(this));
        sybil.forgeBatch(
            initialLastIdx, 
            0xabc, 
            0, 
            0, 
            0, 
            0, 
            false, 
            proofA,
            proofB,
            proofC,
            input
        );

        vm.expectRevert(ISybil.InvalidDepositTransaction.selector);
        vm.prank(address(this));
        sybil.addL1Transaction {
            value: loadAmount
        }(params.babyPubKey, params.fromIdx, params.loadAmountF, params.amountF, params.toIdx);
    }

    // ForceExit transactions tests
    function testForceExitTransaction() public {
        TxParams memory params = validForceExit();
        uint256 loadAmount = (params.loadAmountF) * 10 ** (18 - 8);
        uint48 initialLastIdx = 256;

        uint256[2] memory proofA = [uint(0),uint(0)];
        uint256[2][2] memory proofB = [[uint(0), uint(0)], [uint(0), uint(0)]];
        uint256[2] memory proofC = [uint(0), uint(0)];
        uint256 input = uint(1);

        vm.prank(address(this));
        sybil.forgeBatch(
            initialLastIdx, 
            0xabc, 
            0, 
            0, 
            0, 
            0, 
            false, 
            proofA,
            proofB,
            proofC,
            input
        );

        vm.prank(address(this));
        sybil.addL1Transaction {
            value: loadAmount
        }(params.babyPubKey, params.fromIdx, params.loadAmountF, params.amountF, params.toIdx);
    }

    function testInvalidForceExitTransaction() public {
        TxParams memory params = invalidForceExit();
        uint256 loadAmount = (params.loadAmountF) * 10 ** (18 - 8);
        uint48 initialLastIdx = 256;

        uint256[2] memory proofA = [uint(0),uint(0)];
        uint256[2][2] memory proofB = [[uint(0), uint(0)], [uint(0), uint(0)]];
        uint256[2] memory proofC = [uint(0), uint(0)];
        uint256 input = uint(1);

        vm.prank(address(this));
        sybil.forgeBatch(
            initialLastIdx, 
            0xabc, 
            0, 
            0, 
            0, 
            0, 
            false, 
            proofA,
            proofB,
            proofC,
            input
        );

        vm.expectRevert(ISybil.InvalidForceExitTransaction.selector);
        vm.prank(address(this));
        sybil.addL1Transaction {
            value: loadAmount
        }(params.babyPubKey, params.fromIdx, params.loadAmountF, params.amountF, params.toIdx);
    }

    // ForceExplode transactions tests
    function testForceExplodeTransaction() public {
        TxParams memory params = validForceExplode();
        uint256 loadAmount = (params.loadAmountF) * 10 ** (18 - 8);
        uint48 initialLastIdx = 256;

        uint256[2] memory proofA = [uint(0),uint(0)];
        uint256[2][2] memory proofB = [[uint(0), uint(0)], [uint(0), uint(0)]];
        uint256[2] memory proofC = [uint(0), uint(0)];
        uint256 input = uint(1);

        vm.prank(address(this));
        sybil.forgeBatch(
            initialLastIdx, 
            0xabc, 
            0, 
            0, 
            0, 
            0, 
            false, 
            proofA,
            proofB,
            proofC,
            input
        );

        vm.prank(address(this));
        sybil.addL1Transaction {
            value: loadAmount
        }(params.babyPubKey, params.fromIdx, params.loadAmountF, params.amountF, params.toIdx);
    }

    // Invalid transaction parameters tests
    function testInvalidTransactionParameters() public {
        uint256 babyPubKey = 0;
        uint48 fromIdx = 5000;
        uint40 loadAmountF = 100;
        uint40 amountF = 0;
        uint48 toIdx = 100;

        uint256 loadAmount = (loadAmountF) * 10 ** (18 - 8);

        vm.expectRevert(ISybil.InvalidTransactionParameters.selector);
        vm.prank(address(this));
        sybil.addL1Transaction {
            value: loadAmount
        }(babyPubKey, fromIdx, loadAmountF, amountF, toIdx);
    }
 // Test initializing with invalid Poseidon addresses
    function testInitializeWithInvalidPoseidonAddresses() public {
        PoseidonUnit2 mockPoseidon2 = new PoseidonUnit2();
        PoseidonUnit3 mockPoseidon3 = new PoseidonUnit3();
        PoseidonUnit4 mockPoseidon4 = new PoseidonUnit4();
        // Deploy verifier stub
        VerifierRollupStub verifierStub = new VerifierRollupStub(); 
        
           address[] memory verifiers = new address[](1);
        uint256[] memory maxTx = new uint256[](1);
        uint256[] memory nLevels = new uint256[](1);

        verifiers[0] = address(verifierStub);
        maxTx[0] = uint(256);
        nLevels[0] = uint(1);


        address invalidAddress = address(0);

        // Expect revert for invalid poseidon2Elements address
        vm.expectRevert(bytes("Invalid poseidon2Elements address"));
        new Sybil(
            verifiers, 
            maxTx, 
            nLevels, 
            120, 
            invalidAddress, 
            address(mockPoseidon3), 
            address(mockPoseidon4)
        );

        // Expect revert for invalid poseidon3Elements address
        vm.expectRevert(bytes("Invalid poseidon3Elements address"));
        new Sybil(
            verifiers, 
            maxTx, 
            nLevels, 
            120, 
            address(mockPoseidon2), 
            invalidAddress, 
            address(mockPoseidon4)
        );

        // Expect revert for invalid poseidon4Elements address
        vm.expectRevert(bytes("Invalid poseidon4Elements address"));
        new Sybil(
            verifiers, 
            maxTx, 
            nLevels, 
            120, 
            address(mockPoseidon2), 
            address(mockPoseidon3), 
            invalidAddress
        );
    }
      // Test setForgeL1L2BatchTimeout called by non-owner
    function testSetForgeL1L2BatchTimeoutNonOwner() public {
        uint8 newTimeout = 100;

        // Set up a different address
        address nonOwner = address(0x123);

        vm.prank(nonOwner);
        vm.expectRevert();
        sybil.setForgeL1L2BatchTimeout(newTimeout);
    }
    // Test withdrawMerkleProof when exit nullifier is already set
    function testWithdrawMerkleProofAlreadyDone() public {
        uint192 amount = 1 ether;
        uint256 babyPubKey = 0x1234;
        uint32 numExitRoot = 1;
        uint48 idx = 0;

        bytes32 slot = keccak256(abi.encode(idx, keccak256(abi.encode(numExitRoot, uint256(keccak256("exitNullifierMap"))))));
        vm.store(address(sybil), slot, bytes32(uint256(1)));

    uint256 [] memory siblings; 

        vm.expectRevert(0x6d963f88);
        sybil.withdrawMerkleProof(
            amount,
            babyPubKey,
            numExitRoot,
            siblings,
            idx
        );
    }
    // Test withdrawMerkleProof with invalid SMT proof
    function testWithdrawMerkleProofInvalidSmtProof() public {
        uint192 amount = 1 ether;
        uint256 babyPubKey = 0x1234;
        uint32 numExitRoot = 1;
        uint48 idx = 0;

        // Directly set exitRootsMap[numExitRoot] to a dummy value
        bytes32 slot = keccak256(abi.encode(numExitRoot, uint256(keccak256("exitRootsMap"))));
        vm.store(address(sybil), slot, bytes32(uint256(0xdeadbeef)));

    uint256 [] memory siblings; // Empty siblings

        vm.expectRevert(0x6d963f88);
        sybil.withdrawMerkleProof(
            amount,
            babyPubKey,
            numExitRoot,
            siblings,
            idx
        );
    }

        // Test withdrawMerkleProof where transfer fails
function testWithdrawMerkleProofTransferFails() public {
    // Deploy RevertingReceiver contract
    RevertingReceiver receiver = new RevertingReceiver();

    uint192 amount = 1 ether;
    uint256 babyPubKey = 0x1234;
    uint32 numExitRoot = 1;
    uint48 idx = 0;

    // Directly set exitRootsMap[numExitRoot] to a dummy value
    bytes32 exitRootSlot = keccak256(abi.encode(numExitRoot, uint256(keccak256("exitRootsMap"))));
    vm.store(address(sybil), exitRootSlot, bytes32(uint256(0xdeadbeef)));

    // Ensure exitNullifierMap[numExitRoot][idx] is false
    bytes32 nullifierSlot = keccak256(abi.encode(idx, keccak256(abi.encode(numExitRoot, uint256(keccak256("exitNullifierMap"))))));
    vm.store(address(sybil), nullifierSlot, bytes32(uint256(0)));

    uint256 [] memory siblings; // Empty siblings

    // Expect revert due to ETH transfer failure
    vm.prank(address(receiver));
    vm.expectRevert(ISybil.EthTransferFailed.selector);
    sybil.withdrawMerkleProof(
        amount,
        babyPubKey,
        numExitRoot,
        siblings,
        idx
    );
}
}
// Helper contract that reverts on receiving ETH
contract RevertingReceiver {
    fallback() external payable {
        revert("Transfer failed");
    }
}