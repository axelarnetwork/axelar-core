package types

import (
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
)

type Snapshotter interface {
	snapshot.Snapshotter
}
