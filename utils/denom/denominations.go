package denom

import (
	"fmt"

	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	Sat     = "sat"
	Satoshi = "satoshi"
	Btc     = "btc"
	Bitcoin = "bitcoin"
)

func ParseSatoshi(rawCoin string) (sdk.Coin, error) {
	var coin sdk.DecCoin
	coin, err := sdk.ParseDecCoin(rawCoin)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("could not parse coin string")
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
