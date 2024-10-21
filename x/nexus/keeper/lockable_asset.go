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

// NewLockableAsset creates a new lockable asset.
// The coin denom can either be an actual bank Coin (e.g. uaxl, ICS20 coin) ,
// or the registered asset name (e.g. base denom for an ICS20 coin)
func (k Keeper) NewLockableAsset(ctx sdk.Context, ibc types.IBCKeeper, bank types.BankKeeper, coin sdk.Coin) (exported.LockableAsset, error) {
	return newLockableAsset(ctx, k, ibc, bank, coin)
}

// lockableAsset provides functionality to lock and release coins
type lockableAsset struct {
	asset    sdk.Coin
	coinType types.CoinType
	nexus    types.Nexus
	ibc      types.IBCKeeper
	bank     types.BankKeeper
}

// newLockableAsset creates a coin struct, assign a coin type and normalize the denom if it's a ICS20 token
func newLockableAsset(ctx sdk.Context, nexus types.Nexus, ibc types.IBCKeeper, bank types.BankKeeper, coin sdk.Coin) (lockableAsset, error) {
	asset, coinType, err := toAssetAndType(ctx, nexus, ibc, coin)
	if err != nil {
		return lockableAsset{}, err
	}

	c := lockableAsset{
		asset:    asset,
		coinType: coinType,
		nexus:    nexus,
		ibc:      ibc,
		bank:     bank,
	}

	if _, err := c.getCoin(ctx); err != nil {
		return lockableAsset{}, err
	}

	return c, nil
}

// GetAsset returns a sdk.Coin using the nexus registered asset as the denom
func (c lockableAsset) GetAsset() sdk.Coin {
	return c.asset
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
		return toICS20(ctx, c.nexus, c.ibc, c.asset)
	case types.Native, types.External:
		return c.asset, nil
	default:
		return sdk.Coin{}, fmt.Errorf("unrecognized coin type %d", c.coinType)
	}
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

// toICS20 converts a registered asset originating from an external cosmos chain to an ICS20 coin
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

// toAssetAndType normalizes the given denom to the corresponding registered asset and returns the coin type
func toAssetAndType(ctx sdk.Context, nexus types.Nexus, ibc types.IBCKeeper, coin sdk.Coin) (sdk.Coin, types.CoinType, error) {
	denom := coin.GetDenom()

	switch {
	// check if the format of token denomination is 'ibc/{hash}' and it's a registered asset
	case isICS20Denom(denom):
		{
			denomTrace, err := ibc.ParseIBCDenom(ctx, denom)
			if err != nil {
				return sdk.Coin{}, types.Unrecognized, err
			}

			path, err := getIBCPath(ctx, nexus, ibc, denomTrace.GetBaseDenom())
			if err != nil {
				return sdk.Coin{}, types.Unrecognized, err
			}
			if path != denomTrace.GetPath() {
				return sdk.Coin{}, types.Unrecognized, fmt.Errorf("expected ics20 coin IBC path %s, got %s", path, denomTrace.GetPath())
			}

			asset := sdk.NewCoin(denomTrace.GetBaseDenom(), coin.Amount)

			return asset, types.ICS20, nil
		}
	// check if the denom is the registered asset name for an ICS20 coin from a cosmos chain
	case isFromExternalCosmosChain(ctx, nexus, ibc, denom):
		return coin, types.ICS20, nil
	case isNativeAssetOnAxelarnet(ctx, nexus, denom):
		return coin, types.Native, nil
	case nexus.IsAssetRegistered(ctx, axelarnet.Axelarnet, denom):
		return coin, types.External, nil
	default:
		return sdk.Coin{}, types.Unrecognized, fmt.Errorf("unrecognized coin %s", denom)
	}
}

// isFromExternalCosmosChain returns true if the denom is a nexus-registered
// asset name for an ICS20 coin originating from a cosmos chain
func isFromExternalCosmosChain(ctx sdk.Context, nexus types.Nexus, ibc types.IBCKeeper, denom string) bool {
	_, err := getIBCPath(ctx, nexus, ibc, denom)

	return err == nil
}

// isICS20Denom validates that the given denomination is a valid ICS token representation (ibc/{hash})
func isICS20Denom(denom string) bool {
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
