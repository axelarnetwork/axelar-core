package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/utils/funcs"
)

type coin struct {
	sdk.Coin
	ctx      sdk.Context
	k        Keeper
	nexusK   types.Nexus
	coinType types.CoinType
}

func newCoin(ctx sdk.Context, k Keeper, nexusK types.Nexus, c sdk.Coin) coin {
	return coin{
		Coin:     c,
		ctx:      ctx,
		k:        k,
		nexusK:   nexusK,
		coinType: getCoinType(ctx, nexusK, c.Denom),
	}
}

// lock locks coin from deposit address to escrow address
func (c coin) lock(ibcK IBCKeeper, bankK types.BankKeeper, depositAddr sdk.AccAddress) error {
	switch c.coinType {
	case types.ICS20:
		// get base denomination and tracing path
		denomTrace, err := ibcK.ParseIBCDenom(c.ctx, c.Denom)
		if err != nil {
			return err
		}

		err = c.validateDenomTrace(denomTrace)
		if err != nil {
			return err
		}

		// lock tokens in escrow address
		escrowAddress := types.GetEscrowAddress(c.Denom)
		if err := bankK.SendCoins(
			c.ctx, depositAddr, escrowAddress, sdk.NewCoins(c.Coin),
		); err != nil {
			return err
		}
	case types.Native:
		// lock tokens in escrow address
		escrowAddress := types.GetEscrowAddress(c.GetDenom())
		if err := bankK.SendCoins(
			c.ctx, depositAddr, escrowAddress, sdk.NewCoins(c.Coin),
		); err != nil {
			return err
		}
	case types.External:
		// transfer coins from linked address to module account and burn them
		if err := bankK.SendCoinsFromAccountToModule(
			c.ctx, depositAddr, types.ModuleName, sdk.NewCoins(c.Coin),
		); err != nil {
			return err
		}

		// NOTE: should not happen as the module account was
		// retrieved on the step above, and it has enough balance
		// to burn.
		funcs.MustNoErr(bankK.BurnCoins(c.ctx, types.ModuleName, sdk.NewCoins(c.Coin)))
	default:
		return fmt.Errorf("unrecognized coin type %d", c.coinType)
	}

	return nil
}

func (c coin) normalizeDenom(ibcK IBCKeeper) (sdk.Coin, error) {
	switch c.coinType {
	case types.ICS20:
		// get base denomination and tracing path
		denomTrace, err := ibcK.ParseIBCDenom(c.ctx, c.Denom)
		if err != nil {
			return sdk.Coin{}, err
		}

		// convert denomination from 'ibc/{hash}' to native asset that recognized by nexus module
		return sdk.NewCoin(denomTrace.GetBaseDenom(), c.Amount), nil
	default:
		return c.Coin, nil
	}
}

func (c coin) validateDenomTrace(denomTrace ibctypes.DenomTrace) error {
	if c.coinType != types.ICS20 {
		return fmt.Errorf("%s is not ICS20 token", c.GetDenom())
	}

	// check if the asset registered with a path
	chain, ok := c.nexusK.GetChainByNativeAsset(c.ctx, denomTrace.GetBaseDenom())
	if !ok {
		return fmt.Errorf("asset %s is not linked to a cosmos chain", denomTrace.GetBaseDenom())
	}

	path, ok := c.k.GetIBCPath(c.ctx, chain.Name)
	if !ok {
		return fmt.Errorf("path not found for chain %s", chain.Name)
	}

	if path != denomTrace.Path {
		return fmt.Errorf("path %s does not match registered path %s for asset %s", denomTrace.GetPath(), path, denomTrace.GetBaseDenom())
	}

	return nil
}

func getCoinType(ctx sdk.Context, nexusK types.Nexus, denom string) types.CoinType {
	switch {
	// check if the format of token denomination is 'ibc/{hash}'
	case isIBCDenom(denom):
		return types.ICS20
	case isNativeAssetOnAxelarnet(ctx, nexusK, denom):
		return types.Native
	case nexusK.IsAssetRegistered(ctx, exported.Axelarnet, denom):
		return types.External
	default:
		return types.Unrecognized
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

func isNativeAssetOnAxelarnet(ctx sdk.Context, n types.Nexus, denom string) bool {
	chain, ok := n.GetChainByNativeAsset(ctx, denom)
	return ok && chain.Name.Equals(exported.Axelarnet.Name)
}
