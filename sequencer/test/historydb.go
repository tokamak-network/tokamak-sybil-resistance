package test

import (
	"math/big"
	"time"
	"tokamak-sybil-resistance/common"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// WARNING: the generators in this file doesn't necessary follow the protocol
// they are intended to check that the parsers between struct <==> DB are correct

// GenBlocks generates block from, to block numbers. WARNING: This is meant for DB/API testing, and
// may not be fully consistent with the protocol.
func GenBlocks(from, to int64) []common.Block {
	var blocks []common.Block
	for i := from; i < to; i++ {
		blocks = append(blocks, common.Block{
			Num: i,
			//nolint:gomnd
			Timestamp: time.Now().Add(time.Second * 13).UTC(),
			Hash:      ethCommon.BigToHash(big.NewInt(int64(i))),
		})
	}
	return blocks
}

// GenAccounts generates accounts. WARNING: This is meant for DB/API testing, and may not be fully
// consistent with the protocol.
func GenAccounts(totalAccounts, userAccounts int,
	userAddr *ethCommon.Address, userBjj *babyjub.PublicKey,
	batches []common.Batch) []common.Account {
	if totalAccounts < userAccounts {
		panic("totalAccounts must be greater than userAccounts")
	}
	accs := []common.Account{}
	for i := 256; i < 256+totalAccounts; i++ {
		var addr ethCommon.Address
		var pubK *babyjub.PublicKey
		if i < 256+userAccounts {
			addr = *userAddr
			pubK = userBjj
		} else {
			addr = ethCommon.BigToAddress(big.NewInt(int64(i)))
			privK := babyjub.NewRandPrivKey()
			pubK = privK.Public()
		}
		accs = append(accs, common.Account{
			Idx:      common.AccountIdx(i),
			EthAddr:  addr,
			BatchNum: batches[i%len(batches)].BatchNum,
			BJJ:      pubK.Compress(),
			Balance:  big.NewInt(int64(i * 10000000)), //nolint:gomnd
		})
	}
	return accs
}

// GenBatches generates batches. WARNING: This is meant for DB/API testing, and may not be fully
// consistent with the protocol.
func GenBatches(nBatches int, blocks []common.Block) []common.Batch {
	batches := []common.Batch{}
	collectedFees := make(map[common.TokenID]*big.Int)
	for i := 0; i < 64; i++ {
		collectedFees[common.TokenID(i)] = big.NewInt(int64(i))
	}
	for i := 0; i < nBatches; i++ {
		batch := common.Batch{
			BatchNum:    common.BatchNum(i + 1),
			EthBlockNum: blocks[i%len(blocks)].Num,
			//nolint:gomnd
			ForgerAddr:    ethCommon.BigToAddress(big.NewInt(6886723)),
			CollectedFees: collectedFees,
			StateRoot:     big.NewInt(int64(i+1) * 5), //nolint:gomnd
			//nolint:gomnd
			NumAccounts: 30,
			ExitRoot:    big.NewInt(int64(i+1) * 16), //nolint:gomnd
			SlotNum:     int64(i),
			GasPrice:    big.NewInt(int64(i + 1)),
		}
		if i%2 == 0 {
			toForge := new(int64)
			*toForge = int64(i + 1)
			batch.ForgeL1TxsNum = toForge
		}
		batches = append(batches, batch)
	}
	return batches
}
