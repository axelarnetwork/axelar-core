package app_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/stretchr/testify/assert"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	abci "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"google.golang.org/grpc/encoding"
	encproto "google.golang.org/grpc/encoding/proto"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

func TestNewAxelarApp(t *testing.T) {
	version.Version = "0.27.0"

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
			assert.NotPanics(t, func() {
				app.NewAxelarApp(
					log.TestingLogger(),
					dbm.NewMemDB(),
					nil,
					true,
					nil,
					"",
					"",
					0,
					app.MakeEncodingConfig(),
					simapp.EmptyAppOptions{},
					[]wasm.Option{},
				)
			})
		})
	}
}

func TestMaxWasmSizeOverride(t *testing.T) {
	version.Version = "0.27.0"

	testCases := []int{1, 100, 3 * 1024 * 1024}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("max wasm code size: %d", testCase), func(t *testing.T) {
			app.MaxWasmSize = fmt.Sprintf("%d", testCase)

			app.NewAxelarApp(
				log.TestingLogger(),
				dbm.NewMemDB(),
				nil,
				true,
				nil,
				"",
				"",
				0,
				app.MakeEncodingConfig(),
				simapp.EmptyAppOptions{},
				[]wasm.Option{},
			)

			assert.Equal(t, testCase, wasmtypes.MaxWasmSize)
		})
	}
}

// check that encoding is set so gogoproto extensions are supported
func TestGRPCEncodingSetDuringInit(t *testing.T) {
	// if the codec is set during the app's init() function, then this should return a codec that can encode gogoproto extensions
	codec := encoding.GetCodec(encproto.Name)

	keyResponse := multisig.KeyResponse{
		KeyID:              "keyID",
		State:              0,
		StartedAt:          0,
		StartedAtTimestamp: time.Now(),
		ThresholdWeight:    sdk.ZeroUint(),
		BondedWeight:       sdk.ZeroUint(),
		Participants: []multisig.KeygenParticipant{{
			Address: "participant",
			Weight:  sdk.OneUint(),
			PubKey:  "pubkey",
		}},
	}

	bz, err := codec.Marshal(&keyResponse)
	assert.NoError(t, err)
	assert.NoError(t, codec.Unmarshal(bz, &keyResponse))
}

func TestAnteHandlersCanHandleWasmMsgsWithoutSigners(t *testing.T) {
	app.SetConfig()
	app.WasmEnabled = "true"
	app.IBCWasmHooksEnabled = "true"
	version.Version = "0.35.0"
	encConfig := app.MakeEncodingConfig()

	tx := prepareTx(encConfig, &exported.WasmMessage{})
	anteHandler := prepareAnteHandler(encConfig)
	ctx := prepareCtx()

	_, err := anteHandler(ctx, tx, true)
	assert.NoError(t, err)
	_, err = anteHandler(ctx, tx, false)
	assert.NoError(t, err)
}

func prepareTx(encConfig params.EncodingConfig, msg sdk.Msg) sdk.Tx {
	sk, _, _ := testdata.KeyTestPubAddr()

	tx := funcs.Must(helpers.GenTx(
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

func prepareAnteHandler(cfg params.EncodingConfig) sdk.AnteHandler {
	axelarApp := app.NewAxelarApp(
		log.TestingLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		nil,
		"",
		"",
		0,
		cfg,
		simapp.EmptyAppOptions{},
		[]wasm.Option{},
	)

	anteHandler := app.InitCustomAnteDecorators(cfg, axelarApp.Keys, axelarApp.Keepers, simapp.EmptyAppOptions{})
	return sdk.ChainAnteDecorators(anteHandler...)
}

func prepareCtx() sdk.Context {
	return sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger()).
		WithConsensusParams(&abcitypes.ConsensusParams{
			Block: &abcitypes.BlockParams{MaxGas: 1000000000},
		})
}
