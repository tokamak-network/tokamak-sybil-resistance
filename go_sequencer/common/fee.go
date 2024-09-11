package common

// FeeSelector is used to select a percentage from the FeePlan.
type FeeSelector uint8

<<<<<<< HEAD
=======
// RecommendedFee is the recommended fee to pay in USD per transaction set by
// the coordinator according to the tx type (if the tx requires to create an
// account and register, only register or he account already esists)
type RecommendedFee struct {
	ExistingAccount        float64 `json:"existingAccount"`
	CreatesAccount         float64 `json:"createAccount"`
	CreatesAccountInternal float64 `json:"createAccountInternal"`
}

>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
// MaxFeePlan is the maximum value of the FeePlan
const MaxFeePlan = 256
