package types

import (
	stExported "github.com/axelarnetwork/axelar-core/x/staking/exported"
)

type Staker interface {
	stExported.Staker
}
