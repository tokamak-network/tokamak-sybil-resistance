// SPDX-License-Identifier: MIT

pragma solidity 0.8.23;


import "lib/openzeppelin-contracts-upgradeable/contracts/proxy/utils/Initializable.sol";
import "lib/openzeppelin-contracts-upgradeable/contracts/access/OwnableUpgradeable.sol";
import "lib/openzeppelin-contracts/contracts/token/ERC20/IERC20.sol";
import "lib/openzeppelin-contracts/contracts/token/ERC20/utils/SafeERC20.sol";




contract SybilVerifier is Initializable, OwnableUpgradeable {
    //TODO: change example values
    struct VerifierRollup {
        address verifierInterface;
        uint256 maxTx; // Maximum rollup transactions in a batch: L2-tx + L1-tx transactions
        uint256 nLevels; // Number of levels of the circuit
    }

    bytes4 constant _TRANSFER_SIGNATURE = 0xa9059cbb;
    bytes4 constant _TRANSFER_FROM_SIGNATURE = 0x23b872dd;
    bytes4 constant _APPROVE_SIGNATURE = 0x095ea7b3;
    bytes4 constant _PERMIT_SIGNATURE = 0xd505accf;

    uint48 constant _RESERVED_IDX = 255;
    uint48 constant _EXIT_IDX = 1;
    uint256 constant _LIMIT_LOAD_AMOUNT = (1 << 128);
    uint256 constant _LIMIT_L2TRANSFER_AMOUNT = (1 << 192);
    uint256 constant _LIMIT_TOKENS = (1 << 32);
    uint256 constant _INPUT_SHA_CONSTANT_BYTES = 20082;
uint256 constant _MAX_L1_USER_TX = 1000; 
uint256 constant _MAX_L1_TX = 1000; 
uint256 constant _L1_USER_TOTALBYTES = 128;
uint256 constant _L1_COORDINATOR_TOTALBYTES = 128;
    uint8 public constant ABSOLUTE_MAX_L1L2BATCHTIMEOUT = 240;
    uint256 constant _RFIELD = 21888242871839275222246405745257275088548364400416034343698204186575808495617;

    address constant _ETH_ADDRESS_INTERNAL_ONLY = address(0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF);

    VerifierRollup[] public rollupVerifiers;
    address public withdrawVerifier;

    uint48 public lastIdx;
    uint32 public lastForgedBatch;
    mapping(uint32 => uint256) public stateRootMap;
    mapping(uint32 => uint256) public exitRootsMap;
    mapping(uint32 => bytes32) public l1L2TxsDataHashMap;
    mapping(uint32 => mapping(uint48 => bool)) public exitNullifierMap;

    address[] public tokenList;
    mapping(address => uint256) public tokenMap;
    uint256 public feeAddToken;

    mapping(uint32 => bytes) public mapL1TxQueue;
    uint64 public lastL1L2Batch;
    uint32 public nextL1ToForgeQueue;
    uint32 public nextL1FillingQueue;
    uint8 public forgeL1L2BatchTimeout;
    address public tokenHEZ;
    

    event L1UserTxEvent(uint32 indexed queueIndex, uint8 indexed position, bytes l1UserTx);
    event AddToken(address indexed tokenAddress, uint32 tokenID);
    event ForgeBatch(uint32 indexed batchNum, uint16 l1UserTxsLen);
    event UpdateForgeL1L2BatchTimeout(uint8 newForgeL1L2BatchTimeout);
    event UpdateFeeAddToken(uint256 newFeeAddToken);
    event WithdrawEvent(uint48 indexed idx, uint32 indexed numExitRoot, bool indexed instantWithdraw);
    event InitializeSybilVerifierEvent(uint8 forgeL1L2BatchTimeout, uint256 feeAddToken, uint64 withdrawalDelay);


function _initializeVerifiers(address[] memory verifiers, uint256[] memory verifiersParams) internal {
    // Implementation for initializing verifiers
}

function _initializeHelpers(address poseidon2Elements, address poseidon3Elements, address poseidon4Elements) internal {
    // Implementation for initializing helpers
}

function _initializeWithdraw(address sybilGovernanceAddress, uint64 withdrawalDelay, address withdrawDelayerContract) internal {
    // Implementation for initializing withdraw parameters
}


//is withdraw delay needed? 
    function initializeSybilVerifier(
        address[] memory _verifiers,
        uint256[] memory _verifiersParams,
        address _withdrawVerifier,
        address _tokenHEZ,
        uint8 _forgeL1L2BatchTimeout,
        uint256 _feeAddToken,
        address _poseidon2Elements,
        address _poseidon3Elements,
        address _poseidon4Elements,
        address _sybilGovernanceAddress,
        uint64 _withdrawalDelay,
        address _withdrawDelayerContract
    ) external initializer {
        require(_withdrawDelayerContract != address(0), "SybilVerifier::initializeSybilVerifier ADDRESS_0_NOT_VALID");

        _initializeVerifiers(_verifiers, _verifiersParams);
        withdrawVerifier = _withdrawVerifier;
        tokenHEZ = _tokenHEZ;
        forgeL1L2BatchTimeout = _forgeL1L2BatchTimeout;
        feeAddToken = _feeAddToken;

        lastIdx = _RESERVED_IDX;
        nextL1FillingQueue = 1;
        tokenList.push(address(0)); 

        _initializeHelpers(_poseidon2Elements, _poseidon3Elements, _poseidon4Elements);
        _initializeWithdraw(_sybilGovernanceAddress, _withdrawalDelay, _withdrawDelayerContract);

        emit InitializeSybilVerifierEvent(_forgeL1L2BatchTimeout, _feeAddToken, _withdrawalDelay);
    }

    function _safeTransferFrom(address token, address from, address to, uint256 value) internal {
    SafeERC20.safeTransferFrom(IERC20(token), from, to, value);
}
//main function 
//TODO requires some adjustment to delete l1 batch queue
//TODO add extra checks 
    function forgeBatch(
        uint48 newLastIdx,
        uint256 newStRoot,
        uint256 newExitRoot,
        bytes calldata encodedL1CoordinatorTx,
        bytes calldata l1L2TxsData,
        bytes calldata feeIdxCoordinator,
        uint8 verifierIdx,
        bool l1Batch,
        uint256[2] calldata proofA,
        uint256[2][2] calldata proofB,
        uint256[2] calldata proofC
    ) external virtual {
        require(msg.sender == tx.origin, "SybilVerifier::forgeBatch: INTERNAL_TX_NOT_ALLOWED");

        if (!l1Batch) {
            require(block.number < (lastL1L2Batch + forgeL1L2BatchTimeout), "SybilVerifier::forgeBatch: L1L2BATCH_REQUIRED");
        }

        uint256 input = _constructCircuitInput(newLastIdx, newStRoot, newExitRoot, l1Batch, verifierIdx);

        require(
            _verifyProof(rollupVerifiers[verifierIdx].verifierInterface, proofA, proofB, proofC, input),
            "SybilVerifier::forgeBatch: INVALID_PROOF"
        );

        lastForgedBatch++;
        lastIdx = newLastIdx;
        stateRootMap[lastForgedBatch] = newStRoot;
        exitRootsMap[lastForgedBatch] = newExitRoot;
        l1L2TxsDataHashMap[lastForgedBatch] = sha256(l1L2TxsData);

        uint16 l1UserTxsLen;
        if (l1Batch) {
            lastL1L2Batch = uint64(block.number);
            l1UserTxsLen = _clearQueue();
        }

        emit ForgeBatch(lastForgedBatch, l1UserTxsLen);
    }

    function addL1Transaction(
        uint256 babyPubKey,
        uint48 fromIdx,
        uint40 loadAmountF,
        uint40 amountF,
        uint32 tokenID,
        uint48 toIdx,
        bytes calldata permit
    ) external payable {
        require(tokenID < tokenList.length, "SybilVerifier::addL1Transaction: TOKEN_NOT_REGISTERED");

        uint256 loadAmount = _float2Fix(loadAmountF);
        require(loadAmount < _LIMIT_LOAD_AMOUNT, "SybilVerifier::addL1Transaction: LOADAMOUNT_EXCEED_LIMIT");

        if (loadAmount > 0) {
            if (tokenID == 0) {
                require(loadAmount == msg.value, "SybilVerifier::addL1Transaction: LOADAMOUNT_ETH_DOES_NOT_MATCH");
            } else {
                require(msg.value == 0, "SybilVerifier::addL1Transaction: MSG_VALUE_NOT_EQUAL_0");
                if (permit.length != 0) {
                    _permit(tokenList[tokenID], loadAmount, permit);
                }
                uint256 prevBalance = IERC20(tokenList[tokenID]).balanceOf(address(this));
                _safeTransferFrom(tokenList[tokenID], msg.sender, address(this), loadAmount);
                uint256 postBalance = IERC20(tokenList[tokenID]).balanceOf(address(this));
                require(postBalance - prevBalance == loadAmount, "SybilVerifier::addL1Transaction: LOADAMOUNT_ERC20_DOES_NOT_MATCH");
            }
        }

        _addL1Transaction(msg.sender, babyPubKey, fromIdx, loadAmountF, amountF, tokenID, toIdx);
    }
//to be used within addL1Transaction and reduce function size
    function _addL1Transaction(
        address ethAddress,
        uint256 babyPubKey,
        uint48 fromIdx,
        uint40 loadAmountF,
        uint40 amountF,
        uint32 tokenID,
        uint48 toIdx
    ) internal {
        uint256 amount = _float2Fix(amountF);
        require(amount < _LIMIT_L2TRANSFER_AMOUNT, "SybilVerifier::_addL1Transaction: AMOUNT_EXCEED_LIMIT");

        if (toIdx == 0) {
            require(amount == 0, "SybilVerifier::_addL1Transaction: AMOUNT_MUST_BE_0_IF_NOT_TRANSFER");
        } else if (toIdx == _EXIT_IDX) {
            require(loadAmountF == 0, "SybilVerifier::_addL1Transaction: LOADAMOUNT_MUST_BE_0_IF_EXIT");
        } else {
            require(toIdx > _RESERVED_IDX && toIdx <= lastIdx, "SybilVerifier::_addL1Transaction: INVALID_TOIDX");
        }

        if (fromIdx == 0) {
            require(babyPubKey != 0, "SybilVerifier::_addL1Transaction: INVALID_CREATE_ACCOUNT_WITH_NO_BABYJUB");
        } else {
            require(fromIdx > _RESERVED_IDX && fromIdx <= lastIdx, "SybilVerifier::_addL1Transaction: INVALID_FROMIDX");
            require(babyPubKey == 0, "SybilVerifier::_addL1Transaction: BABYJUB_MUST_BE_0_IF_NOT_CREATE_ACCOUNT");
        }

        _l1QueueAddTx(ethAddress, babyPubKey, fromIdx, loadAmountF, amountF, tokenID, toIdx);
    }

//this function is used to add all L1s inside a queue which will then be deleted after the batch is forged 
//TODO: verify time of deletion. Should the queue be emptied right after the batch was forged? What check do we need to ensure the batch was forged succesfully and not needed to revert back

    function _l1QueueAddTx(
        address ethAddress,
        uint256 babyPubKey,
        uint48 fromIdx,
        uint40 loadAmountF,
        uint40 amountF,
        uint32 tokenID,
        uint48 toIdx
    ) internal {
        bytes memory l1Tx = abi.encodePacked(ethAddress, babyPubKey, fromIdx, loadAmountF, amountF, tokenID, toIdx);

        uint256 currentPosition = mapL1TxQueue[nextL1FillingQueue].length / _L1_USER_TOTALBYTES;

        _concatStorage(mapL1TxQueue[nextL1FillingQueue], l1Tx);

        emit L1UserTxEvent(nextL1FillingQueue, uint8(currentPosition), l1Tx);
        if (currentPosition + 1 >= _MAX_L1_USER_TX) {
            nextL1FillingQueue++;
        }
    }
//this function is just called after a batch is formed to delete all previous L1 data that was in queue
    function _clearQueue() internal returns (uint16) {
        uint16 l1UserTxsLen = uint16(mapL1TxQueue[nextL1ToForgeQueue].length / _L1_USER_TOTALBYTES);
        delete mapL1TxQueue[nextL1ToForgeQueue];
        nextL1ToForgeQueue++;
        if (nextL1ToForgeQueue == nextL1FillingQueue) {
            nextL1FillingQueue++;
        }
        return l1UserTxsLen;
    }
//TODO check functionality with circuit
    function _constructCircuitInput(
        uint48 newLastIdx,
        uint256 newStRoot,
        uint256 newExitRoot,
        bool l1Batch,
        uint8 verifierIdx
    ) internal view returns (uint256) {
        uint256 oldStRoot = stateRootMap[lastForgedBatch];
        uint256 oldLastIdx = lastIdx;

        uint256 l1L2TxsDataLength = ((rollupVerifiers[verifierIdx].nLevels / 8) * 2 + 5 + 1) * rollupVerifiers[verifierIdx].maxTx;
        uint256 feeIdxCoordinatorLength = (rollupVerifiers[verifierIdx].nLevels / 8) * 64;

        bytes memory inputBytes;
        uint256 ptr;

        assembly {
            let inputBytesLength := add(add(_INPUT_SHA_CONSTANT_BYTES, l1L2TxsDataLength), feeIdxCoordinatorLength)

            inputBytes := mload(0x40)
            mstore(0x40, add(add(inputBytes, 0x40), inputBytesLength))

            mstore(inputBytes, inputBytesLength)
            ptr := add(inputBytes, 32)

            mstore(ptr, shl(208, oldLastIdx))
            ptr := add(ptr, 6)

            mstore(ptr, shl(208, newLastIdx))
            ptr := add(ptr, 6)

            mstore(ptr, oldStRoot)
            ptr := add(ptr, 32)

            mstore(ptr, newStRoot)
            ptr := add(ptr, 32)

            mstore(ptr, newExitRoot)
            ptr := add(ptr, 32)
        }

        _buildL1Data(ptr, l1Batch);
        ptr += _MAX_L1_TX * _L1_USER_TOTALBYTES;

        (uint256 dPtr, uint256 dLen) = _getCallData(4);
        require(dLen <= l1L2TxsDataLength, "SybilVerifier::_constructCircuitInput: L2_TX_OVERFLOW");

        assembly {
            calldatacopy(ptr, dPtr, dLen)
        }
        ptr += dLen;
        _fillZeros(ptr, l1L2TxsDataLength - dLen);
        ptr += l1L2TxsDataLength - dLen;

        (dPtr, dLen) = _getCallData(5);
        require(dLen <= feeIdxCoordinatorLength, "SybilVerifier::_constructCircuitInput: INVALID_FEEIDXCOORDINATOR_LENGTH");

        assembly {
            calldatacopy(ptr, dPtr, dLen)
        }
        ptr += dLen;
        _fillZeros(ptr, feeIdxCoordinatorLength - dLen);
        ptr += feeIdxCoordinatorLength - dLen;

        assembly {
            mstore(ptr, shl(240, chainid()))
        }
        ptr += 2;

        uint256 batchNum = lastForgedBatch + 1;

        assembly {
            mstore(ptr, shl(224, batchNum))
        }

        return uint256(sha256(inputBytes)) % _RFIELD;
    }
//TODO maybe we can have this within another function
    function _buildL1Data(uint256 ptr, bool l1Batch) internal view {
        uint256 dPtr;
        uint256 dLen;

        (dPtr, dLen) = _getCallData(3);
        uint256 l1CoordinatorLength = dLen / _L1_COORDINATOR_TOTALBYTES;

        uint256 l1UserLength;
        bytes memory l1UserTxQueue;
        if (l1Batch) {
            l1UserTxQueue = mapL1TxQueue[nextL1ToForgeQueue];
            l1UserLength = l1UserTxQueue.length / _L1_USER_TOTALBYTES;
        } else {
            l1UserLength = 0;
        }

        require(l1UserLength + l1CoordinatorLength <= _MAX_L1_TX, "SybilVerifier::_buildL1Data: L1_TX_OVERFLOW");

        if (l1UserLength > 0) {
            assembly {
                let ptrFrom := add(l1UserTxQueue, 0x20)
                let ptrTo := ptr
                ptr := add(ptr, mul(l1UserLength, _L1_USER_TOTALBYTES))
                for {} lt(ptrTo, ptr) { ptrTo := add(ptrTo, 32) ptrFrom := add(ptrFrom, 32) } {
                    mstore(ptrTo, mload(ptrFrom))
                }
            }
        }
//TODO: avoid use of i++ and verify functionality of assembly s
        for (uint256 i = 0; i < l1CoordinatorLength; i++) {
            uint8 v;
            bytes32 s;
            bytes32 r;
            bytes32 babyPubKey;
            uint256 tokenID;

            assembly {
                v := byte(0, calldataload(dPtr))
                dPtr := add(dPtr, 1)

                s := calldataload(dPtr)
                dPtr := add(dPtr, 32)

                r := calldataload(dPtr)
                dPtr := add(dPtr, 32)

                babyPubKey := calldataload(dPtr)
                dPtr := add(dPtr, 32)

                tokenID := shr(224, calldataload(dPtr))
                dPtr := add(dPtr, 4)
            }

            require(tokenID < tokenList.length, "SybilVerifier::_buildL1Data: TOKEN_NOT_REGISTERED");

            address ethAddress = _ETH_ADDRESS_INTERNAL_ONLY;

            if (v != 0) {
                ethAddress = _checkSig(babyPubKey, r, s, v);
            }

            assembly {
                mstore(ptr, shl(96, ethAddress))
                ptr := add(ptr, 20)

                mstore(ptr, babyPubKey)
                ptr := add(ptr, 32)

                mstore(ptr, 0)
                ptr := add(ptr, 16)

                mstore(ptr, shl(224, tokenID))
                ptr := add(ptr, 4)

                mstore(ptr, 0)
                ptr := add(ptr, 6)
            }
        }

        _fillZeros(ptr, (_MAX_L1_TX - l1UserLength - l1CoordinatorLength) * _L1_USER_TOTALBYTES);
    }

    function _permit(
        address token,
        uint256 _amount,
        bytes calldata _permitData
    ) internal {
        bytes4 sig = abi.decode(_permitData, (bytes4));
        require(sig == _PERMIT_SIGNATURE, "SybilVerifier::_permit: NOT_VALID_CALL");

        (address owner, address spender, uint256 value, uint256 deadline, uint8 v, bytes32 r, bytes32 s) =
            abi.decode(_permitData[4:], (address, address, uint256, uint256, uint8, bytes32, bytes32));
        require(owner == msg.sender, "SybilVerifier::_permit: PERMIT_OWNER_MUST_BE_THE_SENDER");
        require(spender == address(this), "SybilVerifier::_permit: SPENDER_MUST_BE_THIS");
        require(value == _amount, "SybilVerifier::_permit: PERMIT_AMOUNT_DOES_NOT_MATCH");

        address(token).call(abi.encodeWithSelector(_PERMIT_SIGNATURE, owner, spender, value, deadline, v, r, s));
    }

    function _verifyProof(
        address verifier,
        uint256[2] calldata proofA,
        uint256[2][2] calldata proofB,
        uint256[2] calldata proofC,
        uint256 input
    ) internal view returns (bool) {
        // TODO this function would call the actual proof verification on the verifier contract
        // has a method `verifyProof` that returns a boolean indicating proof validity.
               (bool success, bytes memory result) = verifier.staticcall(
            abi.encodeWithSelector(bytes4(keccak256("verifyProof(uint256[2],uint256[2][2],uint256[2],uint256)")), proofA, proofB, proofC, input)
        );
        //TODO add a check here that it returned true
        return true;
    }

//TODO remove use of ecrecover here due to security issue.
    function _checkSig(
        bytes32 babyPubKey,
        bytes32 r,
        bytes32 s,
        uint8 v
    ) internal pure returns (address) {
        // Verify the signature using ecrecover
        bytes32 hash = keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n32", babyPubKey));
        return ecrecover(hash, v, r, s);
    }

    function _concatStorage(bytes storage _preBytes, bytes memory _postBytes) internal {
        assembly {
            let fslot := sload(_preBytes.slot)
            let slength := div(and(fslot, sub(mul(0x100, iszero(and(fslot, 1))), 1)), 2)
            let mlength := mload(_postBytes)
            let newlength := add(slength, mlength)
            sstore(_preBytes.slot, or(mul(newlength, 2), 1))

            let slengthmod := mod(slength, 32)
            let mlengthmod := mod(mlength, 32)
            let mc := add(_postBytes, 0x20)
            let end := add(mc, mlength)
            let mask := sub(exp(0x100, sub(32, slengthmod)), 1)
            let fslot2 := add(fslot, div(mload(mc), exp(0x100, sub(32, slengthmod))))
            sstore(add(_preBytes.slot, div(slength, 32)), and(fslot2, mask))
            mc := add(mc, sub(32, slengthmod))
            for { } lt(mc, end) { mc := add(mc, 32) } {
                sstore(add(_preBytes.slot, div(add(slength, sub(mc, add(_postBytes, 0x20))), 32)), mload(mc))
            }
            if mlengthmod {
                sstore(add(_preBytes.slot, div(newlength, 32)), mul(div(mload(mc), exp(0x100, mlengthmod)), exp(0x100, sub(32, mlengthmod))))
            }
        }
    }

    function _float2Fix(uint40 floatVal) internal pure returns (uint256) {
        return uint256(floatVal) * 10**(18 - 8); 
    }

    function _getCallData(uint256 position) internal pure returns (uint256 dPtr, uint256 dLen) {
        assembly {
            dPtr := calldataload(add(0x24, mul(position, 0x20)))
            dLen := calldataload(add(0x44, mul(position, 0x20)))
        }
    }

    function _fillZeros(uint256 ptr, uint256 len) internal pure {
        assembly {
            for { let end := add(ptr, len) } lt(ptr, end) { ptr := add(ptr, 32) } {
                mstore(ptr, 0)
            }
        }
    }

    // Uniqueness Score Calculation and Other Related Functions

    struct AccountData {
        uint256 uniquenessScore;
        uint48 accountIndex;
    }

    mapping(address => AccountData) public accountData;
    mapping(uint48 => address) public accountIndexToAddress;

    event AccountRegistered(address indexed account, uint48 accountIndex);
    event UniquenessScoreUpdated(address indexed account, uint256 newScore);

    function registerAccount(address account) external {
        require(accountData[account].accountIndex == 0, "SybilVerifier::registerAccount: ALREADY_REGISTERED");

        lastIdx++;
        accountData[account] = AccountData({uniquenessScore: 0, accountIndex: lastIdx});
        accountIndexToAddress[lastIdx] = account;

        emit AccountRegistered(account, lastIdx);
    }

    function updateUniquenessScore(address account, uint256 newScore) external {
        require(accountData[account].accountIndex != 0, "SybilVerifier::updateUniquenessScore: ACCOUNT_NOT_REGISTERED");

        accountData[account].uniquenessScore = newScore;

        emit UniquenessScoreUpdated(account, newScore);
    }

    function getUniquenessScore(address account) external view returns (uint256) {
        return accountData[account].uniquenessScore;
    }

    function getAccountIndex(address account) external view returns (uint48) {
        return accountData[account].accountIndex;
    }

    function getAddressByIndex(uint48 index) external view returns (address) {
        return accountIndexToAddress[index];
    }

//TODO verify if we require governance for this or not 
    function updateForgeL1L2BatchTimeout(uint8 newForgeL1L2BatchTimeout) external onlyGovernance {
        require(newForgeL1L2BatchTimeout <= ABSOLUTE_MAX_L1L2BATCHTIMEOUT, "SybilVerifier::updateForgeL1L2BatchTimeout: MAX_TIMEOUT_EXCEEDED");
        forgeL1L2BatchTimeout = newForgeL1L2BatchTimeout;

        emit UpdateForgeL1L2BatchTimeout(newForgeL1L2BatchTimeout);
    }

    function updateFeeAddToken(uint256 newFeeAddToken) external onlyGovernance {
        feeAddToken = newFeeAddToken;

        emit UpdateFeeAddToken(newFeeAddToken);
    }

  
    modifier onlyGovernance() {
        require(msg.sender == address(this), "SybilVerifier::onlyGovernance: CALLER_NOT_GOVERNANCE");
        _;
    }
}
