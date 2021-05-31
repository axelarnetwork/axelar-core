package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterLegacyAminoCodec registers concrete types on codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&LinkRequest{}, "ethereum/MsgLink", nil)
	cdc.RegisterConcrete(&VoteConfirmTokenRequest{}, "ethereum/VoteConfirmToken", nil)
	cdc.RegisterConcrete(&VoteConfirmDepositRequest{}, "ethereum/VoteConfirmDeposit", nil)
	cdc.RegisterConcrete(&ConfirmTokenRequest{}, "ethereum/ConfirmToken", nil)
	cdc.RegisterConcrete(&ConfirmDepositRequest{}, "ethereum/ConfirmDeposit", nil)
	cdc.RegisterConcrete(&SignTxRequest{}, "ethereum/SignTx", nil)
	cdc.RegisterConcrete(&SignPendingTransfersRequest{}, "ethereum/SignPendingTransfers", nil)
	cdc.RegisterConcrete(&SignDeployTokenRequest{}, "ethereum/SignDeployToken", nil)
	cdc.RegisterConcrete(&SignBurnTokensRequest{}, "ethereum/SignBurnTokens", nil)
	cdc.RegisterConcrete(&SignTransferOwnershipRequest{}, "ethereum/SignTransferOwnership", nil)
	cdc.RegisterConcrete(&AddChainRequest{}, "ethereum/AddChainRequest", nil)
}

// RegisterInterfaces registers types and interfaces with the given registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&LinkRequest{},
		&VoteConfirmTokenRequest{},
		&VoteConfirmDepositRequest{},
		&ConfirmTokenRequest{},
		&ConfirmDepositRequest{},
		&SignTxRequest{},
		&SignPendingTransfersRequest{},
		&SignDeployTokenRequest{},
		&SignBurnTokensRequest{},
		&SignTransferOwnershipRequest{},
		&AddChainRequest{},
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
