package exported_test

import (
	"bytes"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
	testRand "github.com/axelarnetwork/utils/test/rand"
)

func TestSnapshot(t *testing.T) {
	var (
		snapshot exported.Snapshot
	)

	givenSnapshot := Given("given any snapshot", func() {
		participantCount := rand.I64Between(1, 100)
		participants := make([]exported.Participant, participantCount)
		for i := range participants {
			participants[i] = exported.NewParticipant(rand.ValAddr(), sdk.NewUint(uint64(rand.I64Between(1, 100))))
		}

		snapshot = exported.NewSnapshot(
			testRand.Time(),
			rand.PosI64(),
			participants,
			sdk.NewUint(uint64(rand.I64Between(10000, 100000))),
		)
	})

	repeat := 20

	t.Run("ValidateBasic", func(t *testing.T) {
		givenSnapshot.
			When("it is valid", func() {}).
			Then("should return nil", func(t *testing.T) {
				assert.NoError(t, snapshot.ValidateBasic())
			}).
			Run(t, repeat)

		givenSnapshot.
			When("there is no participant", func() {
				snapshot.Participants = nil
			}).
			Then("should return error", func(t *testing.T) {
				assert.ErrorContains(t, snapshot.ValidateBasic(), "no participant")
			}).
			Run(t, repeat)

		givenSnapshot.
			When("bonded weight is zero", func() {
				snapshot.BondedWeight = sdk.ZeroUint()
			}).
			Then("should return error", func(t *testing.T) {
				assert.ErrorContains(t, snapshot.ValidateBasic(), "bonded weight >0")
			}).
			Run(t, repeat)

		givenSnapshot.
			When("height<=0", func() {
				snapshot.Height = -rand.I64Between(0, 100)
			}).
			Then("should return error", func(t *testing.T) {
				assert.ErrorContains(t, snapshot.ValidateBasic(), "height >0")
			}).
			Run(t, repeat)

		givenSnapshot.
			When("timestamp is not set", func() {
				snapshot.Timestamp = time.Time{}
			}).
			Then("should return error", func(t *testing.T) {
				assert.ErrorContains(t, snapshot.ValidateBasic(), "timestamp >0")
			}).
			Run(t, repeat)

		givenSnapshot.
			When("some participant has invalid address", func() {
				snapshot.Participants[rand.ValAddr().String()] = exported.Participant{Address: rand.Bytes(300)}
			}).
			Then("should return error", func(t *testing.T) {
				assert.ErrorContains(t, snapshot.ValidateBasic(), "address")
			}).
			Run(t, repeat)

		givenSnapshot.
			When("some participant does not have the correct key in the map", func() {
				snapshot.Participants[rand.ValAddr().String()] = exported.Participant{Address: rand.ValAddr()}
			}).
			Then("should return error", func(t *testing.T) {
				assert.ErrorContains(t, snapshot.ValidateBasic(), "invalid snapshot")
			}).
			Run(t, repeat)

		givenSnapshot.
			When("some participant does not have the correct key in the map", func() {
				address := rand.ValAddr()
				snapshot.Participants[address.String()] = exported.NewParticipant(address, snapshot.BondedWeight)
			}).
			Then("should return error", func(t *testing.T) {
				assert.ErrorContains(t, snapshot.ValidateBasic(), "participants weight greater than bonded weight")
			}).
			Run(t, repeat)
	})

	t.Run("CalculateMinPassingWeight", func(t *testing.T) {
		givenSnapshot.
			When("", func() {}).
			Then("should calculate correct minimum weight to pass the threshold", func(t *testing.T) {
				threshold := utils.OneThreshold
				assert.Equal(t, snapshot.BondedWeight, snapshot.CalculateMinPassingWeight(threshold))

				threshold = utils.NewThreshold(1, 3)
				snapshot.BondedWeight = sdk.NewUint(10)
				assert.Equal(t, sdk.NewUint(4), snapshot.CalculateMinPassingWeight(threshold))
			}).
			Run(t)
	})

	t.Run("GetParticipantAddresses", func(t *testing.T) {
		givenSnapshot.
			When("it is valid", func() {}).
			Then("should return addresses of all participants in asc order", func(t *testing.T) {
				actual := snapshot.GetParticipantAddresses()
				assert.ElementsMatch(t, maps.Keys(snapshot.Participants), slices.Map(actual, sdk.ValAddress.String))

				for i := 0; i < len(actual)-1; i++ {
					assert.True(t, bytes.Compare(actual[i], actual[i+1]) < 0)
				}
			}).
			Run(t, repeat)
	})

	t.Run("GetParticipantWeight", func(t *testing.T) {
		givenSnapshot.
			When("it is valid", func() {}).
			Then("should return the correct weight for the given participant", func(t *testing.T) {
				for address, participant := range snapshot.Participants {
					addr, err := sdk.ValAddressFromBech32(address)
					if err != nil {
						panic(err)
					}

					assert.Equal(t, participant.Weight, snapshot.GetParticipantWeight(addr))
				}

				assert.Equal(t, sdk.ZeroUint(), snapshot.GetParticipantWeight(rand.ValAddr()))
			}).
			Run(t, repeat)
	})

	t.Run("GetParticipantsWeight", func(t *testing.T) {
		givenSnapshot.
			When("participants weight is 10", func() {
				snapshot.Participants = map[string]exported.Participant{
					"1": exported.NewParticipant(rand.ValAddr(), sdk.NewUint(2)),
					"2": exported.NewParticipant(rand.ValAddr(), sdk.NewUint(3)),
					"3": exported.NewParticipant(rand.ValAddr(), sdk.NewUint(5)),
				}
			}).
			Then("should calculate the correct participants weight", func(t *testing.T) {
				expected := sdk.NewUint(10)
				actual := snapshot.GetParticipantsWeight()

				assert.Equal(t, expected, actual)
			}).
			Run(t)
	})
}

func TestGetValidatorIllegibilities(t *testing.T) {
	expected := []exported.ValidatorIllegibility{exported.Tombstoned, exported.Jailed, exported.MissedTooManyBlocks, exported.NoProxyRegistered, exported.TssSuspended, exported.ProxyInsuficientFunds}
	actual := exported.GetValidatorIllegibilities()

	assert.Equal(t, expected, actual)
}

func TestFilterIllegibilityForNewKey(t *testing.T) {
	for _, illegibility := range exported.GetValidatorIllegibilities() {
		actual := illegibility.FilterIllegibilityForNewKey()

		assert.NotEqual(t, exported.None, actual)
	}
}

func TestFilterIllegibilityForTssSigning(t *testing.T) {
	for _, illegibility := range exported.GetValidatorIllegibilities() {
		actual := illegibility.FilterIllegibilityForTssSigning()

		assert.NotEqual(t, exported.None, actual)
	}
}

func TestFilterIllegibilityForMultisigSigning(t *testing.T) {
	for _, illegibility := range exported.GetValidatorIllegibilities() {
		actual := illegibility.FilterIllegibilityForMultisigSigning()

		switch illegibility {
		case exported.MissedTooManyBlocks:
			assert.Equal(t, exported.None, actual)
		case exported.ProxyInsuficientFunds:
			assert.Equal(t, exported.None, actual)
		default:
			assert.NotEqual(t, exported.None, actual)
		}
	}
}
