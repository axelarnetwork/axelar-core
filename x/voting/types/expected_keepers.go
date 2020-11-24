package types

import (
	bcExported "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
)

type Broadcaster interface {
	bcExported.Broadcaster
}
