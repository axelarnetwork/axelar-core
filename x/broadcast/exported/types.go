package exported

import sdk "github.com/cosmos/cosmos-sdk/types"

// This interface is exposed for convenience, otherwise all other modules would have to reimplement it.
// Recommended pattern: In other modules, define a keeper interface in the respective expected_keepers.go file and
// embed this interface into it
type Broadcaster interface {
	// BroadcastSync sends the passed messages synchronously to the network.
	// Do not call it from the main thread or risk a deadlock (the main thread is needed to validate incoming messages)
	BroadcastSync(ctx sdk.Context, msgs []MsgWithSenderSetter) error

	// RegisterProxy registers a proxy address for a given principal, which can broadcast messages in the principal's name
	RegisterProxy(ctx sdk.Context, principal sdk.ValAddress, proxy sdk.AccAddress) error

	// GetPrincipal returns the principal address for a given proxy address. Returns nil if not set.
	GetPrincipal(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress

	// GetProxy returns the proxy address for a given principal address. Returns nil if not set.
	GetProxy(ctx sdk.Context, principal sdk.ValAddress) sdk.AccAddress

	// GetLocalPrincipal returns the address of the local validator account. Returns nil if not set.
	//
	// WARNING: Handle with care, this call is non-deterministic because it exposes local information that is DIFFERENT for each validator
	GetLocalPrincipal(ctx sdk.Context) sdk.ValAddress
}

type MsgWithSenderSetter interface {
	sdk.Msg
	SetSender(address sdk.AccAddress)
}
