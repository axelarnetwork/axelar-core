package types

import (
	"crypto/ecdsa"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/balance/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . Voter Signer Balancer Snapshotter

type Voter interface {
	voting.Voter
}

type Snapshotter interface {
	GetSnapshot(ctx sdk.Context, round int64) (snapshot.Snapshot, bool)
}

type Signer interface {
	StartSign(ctx sdk.Context, keyID string, sigID string, msg []byte, validators []snapshot.Validator) error
	GetCurrentMasterKeyID(ctx sdk.Context, chain exported.Chain) (string, bool)
	GetSig(ctx sdk.Context, sigID string) (tss.Signature, bool)
	GetKey(ctx sdk.Context, keyID string) (ecdsa.PublicKey, bool)
	GetCurrentMasterKey(ctx sdk.Context, chain exported.Chain) (ecdsa.PublicKey, bool)
	GetNextMasterKey(ctx sdk.Context, chain exported.Chain) (ecdsa.PublicKey, bool)
	GetKeyForSigID(ctx sdk.Context, sigID string) (ecdsa.PublicKey, bool)
	GetSnapshotRoundForKeyID(ctx sdk.Context, keyID string) (int64, bool)
}

type Balancer interface {
	LinkAddresses(ctx sdk.Context, sender exported.CrossChainAddress, recipient exported.CrossChainAddress)
	PrepareForTransfer(ctx sdk.Context, sender exported.CrossChainAddress, amount sdk.Coin) error
	GetPendingTransfersForChain(ctx sdk.Context, chain exported.Chain) []exported.CrossChainTransfer
	GetArchivedTransfersForChain(ctx sdk.Context, chain exported.Chain) []exported.CrossChainTransfer
	ArchivePendingTransfers(ctx sdk.Context, recipient exported.CrossChainAddress)
}
