package types

import (
	"crypto/ecdsa"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tendermint/tendermint/libs/log"

	params "github.com/cosmos/cosmos-sdk/x/params/types"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	votetypes "github.com/axelarnetwork/axelar-core/x/vote/types"
)

//go:generate moq -out ./mock/expected_keepers.go -pkg mock . TSS Voter Signer Nexus Snapshotter BaseKeeper ChainKeeper

// BaseKeeper is implemented by this module's base keeper
type BaseKeeper interface {
	Logger(ctx sdk.Context) log.Logger

	GetParams(ctx sdk.Context) []Params
	SetParams(ctx sdk.Context, params ...Params)

	ForChain(ctx sdk.Context, chain string) ChainKeeper
	SetPendingChain(ctx sdk.Context, chain nexus.Chain)
	GetPendingChain(ctx sdk.Context, chain string) (nexus.Chain, bool)
	DeletePendingChain(ctx sdk.Context, chain string)
}

// ChainKeeper is implemented by this module's chain keeper
type ChainKeeper interface {
	Logger(ctx sdk.Context) log.Logger

	GetName() string
	AssembleEthTx(ctx sdk.Context, txID string, pk ecdsa.PublicKey, sig tss.Signature) (*ethTypes.Transaction, error)
	GetCommandData(ctx sdk.Context, commandID CommandID) []byte
	GetNetwork(ctx sdk.Context) (string, bool)
	GetRequiredConfirmationHeight(ctx sdk.Context) (uint64, bool)
	GetRevoteLockingPeriod(ctx sdk.Context) (int64, bool)
	GetGatewayByteCodes(ctx sdk.Context) ([]byte, bool)
	GetBurnerByteCodes(ctx sdk.Context) ([]byte, bool)
	GetTokenByteCodes(ctx sdk.Context) ([]byte, bool)
	GetGatewayAddress(ctx sdk.Context) (common.Address, bool)
	GetTokenAddress(ctx sdk.Context, symbol string, gatewayAddr common.Address) (common.Address, error)
	SetPendingTokenDeployment(ctx sdk.Context, poll vote.PollMeta, tokenDeploy ERC20TokenDeployment)
	GetDeposit(ctx sdk.Context, txID common.Hash, burnerAddr common.Address) (ERC20Deposit, DepositState, bool)
	GetBurnerInfo(ctx sdk.Context, address common.Address) *BurnerInfo
	SetPendingDeposit(ctx sdk.Context, poll vote.PollMeta, deposit *ERC20Deposit)
	GetBurnerAddressAndSalt(ctx sdk.Context, tokenAddr common.Address, recipient string, gatewayAddr common.Address) (common.Address, common.Hash, error)
	SetBurnerInfo(ctx sdk.Context, burnerAddr common.Address, burnerInfo *BurnerInfo)
	GetPendingDeposit(ctx sdk.Context, poll vote.PollMeta) (ERC20Deposit, bool)
	DeletePendingDeposit(ctx sdk.Context, poll vote.PollMeta)
	DeleteDeposit(ctx sdk.Context, deposit ERC20Deposit)
	SetDeposit(ctx sdk.Context, deposit ERC20Deposit, state DepositState)
	GetPendingTokenDeployment(ctx sdk.Context, poll vote.PollMeta) (ERC20TokenDeployment, bool)
	DeletePendingToken(ctx sdk.Context, poll vote.PollMeta)
	SetCommandData(ctx sdk.Context, commandID CommandID, commandData []byte)
	SetTokenInfo(ctx sdk.Context, msg *SignDeployTokenRequest)
	GetConfirmedDeposits(ctx sdk.Context) []ERC20Deposit
	SetUnsignedTx(ctx sdk.Context, txID string, tx *ethTypes.Transaction)
	GetHashToSign(ctx sdk.Context, txID string) (common.Hash, error)
	SetGatewayAddress(ctx sdk.Context, addr common.Address)
	GetPendingTransferOwnership(ctx sdk.Context, poll vote.PollMeta) (TransferOwnership, bool)
	SetPendingTransferOwnership(ctx sdk.Context, poll vote.PollMeta, transferOwnership *TransferOwnership)
	GetArchivedTransferOwnership(ctx sdk.Context, poll vote.PollMeta) (TransferOwnership, bool)
	ArchiveTransferOwnership(ctx sdk.Context, poll vote.PollMeta)
	GetNetworkByID(ctx sdk.Context, id *big.Int) (string, bool)
	GetChainIDByNetwork(ctx sdk.Context, network string) *big.Int
}

// ParamsKeeper represents a global paramstore
type ParamsKeeper interface {
	Subspace(s string) params.Subspace
	GetSubspace(s string) (params.Subspace, bool)
}

// TSS exposes key functionality
type TSS interface {
	SetKeyRequirement(ctx sdk.Context, keyRequirement tss.KeyRequirement)
}

// Voter exposes voting functionality
type Voter interface {
	InitPoll(ctx sdk.Context, poll vote.PollMeta, snapshotCounter int64, expireAt int64) error
	DeletePoll(ctx sdk.Context, poll vote.PollMeta)
	TallyVote(ctx sdk.Context, sender sdk.AccAddress, pollMeta vote.PollMeta, data vote.VotingData) (*votetypes.Poll, error)
}

// Nexus provides functionality to manage cross-chain transfers
type Nexus interface {
	LinkAddresses(ctx sdk.Context, sender nexus.CrossChainAddress, recipient nexus.CrossChainAddress)
	GetRecipient(ctx sdk.Context, sender nexus.CrossChainAddress) (nexus.CrossChainAddress, bool)
	EnqueueForTransfer(ctx sdk.Context, sender nexus.CrossChainAddress, amount sdk.Coin) error
	GetTransfersForChain(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState) []nexus.CrossChainTransfer
	ArchivePendingTransfer(ctx sdk.Context, transfer nexus.CrossChainTransfer)
	SetChain(ctx sdk.Context, chain nexus.Chain)
	GetChain(ctx sdk.Context, chain string) (nexus.Chain, bool)
	IsAssetRegistered(ctx sdk.Context, chainName, denom string) bool
	RegisterAsset(ctx sdk.Context, chainName, denom string)
}

// InitPoller is a minimal interface to start a poll. This must be a type alias instead of a type definition,
// because the concrete implementation of Signer (specifically StartSign) is defined in a different package using another (identical)
// InitPoller interface. Go cannot match the types otherwise
type InitPoller = interface {
	InitPoll(ctx sdk.Context, poll vote.PollMeta, snapshotCounter int64, expireAt int64) error
}

// Signer provides keygen and signing functionality
type Signer interface {
	StartSign(ctx sdk.Context, initPoll InitPoller, keyID string, sigID string, msg []byte, snapshot snapshot.Snapshot) error
	GetSig(ctx sdk.Context, sigID string) (tss.Signature, bool)
	GetKey(ctx sdk.Context, keyID string) (tss.Key, bool)
	GetCurrentKeyID(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (string, bool)
	GetCurrentKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool)
	GetNextKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool)
	GetKeyForSigID(ctx sdk.Context, sigID string) (tss.Key, bool)
	GetSnapshotCounterForKeyID(ctx sdk.Context, keyID string) (int64, bool)
	AssignNextKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole, keyID string) error
	AssertMatchesRequirements(ctx sdk.Context, snapshot snapshot.Snapshot, chain nexus.Chain, keyID string, keyRole tss.KeyRole) error
}

// Snapshotter provides access to the snapshot functionality
type Snapshotter interface {
	GetSnapshot(ctx sdk.Context, counter int64) (snapshot.Snapshot, bool)
	GetLatestCounter(ctx sdk.Context) int64
	TakeSnapshot(ctx sdk.Context, subsetSize int64, keyShareDistributionPolicy tss.KeyShareDistributionPolicy) (snapshotConsensusPower sdk.Int, totalConsensusPower sdk.Int, err error)
}
