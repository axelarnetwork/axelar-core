package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"

	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/funcs"
)

// lockableAsset provides functionality to lock and release coins
type lockableAsset struct {
	sdk.Coin
	coinType types.CoinType
	nexus    types.Nexus
	ibc      types.IBCKeeper
	bank     types.BankKeeper
}

// NewLockableAsset creates a new lockable asset from a nexus registered asset name.
func (k Keeper) NewLockableAsset(ctx sdk.Context, ibc types.IBCKeeper, bank types.BankKeeper, coin sdk.Coin) (exported.LockableAsset, error) {
	return newLockableAsset(ctx, k, ibc, bank, coin)
}

// NewLockableAssetFromCosmosCoin creates a new lockable asset from a Cosmos x/bank Coin (e.g. uaxl, ICS20 coin).
func (k Keeper) NewLockableAssetFromCosmosCoin(ctx sdk.Context, ibc types.IBCKeeper, bank types.BankKeeper, coin sdk.Coin) (exported.LockableAsset, error) {
	if isIBCDenom(coin.GetDenom()) {
		denomTrace, err := ibc.ParseIBCDenom(ctx, coin.GetDenom())
		if err != nil {
			return lockableAsset{}, err
		}

		coin = sdk.NewCoin(denomTrace.GetBaseDenom(), coin.Amount)
	}

	asset, err := newLockableAsset(ctx, k, ibc, bank, coin)
	if err != nil {
		return lockableAsset{}, err
	}

	if convertedCoin, err := asset.getCoin(ctx); err != nil {
		return lockableAsset{}, err
	} else if convertedCoin != coin {
		// validate that the converted coin denom is the same as the original denom
		// this is needed for ICS20 coins whose IBC denom trace must correspond to the registered IBC path
		return lockableAsset{}, fmt.Errorf("converted coin %s is different from the original coin %s", convertedCoin, coin)
	}

	return asset, nil
}

// newLockableAsset creates a coin struct, assign a coin type and normalize the denom if it's a ICS20 token
func newLockableAsset(ctx sdk.Context, nexus types.Nexus, ibc types.IBCKeeper, bank types.BankKeeper, coin sdk.Coin) (lockableAsset, error) {
	denom := coin.GetDenom()

	coinType, err := getCoinType(ctx, nexus, ibc, denom)
	if err != nil {
		return lockableAsset{}, err
	}

	c := lockableAsset{
		Coin:     coin,
		coinType: coinType,
		nexus:    nexus,
		ibc:      ibc,
		bank:     bank,
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
		return toICS20(ctx, c.nexus, c.ibc, c.Coin)
	case types.Native, types.External:
		return c.Coin, nil
	default:
		return sdk.Coin{}, fmt.Errorf("unrecognized coin type %d", c.coinType)
	}
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

func getCoinType(ctx sdk.Context, nexus types.Nexus, ibc types.IBCKeeper, denom string) (types.CoinType, error) {
	switch {
	// check if the denom is the registered asset name for an ICS20 coin from a cosmos chain
	case isFromExternalCosmosChain(ctx, nexus, ibc, denom):
		return types.ICS20, nil
	case isNativeAssetOnAxelarnet(ctx, nexus, denom):
		return types.Native, nil
	case nexus.IsAssetRegistered(ctx, axelarnet.Axelarnet, denom):
		return types.External, nil
	default:
		return types.Unrecognized, fmt.Errorf("unrecognized coin %s", denom)
	}
}

// isFromExternalCosmosChain returns true if the denom is a nexus-registered
// asset name for an ICS20 coin originating from a cosmos chain
func isFromExternalCosmosChain(ctx sdk.Context, nexus types.Nexus, ibc types.IBCKeeper, denom string) bool {
	if _, err := getIBCPath(ctx, nexus, ibc, denom); err != nil {
		return false
	}

	return true
}

// isIBCDenom validates that the given denomination is a valid ICS token representation (ibc/{hash})
func isIBCDenom(denom string) bool {
	if err := ibctypes.ValidateIBCDenom(denom); err != nil {
		return false
	}

	return true
}

func isNativeAssetOnAxelarnet(ctx sdk.Context, nexus types.Nexus, denom string) bool {
	chain, ok := nexus.GetChainByNativeAsset(ctx, denom)

	return ok && chain.Name.Equals(axelarnet.Axelarnet.Name)
}

func getIBCPath(ctx sdk.Context, nexus types.Nexus, ibc types.IBCKeeper, asset string) (string, error) {
	// check if the asset registered with a path
	chain, ok := nexus.GetChainByNativeAsset(ctx, asset)
	if !ok {
		return "", fmt.Errorf("asset %s is not linked to a cosmos chain", asset)
	}

	path, ok := ibc.GetIBCPath(ctx, chain.Name)
	if !ok {
		return "", fmt.Errorf("path not found for chain %s", chain.Name)
	}

	return path, nil
}

func toICS20(ctx sdk.Context, nexus types.Nexus, ibc types.IBCKeeper, coin sdk.Coin) (sdk.Coin, error) {
	path, err := getIBCPath(ctx, nexus, ibc, coin.GetDenom())
	if err != nil {
		return sdk.Coin{}, err
	}

	trace := ibctypes.DenomTrace{
		Path:      path,
		BaseDenom: coin.GetDenom(),
	}

	return sdk.NewCoin(trace.IBCDenom(), coin.Amount), nil
}
