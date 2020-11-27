package types

import (
	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	staking "github.com/axelarnetwork/axelar-core/x/staking/exported"
)

type Broadcaster interface {
	broadcast.Broadcaster
}

type Staker interface {
	staking.Staker
}
