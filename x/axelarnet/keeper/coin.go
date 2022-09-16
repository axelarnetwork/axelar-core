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
	ibcK     IBCKeeper
	nexusK   types.Nexus
	coinType types.CoinType
}

// newCoin creates a coin struct, assign a coin type and normalize the denom if it's a ICS20 token
func newCoin(ctx sdk.Context, ibcK IBCKeeper, nexusK types.Nexus, c sdk.Coin) (coin, error) {
	c2 := coin{
		Coin:     c,
		ctx:      ctx,
		ibcK:     ibcK,
		nexusK:   nexusK,
		coinType: getCoinType(ctx, nexusK, c.Denom),
	}
	err := c2.normalizeDenom()

	return c2, err
}

// Lock locks coin from deposit address to escrow address
func (c coin) Lock(bankK types.BankKeeper, depositAddr sdk.AccAddress) error {
	switch c.coinType {
	case types.ICS20:
		// convert to IBC denom
		ics20, err := c.toICS20(c.GetDenom())
		if err != nil {
			return err
		}

		// lock tokens in escrow address
		escrowAddress := types.GetEscrowAddress(c.Denom)
		if err := bankK.SendCoins(
			c.ctx, depositAddr, escrowAddress, sdk.NewCoins(ics20),
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

// normalizeDenom converts from 'ibc/{hash}' to native asset that recognized by nexus module,
// returns error if trace is not found in IBC transfer store
func (c *coin) normalizeDenom() error {
	if !isIBCDenom(c.GetDenom()) || c.coinType != types.ICS20 {
		return nil
	}

	// get base denomination and tracing path
	denomTrace, err := c.ibcK.ParseIBCDenom(c.ctx, c.Denom)
	if err != nil {
		return err
	}

	// convert denomination from 'ibc/{hash}' to native asset that recognized by nexus module
	c.Coin = sdk.NewCoin(denomTrace.GetBaseDenom(), c.Amount)

	return nil
}

// toICS20 creates an ICS20 token from base denom, returns error if the asset is not registered on Axelarnet
func (c coin) toICS20(denom string) (sdk.Coin, error) {
	if c.coinType != types.ICS20 {
		return sdk.Coin{}, fmt.Errorf("%s is not ICS20 token", c.GetDenom())
	}

	// check if the asset registered with a path
	chain, ok := c.nexusK.GetChainByNativeAsset(c.ctx, denom)
	if !ok {
		return sdk.Coin{}, fmt.Errorf("asset %s is not linked to a cosmos chain", denom)
	}

	path, ok := c.ibcK.GetIBCPath(c.ctx, chain.Name)
	if !ok {
		return sdk.Coin{}, fmt.Errorf("path not found for chain %s", chain.Name)
	}

	trace := ibctypes.DenomTrace{
		Path:      path,
		BaseDenom: denom,
	}

	return sdk.NewCoin(trace.IBCDenom(), c.Amount), nil
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
