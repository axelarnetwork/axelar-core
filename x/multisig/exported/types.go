package exported

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
)

// Sig defines the interface to work with the generated signature
type Sig interface {
	GetID() uint64
}

// SigHandler defines the interface for the requesting module to implement in
// order to handle the different results of signing session
type SigHandler interface {
	HandleCompleted(ctx sdk.Context, sig Sig, moduleMetadata codec.ProtoMarshaler) error
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

// Key is an interface for the key
type Key interface{}
