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
	SetParams(ctx sdk.Context, n Nexus, p Params)
	GetRouteTimeoutWindow(ctx sdk.Context) uint64

	RegisterIBCPath(ctx sdk.Context, asset, path string) error
	GetIBCPath(ctx sdk.Context, chain string) (string, bool)
	GetPendingRefund(ctx sdk.Context, req RefundMsgRequest) (sdk.Coin, bool)
	DeletePendingRefund(ctx sdk.Context, req RefundMsgRequest)
	GetFeeCollector(ctx sdk.Context) (sdk.AccAddress, bool)
	SetFeeCollector(ctx sdk.Context, address sdk.AccAddress)
	SetPendingIBCTransfer(ctx sdk.Context, portID, channelID string, sequence uint64, value IBCTransfer)
	GetPendingIBCTransfer(ctx sdk.Context, portID, channelID string, sequence uint64) (IBCTransfer, bool)
	DeletePendingIBCTransfer(ctx sdk.Context, portID, channelID string, sequence uint64)
	GetCosmosChains(ctx sdk.Context) []string
	RegisterAssetToCosmosChain(ctx sdk.Context, asset string, chain string)
	GetCosmosChain(ctx sdk.Context, asset string) (string, bool)
}

// Nexus provides functionality to manage cross-chain transfers
type Nexus interface {
	EnqueueForTransfer(ctx sdk.Context, sender nexus.CrossChainAddress, amount sdk.Coin) error
	GetTransfersForChain(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState) []nexus.CrossChainTransfer
	ArchivePendingTransfer(ctx sdk.Context, transfer nexus.CrossChainTransfer)
	GetChain(ctx sdk.Context, chain string) (nexus.Chain, bool)
	IsAssetRegistered(ctx sdk.Context, chainName, denom string) bool
	RegisterAsset(ctx sdk.Context, chainName, denom string)
	LinkAddresses(ctx sdk.Context, sender nexus.CrossChainAddress, recipient nexus.CrossChainAddress)
	GetRecipient(ctx sdk.Context, sender nexus.CrossChainAddress) (nexus.CrossChainAddress, bool)
	AddToChainTotal(ctx sdk.Context, chain nexus.Chain, amount sdk.Coin)
	SetChain(ctx sdk.Context, chain nexus.Chain)
}

// BankKeeper defines the expected interface contract the vesting module requires
// for creating vesting accounts with funds.
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
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
