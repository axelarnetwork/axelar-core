package keeper

import (
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

var _ types.QueryServer = Keeper{}
