package types

import (
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Tss interface {
	GetCurrentMasterKeyID(ctx sdk.Context, chain nexus.Chain) (string, bool)
	GetNextMasterKeyID(ctx sdk.Context, chain nexus.Chain) (string, bool)
	GetSnapshotCounterForKeyID(ctx sdk.Context, keyID string) (int64, bool)
}

type Nexus interface {
	GetChains(ctx sdk.Context) []nexus.Chain
}

type Snapshotter interface {
	GetSnapshot(ctx sdk.Context, counter int64) (snapshot.Snapshot, bool)
}
