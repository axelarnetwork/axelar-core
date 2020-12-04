package types

import (
	"crypto/ecdsa"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	voting "github.com/axelarnetwork/axelar-core/x/voting/exported"
)

type Voter interface {
	InitPoll(ctx sdk.Context, poll voting.PollMeta) error
	Vote(ctx sdk.Context, vote voting.MsgVote) error
	TallyVote(ctx sdk.Context, vote voting.MsgVote) error
	Result(ctx sdk.Context, poll voting.PollMeta) voting.Vote
}

type Signer interface {
	GetSig(ctx sdk.Context, sigID string) (tss.Signature, error)
	GetKey(ctx sdk.Context, keyID string) (ecdsa.PublicKey, error)
	GetLatestMasterKey(ctx sdk.Context, chain string) (ecdsa.PublicKey, error)
}
