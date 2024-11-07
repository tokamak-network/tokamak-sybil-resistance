require('dotenv').config();
const {
  ethers
} = require("ethers");
const txUtils = require("../utils/tx-utils");
const {
  fix2Float
} = require('../utils/float40');

// Global setup for ethers.js
const provider = new ethers.JsonRpcProvider(process.env.PROVIDER_URL) // Add your RPC URL if needed
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
const contract = new ethers.Contract(process.env.CONTRACT_ADDRESS, contractABI, wallet); // Initialize contract globally

const babyjub0 = '';
const fromIdx0 = 0;
const loadAmountF0 = 0;
const amountF0 = 0;
const tokenID0 = 0;
const toIdx0 = 0;

// Create Account Deposit Transaction
let l1UserTxCreateAccountDeposit = async (loadAmount, babyjub) => {
  const loadAmountF = loadAmount * 10 ** (18 - 8);

  // equivalent L1 transaction:
  const l1TxcreateAccountDeposit = {
    toIdx: 0,
    amountF: 0,
    loadAmountF: loadAmount,
    fromIdx: 0,
    fromBjjCompressed: babyjub,
    fromEthAddr: await wallet.getAddress(),
  };
  const l1Txbytes = `0x${txUtils.encodeL1TxFull(l1TxcreateAccountDeposit)}`;

  // Send transaction using ethers.js
  const txRes = await contract.addL1Transaction(
    babyjub,
    0,
    loadAmount,
    0,
    0, {
      value: loadAmountF
    }
  );

  const txReceipt = await txRes.wait(); // Wait for the transaction to be mined
  console.log('Transaction successful:', txReceipt.hash); // Log transaction hash

  return l1Txbytes;
}

let l1UserTxDepositTransfer = async (
  loadAmount,
  fromIdx,
  toIdx,
  amountF
) => {
  const loadAmountF = loadAmount * 10 ** (18 - 8);;

  // Send transaction using ethers.js
  const txRes = await contract.addL1Transaction(
    babyjub0,
    fromIdx,
    loadAmount,
    amountF,
    toIdx, {
      value: loadAmountF
    }
  );

  const txReceipt = await txRes.wait(); // Wait for the transaction to be mined
  console.log('L1 User Tx Deposit transaction successful:', txReceipt.hash); // Log transaction hash

  return txReceipt.hash;
}

let l1UserTxDeposit = async (
  loadAmount,
  fromIdx,
) => {
  const loadAmountF = loadAmount * 10 ** (18 - 8);;

  // Send transaction using ethers.js
  const txRes = await contract.addL1Transaction(
    babyjub0,
    fromIdx,
    loadAmount,
    amountF0,
    toIdx0, {
      value: loadAmountF
    }
  );

  const txReceipt = await txRes.wait(); // Wait for the transaction to be mined
  console.log('L1 User Tx Deposit transaction successful:', txReceipt.hash); // Log transaction hash

  return txReceipt.hash;
}

async function l1UserTxCreateAccountDepositTransfer(
  loadAmount,
  toIdx,
  amountF,
  babyjub,
) {
  const loadAmountF = loadAmount * 10 ** (18 - 8);;

  // Send transaction using ethers.js
  const txRes = await contract.addL1Transaction(
    babyjub,
    fromIdx0,
    loadAmount,
    amountF,
    toIdx, {
      value: loadAmountF
    }
  );

  const txReceipt = await txRes.wait(); // Wait for the transaction to be mined
  console.log('L1 User Tx Deposit transaction tranfer successful:', txReceipt.hash); // Log transaction hash

  return txReceipt.hash;
}

module.exports = {
  l1UserTxCreateAccountDeposit,
  l1UserTxCreateAccountDepositTransfer,
  l1UserTxDeposit,
  l1UserTxDepositTransfer
};