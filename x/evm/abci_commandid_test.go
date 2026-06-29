package evm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
)

// CommandID regression tests for the message-delivery route (nexus -> EVM).
//
// A CommandID is only ever produced when the *destination* is an EVM chain:
// deliverPendingMessages (x/evm/abci.go) iterates only IsEVMChain destinations
// and calls deliverMessage / deliverMessageWithToken, which derive the id via
//
//	NewCommandID([]byte(msg.ID), <destination chain id>)   // x/evm/types/command.go
//
// Cosmos and amplifier *destinations* never reach this code (they route over
// IBC / the multisig-prover respectively) and therefore have no CommandID.
//
// So the only thing the *source* chain type changes is the FORMAT of msg.ID
// that gets hashed:
//
//	EVM source       -> "{txID.Hex()}-{index}"  (types.NewEventID)
//	Cosmos source    -> "0x{hash}-{nonce}"       (nexus Keeper.GenerateMessageID)
//	amplifier source -> source-supplied id string
//
// These tests drive the real route end-to-end (EndBlocker -> deliverPending...)
// and pin cmd.ID to a hard-coded golden value per source format. The goldens are
// frozen literals (NOT recomputed with NewCommandID in-test) on purpose: a buggy
// refactor of either the route's msg.ID wiring OR NewCommandID itself would change
// a recomputed value in lockstep and stay green. Frozen literals catch both — this
// is exactly the EVM->EVM CommandID change in #2321 that no test flagged.
//
// The destination chain id is math.NewInt(1) (newDeliveryTestSetup, abci_test.go);
// the appended chain-id byte is therefore 0x01. To regenerate a golden after an
// *intentional* change, compute types.NewCommandID([]byte(id), math.NewInt(1)).Hex().

// commandIDRoute is one source-format -> golden-CommandID expectation.
type commandIDRoute struct {
	name  string // source chain type the msg.ID format represents
	msgID string // msg.ID in the exact format that source produces
	cmdID string // golden CommandID (hex, no 0x), for destination chain id == 1
}

var commandIDRoutes = []commandIDRoute{
	{
		// EVM source: types.NewEventID(txID, index) == "{txID.Hex()}-{index}"
		// Note: Express executor depends on this. If this changes, also adjust it there!
		name:  "evm",
		msgID: "0x00000000000000000000000000000000000000000000000000000000000000ab-7",
		cmdID: "1f4190d1a72df4ba178c631976f917fb424a17f454fdb5dc532a5d73bb0ca532",
	},
	{
		// Cosmos source: nexus Keeper.GenerateMessageID == "0x{hash}-{nonce}"
		name:  "cosmos",
		msgID: "0x00000000000000000000000000000000000000000000000000000000000000cd-42",
		cmdID: "531c12d89fc2ae6078679ce1fbfdf7204ae0142d17db347b73c3b4b5c44c34e3",
	},
	{
		// amplifier source: source-supplied id string (e.g. stellar/sui style)
		name:  "amplifier",
		msgID: "stellar-0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef-3",
		cmdID: "fbac615bc40eadbfd5b05f5c9d53da71fc9bc258488d4d8d88e6c04e88fc6bb5",
	},
}

// Source {EVM, Cosmos, amplifier} -> destination EVM, no token.
func TestDeliverMessage_CommandIDPerSourceRoute(t *testing.T) {
	for _, tc := range commandIDRoutes {
		t.Run(tc.name, func(t *testing.T) {
			s := newDeliveryTestSetup(t)

			msg := s.createGeneralMessage()
			msg.ID = tc.msgID
			s.queueMessages(msg)

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			calls := s.destChainKeeper.EnqueueCommandCalls()
			assert.Len(t, calls, 1, "exactly one command should be enqueued")
			cmd := calls[0].Cmd

			assert.Equal(t, types.COMMAND_TYPE_APPROVE_CONTRACT_CALL, cmd.Type)
			assert.Equal(t, tc.cmdID, cmd.ID.Hex(), "CommandID for %s-source route changed", tc.name)

			assert.Len(t, s.nexus.SetMessageExecutedCalls(), 1)
			assert.Equal(t, msg.ID, s.nexus.SetMessageExecutedCalls()[0].ID)
			assert.Len(t, s.nexus.SetMessageFailedCalls(), 0)
		})
	}
}

// Source {EVM, Cosmos, amplifier} -> destination EVM, with token.
// The with-mint route derives the CommandID the same way (NewCommandID over
// msg.ID), so for a given msg.ID it MUST equal the no-token golden above —
// presence of a token must not perturb the id. We assert that explicitly.
func TestDeliverMessageWithToken_CommandIDPerSourceRoute(t *testing.T) {
	for _, tc := range commandIDRoutes {
		t.Run(tc.name, func(t *testing.T) {
			s := newDeliveryTestSetup(t)

			msg := s.createGeneralMessageWithToken()
			msg.ID = tc.msgID
			s.queueMessages(msg)

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			calls := s.destChainKeeper.EnqueueCommandCalls()
			assert.Len(t, calls, 1, "exactly one command should be enqueued")
			cmd := calls[0].Cmd

			assert.Equal(t, types.COMMAND_TYPE_APPROVE_CONTRACT_CALL_WITH_MINT, cmd.Type)
			assert.Equal(t, tc.cmdID, cmd.ID.Hex(),
				"with-token CommandID for %s-source route changed (must match the no-token golden)", tc.name)

			assert.Len(t, s.nexus.SetMessageExecutedCalls(), 1)
			assert.Equal(t, msg.ID, s.nexus.SetMessageExecutedCalls()[0].ID)
			assert.Len(t, s.nexus.SetMessageFailedCalls(), 0)
		})
	}
}
