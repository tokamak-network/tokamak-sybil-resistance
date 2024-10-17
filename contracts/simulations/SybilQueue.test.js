async function SybilQueueSimulations() {
    console.log("-------------- SETUP BEGIN ---------------")
    const {
        PRIVATE_KEY, 
        RPC_ENDPOINT,
        SYBIL_CONTRACT_ADDRESS
    } = require("../helpers/constants");

    const { expect } = require("chai");
    const SybilAccount = require("../helpers/babyjub/sybilAccount/sybil-account");
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

    const account = new SybilAccount(ownerAccount.address.toString());
    const accountInfo = await account.initialize();

    const babyjubjub = accountInfo.bjjCompressed;   
    const fromIdx0 = 0;
    const loadAmountF0 = 0;
    const amountF0 = 0;
    const toIdx0 = 0;

    console.log("-------------- SETUP COMPLETE ---------------")

    console.log("-------------- TEST EXCEED 128 L1-USER-TX BEGIN ---------------")
    // ---------- Test exceed 128 l1-user-tx ----------- //
    const initialLastForge = await sybil.nextL1FillingQueue();
    const initialCurrentForge = await sybil.nextL1ToForgeQueue();

    // add l1-user-tx
    for (let i = 0; i < 127; i++) {
        let tx = await sybil.addL1Transaction(
            babyjubjub,
            fromIdx0,
            loadAmountF0,
            amountF0,
            toIdx0,
            {value: ethers.parseEther(loadAmountF0.toString())}
        );
    
        tx.wait()
    }

    // after 127 l1-user-tx still in the same queue
    const intermidiateLastForge = await sybil.nextL1FillingQueue();
    expect(initialLastForge).to.equal(
      intermidiateLastForge
    );

    const intermidiateCurrentForge = await sybil.nextL1ToForgeQueue();
    expect(initialCurrentForge).to.equal(
      intermidiateCurrentForge
    );

    // exceed max tx in queue
    await sybil.addL1Transaction(
        babyjubjub,
        fromIdx0,
        loadAmountF0,
        amountF0,
        toIdx0
    )

    // last Forge is updated at transaction 114
    const afterL1LastForge = await sybil.nextL1FillingQueue();
    const afterL1CurrentForge = await sybil.nextL1ToForgeQueue();
    expect(parseInt(initialLastForge) + 1).to.equal(afterL1LastForge);
    expect(parseInt(initialCurrentForge)).to.equal(afterL1CurrentForge);

    console.log("-------------- TEST EXCEED 128 L1-USER-TX COMPLETE ---------------")
}

try {
    SybilQueueSimulations();
} catch (error) {
    console.error("Failed to run Sybil Queue simulation with error: ", error);
}
