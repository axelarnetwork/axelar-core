package types

import (
	"fmt"

	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	sat      = "sat"
	satoshi  = "satoshi"
	btc      = "btc"
	bitcoin  = "bitcoin"
	satToBtc = 100_000_000
)

func ParseBtc(rawCoin string) (btcutil.Amount, error) {
	var coin sdk.DecCoin
	coin, err := sdk.ParseDecCoin(rawCoin)
	if err != nil {
		return 0, fmt.Errorf("could not parse coin string")
	}

	switch coin.Denom {
	case sat, satoshi:
		if !coin.Amount.IsInteger() {
			return 0, fmt.Errorf("satoshi must be an integer value")
		}
		return btcutil.Amount(coin.Amount.Int64()), nil
	case btc, bitcoin:
		// sdk.Coin does not reduce precision, even if all decimal places are 0,
		// so need to call RoundInt64 to return the correct value
		return btcutil.Amount(coin.Amount.MulInt64(satToBtc).RoundInt64()), nil
	default:
		return 0, fmt.Errorf("choose a correct denomination: %s (%s), %s (%s)", satoshi, sat, bitcoin, btc)
	}
}
