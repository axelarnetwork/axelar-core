package app

import (
	"fmt"
	"os"
	"sort"
	"testing"

	"cosmossdk.io/log"
	"github.com/CosmWasm/wasmd/x/wasm"
	dbm "github.com/cosmos/cosmos-db"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	"github.com/axelarnetwork/utils/funcs"
)

// TestPrintVersionMap prints the consensus version of every module in the
// binary. Compare against the on-chain map from
// /cosmos/upgrade/v1beta1/module_versions to see which migrations run at the
// next upgrade height.
func TestPrintVersionMap(t *testing.T) {
	WasmEnabled, IBCWasmHooksEnabled = "true", "false"
	t.Cleanup(func() { funcs.MustNoErr(os.RemoveAll("wasm")) })

	axelarApp := NewAxelarApp(
		log.NewTestLogger(t),
		dbm.NewMemDB(),
		nil,
		true,
		MakeEncodingConfig(),
		simtestutil.EmptyAppOptions{},
		[]wasm.Option{},
	)

	vm := axelarApp.mm.GetVersionMap()
	names := make([]string, 0, len(vm))
	for name := range vm {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Printf("%s: %d\n", name, vm[name])
	}
}
