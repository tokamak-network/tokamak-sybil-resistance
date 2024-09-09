const fs = require("fs");
const { buildBabyjub, buildPoseidon, buildEddsa } = require("circomlibjs");

async function main() {
  const babyJub = await buildBabyjub();
  const poseidon = await buildPoseidon();
  const eddsa = await buildEddsa();
  const F = babyJub.F;
  
  // setup accounts
  const alicePrvKey = Buffer.from("1".toString().padStart(64, "0"), "hex");
  const bobPrvKey = Buffer.from("2".toString().padStart(64, "0"), "hex");
  const alicePubKey = eddsa.prv2pub(alicePrvKey);
  const bobPubKey = eddsa.prv2pub(bobPrvKey);

  const alice = {
    pubkey: alicePubKey,
    balance: 500,
    ethAddr: "0x1111111111111111111111111111111111111111",
    nonce: 0
  };

  const bob = {
    pubkey: bobPubKey,
    balance: 0,
    ethAddr: "0x2222222222222222222222222222222222222222",
    nonce: 0
  };

  // setup accounts and root hash
  const aliceHash = poseidon([alice.pubkey[0], alice.pubkey[1], alice.balance, BigInt(alice.ethAddr), alice.nonce]);
  const bobHash = poseidon([bob.pubkey[0], bob.pubkey[1], bob.balance, BigInt(bob.ethAddr), bob.nonce]);

  const accounts_root = poseidon([aliceHash, bobHash]);

  // transaction
  const tx = {
    from: alice.pubkey,
    to: bob.pubkey,
    amount: 500,
  };

  // sign tx
  const txHash = poseidon([tx.from[0], tx.from[1], tx.to[0], tx.to[1], tx.amount]);

  const signature = eddsa.signPoseidon(alicePrvKey, txHash);

  // new accounts state
  const newAlice = {
    ...alice,
    balance: 0,
    nonce: 1
  };

  const newBob = {
    ...bob,
    balance: 500
  };

  // new accounts and root hash
  const newAliceHash = poseidon([newAlice.pubkey[0], newAlice.pubkey[1], newAlice.balance, BigInt(newAlice.ethAddr), newAlice.nonce]);
  const newBobHash = poseidon([newBob.pubkey[0], newBob.pubkey[1], newBob.balance, BigInt(newBob.ethAddr), newBob.nonce]);
  const intermediate_root = poseidon([newAliceHash, bobHash]);
  const new_accounts_root = poseidon([newAliceHash, newBobHash]);

  const inputs = {
    accounts_root: F.toObject(accounts_root).toString(),
    intermediate_root: F.toObject(intermediate_root).toString(),
    sender_pubkey: [
      F.toObject(alice.pubkey[0]).toString(),
      F.toObject(alice.pubkey[1]).toString(),
    ],
    sender_balance: alice.balance.toString(),
    sender_ethAddr: alice.ethAddr,
    sender_nonce: alice.nonce.toString(),
    receiver_pubkey: [
      F.toObject(bob.pubkey[0]).toString(),
      F.toObject(bob.pubkey[1]).toString(),
    ],
    receiver_balance: bob.balance.toString(),
    receiver_ethAddr: bob.ethAddr,
    receiver_nonce: bob.nonce.toString(),
    amount: tx.amount.toString(),
    signature_R8x: F.toObject(signature.R8[0]).toString(),
    signature_R8y: F.toObject(signature.R8[1]).toString(),
    signature_S: signature.S.toString(),
    sender_proof: [F.toObject(bobHash).toString()],
    sender_proof_pos: ["0"],
    receiver_proof: [F.toObject(newAliceHash).toString()],
    receiver_proof_pos: ["1"],
    enabled: "1"
  };

  console.log("New accounts root:", F.toObject(new_accounts_root).toString());
  fs.writeFileSync("./inputs/singleTx_input.json", JSON.stringify(inputs), 'utf-8');
}

main();