package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type Broadcaster interface {
	Broadcast(msgs ...sdk.Msg) error
}
