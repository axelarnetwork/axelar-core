package types

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	ibctypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	ibc "github.com/cosmos/ibc-go/v4/modules/core/exported"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

//go:generate moq -out ./mock/expected_keepers.go -pkg mock . BaseKeeper Nexus BankKeeper IBCTransferKeeper ChannelKeeper AccountKeeper PortKeeper GovKeeper FeegrantKeeper IBCKeeper

// BaseKeeper is implemented by this module's base keeper
type BaseKeeper interface {
	Logger(ctx sdk.Context) log.Logger
	GetParams(ctx sdk.Context) (params Params)
	GetRouteTimeoutWindow(ctx sdk.Context) uint64
	GetTransferLimit(ctx sdk.Context) uint64
	GetEndBlockerLimit(ctx sdk.Context) uint64
	GetCosmosChains(ctx sdk.Context) []nexus.ChainName
	GetCosmosChainByName(ctx sdk.Context, chain nexus.ChainName) (CosmosChain, bool)
	GetIBCPath(ctx sdk.Context, chain nexus.ChainName) (string, bool)
	GetChainNameByIBCPath(ctx sdk.Context, ibcPath string) (nexus.ChainName, bool)
	EnqueueIBCTransfer(ctx sdk.Context, transfer IBCTransfer) error
	GetIBCTransferQueue(ctx sdk.Context) utils.KVQueue
	SetSeqIDMapping(ctx sdk.Context, t IBCTransfer) error
	SetTransferFailed(ctx sdk.Context, transferID nexus.TransferID) error
}

// Nexus provides functionality to manage cross-chain transfers
type Nexus interface {
	EnqueueForTransfer(ctx sdk.Context, sender nexus.CrossChainAddress, amount sdk.Coin) (nexus.TransferID, error)
	EnqueueTransfer(ctx sdk.Context, senderChain nexus.Chain, recipient nexus.CrossChainAddress, asset sdk.Coin) (nexus.TransferID, error)
	GetTransfersForChainPaginated(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState, pageRequest *query.PageRequest) ([]nexus.CrossChainTransfer, *query.PageResponse, error)
	ArchivePendingTransfer(ctx sdk.Context, transfer nexus.CrossChainTransfer)
	GetChain(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool)
	LinkAddresses(ctx sdk.Context, sender nexus.CrossChainAddress, recipient nexus.CrossChainAddress) error
	IsAssetRegistered(ctx sdk.Context, chain nexus.Chain, denom string) bool
	RegisterAsset(ctx sdk.Context, chain nexus.Chain, asset nexus.Asset, limit sdk.Uint, window time.Duration) error
	GetRecipient(ctx sdk.Context, sender nexus.CrossChainAddress) (nexus.CrossChainAddress, bool)
	SetChain(ctx sdk.Context, chain nexus.Chain)
	GetTransferFees(ctx sdk.Context) sdk.Coins
	SubTransferFee(ctx sdk.Context, coin sdk.Coin)
	ActivateChain(ctx sdk.Context, chain nexus.Chain)
	GetChainByNativeAsset(ctx sdk.Context, asset string) (nexus.Chain, bool)
	IsChainActivated(ctx sdk.Context, chain nexus.Chain) bool
	RateLimitTransfer(ctx sdk.Context, chain nexus.ChainName, asset sdk.Coin, direction nexus.TransferDirection) error
	GetMessage(ctx sdk.Context, id string) (m nexus.GeneralMessage, found bool)
	SetNewMessage(ctx sdk.Context, m nexus.GeneralMessage) error
	SetMessageExecuted(ctx sdk.Context, id string) error
	SetMessageFailed(ctx sdk.Context, id string) error
	GenerateMessageID(ctx sdk.Context) (string, []byte, uint64)
	ValidateAddress(ctx sdk.Context, address nexus.CrossChainAddress) error
	RouteMessage(ctx sdk.Context, id string, routingCtx ...nexus.RoutingContext) error
}

// BankKeeper defines the expected interface contract the vesting module requires
// for creating vesting accounts with funds.
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	IsSendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool
	IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error
	BlockedAddr(addr sdk.AccAddress) bool
	SpendableBalance(ctx sdk.Context, address sdk.AccAddress, denom string) sdk.Coin
}

// IBCTransferKeeper provides functionality to manage IBC transfers
type IBCTransferKeeper interface {
	GetDenomTrace(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (ibctypes.DenomTrace, bool)
	SendTransfer(ctx sdk.Context, sourcePort, sourceChannel string, token sdk.Coin, sender sdk.AccAddress, receiver string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64) error
	Transfer(goCtx context.Context, msg *ibctypes.MsgTransfer) (*ibctypes.MsgTransferResponse, error)
}

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, ibc.ClientState, error)

	GetChannel(ctx sdk.Context, srcPort string, srcChan string) (channel channeltypes.Channel, found bool) // used in module_test
	SendPacket(ctx sdk.Context, channelCap *capabilitytypes.Capability, packet ibc.PacketI) error          // used in module_test
	WriteAcknowledgement(
		ctx sdk.Context,
		chanCap *capabilitytypes.Capability,
		packet ibc.PacketI,
		ack ibc.Acknowledgement,
	) error
	GetAppVersion(ctx sdk.Context, portID string, channelID string) (string, bool) // used in module_test
}

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/bank keeper.
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress

	GetModuleAccount(ctx sdk.Context, moduleName string) types.ModuleAccountI // used in module_test
}

// CosmosChainGetter exposes GetCosmosChainByName
type CosmosChainGetter func(ctx sdk.Context, chain nexus.ChainName) (CosmosChain, bool)

// PortKeeper used in module_test
type PortKeeper interface {
	BindPort(ctx sdk.Context, portID string) *capabilitytypes.Capability
}

// GovKeeper provides functionality to the gov module
type GovKeeper interface {
	GetProposal(ctx sdk.Context, proposalID uint64) (govtypes.Proposal, bool)
}

// FeegrantKeeper defines the expected feegrant keeper.
type FeegrantKeeper interface {
	UseGrantedFees(ctx sdk.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error
}

// IBCKeeper defines the expected IBC keeper
type IBCKeeper interface {
	SendMessage(c context.Context, recipient nexus.CrossChainAddress, asset sdk.Coin, payload string, id string) error
}
