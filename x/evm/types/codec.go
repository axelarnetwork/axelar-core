package types

import (
	exported "github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
)

// RegisterLegacyAminoCodec registers concrete types on codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&LinkRequest{}, "evm/MsgLink", nil)
	cdc.RegisterConcrete(&VoteConfirmTokenRequest{}, "evm/VoteConfirmToken", nil)
	cdc.RegisterConcrete(&VoteConfirmDepositRequest{}, "evm/VoteConfirmDeposit", nil)
	cdc.RegisterConcrete(&ConfirmTokenRequest{}, "evm/ConfirmToken", nil)
	cdc.RegisterConcrete(&ConfirmDepositRequest{}, "evm/ConfirmDeposit", nil)
	cdc.RegisterConcrete(&SignTxRequest{}, "evm/SignTx", nil)
	cdc.RegisterConcrete(&SignPendingTransfersRequest{}, "evm/SignPendingTransfers", nil)
	cdc.RegisterConcrete(&SignDeployTokenRequest{}, "evm/SignDeployToken", nil)
	cdc.RegisterConcrete(&SignBurnTokensRequest{}, "evm/SignBurnTokens", nil)
	cdc.RegisterConcrete(&SignTransferOwnershipRequest{}, "evm/SignTransferOwnership", nil)
	cdc.RegisterConcrete(&AddChainRequest{}, "evm/AddChainRequest", nil)
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
	registry.RegisterImplementations((*exported.VotingData)(nil),
		&gogoprototypes.BoolValue{},
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
