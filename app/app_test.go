package app_test

import (
	"testing"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	"google.golang.org/grpc/encoding"
	encproto "google.golang.org/grpc/encoding/proto"

	"github.com/axelarnetwork/axelar-core/app"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/types"
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
					0,
					app.MakeEncodingConfig(),
					simapp.EmptyAppOptions{},
					[]wasm.Option{},
				)
			})
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
