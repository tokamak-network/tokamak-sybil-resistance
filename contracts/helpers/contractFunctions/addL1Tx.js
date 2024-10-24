require('dotenv').config();
const {
  ethers
} = require("ethers");
const txUtils = require("../utils/tx-utils");

// Global setup for ethers.js
const provider = new ethers.JsonRpcProvider("https://rpc.thanos-sepolia.tokamak.network") // Add your RPC URL if needed
const wallet = new ethers.Wallet(process.env.PRIVATE_KEY, provider); // Load private key from .env
const contractAddress = "0x7d530f9b81f7e170472f738296d1b581b88fed31"; // Load contract address from .env
const contractABI = [{
  "inputs": [{
      "internalType": "string",
      "name": "babyPubKey",
      "type": "string"
    },
    {
      "internalType": "uint48",
      "name": "fromIdx",
      "type": "uint48"
    },
    {
      "internalType": "uint40",
      "name": "loadAmountF",
      "type": "uint40"
    },
    {
      "internalType": "uint40",
      "name": "amountF",
      "type": "uint40"
    },
    {
      "internalType": "uint48",
      "name": "toIdx",
      "type": "uint48"
    }
  ],
  "name": "addL1Transaction",
  "outputs": [],
  "stateMutability": "payable",
  "type": "function"
}];
const contract = new ethers.Contract(contractAddress, contractABI, wallet); // Initialize contract globally

// Create Account Deposit Transaction
async function l1UserTxCreateAccountDeposit(loadAmount, babyjub) {
  const loadAmountF = loadAmount * 10 ** (18 - 8);

  // Send transaction using ethers.js
  const txRes = await contract.addL1Transaction(
    babyjub,
    0,
    loadAmount,
    0,
    0, {
      value: loadAmountF
    } // Add the value for load amount
  );

  const txReceipt = await txRes.wait(); // Wait for the transaction to be mined
  console.log('Transaction successful:', txReceipt.hash); // Log transaction hash

  return txReceipt.hash;
}

// Force Exit Transaction
async function l1UserTxForceExit(fromIdx, amountF) {
  const exitIdx = 1;

  // equivalent L1 transaction:
  const l1TxForceExit = {
    toIdx: exitIdx,
    amountF: amountF,
    loadAmountF: 0,
    fromIdx: fromIdx,
    fromBjjCompressed: 0,
    fromEthAddr: await wallet.getAddress(),
  };
  const l1Txbytes = `0x${txUtils.encodeL1TxFull(l1TxForceExit)}`;

  const lastQueue = await contract.nextL1FillingQueue();
  const lastQueueBytes = await contract.mapL1TxQueue(lastQueue);
  const currentIndex = (lastQueueBytes.length - 2) / 2 / L1_USER_BYTES;

  // Send transaction using ethers.js
  const txRes = await contract.addL1Transaction(
    '', // babyjub0 as an empty string
    fromIdx,
    0, // loadAmountF0
    amountF,
    exitIdx, {
      value: ethers.utils.parseEther("0")
    } // No ETH value for force exit
  );

  const txReceipt = await txRes.wait(); // Wait for the transaction to be mined
  console.log('Force exit transaction successful:', txReceipt.transactionHash); // Log transaction hash

  return l1Txbytes;
}

module.exports = {
  l1UserTxCreateAccountDeposit,
  l1UserTxForceExit,
};