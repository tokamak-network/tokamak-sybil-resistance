package common

// Vouch is a struct that gives an information about vouch
// between accounts. Each Idx is represented by fromIdx and toIdx
// of each accounts.
type Vouch struct {
	Idx      Idx      `meddler:"idx"`
	BatchNum BatchNum `meddler:"batch_num"`
	Nonce    Nonce    `meddler:"-"` // max of 40 bits used
	Value    bool     `meddler:"value"`
}
