// SPDX-License-Identifier: MIT
pragma solidity 0.8.23;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "../interfaces/IMVPSybil.sol";
import "../interfaces/IVerifierRollup.sol";
import "./SybilHelpers.sol";

contract Sybil is Initializable, OwnableUpgradeable, IMVPSybil, MVPSybilHelpers {
    uint48 constant _RESERVED_IDX = 255;
    uint48 constant _EXIT_IDX = 1;
    uint48 constant _EXPLODE_IDX = 2;
    uint256 constant _TXN_TOTALBYTES = 128; // Total bytes per transaction
    uint256 constant _MAX_TXNS = 1000; // Max transactions per batch

    uint8 public constant ABSOLUTE_MAX_L1BATCHTIMEOUT = 240;

    uint48 public lastIdx;
    uint32 public lastForgedBatch;
    uint32 public currentFillingBatch;

    mapping(uint32 => uint256) public accountRootMap;
    mapping(uint32 => uint256) public vouchRootMap; 
    mapping(uint32 => uint256) public scoreRootMap;
    mapping(uint32 => uint256) public exitRootMap;
    mapping(uint32 => bytes) public unprocessedBatchesMap;

    // Mapping of exit nullifiers, only allowing each withdrawal to be made once
    mapping(uint32 => mapping(uint48 => bool)) public exitNullifierMap;

    struct VerifierRollup {
        VerifierRollupInterface verifierInterface;
        uint256 maxTxs; // maximum rollup transactions in a batch: L1-tx transactions
        uint256 nLevels; // number of levels of the circuit
    }

    // Verifiers array
    VerifierRollup[] public rollupVerifiers;

    event L1UserTxEvent(
        uint32 indexed queueIndex,
        uint8 indexed position,
        bytes l1UserTx
    );

    event ForgeBatch(uint32 indexed batchNum, uint16 l1UserTxsLen);
    event WithdrawEvent(
        uint48 indexed idx,
        uint32 indexed numExitRoot
    );
    event Initialize(uint8 forgeL1BatchTimeout);

    function initialize(
        address[] memory verifiers,
        uint256[] memory maxTxs,
        uint256[] memory nLevels,
        uint8 _forgeL1BatchTimeout, 
        address _poseidon2Elements,
        address _poseidon3Elements,
        address _poseidon4Elements
    ) public initializer {
        lastIdx = _RESERVED_IDX;
        currentFillingBatch = 1;

        _initializeVerifiers(
            verifiers,
            maxTxs,
            nLevels
        );

        _initializeHelpers(
            _poseidon2Elements,
            _poseidon3Elements,
            _poseidon4Elements
        );

        emit Initialize(_forgeL1BatchTimeout);
    }

    function _addTx(
        address ethAddress,
        uint48 fromIdx,
        uint40 loadAmountF,
        uint40 amountF,   
        uint48 toIdx
    ) public override {
        bytes memory l1Tx = abi.encodePacked(
            ethAddress,
            fromIdx,
            loadAmountF,
            amountF,
            toIdx
        );

        uint256 currentPosition = unprocessedBatchesMap[currentFillingBatch].length /
            _TXN_TOTALBYTES;

        unprocessedBatchesMap[currentFillingBatch] = bytes.concat(
            unprocessedBatchesMap[currentFillingBatch],
            l1Tx
        );

        emit L1UserTxEvent(currentFillingBatch, uint8(currentPosition), l1Tx);

        if (currentPosition + 1 >= _MAX_TXNS) {
            currentFillingBatch++;
        }
    }

    function createAccount() external payable override{
        _addTx(msg.sender, 0, 0, 0, 0);
    }

    function deposit(uint48 fromIdx, uint40 loadAmountF, uint40 amountF) external payable override {
        _addTx(msg.sender, fromIdx, loadAmountF, amountF, 0);
    }

    function exit(uint48 fromIdx, uint40 loadAmountF, uint40 amountF) external payable override {
        _addTx(msg.sender, fromIdx, loadAmountF, amountF, _EXIT_IDX);
    }

    function explode(uint48 fromIdx, uint40 loadAmountF, uint40 amountF) external payable override {
        _addTx(msg.sender, fromIdx, loadAmountF, amountF, _EXPLODE_IDX);
    }

    // Implement the missing function from the IMvp interface
    function forgeBatch(
        uint48 newLastIdx,
        uint256 newAccountRoot,
        uint256 newVouchRoot,
        uint256 newScoreRoot,
        uint256 newExitRoot,
        uint8 verifierIdx,
        uint256[2] calldata proofA,
        uint256[2][2] calldata proofB,
        uint256[2] calldata proofC,
        uint256 input
    ) external override {
        // Verify the proof using the specific rollup verifier
        if (
            !rollupVerifiers[verifierIdx].verifierInterface.verifyProof(
                proofA,
                proofB,
                proofC,
                [input]
            )
        ) {
            revert InvalidProof();
        }

        lastForgedBatch++;
        lastIdx = newLastIdx;
        accountRootMap[lastForgedBatch] = newAccountRoot;
        vouchRootMap[lastForgedBatch] = newVouchRoot;
        scoreRootMap[lastForgedBatch] = newScoreRoot;
        exitRootMap[lastForgedBatch] = newExitRoot;

        uint16 l1UserTxsLen = _clearBatchFromQueue();

        emit ForgeBatch(lastForgedBatch, l1UserTxsLen);
    }

    function setForgeL1BatchTimeout(uint8 newTimeout) external pure override {
        // Timeout logic
        if (newTimeout > ABSOLUTE_MAX_L1BATCHTIMEOUT) {
            revert BatchTimeoutExceeded();
        }
    }

    function _clearBatchFromQueue() internal returns (uint16) {
        uint16 l1UserTxsLen = uint16(
            unprocessedBatchesMap[lastForgedBatch].length / _TXN_TOTALBYTES
        );
        delete unprocessedBatchesMap[lastForgedBatch];
        if (lastForgedBatch + 1 == currentFillingBatch) {
            currentFillingBatch++;
        }
        return l1UserTxsLen;
    }

    function withdrawMerkleProof(
        uint192 amount,
        uint32 numExitRoot,
        uint256[] calldata siblings,
        uint48 idx
    ) external {
        uint256[4] memory arrayState = _buildTreeState(amount, msg.sender);
        uint256 stateHash = _hash4Elements(arrayState);

        uint256 exitRoot = exitRootMap[numExitRoot];

        if (exitNullifierMap[numExitRoot][idx]) {
            revert WithdrawAlreadyDone();
        }

        if (!_smtVerifier(exitRoot, siblings, idx, stateHash)) {
            revert SmtProofInvalid();
        }

        exitNullifierMap[numExitRoot][idx] = true;

        _withdrawFunds(amount);
        emit WithdrawEvent(idx, numExitRoot);
    }

    function _buildTreeState(uint192 amount, address user) internal pure returns (uint256[4] memory) {
        uint256[4] memory state;
        state[0] = amount;
        state[1] = uint256(uint160(user)); // Convert address to uint256
        state[2] = 0;
        state[3] = 0;
        return state;
    }

    function _withdrawFunds(uint192 amount) internal {
        _safeTransfer(amount);
    }

    function _safeTransfer(uint256 value) internal {
        (bool success, ) = msg.sender.call{value: value}(new bytes(0));
        if (!success) {
            revert EthTransferFailed();
        }
    }

    function getStateRoot(uint32 batchNum) external view override returns (uint256) {
        return accountRootMap[batchNum];
    }

    function getLastForgedBatch() external view override returns (uint32) {
        return lastForgedBatch;
    }

    function getL1TransactionQueue(uint32 queueIndex) external view override returns (bytes memory) {
        return unprocessedBatchesMap[queueIndex];
    }

    function getQueueLength() external view override returns (uint32) {
        return currentFillingBatch - lastForgedBatch;
    }

    function _initializeVerifiers(
        address[] memory _verifiers,
        uint256[] memory _maxTxs,
        uint256[] memory _nLevels
    ) internal {
        uint256 len = _verifiers.length;
        for (uint256 i = 0; i < len; ++i) {
            if (_verifiers[i] == address(0)) {
                revert InvalidVerifierAddress();
            }

            rollupVerifiers.push(
                VerifierRollup({
                    verifierInterface: VerifierRollupInterface(_verifiers[i]),
                    maxTxs: _maxTxs[i],
                    nLevels: _nLevels[i]
                })
            );
        }
    }
}