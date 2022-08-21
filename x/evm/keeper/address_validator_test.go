package keeper_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

func TestAddressValidator(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	validator := keeper.NewAddressValidator()
	assert.NotNil(t, validator)

	addr := nexus.CrossChainAddress{Chain: exported.Ethereum, Address: "0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"}
	assert.NoError(t, validator(ctx, addr))

	addr = nexus.CrossChainAddress{Chain: exported.Ethereum, Address: "68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"}
	assert.NoError(t, validator(ctx, addr))

	addr = nexus.CrossChainAddress{Chain: exported.Ethereum, Address: "0X68B93045FE7D8794A7CAF327E7F855CD6CD03BB8"}
	assert.NoError(t, validator(ctx, addr))

	addr = nexus.CrossChainAddress{Chain: exported.Ethereum, Address: "0x" + strings.Repeat("0", 40)}
	assert.NoError(t, validator(ctx, addr))

	addr = nexus.CrossChainAddress{Chain: exported.Ethereum, Address: strings.Repeat("0", 40)}
	assert.NoError(t, validator(ctx, addr))

	addr = nexus.CrossChainAddress{Chain: nexus.Chain{}, Address: "0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"}
	assert.NoError(t, validator(ctx, addr))

	addr = nexus.CrossChainAddress{Chain: exported.Ethereum, Address: ""}
	assert.Error(t, validator(ctx, addr))

	addr = nexus.CrossChainAddress{Chain: exported.Ethereum, Address: ""}
	assert.Error(t, validator(ctx, addr))

	addr = nexus.CrossChainAddress{Chain: exported.Ethereum, Address: "0x" + strings.Repeat("0", 41)}
	assert.Error(t, validator(ctx, addr))

	addr = nexus.CrossChainAddress{Chain: exported.Ethereum, Address: "0x6ZB93045fe7D8794a7cAF327e7f855CD6Cd03BB8"}
	assert.Error(t, validator(ctx, addr))
}
