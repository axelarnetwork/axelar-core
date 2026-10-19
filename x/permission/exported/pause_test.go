package exported_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/x/permission/exported"
	permissiontypes "github.com/axelarnetwork/axelar-core/x/permission/types"
)

func TestEmergencyPauseAllowlist_TypeURLsAreRegisteredMessages(t *testing.T) {
	registry := app.MakeEncodingConfig().InterfaceRegistry

	for level, urls := range exported.EmergencyPauseAllowlist {
		assert.NoError(t, level.ValidatePauseLevel())
		assert.True(t, level.IsPaused())

		for _, url := range urls {
			msg, err := registry.Resolve(url)
			if assert.NoError(t, err, url) {
				_, ok := msg.(sdk.Msg)
				assert.True(t, ok, url)
			}
		}
	}
}

func TestEmergencyPauseAllowlist_Levels(t *testing.T) {
	all := exported.EmergencyPauseAllowlist[exported.PAUSE_LEVEL_ALL]
	exceptValidators := exported.EmergencyPauseAllowlist[exported.PAUSE_LEVEL_EXCEPT_VALIDATORS]

	// the strictest level only lets the chain get out of the pause or escalate it
	assert.ElementsMatch(t, []string{
		sdk.MsgTypeURL(&permissiontypes.PauseRequest{}),
		sdk.MsgTypeURL(&permissiontypes.UnpauseRequest{}),
		sdk.MsgTypeURL(&permissiontypes.HaltChainRequest{}),
	}, all)

	assert.Subset(t, exceptValidators, all)
	assert.Contains(t, exceptValidators, "/axelar.vote.v1beta1.VoteRequest")
	assert.Contains(t, exceptValidators, "/axelar.multisig.v1beta1.SubmitPubKeyRequest")
	assert.Contains(t, exceptValidators, "/axelar.multisig.v1beta1.SubmitSignatureRequest")
	assert.Contains(t, exceptValidators, "/axelar.reward.v1beta1.RefundMsgRequest")
	assert.Contains(t, exceptValidators, "/axelar.auxiliary.v1beta1.BatchRequest")
	assert.Contains(t, exceptValidators, sdk.MsgTypeURL(&permissiontypes.UpdateGovernanceKeyRequest{}))
	assert.Contains(t, exceptValidators, sdk.MsgTypeURL(&permissiontypes.RegisterControllerRequest{}))
	assert.Contains(t, exceptValidators, sdk.MsgTypeURL(&permissiontypes.DeregisterControllerRequest{}))
}

func TestIsAllowedWhilePaused(t *testing.T) {
	bankSend := sdk.MsgTypeURL(&banktypes.MsgSend{})
	unpause := sdk.MsgTypeURL(&permissiontypes.UnpauseRequest{})
	unknownLevel := exported.PauseLevel(1000)

	assert.True(t, exported.IsAllowedWhilePaused(exported.PAUSE_LEVEL_UNSPECIFIED, bankSend))
	assert.False(t, exported.IsAllowedWhilePaused(exported.PAUSE_LEVEL_ALL, bankSend))
	assert.False(t, exported.IsAllowedWhilePaused(exported.PAUSE_LEVEL_EXCEPT_VALIDATORS, bankSend))
	assert.False(t, exported.IsAllowedWhilePaused(unknownLevel, bankSend))

	// the chain can always get out of a pause
	assert.True(t, exported.IsAllowedWhilePaused(exported.PAUSE_LEVEL_ALL, unpause))
	assert.True(t, exported.IsAllowedWhilePaused(unknownLevel, unpause))
	assert.Error(t, unknownLevel.ValidatePauseLevel())
	assert.Error(t, exported.PAUSE_LEVEL_UNSPECIFIED.ValidatePauseLevel())
}

func TestEmergencyHaltPlanName(t *testing.T) {
	assert.Equal(t, "emergency-halt-123", exported.EmergencyHaltPlanName(123))
}
