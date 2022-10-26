package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	reward "github.com/axelarnetwork/axelar-core/x/reward/exported"
)

// RegisterInterfaces registers types and interfaces with the given registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&StartKeygenRequest{},
		&SubmitPubKeyRequest{},
		&RotateKeyRequest{},
		&SubmitSignatureRequest{},
		&KeygenOptInRequest{},
		&KeygenOptOutRequest{},
	)

	registry.RegisterImplementations((*reward.Refundable)(nil),
		&SubmitPubKeyRequest{},
		&SubmitSignatureRequest{},
	)

	registry.RegisterImplementations((*codec.ProtoMarshaler)(nil),
		&MultiSig{},
	)
}
