package app_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"cosmossdk.io/log"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	abci "github.com/cometbft/cometbft/proto/tendermint/types"
	abcitypes "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

func init() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(app.AccountAddressPrefix, app.AccountPubKeyPrefix)
	config.SetBech32PrefixForValidator(app.ValidatorAddressPrefix, app.ValidatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(app.ConsNodeAddressPrefix, app.ConsNodePubKeyPrefix)

	version.Version = "v1.3.0"
}

func TestNewAxelarApp(t *testing.T) {
	testCases := []struct {
		wasm  string
		hooks string
	}{
		{"false", "false"},
		{"true", "false"},
		{"true", "true"}}

	for _, testCase := range testCases {
		app.WasmEnabled, app.IBCWasmHooksEnabled = testCase.wasm, testCase.hooks

		t.Run("wasm_enabled:"+testCase.wasm+"-hooks_enabled:"+testCase.hooks, func(t *testing.T) {
			t.Cleanup(cleanup)

			assert.NotPanics(t, func() {
				app.NewAxelarApp(
					log.NewTestLogger(t),
					dbm.NewMemDB(),
					nil,
					true,
					app.MakeEncodingConfig(),
					simtestutil.EmptyAppOptions{},
					[]wasm.Option{},
				)
			})
		})
	}
}

func TestMaxWasmSizeOverride(t *testing.T) {
	app.WasmEnabled = "true"

	testCases := []int{1, 100, 3 * 1024 * 1024}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("max wasm code size: %d", testCase), func(t *testing.T) {
			app.MaxWasmSize = fmt.Sprintf("%d", testCase)
			t.Cleanup(cleanup)

			app.NewAxelarApp(
				log.NewTestLogger(t),
				dbm.NewMemDB(),
				nil,
				true,
				app.MakeEncodingConfig(),
				simtestutil.EmptyAppOptions{},
				[]wasm.Option{},
			)

			assert.Equal(t, testCase, wasmtypes.MaxWasmSize)
		})
	}
}

func TestAnteHandlersCanHandleWasmMsgsWithoutSigners(t *testing.T) {
	app.WasmEnabled = "true"
	app.IBCWasmHooksEnabled = "true"

	encConfig := app.MakeEncodingConfig()

	tx := prepareTx(encConfig, &exported.WasmMessage{})
	anteHandler := prepareAnteHandler(encConfig, t)
	ctx := prepareCtx(t)

	_, err := anteHandler(ctx, tx, true)
	assert.NoError(t, err)
	_, err = anteHandler(ctx, tx, false)
	assert.NoError(t, err)
}

func prepareTx(encConfig params.EncodingConfig, msg sdk.Msg) sdk.Tx {
	sk, _, _ := testdata.KeyTestPubAddr()

	tx := funcs.Must(simtestutil.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		encConfig.TxConfig,
		[]sdk.Msg{msg},
		sdk.NewCoins(sdk.NewInt64Coin("testcoin", 0)),
		1000000000,
		"testchain",
		[]uint64{0},
		[]uint64{0},
		sk,
	))
	return tx
}

func prepareAnteHandler(cfg params.EncodingConfig, t log.TestingT) sdk.AnteHandler {
	axelarApp := app.NewAxelarApp(
		log.NewTestLogger(t),
		dbm.NewMemDB(),
		nil,
		true,
		cfg,
		simtestutil.EmptyAppOptions{},
		[]wasm.Option{},
	)

	anteHandler := app.InitCustomAnteDecorators(cfg, axelarApp.Keys, axelarApp.Keepers, simtestutil.EmptyAppOptions{})
	return sdk.ChainAnteDecorators(anteHandler...)
}

func prepareCtx(t log.TestingT) sdk.Context {
	return sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.NewTestLogger(t)).
		WithConsensusParams(abcitypes.ConsensusParams{
			Block: &abcitypes.BlockParams{MaxGas: 1000000000},
		})
}

func cleanup() {
	funcs.MustNoErr(os.RemoveAll("wasm"))
}
