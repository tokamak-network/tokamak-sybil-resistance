const {buildPoseidon} = require("circomlibjs");
const utils = require("ffjavascript").utils;

/**
 * Convert string to a big integer
 * @param {String} str - string
 * @return {Scalar} big integer
 */
function string2Int(str) {
    return utils.leBuff2int(Buffer.from(str));
}

module.exports.DB_Master = buildPoseidon([string2Int("Rollup_DB_Master")]);
module.exports.DB_Batch = buildPoseidon([string2Int("Rollup_DB_Batch")]);
module.exports.DB_Idx = buildPoseidon([string2Int("Rollup_DB_Idx")]);
module.exports.DB_AxAy = buildPoseidon([string2Int("Rollup_DB_AxAy")]);
module.exports.DB_EthAddr = buildPoseidon([string2Int("Rollup_DB_EthAddr")]);
module.exports.DB_TxPoolSlotsMap = buildPoseidon([string2Int("Rollup_DB_TxPoolSlots")]);
module.exports.DB_TxPollTx = buildPoseidon([string2Int("Rollup_DB_TxPollTx")]);
module.exports.DB_TxPoolDepositTx = buildPoseidon([string2Int("Rollup_DB_TxPoolDepositTx")]);
module.exports.DB_NumBatch_Idx = buildPoseidon([string2Int("Rollup_DB_NumBatch_Idx")]);
module.exports.DB_NumBatch_AxAy = buildPoseidon([string2Int("Rollup_DB_NumBatch_AxAy")]);
module.exports.DB_NumBatch_EthAddr = buildPoseidon([string2Int("Rollup_DB_NumBatch_EthAddr")]);
module.exports.DB_InitialIdx = buildPoseidon([string2Int("Rollup_DB_Initial_Idx")]);
module.exports.DB_ChainID = buildPoseidon([string2Int("Rollup_DB_ChainID")]);

module.exports.defaultChainID = 0;
module.exports.firstIdx = 255;
module.exports.exitIdx = 1;
module.exports.nullIdx = 0;
module.exports.maxNlevels = 48;
module.exports.exitAx = "0x0000000000000000000000000000000000000000000000000000000000000000";
module.exports.exitAy = "0x0000000000000000000000000000000000000000000000000000000000000000";
module.exports.nullEthAddr = "0xffffffffffffffffffffffffffffffffffffffff";
module.exports.createAccountMsg = "Account creation";
module.exports.EIP712Version = "1";
module.exports.EIP712Provider = "Sybil Network";