package til

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBlockchainTxs(t *testing.T) {
	s := `
		Type: Blockchain

		Deposit A: 10
		Deposit A: 20
		Deposit B: 5
		CreateVouch A-B

		// set new batch
		> batch

		CreateAccountDeposit C: 5
		CreateVouch B-C
		Deposit User0: 20
		Deposit User1: 20

		> batch
		> block

		DeleteVouch A-B

		// Exits
		Exit A: 5
	`

	parser := newParser(strings.NewReader(s))
	instructions, err := parser.parse()
	require.NoError(t, err)
	assert.Equal(t, 13, len(instructions.instructions))
	assert.Equal(t, 5, len(instructions.users))

	if Debug {
		for _, instruction := range instructions.instructions {
			fmt.Println(instruction.raw())
		}
	}

	assert.Equal(t, TypeNewBatch, instructions.instructions[4].Typ)
	assert.Equal(t, "DepositUser0:20", instructions.instructions[7].raw())
	assert.Equal(t, "ExitA:5", instructions.instructions[12].raw())
	assert.Equal(t, "Type: Exit, From: A, Amount: 5\n",
		instructions.instructions[12].String())
}

func TestParsePoolTxs(t *testing.T) {
	s := `
		Type: PoolL2
		PoolCreateVouch A-B
		PoolCreateVouch B-C
		PoolExit A: 5
		PoolDeleteVouch A-B
	`

	parser := newParser(strings.NewReader(s))
	instructions, err := parser.parse()
	require.NoError(t, err)

	if Debug {
		for _, instruction := range instructions.instructions {
			fmt.Println(instruction.String())
		}
	}

	assert.Equal(t, 4, len(instructions.instructions))
	assert.Equal(t, 2, len(instructions.users))

	assert.Equal(t, "ExitA:5", instructions.instructions[2].raw())
}

func TestParseErrors(t *testing.T) {
	s := `
		Type: Blockchain
		Deposit A:: 10
	`
	parser := newParser(strings.NewReader(s))
	_, err := parser.parse()
	assert.Equal(t, "line 2: DepositA:: 10\n, err: can not parse number for Amount: :", err.Error())

	s = `
		Type: Blockchain
		Deposit A: 10 20
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "line 3: 20, err: unexpected Blockchain tx type: 20", err.Error())

	s = `
		Type: Blockchain
		> btch
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t,
		"line 2: >, err: unexpected '> btch', expected '> batch' or '> block'",
		err.Error())

	// check definition of set Type
	s = `PoolExit A: 10`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "line 1: PoolExit, err: set type not defined", err.Error())

	s = `Type: PoolL1`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t,
		"line 1: Type:, err: invalid set type: 'PoolL1'. Valid set types: 'Blockchain', 'PoolL2'",
		err.Error())

	s = `Type: PoolL1
		Type: Blockchain`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t,
		"line 1: Type:, err: invalid set type: 'PoolL1'. Valid set types: 'Blockchain', 'PoolL2'",
		err.Error())

	s = `Type: PoolL2
		Type: Blockchain`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t,
		"line 2: Instruction of 'Type: Blockchain' when there is already a previous "+
			"instruction 'Type: PoolL2' defined", err.Error())
}
