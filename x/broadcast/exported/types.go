package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//go:generate moq -out ./mock/types.go -pkg mock . Broadcaster

// Broadcaster -
// This interface is exposed for convenience, otherwise all other modules would have to reimplement it.
// Recommended pattern: In other modules, define a keeper interface in the respective expected_keepers.go file and
// embed this interface into it
type Broadcaster interface {
	// RegisterProxy registers a proxy address for a given principal, which can broadcast messages in the principal's name
	RegisterProxy(ctx sdk.Context, principal sdk.ValAddress, proxy sdk.AccAddress) error

	// GetPrincipal returns the principal address for a given proxy address. Returns nil if not set.
	GetPrincipal(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress

	// GetProxy returns the proxy address for a given principal address. Returns nil if not set.
	GetProxy(ctx sdk.Context, principal sdk.ValAddress) sdk.AccAddress
}
