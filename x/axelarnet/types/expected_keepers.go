package types

import (
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	ibcclient "github.com/cosmos/ibc-go/modules/core/exported"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/modules/apps/transfer/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/log"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

//go:generate moq -out ./mock/expected_keepers.go -pkg mock . BaseKeeper  Nexus  BankKeeper IBCTransferKeeper ChannelKeeper AccountKeeper

// BaseKeeper is implemented by this module's base keeper
type BaseKeeper interface {
	Logger(ctx sdk.Context) log.Logger
	GetRouteTimeoutWindow(ctx sdk.Context) uint64
	GetTransactionFeeRate(ctx sdk.Context) sdk.Dec

	RegisterIBCPath(ctx sdk.Context, asset, path string) error
	GetIBCPath(ctx sdk.Context, chain string) (string, bool)
	GetFeeCollector(ctx sdk.Context) (sdk.AccAddress, bool)
	SetFeeCollector(ctx sdk.Context, address sdk.AccAddress) error
	SetPendingIBCTransfer(ctx sdk.Context, transfer IBCTransfer)
	GetPendingIBCTransfer(ctx sdk.Context, portID, channelID string, sequence uint64) (IBCTransfer, bool)
	DeletePendingIBCTransfer(ctx sdk.Context, portID, channelID string, sequence uint64)
	GetCosmosChains(ctx sdk.Context) []string
	GetCosmosChainByName(ctx sdk.Context, chain string) (CosmosChain, bool)
	SetCosmosChain(ctx sdk.Context, chain CosmosChain)
}

// Nexus provides functionality to manage cross-chain transfers
type Nexus interface {
	EnqueueForTransfer(ctx sdk.Context, sender nexus.CrossChainAddress, amount sdk.Coin, feeRate sdk.Dec) (nexus.TransferID, error)
	GetTransfersForChain(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState) []nexus.CrossChainTransfer
	ArchivePendingTransfer(ctx sdk.Context, transfer nexus.CrossChainTransfer)
	GetChain(ctx sdk.Context, chain string) (nexus.Chain, bool)
	LinkAddresses(ctx sdk.Context, sender nexus.CrossChainAddress, recipient nexus.CrossChainAddress) error
	IsAssetRegistered(ctx sdk.Context, chain nexus.Chain, denom string) bool
	RegisterAsset(ctx sdk.Context, chain nexus.Chain, asset nexus.Asset)
	GetRecipient(ctx sdk.Context, sender nexus.CrossChainAddress) (nexus.CrossChainAddress, bool)
	SetChain(ctx sdk.Context, chain nexus.Chain)
	GetTransferFees(ctx sdk.Context) sdk.Coins
	SubTransferFee(ctx sdk.Context, coin sdk.Coin)
	ActivateChain(ctx sdk.Context, chain nexus.Chain)
	GetChainByNativeAsset(ctx sdk.Context, asset string) (nexus.Chain, bool)
	RegisterNativeAsset(ctx sdk.Context, chain nexus.Chain, nativeAsset string) error
}

// BankKeeper defines the expected interface contract the vesting module requires
// for creating vesting accounts with funds.
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// IBCTransferKeeper provides functionality to manage IBC transfers
type IBCTransferKeeper interface {
	GetDenomTrace(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (ibctypes.DenomTrace, bool)
	SendTransfer(ctx sdk.Context, sourcePort, sourceChannel string, token sdk.Coin, sender sdk.AccAddress, receiver string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64) error
}

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, ibcclient.ClientState, error)
}

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/bank keeper.
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// CosmosChainGetter exposes GetCosmosChainByName
type CosmosChainGetter interface {
	GetCosmosChainByName(ctx sdk.Context, chain string) (CosmosChain, bool)
}
