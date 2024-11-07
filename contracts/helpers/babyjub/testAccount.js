const SybilAccount = require("./sybilAccount/sybil-account");

(async () => {
    const account = new SybilAccount();
    const accountInfo = await account.initialize();

    console.log(accountInfo);
    console.log("Private Key:", accountInfo.privateKey);
    console.log("Ethereum Address:", accountInfo.ethAddr);
    console.log("BabyJubJub Public Key (ax):", accountInfo.ax);
    console.log("BabyJubJub Public Key (ay):", accountInfo.ay);
    console.log("Compressed BabyJubJub Public Key:", accountInfo.bjjCompressed);
})();