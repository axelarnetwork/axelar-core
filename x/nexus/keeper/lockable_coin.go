package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"

	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/funcs"
)

// NewLockableAsset creates a new lockable asset
func (k Keeper) NewLockableAsset(ctx sdk.Context, ibc types.IBCKeeper, bank types.BankKeeper, coin sdk.Coin) (exported.LockableAsset, error) {
	return newLockableAsset(ctx, k, ibc, bank, coin)
}

// lockableAsset provides functionality to lock and release coins
type lockableAsset struct {
	sdk.Coin
	coinType types.CoinType
	nexus    types.Nexus
	ibc      types.IBCKeeper
	bank     types.BankKeeper
}

// newLockableAsset creates a coin struct, assign a coin type and normalize the denom if it's a ICS20 token
func newLockableAsset(ctx sdk.Context, nexus types.Nexus, ibc types.IBCKeeper, bank types.BankKeeper, coin sdk.Coin) (lockableAsset, error) {
	denom := coin.GetDenom()

	coinType, err := getCoinType(ctx, nexus, denom)
	if err != nil {
		return lockableAsset{}, err
	}

	// If coin type is ICS20, we need to normalize it to convert from 'ibc/{hash}'
	// to native asset denom so that nexus could recognize it
	if coinType == types.ICS20 {
		denomTrace, err := ibc.ParseIBCDenom(ctx, denom)
		if err != nil {
			return lockableAsset{}, err
		}

		coin = sdk.NewCoin(denomTrace.GetBaseDenom(), coin.Amount)
	}

	c := lockableAsset{
		Coin:     coin,
		coinType: coinType,
		nexus:    nexus,
		ibc:      ibc,
		bank:     bank,
	}

	originalCoin, err := c.getCoin(ctx)
	if err != nil {
		return lockableAsset{}, err
	}
	if originalCoin.GetDenom() != denom {
		return lockableAsset{}, fmt.Errorf("denom mismatch, expected %s, got %s", denom, originalCoin.GetDenom())
	}

	return c, nil
}

// GetAsset returns a sdk.Coin using the nexus registered asset as the denom
func (c lockableAsset) GetAsset() sdk.Coin {
	return c.Coin
}

// GetCoin returns a sdk.Coin with the actual denom used by x/bank (e.g. ICS20 coins)
func (c lockableAsset) GetCoin(ctx sdk.Context) sdk.Coin {
	// NOTE: must not fail since it's already checked in NewCoin
	return funcs.Must(c.getCoin(ctx))
}

// LockFrom locks the given coin from the given address
func (c lockableAsset) LockFrom(ctx sdk.Context, fromAddr sdk.AccAddress) error {
	coin := c.GetCoin(ctx)

	switch c.coinType {
	case types.ICS20, types.Native:
		return lock(ctx, c.bank, fromAddr, coin)
	case types.External:
		return burn(ctx, c.bank, fromAddr, coin)
	default:
		return fmt.Errorf("unrecognized coin type %d", c.coinType)
	}
}

// UnlockTo unlocks the given coin to the given address
func (c lockableAsset) UnlockTo(ctx sdk.Context, toAddr sdk.AccAddress) error {
	coin := c.GetCoin(ctx)

	switch c.coinType {
	case types.ICS20, types.Native:
		return unlock(ctx, c.bank, toAddr, coin)
	case types.External:
		return mint(ctx, c.bank, toAddr, coin)
	default:
		return fmt.Errorf("unrecognized coin type %d", c.coinType)
	}
}

func (c lockableAsset) getCoin(ctx sdk.Context) (sdk.Coin, error) {
	switch c.coinType {
	case types.ICS20:
		return c.toICS20(ctx)
	case types.Native, types.External:
		return c.Coin, nil
	default:
		return sdk.Coin{}, fmt.Errorf("unrecognized coin type %d", c.coinType)
	}
}

func (c lockableAsset) toICS20(ctx sdk.Context) (sdk.Coin, error) {
	if c.coinType != types.ICS20 {
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

func lock(ctx sdk.Context, bank types.BankKeeper, fromAddr sdk.AccAddress, coin sdk.Coin) error {
	return bank.SendCoins(ctx, fromAddr, exported.GetEscrowAddress(coin.GetDenom()), sdk.NewCoins(coin))
}

func unlock(ctx sdk.Context, bank types.BankKeeper, toAddr sdk.AccAddress, coin sdk.Coin) error {
	return bank.SendCoins(ctx, exported.GetEscrowAddress(coin.GetDenom()), toAddr, sdk.NewCoins(coin))
}

func burn(ctx sdk.Context, bank types.BankKeeper, fromAddr sdk.AccAddress, coin sdk.Coin) error {
	coins := sdk.NewCoins(coin)

	if err := bank.SendCoinsFromAccountToModule(ctx, fromAddr, types.ModuleName, coins); err != nil {
		return err
	}

	// NOTE: should never fail since the coin is just transfered to the module
	// account before the burn
	funcs.MustNoErr(bank.BurnCoins(ctx, types.ModuleName, coins))

	return nil
}

func mint(ctx sdk.Context, bank types.BankKeeper, toAddr sdk.AccAddress, coin sdk.Coin) error {
	coins := sdk.NewCoins(coin)

	if err := bank.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return err
	}

	// NOTE: should never fail since the coin is just minted to the module
	// account before the transfer
	funcs.MustNoErr(bank.SendCoinsFromModuleToAccount(ctx, types.ModuleName, toAddr, coins))

	return nil
}

func getCoinType(ctx sdk.Context, nexus types.Nexus, denom string) (types.CoinType, error) {
	switch {
	// check if the format of token denomination is 'ibc/{hash}'
	case isIBCDenom(denom):
		return types.ICS20, nil
	case isNativeAssetOnAxelarnet(ctx, nexus, denom):
		return types.Native, nil
	case nexus.IsAssetRegistered(ctx, axelarnet.Axelarnet, denom):
		return types.External, nil
	default:
		return types.Unrecognized, fmt.Errorf("unrecognized coin %s", denom)
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

func isNativeAssetOnAxelarnet(ctx sdk.Context, nexus types.Nexus, denom string) bool {
	chain, ok := nexus.GetChainByNativeAsset(ctx, denom)

	return ok && chain.Name.Equals(axelarnet.Axelarnet.Name)
}
