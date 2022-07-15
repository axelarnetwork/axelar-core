package exported

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
)

//go:generate moq -out ./mock/types.go -pkg mock . SigHandler

// Key provides an interface to work with the key
type Key interface {
	GetParticipants() []sdk.ValAddress
	GetPubKey(sdk.ValAddress) (PublicKey, bool)
	GetWeight(sdk.ValAddress) sdk.Uint
	GetMinPassingWeight() sdk.Uint
}

// SigHandler defines the interface for the requesting module to implement in
// order to handle the different results of signing session
type SigHandler interface {
	HandleCompleted(ctx sdk.Context, sig codec.ProtoMarshaler, moduleMetadata codec.ProtoMarshaler) error
	HandleFailed(ctx sdk.Context, moduleMetadata codec.ProtoMarshaler) error
}

// key id length range bounds dictated by tofnd
const (
	KeyIDLengthMin = 4
	KeyIDLengthMax = 256
)

// KeyID ensures a correctly formatted tss key ID
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

// GetECDSAPubKey returns the ECDSA public key
func (pk PublicKey) GetECDSAPubKey() (ecdsa.PublicKey, error) {
	btcecKey, err := btcec.ParsePubKey(pk, btcec.S256())
	if err != nil {
		return ecdsa.PublicKey{}, err
	}

	return *btcecKey.ToECDSA(), nil
}
