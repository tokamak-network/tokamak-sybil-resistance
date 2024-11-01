const {
    ethers
} = require('ethers');
require('dotenv').config();
const fs = require('fs');
const path = require('path');
const SybilAccount = require("../helpers/babyjub/sybilAccount/sybil-account");
const {
    l1UserTxCreateAccountDeposit,
    l1UserTxDeposit,
    l1UserTxDepositTransfer,
    l1UserTxCreateAccountDepositTransfer
} = require('../helpers/contractFunctions/addL1Tx');
const RollupDB = require('../helpers/rollupdb/rollup-db');
const {SMTMemDb}  = require("circomlibjs");
const ffjavascript = require("ffjavascript");
const { ForgerTest } = require('../helpers/contractFunctions/forgeBatch')

const provider = new ethers.JsonRpcProvider("https://rpc.thanos-sepolia.tokamak.network");

const privateKey = process.env.PRIVATE_KEY;
const wallet = new ethers.Wallet(privateKey, provider);

// Contract address and ABI (Replace with your own)
const contractAddress = '0x7d530f9b81f7e170472f738296d1b581b88fed31'; // Replace with your contract address
const abiPath = path.join(__dirname, `../out/Sybil.sol/Sybil.json`); // Path to ABI

const abiFile = fs.readFileSync(abiPath, 'utf-8');
const contractABI = JSON.parse(abiFile).abi;

// Create a contract instance
const contract = new ethers.Contract(contractAddress, contractABI, wallet);


let executeL1Txs = async () => {
    const account = new SybilAccount();
    const accountInfo = await account.initialize();

    // l1UserTxCreateAccountDeposit(1000, accountInfo.bjjCompressed) //working

    l1UserTxDeposit(1000, 257); //working

    // l1UserTxDepositTransfer(1000, 256, 257, 100)

    // l1UserTxCreateAccountDepositTransfer(1000, 256, 10, accountInfo.bjjCompressed)
}


async function executeForgeBatch() {
    const maxL1Tx = 256;
    const maxTx = 512;
    const nLevels = 32;

    const F = ffjavascript.F1Field; // Or appropriate finite field

    let chainId = 111551119090;

    const rollupDB = await RollupDB(new SMTMemDb(F), chainId);
    const forgerTest = new ForgerTest(
        maxTx,
        maxL1Tx,
        nLevels,
        rollupDB
    );

    await forgerTest.forgeBatch(true, [], []);
}

async function executeWithdrawMerkleProof() {
    const amount = 1234567890123456789n; // Example values
    const babyPubKey = 9876543210123456789n;
    const numExitRoot = 12345;
    const siblings = [1234567890123456789n, 9876543210123456789n]; // Example Merkle proof siblings
    const idx = 654321n;

    try {
        const tx = await contract.withdrawMerkleProof(
            amount,
            babyPubKey,
            numExitRoot,
            siblings,
            idx
        );
        console.log("Transaction sent:", tx.hash);
        const receipt = await tx.wait();
        console.log("Transaction confirmed in block:", receipt.blockNumber);
    } catch (error) {
        console.error("Error executing withdrawMerkleProof:", error);
    }
}

// Run one of the functions
(async () => {
    // await executeL1Txs();
    await executeForgeBatch();
    // await executeWithdrawMerkleProof();
})();