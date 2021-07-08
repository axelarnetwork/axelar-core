package cli

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
)

func TestEthToWei_IsInteger(t *testing.T) {
	amount, _ := sdk.NewDecFromStr("3.2")
	eth := sdk.DecCoin{
		Denom:  "eth",
		Amount: amount,
	}
	wei := eth
	wei.Amount = eth.Amount.MulInt64(params.Ether)

	assert.True(t, wei.Amount.IsInteger())
}

func TestGweiToWei_IsNotInteger(t *testing.T) {
	amount, _ := sdk.NewDecFromStr("3.0000000000002")
	gwei := sdk.DecCoin{
		Denom:  "gwei",
		Amount: amount,
	}
	wei := gwei
	wei.Amount = gwei.Amount.MulInt64(params.GWei)

	assert.False(t, wei.Amount.IsInteger())
}
