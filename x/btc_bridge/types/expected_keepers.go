package types

import (
	"crypto/ecdsa"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

type Voter interface {
	SetFutureVote(ctx sdk.Context, vote exported.FutureVote)
}

type Signer interface {
	// TODO: StartSign should not depend on a msg type from a different module
	StartSign(ctx sdk.Context, info types.MsgSignStart) error
	GetSig(ctx sdk.Context, sigID string) (r *big.Int, s *big.Int)
	GetKey(ctx sdk.Context, keyID string) ecdsa.PublicKey
}
