package txprocessor

import (
	"math/big"

	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree"
)

// BJJCompressedTo256BigInts returns a [256]*big.Int array with the bit
// representation of the babyjub.PublicKeyComp
func BJJCompressedTo256BigInts(pkComp babyjub.PublicKeyComp) [256]*big.Int {
	var r [256]*big.Int
	b := pkComp[:]

	for i := 0; i < 256; i++ {
		if b[i/8]&(1<<(i%8)) == 0 { //nolint:gomnd
			r[i] = big.NewInt(0)
		} else {
			r[i] = big.NewInt(1)
		}
	}

	return r
}

func siblingsToZKInputFormat(s []*merkletree.Hash) []*big.Int {
	b := make([]*big.Int, len(s))
	for i := 0; i < len(s); i++ {
		b[i] = s[i].BigInt()
	}
	return b
}
