// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.23;

import "../interfaces/IVerifierRollup.sol";

contract VerifierRollupStub is VerifierRollupInterface {
    function verifyProof(
        uint256[2] calldata a,
        uint256[2][2] calldata b,
        uint256[2] calldata c,
        uint256[1] calldata input
    ) public  override pure returns (bool) {
        return true;
    }
}
