package tokamak

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TODO: Update this
const TokamakABI = ""

type Tokamak struct {
	TokamakCaller     // Read-only binding to the contract
	TokamakTransactor // Write-only binding to the contract
	TokamakFilterer   // Log filterer for contract events
}

// TokamakCaller is an auto generated read-only Go binding around an Ethereum contract.
type TokamakCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TokamakTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TokamakTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TokamakFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TokamakFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Solidity: function ABSOLUTE_MAX_L1L2BATCHTIMEOUT() view returns(uint8)
func (_Tokamak *TokamakCaller) ABSOLUTEMAXL1L2BATCHTIMEOUT(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Tokamak.contract.Call(opts, &out, "ABSOLUTE_MAX_L1L2BATCHTIMEOUT")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Solidity: function rollupVerifiersLength() view returns(uint256)
func (_Tokamak *TokamakCaller) RollupVerifiersLength(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Tokamak.contract.Call(opts, &out, "rollupVerifiersLength")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err
}

// Solidity: function rollupVerifiers(uint256 ) view returns(address verifierInterface, uint256 maxTx, uint256 nLevels)
func (_Tokamak *TokamakCaller) RollupVerifiers(opts *bind.CallOpts, arg0 *big.Int) (struct {
	VerifierInterface common.Address
	MaxTx             *big.Int
	NLevels           *big.Int
}, error) {
	var out []interface{}
	err := _Tokamak.contract.Call(opts, &out, "rollupVerifiers", arg0)

	outstruct := new(struct {
		VerifierInterface common.Address
		MaxTx             *big.Int
		NLevels           *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.VerifierInterface = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.MaxTx = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.NLevels = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// LastForgedBatch is a free data retrieval call binding the contract method 0x44e0b2ce.
//
// Solidity: function lastForgedBatch() view returns(uint32)
func (_Tokamak *TokamakCaller) LastForgedBatch(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _Tokamak.contract.Call(opts, &out, "lastForgedBatch")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// NewTokamak creates a new instance of Tokamak, bound to a specific deployed contract.
func NewTokamak(address common.Address, backend bind.ContractBackend) (*Tokamak, error) {
	contract, err := bindTokamak(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Tokamak{TokamakCaller: TokamakCaller{contract: contract}, TokamakTransactor: TokamakTransactor{contract: contract}, TokamakFilterer: TokamakFilterer{contract: contract}}, nil
}

// bindTokamak binds a generic wrapper to an already deployed contract.
func bindTokamak(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(TokamakABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// TODO: Need to check this, with sm team
// Solidity: function forgeBatch(uint48 newLastIdx, uint256 newStRoot, uint256 newExitRoot, bytes encodedL1CoordinatorTx, bytes l1L2TxsData, bytes feeIdxCoordinator, uint8 verifierIdx, bool l1Batch, uint256[2] proofA, uint256[2][2] proofB, uint256[2] proofC) returns()
func (_Tokamak *TokamakTransactor) ForgeBatch(opts *bind.TransactOpts, newLastIdx *big.Int, newStRoot *big.Int, newExitRoot *big.Int, l1L2TxsData []byte, verifierIdx uint8, l1Batch bool, proofA [2]*big.Int, proofB [2][2]*big.Int, proofC [2]*big.Int) (*types.Transaction, error) {
	return _Tokamak.contract.Transact(opts, "forgeBatch", newLastIdx, newStRoot, newExitRoot, l1L2TxsData, verifierIdx, l1Batch, proofA, proofB, proofC)
}
