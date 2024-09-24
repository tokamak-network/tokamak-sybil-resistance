package statedb

import (
	"bytes"
	"fmt"
	"tokamak-sybil-resistance/common"
	"tokamak-sybil-resistance/log"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree/db"
)

func concatEthAddr(addr ethCommon.Address) []byte {
	var b []byte
	b = append(b, addr.Bytes()...)
	return b
}
func concatEthAddrBJJ(addr ethCommon.Address, pk babyjub.PublicKeyComp) []byte {
	pkComp := pk
	var b []byte
	b = append(b, addr.Bytes()...)
	b = append(b[:], pkComp[:]...)
	return b
}

// GetIdxByEthAddr returns the smallest Idx in the StateDB for the given
// Ethereum Address. Will return common.Idx(0) and error in case that Idx is
// not found in the StateDB.
func (s *StateDB) GetIdxByEthAddr(addr ethCommon.Address) (common.Idx,
	error) {
	k := concatEthAddr(addr)
	b, err := s.db.DB().Get(append(PrefixKeyAddr, k...))
	if err != nil {
		return common.Idx(0), common.Wrap(fmt.Errorf("GetIdxByEthAddr: %s: ToEthAddr: %s",
			ErrIdxNotFound, addr.Hex()))
	}
	idx, err := common.IdxFromBytes(b)
	if err != nil {
		return common.Idx(0), common.Wrap(fmt.Errorf("GetIdxByEthAddr: %s: ToEthAddr: %s",
			err, addr.Hex()))
	}
	return idx, nil
}

// GetIdxByEthAddrBJJ returns the smallest Idx in the StateDB for the given
// Ethereum Address AND the given BabyJubJub PublicKey. If `addr` is the zero
// address, it's ignored in the query.  If `pk` is nil, it's ignored in the
// query.  Will return common.Idx(0) and error in case that Idx is not found in
// the StateDB.
func (s *StateDB) GetIdxByEthAddrBJJ(addr ethCommon.Address, pk babyjub.PublicKeyComp) (common.Idx, error) {
	if !bytes.Equal(addr.Bytes(), common.EmptyAddr.Bytes()) && pk == common.EmptyBJJComp {
		// ToEthAddr
		// case ToEthAddr!=0 && ToBJJ=0
		return s.GetIdxByEthAddr(addr)
	} else if !bytes.Equal(addr.Bytes(), common.EmptyAddr.Bytes()) &&
		pk != common.EmptyBJJComp {
		// case ToEthAddr!=0 && ToBJJ!=0
		k := concatEthAddrBJJ(addr, pk)
		b, err := s.db.DB().Get(append(PrefixKeyAddrBJJ, k...))
		if common.Unwrap(err) == db.ErrNotFound {
			// return the error (ErrNotFound), so can be traced at upper layers
			return common.Idx(0), common.Wrap(ErrIdxNotFound)
		} else if err != nil {
			return common.Idx(0),
				common.Wrap(fmt.Errorf("GetIdxByEthAddrBJJ: %s: ToEthAddr: %s, ToBJJ: %s",
					ErrIdxNotFound, addr.Hex(), pk))
		}
		idx, err := common.IdxFromBytes(b)
		if err != nil {
			return common.Idx(0),
				common.Wrap(fmt.Errorf("GetIdxByEthAddrBJJ: %s: ToEthAddr: %s, ToBJJ: %s",
					err, addr.Hex(), pk))
		}
		return idx, nil
	}
	// rest of cases (included case ToEthAddr==0) are not possible
	return common.Idx(0),
		common.Wrap(
			fmt.Errorf("GetIdxByEthAddrBJJ: Not found, %s: ToEthAddr: %s, ToBJJ: %s",
				ErrGetIdxNoCase, addr.Hex(), pk))
}

// setIdxByEthAddrBJJ stores the given Idx in the StateDB as follows:
// - key: Eth Address, value: idx
// - key: EthAddr & BabyJubJub PublicKey Compressed, value: idx
// If Idx already exist for the given EthAddr & BJJ, the remaining Idx will be
// always the smallest one.
func (s *StateDB) setIdxByEthAddrBJJ(idx common.Idx, addr ethCommon.Address,
	pk babyjub.PublicKeyComp) error {
	oldIdx, err := s.GetIdxByEthAddrBJJ(addr, pk)
	if err == nil {
		// EthAddr & BJJ already have an Idx
		// check which Idx is smaller
		// if new idx is smaller, store the new one
		// if new idx is bigger, don't store and return, as the used one will be the old
		if idx >= oldIdx {
			log.Debug("StateDB.setIdxByEthAddrBJJ: Idx not stored because there " +
				"already exist a smaller Idx for the given EthAddr & BJJ")
			return nil
		}
	}

	// store idx for EthAddr & BJJ assuming that EthAddr & BJJ still don't
	// have an Idx stored in the DB, and if so, the already stored Idx is
	// bigger than the given one, so should be updated to the new one
	// (smaller)
	tx, err := s.db.DB().NewTx()
	if err != nil {
		return common.Wrap(err)
	}
	idxBytes, err := idx.Bytes()
	if err != nil {
		return common.Wrap(err)
	}
	// store Addr&BJJ-idx
	k := concatEthAddrBJJ(addr, pk)
	err = tx.Put(append(PrefixKeyAddrBJJ, k...), idxBytes[:])
	if err != nil {
		return common.Wrap(err)
	}

	// store Addr-idx
	k = concatEthAddr(addr)
	err = tx.Put(append(PrefixKeyAddr, k...), idxBytes[:])
	if err != nil {
		return common.Wrap(err)
	}
	err = tx.Commit()
	if err != nil {
		return common.Wrap(err)
	}
	return nil
}
