package types

import (
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

var _ codectypes.UnpackInterfacesMessage = MultiSig{}

// NewSigningSession is the contructor for signing session
func NewSigningSession(id uint64, key Key, payloadHash Hash, expiresAt int64, gracePeriod int64, module string, moduleMetadataProto ...codec.ProtoMarshaler) SigningSession {
	var moduleMetadata *codectypes.Any
	if len(moduleMetadataProto) > 0 {
		moduleMetadata = funcs.Must(codectypes.NewAnyWithValue(moduleMetadataProto[0]))
	}

	return SigningSession{
		MultiSig: MultiSig{
			ID:             id,
			KeyID:          key.ID,
			PayloadHash:    payloadHash,
			Module:         module,
			ModuleMetadata: moduleMetadata,
		},
		State:       exported.Pending,
		Key:         key,
		ExpiresAt:   expiresAt,
		GracePeriod: gracePeriod,
	}
}

// ValidateBasic returns an error if the given signing session is invalid; nil otherwise
func (m SigningSession) ValidateBasic() error {
	if err := m.MultiSig.ValidateBasic(); err != nil {
		return err
	}

	if err := m.Key.ValidateBasic(); err != nil {
		return err
	}

	if m.Key.ID != m.MultiSig.KeyID {
		return fmt.Errorf("key ID mismatch")
	}

	if m.ExpiresAt <= 0 {
		return fmt.Errorf("expires at must be >0")
	}

	switch m.State {
	case exported.Pending:
		if m.CompletedAt != 0 {
			return fmt.Errorf("pending signing session must not have completed at set")
		}
	case exported.Completed:
		if m.CompletedAt == 0 {
			return fmt.Errorf("completed signing session must have completed at set")
		}

		if m.getParticipantsWeight().LT(m.Key.Snapshot.CalculateMinPassingWeight(m.Key.SigningThreshold)) {
			return fmt.Errorf("completed signing session must have completed multi signature")
		}

	default:
		return fmt.Errorf("unexpected state %s", m.State)
	}

	for addr, sig := range m.MultiSig.Sigs {
		pubKey, ok := m.Key.PubKeys[addr]
		if !ok {
			return fmt.Errorf("participant %s does not have public key submitted", addr)
		}

		if !sig.Verify(m.MultiSig.PayloadHash, pubKey) {
			return fmt.Errorf("signature does not match the public key")
		}
	}

	return nil
}

func (m SigningSession) getParticipantsWeight() sdk.Uint {
	return slices.Reduce(m.MultiSig.getParticipants(), sdk.ZeroUint(), func(total sdk.Uint, p sdk.ValAddress) sdk.Uint {
		return total.Add(m.Key.Snapshot.GetParticipantWeight(p))
	})
}

// ValidateBasic returns an error if the given sig is invalid; nil otherwise
func (m MultiSig) ValidateBasic() error {
	if err := m.KeyID.ValidateBasic(); err != nil {
		return err
	}

	if err := m.PayloadHash.ValidateBasic(); err != nil {
		return err
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
func (m MultiSig) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var data codec.ProtoMarshaler

	return unpacker.UnpackAny(m.ModuleMetadata, &data)
}

func (m MultiSig) getParticipants() []sdk.ValAddress {
	return sortAddresses(
		slices.Map(maps.Keys(m.Sigs), func(a string) sdk.ValAddress { return funcs.Must(sdk.ValAddressFromBech32(a)) }),
	)
}
