package testutils

import (
	"time"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	utilstestutils "github.com/axelarnetwork/axelar-core/utils/testutils"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigtestutils "github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	snapshottypes "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapshottestutils "github.com/axelarnetwork/axelar-core/x/snapshot/exported/testutils"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// PublicKey returns a random public key
func PublicKey() exported.PublicKey {
	privKey, _ := btcec.NewPrivateKey(btcec.S256())

	return privKey.PubKey().SerializeCompressed()
}

// Key returns a random multisig key
func Key() types.Key {
	threshold := utilstestutils.RandThreshold()
	snapshot := snapshottestutils.Snapshot(uint64(rand.I64Between(10, 20)), threshold)
	participants := snapshot.GetParticipantAddresses()
	pubKeys := make(map[string]exported.PublicKey, len(participants))

	for _, p := range participants {
		pubKeys[p.String()] = PublicKey()
	}

	return types.Key{
		ID:               multisigtestutils.KeyID(),
		Snapshot:         snapshot,
		PubKeys:          pubKeys,
		SigningThreshold: threshold,
	}
}

// KeyWithMissingParticipants returns a random multisig key with some missing participants
func KeyWithMissingParticipants() types.Key {
	participantCount := uint64(rand.I64Between(10, 20))
	missingCount := uint64(rand.I64Between(1, int64(participantCount)))
	participants := slices.Expand(func(_ int) snapshottypes.Participant {
		return snapshottypes.Participant{Address: rand.ValAddr(), Weight: sdk.NewUint(uint64(rand.I64Between(1, 100)))}
	}, int(participantCount))

	missingParticipants := slices.Expand(func(_ int) snapshottypes.Participant {
		return snapshottypes.Participant{Address: rand.ValAddr(), Weight: sdk.NewUint(uint64(rand.I64Between(1, 100)))}
	}, int(missingCount))

	participants = append(participants, missingParticipants...)
	weightAdder := func(total sdk.Uint, p snapshottypes.Participant) sdk.Uint { total = total.Add(p.Weight); return total }
	participantWeight := slices.Reduce(participants, sdk.ZeroUint(), weightAdder)
	missingParticipantWeight := slices.Reduce(missingParticipants, sdk.ZeroUint(), weightAdder)

	bondedWeight := rand.UintBetween(participantWeight, participantWeight.MulUint64(2))
	threshold := utils.NewThreshold(rand.I64Between(1, participantWeight.Sub(missingParticipantWeight).BigInt().Int64()), bondedWeight.BigInt().Int64())

	snapshot := snapshottypes.NewSnapshot(time.Now(), rand.I64Between(1, 100), participants, bondedWeight)

	pubKeys := make(map[string]exported.PublicKey, len(participants))
	for _, p := range snapshot.GetParticipantAddresses() {
		pubKeys[p.String()] = PublicKey()
	}

	for _, p := range missingParticipants {
		delete(pubKeys, p.String())
	}

	return types.Key{
		ID:               multisigtestutils.KeyID(),
		Snapshot:         snapshot,
		PubKeys:          pubKeys,
		SigningThreshold: threshold,
	}
}

// MultiSig returns a random multi sig
func MultiSig() types.MultiSig {
	payloadHash := rand.Bytes(exported.HashLength)
	participants := slices.Expand(func(int) sdk.ValAddress { return rand.ValAddr() }, int(rand.I64Between(5, 10)))
	sigs := make(map[string]types.Signature, len(participants))

	for _, p := range participants {
		sk := funcs.Must(btcec.NewPrivateKey(btcec.S256()))
		sigs[p.String()] = funcs.Must(sk.Sign(payloadHash)).Serialize()
	}

	return types.MultiSig{
		KeyID:       multisigtestutils.KeyID(),
		Sigs:        sigs,
		PayloadHash: payloadHash,
	}
}
