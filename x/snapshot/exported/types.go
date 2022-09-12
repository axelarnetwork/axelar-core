package exported

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/utils/slices"
)

//go:generate moq -out ./mock/types.go -pkg mock . ValidatorI

// QuadraticWeightFunc returns floor(sqrt(consensusPower)) as the weight
func QuadraticWeightFunc(consensusPower sdk.Uint) sdk.Uint {
	bigInt := consensusPower.BigInt()

	return sdk.NewUintFromBigInt(bigInt.Sqrt(bigInt))
}

// ValidatorI provides necessary functions to the validator information
type ValidatorI interface {
	GetConsensusPower(sdk.Int) int64       // validation power in tendermint
	GetOperator() sdk.ValAddress           // operator address to receive/return validators coins
	GetConsAddr() (sdk.ConsAddress, error) // validation consensus address
	IsJailed() bool                        // whether the validator is jailed
	IsBonded() bool                        // whether the validator is bonded
}

// NewSnapshot is the constructor of Snapshot
func NewSnapshot(timestamp time.Time, height int64, participants []Participant, bondedWeight sdk.Uint) Snapshot {
	return Snapshot{
		Timestamp:    timestamp,
		Height:       height,
		Participants: slices.ToMap(participants, func(p Participant) string { return p.Address.String() }),
		BondedWeight: bondedWeight,
	}
}

// ValidateBasic returns an error if the given snapshot is invalid; nil otherwise
func (m Snapshot) ValidateBasic() error {
	if len(m.Participants) == 0 {
		return fmt.Errorf("snapshot cannot have no participant")
	}

	if m.BondedWeight.IsZero() {
		return fmt.Errorf("snapshot must have bonded weight >0")
	}

	if m.Height <= 0 {
		return fmt.Errorf("snapshot must have height >0")
	}

	if m.Timestamp.IsZero() {
		return fmt.Errorf("snapshot must have timestamp >0")
	}

	for addr, p := range m.Participants {
		if err := p.ValidateBasic(); err != nil {
			return err
		}

		if addr != p.Address.String() {
			return fmt.Errorf("invalid snapshot")
		}
	}

	if m.GetParticipantsWeight().GT(m.BondedWeight) {
		return fmt.Errorf("snapshot cannot have sum of participants weight greater than bonded weight")
	}

	return nil
}

// NewParticipant is the constructor of Participant
func NewParticipant(address sdk.ValAddress, weight sdk.Uint) Participant {
	return Participant{
		Address: address,
		Weight:  weight,
	}
}

// GetAddress returns the address of the participant
func (m Participant) GetAddress() sdk.ValAddress {
	return m.Address
}

// ValidateBasic returns an error if the given participant is invalid; nil otherwise
func (m Participant) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Address); err != nil {
		return err
	}

	return nil
}

// GetParticipantAddresses returns the addresses of all participants in the snapshot
func (m Snapshot) GetParticipantAddresses() []sdk.ValAddress {
	addresses := slices.Map(maps.Values(m.Participants), Participant.GetAddress)
	sort.SliceStable(addresses, func(i, j int) bool { return bytes.Compare(addresses[i], addresses[j]) < 0 })

	return addresses
}

// GetParticipantsWeight returns the sum of all participants' weights
func (m Snapshot) GetParticipantsWeight() sdk.Uint {
	weight := sdk.ZeroUint()
	for _, p := range m.Participants {
		weight = weight.Add(p.Weight)
	}

	return weight
}

// GetParticipantWeight returns the weight of the given participant
func (m Snapshot) GetParticipantWeight(participant sdk.ValAddress) sdk.Uint {
	if participant, ok := m.Participants[participant.String()]; ok {
		return participant.Weight
	}

	return sdk.ZeroUint()
}

// CalculateMinPassingWeight returns the minimum amount of weights to pass the given threshold
func (m Snapshot) CalculateMinPassingWeight(threshold utils.Threshold) sdk.Uint {
	minPassingWeight := m.BondedWeight.
		MulUint64(uint64(threshold.Numerator)).
		QuoUint64(uint64(threshold.Denominator))

	if minPassingWeight.MulUint64(uint64(threshold.Denominator)).GTE(m.BondedWeight.MulUint64(uint64(threshold.Numerator))) {
		return minPassingWeight
	}

	return minPassingWeight.AddUint64(1)
}
