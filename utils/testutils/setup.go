package testutils

import (
	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"
	abci "github.com/tendermint/tendermint/proto/tendermint/types"
)

func NewSubspace(cfg params.EncodingConfig) paramstypes.Subspace {
	return paramstypes.NewSubspace(cfg.Codec, cfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), rand.Str(10))
}

func NewContext() sdk.Context {
	return sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
}
