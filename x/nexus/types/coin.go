package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"

	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

// Coin provides functionality to lock and release coins
type Coin struct {
	sdk.Coin
	coinType CoinType
	nexus    Nexus
	ibc      IBCKeeper
	bank     BankKeeper
}

// NewCoin creates a coin struct, assign a coin type and normalize the denom if it's a ICS20 token
func NewCoin(ctx sdk.Context, nexus Nexus, ibc IBCKeeper, bank BankKeeper, coin sdk.Coin) (Coin, error) {
	coinType, err := getCoinType(ctx, nexus, coin.GetDenom())
	if err != nil {
		return Coin{}, err
	}

	// If coin type is ICS20, we need to normalize it to convert from 'ibc/{hash}'
	// to native asset denom so that nexus could recognize it
	if coinType == ICS20 {
		denomTrace, err := ibc.ParseIBCDenom(ctx, coin.GetDenom())
		if err != nil {
			return Coin{}, err
		}

		coin = sdk.NewCoin(denomTrace.GetBaseDenom(), coin.Amount)
	}

	c := Coin{
		Coin:     coin,
		coinType: coinType,
		nexus:    nexus,
		ibc:      ibc,
		bank:     bank,
	}
	if _, err := c.getOriginalCoin(ctx); err != nil {
		return Coin{}, err
	}

	return c, nil
}

// GetOriginalCoin returns the original coin
func (c Coin) GetOriginalCoin(ctx sdk.Context) sdk.Coin {
	// NOTE: must not fail since it's already checked in NewCoin
	return funcs.Must(c.getOriginalCoin(ctx))
}

// Lock locks the given coin from the given address
func (c Coin) Lock(ctx sdk.Context, fromAddr sdk.AccAddress) error {
	coin := c.GetOriginalCoin(ctx)

	switch c.coinType {
	case ICS20, Native:
		return lock(ctx, c.bank, fromAddr, coin)
	case External:
		return burn(ctx, c.bank, fromAddr, coin)
	default:
		return fmt.Errorf("unrecognized coin type %d", c.coinType)
	}
}

// Unlock unlocks the given coin to the given address
func (c Coin) Unlock(ctx sdk.Context, toAddr sdk.AccAddress) error {
	coin := c.GetOriginalCoin(ctx)

	switch c.coinType {
	case ICS20, Native:
		return unlock(ctx, c.bank, toAddr, coin)
	case External:
		return mint(ctx, c.bank, toAddr, coin)
	default:
		return fmt.Errorf("unrecognized coin type %d", c.coinType)
	}
}

func (c Coin) getOriginalCoin(ctx sdk.Context) (sdk.Coin, error) {
	switch c.coinType {
	case ICS20:
		return c.toICS20(ctx)
	case Native, External:
		return c.Coin, nil
	default:
		return sdk.Coin{}, fmt.Errorf("unrecognized coin type %d", c.coinType)
	}
}

func (c Coin) toICS20(ctx sdk.Context) (sdk.Coin, error) {
	if c.coinType != ICS20 {
		return sdk.Coin{}, fmt.Errorf("%s is not ICS20 token", c.GetDenom())
	}

	// check if the asset registered with a path
	chain, ok := c.nexus.GetChainByNativeAsset(ctx, c.GetDenom())
	if !ok {
		return sdk.Coin{}, fmt.Errorf("asset %s is not linked to a cosmos chain", c.GetDenom())
	}

	path, ok := c.ibc.GetIBCPath(ctx, chain.Name)
	if !ok {
		return sdk.Coin{}, fmt.Errorf("path not found for chain %s", chain.Name)
	}

	trace := ibctypes.DenomTrace{
		Path:      path,
		BaseDenom: c.GetDenom(),
	}

	return sdk.NewCoin(trace.IBCDenom(), c.Amount), nil
}

func lock(ctx sdk.Context, bank BankKeeper, fromAddr sdk.AccAddress, coin sdk.Coin) error {
	return bank.SendCoins(ctx, fromAddr, exported.GetEscrowAddress(coin.GetDenom()), sdk.NewCoins(coin))
}

func unlock(ctx sdk.Context, bank BankKeeper, toAddr sdk.AccAddress, coin sdk.Coin) error {
	return bank.SendCoins(ctx, exported.GetEscrowAddress(coin.GetDenom()), toAddr, sdk.NewCoins(coin))
}

func burn(ctx sdk.Context, bank BankKeeper, fromAddr sdk.AccAddress, coin sdk.Coin) error {
	coins := sdk.NewCoins(coin)

	if err := bank.SendCoinsFromAccountToModule(ctx, fromAddr, ModuleName, coins); err != nil {
		return err
	}

	// NOTE: should never fail since the coin is just transfered to the module
	// account before the burn
	funcs.MustNoErr(bank.BurnCoins(ctx, ModuleName, coins))

	return nil
}

func mint(ctx sdk.Context, bank BankKeeper, toAddr sdk.AccAddress, coin sdk.Coin) error {
	coins := sdk.NewCoins(coin)

	if err := bank.MintCoins(ctx, ModuleName, coins); err != nil {
		return err
	}

	// NOTE: should never fail since the coin is just minted to the module
	// account before the transfer
	funcs.MustNoErr(bank.SendCoinsFromModuleToAccount(ctx, ModuleName, toAddr, coins))

	return nil
}

func getCoinType(ctx sdk.Context, nexus Nexus, denom string) (CoinType, error) {
	switch {
	// check if the format of token denomination is 'ibc/{hash}'
	case isIBCDenom(denom):
		return ICS20, nil
	case isNativeAssetOnAxelarnet(ctx, nexus, denom):
		return Native, nil
	case nexus.IsAssetRegistered(ctx, axelarnet.Axelarnet, denom):
		return External, nil
	default:
		return Unrecognized, fmt.Errorf("unrecognized coin %s", denom)
	}
}

// isIBCDenom validates that the given denomination is a valid ICS token representation (ibc/{hash})
func isIBCDenom(denom string) bool {
	if err := sdk.ValidateDenom(denom); err != nil {
		return false
	}

	denomSplit := strings.SplitN(denom, "/", 2)
	if len(denomSplit) != 2 || denomSplit[0] != ibctypes.DenomPrefix {
		return false
	}

	if _, err := ibctypes.ParseHexHash(denomSplit[1]); err != nil {
		return false
	}

	return true
}

func isNativeAssetOnAxelarnet(ctx sdk.Context, nexus Nexus, denom string) bool {
	chain, ok := nexus.GetChainByNativeAsset(ctx, denom)

	return ok && chain.Name.Equals(axelarnet.Axelarnet.Name)
}
