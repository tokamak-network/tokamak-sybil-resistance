package nonce

const (
	// MaxNonceValue is the maximum value that the Account.Nonce can have
	// (40 bits: MaxNonceValue=2**40-1)
	MaxNonceValue = 0xffffffffff
)

// Nonce represents the nonce value in a uint64, which has the method Bytes
// that returns a byte array of length 5 (40 bits).
type Nonce uint64
