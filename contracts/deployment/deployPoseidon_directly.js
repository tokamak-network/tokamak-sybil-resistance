const { ethers } = require("ethers");
const poseidonGenContract = require("circomlibjs").poseidonContract;
require('dotenv').config();

async function main() {
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
  const deployer = wallet;

  // Poseidon with 2 elements
  const Poseidon2Elements = new ethers.ContractFactory(
    poseidonGenContract.generateABI(2),
    poseidonGenContract.createCode(2),
    deployer
  );
  const poseidon2Contract = await Poseidon2Elements.deploy();
  await poseidon2Contract.waitForDeployment();
  console.log("Poseidon2Elements deployed at:", await poseidon2Contract.getAddress());

  // Poseidon with 3 elements
  const Poseidon3Elements = new ethers.ContractFactory(
    poseidonGenContract.generateABI(3),
    poseidonGenContract.createCode(3),
    deployer
  );
  const poseidon3Contract = await Poseidon3Elements.deploy();
  await poseidon3Contract.waitForDeployment();
  console.log("Poseidon3Elements deployed at:", await poseidon3Contract.getAddress());

  // Poseidon with 4 elements
  const Poseidon4Elements = new ethers.ContractFactory(
    poseidonGenContract.generateABI(4),
    poseidonGenContract.createCode(4),
    deployer
  );
  const poseidon4Contract = await Poseidon4Elements.deploy();
  await poseidon4Contract.waitForDeployment();
  console.log("Poseidon4Elements deployed at:", await poseidon4Contract.getAddress());
}

main().catch((error) => {
  console.error("Error deploying contracts:", error);
});
