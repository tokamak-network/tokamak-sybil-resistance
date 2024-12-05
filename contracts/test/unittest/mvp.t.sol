// SPDX-License-Identifier: MIT
pragma solidity 0.8 .23;

import "forge-std/Test.sol";
import "../../src/mvp/Sybil.sol";
import "../../src/interfaces/IMVPSybil.sol";
import "../_helpers/constants.sol";
import "../_helpers/MVPTransactionTypes.sol";
import "../../src/stub/VerifierRollupStub.sol";

contract MvpTest is Test, TransactionTypeHelper{
    Sybil public sybil;
    bytes32[] public hashes;

    function setUp() public {
        PoseidonUnit2 mockPoseidon2 = new PoseidonUnit2();
        PoseidonUnit3 mockPoseidon3 = new PoseidonUnit3();
        PoseidonUnit4 mockPoseidon4 = new PoseidonUnit4();
        emit log_address(address(mockPoseidon2));
        emit log_address(address(mockPoseidon3));
        emit log_address(address(mockPoseidon4));

        VerifierRollupStub verifierStub = new VerifierRollupStub(); 

        address[] memory verifiers = new address[](1);
        uint256[] memory maxTx = new uint256[](1);
        uint256[] memory nLevels = new uint256[](1);

        verifiers[0] = address(verifierStub);
        maxTx[0] = uint(256);
        nLevels[0] = uint(1);

        sybil = new Sybil();

        sybil.initialize(
            verifiers, 
            maxTx, 
            nLevels, 
            120, 
            address(mockPoseidon2), 
            address(mockPoseidon3), 
            address(mockPoseidon4)
        );
    }

    function testGetStateRoot() public {
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
            proofA,
            proofB,
            proofC,
            input
        );
        uint256 stateRoot = sybil.getStateRoot(batchNum);
        assertEq(stateRoot, 0xabc);
    }

    function testGetLastForgedBatch() public {
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
            proofA,
            proofB,
            proofC,
            input
        );

        lastForged = sybil.getLastForgedBatch();
        assertEq(lastForged, 1);
    }

    function testGetL1TransactionQueue() public {
        uint48 fromIdx = 0;
        uint40 loadAmountF = 100;
        uint40 amountF = 0;
        uint48 toIdx = 0;

        vm.prank(address(this));
        sybil.deposit(fromIdx, loadAmountF, amountF);

        bytes memory txData = sybil.getL1TransactionQueue(1);
        bytes memory expectedTxData = abi.encodePacked(address(this), fromIdx, loadAmountF, amountF, toIdx);
        assertEq(txData, expectedTxData);
    }

    function testGetQueueLength() public {
        uint32 queueLength = sybil.getQueueLength();
        assertEq(queueLength, 1);

        TxParams memory params = valid();
        uint256 loadAmount = (params.loadAmountF) * 10 ** (18 - 8);

        vm.prank(address(this));
        sybil.deposit {
            value: loadAmount
        }(params.fromIdx, params.loadAmountF, params.amountF);

        queueLength = sybil.getQueueLength();
        assertEq(queueLength, 1);
    }

    function testSetForgeL1BatchTimeout() public {
        uint8 newTimeout = 255;
        vm.expectRevert(IMVPSybil.BatchTimeoutExceeded.selector);
        sybil.setForgeL1BatchTimeout(newTimeout);
    }

    function testClearQueue() public {
        TxParams memory params = valid();
        uint256 loadAmount = (params.loadAmountF) * 10 ** (18 - 8);

        vm.prank(address(this));
        sybil.deposit {
            value: loadAmount
        }( params.fromIdx, params.loadAmountF, params.amountF);
        sybil.deposit {
            value: loadAmount
        }(params.fromIdx, params.loadAmountF, params.amountF);

        uint256[2] memory proofA = [uint(0),uint(0)];
        uint256[2][2] memory proofB = [[uint(0), uint(0)], [uint(0), uint(0)]];
        uint256[2] memory proofC = [uint(0), uint(0)];
        uint256 input = uint(1);

        uint32 queueAfter = sybil.getQueueLength();
        assertEq(queueAfter, 1);
        vm.prank(address(this));
        sybil.forgeBatch(
            256, 
            0xabc, 
            0, 
            0, 
            0, 
            0, 
            proofA,
            proofB,
            proofC,
            input
        );

        queueAfter = sybil.getQueueLength();
        assertEq(sybil.getLastForgedBatch(),1);
        assertEq(queueAfter, 0);
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
        emit Sybil.L1UserTxEvent(1, 0, abi.encodePacked(address(this), params.fromIdx, params.loadAmountF, params.amountF, params.toIdx));

        vm.prank(address(this));
        sybil.deposit {
            value: loadAmount
        }(params.fromIdx, params.loadAmountF, params.amountF);
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

        Sybil newSybil = new Sybil();
        newSybil.initialize(
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
        sybil.createAccount {
            value: loadAmount
        }();
    }

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
            proofA,
            proofB,
            proofC,
            input
        );

        vm.prank(address(this));
        sybil.deposit {
            value: loadAmount
        }(params.fromIdx, params.loadAmountF, params.amountF);
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
            proofA,
            proofB,
            proofC,
            input
        );

        vm.prank(address(this));
        sybil.exit {
            value: loadAmount
        }(params.fromIdx, params.loadAmountF, params.amountF);
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
            proofA,
            proofB,
            proofC,
            input
        );

        vm.prank(address(this));
        sybil.explode {
            value: loadAmount
        }(params.fromIdx, params.loadAmountF, params.amountF);
    }

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
        Sybil newSybil = new Sybil();
        vm.expectRevert();
        newSybil.initialize(
            verifiers, 
            maxTx, 
            nLevels, 
            120, 
            invalidAddress, 
            address(mockPoseidon3), 
            address(mockPoseidon4)
        );

        // Expect revert for invalid poseidon3Elements address
        vm.expectRevert();
        newSybil.initialize(
            verifiers, 
            maxTx, 
            nLevels, 
            120, 
            address(mockPoseidon2), 
            invalidAddress, 
            address(mockPoseidon4)
        );

        // Expect revert for invalid poseidon4Elements address
        vm.expectRevert();
        newSybil.initialize(
            verifiers, 
            maxTx, 
            nLevels, 
            120, 
            address(mockPoseidon2), 
            address(mockPoseidon3),
            invalidAddress
        );
    }

        // Test initializing with invalid verifier address
    function testInitializeWithInvalidVerifierAddresses() public {
        PoseidonUnit2 mockPoseidon2 = new PoseidonUnit2();
        PoseidonUnit3 mockPoseidon3 = new PoseidonUnit3();
        PoseidonUnit4 mockPoseidon4 = new PoseidonUnit4();
        // Deploy verifier stub
        VerifierRollupStub verifierStub = new VerifierRollupStub(); 
        
        address[] memory verifiers = new address[](1);
        uint256[] memory maxTx = new uint256[](1);
        uint256[] memory nLevels = new uint256[](1);

        verifiers[0] = address(0);
        maxTx[0] = uint(256);
        nLevels[0] = uint(1);


        address invalidAddress = address(0);

        // Expect revert for invalid verifier address
        Sybil newSybil = new Sybil();
        vm.expectRevert(IMVPSybil.InvalidVerifierAddress.selector);
        newSybil.initialize(
            verifiers, 
            maxTx, 
            nLevels, 
            120, 
            invalidAddress, 
            address(mockPoseidon3), 
            address(mockPoseidon4)
        );
    }

    function testWithdrawMerkleProofTransferFails() public {
        // Deploy RevertingReceiver contract

        uint192 amount = 1 ether;
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
        vm.expectRevert(IMVPSybil.EthTransferFailed.selector);
        sybil.withdrawMerkleProof(
            amount,
            numExitRoot,
            siblings,
            idx
        );
    }

    function testWithdrawMerkleProofTransferPasses() public {
        uint192 amount = 1 ether;
        uint256 babyPubKey = 0x1234;
        uint32 numExitRoot = 1;
        uint48 idx = 2;
        
        // Calcuate exit root
        bytes32 exitRoot = calculateTestExitTreeRoot();
        
        // forge batch with exit root
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
            uint(exitRoot), 
            0, 
            proofA,
            proofB,
            proofC,
            input
        );

        /* verify
            3rd leaf
            0xdca3326ad7e8121bf9cf9c12333e6b2271abe823ec9edfe42f813b1e768fa57b

            root
            0xcc086fcc038189b4641db2cc4f1de3bb132aefbd65d510d817591550937818c7

            index
            2

            proof
            0x8da9e1c820f9dbd1589fd6585872bc1063588625729e7ab0797cfc63a00bd950
            0x995788ffc103b987ad50f5e5707fd094419eb12d9552cc423bd0cd86a3861433
        */
        // Calculate proof (sibling)
        uint[] memory siblings = new uint[](2);
        siblings[0] = uint(0x8da9e1c820f9dbd1589fd6585872bc1063588625729e7ab0797cfc63a00bd950);
        siblings[1] = uint(0x995788ffc103b987ad50f5e5707fd094419eb12d9552cc423bd0cd86a3861433);

        bytes32 leaf = bytes32(0xdca3326ad7e8121bf9cf9c12333e6b2271abe823ec9edfe42f813b1e768fa57b);

        // verify proof
        bool isVerified = verify(
            siblings,
            exitRoot,
            leaf,
            idx
        );

        assert(isVerified == true);

        // call withdrawMerkleProof
        vm.expectRevert(IMVPSybil.EthTransferFailed.selector);
        sybil.withdrawMerkleProof(
            amount,
            numExitRoot,
            siblings,
            idx
        );   
    }

    function calculateTestExitTreeRoot() internal returns (bytes32) {
        uint256[4] memory transactions = [uint(0), uint(1), uint(2), uint(3)];
        uint256[4] memory keys = [uint(0), uint(1), uint(2), uint(3)];

        for (uint256 i = 0; i < transactions.length; i++) {
            uint256 hashValue = sybil._hashNode(keys[i], transactions[i]);
            hashes.push(bytes32(hashValue));
        }

        uint256 n = transactions.length;
        uint256 offset = 0;

        while (n > 0) {
            for (uint256 i = 0; i < n - 1; i += 2) {
                uint256 res = sybil._hashNode(uint(hashes[offset + i]), uint(hashes[offset + i + 1]));
                hashes.push(
                    bytes32(res)
                );
            }
            offset += n;
            n = n / 2;
        }

        return hashes[hashes.length - 1];
    }

    function verify(
        uint[] memory proof,
        bytes32 root,
        bytes32 leaf,
        uint256 index
    ) internal view returns (bool) {
        uint256 hash = uint(leaf);

        for (uint256 i = 0; i < proof.length; i++) {
            uint256 proofElement = uint(proof[i]);

            if (index % 2 == 0) {
                hash = sybil._hashNode(hash, proofElement);
            } else {
                hash = sybil._hashNode(proofElement, hash);
            }

            index = index / 2;
        }

        return hash == uint(root);
    }
    
}
