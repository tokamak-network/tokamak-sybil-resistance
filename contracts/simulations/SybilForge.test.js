async function SybilForgeSimulations() {
    console.log("-------------- SETUP BEGIN ---------------");
    const {
        PRIVATE_KEY, 
        RPC_ENDPOINT,
        SYBIL_CONTRACT_ADDRESS
    } = require("../helpers/constants");
    const ethers = require("ethers");

    // set up test wallet
    const owner = new ethers.Wallet(PRIVATE_KEY);
    const provider = new ethers.JsonRpcProvider(RPC_ENDPOINT);
    const ownerAccount = owner.connect(provider);

    // get sybil contract
    const sybilOutput = require("../out/Sybil.sol/Sybil.json");
    const sybil = new ethers.Contract(
        SYBIL_CONTRACT_ADDRESS,
        sybilOutput.abi,
        ownerAccount
    );
    console.log("-------------- SETUP COMPLETE ---------------");
    
    console.log("-------------- FORGE EMPTY BATCH BEGIN ---------------");
    // forge batch
    await sybil.forgeBatch(
        newLastIdx = 256,
        newStRoot = 256,
        newVouchRoot = 256,
        newScoreRoot = 256,
        newExitRoot = 256,
        verifierIdx = 0,
        l1Batch = true,
        proofA = ["0", "0"],
        proofB = [
          ["0", "0"],
          ["0", "0"],
        ],
        proofC = ["0", "0"],
        input = 0
    );
    console.log("-------------- FORGE EMPTY BATCH COMPLETE ---------------");
}

try {
    SybilForgeSimulations();
} catch (error) {
    console.error("Failed to run Sybil Forge simulation with error: ", error);
}
