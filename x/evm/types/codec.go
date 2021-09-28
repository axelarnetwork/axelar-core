package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"

	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
)

// RegisterLegacyAminoCodec registers concrete types on codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&LinkRequest{}, "evm/MsgLink", nil)
	cdc.RegisterConcrete(&VoteConfirmTokenRequest{}, "evm/VoteConfirmToken", nil)
	cdc.RegisterConcrete(&VoteConfirmDepositRequest{}, "evm/VoteConfirmDeposit", nil)
	cdc.RegisterConcrete(&VoteConfirmChainRequest{}, "evm/VoteConfirmChain", nil)
	cdc.RegisterConcrete(&VoteConfirmTransferKeyRequest{}, "evm/VoteConfirmTransferKey", nil)
	cdc.RegisterConcrete(&ConfirmTokenRequest{}, "evm/ConfirmToken", nil)
	cdc.RegisterConcrete(&ConfirmDepositRequest{}, "evm/ConfirmDeposit", nil)
	cdc.RegisterConcrete(&ConfirmChainRequest{}, "evm/ConfirmChain", nil)
	cdc.RegisterConcrete(&ConfirmTransferKeyRequest{}, "evm/ConfirmTransferKey", nil)
	cdc.RegisterConcrete(&SignTxRequest{}, "evm/SignTx", nil)
	cdc.RegisterConcrete(&CreatePendingTransfersRequest{}, "evm/CreatePendingTransfers", nil)
	cdc.RegisterConcrete(&CreateDeployTokenRequest{}, "evm/CreateDeployToken", nil)
	cdc.RegisterConcrete(&CreateBurnTokensRequest{}, "evm/CreateBurnTokens", nil)
	cdc.RegisterConcrete(&CreateTransferOwnershipRequest{}, "evm/CreateTransferOwnership", nil)
	cdc.RegisterConcrete(&CreateTransferOperatorshipRequest{}, "evm/CreateTransferOperatorship", nil)
	cdc.RegisterConcrete(&SignCommandsRequest{}, "evm/SignCommands", nil)
	cdc.RegisterConcrete(&AddChainRequest{}, "evm/AddChainRequest", nil)
}

// RegisterInterfaces registers types and interfaces with the given registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&LinkRequest{},
		&VoteConfirmTokenRequest{},
		&VoteConfirmDepositRequest{},
		&VoteConfirmChainRequest{},
		&VoteConfirmTransferKeyRequest{},
		&ConfirmTokenRequest{},
		&ConfirmDepositRequest{},
		&ConfirmChainRequest{},
		&ConfirmTransferKeyRequest{},
		&SignTxRequest{},
		&CreatePendingTransfersRequest{},
		&CreateDeployTokenRequest{},
		&CreateBurnTokensRequest{},
		&CreateTransferOwnershipRequest{},
		&CreateTransferOperatorshipRequest{},
		&SignCommandsRequest{},
		&AddChainRequest{},
	)
	registry.RegisterImplementations((*codec.ProtoMarshaler)(nil),
		&gogoprototypes.BoolValue{},
	)

	registry.RegisterImplementations((*axelarnet.Refundable)(nil),
		&VoteConfirmTokenRequest{},
		&VoteConfirmDepositRequest{},
		&VoteConfirmChainRequest{},
		&VoteConfirmTransferKeyRequest{},
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
