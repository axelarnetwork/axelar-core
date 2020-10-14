package exported

import sdk "github.com/cosmos/cosmos-sdk/types"

type Broadcaster interface {
	Broadcast(ctx sdk.Context, msgs []ValidatorMsg) error
}

type ValidatorMsg interface {
	sdk.Msg
	SetSender(address sdk.AccAddress)
}
