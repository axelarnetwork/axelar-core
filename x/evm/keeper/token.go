package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

type erc20Token struct {
	types.ERC20TokenMetadata

	ctx     sdk.Context
	setMeta func(ctx sdk.Context, asset string, meta types.ERC20TokenMetadata)
}

func renderERC20Token(ctx sdk.Context, k keeper, meta types.ERC20TokenMetadata) *erc20Token {
	token := &erc20Token{
		ctx:                ctx,
		ERC20TokenMetadata: meta,
		setMeta:            k.setTokenMetadata,
	}

	return token
}

func (t *erc20Token) initialize(k keeper) error {
	// perform a few checks now, so that it is impossible to get errors later
	if other := k.GetERC20Token(t.ctx, t.Asset); !other.Is(types.NonExistent) {
		return fmt.Errorf("token '%s' already set", t.Asset)
	}

	gatewayAddr, found := k.GetGatewayAddress(t.ctx)
	if !found {
		return fmt.Errorf("axelar gateway address for chain '%s' not set", k.chain)
	}

	_, found = k.GetTokenByteCodes(t.ctx)
	if !found {
		return fmt.Errorf("bytecodes for token contract for chain '%s' not found", k.chain)
	}

	if err := t.ERC20TokenMetadata.Details.Validate(); err != nil {
		return err
	}

	var network string
	subspace, ok := k.getSubspace(t.ctx, k.chain)
	if !ok {
		return fmt.Errorf("could not find subspace for chain '%s'", k.chain)
	}

	subspace.Get(t.ctx, types.KeyNetwork, &network)

	chainID := k.GetChainIDByNetwork(t.ctx, network)
	if chainID == nil {
		return fmt.Errorf("could not find chain ID for chain '%s'", k.chain)
	}

	tokenAddr, err := k.getTokenAddress(t.ctx, t.Asset, t.Details, gatewayAddr)
	if err != nil {
		return err
	}

	// all good
	t.ERC20TokenMetadata.TokenAddress = types.Address(tokenAddr)
	t.ERC20TokenMetadata.ChainID = sdk.NewIntFromBigInt(chainID)
	t.setMeta(t.ctx, t.Asset, t.ERC20TokenMetadata)
	return nil
}

func (t *erc20Token) TokenDetails() types.TokenDetails {
	return t.Details
}

func (t *erc20Token) Is(status types.Status) bool {
	// this special case check is needed, because 0 & x == 0 is true for any x
	if status == types.NonExistent {
		return t.Status == types.NonExistent
	}
	return status&t.Status == status
}

func (t *erc20Token) DeployCommand(key tss.KeyID) types.Command {
	// should not return any error
	command, _ := types.CreateDeployTokenCommand(
		t.ERC20TokenMetadata.ChainID.BigInt(),
		key,
		t.Details,
	)
	return command
}

func (t *erc20Token) TokenAddress() types.Address {
	return t.ERC20TokenMetadata.TokenAddress

}

func (t *erc20Token) StartVoting(txID types.Hash) (vote.PollKey, error) {
	switch {
	case t.Is(types.Confirmed):
		return vote.PollKey{}, fmt.Errorf("token %s is already confirmed", t.Asset)
	case t.Is(types.Voting):
		return vote.PollKey{}, fmt.Errorf("voting for token %s is already underway", t.Asset)
	case t.Is(types.NonExistent):
		return vote.PollKey{}, fmt.Errorf("non-existent token for asset '%s'", t.Asset)
	}

	t.ERC20TokenMetadata.TxHash = txID
	t.Status |= types.Voting
	t.setMeta(t.ctx, t.Asset, t.ERC20TokenMetadata)

	return t.getPollKey(txID), nil
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

func (t *erc20Token) ConfirmationFailed() {
	if !t.Is(types.Initialized) {
		return
	}

	t.Status = types.Initialized
	t.setMeta(t.ctx, t.Asset, t.ERC20TokenMetadata)
}

func (t *erc20Token) ConfirmationSuccessful() {
	if !t.Is(types.Initialized) {
		return
	}

	t.Status |= types.Confirmed
	t.setMeta(t.ctx, t.Asset, t.ERC20TokenMetadata)
}
