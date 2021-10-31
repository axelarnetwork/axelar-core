package types

import (
	"crypto/ecdsa"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
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
	GetMinOutputAmount(ctx sdk.Context) btcutil.Amount
	GetMaxInputCount(ctx sdk.Context) int64
	GetMaxSecondaryOutputAmount(ctx sdk.Context) btcutil.Amount
	GetMasterKeyRetentionPeriod(ctx sdk.Context) int64
	GetMasterAddressInternalKeyLockDuration(ctx sdk.Context) time.Duration
	GetMasterAddressExternalKeyLockDuration(ctx sdk.Context) time.Duration
	GetVotingThreshold(ctx sdk.Context) utils.Threshold
	GetMinVoterCount(ctx sdk.Context) int64
	GetMaxTxSize(ctx sdk.Context) int64

	SetPendingOutpointInfo(ctx sdk.Context, key vote.PollKey, info OutPointInfo)
	GetPendingOutPointInfo(ctx sdk.Context, key vote.PollKey) (OutPointInfo, bool)
	DeletePendingOutPointInfo(ctx sdk.Context, key vote.PollKey)
	GetOutPointInfo(ctx sdk.Context, outPoint wire.OutPoint) (OutPointInfo, OutPointState, bool)
	DeleteOutpointInfo(ctx sdk.Context, outPoint wire.OutPoint)
	SetSpentOutpointInfo(ctx sdk.Context, info OutPointInfo)
	SetConfirmedOutpointInfo(ctx sdk.Context, keyID tss.KeyID, info OutPointInfo)
	GetConfirmedOutpointInfoQueueForKey(ctx sdk.Context, keyID tss.KeyID) utils.KVQueue

	SetUnsignedTx(ctx sdk.Context, tx UnsignedTx)
	GetUnsignedTx(ctx sdk.Context, txType TxType) (UnsignedTx, bool)
	DeleteUnsignedTx(ctx sdk.Context, txType TxType)
	SetSignedTx(ctx sdk.Context, tx SignedTx)
	GetSignedTx(ctx sdk.Context, txHash chainhash.Hash) (SignedTx, bool)
	SetLatestSignedTxHash(ctx sdk.Context, txType TxType, txHash chainhash.Hash)
	GetLatestSignedTxHash(ctx sdk.Context, txType TxType) (*chainhash.Hash, bool)

	SetAddress(ctx sdk.Context, address AddressInfo)
	GetAddress(ctx sdk.Context, encodedAddress string) (AddressInfo, bool)

	GetDustAmount(ctx sdk.Context, encodedAddress string) btcutil.Amount
	SetDustAmount(ctx sdk.Context, encodedAddress string, amount btcutil.Amount)
	DeleteDustAmount(ctx sdk.Context, encodedAddress string)

	SetUnconfirmedAmount(ctx sdk.Context, keyID tss.KeyID, amount btcutil.Amount)
	GetUnconfirmedAmount(ctx sdk.Context, keyID tss.KeyID) btcutil.Amount
}

// Voter is the interface that provides voting functionality
type Voter interface {
	InitializePoll(ctx sdk.Context, key vote.PollKey, voters []sdk.ValAddress, pollProperties ...vote.PollProperty) error
	// Deprecated: InitializePollWithSnapshot will be removed soon
	InitializePollWithSnapshot(ctx sdk.Context, key vote.PollKey, snapshotSeqNo int64, pollProperties ...vote.PollProperty) error
	GetPoll(ctx sdk.Context, pollKey vote.PollKey) vote.Poll
}

// InitPoller is a minimal interface to start a poll. This must be a type alias instead of a type definition,
// because the concrete implementation of Signer (specifically StartSign) is defined in a different package using another (identical)
// InitPoller interface. Go cannot match the types otherwise
type InitPoller = interface {
	// Deprecated: InitializePollWithSnapshot will be removed soon
	InitializePollWithSnapshot(ctx sdk.Context, key vote.PollKey, snapshotSeqNo int64, pollProperties ...vote.PollProperty) error
}

// Signer provides keygen and signing functionality
type Signer interface {
	StartSign(ctx sdk.Context, info exported.SignInfo, snapshotter Snapshotter, voter InitPoller) error
	SetSig(ctx sdk.Context, sigID string, signature []byte)
	GetSig(ctx sdk.Context, sigID string) (tss.Signature, tss.SigStatus)
	SetSigStatus(ctx sdk.Context, sigID string, status tss.SigStatus)
	SetInfoForSig(ctx sdk.Context, sigID string, info tss.SignInfo)
	GetCurrentKeyID(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool)
	GetCurrentKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool)
	GetNextKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool)
	GetSnapshotCounterForKeyID(ctx sdk.Context, keyID tss.KeyID) (int64, bool)
	SetKey(ctx sdk.Context, keyID tss.KeyID, key ecdsa.PublicKey)
	GetKey(ctx sdk.Context, keyID tss.KeyID) (tss.Key, bool)
	GetKeyForSigID(ctx sdk.Context, sigID string) (tss.Key, bool)
	AssignNextKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole, keyID tss.KeyID) error
	RotateKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) error
	AssertMatchesRequirements(ctx sdk.Context, snapshotter Snapshotter, chain nexus.Chain, keyID tss.KeyID, keyRole tss.KeyRole) error
	GetRotationCount(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) int64
	GetKeyByRotationCount(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole, rotationCount int64) (tss.Key, bool)
	GetRotationCountOfKeyID(ctx sdk.Context, keyID tss.KeyID) (int64, bool)
	GetKeyUnbondingLockingKeyRotationCount(ctx sdk.Context) int64
	GetOldActiveKeys(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) ([]tss.Key, error)
	GetExternalKeyIDs(ctx sdk.Context, chain nexus.Chain) ([]exported.KeyID, bool)
	GetExternalMultisigThreshold(ctx sdk.Context) utils.Threshold
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
	GetChainMaintainers(ctx sdk.Context, chain nexus.Chain) []sdk.ValAddress
	IsChainActivated(ctx sdk.Context, chain nexus.Chain) bool
}

// Snapshotter provides snapshot functionality
type Snapshotter = snapshot.Snapshotter
