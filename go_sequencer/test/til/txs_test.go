package til

import (
	"encoding/hex"
	"fmt"
	"testing"
	"tokamak-sybil-resistance/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateKeys(t *testing.T) {
	tc := NewContext(0, common.RollupConstMaxL1UserTx)
	usernames := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}
	tc.generateKeys(usernames)
	debug := false
	if debug {
		for i, username := range usernames {
			fmt.Println(i, username)
			// sk := crypto.FromECDSA(tc.accounts[username].EthSk)
			// fmt.Println("	eth_sk", hex.EncodeToString(sk))
			fmt.Println("	eth_addr", tc.Accounts[username].Addr)
			fmt.Println("	bjj_sk", hex.EncodeToString(tc.Accounts[username].BJJ[:]))
			fmt.Println("	bjj_pub", tc.Accounts[username].BJJ.Public().Compress())
		}
	}
}

func TestGenerateBlocksNoBatches(t *testing.T) {
	set := `
		Type: Blockchain

		CreateAccountDeposit A: 11
		CreateAccountDeposit B: 22

		> block
	`
	tc := NewContext(0, common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(set)
	require.NoError(t, err)
	assert.Equal(t, 1, len(blocks))
	assert.Equal(t, 0, len(blocks[0].Rollup.Batches))
	assert.Equal(t, 2, len(blocks[0].Rollup.L1UserTxs))
}

func TestGenerateBlocks(t *testing.T) {
	set := `
		Type: Blockchain
	
		CreateAccountDeposit A: 10
		CreateAccountDeposit B: 5
		CreateAccountDeposit C: 5
		CreateAccountDeposit D: 5

		> batchL1 // batchNum = 1
		> batchL1 // batchNum = 2

		CreateVouch A-B
		CreateVouch B-A
		CreateVouch A-D
		DeleteVouch A-B

		// set new batch
		> batch // batchNum = 3

		> block
	`
	tc := NewContext(0, common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(set)
	require.NoError(t, err)
	assert.Equal(t, 1, len(blocks))
	assert.Equal(t, 3, len(blocks[0].Rollup.Batches))
	assert.Equal(t, 4, len(blocks[0].Rollup.L1UserTxs))

}
