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
	cdc.RegisterConcrete(&VoteConfirmOutpointRequest{}, "bitcoin/VoteConfirmOutpoint", nil)
	cdc.RegisterConcrete(&ConfirmOutpointRequest{}, "bitcoin/ConfirmOutpoint", nil)
	cdc.RegisterConcrete(&LinkRequest{}, "bitcoin/Link", nil)
	cdc.RegisterConcrete(&SignPendingTransfersRequest{}, "bitcoin/SignPendingTransfers", nil)
	cdc.RegisterConcrete(&SignMasterConsolidationTransactionRequest{}, "bitcoin/SignMasterConsolidationTransaction", nil)
	cdc.RegisterConcrete(&RegisterExternalKeyRequest{}, "bitcoin/RegisterExternalKey", nil)
}

// RegisterInterfaces registers types and interfaces with the given registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&VoteConfirmOutpointRequest{},
		&ConfirmOutpointRequest{},
		&LinkRequest{},
		&SignPendingTransfersRequest{},
		&SignMasterConsolidationTransactionRequest{},
		&RegisterExternalKeyRequest{},
	)
	registry.RegisterImplementations((*codec.ProtoMarshaler)(nil),
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
