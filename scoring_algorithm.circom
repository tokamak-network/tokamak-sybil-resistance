pragma circom 2.0.0;

/**
* all inputs are scaled up 10^6 since circom only accepts integers as inputs
* @param num_verts - number of vertices in part of graph where scoring algorithm will be run
* @param num_subsets - number of subsets of vertices used in calculating score
* @input weights[num_verts][num_verts] - {Array(Uint180)} - scaled integer representing stake on a link
* @input subsets[num_verts][num_subsets] - {Bool} - Boolean of whether a particular vertex is an element of a particular subset 

**/

template ScoringAlgorithm (num_verts , num_subsets) {
	signal input subsets[num_verts][num_subsets];
	signal input weights[num_verts][num_verts];
	signal output scores[num_verts];
	signal bdry[num_subsets];

    var sum = 0;
    var size = 0;
	for (var a = 0; a<num_subsets; a+=1){   
	    sum = 0;
	    size = 0;
	    for (var i = 0; i<num_verts; i+=1){
		    for (var j = 0; j<num_verts; j+=1){

		      sum = sum + subsets[i][a]*(1-subsets[j][a])*weights[i][j];
	        
	        }
	        size = size + subsets[i][a];
        }
        bdry[a] <== sum \ size;
	}


    for(var k = 0; k<num_verts; k+=1){
        var min = 10**9;
        for(var b = 0; b<num_subsets; b+=1){

            min = min>bdry[b]? bdry[b] : min;
            
        }
        scores[k] <== min;
    }


}

component main = ScoringAlgorithm();
