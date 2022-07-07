package exported

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
)

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
