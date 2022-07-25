package types

import (
	"bytes"
	"encoding/hex"
	"sort"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

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
func (sig Signature) Verify(payloadHash exported.Hash, pk exported.PublicKey) bool {
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

func (sig Signature) toECDSASignature() btcec.Signature {
	return *funcs.Must(btcec.ParseDERSignature(sig, btcec.S256()))
}

func sortAddresses[T sdk.Address](addrs []T) []T {
	sorted := make([]T, len(addrs))
	copy(sorted, addrs)

	sort.SliceStable(sorted, func(i, j int) bool { return bytes.Compare(sorted[i].Bytes(), sorted[j].Bytes()) < 0 })

	return sorted
}

// NewKeyEpoch is the constructor for key rotation
func NewKeyEpoch(epoch uint64, chain nexus.ChainName, keyID exported.KeyID) KeyEpoch {
	return KeyEpoch{
		Epoch: epoch,
		Chain: chain,
		KeyID: keyID,
	}
}
