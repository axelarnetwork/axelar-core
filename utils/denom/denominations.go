package denom

import (
	"fmt"

	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Known denominations (and abbreviations)
const (
	Sat     = "sat"
	Satoshi = "satoshi"
	Btc     = "btc"
	Bitcoin = "bitcoin"
	Wei     = "wei"
)

// ParseSatoshi parses a string to Satoshi, returning errors if invalid. Inputs in Bitcoin are automatically converted.
// This returns an error on an empty string as well.
func ParseSatoshi(rawCoin string) (sdk.Coin, error) {
	var coin sdk.DecCoin

	if intCoin, err := sdk.ParseCoin(rawCoin); err == nil {
		coin = sdk.NewDecCoinFromCoin(intCoin)
	} else {
		coin, err = sdk.ParseDecCoin(rawCoin)
		if err != nil {
			return sdk.Coin{}, fmt.Errorf("could not parse coin string")
		}
	}

	switch coin.Denom {
	case Sat, Satoshi:
	case Btc, Bitcoin:
		coin = sdk.NewDecCoinFromDec(Sat, coin.Amount.MulInt64(btcutil.SatoshiPerBitcoin))
	default:
		return sdk.Coin{}, fmt.Errorf("choose a correct denomination: %s (%s), %s (%s)", Satoshi, Sat, Bitcoin, Btc)
	}

	sat, remainder := coin.TruncateDecimal()
	if !remainder.Amount.IsZero() {
		return sdk.Coin{}, fmt.Errorf("amount in satoshi must be an integer value")
	}
	return sat, nil
}
