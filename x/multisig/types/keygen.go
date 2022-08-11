package types

import (
	fmt "fmt"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// NewKeygenSession is the contructor for keygen session
func NewKeygenSession(id exported.KeyID, keygenThreshold utils.Threshold, signingThreshold utils.Threshold, snapshot snapshot.Snapshot, expiresAt int64, gracePeriod int64) KeygenSession {
	return KeygenSession{
		Key: Key{
			ID:               id,
			Snapshot:         snapshot,
			SigningThreshold: signingThreshold,
		},
		State:           exported.Pending,
		KeygenThreshold: keygenThreshold,
		ExpiresAt:       expiresAt,
		GracePeriod:     gracePeriod,
	}
}

// ValidateBasic returns an error if the given keygen session is invalid; nil otherwise
func (m KeygenSession) ValidateBasic() error {
	if m.KeygenThreshold.LT(m.Key.SigningThreshold) {
		return fmt.Errorf("keygen threshold must be >=signing threshold")
	}

	if m.ExpiresAt <= 0 {
		return fmt.Errorf("expires at must be >0")
	}

	if m.CompletedAt >= m.ExpiresAt {
		return fmt.Errorf("completed at must be < expires at")
	}

	if m.GracePeriod < 0 {
		return fmt.Errorf("grace period must be >=0")
	}

	switch m.GetState() {
	case exported.Pending:
		if m.CompletedAt != 0 {
			return fmt.Errorf("pending keygen session must not have completed at set")
		}

		if err := validateBasicPendingKey(m.Key); err != nil {
			return err
		}
	case exported.Completed:
		if m.CompletedAt <= 0 {
			return fmt.Errorf("completed keygen session must have completed at set")
		}

		if err := m.Key.ValidateBasic(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unexpected state %s", m.GetState())
	}

	return nil
}

// GetKeyID returns the key ID of the given keygen session
func (m KeygenSession) GetKeyID() exported.KeyID {
	return m.Key.ID
}

// AddKey adds a new public key for the given participant into the keygen session
func (m *KeygenSession) AddKey(blockHeight int64, participant sdk.ValAddress, pubKey exported.PublicKey) error {
	if m.Key.PubKeys == nil {
		m.Key.PubKeys = make(map[string]exported.PublicKey)
		m.IsPubKeyReceived = make(map[string]bool)
	}

	if m.isExpired(blockHeight) {
		return fmt.Errorf("keygen session %s has expired", m.GetKeyID())
	}

	if m.Key.Snapshot.GetParticipantWeight(participant).IsZero() {
		return fmt.Errorf("%s is not a participant of keygen %s", participant.String(), m.GetKeyID())
	}

	if _, ok := m.Key.PubKeys[participant.String()]; ok {
		return fmt.Errorf("participant %s already submitted its public key for keygen %s", participant.String(), m.GetKeyID())
	}

	if m.IsPubKeyReceived[pubKey.String()] {
		return fmt.Errorf("duplicate public key received")
	}

	if m.State == exported.Completed && !m.isWithinGracePeriod(blockHeight) {
		return fmt.Errorf("keygen session %s has closed", m.GetKeyID())
	}

	m.addKey(participant, pubKey)

	if m.State != exported.Completed && m.Key.GetParticipantsWeight().GTE(m.Key.Snapshot.CalculateMinPassingWeight(m.KeygenThreshold)) {
		m.CompletedAt = blockHeight
		m.State = exported.Completed
	}

	return nil
}

// GetMissingParticipants returns all participants who failed to submit their public keys
func (m KeygenSession) GetMissingParticipants() []sdk.ValAddress {
	participants := m.Key.Snapshot.GetParticipantAddresses()

	return slices.Filter(participants, func(p sdk.ValAddress) bool {
		_, ok := m.Key.PubKeys[p.String()]

		return !ok
	})
}

// Result returns the generated key if the session is completed and the key is valid
func (m KeygenSession) Result() (Key, error) {
	if m.GetState() != exported.Completed {
		return Key{}, fmt.Errorf("keygen %s is not completed yet", m.GetKeyID())
	}

	funcs.MustNoErr(m.Key.ValidateBasic())

	return m.Key, nil
}

func (m KeygenSession) isWithinGracePeriod(blockHeight int64) bool {
	return blockHeight <= m.CompletedAt+m.GracePeriod
}

func (m KeygenSession) isExpired(blockHeight int64) bool {
	return blockHeight >= m.ExpiresAt
}

func (m *KeygenSession) addKey(participant sdk.ValAddress, pubKey exported.PublicKey) {
	m.Key.PubKeys[participant.String()] = pubKey
	m.IsPubKeyReceived[pubKey.String()] = true
}

// GetParticipants returns the participants of the given key
func (m Key) GetParticipants() []sdk.ValAddress {
	return sortAddresses(
		slices.Map(maps.Keys(m.PubKeys), func(a string) sdk.ValAddress { return funcs.Must(sdk.ValAddressFromBech32(a)) }),
	)
}

// GetParticipantsWeight returns the total weight of all participants who have submitted their public keys
func (m Key) GetParticipantsWeight() sdk.Uint {
	return slices.Reduce(m.GetParticipants(), sdk.ZeroUint(), func(total sdk.Uint, p sdk.ValAddress) sdk.Uint {
		return total.Add(m.Snapshot.GetParticipantWeight(p))
	})
}

// GetMinPassingWeight returns the minimum amount of weights required for the
// key to sign
func (m Key) GetMinPassingWeight() sdk.Uint {
	return m.Snapshot.CalculateMinPassingWeight(m.SigningThreshold)
}

// GetPubKey returns the public key of the given participant
func (m Key) GetPubKey(p sdk.ValAddress) (exported.PublicKey, bool) {
	pubKey, ok := m.PubKeys[p.String()]

	return pubKey, ok
}

// GetWeight returns the weight of the given participant
func (m Key) GetWeight(p sdk.ValAddress) sdk.Uint {
	return m.Snapshot.GetParticipantWeight(p)
}

// GetHeight returns the height of the key snapshot
func (m Key) GetHeight() int64 {
	return m.Snapshot.Height
}

// GetTimestamp returns the timestamp of the key snapshot
func (m Key) GetTimestamp() time.Time {
	return m.Snapshot.Timestamp
}

// GetBondedWeight returns the bonded weight of the key snapshot
func (m Key) GetBondedWeight() sdk.Uint {
	return m.Snapshot.BondedWeight
}

// ValidateBasic returns an error if the given key is invalid; nil otherwise
func (m Key) ValidateBasic() error {
	if err := validateBasicPendingKey(m); err != nil {
		return err
	}

	if m.GetParticipantsWeight().LT(m.GetMinPassingWeight()) {
		return fmt.Errorf("invalid signing threshold")
	}

	return nil
}

func validateBasicPendingKey(key Key) error {
	if err := key.ID.ValidateBasic(); err != nil {
		return err
	}

	if err := key.Snapshot.ValidateBasic(); err != nil {
		return err
	}

	pubKeySeen := make(map[string]bool, len(key.PubKeys))
	for address, pubkey := range key.PubKeys {
		pubkeyStr := pubkey.String()
		if pubKeySeen[pubkeyStr] {
			return fmt.Errorf("duplicate public key seen")
		}
		pubKeySeen[pubkeyStr] = true

		p, err := sdk.ValAddressFromBech32(address)
		if err != nil {
			return err
		}

		if err := pubkey.ValidateBasic(); err != nil {
			return err
		}

		if key.Snapshot.GetParticipantWeight(p).IsZero() {
			return fmt.Errorf("invalid participant with public key submitted")
		}
	}

	return nil
}
