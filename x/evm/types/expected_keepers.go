package types

import (
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

//go:generate moq -out ./mock/expected_keepers.go -pkg mock . TSS Voter Signer Nexus Snapshotter EVMKeeper

// EVMKeeper is implemented by this module's keeper
type EVMKeeper interface {
	Logger(ctx sdk.Context) log.Logger

	GetParams(ctx sdk.Context) []Params
	SetParams(ctx sdk.Context, params ...Params)
	GetNetwork(ctx sdk.Context, chain string) (string, bool)
	GetRequiredConfirmationHeight(ctx sdk.Context, chain string) (uint64, bool)
	GetRevoteLockingPeriod(ctx sdk.Context, chain string) (int64, bool)
	GetGatewayByteCodes(ctx sdk.Context, chain string) ([]byte, bool)

	GetGatewayAddress(ctx sdk.Context, chain string) (common.Address, bool)
	GetTokenAddress(ctx sdk.Context, chain, symbol string, gatewayAddr common.Address) (common.Address, error)
	SetPendingTokenDeployment(ctx sdk.Context, chain string, poll vote.PollMeta, tokenDeploy ERC20TokenDeployment)
	GetDeposit(ctx sdk.Context, chain string, txID common.Hash, burnerAddr common.Address) (ERC20Deposit, DepositState, bool)
	GetBurnerInfo(ctx sdk.Context, chain string, address common.Address) *BurnerInfo
	SetPendingDeposit(ctx sdk.Context, chain string, poll vote.PollMeta, deposit *ERC20Deposit)
	GetBurnerAddressAndSalt(ctx sdk.Context, chain string, tokenAddr common.Address, recipient string, gatewayAddr common.Address) (common.Address, common.Hash, error)
	SetBurnerInfo(ctx sdk.Context, chain string, burnerAddr common.Address, burnerInfo *BurnerInfo)
	GetPendingDeposit(ctx sdk.Context, chain string, poll vote.PollMeta) (ERC20Deposit, bool)
	DeletePendingDeposit(ctx sdk.Context, chain string, poll vote.PollMeta)
	DeleteDeposit(ctx sdk.Context, chain string, deposit ERC20Deposit)
	SetDeposit(ctx sdk.Context, chain string, deposit ERC20Deposit, state DepositState)
	GetPendingTokenDeployment(ctx sdk.Context, chain string, poll vote.PollMeta) (ERC20TokenDeployment, bool)
	DeletePendingToken(ctx sdk.Context, chain string, poll vote.PollMeta)
	SetCommandData(ctx sdk.Context, chain string, commandID CommandID, commandData []byte)
	SetTokenInfo(ctx sdk.Context, chain string, msg *SignDeployTokenRequest)
	GetConfirmedDeposits(ctx sdk.Context, chain string) []ERC20Deposit
	SetUnsignedTx(ctx sdk.Context, chain, txID string, tx *ethTypes.Transaction)
	GetHashToSign(ctx sdk.Context, chain, txID string) (common.Hash, error)
	SetGatewayAddress(ctx sdk.Context, chain string, addr common.Address)
	DeletePendingChain(ctx sdk.Context, chain string)
	SetPendingChain(ctx sdk.Context, chain nexus.Chain)
	GetPendingChain(ctx sdk.Context, chain string) (nexus.Chain, bool)
	GetPendingTransferOwnership(ctx sdk.Context, chain string, poll vote.PollMeta) (TransferOwnership, bool)
	SetPendingTransferOwnership(ctx sdk.Context, chain string, poll vote.PollMeta, transferOwnership *TransferOwnership)
	DeletePendingTransferOwnership(ctx sdk.Context, chain string, poll vote.PollMeta)
	GetNetworkByID(ctx sdk.Context, chain string, id *big.Int) (string, bool)
	GetChainIDByNetwork(ctx sdk.Context, chain, network string) *big.Int
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
}

// Snapshotter provides access to the snapshot functionality
type Snapshotter interface {
	GetSnapshot(ctx sdk.Context, counter int64) (snapshot.Snapshot, bool)
}
