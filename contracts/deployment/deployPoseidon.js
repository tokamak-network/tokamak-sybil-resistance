// deploy-poseidon.js
const { ethers } = require("ethers");
const poseidonGenContract = require("circomlibjs").poseidonContract;
const fs = require('fs'); // Import the fs module to handle file system operations
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

  const contractAddress = await poseidonContract.getAddress();

  console.log(contractAddress);

  // Ensure the output directory exists
  const outputDir = './broadcast/DeployPoseidon.s.sol';
  if (!fs.existsSync(outputDir)){
      fs.mkdirSync(outputDir, { recursive: true }); // Create the directory if it doesn't exist
  }

  // Prepare data to write
  const jsonData = {
    [`${elements}_elements`]: contractAddress,
  };

  const filePath = `${outputDir}/deployments.json`;

  // Read existing data if the file exists
  let existingData = {};
  if (fs.existsSync(filePath)) {
    const fileContent = fs.readFileSync(filePath);
    existingData = JSON.parse(fileContent);
  }

  // Update existing data with new data, replacing the entry for the same key
  existingData = { ...existingData, ...jsonData };

  // Write the updated data back to the file
  fs.writeFileSync(filePath, JSON.stringify(existingData, null, 2));

  return contractAddress;
}

// Get the number of elements from the command line arguments
const elements = parseInt(process.argv[2], 10) || 2;

deployPoseidon(elements).catch((error) => {
  console.error("Error deploying Poseidon contract:", error);
});
