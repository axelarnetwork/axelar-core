package types

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetAccountAddress returns the account address and name from a keys.Keybase for a given Bech32 encoded address or account moniker.
// Returns an error if the account is unknown.
func GetAccountAddress(from string, keybase keys.Keybase) (sdk.AccAddress, string, error) {
	var info keys.Info
	addr, err := sdk.AccAddressFromBech32(from)
	switch err {
	// string represents a Bech32 encoded address
	case nil:
		info, err = keybase.GetByAddress(addr)
		if err != nil {
			return nil, "", err
		}
		return info.GetAddress(), info.GetName(), nil
	// string represents an account moniker
	default:
		info, err = keybase.Get(from)
		if err != nil {
			return nil, "", err
		}
		return info.GetAddress(), info.GetName(), nil
	}
}
