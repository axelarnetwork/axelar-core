package distribution_test

import (
	"fmt"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/stretchr/testify/assert"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"runtime"
	"testing"

	"github.com/axelarnetwork/axelar-core/x/distribution"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
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
