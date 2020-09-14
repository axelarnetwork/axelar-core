package axelar

import (
	"github.com/axelarnetwork/axelar-net/x/axelar/internal/keeper"
	"github.com/axelarnetwork/axelar-net/x/axelar/internal/types"
)

const (
	ModuleName   = types.ModuleName
	QuerierRoute = types.QuerierRoute
	RouterKey    = types.RouterKey
	StoreKey     = types.StoreKey
)

type (
	Keeper          = keeper.Keeper
	MsgTrackAddress = types.MsgTrackAddress
)

var (
	NewKeeper = keeper.NewKeeper
)
