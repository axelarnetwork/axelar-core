package types

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
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

	GetAnyoneCanSpendAddress(ctx sdk.Context) AddressInfo
	GetRequiredConfirmationHeight(ctx sdk.Context) uint64
	GetRevoteLockingPeriod(ctx sdk.Context) int64
	GetSigCheckInterval(ctx sdk.Context) int64
	GetNetwork(ctx sdk.Context) Network
	GetMinimumWithdrawalAmount(ctx sdk.Context) btcutil.Amount
	GetMaxInputCount(ctx sdk.Context) int64

	SetPendingOutpointInfo(ctx sdk.Context, key vote.PollKey, info OutPointInfo)
	GetPendingOutPointInfo(ctx sdk.Context, key vote.PollKey) (OutPointInfo, bool)
	DeletePendingOutPointInfo(ctx sdk.Context, key vote.PollKey)
	GetOutPointInfo(ctx sdk.Context, outPoint wire.OutPoint) (OutPointInfo, OutPointState, bool)
	DeleteOutpointInfo(ctx sdk.Context, outPoint wire.OutPoint)
	SetSpentOutpointInfo(ctx sdk.Context, info OutPointInfo)
	SetConfirmedOutpointInfo(ctx sdk.Context, keyID string, info OutPointInfo)
	GetConfirmedOutpointInfoQueueForKey(ctx sdk.Context, keyID string) utils.KVQueue

	SetUnsignedTx(ctx sdk.Context, tx *Transaction)
	GetUnsignedTx(ctx sdk.Context) (*Transaction, bool)
	DeleteUnsignedTx(ctx sdk.Context)
	SetSignedTx(ctx sdk.Context, tx *wire.MsgTx)
	GetSignedTx(ctx sdk.Context, txHash chainhash.Hash) (*wire.MsgTx, bool)
	GetLatestSignedTxHash(ctx sdk.Context) (*chainhash.Hash, bool)

	SetAddress(ctx sdk.Context, address AddressInfo)
	GetAddress(ctx sdk.Context, encodedAddress string) (AddressInfo, bool)

	GetDustAmount(ctx sdk.Context, encodedAddress string) btcutil.Amount
	SetDustAmount(ctx sdk.Context, encodedAddress string, amount btcutil.Amount)
	DeleteDustAmount(ctx sdk.Context, encodedAddress string)
}

// Voter is the interface that provides voting functionality
type Voter interface {
	NewPoll(ctx sdk.Context, metadata vote.PollMetadata) vote.Poll
	GetPoll(ctx sdk.Context, pollKey vote.PollKey) vote.Poll
	GetDefaultVotingThreshold(ctx sdk.Context) utils.Threshold
}

// InitPoller is a minimal interface to start a poll. This must be a type alias instead of a type definition,
// because the concrete implementation of Signer (specifically StartSign) is defined in a different package using another (identical)
// InitPoller interface. Go cannot match the types otherwise
type InitPoller = interface {
	NewPoll(ctx sdk.Context, metadata vote.PollMetadata) vote.Poll
	GetDefaultVotingThreshold(ctx sdk.Context) utils.Threshold
}

// Signer provides keygen and signing functionality
type Signer interface {
	StartSign(ctx sdk.Context, voter InitPoller, keyID string, sigID string, msg []byte, snapshot snapshot.Snapshot) error
	GetSig(ctx sdk.Context, sigID string) (tss.Signature, bool)
	GetCurrentKeyID(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (string, bool)
	GetCurrentKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool)
	GetNextKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool)
	GetSnapshotCounterForKeyID(ctx sdk.Context, keyID string) (int64, bool)
	GetKey(ctx sdk.Context, keyID string) (tss.Key, bool)
	AssignNextKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole, keyID string) error
	AssertMatchesRequirements(ctx sdk.Context, snapshotter Snapshotter, chain nexus.Chain, keyID string, keyRole tss.KeyRole) error
}

// Nexus provides functionality to manage cross-chain transfers
type Nexus interface {
	LinkAddresses(ctx sdk.Context, sender nexus.CrossChainAddress, recipient nexus.CrossChainAddress)
	GetRecipient(ctx sdk.Context, sender nexus.CrossChainAddress) (nexus.CrossChainAddress, bool)
	EnqueueForTransfer(ctx sdk.Context, sender nexus.CrossChainAddress, amount sdk.Coin) error
	GetTransfersForChain(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState) []nexus.CrossChainTransfer
	ArchivePendingTransfer(ctx sdk.Context, transfer nexus.CrossChainTransfer)
	GetChain(ctx sdk.Context, chain string) (nexus.Chain, bool)
	IsAssetRegistered(ctx sdk.Context, chainName, denom string) bool
}

// Snapshotter provides snapshot functionality
type Snapshotter = snapshot.Snapshotter
