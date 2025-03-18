package testutils

import (
	"cosmossdk.io/log"
	store "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
)

// NewSubspace returns a new subspace with a random name
func NewSubspace(cfg params.EncodingConfig) paramstypes.Subspace {
	return paramstypes.NewSubspace(cfg.Codec, cfg.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), rand.Str(10))
}

// NewContext returns a basic context with a fake store and a test logger
func NewContext(t log.TestingT) sdk.Context {
	return sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.NewTestLogger(t))
}
