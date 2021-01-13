package exported

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
)

func PubkeyToAddress(pk ecdsa.PublicKey) (string, error) {
	return crypto.PubkeyToAddress(pk).String(), nil
}
