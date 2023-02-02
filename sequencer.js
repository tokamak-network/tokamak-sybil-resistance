const { MerkleTree } = require('merkletreejs')
const SHA256 = require('crypto-js/sha256')
const fs = require('fs');
let tree_data = require('/Users/mac/Desktop/zksnarks/scoring_circuit_js/input.json');
let link_data = tree_data['weights']
let scores = require('/Users/mac/Desktop/zksnarks/public.json');
module.exports.view_scores = function () {
      console.log(tree_data);
    };

module.exports.create_link_tree = function () {
	link_tree_roots = []
	for(var i = 0; i < link_data.length; i++){
		let leaves = link_data[i].map(x => SHA256(x))
		let link_tree = new MerkleTree(leaves, SHA256)
		let root = link_tree.getRoot().toString('hex')
		link_tree_roots.push(root)
	}
    let main_tree_leaves = link_tree_roots.map(x => SHA256(x))
    let main_tree = new MerkleTree(main_tree_leaves, SHA256)
    let main_root = main_tree.getRoot().toString('hex')
    console.log('link_tree root hash:',main_root)
};

module.exports.create_scores_tree = function () {
    let scores_tree_leaves = scores.map(x => SHA256(x))
    let scores_tree = new MerkleTree(scores_tree_leaves, SHA256)
    let scores_root = scores_tree.getRoot().toString('hex')
    console.log('scores_tree root hash:',scores_root)
};

module.exports.update_link_weight = function (i,j,w) {
	  let content = tree_data
	  content['weights'][i][j] = String(w)
	  content['weights'][j][i] = String(w)
      fs.writeFileSync('./scoring_circuit_js/input.json', JSON.stringify(content));
    };




