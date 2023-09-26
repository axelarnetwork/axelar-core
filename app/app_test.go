package app_test

import (
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/axelarnetwork/axelar-core/app"
)

func TestNewAxelarApp(t *testing.T) {
	version.Version = "0.27.0"

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
}
