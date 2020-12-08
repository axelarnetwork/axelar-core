package types

import (
	ssExported "github.com/axelarnetwork/axelar-core/x/snapshotting/exported"
)

type Snapshotter interface {
	ssExported.Snapshotter
}
