package types

import (
	"crypto/ecdsa"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
)

type Voter interface {
	SetFutureVote(ctx sdk.Context, vote exported.FutureVote)
	IsVerified(ctx sdk.Context, tx exported.ExternalTx) bool
}

type Signer interface {
	GetSig(ctx sdk.Context, sigID string) (r *big.Int, s *big.Int, e error)
	GetKey(ctx sdk.Context, keyID string) (ecdsa.PublicKey, error)
}
