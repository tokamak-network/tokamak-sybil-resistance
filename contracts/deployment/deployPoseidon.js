// deploy-poseidon.js
const { ethers } = require("ethers");
const poseidonGenContract = require("circomlibjs").poseidonContract;
require('dotenv').config();

async function deployPoseidon(elements) {
  const privateKey = process.env.PRIVATE_KEY;
  const providerUrl = process.env.PROVIDER_URL;

  if (!privateKey) {
    throw new Error("Private key not set in environment variables");
  }

  if (!providerUrl) {
    throw new Error("Provider URL not set in environment variables");
  }

  const provider = new ethers.JsonRpcProvider(providerUrl);
  const wallet = new ethers.Wallet(privateKey, provider);

  // Generate Poseidon contract
  const PoseidonFactory = new ethers.ContractFactory(
    poseidonGenContract.generateABI(elements),
    poseidonGenContract.createCode(elements),
    wallet
  );

  const poseidonContract = await PoseidonFactory.deploy();
  
  // Wait for deployment confirmation
  await poseidonContract.waitForDeployment(); 

  let contractAddress = await poseidonContract.getAddress();

  console.log(contractAddress);

  return contractAddress;
}

// Get the number of elements from the command line arguments
const elements = parseInt(process.argv[2], 10) || 2;

deployPoseidon(elements).catch((error) => {
  console.error("Error deploying Poseidon contract:", error);
});
