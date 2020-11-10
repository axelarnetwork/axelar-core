package cli

import (
	"math/big"
	"testing"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/binance-chain/tss-lib/crypto"
	"github.com/binance-chain/tss-lib/tss"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
)

func TestGetCmdGetKey_UnmarshalKey(t *testing.T) {

	// initialize point
	x, _ := new(big.Int).SetString("35352920749338578287595792479852147230262762809356716696099786166951567525204", 10)
	y, _ := new(big.Int).SetString("7746337099917598860765814626382730010431714461234379120852704808357626213942", 10)
	c := tss.EC()
	point, err := crypto.NewECPoint(c, x, y)
	if err != nil {
		t.Fatalf("failed to initialize ECPoint [%v]", err)
	}

	// initialize codec
	// cdc := codec.New()
	cliCtx := context.NewCLIContext().WithCodec(types.ModuleCdc)

	// test: does it marshal/unmarshal?
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, point)
	if err != nil {
		t.Fatalf("MarshalJSONIndent error [%v]", err)
	}
	types.ModuleCdc.MustUnmarshalJSON(bz, &point)

	// TEST: does it print?
	t.Logf("PrintOutput format [%v]", cliCtx.OutputFormat)
	if err := cliCtx.PrintOutput(point); err != nil {
		t.Fatalf("PrintOutput error [%v]", err)
	}

	cliCtx.OutputFormat = "text"
	t.Logf("PrintOutput format [%v]", cliCtx.OutputFormat)
	if err := cliCtx.PrintOutput(point); err != nil {
		t.Fatalf("PrintOutput error [%v]", err)
	}

	cliCtx.OutputFormat = "json"
	t.Logf("PrintOutput format [%v]", cliCtx.OutputFormat)
	if err := cliCtx.PrintOutput(point); err != nil {
		t.Fatalf("PrintOutput error [%v]", err)
	}
}
