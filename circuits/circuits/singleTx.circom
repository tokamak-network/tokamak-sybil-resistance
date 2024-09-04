pragma circom 2.0.0;
include "./check_leaf_existence.circom";
include "./get_merkle_root.circom";
include "../node_modules/circomlib/circuits/poseidon.circom";
include "../node_modules/circomlib/circuits/eddsaposeidon.circom";

template singleTx(k){
    // k is the depth of accounts tree

    // accounts tree info
    signal input accounts_root;
    signal input intermediate_root;
    //@TODO: signal input links_root;

    // account State: sender
    signal input sender_pubkey[2];
    signal input sender_balance;
    signal input sender_ethAddr;
    signal input sender_nonce;
    signal input sender_score; //@TODO: consider removing score in account state

    // account State: receiver
    signal input receiver_pubkey[2];
    signal input receiver_balance;
    signal input receiver_ethAddr;
    signal input receiver_nonce;
    signal input receiver_score;

    //tx
    signal input amount;
    signal input signature_R8x;
    signal input signature_R8y;
    signal input signature_S;
    signal input sender_proof[k];
    signal input sender_proof_pos[k];
    signal input receiver_proof[k];
    signal input receiver_proof_pos[k];
    signal input enabled;
    
    signal output new_accounts_root;
    //@TODO: signal output new_links_root;

    // verify sender account exists in accounts_root
    component senderExistence = LeafExistence(k,6);
    senderExistence.preimage[0] <== sender_pubkey[0];
    senderExistence.preimage[1] <== sender_pubkey[1];
    senderExistence.preimage[2] <== sender_balance;
    senderExistence.preimage[3] <== sender_ethAddr;
    senderExistence.preimage[4] <== sender_nonce;
    senderExistence.preimage[5] <== sender_score;
    senderExistence.root <== accounts_root;
    for(var i = 0; i < k; i++){
        senderExistence.paths2_root_pos[i] <== sender_proof_pos[i];
        senderExistence.paths2_root[i] <== sender_proof[i];
    }

    // hash msg
    component msg = Poseidon(5);
    msg.inputs[0] <== sender_pubkey[0];
    msg.inputs[1] <== sender_pubkey[1];
    msg.inputs[2] <== receiver_pubkey[0];
    msg.inputs[3] <== receiver_pubkey[1];
    msg.inputs[4] <== amount;

    // check that transaction was signed by sender
    component signatureCheck = EdDSAPoseidonVerifier();
    signatureCheck.enabled <== enabled;
    signatureCheck.Ax <== sender_pubkey[0];
    signatureCheck.Ay <== sender_pubkey[1];
    signatureCheck.R8x <== signature_R8x;
    signatureCheck.R8y <== signature_R8y;
    signatureCheck.S <== signature_S;
    signatureCheck.M <== msg.out;

    // debit sender account and hash new sender leaf
    component newSenderLeaf = Poseidon(6);
    newSenderLeaf.inputs[0] <== sender_pubkey[0];
    newSenderLeaf.inputs[1] <== sender_pubkey[1];
    newSenderLeaf.inputs[2] <== (sender_balance - amount);
    newSenderLeaf.inputs[3] <== sender_ethAddr;
    newSenderLeaf.inputs[4] <== sender_nonce + 1;
    newSenderLeaf.inputs[5] <== sender_score;
    
    component compute_intermediate_root = GetMerkleRoot(k);
    compute_intermediate_root.leaf <== newSenderLeaf.out;
    for(var i = 0; i < k; i++){
        compute_intermediate_root.paths2_root_pos[i] <== sender_proof_pos[i];
        compute_intermediate_root.paths2_root[i] <== sender_proof[i];
    }

    // check that computed_intermediate_root.out === intermediate_root
    compute_intermediate_root.out === intermediate_root;

    // verify receiver account exists in intermediate_root
    component receiverExistence = LeafExistence(k,6);
    receiverExistence.preimage[0] <== receiver_pubkey[0];
    receiverExistence.preimage[1] <== receiver_pubkey[1];
    receiverExistence.preimage[2] <== receiver_balance;
    receiverExistence.preimage[3] <== receiver_ethAddr;
    receiverExistence.preimage[4] <== receiver_nonce;
    receiverExistence.preimage[5] <== receiver_score;
    receiverExistence.root <== intermediate_root;
    for(var i = 0; i < k; i++){
        receiverExistence.paths2_root_pos[i] <== receiver_proof_pos[i];
        receiverExistence.paths2_root[i] <== receiver_proof[i];
    }

    // credit receiver account and hash new receiver leaf
    component newReceiverLeaf = Poseidon(6);
    newReceiverLeaf.inputs[0] <== receiver_pubkey[0];
    newReceiverLeaf.inputs[1] <== receiver_pubkey[1];
    newReceiverLeaf.inputs[2] <== (receiver_balance + amount);
    newReceiverLeaf.inputs[3] <== receiver_ethAddr;
    newReceiverLeaf.inputs[4] <== receiver_nonce;
    newReceiverLeaf.inputs[5] <== receiver_score;

    // update accounts_root
    component compute_final_root = GetMerkleRoot(k);
    compute_final_root.leaf <== newReceiverLeaf.out;
    for(var i = 0; i < k; i++){
        compute_final_root.paths2_root_pos[i] <== receiver_proof_pos[i];
        compute_final_root.paths2_root[i] <== receiver_proof[i];
    }

    //output final accounts_root
    new_accounts_root <== compute_final_root.out;
}
component main = LinkTx(1);