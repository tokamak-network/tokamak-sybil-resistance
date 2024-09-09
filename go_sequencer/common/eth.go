package common

// SCVariablesPtr joins all the smart contract variables as pointers in a single
// struct
type SCVariablesPtr struct {
	Rollup *RollupVariables `validate:"required"`
}

// SCConsts joins all the smart contract constants in a single struct
type SCConsts struct {
	Rollup RollupConstants
}
