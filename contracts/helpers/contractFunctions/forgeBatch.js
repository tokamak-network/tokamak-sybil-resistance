const { ethers } = require("ethers");
require("dotenv").config();

const contractABI = [
  {
    "inputs": [
      { "internalType": "uint48", "name": "newLastIdx", "type": "uint48" },
      { "internalType": "uint256", "name": "newStRoot", "type": "uint256" },
      { "internalType": "uint256", "name": "newVouchRoot", "type": "uint256" },
      { "internalType": "uint256", "name": "newScoreRoot", "type": "uint256" },
      { "internalType": "uint256", "name": "newExitRoot", "type": "uint256" },
      { "internalType": "uint8", "name": "verifierIdx", "type": "uint8" },
      { "internalType": "bool", "name": "l1Batch", "type": "bool" },
      { "internalType": "uint256[2]", "name": "proofA", "type": "uint256[2]" },
      { "internalType": "uint256[2][2]", "name": "proofB", "type": "uint256[2][2]" },
      { "internalType": "uint256[2]", "name": "proofC", "type": "uint256[2]" },
      { "internalType": "uint256", "name": "input", "type": "uint256" }
    ],
    "name": "forgeBatch",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  }
];

class ForgerTest {
  constructor(maxTx, maxL1Tx, nLevels, rollupDB) {
    this.maxTx = maxTx;
    this.maxL1Tx = maxL1Tx;
    this.nLevels = nLevels;
    this.rollupDB = rollupDB;
    this.L1TxB = 544;

    this.provider = new ethers.JsonRpcProvider(process.env.PROVIDER_URL);
    const wallet = new ethers.Wallet(process.env.PRIVATE_KEY, this.provider); 
    this.contract = new ethers.Contract(process.env.CONTRACT_ADDRESS, contractABI, wallet);
  }

  async forgeBatch(l1Batch, l1TxUserArray, l1TxCoordiatorArray, l2txArray) {
    // 1. Build batch with rollupDB
    const bb = await this.rollupDB.buildBatch(this.maxTx, this.nLevels, this.maxL1Tx);
    
    console.log("tx:", l1TxUserArray);

    // Add L1 and L2 transactions to the batch builder
    l1TxUserArray.forEach(tx => bb.addTx(tx));
    l1TxCoordiatorArray.forEach(tx => bb.addTx(tx));
    // l2txArray.forEach(tx => bb.addTx(tx));
    if (l2txArray) {
      for (let tx of l2txArray) {
        bb.addTx(tx);
      }
    }

    await bb.build();

    console.log("bb:", bb);

    // 2. Retrieve data from rollupDB batch builder
    const newLastIdx = bb.getNewLastIdx();
    const newStRoot = bb.getNewStateRoot();
    const newVouchRoot = bb.getNewVouchRoot();
    const newScoreRoot = bb.getNewScoreRoot();
    const newExitRoot = bb.getNewExitRoot();
    const proofA = ["0", "0"];
    const proofB = [["0", "0"], ["0", "0"]];
    const proofC = ["0", "0"];
    const input = bb.getHashInputs();
    const verifierIdx = 0;
    // const l1Batch = true;

    // 3. Send forgeBatch transaction
    try {
      const tx = await this.contract.forgeBatch(
        newLastIdx,
        newStRoot,
        newVouchRoot,
        newScoreRoot,
        newExitRoot,
        verifierIdx,
        l1Batch,
        proofA,
        proofB,
        proofC,
        input
      );

      console.log("forgeBatch transaction submitted: ", tx.hash);

      await tx.wait();
      console.log("forgeBatch transaction confirmed");


      // 4. Consolidate batch in rollupDB
      // await this.rollupDB.consolidate(await bb);
      // console.log("Batch consolidated in rollupDB");
    } catch (error) {
      console.error("Error in forgeBatch: ", error);
    }
  }
}

module.exports = {
  ForgerTest
}