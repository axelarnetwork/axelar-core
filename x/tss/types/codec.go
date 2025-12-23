// Package types provides legacy TSS type registrations for historical transaction decoding.
package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
)

// RegisterLegacyAminoCodec registers concrete types on codec.
// These registrations are required to decode historical transactions that used TSS message types.
// Block explorers and other tools querying historical data depend on these registrations.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&HeartBeatRequest{}, "tss/HeartBeatRequest", nil)
	cdc.RegisterConcrete(&UpdateParamsRequest{}, "tss/UpdateParams", nil)
}

// RegisterInterfaces registers types and interfaces with the given registry.
// No active message types remain as the TSS module's runtime functionality has been removed.
func RegisterInterfaces(_ cdctypes.InterfaceRegistry) {}

var amino = codec.NewLegacyAmino()

// ModuleCdc defines the module codec
var ModuleCdc = codec.NewAminoCodec(amino)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
