package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
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
	if t.Is(types.NonExistent) {
		return types.Command{}, fmt.Errorf("token is non-existent")
	}
	if t.Is(types.Confirmed) {
		return types.Command{}, fmt.Errorf("token is already confirmed")
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

func (t *erc20Token) StartVoting(txID types.Hash) (vote.PollKey, []vote.PollProperty, error) {
	switch {
	case t.Is(types.Confirmed):
		return vote.PollKey{}, nil, fmt.Errorf("token %s is already confirmed", t.Asset)
	case t.Is(types.Voting):
		return vote.PollKey{}, nil, fmt.Errorf("voting for token %s is already underway", t.Asset)
	case t.Is(types.NonExistent):
		return vote.PollKey{}, nil, fmt.Errorf("non-existent token for asset '%s'", t.Asset)
	}

	period, ok := t.getRevoteLockingPeriod(t.ctx)
	if !ok {
		return vote.PollKey{}, nil, fmt.Errorf("could not retrieve revote locking period")
	}

	votingThreshold, ok := t.getVotingThreshold(t.ctx)
	if !ok {
		return vote.PollKey{}, nil, fmt.Errorf("voting threshold not found")
	}

	minVoterCount, ok := t.getMinVoterCount(t.ctx)
	if !ok {
		return vote.PollKey{}, nil, fmt.Errorf("min voter count not found")
	}

	properties := []vote.PollProperty{
		vote.ExpiryAt(t.ctx.BlockHeight() + period),
		vote.Threshold(votingThreshold),
		vote.MinVoterCount(minVoterCount),
	}

	t.ERC20TokenMetadata.TxHash = txID
	t.Status |= types.Voting
	t.setMeta(t.ctx, t.Asset, t.ERC20TokenMetadata)

	return t.getPollKey(txID), properties, nil
}

func (t *erc20Token) getPollKey(txID types.Hash) vote.PollKey {
	return vote.NewPollKey(types.ModuleName, txID.Hex()+"_"+t.Asset)
}

func (t *erc20Token) ValidatePollKey(key vote.PollKey) error {
	switch {
	case t.Is(types.Confirmed):
		return fmt.Errorf("token %s already confirmed", t.Asset)
	case !t.Is(types.Voting):
		return fmt.Errorf("voting for token not underway %s", t.Asset)
	case t.getPollKey(t.TxHash) != key:
		return fmt.Errorf("poll key mismatch (expected %s, got %s)", t.getPollKey(t.TxHash).String(), key.String())
	default:
		// assert: the token is known and has not been confirmed before
		return nil
	}
}

func (t *erc20Token) Reset() {
	if !t.Is(types.Initialized) {
		return
	}

	t.Status = types.Initialized
	t.ERC20TokenMetadata.TxHash = types.Hash{}
	t.setMeta(t.ctx, t.Asset, t.ERC20TokenMetadata)
}

func (t *erc20Token) Confirm() {
	if !t.Is(types.Initialized) {
		return
	}

	t.Status |= types.Confirmed
	t.setMeta(t.ctx, t.Asset, t.ERC20TokenMetadata)
}
