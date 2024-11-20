package txselector

const (
	// ErrInvalidAtomicGroup error message returned if an atomic group is malformed
	ErrInvalidAtomicGroup = "Tx not selected because it belongs to an atomic group with missing transactions or bad requested transaction"
	// ErrInvalidAtomicGroupCode error code
	ErrInvalidAtomicGroupCode int = 18
	// ErrInvalidAtomicGroupType error type
	ErrInvalidAtomicGroupType string = "ErrInvalidAtomicGroup"
)
