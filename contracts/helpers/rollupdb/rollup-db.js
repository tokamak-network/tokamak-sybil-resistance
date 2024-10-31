const SMT = require("circomlib").SMT;
const poseidonHash = require("circomlib").poseidon;
const Scalar = require("ffjavascript").Scalar;
const BabyJubJub = require("circomlib").babyJub;

const SMTTmpDb = require("./smt-tmp-db");
const BatchBuilder = require("./batch-builder");
const Constants = require("../utils/constants");
const stateUtils = require("../utils/state-utils");

class RollupDB {

    constructor(db, lastBatch, stateRoot, initialIdx, chainID) {
        this.db = db;
        this.lastBatch = lastBatch || 0;
        this.stateRoot = stateRoot || 0;
        this.initialIdx = initialIdx || 255;
        this.chainID = chainID;
    }

    /**
     * Return a new Batchbuilder with the current RollupDb state
     * @param {Scalar} maxNTx - Maximum number of transactions
     * @param {Scalar} nLevels - Number of levels in the merkle tree
     * @param {Scalar} maxL1Tx - Maximum number of L1 transactions
     * @param {Scalar} nFeeTx - Number of fee Tx (maximum number of txs)
     */
    async buildBatch(maxNTx, nLevels, maxL1Tx, nFeeTx) {
        return new BatchBuilder(this, this.lastBatch+1, this.stateRoot,
            this.initialIdx, maxNTx, nLevels, maxL1Tx, this.chainID, nFeeTx || 64);
    }

    /**
     * Consolidate a batch by writing it in the DB
     * @param {Object} bb - Batchbuilder object
     */
    async consolidate(bb) {
        if (bb.batchNumber != this.lastBatch +1) {
            throw new Error("Updating the wrong batch");
        }
        if (!bb.builded) {
            await bb.build();
        }
        const insertsState = Object.keys(bb.dbState.inserts).reverse().map(function(key) {
            return [Scalar.e(key), bb.dbState.inserts[key]];
        });
        const insertsExit = Object.keys(bb.dbExit.inserts).map(function(key) {
            return [Scalar.e(key), bb.dbExit.inserts[key]];
        });
        await this.db.multiIns([
            ...insertsState,
            ...insertsExit,
            [ Scalar.add(Constants.DB_Batch, bb.batchNumber), [bb.stateTree.root, bb.exitTree.root]],
            [ Scalar.add(Constants.DB_InitialIdx, bb.batchNumber), bb.finalIdx],
            [ Constants.DB_Master, bb.batchNumber]
        ]);
        this.lastBatch = bb.batchNumber;
        this.stateRoot = bb.stateTree.root;
        this.initialIdx = bb.finalIdx;
    }

    /**
     * Revert the RollupDb State and the DB state to a specific batch number
     * @param {Scalar} numBatch - Batch number
     */
    async rollbackToBatch(numBatch){
        if (numBatch > this.lastBatch)
            throw new Error("Cannot rollback to future state");

        // Update Idx database
        await this._updateIdx(numBatch);
        // Update AxAy database
        await this._updateAxAy(numBatch);
        // Update ethAddr databse
        await this._updateEthAddr(numBatch);

        // Update num batch and root
        await this.db.multiIns([
            [Constants.DB_Master, numBatch]
        ]);
        const roots = await this.db.get(Scalar.add(Constants.DB_Batch, numBatch));
        this.lastBatch = numBatch;
        if (numBatch === 0)
            this.stateRoot = Scalar.e(0);
        else
            this.stateRoot = roots[0];
    }

    /**
     * Get the state of the leaf from the Idx
     * @param {Scalar} idx - Leaf identifier
     * @returns {Object} State of the leaf
     */
    async getStateByIdx(idx) {
        const key = Scalar.add(Constants.DB_Idx, idx);
        const valStates = await this.db.get(key);
        if (!valStates) return null;
        // Get last state
        const lastState = valStates.slice(-1)[0];
        if (!lastState) return null;
        // Last state key
        const keyLastState = poseidonHash([idx, lastState]);
        const keyValueState = await this.db.get(keyLastState);
        if (!keyValueState) return null;
        const stateArray = await this.db.get(keyValueState);
        if (!stateArray) return null;
        const st = stateUtils.array2State(stateArray);
        st.idx = Number(idx);
        // st.rollupAddress = this.pointToCompress(st.ax, st.ay);
        return st;
    }

    /**
     * Return all the states that matches some babyjub public key
     * @param {Number} sign - babyjubjub sign
     * @param {String} ay - babyjubjub ay coordinate encoded as hexadecimal string
     * @returns {Array} States of the leafs
     */
    async getStateBySignAy(sign, ay) {
        const idxs = await this.getIdxsBySignAy(sign, ay);
        if (!idxs) return null;
        const promises = [];
        for (let i=0; i<idxs.length; i++) {
            promises.push(this.getStateByIdx(idxs[i]));
        }
        return Promise.all(promises);
    }

    /**
     * Return all the states that matches some ethereum address
     * @param {String} ethAddr - Ethereum address
     * @returns {Array} States of the leafs
     */
    async getStateByEthAddr(ethAddr) {
        const idxs = await this.getIdxsByEthAddr(ethAddr);

        if (!idxs) return null;
        const promises = [];
        for (let i=0; i<idxs.length; i++) {
            promises.push(this.getStateByIdx(idxs[i]));
        }
        return Promise.all(promises);
    }

    /**
     * Return all the idxs that matches babyjubjub public key
     * @param {Number} sign - sign Bjj
     * @param {String} ay - babyjubjub ay coordinate encoded as hexadecimal string
     * @returns {Array} indexes
     */
    async getIdxsBySignAy(sign, ay){
        let keyAxAy = Scalar.add(Constants.DB_AxAy, sign);
        keyAxAy = Scalar.add(keyAxAy, Scalar.fromString(ay, 16));

        const valStates = await this.db.get(keyAxAy);
        if (!valStates) return null;
        // Get last state
        const lastState = valStates.slice(-1)[0];
        if (!lastState) return null;
        // Last state key
        const keyLastState = poseidonHash([keyAxAy, lastState]);
        const idxs = await this.db.get(keyLastState);
        return idxs;
    }

    /**
     * Return all the states that matches some ethereum address
     * @param {String} ethAddr - Ethereum address
     * @returns {Array} indexes
     */
    async getIdxsByEthAddr(ethAddr){
        const keyEth = Scalar.add(Constants.DB_EthAddr, Scalar.fromString(ethAddr, 16));
        const valStates = await this.db.get(keyEth);
        if (!valStates) return null;
        // Get last state
        const lastState = valStates.slice(-1)[0];
        if (!lastState) return null;
        // Last state key
        const keyLastState = poseidonHash([keyEth, lastState]);
        const idxs = await this.db.get(keyLastState);
        return idxs;
    }

    /**
     * Get exit tree information for some account in an specific batch
     * @param {Number} idx - merkle tree index
     * @param {Scalar} numBatch - Batch number
     * @returns {Object} Exit tree information
     */
    async getExitTreeInfo(idx, numBatch){
        const rootExitTree = await this.getExitRoot(numBatch);
        if (!rootExitTree) return null;
        const dbExit = new SMTTmpDb(this.db);
        const tmpExitTree = new SMT(dbExit, rootExitTree);
        const resFindExit = await tmpExitTree.find(Scalar.e(idx));
        // Get leaf information
        if (resFindExit.found) {
            const foundValueId = poseidonHash([resFindExit.foundValue, idx]);
            const stateArray = await this.db.get(foundValueId);
            const state = stateUtils.array2State(stateArray);
            state.idx = Number(idx);
            resFindExit.state = state;
            delete resFindExit.foundValue;
        }
        delete resFindExit.isOld0;
        return resFindExit;
    }

    /**
     * Given an array of states, return the last state, before or equal to the lastBatch
     * @param {Array} valueStates - Array of states
     * @returns {Object} State object
     */
    _findLastState(valueStates){
        const lastBatch = Scalar.e(this.lastBatch);
        for (let i = valueStates.length - 1; i >= 0; i--){
            if (Scalar.leq(valueStates[i], lastBatch))
                return valueStates[i];
        }
        return null;
    }

    /**
     * Return the last batch saved in rollupDb
     * @returns {Scalar} Last batch
     */
    getLastBatchId(){
        return this.lastBatch;
    }

    /**
     * Return the last root saved in rollupDb
     * @returns {Scalar} Last root
     */
    getRoot(){
        return this.stateRoot;
    }

    /**
     * Return the exit root saved in rollupDb depending on batch number
     * @param {Scalar} numBatch - Batch number
     * @returns {Scalar} exit root
     */
    async getExitRoot(numBatch){
        if (numBatch > this.lastBatch)
            return null;

        const keyRoot = Scalar.add(Constants.DB_Batch, Scalar.e(numBatch));
        const rootValues = await this.db.get(keyRoot);
        if (!rootValues) return null;
        const rootExitTree = rootValues[1];
        return rootExitTree;
    }

    /**
     * Return the state root saved in rollupDb depending on batch number
     * @param {Scalar} numBatch - Batch number
     * @returns {Scalar} state root
     */
    async getStateRoot(numBatch){
        if (numBatch > this.lastBatch)
            return null;

        const keyRoot = Scalar.add(Constants.DB_Batch, Scalar.e(numBatch));
        const rootValues = await this.db.get(keyRoot);
        if (!rootValues) return null;
        const rootExitTree = rootValues[0];
        return rootExitTree;
    }

    /**
     * Updates the array of idx states in the DDBB deleting all the information above the numBatch
     * @param {Scalar} numBatch - Batch number
     */
    async _updateIdx(numBatch) {
        // Update idx states
        const alreadyUpdated = [];
        for (let i = this.lastBatch; i > numBatch; i--){
            const keyNumBatchIdx = Scalar.add(Constants.DB_NumBatch_Idx, i);
            const idxToUpdate = await this.db.get(keyNumBatchIdx);
            if (!idxToUpdate) continue;
            for (const idx of idxToUpdate) {
                if (!alreadyUpdated.includes(idx)){
                    const keyIdx = Scalar.add(Constants.DB_Idx, idx);
                    const states = await this.db.get(keyIdx);
                    this._purgeStates(states, numBatch);
                    await this.db.multiIns([
                        [keyIdx, states],
                    ]);
                    alreadyUpdated.push(idx);
                }
            }
        }

        // Reset numBatch-idx for future states
        const keysToDel = [];
        for (let i = this.lastBatch; i > numBatch; i--){
            const keyNumBatchIdx = Scalar.add(Constants.DB_NumBatch_Idx, i);
            keysToDel.push(keyNumBatchIdx);
        }
        await this.db.multiDel(keysToDel);
    }

    /**
     * Updates the array of babyjubjub states in the DB deleting all the information above the numBatch
     * @param {Scalar} numBatch - Batch number
     */
    async _updateAxAy(numBatch) {
        // Update axAy states
        const alreadyUpdated = [];
        for (let i = this.lastBatch; i > numBatch; i--){
            const keyNumBatchAxAy = Scalar.add(Constants.DB_NumBatch_AxAy, i);
            const axAyToUpdate = await this.db.get(keyNumBatchAxAy);
            if (!axAyToUpdate) continue;
            for (const hashAxAy of axAyToUpdate) {
                if (!alreadyUpdated.includes(hashAxAy)){
                    const valueHashAxAy = await this.db.get(hashAxAy);
                    const ax = valueHashAxAy[0];
                    const ay = valueHashAxAy[1];
                    const keyAxAy = Scalar.add(Scalar.add(Constants.DB_AxAy, ax), ay);
                    const states = await this.db.get(keyAxAy);
                    this._purgeStates(states, numBatch);
                    await this.db.multiIns([
                        [keyAxAy, states],
                    ]);
                    alreadyUpdated.push(hashAxAy);
                }
            }
        }

        // Reset numBatch-AxAy for future states
        const keysToDel = [];
        for (let i = this.lastBatch; i > numBatch; i--){
            const keyNumBatchAxAy = Scalar.add(Constants.DB_NumBatch_AxAy, i);
            keysToDel.push(keyNumBatchAxAy);
        }
        await this.db.multiDel(keysToDel);
    }

    /**
     * Updates the array of ethereum address states in the DB deleting all the information above the numBatch
     * @param {Scalar} numBatch - Batch number
     */
    async _updateEthAddr(numBatch) {
        // Update ethAddr states
        const alreadyUpdated = [];
        for (let i = this.lastBatch; i > numBatch; i--){
            const keyNumBatchEthAddr = Scalar.add(Constants.DB_NumBatch_EthAddr, i);
            const ethAddrToUpdate = await this.db.get(keyNumBatchEthAddr);
            if (!ethAddrToUpdate) continue;
            for (const ethAddr of ethAddrToUpdate) {
                if (!alreadyUpdated.includes(ethAddr)){
                    const keyEthAddr = Scalar.add(Constants.DB_EthAddr, ethAddr);
                    const states = await this.db.get(keyEthAddr);
                    this._purgeStates(states, numBatch);
                    await this.db.multiIns([
                        [keyEthAddr, states],
                    ]);
                    alreadyUpdated.push(ethAddr);
                }
            }
        }

        // Reset numBatch-ethAddr for future states
        const keysToDel = [];
        for (let i = this.lastBatch; i > numBatch; i--){
            const keyNumBatchEthAddr = Scalar.add(Constants.DB_NumBatch_EthAddr, i);
            keysToDel.push(keyNumBatchEthAddr);
        }
        await this.db.multiDel(keysToDel);
    }

    /**
     * Removes all the States that are above or equal to the _numBatch
     * @param {Object} states - Object containing an array of states, indexes by batchNum
     * @param {Scalar} _numBatch - Batch number
     */
    async _purgeStates(states, _numBatch){
        const numBatch = Scalar.e(_numBatch);
        if (states.length === 0) return;
        if (Scalar.lt(states.slice(-1)[0], numBatch)) return;
        if (Scalar.gt(states[0], numBatch)) {
            states.splice(0, states.length);
            return;
        }
        let indexFound = null;
        for (let i = states.length - 1; i >= 0; i--){
            if (Scalar.leq(states[i], numBatch)){
                indexFound = i+1;
                break;
            }
        }
        if (indexFound !== null){
            states.splice(indexFound);
        }
    }

    /**
     * Compute babyjubjub compressed address
     * @param {String} ax - babyjubjub ax coordinate encoded as hexadecimal string
     * @param {String} ay - babyjubjub ay coordinate encoded as hexadecimal string
     * @returns {String} Compressed babyjubjub address encoded as hexadecimal string
     */
    pointToCompress(axStr, ayStr){
        const ax = Scalar.fromString(axStr, 16);
        const ay = Scalar.fromString(ayStr, 16);
        const compress = BabyJubJub.packPoint([ax, ay]);

        return `0x${compress.toString("hex")}`;
    }

    /**
     * get babyjubjub sign
     * @param {String} axStr - babyjubjub ax coordinate encoded as hexadecimal string
     * @param {String} ayStr - babyjubjub ay coordinate encoded as hexadecimal string
     * @returns {Number} babyjubjub sign
     */
    getSignFromAxAy(axStr, ayStr){
        const ax = Scalar.fromString(axStr, 16);
        const ay = Scalar.fromString(ayStr, 16);
        const compressedBuff = BabyJubJub.packPoint([ax, ay]);

        let sign = 0;
        if (compressedBuff[31] & 0x80) {
            sign = 1;
        }
        return sign;
    }
}

module.exports = async function(db, chainID) {
    const master = await db.get(Constants.DB_Master);
    if (!master) {
        const setChainID = chainID || Constants.defaultChainID;
        await db.multiIns([
            [ Constants.DB_ChainID, setChainID]
        ]);
        return new RollupDB(db, 0, Scalar.e(0), Constants.firstIdx, setChainID);
    }
    const roots = await db.get(Scalar.add(Constants.DB_Batch, Scalar.e(master)));
    const initialIdx = await db.get(Scalar.add(Constants.DB_InitialIdx, Scalar.e(master)));
    const dBchainID = await db.get(Scalar.add(Constants.DB_ChainID));
    if (!roots) {
        throw new Error("Database corrupted");
    }
    return new RollupDB(db, master, roots[0], initialIdx, dBchainID);
};
