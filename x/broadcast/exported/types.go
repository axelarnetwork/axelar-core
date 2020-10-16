package exported

import sdk "github.com/cosmos/cosmos-sdk/types"

type Broadcaster interface {
	Broadcast(ctx sdk.Context, msgs []ValidatorMsg) error
	RegisterProxy(ctx sdk.Context, principal sdk.ValAddress, proxy sdk.AccAddress) error
	GetPrincipal(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress
	GetProxyCount(ctx sdk.Context) uint32

	// WARNING: Handle with care, this exposes local information that is DIFFERENT for each validator
	GetLocalPrincipal(ctx sdk.Context) sdk.ValAddress
}

type ValidatorMsg interface {
	sdk.Msg
	SetSender(address sdk.AccAddress)
}
