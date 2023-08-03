package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
)

// RegisterLegacyAminoCodec registers concrete types on codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&LinkRequest{}, "evm/MsgLink", nil)
	cdc.RegisterConcrete(&ConfirmTokenRequest{}, "evm/ConfirmToken", nil)
	cdc.RegisterConcrete(&ConfirmDepositRequest{}, "evm/ConfirmDeposit", nil)
	cdc.RegisterConcrete(&ConfirmTransferKeyRequest{}, "evm/ConfirmTransferKey", nil)
	cdc.RegisterConcrete(&CreatePendingTransfersRequest{}, "evm/CreatePendingTransfers", nil)
	cdc.RegisterConcrete(&CreateDeployTokenRequest{}, "evm/CreateDeployToken", nil)
	cdc.RegisterConcrete(&CreateBurnTokensRequest{}, "evm/CreateBurnTokens", nil)
	cdc.RegisterConcrete(&CreateTransferOperatorshipRequest{}, "evm/CreateTransferOperatorship", nil)
	cdc.RegisterConcrete(&SignCommandsRequest{}, "evm/SignCommands", nil)
	cdc.RegisterConcrete(&AddChainRequest{}, "evm/AddChainRequest", nil)
	cdc.RegisterConcrete(&SetGatewayRequest{}, "evm/SetGatewayRequest", nil)
	cdc.RegisterConcrete(&ConfirmGatewayTxRequest{}, "evm/ConfirmGatewayTxRequest", nil)
	cdc.RegisterConcrete(&ConfirmGatewayTxsRequest{}, "evm/ConfirmGatewayTxsRequest", nil)
	cdc.RegisterConcrete(&RetryFailedEventRequest{}, "evm/RetryFailedEvent", nil)
}

// RegisterInterfaces registers types and interfaces with the given registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&LinkRequest{},
		&ConfirmTokenRequest{},
		&ConfirmDepositRequest{},
		&ConfirmTransferKeyRequest{},
		&CreatePendingTransfersRequest{},
		&CreateDeployTokenRequest{},
		&CreateBurnTokensRequest{},
		&CreateTransferOperatorshipRequest{},
		&SignCommandsRequest{},
		&AddChainRequest{},
		&SetGatewayRequest{},
		&ConfirmGatewayTxRequest{},
		&ConfirmGatewayTxsRequest{},
		&RetryFailedEventRequest{},
	)
	registry.RegisterImplementations((*codec.ProtoMarshaler)(nil),
		&gogoprototypes.BoolValue{},
		&SigMetadata{},
		&Event{},
		&VoteEvents{},
		&PollMetadata{},
	)
}

var amino = codec.NewLegacyAmino()

// ModuleCdc defines the module codec
var ModuleCdc = codec.NewAminoCodec(amino)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
