package app

import (
	"encoding/json"
	"os"
	"testing"

	"cosmossdk.io/log"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	permissionexported "github.com/axelarnetwork/axelar-core/x/permission/exported"
	permissionKeeper "github.com/axelarnetwork/axelar-core/x/permission/keeper"
	permissiontypes "github.com/axelarnetwork/axelar-core/x/permission/types"
	"github.com/axelarnetwork/utils/funcs"
)

// messages dispatched by wasm contracts never pass through the tx ante handler, and nexus custom messages
// never reach the msg service router, so the emergency pause must be enforced by the wasm messenger
func TestEmergencyPause_WasmDispatchedMessagesAreRejected(t *testing.T) {
	WasmEnabled, IBCWasmHooksEnabled = "true", "false"
	t.Cleanup(func() { funcs.MustNoErr(os.RemoveAll("wasm")) })

	encCfg := MakeEncodingConfig()
	axelarApp := NewAxelarApp(log.NewTestLogger(t), dbm.NewMemDB(), nil, true, encCfg, simtestutil.EmptyAppOptions{}, []wasm.Option{})
	ctx := axelarApp.NewUncachedContext(false, cmtproto.Header{})

	permissionK := GetKeeper[permissionKeeper.Keeper](axelarApp.Keepers)
	permissionK.InitGenesis(ctx, &permissiontypes.GenesisState{PauseLevel: permissionexported.PAUSE_LEVEL_ALL})

	encoders := wasm.DefaultEncoders(encCfg.Codec, nil)
	encoders.Custom = nexusKeeper.EncodeRoutingMessage

	var dispatched int
	messenger := WithAnteHandlers(encoders, initMessageAnteDecorators(encCfg, axelarApp.Keepers),
		wasmkeeper.MessageHandlerFunc(func(sdk.Context, sdk.AccAddress, string, wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
			dispatched++
			return nil, nil, nil, nil
		}))

	contract := rand.AccAddr()
	bankSend := wasmvmtypes.CosmosMsg{Bank: &wasmvmtypes.BankMsg{Send: &wasmvmtypes.SendMsg{
		ToAddress: rand.AccAddr().String(),
		Amount:    wasmvmtypes.Array[wasmvmtypes.Coin]{{Denom: "uaxl", Amount: "1"}},
	}}}
	routing := wasmvmtypes.CosmosMsg{Custom: funcs.Must(json.Marshal(exported.WasmMessage{}))}

	for _, msg := range []wasmvmtypes.CosmosMsg{bankSend, routing} {
		_, _, _, err := messenger.DispatchMsg(ctx, contract, "", msg)
		assert.ErrorIs(t, err, permissiontypes.ErrChainPaused)
	}
	assert.Zero(t, dispatched)

	permissionK.InitGenesis(ctx, &permissiontypes.GenesisState{PauseLevel: permissionexported.PAUSE_LEVEL_UNSPECIFIED})
	_, _, _, err := messenger.DispatchMsg(ctx, contract, "", bankSend)
	require.NoError(t, err)
	assert.Equal(t, 1, dispatched)
}
