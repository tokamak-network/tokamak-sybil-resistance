package models

// Account represents an account with specified fields.
type Account struct {
	Idx     int
	EthAddr string
	Sign    bool
	Ay      string
	Balance int
	Score   int
	Nonce   int
}
