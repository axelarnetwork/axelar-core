package exported

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/utils/funcs"
)

//go:generate moq -out ./mock/types.go -pkg mock . SigHandler Key MultiSig

// Key provides an interface to work with the key
type Key interface {
	GetParticipants() []sdk.ValAddress
	GetPubKey(sdk.ValAddress) (PublicKey, bool)
	GetWeight(sdk.ValAddress) sdk.Uint
	GetMinPassingWeight() sdk.Uint
	GetState() KeyState
	GetHeight() int64
	GetTimestamp() time.Time
	GetBondedWeight() sdk.Uint
	GetSnapshot() exported.Snapshot
}

// MultiSig provides an interface to work with the multi sig
type MultiSig interface {
	GetSignature(p sdk.ValAddress) (btcec.Signature, bool)
	GetPayloadHash() Hash
	GetKeyID() KeyID
	ValidateBasic() error
}

// SigHandler defines the interface for the requesting module to implement in
// order to handle the different results of signing session
type SigHandler interface {
	HandleCompleted(ctx sdk.Context, sig utils.ValidatedProtoMarshaler, moduleMetadata codec.ProtoMarshaler) error
	HandleFailed(ctx sdk.Context, moduleMetadata codec.ProtoMarshaler) error
}

// key id length range bounds dictated by tofnd
const (
	KeyIDLengthMin = 4
	KeyIDLengthMax = 256
)

// KeyID ensures a correctly formatted key ID
type KeyID string

// ValidateBasic returns an error if the given key ID is invalid; nil otherwise
func (id KeyID) ValidateBasic() error {
	if err := utils.ValidateString(string(id)); err != nil {
		return sdkerrors.Wrap(err, "invalid key id")
	}

	if len(id) < KeyIDLengthMin || len(id) > KeyIDLengthMax {
		return fmt.Errorf("key id length %d not in range [%d,%d]", len(id), KeyIDLengthMin, KeyIDLengthMax)
	}

	return nil
}

func (id KeyID) String() string {
	return string(id)
}

// PublicKey is an alias for compressed public key in raw bytes
type PublicKey []byte

// ValidateBasic returns an error if the given public key is invalid; nil otherwise
func (pk PublicKey) ValidateBasic() error {
	btcecPubKey, err := btcec.ParsePubKey(pk, btcec.S256())
	if err != nil {
		return err
	}

	if !bytes.Equal(pk, btcecPubKey.SerializeCompressed()) {
		return fmt.Errorf("public key is not compressed")
	}

	return nil
}

// String returns the hex encoding of the given public key
func (pk PublicKey) String() string {
	return hex.EncodeToString(pk)
}

// ToECDSAPubKey returns the ECDSA public key
func (pk PublicKey) ToECDSAPubKey() ecdsa.PublicKey {
	btcecKey := funcs.Must(btcec.ParsePubKey(pk, btcec.S256()))

	return *btcecKey.ToECDSA()
}

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
