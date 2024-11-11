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

// SCVariables joins all the smart contract variables in a single struct
type SCVariables struct {
	Rollup RollupVariables `validate:"required"`
}

// AsPtr returns the SCVariables as a SCVariablesPtr using pointers to the
// original SCVariables
func (v *SCVariables) AsPtr() *SCVariablesPtr {
	return &SCVariablesPtr{
		Rollup: &v.Rollup,
	}
}
