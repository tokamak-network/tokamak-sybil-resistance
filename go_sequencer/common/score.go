package common

// Score is a struct that gives an information about score
// of each accounts.
type Score struct {
	Idx      Idx      `meddler:"idx"`
	BatchNum BatchNum `meddler:"batch_num"`
	Nonce    Nonce    `meddler:"-"` // max of 40 bits used
	Score    bool     `meddler:"score"`
}
