package types

import (
	"crypto/ecdsa"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/voting/exported"
)

type Voter interface {
	InitPoll(ctx sdk.Context, poll exported.PollMeta) error
	Vote(ctx sdk.Context, vote exported.MsgVote) error
	TallyVote(ctx sdk.Context, vote exported.MsgVote) error
	Result(ctx sdk.Context, poll exported.PollMeta) exported.Vote
}

type Signer interface {
	GetSig(ctx sdk.Context, sigID string) (r *big.Int, s *big.Int, e error)
	GetKey(ctx sdk.Context, keyID string) (ecdsa.PublicKey, error)
}
