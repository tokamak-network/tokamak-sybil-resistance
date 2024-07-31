package statedb

import (
	"encoding/json"
	"fmt"
)

type enum string

const (
	Account enum = "Account"
	Link    enum = "Link"
)

type TreeNodeHash interface {
}

// buildMerkleTree constructs a Merkle tree from a list of leaf nodes.
func buildMerkleTree(leaves []*TreeNode) *TreeNode {
	if len(leaves) == 1 {
		return leaves[0]
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
func (sdb *StateDB) GetRootForLeaf(leafHash string, treeType enum, key string) (string, bool) {
	if treeType == Account {
		return VerifyMerklePath(sdb.AccountTree, leafHash)
	} else {
		return VerifyMerklePath(sdb.LinkTree[key[:len(key)/2]], leafHash)
	}
}

func GetRootHash(s *StateDB, key string, treeType enum) string {
	leafHash, _ := s.GetTreeNodeHash(key, treeType)
	rootHash, verified := s.GetRootForLeaf(leafHash, treeType, key)
	if verified {
		fmt.Printf("Root hash for leaf %s: %s\n", leafHash, rootHash)
	} else {
		fmt.Println("Leaf not found in the Merkle tree")
	}
	return rootHash
}

func GetMerkelTreePath(s *StateDB, key string, treeType enum) ([]string, error) {
	nodeHash, _ := s.GetTreeNodeHash(key, treeType)
	var path []string
	var found bool
	if treeType == Account {
		path, found = FindPathToRoot(s.AccountTree.Root, nodeHash)
	} else {
		path, found = FindPathToRoot(s.LinkTree[key[:len(key)/2]].Root, nodeHash)
	}
	if !found {
		return nil, fmt.Errorf("path not found for key: %s", key)
	}
	return path, nil
}

func (s *StateDB) GetTreeNodeHash(key string, treeType enum) (string, error) {
	var (
		data interface{}
		err  error
	)
	if treeType == Account {
		data, err = s.GetAccount(key)
	} else {
		data, err = s.GetLink(key)
	}

	if err != nil {
		return "", err
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return hashData(string(bytes)), nil
}
