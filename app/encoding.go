package app

import (
	"github.com/cosmos/cosmos-sdk/std"

	"github.com/axelarnetwork/axelar-core/app/codec"
	"github.com/axelarnetwork/axelar-core/app/params"
)

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() params.EncodingConfig {
	encodingConfig := params.MakeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	GetModuleBasics().RegisterLegacyAminoCodec(encodingConfig.Amino)
	GetModuleBasics().RegisterInterfaces(encodingConfig.InterfaceRegistry)

	codec.RegisterLegacyMsgInterfaces(encodingConfig.InterfaceRegistry)

	return encodingConfig
}
