pragma circom 2.0.0;

/**
* all inputs are scaled up 10^6 since circom only accepts integers as inputs 
* @param num_verts - number of vertices in part of graph where scoring algorithm will be run
* @param num_subsets - number of subsets of vertices used in calculating score
* @input weights[num_verts][num_verts] - {Array(Uint180)} - scaled integer representing stake on a link
* @input subsets[num_verts][num_subsets] - {Bool} - Boolean of whether a particular vertex is an element of a particular subset 

**/
template Num2Bits(n) {
    signal input in;
    signal output out[n];
    var lc1=0;

    var e2=1;
    for (var i = 0; i<n; i++) {
        out[i] <-- (in >> i) & 1;
        out[i] * (out[i] -1 ) === 0;
        lc1 += out[i] * e2;
        e2 = e2+e2;
    }
    lc1 === in;
}
template LessThan(n) {
    assert(n <= 252);
    signal input in[2];
    signal output out;



    component n2b = Num2Bits(n+1);

    n2b.in <== in[0]+ (1<<n) - in[1];

    out <== 1-n2b.out[n];
}

template NewScoringAlgorithm (num_verts, p, q, num_steps) {
	signal input weights[num_verts][num_verts];
	signal output noderanks[num_verts][num_verts];

   
    signal residual[num_verts][num_steps][num_verts];
    signal rank[num_verts][num_steps][num_verts];




    signal deg[num_verts];


    signal sort[num_verts][num_verts-1][num_verts];



   
    component comp[num_verts][num_steps][num_verts];


    var sum = 0;
    for(var k = 0; k<num_verts; k+=1){
        sum = 0;
        for(var j = 0; j<num_verts; j+=1){
            sum = sum + weights[k][j];
        }   
        deg[k] <== sum;
        log(deg[k]);
    }


    for(var k = 0; k<num_verts; k+=1){

        for(var v = 0; v<k; v+=1){
            rank[k][0][v] <== 0;
            residual[k][0][v] <== 0;
        }

        residual[k][0][k] <== 10;
        rank[k][0][k] <== 0;

        for(var v = k+1; v<num_verts; v+=1){
            rank[k][0][v] <== 0;
            residual[k][0][v] <== 0;
        }

        for(var step = 0; step<num_steps-1; step+=1){
            for(var j = 0; j<num_verts; j+=1){
                comp[k][step][j] = LessThan(7);
                comp[k][step][j].in[1] <== residual[k][step][j];
                comp[k][step][j].in[0] <== q * deg[j];

                log(1000000000000000);
                log(step);
                log(j);

                log(residual[k][step][j]);
                log(q * deg[j]);

                log(p * comp[k][step][j].out * residual[k][step][j] \ 10);
                log((10+p)*residual[k][step][j]*comp[k][step][j].out \ 20);

                rank[k][step+1][j] <== rank[k][step][j] + p * comp[k][step][j].out * residual[k][step][j] \ 10;
                residual[k][step+1][j] <== residual[k][step][j] - (10+p)*residual[k][step][j]*comp[k][step][j].out \ 20;
            }

        }

        for(var i = 0; i<num_verts; i+=1){
            noderanks[k][i] <== rank[k][num_steps-1][i];
        }

            
    } 

}

component main = NewScoringAlgorithm(5,2,1,2);






