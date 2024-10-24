const {
    ethers
} = require('ethers');
require('dotenv').config();
const fs = require('fs');
const path = require('path');

const SybilAccount = require("../helpers/babyjub/sybilAccount/sybil-account");
const {
    l1UserTxCreateAccountDeposit
} = require('../helpers/contractFunctions/addL1Tx');

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

let testFun = async () => {
    const account = new SybilAccount();
    const accountInfo = await account.initialize();

    l1UserTxCreateAccountDeposit(1000, accountInfo.bjjCompressed)
}



// async function executeAddL1Transaction() {
//     const account = new SybilAccount();
//     const accountInfo = await account.initialize();

//     const babyPubKey = `0x${accountInfo.bjjCompressed}`;
//     console.log("babay key:", babyPubKey);
//     const fromIdx = 256;
//     const loadAmountF = 1000;
//     const amountF = 0;
//     const toIdx = 0;

//     try {
//         let loadAmount = loadAmountF * 10 ** (18 - 8);

//         const tx = await contract.addL1Transaction(babyPubKey, fromIdx, loadAmountF, amountF, toIdx, {
//             value: loadAmount
//         });
//         console.log("Transaction sent:", tx.hash);
//         const receipt = await tx.wait();
//         console.log("Transaction confirmed in block:", receipt.blockNumber);
//     } catch (error) {
//         console.error("Error executing addL1Transaction:", error);
//     }

// try {
//     // Example of calling a view function (read-only)
//     const result = await contract.getLastForgedBatch();
//     console.log("Function call result:", result);
// } catch (error) {
//     console.error("Error calling get function:", error);
// }
// }

// async function executeForgeBatch() {
//     const newLastIdx = 123456n;
//     const newStRoot = 9876543210123456789n;
//     const newVouchRoot = 567890123456789n;
//     const newScoreRoot = 123450987654321n;
//     const newExitRoot = 876543210123456789n;
//     const verifierIdx = 1;
//     const l1Batch = true;
//     const proofA = [0, 0]; // Example proof arrays
//     const proofB = [
//         [0, 0],
//         [0, 0]
//     ];
//     const proofC = [0, 0];
//     const input = 1234567890123456789n;

//     try {
//         const tx = await contract.forgeBatch(
//             newLastIdx, 
//             newStRoot, 
//             newVouchRoot, 
//             newScoreRoot, 
//             newExitRoot, 
//             verifierIdx, 
//             l1Batch, 
//             proofA, 
//             proofB, 
//             proofC, 
//             input
//         );
//         console.log("Transaction sent:", tx.hash);
//         const receipt = await tx.wait();
//         console.log("Transaction confirmed in block:", receipt.blockNumber);
//     } catch (error) {
//         console.error("Error executing forgeBatch:", error);
//     }
// }

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
    await testFun();
    // await executeForgeBatch();
    // await executeWithdrawMerkleProof();
})();