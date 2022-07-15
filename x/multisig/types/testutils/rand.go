package testutils

import (
	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	utilstestutils "github.com/axelarnetwork/axelar-core/utils/testutils"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigtestutils "github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
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

// MultiSig returns a random multi sig
func MultiSig() types.MultiSig {
	payloadHash := rand.Bytes(types.HashLength)
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
