package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
)

type Voter interface {
	SetFutureVote(ctx sdk.Context, vote exported.FutureVote)
}
