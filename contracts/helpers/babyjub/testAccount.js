const SybilAccount = require("./sybilAccount/sybil-account");

(async () => {
    const account = new SybilAccount("99cb542c018ed53489ac6aa2afdbfd23a25a32d7d1ef9d1cf9060bd4dc0f5927");
    const accountInfo = await account.initialize();

    console.log("Private Key:", accountInfo.privateKey);
    console.log("Ethereum Address:", accountInfo.ethAddr);
    console.log("BabyJubJub Public Key (ax):", accountInfo.ax);
    console.log("BabyJubJub Public Key (ay):", accountInfo.ay);
    console.log("Compressed BabyJubJub Public Key:", accountInfo.bjjCompressed);
})();