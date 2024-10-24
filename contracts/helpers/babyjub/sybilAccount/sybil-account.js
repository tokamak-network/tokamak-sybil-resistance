const EC = require("elliptic").ec;
const ec = new EC("secp256k1");
const keccak256 = require("js-sha3").keccak256;
const crypto = require("crypto");
const Scalar = require("ffjavascript").Scalar;
const utilsScalar = require("ffjavascript").utils;
const circomlibjs = require("circomlibjs");

const txUtils = require("../../utils/tx-utils");
const utils = require("../../utils/utils");

module.exports = class SybilAccount {
    constructor(privateKey) {
        if (privateKey) {
            if (typeof privateKey !== "string") {
                this.privateKey = Scalar.e(privateKey).toString(16);
            } else {
                this.privateKey = privateKey;
            }
            while (this.privateKey.length < 64) this.privateKey = "0" + this.privateKey;
        } else {
            this.privateKey = crypto.randomBytes(32).toString("hex");
        }

        // Get secp256k1 generator point
        const generatorPoint = ec.g;

        // Public Key Coordinates calculated via Elliptic Curve Multiplication
        const pubKeyCoordinates = generatorPoint.mul(this.privateKey);

        const x = pubKeyCoordinates.getX().toString("hex");
        const y = pubKeyCoordinates.getY().toString("hex");

        // Public Key = X and Y concatenated
        const publicKey = x + y;

        // Use Keccak-256 hash function to get public key hash
        const hashOfPublicKey = keccak256(Buffer.from(publicKey, "hex"));

        // Convert hash to buffer
        const ethAddressBuffer = Buffer.from(hashOfPublicKey, "hex");

        // Ethereum Address is '0x' concatenated with the last 20 bytes
        const ethAddress = ethAddressBuffer.slice(-20).toString("hex");
        this.ethAddr = `0x${ethAddress}`;
    }

    // Asynchronous initialization with circomlib components
    async initialize() {
        try {
            // Load circomlibjs asynchronously
            const circomlib = await circomlibjs;

            // Build BabyJubJub and Eddsa
            const babyJub = await circomlib.buildBabyjub();
            const eddsa = await circomlib.buildEddsa();

            // Save eddsa instance to this for later use
            this.eddsa = eddsa;

            // Derive a private key with a hash
            this.rollupPrvKey = Buffer.from(keccak256("SYBIL_MOCK_ACCOUNT" + this.privateKey), "hex");

            const bjPubKey = eddsa.prv2pub(this.rollupPrvKey);

            // Convert the BabyJubJub public key ax and ay to string representations of comma-separated numbers
            this.ax = bjPubKey[0].toString(16);
            this.ay = bjPubKey[1].toString(16);

            // Compress the BabyJubJub public key
            const compressedBuff = babyJub.packPoint(bjPubKey);

            this.sign = 0;
            if (compressedBuff[31] & 0x80) {
                this.sign = 1;
            }

            this.bjjCompressed = utils.padZeros(utilsScalar.leBuff2int(compressedBuff).toString(16), 64);

            // Return the computed values along with the private key
            return {
                privateKey: this.privateKey, // Include private key here
                ethAddr: this.ethAddr,
                ax: this.ax,  // String format of ax
                ay: this.ay,  // String format of ay
                bjjCompressed: this.bjjCompressed,
            };
        } catch (error) {
            console.error("Error during account initialization:", error);
            throw error;
        }
    }

    /**
     * Sign rollup transaction 
     * Adds signature and sender data to the transaction
     * @param {Object} tx - Transaction object
     */
    signTx(tx) {
        const h = txUtils.buildHashSig(tx);

        const signature = this.eddsa.signPoseidon(this.rollupPrvKey, h);
        tx.r8x = signature.R8[0];
        tx.r8y = signature.R8[1];
        tx.s = signature.S;
        tx.fromAx = this.ax;
        tx.fromAy = this.ay;
    }
};
