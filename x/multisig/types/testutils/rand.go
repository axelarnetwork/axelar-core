package testutils

import (
	"github.com/btcsuite/btcd/btcec"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	utilstestutils "github.com/axelarnetwork/axelar-core/utils/testutils"
	multisigtestutils "github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	snapshottestutils "github.com/axelarnetwork/axelar-core/x/snapshot/exported/testutils"
)

// PublicKey returns a random public key
func PublicKey() types.PublicKey {
	privKey, _ := btcec.NewPrivateKey(btcec.S256())

	return privKey.PubKey().SerializeCompressed()
}

// Key returns a random key
func Key() types.Key {
	threshold := utilstestutils.RandThreshold()
	snapshot := snapshottestutils.Snapshot(uint64(rand.I64Between(10, 20)), threshold)
	participants := snapshot.GetParticipantAddresses()
	pubKeys := make(map[string]types.PublicKey, len(participants))

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
