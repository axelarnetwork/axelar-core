package types

import (
	fmt "fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ codectypes.UnpackInterfacesMessage = Sig{}

// NewSigningSession is the contructor for signing session
func NewSigningSession(id uint64, key Key, payloadHash Hash, expiresAt int64, gracePeriod int64, module string, moduleMetadataProto ...codec.ProtoMarshaler) SigningSession {
	var moduleMetadata *codectypes.Any
	if len(moduleMetadataProto) > 0 {
		moduleMetadata = funcs.Must(codectypes.NewAnyWithValue(moduleMetadataProto[0]))
	}

	return SigningSession{
		Signature: Sig{
			ID:             id,
			KeyID:          key.ID,
			PayloadHash:    payloadHash,
			Module:         module,
			ModuleMetadata: moduleMetadata,
		},
		Key:         key,
		ExpiresAt:   expiresAt,
		GracePeriod: gracePeriod,
	}
}

// ValidateBasic returns an error if the given signing session is invalid; nil otherwise
func (m SigningSession) ValidateBasic() error {
	if err := m.Signature.ValidateBasic(); err != nil {
		return err
	}

	if err := m.Key.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

// ValidateBasic returns an error if the given sig is invalid; nil otherwise
func (m Sig) ValidateBasic() error {
	if err := m.KeyID.ValidateBasic(); err != nil {
		return err
	}

	if err := m.PayloadHash.ValidateBasic(); err != nil {
		return nil
	}

	signatureSeen := make(map[string]bool, len(m.Sigs))
	for address, sig := range m.Sigs {
		sigHex := sig.String()
		if signatureSeen[sigHex] {
			return fmt.Errorf("duplicate signature seen")
		}
		signatureSeen[sigHex] = true

		if _, err := sdk.ValAddressFromBech32(address); err != nil {
			return err
		}

		if err := sig.ValidateBasic(); err != nil {
			return err
		}
	}

	if err := utils.ValidateString(m.Module); err != nil {
		return err
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m Sig) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var data codec.ProtoMarshaler

	return unpacker.UnpackAny(m.ModuleMetadata, &data)
}
