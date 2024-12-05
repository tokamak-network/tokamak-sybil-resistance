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
    
}
