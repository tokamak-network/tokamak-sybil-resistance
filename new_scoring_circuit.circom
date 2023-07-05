pragma circom 2.0.0;

/**
* all inputs are scaled up 10^6 since circom only accepts integers as inputs 
* @param num_verts - number of vertices in part of graph where scoring algorithm will be run
* @param num_steps - number of steps used in calculating score
* @param p - residual parameter
* @param q - cutoff parameter
* @input weights[num_verts][num_verts] - {Array(Uint180)} - scaled integer representing stake on a link

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

template NewScoringAlgorithm (num_verts , p, q, num_steps) {
	signal input weights[num_verts][num_verts];
	signal output scores[num_verts];

   
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
    }


    for(var k = 0; k<num_verts; k+=1){

        residual[k][0][k] <== 1000;
        rank[k][0][k] <== 0;
        for(var v = 0; v<k; v+=1){
            rank[k][0][v] <== 0;
            residual[k][0][v] <== 0;
        }


        for(var v = k+1; v<num_verts; v+=1){

            rank[k][0][v] <== 0;
            residual[k][0][v] <== 0;
        }

        for(var step = 0; step<num_steps-1; step+=1){
            for(var j = 0; j<num_verts; j+=1){
                comp[k][step][j] = LessThan(5);
                comp[k][step][j].in[0] <== q * deg[j];
                comp[k][step][j].in[1] <== residual[k][step][j];

                rank[k][step+1][j] <== rank[k][step][j] + p * comp[k][step][j].out * residual[k][step][j];
                residual[k][step+1][j] <== residual[k][step][j] - (1+p)*residual[k][step][j]*comp[k][step][j].out \ 2;
            }

        }

    for (var i = 0; i<num_verts; i+=1){

            for (var j = 0; j<num_verts; j+=1){

              subset_indicator[a][i][j] <== subsets[i][a]*(1-subsets[j][a]);
              weighted_subset_indicator[a][i][j] <== subset_indicator[a][i][j]*weights[i][j];
              sum = sum + weighted_subset_indicator[a][i][j];
            
            }

            size = size + subsets[i][a];
        }


        bdry[a] <== sum;
        scaled_bdry[a] <-- bdry[a]\size;
        rem = sum - (size*scaled_bdry[a]);
        bdry_checks[a] <== sum - rem;
        scaled_bdry[a]*size === bdry_checks[a];
            
    } 

}

component main = NewScoringAlgorithm(12,1,1,10);






