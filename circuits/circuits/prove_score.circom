include "./mimc.circom";



template merkleroot(n){
    signal input path[n];
    signal input path_lr[n];
    signal input leaf;
    signal output root;
    component tmp_root[n];
    tmp_root[0] = MultiMiMC7(2,91);
    tmp_root[0].in[0] <== leaf - path_lr[0]* (leaf - path[0]);
    tmp_root[0].in[1] <== path[0] - path_lr[0]* (path[0] - leaf);
    tmp_root[0].k <== 0;
    for (var i = 1; i < n; i++){
        tmp_root[i] = MultiMiMC7(2,91);
        tmp_root[i].in[0] <== path[i] - path_lr[i]* (path[i] - tmp_root[i-1].out);
        tmp_root[i].in[1] <== tmp_root[i-1].out - path_lr[i]*(tmp_root[i-1].out - path[i]);
        tmp_root[i].k <== 0;

    }

    root <== tmp_root[n-1].out;

}


component main = merkleroot(2);
