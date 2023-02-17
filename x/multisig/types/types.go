package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/btcsuite/btcd/btcec/v2"
	ec "github.com/btcsuite/btcd/btcec/v2/ecdsa"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

// Signature is an alias for signature in raw bytes
type Signature []byte

// ValidateBasic returns an error if the signature is not a valid S256 elliptic curve signature
func (sig Signature) ValidateBasic() error {
	_, err := ec.ParseDERSignature(sig)
	if err != nil {
		return err
	}

	return nil
}

// Verify checks if the signature matches the payload and public key
func (sig Signature) Verify(payloadHash exported.Hash, pk exported.PublicKey) bool {
	s, err := ec.ParseDERSignature(sig)
	if err != nil {
		return false
	}

	parsedKey, err := btcec.ParsePubKey(pk)
	if err != nil {
		return false
	}

	return s.Verify(payloadHash[:], parsedKey)
}

// String returns the hex-encoding of signature
func (sig Signature) String() string {
	return hex.EncodeToString(sig)
}

func (sig Signature) toECDSASignature() ec.Signature {
	return *funcs.Must(ec.ParseDERSignature(sig))
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

// ValidateBasic returns an error if the key epoch is invalid
func (m KeyEpoch) ValidateBasic() error {
	if m.Epoch == 0 {
		return fmt.Errorf("epoch must be >0")
	}

	if err := m.Chain.Validate(); err != nil {
		return err
	}

	if err := m.KeyID.ValidateBasic(); err != nil {
		return err
	}

	return nil
}
