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

template NewScoringAlgorithm (num_verts , num_subsets, p, epsilon, num_steps) {
	signal input subsets[num_verts][num_subsets];
	signal input weights[num_verts][num_verts];
	signal output scores[num_verts];

    signal bdry[num_subsets];
    signal bdry_checks[num_subsets];
	signal scaled_bdry[num_subsets];
	signal subset_indicator[num_subsets][num_verts][num_verts];
	signal weighted_subset_indicator[num_subsets][num_verts][num_verts];
	signal selector[num_verts][num_subsets];
    signal minimizing_vector[num_verts][num_subsets];

    signal mass[num_verts][num_verts][num_steps];
    signal rank[num_verts][num_verts][num_steps];
    signal queue[num_verts][num_verts][num_steps];
    signal path[num_verts][num_verts];
    signal minimizing_sweep[num_verts][num_verts];

    signal subset_indicator[num_subsets][num_verts][num_verts];

    component lt[num_verts][num_verts];
    component lt2[num_verts][num_verts];


    for(var k = 0; k<num_verts; k+=1){
        for(var v = 0; v<num_verts; v+=1){
            rank[k][v][0] <== 0;
            mass[k][v][0] <== 0;
        }

        mass[k][k][0] <== 1;
        queue[k][k][0] <== 1;

        for(var step = 0; step<num_steps; step+=1){

            rank[k][k][step+1] <== rank[k][k][step] + p*mass[k][k][step];
            mass[k][k][step+1] <== 0.5*(1-p)*mass[k][k][step];

            for(var j = 0; j<num_verts; j+=1){
                mass[k][j][step+1] <== mass[k][j][step] + 0.5*(1-p)*weights[k][j]
                for(var i = 0; i<num_verts; i+=1){
          
                   lt[k][i] = LessThan(5);
                   lt[k][i].in[1] <== mass[k][j][step+1];
                   lt[k][i].in[0] <== epsilon*weights[k][i]; 

                   queue[k][j][step+1] <== lt[k][i].out

                }

            }
            
        } 

        for(var b = 0; b<num_verts; b+=1){
            lt2[k][b] = LessThan(5);
            lt2[k][b].in[1] <== rank[k][b][num_step-1];
            lt2[k][b].in[0] <== minimizing_sweep[k][b]; 

            path[k][b] <== lt2[k][i].out

        }

        //convert path to subset indicator

        sum = 0;
        size = 0;

        for (var c = 0; c<num_verts; c+=1){

            for (var d = 0; d<num_verts; d+=1){

              subset_indicator[a][c][d] <== subsets[c][a]*(1-subsets[d][a]);
              weighted_subset_indicator[a][c][d] <== subset_indicator[a][c][d]*weights[c][d];
              sum = sum + weighted_subset_indicator[a][c][d];
            
            }

            size = size + subsets[c][a];
        }
        bdry[a] <== sum;
        scaled_bdry[a] <-- bdry[a]\size;
        rem = sum - (size*scaled_bdry[a]);
        bdry_checks[a] <== sum - rem;
        scaled_bdry[a]*size === bdry_checks[a];

    }

}

component main = ScoringAlgorithm(7,63);






