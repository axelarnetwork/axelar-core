package distribution_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"

	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/distribution"
	"github.com/axelarnetwork/utils/funcs"
)

// TestEnsureIdentical ensure that our BeginBlocker implementation exactly matches the cosmos-sdk version.
func TestEnsureIdentical(t *testing.T) {
	customFn := runtime.FuncForPC(reflect.ValueOf(distribution.BeginBlocker).Pointer())
	sdkFn := runtime.FuncForPC(reflect.ValueOf(distr.BeginBlocker).Pointer())

	localFile, _ := customFn.FileLine(customFn.Entry())
	sdkFile, _ := sdkFn.FileLine(sdkFn.Entry())

	customContent := funcs.Must(os.ReadFile(localFile))
	sdkContent := funcs.Must(os.ReadFile(sdkFile))

	funcName := "BeginBlocker"
	customImpl := extractFunctionBody(string(customContent), funcName)
	sdkImpl := extractFunctionBody(string(sdkContent), funcName)

	assert.Equal(t, customImpl, sdkImpl, fmt.Sprintf("%s implementation differs from SDK, please update abci.go to match the SDK version", funcName))
}

// TestSDKKeeperBeginBlockerUnchanged pins the body of the SDK distribution
// keeper's BeginBlocker. Our keeper mirrors it (minus the external community
// pool branch, which is not wired) in x/distribution/keeper.Keeper.BeginBlocker.
// If this test fails after an SDK bump, re-derive that mirror from the new SDK
// source and update the pinned body below.
func TestSDKKeeperBeginBlockerUnchanged(t *testing.T) {
	sdkFn := runtime.FuncForPC(reflect.ValueOf(distrkeeper.Keeper.BeginBlocker).Pointer())
	sdkFile, _ := sdkFn.FileLine(sdkFn.Entry())
	sdkImpl := strings.TrimSpace(extractFunctionBody(string(funcs.Must(os.ReadFile(sdkFile))), "BeginBlocker"))

	expected := `{
	start := telemetry.Now()
	defer telemetry.ModuleMeasureSince(types.ModuleName, start, telemetry.MetricKeyBeginBlocker)

	// determine the total power signing the block
	var previousTotalPower int64
	// determine the total power signing the block
	for _, voteInfo := range ctx.VoteInfos() {
		previousTotalPower += voteInfo.Validator.Power
	}

	// TODO this is Tendermint-dependent
	// ref https://github.com/cosmos/cosmos-sdk/issues/3095
	height := ctx.BlockHeight()
	if height > 1 {
		if err := k.AllocateTokens(ctx, previousTotalPower, ctx.VoteInfos()); err != nil {
			return err
		}

		// send whole coins from community pool to x/protocolpool if enabled
		if k.HasExternalCommunityPool() {
			if err := k.sendCommunityPoolToExternalPool(ctx); err != nil {
				return err
			}
		}
	}

	// record the proposer for when we pay out on the next block
	consAddr := sdk.ConsAddress(ctx.BlockHeader().ProposerAddress)
	return k.SetPreviousProposerConsAddr(ctx, consAddr)
}`

	assert.Equal(t, expected, sdkImpl,
		"the SDK distribution keeper BeginBlocker changed; re-derive x/distribution/keeper.Keeper.BeginBlocker to match")
}

func extractFunctionBody(content, funcName string) string {
	fset := token.NewFileSet()
	file := funcs.Must(parser.ParseFile(fset, "", content, parser.ParseComments))

	var functionBody string
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == funcName {
			functionBody = content[fn.Body.Pos()-1 : fn.Body.End()]
			return false
		}
		return true
	})
	return functionBody
}
