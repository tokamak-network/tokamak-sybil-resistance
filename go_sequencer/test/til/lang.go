package til

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/big"
	"sort"
	"strconv"
	"tokamak-sybil-resistance/common"
<<<<<<< HEAD
	"tokamak-sybil-resistance/log"
=======
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
)

var eof = rune(0)
var errof = fmt.Errorf("eof in parseline")
var commentLine = fmt.Errorf("comment in parseline") //nolint:golint
var newEventLine = fmt.Errorf("newEventLine")        //nolint:golint
var setTypeLine = fmt.Errorf("setTypeLine")          //nolint:golint

// setType defines the type of the set
type setType string

// SetTypeBlockchain defines the type 'Blockchain' of the set
var SetTypeBlockchain = setType("Blockchain")

// SetTypePoolL2 defines the type 'PoolL2' of the set
var SetTypePoolL2 = setType("PoolL2")

// TypeNewBatch is used for testing purposes only, and represents the
// common.TxType of a new batch
var TypeNewBatch common.TxType = "InstrTypeNewBatch"

// TypeNewBatchL1 is used for testing purposes only, and represents the
// common.TxType of a new batch
var TypeNewBatchL1 common.TxType = "InstrTypeNewBatchL1"

// TypeNewBlock is used for testing purposes only, and represents the
// common.TxType of a new ethereum block
var TypeNewBlock common.TxType = "InstrTypeNewBlock"

<<<<<<< HEAD
=======
// TypeAddToken is used for testing purposes only, and represents the
// common.TxType of a new Token regsitration.
// It has 'nolint:gosec' as the string 'Token' triggers gosec as a potential
// leaked Token (which is not the case).
var TypeAddToken common.TxType = "InstrTypeAddToken" //nolint:gosec

>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
// TxTypeCreateAccountDepositCoordinator  is used for testing purposes only,
// and represents the common.TxType of a create acount deposit made by the
// coordinator
var TxTypeCreateAccountDepositCoordinator common.TxType = "TypeCreateAccountDepositCoordinator"

// nolint
const (
	ILLEGAL token = iota
	WS
	EOF

	IDENT // val
)

// Instruction is the data structure that represents one line of code
type Instruction struct {
	LineNum       int
	Literal       string
	From          string
	To            string
	Amount        *big.Int
	DepositAmount *big.Int
	Fee           uint8
<<<<<<< HEAD
	// TokenID       common.TokenID
	Typ common.TxType // D: Deposit, T: Transfer, E: ForceExit
=======
	TokenID       common.TokenID
	Typ           common.TxType // D: Deposit, T: Transfer, E: ForceExit
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
}

// parsedSet contains the full Set of Instructions representing a full code
type parsedSet struct {
	typ          setType
	instructions []Instruction
	users        []string
}

func (i Instruction) String() string {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "Type: %s, ", i.Typ)
	fmt.Fprintf(buf, "From: %s, ", i.From)
<<<<<<< HEAD
	if i.Typ == common.TxTypeCreateVouch ||
		i.Typ == common.TxTypeDeleteVouch {
=======
	if i.Typ == common.TxTypeTransfer ||
		i.Typ == common.TxTypeDepositTransfer ||
		i.Typ == common.TxTypeCreateAccountDepositTransfer {
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
		fmt.Fprintf(buf, "To: %s, ", i.To)
	}

	if i.Typ == common.TxTypeDeposit ||
<<<<<<< HEAD
		i.Typ == common.TxTypeDepositTransfer ||
<<<<<<< HEAD
=======
>>>>>>> 3523c13 (Fixed and removed redundant part)
		i.Typ == common.TxTypeCreateAccountDeposit {
		fmt.Fprintf(buf, "DepositAmount: %d, ", i.DepositAmount)
	}
	// if i.Typ != common.TxTypeDeposit {
	// 	fmt.Fprintf(buf, "Amount: %d, ", i.Amount)
	// }
	// if i.Typ == common.TxTypeTransfer ||
	// 	i.Typ == common.TxTypeDepositTransfer ||
	// 	i.Typ == common.TxTypeCreateAccountDepositTransfer {
	// 	fmt.Fprintf(buf, "Fee: %d, ", i.Fee)
	// }
	// fmt.Fprintf(buf, "TokenID: %d\n", i.TokenID)
=======
		i.Typ == common.TxTypeCreateAccountDepositTransfer {
		fmt.Fprintf(buf, "DepositAmount: %d, ", i.DepositAmount)
	}
	if i.Typ != common.TxTypeDeposit {
		fmt.Fprintf(buf, "Amount: %d, ", i.Amount)
	}
	if i.Typ == common.TxTypeTransfer ||
		i.Typ == common.TxTypeDepositTransfer ||
		i.Typ == common.TxTypeCreateAccountDepositTransfer {
		fmt.Fprintf(buf, "Fee: %d, ", i.Fee)
	}
	fmt.Fprintf(buf, "TokenID: %d\n", i.TokenID)
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
	return buf.String()
}

// Raw returns a string with the raw representation of the Instruction
func (i Instruction) raw() string {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "%s", i.Typ)
<<<<<<< HEAD
	// fmt.Fprintf(buf, "(%d)", i.TokenID)
=======
	fmt.Fprintf(buf, "(%d)", i.TokenID)
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
	fmt.Fprintf(buf, "%s", i.From)
	fmt.Fprintf(buf, ":")
	if i.Typ == common.TxTypeDeposit {
		fmt.Fprintf(buf, "%d", i.DepositAmount)
	}
	if i.Typ != common.TxTypeDeposit {
		fmt.Fprintf(buf, "%d", i.Amount)
	}
	return buf.String()
}

type token int

type scanner struct {
	r *bufio.Reader
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\v' || ch == '\f'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isComment(ch rune) bool {
	return ch == '/'
}

func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

// newScanner creates a new scanner with the given io.Reader
func newScanner(r io.Reader) *scanner {
	return &scanner{r: bufio.NewReader(r)}
}

func (s *scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *scanner) unread() {
	_ = s.r.UnreadRune()
}

// scan returns the token and literal string of the current value
func (s *scanner) scan() (tok token, lit string) {
	ch := s.read()

	if isWhitespace(ch) {
		// space
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) || isDigit(ch) {
		// letter/digit
		s.unread()
		return s.scanIndent()
	} else if isComment(ch) {
		// comment
		s.unread()
		return s.scanIndent()
	}

	if ch == eof {
		return EOF, ""
	}

	return ILLEGAL, string(ch)
}

func (s *scanner) scanWhitespace() (token token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}
	return WS, buf.String()
}

func (s *scanner) scanIndent() (tok token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	if len(buf.String()) == 1 {
		return token(rune(buf.String()[0])), buf.String()
	}
	return IDENT, buf.String()
}

// parser defines the parser
type parser struct {
	s   *scanner
	buf struct {
		tok token
		lit string
		n   int
	}
}

// newParser creates a new parser from a io.Reader
func newParser(r io.Reader) *parser {
	return &parser{s: newScanner(r)}
}

func (p *parser) scan() (tok token, lit string) {
	// if there is a token in the buffer return it
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}
	tok, lit = p.s.scan()

	p.buf.tok, p.buf.lit = tok, lit

	return
}

func (p *parser) scanIgnoreWhitespace() (tok token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

// parseLine parses the current line
func (p *parser) parseLine(setType setType) (*Instruction, error) {
	c := &Instruction{}
	tok, lit := p.scanIgnoreWhitespace()
	if tok == EOF {
		return nil, common.Wrap(errof)
	}
	c.Literal += lit
	if lit == "/" {
		_, _ = p.s.r.ReadString('\n')
		return nil, commentLine
	} else if lit == ">" {
		if setType == SetTypePoolL2 {
			return c, common.Wrap(fmt.Errorf("Unexpected '>' at PoolL2Txs set"))
		}
		_, lit = p.scanIgnoreWhitespace()
		if lit == "batch" {
			_, _ = p.s.r.ReadString('\n')
			return &Instruction{Typ: TypeNewBatch}, newEventLine
		} else if lit == "batchL1" {
			_, _ = p.s.r.ReadString('\n')
			return &Instruction{Typ: TypeNewBatchL1}, newEventLine
		} else if lit == "block" {
			_, _ = p.s.r.ReadString('\n')
			return &Instruction{Typ: TypeNewBlock}, newEventLine
		} else {
			return c, common.Wrap(fmt.Errorf("Unexpected '> %s', expected '> batch' or '> block'", lit))
		}
	} else if lit == "Type" {
		if err := p.expectChar(c, ":"); err != nil {
			return c, common.Wrap(err)
		}
		_, lit = p.scanIgnoreWhitespace()
		if lit == "Blockchain" {
			return &Instruction{Typ: "Blockchain"}, setTypeLine
		} else if lit == "PoolL2" {
			return &Instruction{Typ: "PoolL2"}, setTypeLine
		} else {
			return c,
				common.Wrap(fmt.Errorf("Invalid set type: '%s'. Valid set types: 'Blockchain', 'PoolL2'", lit))
		}
<<<<<<< HEAD
=======
	} else if lit == "AddToken" {
		if err := p.expectChar(c, "("); err != nil {
			return c, common.Wrap(err)
		}
		_, lit = p.scanIgnoreWhitespace()
		c.Literal += lit
		tidI, err := strconv.Atoi(lit)
		if err != nil {
			line, _ := p.s.r.ReadString('\n')
			c.Literal += line
			return c, common.Wrap(err)
		}
		c.TokenID = common.TokenID(tidI)
		if err := p.expectChar(c, ")"); err != nil {
			return c, common.Wrap(err)
		}
		c.Typ = TypeAddToken
		line, _ := p.s.r.ReadString('\n')
		c.Literal += line
		return c, newEventLine
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
	}

	if setType == "" {
		return c, common.Wrap(fmt.Errorf("Set type not defined"))
	}
<<<<<<< HEAD
	vouch := false
=======
	transferring := false
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
	fee := false

	if setType == SetTypeBlockchain {
		switch lit {
		case "Deposit":
			c.Typ = common.TxTypeDeposit
		case "Exit":
			c.Typ = common.TxTypeExit
<<<<<<< HEAD
			// fee = true
		case "CreateVouch":
			c.Typ = common.TxTypeCreateVouch
			vouch = true
		case "DeleteVouch":
			c.Typ = common.TxTypeDeleteVouch
			vouch = true
		case "CreateAccountDeposit":
			c.Typ = common.TxTypeCreateAccountDeposit
=======
			fee = true
		case "Transfer":
			c.Typ = common.TxTypeTransfer
			transferring = true
			fee = true
		case "CreateAccountDeposit":
			c.Typ = common.TxTypeCreateAccountDeposit
		case "CreateAccountDepositTransfer":
			c.Typ = common.TxTypeCreateAccountDepositTransfer
			transferring = true
		case "CreateAccountCoordinator":
			c.Typ = TxTypeCreateAccountDepositCoordinator
			// transferring is false, as the Coordinator tx transfer will be 0
		case "DepositTransfer":
			c.Typ = common.TxTypeDepositTransfer
			transferring = true
		case "ForceTransfer":
			c.Typ = common.TxTypeForceTransfer
			transferring = true
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
		case "ForceExit":
			c.Typ = common.TxTypeForceExit
		default:
			return c, common.Wrap(fmt.Errorf("Unexpected Blockchain tx type: %s", lit))
		}
	} else if setType == SetTypePoolL2 {
		switch lit {
<<<<<<< HEAD
		case "PoolCreateVouch":
			c.Typ = common.TxTypeCreateVouch
			vouch = true
			fee = true
		case "PoolDeleteVouch":
			c.Typ = common.TxTypeDeleteVouch
			vouch = true
=======
		case "PoolTransfer":
			c.Typ = common.TxTypeTransfer
			transferring = true
			fee = true
		case "PoolTransferToEthAddr":
			c.Typ = common.TxTypeTransferToEthAddr
			transferring = true
			fee = true
		case "PoolTransferToBJJ":
			c.Typ = common.TxTypeTransferToBJJ
			transferring = true
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
			fee = true
		case "PoolExit":
			c.Typ = common.TxTypeExit
			fee = true
		default:
			return c, common.Wrap(fmt.Errorf("Unexpected PoolL2 tx type: %s", lit))
		}
	} else {
		return c,
			common.Wrap(fmt.Errorf("Invalid set type: '%s'. Valid set types: 'Blockchain', 'PoolL2'",
				setType))
	}

<<<<<<< HEAD
	_, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	c.From = lit

	_, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	if vouch {
=======
	if err := p.expectChar(c, "("); err != nil {
		return c, common.Wrap(err)
	}
	_, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	tidI, err := strconv.Atoi(lit)
	if err != nil {
		line, _ := p.s.r.ReadString('\n')
		c.Literal += line
		return c, common.Wrap(err)
	}
	c.TokenID = common.TokenID(tidI)
	if err := p.expectChar(c, ")"); err != nil {
		return c, common.Wrap(err)
	}
	_, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	c.From = lit
	if c.Typ == TxTypeCreateAccountDepositCoordinator {
		line, _ := p.s.r.ReadString('\n')
		c.Literal += line
		return c, nil
	}
	_, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	if transferring {
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
		if lit != "-" {
			return c, common.Wrap(fmt.Errorf("Expected '-', found '%s'", lit))
		}
		_, lit = p.scanIgnoreWhitespace()
		c.Literal += lit
		c.To = lit
<<<<<<< HEAD
		line, _ := p.s.r.ReadString('\n')
		c.Literal += line
	} else {
		if lit != ":" {
			line, _ := p.s.r.ReadString('\n')
			c.Literal += line
			return c, common.Wrap(fmt.Errorf("Expected ':', found '%s'", lit))
		}
		_, lit = p.scanIgnoreWhitespace()
		c.Literal += lit
		amount, ok := new(big.Int).SetString(lit, 10)
		if !ok {
			line, _ := p.s.r.ReadString('\n')
			c.Literal += line
			return c, common.Wrap(fmt.Errorf("Can not parse number for Amount: %s", lit))
		}
		if c.Typ == common.TxTypeDeposit ||
			c.Typ == common.TxTypeCreateAccountDeposit {
			c.DepositAmount = amount
		} else {
			c.Amount = amount
		}
		if fee {
			if err := p.expectChar(c, "("); err != nil {
				return c, common.Wrap(err)
			}
			_, lit = p.scanIgnoreWhitespace()
			c.Literal += lit
			fee, err := strconv.Atoi(lit)
			if err != nil {
				line, _ := p.s.r.ReadString('\n')
				c.Literal += line
				return c, common.Wrap(err)
			}
			if fee > common.MaxFeePlan-1 {
				line, _ := p.s.r.ReadString('\n')
				c.Literal += line
				return c, common.Wrap(fmt.Errorf("Fee %d can not be bigger than 255", fee))
			}
			c.Fee = uint8(fee)

			if err := p.expectChar(c, ")"); err != nil {
				return c, common.Wrap(err)
			}
		}

		if tok == EOF {
			return nil, common.Wrap(errof)
		}
=======
		_, lit = p.scanIgnoreWhitespace()
		c.Literal += lit
	}
	if lit != ":" {
		line, _ := p.s.r.ReadString('\n')
		c.Literal += line
		return c, common.Wrap(fmt.Errorf("Expected ':', found '%s'", lit))
	}
	if c.Typ == common.TxTypeDepositTransfer ||
		c.Typ == common.TxTypeCreateAccountDepositTransfer {
		// deposit case
		_, lit = p.scanIgnoreWhitespace()
		c.Literal += lit
		depositAmount, ok := new(big.Int).SetString(lit, 10)
		if !ok {
			line, _ := p.s.r.ReadString('\n')
			c.Literal += line
			return c, common.Wrap(fmt.Errorf("Can not parse number for DepositAmount"))
		}
		c.DepositAmount = depositAmount
		if err := p.expectChar(c, ","); err != nil {
			return c, common.Wrap(err)
		}
	}
	_, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	amount, ok := new(big.Int).SetString(lit, 10)
	if !ok {
		line, _ := p.s.r.ReadString('\n')
		c.Literal += line
		return c, common.Wrap(fmt.Errorf("Can not parse number for Amount: %s", lit))
	}
	if c.Typ == common.TxTypeDeposit ||
		c.Typ == common.TxTypeCreateAccountDeposit {
		c.DepositAmount = amount
	} else {
		c.Amount = amount
	}
	if fee {
		if err := p.expectChar(c, "("); err != nil {
			return c, common.Wrap(err)
		}
		_, lit = p.scanIgnoreWhitespace()
		c.Literal += lit
		fee, err := strconv.Atoi(lit)
		if err != nil {
			line, _ := p.s.r.ReadString('\n')
			c.Literal += line
			return c, common.Wrap(err)
		}
		if fee > common.MaxFeePlan-1 {
			line, _ := p.s.r.ReadString('\n')
			c.Literal += line
			return c, common.Wrap(fmt.Errorf("Fee %d can not be bigger than 255", fee))
		}
		c.Fee = uint8(fee)

		if err := p.expectChar(c, ")"); err != nil {
			return c, common.Wrap(err)
		}
	}

	if tok == EOF {
		return nil, common.Wrap(errof)
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
	}
	return c, nil
}

func (p *parser) expectChar(c *Instruction, ch string) error {
	_, lit := p.scanIgnoreWhitespace()
	c.Literal += lit
	if lit != ch {
		line, _ := p.s.r.ReadString('\n')
		c.Literal += line
		return common.Wrap(fmt.Errorf("Expected '%s', found '%s'", ch, lit))
	}
	return nil
}

<<<<<<< HEAD
=======
func idxTokenIDToString(idx string, tid common.TokenID) string {
	return idx + strconv.Itoa(int(tid))
}

>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
// parse parses through reader
func (p *parser) parse() (*parsedSet, error) {
	ps := &parsedSet{}
	i := 0 // lines will start counting at line 1
	users := make(map[string]bool)
	for {
		i++
		instruction, err := p.parseLine(ps.typ)
		if common.Unwrap(err) == errof {
			break
		}
		if common.Unwrap(err) == setTypeLine {
			if ps.typ != "" {
				return ps,
					common.Wrap(fmt.Errorf("Line %d: Instruction of 'Type: %s' when "+
						"there is already a previous instruction 'Type: %s' defined",
						i, instruction.Typ, ps.typ))
			}
			if instruction.Typ == "PoolL2" {
				ps.typ = SetTypePoolL2
			} else if instruction.Typ == "Blockchain" {
				ps.typ = SetTypeBlockchain
			} else {
				log.Fatalf("Line %d: Invalid set type: '%s'. Valid set types: "+
					"'Blockchain', 'PoolL2'", i, instruction.Typ)
			}
			continue
		}
		if common.Unwrap(err) == commentLine {
			continue
		}
		instruction.LineNum = i
		if common.Unwrap(err) == newEventLine {
<<<<<<< HEAD
=======
			if instruction.Typ == TypeAddToken && instruction.TokenID == common.TokenID(0) {
				return ps, common.Wrap(fmt.Errorf("Line %d: AddToken can not register TokenID 0", i))
			}
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
			ps.instructions = append(ps.instructions, *instruction)
			continue
		}
		if err != nil {
			return ps, common.Wrap(fmt.Errorf("Line %d: %s, err: %s", i, instruction.Literal, err.Error()))
		}
		if ps.typ == "" {
			return ps, common.Wrap(fmt.Errorf("Line %d: Set type not defined", i))
		}
		ps.instructions = append(ps.instructions, *instruction)
		users[instruction.From] = true
<<<<<<< HEAD
=======
		if instruction.Typ == common.TxTypeTransfer ||
			instruction.Typ == common.TxTypeTransferToEthAddr ||
			instruction.Typ == common.TxTypeTransferToBJJ { // type: Transfer
			users[instruction.To] = true
		}
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
	}
	for u := range users {
		ps.users = append(ps.users, u)
	}
	sort.Strings(ps.users)
	return ps, nil
}
