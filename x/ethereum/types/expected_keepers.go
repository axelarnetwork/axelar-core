package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

//go:generate moq -out ./mock/expected_keepers.go -pkg mock . Voter Signer Nexus Snapshotter

// Voter wraps around the existing exported.Voter interface to adhere to the Cosmos convention of keeping all
// expected keepers from other modules in the expected_keepers.go file
type Voter interface {
	InitPoll(ctx sdk.Context, poll vote.PollMeta) error
	DeletePoll(ctx sdk.Context, poll vote.PollMeta)
	RecordVote(vote vote.MsgVote)
	TallyVote(ctx sdk.Context, sender sdk.AccAddress, pollMeta vote.PollMeta, data vote.VotingData) error
	Result(ctx sdk.Context, poll vote.PollMeta) vote.VotingData
}

// Nexus provides functionality to manage cross-chain transfers
type Nexus interface {
	LinkAddresses(ctx sdk.Context, sender exported.CrossChainAddress, recipient exported.CrossChainAddress)
	GetRecipient(ctx sdk.Context, sender exported.CrossChainAddress) (exported.CrossChainAddress, bool)
	EnqueueForTransfer(ctx sdk.Context, sender exported.CrossChainAddress, amount sdk.Coin) error
	GetPendingTransfersForChain(ctx sdk.Context, chain exported.Chain) []exported.CrossChainTransfer
	GetArchivedTransfersForChain(ctx sdk.Context, chain exported.Chain) []exported.CrossChainTransfer
	ArchivePendingTransfer(ctx sdk.Context, transfer exported.CrossChainTransfer)
	GetChain(ctx sdk.Context, chain string) (exported.Chain, bool)
	IsAssetRegistered(ctx sdk.Context, chainName, denom string) bool
	RegisterAsset(ctx sdk.Context, chainName, denom string)
}

// InitPoller is a minimal interface to start a poll. This must be a type alias instead of a type definition,
// because the concrete implementation of Signer (specifically StartSign) is defined in a different package using another (identical)
// InitPoller interface. Go cannot match the types otherwise
type InitPoller = interface {
	InitPoll(ctx sdk.Context, poll vote.PollMeta) error
}

// Signer provides keygen and signing functionality
type Signer interface {
	StartSign(ctx sdk.Context, initPoll InitPoller, keyID string, sigID string, msg []byte, snapshot snapshot.Snapshot) error
	GetCurrentMasterKeyID(ctx sdk.Context, chain exported.Chain) (string, bool)
	GetSig(ctx sdk.Context, sigID string) (tss.Signature, bool)
	GetKey(ctx sdk.Context, keyID string) (tss.Key, bool)
	GetCurrentMasterKey(ctx sdk.Context, chain exported.Chain) (tss.Key, bool)
	GetNextMasterKey(ctx sdk.Context, chain exported.Chain) (tss.Key, bool)
	GetKeyForSigID(ctx sdk.Context, sigID string) (tss.Key, bool)
	GetSnapshotCounterForKeyID(ctx sdk.Context, keyID string) (int64, bool)
}

// Snapshotter provides access to the snapshot functionality
type Snapshotter interface {
	GetSnapshot(ctx sdk.Context, counter int64) (snapshot.Snapshot, bool)
}
