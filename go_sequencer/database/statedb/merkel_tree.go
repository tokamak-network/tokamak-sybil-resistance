package statedb

import (
	"encoding/json"
	"fmt"
)

<<<<<<< HEAD
// buildMerkleTree constructs a Merkle tree from a list of leaf nodes.
func buildMerkleTree(leaves []*TreeNode) *TreeNode {
	if len(leaves) == 1 {
		return leaves[0]
=======
type enum string

const (
	Account enum = "Account"
	Link    enum = "Link"
)

type TreeNodeHash interface {
}

// BigInt returns a *big.Int representing the Idx
func BigInt(idx int) *big.Int {
	return big.NewInt(int64(idx))
}

// GetMTRoot returns the root of the Merkle Tree
func (s *StateDB) GetMTRoot(treeType enum) *big.Int {
	var root *big.Int
	if treeType == Account {
		root = s.AccountTree.Root().BigInt()
	} else {
		root = s.VouchTree.Root().BigInt()
>>>>>>> d31cee2 (feat/go-synchronizer initial construction of stateDB)
	}

	var nextLevel []*TreeNode

	for i := 0; i < len(leaves); i += 2 {
		var left, right *TreeNode
		left = leaves[i]
		if i+1 < len(leaves) {
			right = leaves[i+1]
		} else {
			right = &TreeNode{Hash: ""}
		}

		combinedHash := hashData(left.Hash + right.Hash)
		parentNode := &TreeNode{
			Hash:  combinedHash,
			Left:  left,
			Right: right,
		}
		nextLevel = append(nextLevel, parentNode)
	}

	return buildMerkleTree(nextLevel)
}

// updateMerkleTree updates the Merkle tree with a new leaf node.
func updateMerkleTree(tree *MerkleTree, newLeaf *TreeNode) {
	leaves := collectLeaves(tree.Root)
	leaves = append(leaves, newLeaf)
	tree.Root = buildMerkleTree(leaves)
}

// collectLeaves collects all the leaf nodes from a Merkle tree.
func collectLeaves(node *TreeNode) []*TreeNode {
	if node == nil {
		return nil
	}
	if node.Left == nil && node.Right == nil {
		return []*TreeNode{node}
	}
	leaves := collectLeaves(node.Left)
	leaves = append(leaves, collectLeaves(node.Right)...)
	return leaves
}

// findPathToRoot finds the path from a leaf node to the root node.
func FindPathToRoot(node *TreeNode, targetHash string) ([]string, bool) {
	if node == nil {
		return nil, false
	}
	if node.Hash == targetHash {
		return []string{node.Hash}, true
	}

	leftPath, found := FindPathToRoot(node.Left, targetHash)
	if found {
		return append(leftPath, node.Hash), true
	}

	rightPath, found := FindPathToRoot(node.Right, targetHash)
	if found {
		return append(rightPath, node.Hash), true
	}

	return nil, false
}

// verifyMerklePath verifies the Merkle path for a given leaf node hash.
func VerifyMerklePath(tree *MerkleTree, leafHash string) (string, bool) {
	path, found := FindPathToRoot(tree.Root, leafHash)
	if !found {
		return "", false
	}

	return path[len(path)-1], true
}

// GetRootForLeaf retrieves the root hash for a provided leaf node hash.
func (sdb *StateDB) GetRootForLeaf(leafHash string) (string, bool) {
	return VerifyMerklePath(sdb.Tree, leafHash)
}

func GetRootHash(s *StateDB, key string) string {
	leafHash, _ := s.GetTreeNodeHash(key)
	rootHash, verified := s.GetRootForLeaf(leafHash)
	if verified {
		fmt.Printf("Root hash for leaf %s: %s\n", leafHash, rootHash)
	} else {
		fmt.Println("Leaf not found in the Merkle tree")
	}
	return rootHash
}

func GetMerkelTreePath(s *StateDB, key string) ([]string, error) {
	nodeHash, _ := s.GetTreeNodeHash(key)
	path, found := FindPathToRoot(s.Tree.Root, nodeHash)
	if !found {
		return nil, fmt.Errorf("path not found for key: %s", key)
	}
	fmt.Printf("Merkle path: %v\n", path)
	return path, nil
}

func (s *StateDB) GetTreeNodeHash(key string) (string, error) {
	account, err := s.GetAccount(key)
	if err != nil {
		return "", err
	}
	accountBytes, err := json.Marshal(account)
	if err != nil {
		return "", err
	}
	return hashData(string(accountBytes)), nil
}
