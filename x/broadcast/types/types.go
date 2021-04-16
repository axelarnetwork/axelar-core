package types

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetAccountAddress returns the account address and name from a keys.Keybase for a given Bech32 encoded address or account moniker.
// Returns an error if the account is unknown.
func GetAccountAddress(from string, kr keyring.Keyring) (sdk.AccAddress, string, error) {
	var info keyring.Info
	addr, err := sdk.AccAddressFromBech32(from)
	switch err {
	// string represents a Bech32 encoded address
	case nil:
		info, err = kr.KeyByAddress(addr)
		if err != nil {
			return nil, "", err
		}
		return info.GetAddress(), info.GetName(), nil
	// string represents an account moniker
	default:
		info, err = kr.Key(from)
		if err != nil {
			return nil, "", err
		}
		return info.GetAddress(), info.GetName(), nil
	}
}
