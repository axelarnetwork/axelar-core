package exported

import (
	"fmt"
	"slices"
)

const EmergencyHaltPlanPrefix = "emergency-halt-"

// EmergencyHaltPlanName returns the name of the upgrade plan that halts the chain at the given height.
func EmergencyHaltPlanName(height int64) string {
	return fmt.Sprintf("%s%d", EmergencyHaltPlanPrefix, height)
}

var (
	// messages needed to get out of a pause, change its level, or escalate it to a halt
	emergencyControlMsgs = []string{
		"/axelar.permission.v1beta1.PauseRequest",
		"/axelar.permission.v1beta1.UnpauseRequest",
		"/axelar.permission.v1beta1.HaltChainRequest",
	}

	// messages needed to rotate or recover the governance key and controller accounts
	permissionKeyRotationMsgs = []string{
		"/axelar.permission.v1beta1.UpdateGovernanceKeyRequest",
		"/axelar.permission.v1beta1.RegisterControllerRequest",
		"/axelar.permission.v1beta1.DeregisterControllerRequest",
	}

	// messages validators must keep submitting so that open polls, keygen and signing sessions
	validatorLivenessMsgs = []string{
		"/axelar.vote.v1beta1.VoteRequest",
		"/axelar.multisig.v1beta1.SubmitPubKeyRequest",
		"/axelar.multisig.v1beta1.SubmitSignatureRequest",
		"/axelar.reward.v1beta1.RefundMsgRequest",
		"/axelar.auxiliary.v1beta1.BatchRequest",
	}
)

// EmergencyPauseAllowlist lists, for every pause level.
var EmergencyPauseAllowlist = map[PauseLevel][]string{
	PAUSE_LEVEL_ALL: emergencyControlMsgs,
	PAUSE_LEVEL_EXCEPT_VALIDATORS: slices.Concat(
		emergencyControlMsgs,
		permissionKeyRotationMsgs,
		validatorLivenessMsgs,
	),
}

// IsPaused returns true if the level pauses the chain
func (x PauseLevel) IsPaused() bool {
	return x != PAUSE_LEVEL_UNSPECIFIED
}

// ValidatePauseLevel returns an error if the level is not a known level that pauses the chain
func (x PauseLevel) ValidatePauseLevel() error {
	if _, ok := EmergencyPauseAllowlist[x]; !ok {
		return fmt.Errorf("invalid pause level %s", x)
	}

	return nil
}

// IsAllowedWhilePaused returns true if the message with the given type URL is accepted at the given pause level
func IsAllowedWhilePaused(level PauseLevel, msgTypeURL string) bool {
	if !level.IsPaused() {
		return true
	}

	// never lock the chain out of its own pause, even at a level without an allowlist entry
	if slices.Contains(emergencyControlMsgs, msgTypeURL) {
		return true
	}

	return slices.Contains(EmergencyPauseAllowlist[level], msgTypeURL)
}
