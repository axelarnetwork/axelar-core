package keeper_test

import (
	"encoding/base64"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/types"
	. "github.com/axelarnetwork/utils/test"
)

func TestAddr(t *testing.T) {
	addr := common.HexToAddress("0xF4DFa637c97a6991b32Fc72b6817c68b16ed04c3")
	t.Logf("%s as base64: %s", addr.Hex(), base64.StdEncoding.EncodeToString(addr.Bytes()))
}

func TestGenesis(t *testing.T) {
	var (
		initialState  types.GenesisState
		exportedState types.GenesisState
	)

	cfg := params.MakeEncodingConfig()

	paramsK := paramskeeper.NewKeeper(cfg.Codec, cfg.Amino, sdk.NewKVStoreKey(paramstypes.StoreKey), sdk.NewKVStoreKey(paramstypes.TStoreKey))
	k := keeper.NewKeeper(cfg.Codec, sdk.NewKVStoreKey(types.StoreKey), paramsK)

	Given("a genesis state", func() {
		cfg.InterfaceRegistry.RegisterImplementations((*codec.ProtoMarshaler)(nil), &multisig.MultiSig{})
		initialState = types.NewGenesisState(testutils.RandomChains(cfg.Codec))

	}).When("it is valid", func() {
		assert.NoError(t, initialState.Validate())
	}).
		When("importing and exporting the state", func() {
			ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
			k.InitGenesis(ctx, initialState)
			exportedState = k.ExportGenesis(ctx)
		}).
		Then("both states are equal", func(t *testing.T) {
			assertChainsEqual(t, initialState, exportedState)
		}).Run(t, 10)

	Given("the default genesis state", func() {
		initialState = types.DefaultGenesisState()
	}).When("it is valid", func() {
		assert.NoError(t, initialState.Validate())
	}).Then("the keeper can be initialized", func(t *testing.T) {
		ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		assert.NotPanics(t, func() { k.InitGenesis(ctx, initialState) })
	}).Run(t)

}

func assertChainsEqual(t *testing.T, initial types.GenesisState, exported types.GenesisState) {
	assert.Equal(t, len(initial.Chains), len(exported.Chains))

	for i := range initial.Chains {
		assert.Equal(t, initial.Chains[i].CommandQueue, exported.Chains[i].CommandQueue)
		assert.ElementsMatch(t, initial.Chains[i].BurnerInfos, exported.Chains[i].BurnerInfos)
		assert.ElementsMatch(t, initial.Chains[i].ConfirmedDeposits, exported.Chains[i].ConfirmedDeposits)
		assert.ElementsMatch(t, initial.Chains[i].BurnedDeposits, exported.Chains[i].BurnedDeposits)
		assert.ElementsMatch(t, initial.Chains[i].Tokens, exported.Chains[i].Tokens)
		assert.ElementsMatch(t, initial.Chains[i].CommandBatches, exported.Chains[i].CommandBatches)
		assert.Equal(t, initial.Chains[i].Gateway, exported.Chains[i].Gateway)
		assert.Equal(t, initial.Chains[i].Params, exported.Chains[i].Params)
		assert.ElementsMatch(t, initial.Chains[i].Events, exported.Chains[i].Events)
		assert.Equal(t, initial.Chains[i].ConfirmedEventQueue, exported.Chains[i].ConfirmedEventQueue)
	}
}
