// SPDX-License-Identifier: MIT
pragma solidity 0.8.23;

contract TransactionTypeHelper {
    struct TxParams {
        string babyPubKey;
        uint48 fromIdx;
        uint40 loadAmountF;
        uint40 amountF;
        uint48 toIdx;
    }


    // Returns valid deposit transaction parameters
    function validDeposit() public pure returns (TxParams memory) {
        return TxParams({
            babyPubKey: "", 
            fromIdx: 256, 
            loadAmountF: 100, 
            amountF: 0, 
            toIdx: 0
        });
    }

    // Returns invalid deposit transaction parameters
    function invalidDeposit() public pure returns (TxParams memory) {
        return TxParams({
            babyPubKey: "12345", 
            fromIdx: 256, 
            loadAmountF: 100, 
            amountF: 100, 
            toIdx: 0
        });
    }

    // Returns valid CreateAccount transaction parameters
    function validCreateAccount() public pure returns (TxParams memory) {
        return TxParams({
            babyPubKey: "12345", 
            fromIdx: 0, 
            loadAmountF: 100, 
            amountF: 0, 
            toIdx: 0
        });
    }

    // Returns invalid CreateAccount transaction parameters
    function invalidCreateAccount() public pure returns (TxParams memory) {
        return TxParams({
            babyPubKey: "", 
            fromIdx: 0, 
            loadAmountF: 100, 
            amountF: 0, 
            toIdx: 0
        });
    }

    // Returns valid ForceExit transaction parameters
    function validForceExit() public pure returns (TxParams memory) {
        return TxParams({
            babyPubKey: "", 
            fromIdx: 256, 
            loadAmountF: 0, 
            amountF: 0, 
            toIdx: 1 
        });
    }

    // Returns invalid ForceExit transaction parameters
    function invalidForceExit() public pure returns (TxParams memory) {
        return TxParams({
            babyPubKey: "12345", 
            fromIdx: 256, 
            loadAmountF: 100, // Invalid non-zero loadAmountF
            amountF: 0, 
            toIdx: 1 
        });
    }

    // Returns valid ForceExplode transaction parameters
    function validForceExplode() public pure returns (TxParams memory) {
        return TxParams({
            babyPubKey: "", 
            fromIdx: 256, 
            loadAmountF: 0, 
            amountF: 0, 
            toIdx: 2 
        });
    }

    // Returns invalid ForceExplode transaction parameters
    function valid() public pure returns (TxParams memory) {
        return TxParams({
            babyPubKey: "2", 
            fromIdx: 0, 
            loadAmountF: 100, 
            amountF: 0, 
            toIdx: 0 
        });
    }

    // Returns invalid ForceExplode transaction parameters
    function invalidForceExplode() public pure returns (TxParams memory) {
        return TxParams({
            babyPubKey: "12345", 
            fromIdx: 256, 
            loadAmountF: 100, 
            amountF: 100, 
            toIdx: 2 
        });
    }
}
