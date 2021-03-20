package exported

import (
	"crypto/ecdsa"
	"math/big"
)

// Signature - an ECDSA signature
type Signature struct {
	R *big.Int
	S *big.Int
}

// Key contains the public key value and corresponding ID
type Key struct {
	ID    string
	Value ecdsa.PublicKey
}
