package types

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
)

const (
	// HashLength is the expected length of the hash
	HashLength = 32
)

// Hash is an alias for a 32-byte hash
type Hash []byte

var zeroHash [HashLength]byte

// ValidateBasic returns an error if the hash is not a valid
func (h Hash) ValidateBasic() error {
	if len(h) != HashLength {
		return fmt.Errorf("hash length must be %d", HashLength)
	}

	if bytes.Equal(h, zeroHash[:]) {
		return fmt.Errorf("hash cannot be zero")
	}

	return nil
}

// Signature is an alias for signature in raw bytes
type Signature []byte

// ValidateBasic returns an error if the signature is not a valid S256 elliptic curve signature
func (sig Signature) ValidateBasic() error {
	_, err := btcec.ParseDERSignature(sig, btcec.S256())
	if err != nil {
		return err
	}

	return nil
}

// Verify checks if the signature matches the payload and public key
func (sig Signature) Verify(payloadHash Hash, pk PublicKey) bool {
	s, err := btcec.ParseDERSignature(sig, btcec.S256())
	if err != nil {
		return false
	}

	parsedKey, err := btcec.ParsePubKey(pk, btcec.S256())
	if err != nil {
		return false
	}

	return s.Verify(payloadHash[:], parsedKey)
}

// String returns the hex-encoding of signature
func (sig Signature) String() string {
	return hex.EncodeToString(sig)
}

// PublicKey is an alias for public key in raw bytes
type PublicKey []byte

// ValidateBasic returns an error if the given public key is invalid; nil otherwise
func (pk PublicKey) ValidateBasic() error {
	if _, err := btcec.ParsePubKey(pk, btcec.S256()); err != nil {
		return err
	}

	return nil
}

// String returns the hex encoding of the given public key
func (pk PublicKey) String() string {
	return hex.EncodeToString(pk)
}
