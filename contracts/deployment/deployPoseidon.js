// deploy-poseidon.js
const { ethers } = require("ethers");
const poseidonGenContract = require("circomlibjs").poseidonContract;
require('dotenv').config();

async function deployPoseidon(elements) {
  const privateKey = process.env.PRIVATE_KEY;
  if (!privateKey) {
    throw new Error("Private key not set in environment variables");
  }

  const provider = new ethers.JsonRpcProvider("https://rpc.titan-sepolia.tokamak.network");
  const wallet = new ethers.Wallet(privateKey, provider); // Use an environment variable for the private key

  // Generate Poseidon contract
  const PoseidonFactory = new ethers.ContractFactory(
    poseidonGenContract.generateABI(elements),
    poseidonGenContract.createCode(elements),
    wallet
  );

  const poseidonContract = await PoseidonFactory.deploy();
  
  // Wait for deployment confirmation
  await poseidonContract.deployTransaction.wait(); // Ensure the deployment transaction is mined

  // Log contract address
  console.log("Poseidon Contract deployed at:", poseidonContract.address); // Use 'address' instead of 'getAddress()'
}

// Get the number of elements from the command line arguments
const elements = parseInt(process.argv[2], 10) || 2;

deployPoseidon(elements).catch((error) => {
  console.error("Error deploying Poseidon contract:", error);
});
