package app

import (
	"github.com/cosmos/cosmos-sdk/std"

	"github.com/axelarnetwork/axelar-core/app/codec"
	"github.com/axelarnetwork/axelar-core/app/params"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() params.EncodingConfig {
	encodingConfig := params.MakeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	GetModuleBasics().RegisterLegacyAminoCodec(encodingConfig.Amino)
	GetModuleBasics().RegisterInterfaces(encodingConfig.InterfaceRegistry)

	codec.RegisterLegacyMsgInterfaces(encodingConfig.InterfaceRegistry)

	// Register the AccountI interface for rosetta compatibility
	encodingConfig.InterfaceRegistry.RegisterInterface(
		"cosmos.auth.v1beta1.AccountI",
		(*sdkclient.Account)(nil),
		&authtypes.BaseAccount{},
		&authtypes.ModuleAccount{},
	)

	return encodingConfig
}
