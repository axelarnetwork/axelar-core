package types

import (
	"crypto/ecdsa"

	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//go:generate moq -out ./mock/expected_keepers.go -pkg mock . Voter Signer

type Voter interface {
	voting.Voter
}

type Signer interface {
	GetSig(ctx sdk.Context, sigID string) (tss.Signature, bool)
	GetKey(ctx sdk.Context, keyID string) (ecdsa.PublicKey, bool)
	GetCurrentMasterKey(ctx sdk.Context, chain string) (ecdsa.PublicKey, bool)
	GetNextMasterKey(ctx sdk.Context, chain string) (ecdsa.PublicKey, bool)
}
