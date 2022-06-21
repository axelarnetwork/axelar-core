package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	ibcclient "github.com/cosmos/ibc-go/v2/modules/core/exported"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

//go:generate moq -out ./mock/expected_keepers.go -pkg mock . BaseKeeper  Nexus  BankKeeper IBCTransferKeeper ChannelKeeper AccountKeeper

// BaseKeeper is implemented by this module's base keeper
type BaseKeeper interface {
	Logger(ctx sdk.Context) log.Logger
	GetRouteTimeoutWindow(ctx sdk.Context) uint64

	RegisterIBCPath(ctx sdk.Context, chain nexus.ChainName, path string) error
	GetIBCPath(ctx sdk.Context, chain nexus.ChainName) (string, bool)
	GetFeeCollector(ctx sdk.Context) (sdk.AccAddress, bool)
	SetFeeCollector(ctx sdk.Context, address sdk.AccAddress) error
	GetCosmosChains(ctx sdk.Context) []nexus.ChainName
	GetCosmosChainByName(ctx sdk.Context, chain nexus.ChainName) (CosmosChain, bool)
	SetCosmosChain(ctx sdk.Context, chain CosmosChain)
	EnqueueTransfer(ctx sdk.Context, transfer IBCTransfer) error
	GetIBCTransferQueue(ctx sdk.Context) utils.KVQueue
}

// Nexus provides functionality to manage cross-chain transfers
type Nexus interface {
	EnqueueForTransfer(ctx sdk.Context, sender nexus.CrossChainAddress, amount sdk.Coin) (nexus.TransferID, error)
	GetTransfersForChain(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState) []nexus.CrossChainTransfer
	ArchivePendingTransfer(ctx sdk.Context, transfer nexus.CrossChainTransfer)
	GetChain(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool)
	LinkAddresses(ctx sdk.Context, sender nexus.CrossChainAddress, recipient nexus.CrossChainAddress) error
	IsAssetRegistered(ctx sdk.Context, chain nexus.Chain, denom string) bool
	RegisterAsset(ctx sdk.Context, chain nexus.Chain, asset nexus.Asset) error
	GetRecipient(ctx sdk.Context, sender nexus.CrossChainAddress) (nexus.CrossChainAddress, bool)
	SetChain(ctx sdk.Context, chain nexus.Chain)
	GetTransferFees(ctx sdk.Context) sdk.Coins
	SubTransferFee(ctx sdk.Context, coin sdk.Coin)
	ActivateChain(ctx sdk.Context, chain nexus.Chain)
	GetChainByNativeAsset(ctx sdk.Context, asset string) (nexus.Chain, bool)
	IsChainActivated(ctx sdk.Context, chain nexus.Chain) bool
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
type CosmosChainGetter func(ctx sdk.Context, chain nexus.ChainName) (CosmosChain, bool)
