package types

import (
	"github.com/btcsuite/btcd/wire"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . Voter Signer Nexus Snapshotter BTCKeeper

// BTCKeeper is implemented by this module's keeper
type BTCKeeper interface {
	Logger(ctx sdk.Context) log.Logger
	SetParams(ctx sdk.Context, p Params)
	GetParams(ctx sdk.Context) Params
	Codec() *codec.Codec

	GetRequiredConfirmationHeight(ctx sdk.Context) uint64
	GetRevoteLockingPeriod(ctx sdk.Context) int64
	GetSigCheckInterval(ctx sdk.Context) int64
	GetNetwork(ctx sdk.Context) Network

	SetUnconfirmedOutpointInfo(ctx sdk.Context, poll vote.PollMeta, info OutPointInfo)
	GetUnconfirmedOutPointInfo(ctx sdk.Context, poll vote.PollMeta) (OutPointInfo, bool)
	DeleteUnconfirmedOutPointInfo(ctx sdk.Context, poll vote.PollMeta)
	SetOutpointInfo(ctx sdk.Context, info OutPointInfo, state OutPointState)
	GetOutPointInfo(ctx sdk.Context, outPoint wire.OutPoint) (OutPointInfo, OutPointState, bool)
	DeleteOutpointInfo(ctx sdk.Context, outPoint wire.OutPoint)
	GetConfirmedOutPointInfos(ctx sdk.Context) []OutPointInfo

	SetUnsignedTx(ctx sdk.Context, tx *wire.MsgTx)
	GetUnsignedTx(ctx sdk.Context) (*wire.MsgTx, bool)
	DeleteUnsignedTx(ctx sdk.Context)
	SetSignedTx(ctx sdk.Context, tx *wire.MsgTx)
	GetSignedTx(ctx sdk.Context) (*wire.MsgTx, bool)
	DeleteSignedTx(ctx sdk.Context)

	SetAddress(ctx sdk.Context, address AddressInfo)
	GetAddress(ctx sdk.Context, encodedAddress string) (AddressInfo, bool)
}

// Voter is the interface that provides voting functionality
type Voter interface {
	InitPoll(ctx sdk.Context, poll vote.PollMeta) error
	DeletePoll(ctx sdk.Context, poll vote.PollMeta)
	TallyVote(ctx sdk.Context, sender sdk.AccAddress, pollMeta vote.PollMeta, data vote.VotingData) error
	Result(ctx sdk.Context, poll vote.PollMeta) vote.VotingData
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
	GetSig(ctx sdk.Context, sigID string) (tss.Signature, bool)
	GetCurrentMasterKey(ctx sdk.Context, chain exported.Chain) (tss.Key, bool)
	GetNextMasterKey(ctx sdk.Context, chain exported.Chain) (tss.Key, bool)
	GetSnapshotCounterForKeyID(ctx sdk.Context, keyID string) (int64, bool)
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
}

// Snapshotter provides snapshot functionality
type Snapshotter interface {
	GetSnapshot(ctx sdk.Context, counter int64) (snapshot.Snapshot, bool)
}
