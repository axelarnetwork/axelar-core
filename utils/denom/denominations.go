package denom

import (
	"fmt"

	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Sat is the allowed abbreviation for the Satoshi denomination label
	Sat = "sat"
	// Satoshi denomination label
	Satoshi = "satoshi"
	// Btc is the allowed abbreviation for the Bitcoin denomination label
	Btc = "btc"
	// Bitcoin denomination label
	Bitcoin = "bitcoin"
)

// ParseSatoshi parses a string to Satoshi, returning errors if invalid. Inputs in Bitcoin are automatically converted.
// This returns an error on an empty string as well.
func ParseSatoshi(rawCoin string) (sdk.Coin, error) {
	var coin sdk.DecCoin

	intCoin, err := sdk.ParseCoin(rawCoin)
	if err != nil {
		coin, err = sdk.ParseDecCoin(rawCoin)
		if err != nil {
			return sdk.Coin{}, fmt.Errorf("could not parse coin string")
		}
	}
	coin = sdk.NewDecCoinFromCoin(intCoin)

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
