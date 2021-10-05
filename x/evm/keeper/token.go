package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

type erc20Token struct {
	types.ERC20TokenMetadata

	ctx                    sdk.Context
	setMeta                func(ctx sdk.Context, asset string, meta types.ERC20TokenMetadata)
	getRevoteLockingPeriod func(ctx sdk.Context) (int64, bool)
	getMinVoterCount       func(ctx sdk.Context) (int64, bool)
	getVotingThreshold     func(ctx sdk.Context) (utils.Threshold, bool)
}

func createERC20Token(ctx sdk.Context, k keeper, meta types.ERC20TokenMetadata) *erc20Token {
	token := &erc20Token{
		ctx:                    ctx,
		ERC20TokenMetadata:     meta,
		setMeta:                k.setTokenMetadata,
		getRevoteLockingPeriod: k.GetRevoteLockingPeriod,
		getMinVoterCount:       k.GetMinVoterCount,
		getVotingThreshold:     k.GetVotingThreshold,
	}

	return token
}

func (t *erc20Token) GetAsset() string {
	return t.Asset
}

func (t *erc20Token) GetTxID() types.Hash {
	return t.TxHash
}

func (t *erc20Token) GetDetails() types.TokenDetails {
	return t.Details
}

func (t *erc20Token) Is(status types.Status) bool {
	// this special case check is needed, because 0 & x == 0 is true for any x
	if status == types.NonExistent {
		return t.Status == types.NonExistent
	}
	return status&t.Status == status
}

func (t *erc20Token) CreateDeployCommand(key tss.KeyID) (types.Command, error) {
	switch {
	case t.Is(types.NonExistent):
		return types.Command{}, fmt.Errorf("token %s non-existent", t.Asset)
	case t.Is(types.Confirmed):
		return types.Command{}, fmt.Errorf("token %s already confirmed", t.Asset)
	}
	if err := key.Validate(); err != nil {
		return types.Command{}, err
	}

	return types.CreateDeployTokenCommand(
		t.ERC20TokenMetadata.ChainID.BigInt(),
		key,
		t.Details,
	)
}

func (t *erc20Token) GetAddress() types.Address {
	return t.ERC20TokenMetadata.TokenAddress

}

func (t *erc20Token) StartVoting(txID types.Hash) error {
	switch {
	case t.Is(types.NonExistent):
		return fmt.Errorf("token %s non-existent", t.Asset)
	case t.Is(types.Confirmed):
		return fmt.Errorf("token %s already confirmed", t.Asset)
	case t.Is(types.Waiting):
		return fmt.Errorf("voting for token %s is already underway", t.Asset)
	}

	t.ERC20TokenMetadata.TxHash = txID
	t.Status |= types.Waiting
	t.setMeta(t.ctx, t.Asset, t.ERC20TokenMetadata)

	return nil
}

func (t *erc20Token) Reset() error {
	switch {
	case t.Is(types.NonExistent):
		return fmt.Errorf("token %s non-existent", t.Asset)
	case !t.Is(types.Waiting):
		return fmt.Errorf("token %s not waiting confirmation (current status: %s)", t.Asset, t.Status.String())
	}

	t.Status = types.Initialized
	t.ERC20TokenMetadata.TxHash = types.Hash{}
	t.setMeta(t.ctx, t.Asset, t.ERC20TokenMetadata)
	return nil
}

func (t *erc20Token) Confirm() error {
	switch {
	case t.Is(types.NonExistent):
		return fmt.Errorf("token %s non-existent", t.Asset)
	case !t.Is(types.Waiting):
		return fmt.Errorf("token %s not waiting confirmation (current status: %s)", t.Asset, t.Status.String())
	}

	t.Status = types.Confirmed
	t.setMeta(t.ctx, t.Asset, t.ERC20TokenMetadata)

	return nil
}
