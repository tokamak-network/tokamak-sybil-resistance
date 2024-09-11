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

template ScoringAlgorithm (num_verts , num_subsets) {
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

	component lt[num_verts][num_subsets];

    var sum = 0;
    var size = 0;
    var rem = 0;

	for (var a = 0; a<num_subsets; a+=1){  

	    sum = 0;
	    size = 0;

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


    for(var k = 0; k<num_verts; k+=1){
        lt[k][0] = LessThan(5);
        lt[k][0].in[1] <== 31;
        lt[k][0].in[0] <== scaled_bdry[0];
        selector[k][0] <== lt[k][0].out * subsets[k][0];
        minimizing_vector[k][0] <== (scaled_bdry[0]-31)*selector[k][0] + 31;



        for(var b = 1; b<num_subsets; b+=1){
           lt[k][b] = LessThan(5);
           lt[k][b].in[1] <== minimizing_vector[k][b-1];
           lt[k][b].in[0] <== scaled_bdry[b];
           selector[k][b] <== lt[k][b].out * subsets[k][b];
           minimizing_vector[k][b] <== (scaled_bdry[b]-minimizing_vector[k][b-1])*selector[k][b] + minimizing_vector[k][b-1];
        }

        scores[k] <== minimizing_vector[k][num_subsets-1];
        log(scores[k]);
    }
    

}
//component main = ScoringAlgorithm(7,63);
