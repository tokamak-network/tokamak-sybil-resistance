const assert = require("assert");
const Scalar = require("ffjavascript").Scalar;
// const utilsScalar = require("ffjavascript").utils;
const poseidonHash = require("circomlib").poseidon;
// const babyJub = require("circomlib").babyJub;
// const SMT = require("circomlib").SMT;

const SMTTmpDb = require("./smt-tmp-db");
const feeUtils = require("./fee-table");
const utils = require("../utils/utils");
const stateUtils = require("../utils/state-utils");
const txUtils = require("../utils/tx-utils");
const float40 = require("../utils/float40");
const Constants = require("../utils/constants");
const { SMT } = require("circomlibjs");

module.exports = class BatchBuilder {
    constructor(rollupDB, batchNumber, root, initialIdx, maxNTx, nLevels, maxL1Tx, chainID, nFeeTx) {
        assert((nLevels % 8) == 0);
        this.rollupDB = rollupDB;
        this.batchNumber = batchNumber;
        this.finalIdx = initialIdx;
        this.maxNTx = maxNTx || 4;
        this.maxL1Tx = maxL1Tx;
        this.nLevels = nLevels;
        this.offChainTxs = [];
        this.onChainTxs = [];
        this.dbState = new SMTTmpDb(rollupDB.db);
        this.stateTree = new SMT(this.dbState, root);
        this.dbExit = new SMTTmpDb(rollupDB.db);
        this.exitTree = new SMT(this.dbExit, Scalar.e(0));
        // this.totalFeeTransactions = nFeeTx;
        // this.feePlanTokens = Array(this.totalFeeTransactions).fill(0);
        // this.feeTotals = Array(this.totalFeeTransactions).fill(0);
        // this.feeIdxs = Array(this.totalFeeTransactions).fill(0);
        // this.nTokens = 0;
        // this.nFeeIdxs = 0;
        this.currentNumBatch = Scalar.e(batchNumber);
        this.chainID = chainID;

        this._initBits();
    }

    /**
     * Initialize parameters
     */
    _initBits(){
        this.maxIdxB = Constants.maxNlevels;
        this.idxB = this.nLevels;
        this.rootB = 256;
        this.chainIDB = 16;
        this.fromEthAddrB = 160;
        this.fromBjjCompressedB = 256;
        this.f40B = 40;
        this.tokenIDB = 32;
        this.feeB = 8;
        this.numBatchB = 32;

        this.L1TxFullB = this.fromEthAddrB + this.fromBjjCompressedB + 2*this.maxIdxB + this.tokenIDB + 2*this.f40B;
        this.L1L2TxDataB = 2*this.idxB + this.f40B + this.feeB;

        const inputL1TxsFullB = this.maxL1Tx * this.L1TxFullB;
        const inputTxsDataB = this.maxNTx * this.L1L2TxDataB;
        const inputFeeTxsB = this.totalFeeTransactions * this.idxB;

        this.sha256InputsB = 2*this.idxB + 3*this.rootB + this.chainIDB + inputL1TxsFullB + inputTxsDataB + inputFeeTxsB;
    }

    /**
     * Add an empty transaction to the batch
     */
    _addNopTx() {
        const i = this.input.txCompressedData.length;
        const txNop = {
            amount: 0,
            tokenID: 0,
            nonce: 0,
            userFee: 0,
            rqOffset: 0,
            onChain: 0,
            newAccount: 0,
            chainID: this.chainID,
        };
        this.input.txCompressedData[i] = txUtils.buildTxCompressedData(txNop);
        this.input.amountF[i] = 0;
        this.input.txCompressedDataV2[i] = txUtils.buildTxCompressedDataV2(txNop);
        this.input.fromIdx[i] = 0;
        this.input.auxFromIdx[i] = 0;
        this.input.toIdx[i] = 0;
        this.input.auxToIdx[i] = 0;
        this.input.maxNumBatch[i] = 0;

        // to
        this.input.toEthAddr[i] = 0,
        this.input.toBjjAy[i] = 0,

        this.input.rqOffset[i] = 0;
        this.input.onChain[i] = 0;
        this.input.newAccount[i] = 0;

        // rqData
        this.input.rqTxCompressedDataV2[i] = 0;
        this.input.rqToEthAddr[i] = 0;
        this.input.rqToBjjAy[i] = 0;

        this.input.s[i] = 0;
        this.input.r8x[i] = 0;
        this.input.r8y[i] = 0;

        // on-chain
        this.input.fromEthAddr[i] = 0;
        this.input.loadAmountF[i] = 0;
        this.input.fromBjjCompressed[i] = [];
        for (let j = 0; j < 256 ; j++) {
            this.input.fromBjjCompressed[i][j] = 0;
        }

        // State 1
        this.input.sign1[i] = 0;
        this.input.ay1[i] = 0;
        this.input.balance1[i] = 0;
        this.input.nonce1[i] = 0;
        this.input.tokenID1[i] = 0;
        this.input.ethAddr1[i] = 0;
        this.input.siblings1[i] = [];
        for (let j=0; j<this.nLevels+1; j++) {
            this.input.siblings1[i][j] = 0;
        }
        this.input.isOld0_1[i] = 0;
        this.input.oldKey1[i] = 0;
        this.input.oldValue1[i] = 0;

        // State 2
        this.input.sign2[i] = 0;
        this.input.ay2[i] = 0;
        this.input.balance2[i] = 0;
        this.input.newExit[i] = 0;
        this.input.nonce2[i] = 0;
        this.input.tokenID2[i] = 0;
        this.input.ethAddr2[i] = 0;
        this.input.siblings2[i] = [];
        for (let j = 0; j < this.nLevels+1; j++) {
            this.input.siblings2[i][j] = 0;
        }
        this.input.isOld0_2[i] = 0;
        this.input.oldKey2[i] = 0;
        this.input.oldValue2[i] = 0;

        if (i < this.maxNTx-1) {
            this.input.imOnChain[i] = 0;
            this.input.imOutIdx[i] = this.finalIdx;

            this.input.imStateRoot[i] = this.stateTree.root;
            this.input.imExitRoot[i] = this.exitTree.root;
            this.input.imAccFeeOut[i] = [];
            for (let j = 0; j < this.totalFeeTransactions; j++) {
                this.input.imAccFeeOut[i][j] = (i == 0) ? 0: this.input.imAccFeeOut[i-1][j];
            }
        }
    }

    /**
     * Process a transaction, update the DB and add the proper inputs for the circuit
     * @param {Object} tx - Object transaction
     */
    async _addTx(tx) {
        const i = this.input.txCompressedData.length;

        // Set auxFromIdx if new account is created
        if (tx.fromIdx == Constants.nullIdx){
            this._addAuxFromIdx(tx);
        }

        // Set auxToIdx if L2 transfer is set to nullIdx
        if (!tx.onChain){
            if (tx.toIdx == Constants.nullIdx){
                await this._addAuxToIdx(tx);
            }
        }

        // Round values
        const amountF = tx.amountF || float40.fix2Float(tx.amount || 0);
        const amount = float40.float2Fix(amountF);
        tx.amountF = amountF;
        tx.amount = amount;

        // Get ay and sign from bjjCompressed
        // let fromAx = 0;
        let fromAy = 0;
        let fromSign = 0;
        if (tx.fromBjjCompressed != undefined && Scalar.eq(0, tx.fromIdx)){
            // const bjjCompresedBuf = utilsScalar.leInt2Buff(Scalar.fromString(tx.fromBjjCompressed || "0", 16));
            // const pointBjj = babyJub.unpackPoint(bjjCompresedBuf);
            // fromAx = pointBjj[0].toString(16);
            const scalarFromBjjCompressed = Scalar.fromString(tx.fromBjjCompressed || "0", 16);
            fromAy = utils.extract(scalarFromBjjCompressed, 0, 254).toString(16);
            fromSign = utils.extract(scalarFromBjjCompressed, 255, 1);
        }

        let loadAmount = Scalar.e(float40.float2Fix(tx.loadAmountF || 0));
        if ((!tx.onChain) && (Scalar.gt(loadAmount, 0))){
            throw new Error("Load amount must be 0 for L2 txs");
        }

        // check max num batch
        const maxNumBatchScalar = Scalar.e(tx.maxNumBatch || 0);
        if (Scalar.gt(maxNumBatchScalar, Scalar.e(0))){
            if (!Scalar.geq(maxNumBatchScalar, this.currentNumBatch)){
                throw new Error("maxNumBatch must be less than currentBatch");
            }
        }

        let oldState1;
        let oldState2;
        let op1 = "NOP";
        let op2 = "INSERT";
        let isExit;
        let newAccount = 0;

        // From leaf
        const resFind1 = await this.stateTree.find(tx.fromIdx);
        if (resFind1.found) {
            const foundValueId = poseidonHash([resFind1.foundValue, tx.fromIdx]);
            oldState1 = stateUtils.array2State(await this.dbState.get(foundValueId));
            op1 = "UPDATE";
        } else {
            // INSERT
            // tx needs (fromAy, fromSign) get from BjjCompressed
            oldState1 = {
                balance: Scalar.e(0),
                tokenID: tx.tokenID,
                nonce: 0,
                sign: fromSign,
                ay: fromAy,
                ethAddr: tx.fromEthAddr
            };
            op1 = "INSERT";
            newAccount = 1;
        }

        // To leaf
        let resFind2;
        let resFindExit;
        const finalToIdx = (tx.toIdx == Constants.nullIdx) ? (tx.auxToIdx || 0) : tx.toIdx;
        if (finalToIdx > Constants.firstIdx) {
            resFind2 = await this.stateTree.find(finalToIdx);
            if (!resFind2.found) {
                throw new Error("trying to send to a non existing account");
            }
            const foundValueId = poseidonHash([resFind2.foundValue, finalToIdx]);
            oldState2 = stateUtils.array2State(await this.dbState.get(foundValueId));
            isExit = false;
            op2 = "UPDATE";
        } else if (tx.toIdx == Constants.exitIdx) {
            resFindExit = await this.exitTree.find(tx.fromIdx);
            if (resFindExit.found) {
                const foundValueId = poseidonHash([resFindExit.foundValue, tx.fromIdx]);
                oldState2 = stateUtils.array2State(await this.dbExit.get(foundValueId));
                op2 = "UPDATE";
            } else {
                oldState2 = {
                    balance: Scalar.e(0),
                    tokenID: oldState1.tokenID,
                    nonce: 0,
                    sign: oldState1.sign,
                    ay: oldState1.ay,
                    ethAddr: oldState1.ethAddr
                };
                op2 = "INSERT";
            }
            isExit = true;
        } else {
            op2 = "NOP";
        }

        let fee2Charge;
        if (tx.onChain){
            fee2Charge = Scalar.e(0);
        } else {
            fee2Charge = feeUtils.computeFee(tx.amount, tx.userFee);
        }

        // compute applyEthAddr
        let nullifyLoadAmount = false;
        let nullifyAmount = false;
        let applyNullifierEthAddr = false;
        let applyNullifierTokenID1 = false;
        let applyNullifierTokenID2 = false;

        if (tx.onChain){
            if (Scalar.gt(amount, 0) && newAccount == false){
                if (tx.fromEthAddr != oldState1.ethAddr){
                    applyNullifierEthAddr = true;
                }
            }

            if (newAccount == false){
                if (tx.tokenID != oldState1.tokenID){
                    applyNullifierTokenID1 = true;
                }
            }

            if (Scalar.gt(amount, 0) && op2 != "INSERT"){
                if (tx.tokenID != oldState2.tokenID){
                    applyNullifierTokenID2 = true;
                }
            }
        }

        if (applyNullifierTokenID1 && Scalar.gt(loadAmount, 0)){
            nullifyLoadAmount = true;
        }

        let applyNullifierTokenID1ToAmount = false;
        if (applyNullifierTokenID1 && Scalar.gt(amount, 0)){
            applyNullifierTokenID1ToAmount = true;
        }

        if (applyNullifierTokenID1ToAmount || applyNullifierEthAddr || applyNullifierTokenID2){
            nullifyAmount = true;
        }

        let effectiveLoadAmount = loadAmount;
        if (nullifyLoadAmount){
            effectiveLoadAmount = Scalar.e(0);
        }

        let effectiveAmount = amount;
        if (nullifyAmount){
            effectiveAmount = Scalar.e(0);
        }

        const underFlowOk = Scalar.geq(Scalar.sub( Scalar.sub( Scalar.add(oldState1.balance, effectiveLoadAmount), amount), fee2Charge), 0);
        if (!underFlowOk) {
            if (tx.onChain) {
                effectiveAmount = Scalar.e(0);
            } else {
                let errStr = "Error ";
                if (!underFlowOk) errStr = "underflow";
                throw new Error(errStr);
            }
        }

        // save isAmountNullified for each transaction
        tx.effectiveAmount = effectiveAmount;
        tx.isAmountNullified = !(!nullifyAmount && underFlowOk);

        // TX INPUT CIRCUIT
        this.input.fromIdx[i] = tx.fromIdx;
        this.input.auxFromIdx[i] = tx.auxFromIdx || 0;
        this.input.toIdx[i] = tx.toIdx;
        this.input.auxToIdx[i] = tx.auxToIdx || 0;
        this.input.txCompressedData[i] = txUtils.buildTxCompressedData(Object.assign({newAccount: newAccount}, tx));
        this.input.amountF[i] = tx.amountF;
        this.input.txCompressedDataV2[i] = tx.onChain ?  0 : txUtils.buildTxCompressedDataV2(tx);
        this.input.toEthAddr[i] = Scalar.fromString(tx.toEthAddr || "0", 16);
        if (tx.toBjjSign !== undefined && tx.toBjjAy !== undefined){
            // this.input.toBjjAx[i]= Scalar.fromString(stateUtils.getAx(tx.toBjjSign, tx.toBjjAy), 16);
            this.input.toBjjAy[i]= Scalar.fromString(tx.toBjjAy, 16);
        } else {
            // this.input.toBjjAx[i]= 0;
            this.input.toBjjAy[i]= 0;
        }
        this.input.maxNumBatch[i] = tx.maxNumBatch || 0;
        this.input.onChain[i] = tx.onChain ? 1 : 0;
        this.input.newAccount[i] = newAccount;
        this.input.rqOffset[i] = tx.rqOffset || 0;

        // rqData
        this.input.rqTxCompressedDataV2[i] = tx.rqTxCompressedDataV2 || 0;
        this.input.rqToEthAddr[i] = tx.rqToEthAddr || 0;
        this.input.rqToBjjAy[i] = tx.rqToBjjAy || 0;

        this.input.s[i] = tx.s || 0;
        this.input.r8x[i] = tx.r8x || 0;
        this.input.r8y[i] = tx.r8y || 0;
        this.input.loadAmountF[i] = tx.loadAmountF || 0;
        this.input.fromEthAddr[i] = Scalar.fromString(tx.fromEthAddr || "0", 16); //Scalar.fromString(oldState1.ethAddr, 16);
        // Input bits BjjCompressed
        this.input.fromBjjCompressed[i] = [];
        const fromBjjCompressedScalar = Scalar.fromString(tx.fromBjjCompressed || "0", 16);
        const bjjCompressedBits = Scalar.bits(fromBjjCompressedScalar);
        while (bjjCompressedBits.length < 256) bjjCompressedBits.push(0);
        for (let j = 0; j < 256; j++){
            this.input.fromBjjCompressed[i][j] = bjjCompressedBits[j];
        }

        const newState1 = Object.assign({}, oldState1);
        newState1.balance = Scalar.sub(Scalar.sub(Scalar.add(oldState1.balance, effectiveLoadAmount), effectiveAmount), fee2Charge);
        if (!tx.onChain) {
            if (oldState1.nonce != tx.nonce) {
                throw new Error("invalid nonce");
            }
            newState1.nonce++;
            this._accumulateFees(tx.tokenID, fee2Charge);
        }

        if (tx.fromIdx === finalToIdx)
            oldState2 = Object.assign({}, newState1);

        let newState2;
        if (op2 != "NOP"){
            newState2 = Object.assign({}, oldState2);
            newState2.balance = Scalar.add(oldState2.balance, effectiveAmount);
        }

        if (op1=="INSERT") {

            this.finalIdx += 1;

            const newValue = stateUtils.hashState(newState1);

            const res = await this.stateTree.insert(tx.auxFromIdx, newValue);
            let siblings = res.siblings;
            while (siblings.length<this.nLevels+1) siblings.push(Scalar.e(0));

            // State 1
            // That first 4 parameters do not matter in the circuit, since it gets the information from the TxData
            this.input.sign1[i]= 0x1234;      // It should not matter
            this.input.ay1[i]= 0x1234;      // It should not matter
            this.input.balance1[i]= 0x1234;  // It should not matter
            this.input.nonce1[i]= 0x1234;   // It should not matter
            this.input.tokenID1[i]= tx.tokenID;   // Must match with tokenID transaction
            this.input.ethAddr1[i]= this.input.fromEthAddr[i]; // In the onChain TX this must match
            this.input.siblings1[i] = siblings;
            this.input.isOld0_1[i]= res.isOld0 ? 1 : 0;
            this.input.oldKey1[i]= res.isOld0 ? 0 : res.oldKey;
            this.input.oldValue1[i]= res.isOld0 ? 0 : res.oldValue;

            // Database AxAy
            const keyAxAy = Scalar.add( Scalar.add(Constants.DB_AxAy, fromSign), Scalar.fromString(fromAy, 16));
            const lastAxAyStates = await this.dbState.get(keyAxAy);

            // get last state and add last batch number
            let valStatesAxAy;
            let lastAxAyState;
            if (!lastAxAyStates) {
                lastAxAyState = null;
                valStatesAxAy = [];
            }
            else {
                valStatesAxAy = [...lastAxAyStates];
                lastAxAyState = valStatesAxAy.slice(-1)[0];
            }
            if (!valStatesAxAy.includes(this.currentNumBatch)){
                valStatesAxAy.push(this.currentNumBatch);
                await this.dbState.multiIns([
                    [keyAxAy, valStatesAxAy],
                ]);
            }

            // get last state
            let valOldAxAy = null;
            if (lastAxAyState){
                const keyOldAxAyBatch = poseidonHash([keyAxAy, lastAxAyState]);
                valOldAxAy = await this.dbState.get(keyOldAxAyBatch);
            }

            let newValAxAy;
            if (!valOldAxAy) newValAxAy = [];
            else newValAxAy = [...valOldAxAy];
            newValAxAy.push(Scalar.e(tx.auxFromIdx));
            // new key newValAxAy
            const newKeyAxAyBatch = poseidonHash([keyAxAy, this.currentNumBatch]);
            await this.dbState.multiIns([
                [newKeyAxAyBatch, newValAxAy],
            ]);

            // Database Ether address
            const keyEth = Scalar.add(Constants.DB_EthAddr, this.input.fromEthAddr[i]);
            const lastEthStates = await this.dbState.get(keyEth);

            // get last state and add last batch number
            let valStatesEth;
            let lastEthState;
            if (!lastEthStates) {
                lastEthState = null;
                valStatesEth = [];
            } else {
                valStatesEth = [...lastEthStates];
                lastEthState = valStatesEth.slice(-1)[0];
            }
            if (!valStatesEth.includes(this.currentNumBatch)){
                valStatesEth.push(this.currentNumBatch);
                await this.dbState.multiIns([
                    [keyEth, valStatesEth],
                ]);
            }

            // get last state
            let valOldEth = null;
            if (lastEthState){
                const keyOldEthBatch = poseidonHash([keyEth, lastEthState]);
                valOldEth = await this.dbState.get(keyOldEthBatch);
            }

            let newValEth;
            if (!valOldEth) newValEth = [];
            else newValEth = [...valOldEth];
            newValEth.push(Scalar.e(tx.auxFromIdx));

            // new key newValEth
            const newKeyEthBatch = poseidonHash([keyEth, this.currentNumBatch]);

            await this.dbState.multiIns([
                [newKeyEthBatch, newValEth],
            ]);

            // Database Idx
            // get array of states saved by batch
            const lastIdStates = await this.dbState.get(Scalar.add(Constants.DB_Idx, tx.auxFromIdx));
            // add last batch number
            let valStatesId;
            if (!lastIdStates) valStatesId = [];
            else valStatesId = [...lastIdStates];
            if (!valStatesId.includes(this.currentNumBatch)) valStatesId.push(this.currentNumBatch);

            // new state for idx
            const newValueId = poseidonHash([newValue, tx.auxFromIdx]);

            // new entry according idx and batchNumber
            const keyIdBatch = poseidonHash([tx.auxFromIdx, this.currentNumBatch]);

            await this.dbState.multiIns([
                [newValueId, stateUtils.state2Array(newState1)],
                [keyIdBatch, newValueId],
                [Scalar.add(Constants.DB_Idx, tx.auxFromIdx), valStatesId],
            ]);
        } else if (op1 == "UPDATE") {
            const newValue = stateUtils.hashState(newState1);

            const res = await this.stateTree.update(tx.fromIdx, newValue);
            let siblings = res.siblings;
            while (siblings.length<this.nLevels+1) siblings.push(Scalar.e(0));

            // State 1
            //It should not matter what the Tx have, because we get the input from the oldState
            this.input.sign1[i]= Scalar.e(oldState1.sign);
            this.input.ay1[i]= Scalar.fromString(oldState1.ay, 16);
            this.input.tokenID1[i]= Scalar.e(oldState1.tokenID);
            this.input.balance1[i]= oldState1.balance;
            this.input.nonce1[i]= oldState1.nonce;
            this.input.ethAddr1[i]= Scalar.fromString(oldState1.ethAddr, 16);

            this.input.siblings1[i] = siblings;
            this.input.isOld0_1[i]= 0;
            this.input.oldKey1[i]= 0x1234;      // It should not matter
            this.input.oldValue1[i]= 0x1234;    // It should not matter

            // get array of states saved by batch
            const lastIdStates = await this.dbState.get(Scalar.add(Constants.DB_Idx, tx.fromIdx));
            // add last batch number
            let valStatesId;
            if (!lastIdStates) valStatesId = [];
            else valStatesId = [...lastIdStates];
            if (!valStatesId.includes(this.currentNumBatch)) valStatesId.push(this.currentNumBatch);

            // new state for idx
            const newValueId = poseidonHash([newValue, tx.fromIdx]);

            // new entry according idx and batchNumber
            const keyIdBatch = poseidonHash([tx.fromIdx, this.currentNumBatch]);

            await this.dbState.multiIns([
                [newValueId, stateUtils.state2Array(newState1)],
                [keyIdBatch, newValueId],
                [Scalar.add(Constants.DB_Idx, tx.fromIdx), valStatesId]
            ]);
        }

        if (op2=="INSERT") {
            const newValue = stateUtils.hashState(newState2);

            const res = await this.exitTree.insert(tx.fromIdx, newValue);
            if (res.found) {
                throw new Error("Invalid Exit account");
            }
            let siblings = res.siblings;
            while (siblings.length<this.nLevels+1) siblings.push(Scalar.e(0));

            // State 1
            this.input.sign2[i] = 0x1234;    // It should not matter
            this.input.ay2[i] = 0x1234;      // It should not matter
            this.input.balance2[i] = 0x1234;      // must be 0 when inserting
            this.input.nonce2[i] = 0x1234;   // It should not matter
            this.input.tokenID2[i] = 0x1234; // It should not matter
            this.input.newExit[i] = 1;       // must be 1 to signal new exit leaf
            this.input.ethAddr2[i] = this.input.fromEthAddr[i]; // In the onChain TX this must match
            this.input.siblings2[i] = siblings;
            this.input.isOld0_2[i] = res.isOld0 ? 1 : 0;
            this.input.oldKey2[i] = res.isOld0 ? 0 : res.oldKey;
            this.input.oldValue2[i] = res.isOld0 ? 0 : res.oldValue;

            const newValueId = poseidonHash([newValue, tx.fromIdx]);
            await this.dbExit.multiIns([[newValueId, stateUtils.state2Array(newState2)]]);

        } else if (op2=="UPDATE") {
            if (isExit) {
                const newValue = stateUtils.hashState(newState2);

                const res = await this.exitTree.update(tx.fromIdx, newValue);
                let siblings = res.siblings;
                while (siblings.length<this.nLevels+1) siblings.push(Scalar.e(0));

                // State 2
                //It should not matter what the Tx have, because we get the input from the oldState
                this.input.sign2[i]= Scalar.e(oldState2.sign);
                this.input.ay2[i]= Scalar.fromString(oldState2.ay, 16);
                this.input.balance2[i]= oldState2.balance;
                this.input.newExit[i]= Scalar.e(0);
                this.input.nonce2[i]= oldState2.nonce;
                this.input.tokenID2[i]= Scalar.e(oldState2.tokenID);
                this.input.ethAddr2[i]= Scalar.fromString(oldState2.ethAddr, 16);

                this.input.siblings2[i] = siblings;
                this.input.isOld0_2[i]= 0;
                this.input.oldKey2[i]= 0x1234;      // It should not matter
                this.input.oldValue2[i]= 0x1234;    // It should not matter

                const newValueId = poseidonHash([newValue, tx.fromIdx]);
                const oldValueId = poseidonHash([resFindExit.foundValue, tx.fromIdx]);
                await this.dbExit.multiDel([oldValueId]);
                await this.dbExit.multiIns([[newValueId, stateUtils.state2Array(newState2)]]);
            } else {
                const newValue = stateUtils.hashState(newState2);

                const res = await this.stateTree.update(finalToIdx, newValue);
                let siblings = res.siblings;
                while (siblings.length<this.nLevels+1) siblings.push(Scalar.e(0));

                // State 2
                //It should not matter what the Tx have, because we get the input from the oldState
                this.input.sign2[i]= Scalar.e(oldState2.sign);
                this.input.ay2[i]= Scalar.fromString(oldState2.ay, 16);
                this.input.balance2[i]= oldState2.balance;
                this.input.newExit[i] = 0;
                this.input.nonce2[i]= oldState2.nonce;
                this.input.tokenID2[i]= Scalar.e(oldState2.tokenID);
                this.input.ethAddr2[i]= Scalar.fromString(oldState2.ethAddr, 16);


                this.input.siblings2[i] = siblings;
                this.input.isOld0_2[i]= 0;
                this.input.oldKey2[i]= 0x1234;      // It should not matter
                this.input.oldValue2[i]= 0x1234;    // It should not matter

                // get array of states saved by batch
                const lastIdStates = await this.dbState.get(Scalar.add(Constants.DB_Idx, finalToIdx));
                // add last batch number
                let valStatesId;
                if (!lastIdStates) valStatesId = [];
                else valStatesId = [...lastIdStates];
                if (!valStatesId.includes(this.currentNumBatch)) valStatesId.push(this.currentNumBatch);

                // new state for idx
                const newValueId = poseidonHash([newValue, finalToIdx]);

                // new entry according idx and batchNumber
                const keyIdBatch = poseidonHash([finalToIdx, this.currentNumBatch]);

                await this.dbState.multiIns([
                    [newValueId, stateUtils.state2Array(newState2)],
                    [keyIdBatch, newValueId],
                    [Scalar.add(Constants.DB_Idx, finalToIdx), valStatesId]
                ]);
            }
        } else if (op2=="NOP") {
            // State 2
            this.input.sign2[i]= 0;
            this.input.ay2[i]= 0;
            this.input.balance2[i]= 0;
            this.input.newExit[i]= 0;
            this.input.nonce2[i]= 0;
            this.input.tokenID2[i] = 0;
            this.input.ethAddr2[i]= 0;
            this.input.siblings2[i] = [];
            for (let j=0; j<this.nLevels+1; j++) {
                this.input.siblings2[i][j]= 0;
            }
            this.input.isOld0_2[i]= 0;
            this.input.oldKey2[i]= 0;
            this.input.oldValue2[i]= 0;
        }

        // intermediary signals
        if (i < this.maxNTx-1) {
            if (tx.onChain) {
                this.input.imOnChain[i] = Scalar.e(1);
            } else {
                this.input.imOnChain[i] = 0;
            }
            this.input.imOutIdx[i] = this.finalIdx;

            this.input.imStateRoot[i] = this.stateTree.root;
            this.input.imExitRoot[i] = this.exitTree.root;
            this.input.imAccFeeOut[i] = [];
            for (let j = 0; j < this.totalFeeTransactions; j++) {
                this.input.imAccFeeOut[i][j] = this.feeTotals[j];
            }
        }

        // Database numBatch - Idx
        const keyNumBatchIdx = Scalar.add(Constants.DB_NumBatch_Idx, this.currentNumBatch);
        let lastBatchIdx = await this.dbState.get(keyNumBatchIdx);

        // get last state and add last batch number
        let newBatchIdx;
        if (!lastBatchIdx) lastBatchIdx = [];
        newBatchIdx = [...lastBatchIdx];

        if (op1 == "INSERT") {
            if (!newBatchIdx.includes(tx.auxFromIdx)) newBatchIdx.push(tx.auxFromIdx);
        }

        if (op1 == "UPDATE") {
            if (!newBatchIdx.includes(tx.fromIdx)) newBatchIdx.push(tx.fromIdx);
        }

        if (op2 == "UPDATE" && !isExit) {
            if (!newBatchIdx.includes(finalToIdx)) newBatchIdx.push(finalToIdx);
        }
        await this.dbState.multiIns([
            [keyNumBatchIdx, newBatchIdx],
        ]);

        // Database NumBatch
        if (op1 == "INSERT") {
            // AxAy
            const hashAxAy = poseidonHash([fromSign, Scalar.fromString(fromAy, 16)]);
            const keyNumBatchAxAy = Scalar.add(Constants.DB_NumBatch_AxAy, this.currentNumBatch);
            let oldStatesAxAy = await this.dbState.get(keyNumBatchAxAy);
            let newStatesAxAy;
            if (!oldStatesAxAy) oldStatesAxAy = [];
            newStatesAxAy = [...oldStatesAxAy];
            if (!newStatesAxAy.includes(hashAxAy)) {
                newStatesAxAy.push(hashAxAy);
                await this.dbState.multiIns([
                    [hashAxAy, [fromSign, fromSign]],
                    [keyNumBatchAxAy, newStatesAxAy],
                ]);
            }
            // EthAddress
            const ethAddr =  this.input.fromEthAddr[i];
            const keyNumBatchEthAddr = Scalar.add(Constants.DB_NumBatch_EthAddr, this.currentNumBatch);
            let oldStatesEthAddr = await this.dbState.get(keyNumBatchEthAddr);
            let newStatesEthAddr;
            if (!oldStatesEthAddr) oldStatesEthAddr = [];
            newStatesEthAddr = [...oldStatesEthAddr];
            if (!newStatesEthAddr.includes(ethAddr)) {
                newStatesEthAddr.push(ethAddr);
                await this.dbState.multiIns([
                    [keyNumBatchEthAddr, newStatesEthAddr],
                ]);
            }
        }
    }

    /**
     * Add nop transaction to collect fees
     * @param {Number} i - fee transaction index
     */
    _addNopTxFee(i){
        // State Fees
        this.input.sign3[i] = 0;
        this.input.ay3[i] = 0;
        this.input.balance3[i] = 0;
        this.input.nonce3[i] = 0;
        this.input.ethAddr3[i] = 0;
        this.input.tokenID3[i] = 0;
        this.input.siblings3[i] = [];
        for (let  j= 0; j < this.nLevels + 1; j++) {
            this.input.siblings3[i][j] = 0;
        }

        if (i < this.totalFeeTransactions-1) {
            this.input.imStateRootFee[i] = this.stateTree.root;
        }
    }

    /**
     * Take all the fees accumulated and transfer them to coordinator idxs
     */
    async _transferFees(){
        this.input.imInitStateRootFee = this.stateTree.root;

        // Fill all fee transactions
        for (let i = 0; i < this.totalFeeTransactions ; i++){

            this.input.imFinalAccFee[i] = this.feeTotals[i];

            // Check if tokenID and feeIdx has been added properly, otherwise add NOP tx
            if ((i < this.nFeeIdxs) && (i < this.nTokens)){

                const feeIdx = this.feeIdxs[i];
                const feeTokenID = this.feePlanTokens[i];
                let op = "NOP";
                let oldState;

                // Checks before update the leaf
                const resFind = await this.stateTree.find(feeIdx);
                if (resFind.found){
                    const foundValueId = poseidonHash([resFind.foundValue, feeIdx]);
                    oldState = stateUtils.array2State(await this.dbState.get(foundValueId));
                    // check tokenID matches with idx provided
                    if (oldState.tokenID == feeTokenID){
                        op = "UPDATE";
                    }
                }

                // Update the leaf if necessary
                if (op == "UPDATE") {
                    const newState = Object.assign({}, oldState);
                    newState.balance = Scalar.add(newState.balance, this.feeTotals[i]);

                    const newValue = stateUtils.hashState(newState);

                    const res = await this.stateTree.update(feeIdx, newValue);
                    let siblings = res.siblings;
                    while (siblings.length < this.nLevels+1) siblings.push(Scalar.e(0));

                    // StateFee i
                    // get the input from the oldState
                    this.input.sign3[i]= Scalar.e(oldState.sign);
                    this.input.ay3[i]= Scalar.fromString(oldState.ay, 16);
                    this.input.balance3[i]= oldState.balance;
                    this.input.nonce3[i]= oldState.nonce;
                    this.input.tokenID3[i]= oldState.tokenID;
                    this.input.ethAddr3[i]= Scalar.fromString(oldState.ethAddr, 16);
                    this.input.siblings3[i] = siblings;

                    // Update DB
                    // get array of states saved by batch
                    const lastIdStates = await this.dbState.get(Scalar.add(Constants.DB_Idx, feeIdx));
                    // add last batch number
                    let valStatesId;
                    if (!lastIdStates) valStatesId = [];
                    else valStatesId = [...lastIdStates];
                    if (!valStatesId.includes(this.currentNumBatch)) valStatesId.push(this.currentNumBatch);

                    // new state for idx
                    const newValueId = poseidonHash([newValue, feeIdx]);

                    // new entry according idx and batchNumber
                    const keyIdBatch = poseidonHash([feeIdx, this.currentNumBatch]);

                    await this.dbState.multiIns([
                        [newValueId, stateUtils.state2Array(newState)],
                        [keyIdBatch, newValueId],
                        [Scalar.add(Constants.DB_Idx, feeIdx), valStatesId]
                    ]);

                    if (i < this.totalFeeTransactions-1) {
                        this.input.imStateRootFee[i] = this.stateTree.root;
                    }

                    // Database numBatch - Idx
                    const keyNumBatchIdx = Scalar.add(Constants.DB_NumBatch_Idx, this.currentNumBatch);
                    let lastBatchIdx = await this.dbState.get(keyNumBatchIdx);

                    // get last state and add last batch number
                    let newBatchIdx;
                    if (!lastBatchIdx) lastBatchIdx = [];
                    newBatchIdx = [...lastBatchIdx];

                    if (!newBatchIdx.includes(feeIdx)) newBatchIdx.push(feeIdx);

                    await this.dbState.multiIns([
                        [keyNumBatchIdx, newBatchIdx],
                    ]);

                } else if (op == "NOP"){
                    this._addNopTxFee(i);
                }
            } else {
                this._addNopTxFee(i);
            }
        }
    }

    /**
     * Add auxiliar identifier to determine fromIdx if new account is created
     * @param {Object} tx - transaction object
     */
    _addAuxFromIdx(tx){
        tx.auxFromIdx = this.finalIdx + 1;
    }

    /**
     * Choose automatically index to deposit the transfer and save it in auxToIdx
     * @param {Object} tx - transaction object
     */
    async _addAuxToIdx(tx){
        const resFind1 = await this.stateTree.find(tx.fromIdx);
        const foundValueId = poseidonHash([resFind1.foundValue, tx.fromIdx]);
        const tokenID1 = stateUtils.array2State(await this.dbState.get(foundValueId)).tokenID;

        // Check if send to ethAddress or Bjj
        if (tx.toEthAddr != Constants.nullEthAddr){
            // Check first temporary states
            const keyEth = Scalar.add(Constants.DB_EthAddr, Scalar.fromString(tx.toEthAddr, 16));
            const newKeyEthBatch = poseidonHash([keyEth, this.currentNumBatch]);
            const newIdxs = await this.dbState.get(newKeyEthBatch);

            // check temporary idxs to find tokenID matching
            if (newIdxs){
                for (let i = newIdxs.length - 1; i >= 0; i--){
                    const resFind2 = await this.stateTree.find(newIdxs[i]);
                    const foundValueId2 = poseidonHash([resFind2.foundValue, newIdxs[i]]);
                    const tokenID2 = stateUtils.array2State(await this.dbState.get(foundValueId2)).tokenID;

                    if (tokenID1 == tokenID2){
                        tx.auxToIdx = newIdxs[i];
                        return;
                    }
                }
            }

            // check in previous consolidated states
            const oldStates = await this.rollupDB.getStateByEthAddr(tx.toEthAddr);
            if (!oldStates) throw new Error("trying to send to a non-existing ethereum address");
            for (let i = oldStates.length - 1; i >= 0; i--){
                const state2 = oldStates[i];
                const tokenID2 = state2.tokenID;

                if (tokenID1 == tokenID2){
                    tx.auxToIdx = state2.idx;
                    return;
                }
            }
            throw new Error("trying to send to a non-existing ethereum address");
        } else {
            // get ax, ay from transaction
            const ay = tx.toBjjAy;
            const sign = tx.toBjjSign;

            // Check first temporary states
            const keyAxAy = Scalar.add( Scalar.add(Constants.DB_AxAy, sign), Scalar.fromString(ay, 16));
            const newKeyAxAyBatch = poseidonHash([keyAxAy, this.currentNumBatch]);
            const newIdxs = await this.dbState.get(newKeyAxAyBatch);

            // check temporary idxs to find tokenID matching
            if (newIdxs){
                for (let i = newIdxs.length - 1; i >= 0; i--){
                    const resFind2 = await this.stateTree.find(newIdxs[i]);
                    const foundValueId2 = poseidonHash([resFind2.foundValue, newIdxs[i]]);
                    const tokenID2 = stateUtils.array2State(await this.dbState.get(foundValueId2)).tokenID;

                    if (tokenID1 == tokenID2){
                        tx.auxToIdx = newIdxs[i];
                        return;
                    }
                }
            }

            // check in previous consolidated states
            const oldStates = await this.rollupDB.getStateBySignAy(sign, ay);
            if (!oldStates) throw new Error("trying to send to a non existing bjj address");
            for (let i = oldStates.length - 1; i >= 0; i--){
                const state2 = oldStates[i];
                const tokenID2 = state2.tokenID;

                if (tokenID1 == tokenID2){
                    tx.auxToIdx = state2.idx;
                    return;
                }
            }
            throw new Error("trying to send to a non existing bjj address");
        }
    }

    /**
     * Add fee of an specific token to the feeTotals object
     * @param {String} tokenID - Token identifier
     * @param {String} fee2Charge - Fee to add
     */
    _accumulateFees(tokenID, fee2Charge){
        // find token index
        const indexToken = this.feePlanTokens.indexOf(tokenID);

        if (indexToken === -1) return;
        this.feeTotals[indexToken] = Scalar.add(this.feeTotals[indexToken], fee2Charge);
    }

    /**
     * Build the batch
     * Adds all the transactions and calculate the inputs for the circuit
     */
    async build() {
        this.txs = [];

        this.input = {
            // inputs for hash input
            oldLastIdx: this.finalIdx,
            oldStateRoot: this.stateTree.root,
            globalChainID: this.chainID,
            currentNumBatch: this.currentNumBatch,
            feeIdxs: [],

            // accumulate fees
            feePlanTokens: [],

            // Intermediary States to parallelize witness computation
            // decode-tx
            imOnChain: [],
            imOutIdx: [],
            // rollup-tx
            imStateRoot: [],
            imExitRoot: [],
            imAccFeeOut: [],
            // fee-tx
            imStateRootFee: [],
            imInitStateRootFee: 0,
            imFinalAccFee: [],

            // transaction L1-L2
            txCompressedData: [],
            amountF: [],
            txCompressedDataV2: [],
            fromIdx: [],
            auxFromIdx: [],
            maxNumBatch: [],

            toIdx: [],
            auxToIdx: [],
            toBjjAy: [],
            toEthAddr: [],

            onChain: [],
            newAccount: [],
            rqOffset: [],

            // transaction L2 request data
            rqTxCompressedDataV2: [],
            rqToEthAddr: [],
            rqToBjjAy: [],

            // transaction L2 signature
            s: [],
            r8x: [],
            r8y: [],

            // transaction L1
            loadAmountF: [],
            fromEthAddr: [],
            fromBjjCompressed: [],

            // state 1
            tokenID1: [],
            nonce1: [],
            sign1: [],
            balance1: [],
            ay1: [],
            ethAddr1: [],
            siblings1: [],
            // Required for inserts and deletes
            isOld0_1: [],
            oldKey1: [],
            oldValue1: [],

            // state 2
            tokenID2: [],
            nonce2: [],
            sign2: [],
            balance2: [],
            ay2: [],
            ethAddr2: [],
            siblings2: [],
            newExit: [],
            // Required for inserts and deletes
            isOld0_2: [],
            oldKey2: [],
            oldValue2: [],

            // fee tx
            // State fees
            tokenID3: [],
            nonce3: [],
            sign3: [],
            balance3: [],
            ay3: [],
            ethAddr3: [],
            siblings3: [],
        };

        if (this.builded) throw new Error("Batch already builded");

        // Add on-chain Tx
        for (let i = 0; i < this.onChainTxs.length; i++) {
            await this._addTx(this.onChainTxs[i]);
            this.txs.push(this.onChainTxs[i]);
        }

        // Add off-chain Tx
        for (let i = 0; i < this.offChainTxs.length; i++) {
            await this._addTx(this.offChainTxs[i]);
            this.txs.push(this.offChainTxs[i]);
        }

        // Add Nop Tx
        for (let i = 0; i < this.maxNTx - this.offChainTxs.length - this.onChainTxs.length; i++) {
            this._addNopTx();
            this.txs.push(0);
        }

        this.stateRootBeforeFees = this.stateTree.root;

        // Add fees Tx
        await this._transferFees();

        // Compute inputs feePlanTokens
        for (let j = 0; j < this.totalFeeTransactions; j++){
            this.input.feePlanTokens.push(this.feePlanTokens[j]);
        }

        // Compute inputs feeIdxs
        for (let j = 0; j < this.totalFeeTransactions; j++){
            this.input.feeIdxs.push(this.feeIdxs[j]);
        }

        this.builded = true;
    }

    /**
     * Return the circuit input
     * @return {Object} Circuit input
     */
    getInput() {
        if (!this.builded) throw new Error("Batch must first be builded");
        return this.input;
    }

    /**
     * Return the last leaf identifier before the batch is builded
     * @return {Scalar} Identifier
     */
    getOldLastIdx() {
        if (!this.builded) throw new Error("Batch must first be builded");
        return this.input.oldLastIdx;
    }

    /**
     * Return the last leaf identifier after the batch is builded
     * @return {Scalar} Identifier
     */
    getNewLastIdx() {
        if (!this.builded) throw new Error("Batch must first be builded");
        return this.finalIdx;
    }

    /**
     * Return the last state root before the batch is builded
     * @return {Scalar} State root
     */
    getOldStateRoot() {
        if (!this.builded) throw new Error("Batch must first be builded");
        return this.input.oldStateRoot;
    }

    /**
     * Return the feePlanTokens object
     * @return {Object} FeePlanTokens object
     */
    getFeePlanTokens() {
        if (!this.builded) throw new Error("Batch must first be builded");
        return this.input.feePlanTokens;
    }

    /**
     * Return the last state root after the batch is builded
     * @return {Scalar} State root
     */
    getNewStateRoot() {
        if (!this.builded) throw new Error("Batch must first be builded");
        return this.stateTree.root;
    }

     /**
     * Return the last state root after the batch is builded
     * @return {Scalar} State root
     */
     getNewVouchRoot() {
        if (!this.builded) throw new Error("Batch must first be builded");
        return this.stateTree.root;
    }

    /**
     * Return the last state root after the batch is builded
     * @return {Scalar} State root
     */
    getNewScoreRoot() {
        if (!this.builded) throw new Error("Batch must first be builded");
        return this.stateTree.root;
    }

    /**
     * Return the last exit root after the batch is builded
     * @return {Scalar} Exit root
     */
    getNewExitRoot() {
        if (!this.builded) throw new Error("Batch must first be builded");
        return this.exitTree.root;
    }

    /**
     * Computes hash of all pretended public inputs
     * @return {Scalar} hash global input
     */
    getHashInputs(){
        if (!this.builded) throw new Error("Batch must first be builded");
        const finalStr = this.getInputsStr();

        return utils.sha256Snark(finalStr);
    }
    /**
     * Computes string in hexadecimal of all pretended public inputs
     * @return {String} Public input string encoded as hexadecimal
     */
    getInputsStr(){
        if (!this.builded) throw new Error("Batch must first be builded");

        const oldLastIdx = this.getOldLastIdx();
        const newLastIdx = this.getNewLastIdx();

        const oldStateRoot = this.getOldStateRoot();
        const newStateRoot = this.getNewStateRoot();
        const newExitRoot = this.getNewExitRoot();

        // L1TxData
        let L1FullTxsData = this.getL1TxsFullData();

        // txsData
        let txsData = this.getL1L2TxsData();

        // feeTxData
        const feeTxsData = this.getFeeTxsData();

        // string hexacecimal chainID
        const chainID = this.chainID;
        let strChainID = utils.padZeros(chainID.toString("16"), this.chainIDB / 4);

        // string hexacecimal currentNumBatch
        const currentNumBatch = this.currentNumBatch;
        let strCurrentNumBatch = utils.padZeros(currentNumBatch.toString("16"), this.numBatchB / 4);

        // string hexacecimal newExitRoot
        let strNewExitRoot = utils.padZeros(newExitRoot.toString("16"), this.rootB / 4);

        // string hexacecimal newStateRoot
        let strNewStateRoot = utils.padZeros(newStateRoot.toString("16"), this.rootB / 4);

        // string hexacecimal oldStateRoot
        let strOldStateRoot = utils.padZeros(oldStateRoot.toString("16"), this.rootB / 4);

        // string newLastIdx and oldLastIdx
        let res = Scalar.e(0);
        res = Scalar.add(res, newLastIdx);
        res = Scalar.add(res, Scalar.shl(oldLastIdx, this.maxIdxB));
        const finalIdxStr = utils.padZeros(res.toString("16"), (2*this.maxIdxB) / 4);

        // build input string
        const finalStr = finalIdxStr.concat(strOldStateRoot).concat(strNewStateRoot).concat(strNewExitRoot)
            .concat(L1FullTxsData).concat(txsData).concat(feeTxsData).concat(strChainID).concat(strCurrentNumBatch);

        return finalStr;
    }

    /**
     * Return the encoded fee data availability
     * @return {String} Encoded fee data encoded as hexadecimal
     */
    getFeeTxsData() {
        if (!this.builded) throw new Error("Batch must first be builded");

        let finalStr = "";

        for (let i = 0; i < this.totalFeeTransactions; i++){
            const idx = this.feeIdxs[i];
            const res = Scalar.e(idx);

            finalStr = finalStr + utils.padZeros(res.toString("16"), this.idxB / 4);
        }
        return finalStr;
    }

    /**
     * Return the encoded data available ready to send to the rollup SC
     * @return {String} Encoded data available as hexadecimal
     */
    getFeeTxsDataSM() {
        return `0x${this.getFeeTxsData()}`;
    }

    /**
     * Return the encoded L1 full data of all used L1 tx
     * Data saved in smart contract for all L1 tx:
     * fromEthAddr | fromBjj-compressed | fromIdx | loadAmountFloat40 | amountFloat40 | tokenID | toIdx
     * @return {String} Encoded data available in hexadecimal
     */
    _L1TxsFullData() {
        if (!this.builded) throw new Error("Batch must first be builded");

        let finalStr = "";

        for (let i = 0; i < this.onChainTxs.length; i++){
            const tx = this.onChainTxs[i];
            finalStr = finalStr + txUtils.encodeL1TxFull(tx, this.nLevels);
        }
        return finalStr;
    }

    /**
     * Return the full L1 data with padding for unused L1 tx
     * @return {String} L1Data encoded as hexadecimal
     */
    getL1TxsFullData() {
        if (!this.builded) throw new Error("Batch must first be builded");

        const dataL1Tx = this._L1TxsFullData();
        const dataL1NopTx = utils.padZeros("", (this.maxL1Tx - this.onChainTxs.length) * (this.L1TxFullB / 4));

        return dataL1Tx.concat(dataL1NopTx);
    }

    /**
     * Return the L1 full tx encoded data available ready to send to the rollup SC
     * @return {String} Encoded data available in hexadecimal
     */
    getL1TxsFullDataSM() {
        return `0x${this._L1TxsFullData()}`;
    }

    /**
     * Return L2 data-availability
     * fromIdx | toIdx | amountF | userFee
     * @return {String} L2 data availability encoded as hexadecimal
     */
    _L2TxsData() {
        if (!this.builded) throw new Error("Batch must first be builded");

        let finalStr = "";

        for (let i = 0; i < this.offChainTxs.length; i++){
            const tx = this.offChainTxs[i];
            finalStr = finalStr + txUtils.encodeL2Tx(tx, this.nLevels);
        }
        return  finalStr;
    }

    /**
     * Return L1 data-availability
     * fromIdx | toIdx | effectiveAmountF | userFee
     * @return {String} L1 data availability encoded as hexadecimal
     */
    _L1TxsData() {
        if (!this.builded) throw new Error("Batch must first be builded");

        let finalStr = "";

        for (let i = 0; i < this.onChainTxs.length; i++){
            const tx = this.onChainTxs[i];
            finalStr = finalStr + txUtils.encodeL1Tx(tx, this.nLevels);
        }
        return  finalStr;
    }

    /**
     * Return nops tx data-availability
     * fromIdx | toIdx | amountF | userFee == 0 | 0 | 0 | 0
     * @return {String} nop tx data availability encoded as hexadecimal
     */
    _nopTxsData() {
        if (!this.builded) throw new Error("Batch must first be builded");

        const dataNopTx = utils.padZeros("",
            (this.maxNTx - this.offChainTxs.length - this.onChainTxs.length) * (this.L1L2TxDataB / 4));
        return  dataNopTx;
    }

    /**
     * Return L1 & L2 data-availability padded with all unused L2 txs
     * @return {String} L1 & L2 data-availability encoded as hexadecimal
     */
    getL1L2TxsData() {
        if (!this.builded) throw new Error("Batch must first be builded");

        const dataL1Tx = this._L1TxsData();
        const dataL2Tx = this._L2TxsData();
        const dataNopTx = this._nopTxsData();

        return dataL1Tx.concat(dataL2Tx).concat(dataNopTx);
    }

    /**
     * Return the L1 & L2 data-availability ready to send to the SC
     * @return {String} L1 & L2 data encoded as hexadecimal
     */
    getL1L2TxsDataSM() {
        if (!this.builded) throw new Error("Batch must first be builded");

        const dataL1Tx = this._L1TxsData();
        const dataL2Tx = this._L2TxsData();

        return `0x${dataL1Tx.concat(dataL2Tx)}`;
    }

    /**
     * Return the states that have been change with the L1 transactions
     * @return {Object} Resulting temporally state
     */
    async getTmpStateOnChain() {
        if (!this.builded) throw new Error("Batch must first be builded");

        const tmpState = {};
        const idxChanges = {};

        for (let i=0; i<this.onChainTxs.length; i++) {
            const fromIdx = this.onChainTxs[i].fromIdx;
            const toIdx = this.onChainTxs[i].toIdx;
            if (idxChanges[fromIdx] == undefined && fromIdx != 0)
                idxChanges[fromIdx] = fromIdx;
            if (idxChanges[toIdx] == undefined && toIdx != 0)
                idxChanges[toIdx] = toIdx;
        }

        for (const idx of Object.keys(idxChanges)) {
            const resFind = await this.stateTree.find(idx);
            const foundValueId = poseidonHash([resFind.foundValue, idx]);
            const st = stateUtils.array2State(await this.dbState.get(foundValueId));
            st.idx = Number(idx);
            tmpState[idx] = st;
        }

        return tmpState;
    }

    /**
     * Add a transaction to the batchbuilder
     * @param {Object} tx - Transaction object
     */
    addTx(tx) {
        if (this.builded) throw new Error("Batch already builded");

        // Check Max Tx
        if (this.onChainTxs.length + this.offChainTxs.length >= this.maxNTx) {
            throw Error("Too many TX per batch");
        }

        if (tx.onChain && (this.onChainTxs.length >= this.maxL1Tx)) {
            throw Error("Too many L1 TX per batch");
        }

        // Add tx
        if (tx.onChain) {
            this.onChainTxs.push(tx);
        } else {
            this.offChainTxs.push(tx);
        }
    }

    /**
     * Add a token identifier to the batchbuilder
     * @param {Scalar} tokenID - Token identifier
     */
    addToken(tokenID) {
        if (this.nTokens >= this.totalFeeTransactions) {
            throw new Error(`Maximum ${this.totalFeeTransactions} tokens per batch`);
        }

        this.feePlanTokens[this.nTokens] = tokenID;
        this.nTokens = this.nTokens + 1;
    }

    /**
     * Add idx to receive accumulate fees
     * @param {Scalar} feeIdx - merkle tree index
     */
    addFeeIdx(feeIdx) {
        if (this.nFeeIdxs >= this.totalFeeTransactions) {
            throw new Error(`Maximum ${this.totalFeeTransactions} tokens per batch`);
        }

        this.feeIdxs[this.nFeeIdxs] = feeIdx;
        this.nFeeIdxs = this.nFeeIdxs + 1;
    }
};
