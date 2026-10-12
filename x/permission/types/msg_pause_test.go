package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/permission/exported"
	"github.com/axelarnetwork/axelar-core/x/permission/types"
)

func TestEmergencyMsgs_ValidateBasic(t *testing.T) {
	sender := rand.AccAddr()

	assert.NoError(t, types.NewPauseRequest(sender, exported.PAUSE_LEVEL_ALL, "incident").ValidateBasic())
	assert.NoError(t, types.NewPauseRequest(sender, exported.PAUSE_LEVEL_EXCEPT_VALIDATORS, "incident").ValidateBasic())
	assert.Error(t, types.NewPauseRequest(sender, exported.PAUSE_LEVEL_UNSPECIFIED, "incident").ValidateBasic())
	assert.Error(t, types.NewPauseRequest(sender, exported.PauseLevel(1000), "incident").ValidateBasic())
	assert.Error(t, types.NewPauseRequest(sender, exported.PAUSE_LEVEL_ALL, "").ValidateBasic())
	assert.Error(t, types.NewPauseRequest(sdk.AccAddress{}, exported.PAUSE_LEVEL_ALL, "incident").ValidateBasic())

	assert.NoError(t, types.NewUnpauseRequest(sender, "resolved").ValidateBasic())
	assert.Error(t, types.NewUnpauseRequest(sender, "").ValidateBasic())
	assert.Error(t, types.NewUnpauseRequest(sdk.AccAddress{}, "resolved").ValidateBasic())

	assert.NoError(t, types.NewHaltChainRequest(sender, "incident", false).ValidateBasic())
	assert.NoError(t, types.NewHaltChainRequest(sender, "incident", true).ValidateBasic())
	assert.Error(t, types.NewHaltChainRequest(sender, "", false).ValidateBasic())
	assert.Error(t, types.NewHaltChainRequest(sdk.AccAddress{}, "incident", false).ValidateBasic())
}

func TestEmergencyMsgs_RequireAccessControlRole(t *testing.T) {
	for _, msg := range []sdk.Msg{&types.PauseRequest{}, &types.UnpauseRequest{}, &types.HaltChainRequest{}} {
		assert.Equal(t, exported.ROLE_ACCESS_CONTROL, exported.GetPermissionRole(msg.(interface {
			sdk.Msg
			Descriptor() ([]byte, []int)
		})), sdk.MsgTypeURL(msg))
	}
}
