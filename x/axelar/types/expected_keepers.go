package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"

	bcExported "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
)

type Broadcaster interface {
	bcExported.Broadcaster
}

type Staker interface {
	GetLastTotalPower(ctx sdk.Context) (power sdk.Int)
	Validator(ctx sdk.Context, address sdk.ValAddress) exported.ValidatorI
}
