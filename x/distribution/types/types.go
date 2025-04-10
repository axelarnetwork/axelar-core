package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const feeBurnedPrefix = "burned"

var ZeroAddress = sdk.AccAddress(make([]byte, 32))

// WithBurnedPrefix converts a coin to a coin with 'burned' prefix
func WithBurnedPrefix(coin sdk.Coin) sdk.Coin {
	return sdk.NewCoin(fmt.Sprintf("%s-%s", feeBurnedPrefix, coin.Denom), coin.Amount)
}
