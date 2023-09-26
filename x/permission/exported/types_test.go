package exported_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/x/permission/exported"
)

func TestMsgRoles(t *testing.T) {
	registry := app.MakeEncodingConfig().InterfaceRegistry
	implementations := registry.ListImplementations(sdk.MsgInterfaceProtoName)
	slices.Sort(implementations)

	var missingRoles []string
	for _, implementation := range implementations {
		if strings.HasPrefix(implementation, "/ibc.") ||
			strings.HasPrefix(implementation, "/cosmos.") ||
			strings.HasPrefix(implementation, "/cosmwasm.") {
			continue
		}

		msg, err := registry.Resolve(implementation)
		assert.NoError(t, err)

		role := exported.GetPermissionRole(msg.(descriptor.Message))
		if role == exported.ROLE_UNSPECIFIED {
			missingRoles = append(missingRoles, implementation)
			continue
		}
	}

	if len(missingRoles) > 0 {
		assert.Fail(t, "Found msgs without defined role", strings.Join(missingRoles, "\n"))
	}
}
