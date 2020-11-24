package types

import (
	bcExported "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	stExported "github.com/axelarnetwork/axelar-core/x/staking/exported"
)

type Broadcaster interface {
	bcExported.Broadcaster
}

type Staker interface {
	stExported.Staker
}
