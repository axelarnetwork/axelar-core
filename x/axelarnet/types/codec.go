package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/reward/exported"
)

// RegisterLegacyAminoCodec registers concrete types on codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&LinkRequest{}, "axelarnet/Link", nil)
	cdc.RegisterConcrete(&ConfirmDepositRequest{}, "axelarnet/ConfirmDeposit", nil)
	cdc.RegisterConcrete(&ExecutePendingTransfersRequest{}, "axelarnet/ExecutePendingTransfers", nil)
	cdc.RegisterConcrete(&RegisterIBCPathRequest{}, "axelarnet/RegisterIBCPath", nil)
	cdc.RegisterConcrete(&AddCosmosBasedChainRequest{}, "axelarnet/AddCosmosBasedChain", nil)
	cdc.RegisterConcrete(&RegisterAssetRequest{}, "axelarnet/RegisterAsset", nil)
	cdc.RegisterConcrete(&RouteIBCTransfersRequest{}, "axelarnet/RouteIBCTransfers", nil)
	cdc.RegisterConcrete(&RegisterFeeCollectorRequest{}, "axelarnet/RegisterFeeCollector", nil)
	cdc.RegisterConcrete(&RetryFailedTransferRequest{}, "axelarnet/RetryFailedTransfer", nil)
}

// RegisterInterfaces registers types and interfaces with the given registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&LinkRequest{},
		&ConfirmDepositRequest{},
		&ExecutePendingTransfersRequest{},
		&RegisterIBCPathRequest{},
		&AddCosmosBasedChainRequest{},
		&RegisterAssetRequest{},
		&RouteIBCTransfersRequest{},
		&RegisterFeeCollectorRequest{},
		&RetryFailedTransferRequest{},
	)
	registry.RegisterInterface("reward.v1beta1.Refundable",
		(*exported.Refundable)(nil))
}

var amino = codec.NewLegacyAmino()

// ModuleCdc defines the module codec
var ModuleCdc = codec.NewAminoCodec(amino)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
