package keeper_test

import (
	"strings"
	"testing"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestAddressValidator(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	axelarnetK := &mock.BaseKeeperMock{
		GetCosmosChainByNameFunc: func(ctx sdk.Context, chain nexus.ChainName) (types.CosmosChain, bool) {
			var prefix string
			switch chain.String() {
			case "Axelarnet":
				prefix = "axelar"
			case "terra":
				prefix = "terra"
			default:
				panic("unknown chain")
			}
			return types.CosmosChain{Name: chain, AddrPrefix: prefix}, true
		},
	}

	bankK := &mock.BankKeeperMock{
		BlockedAddrFunc: func(addr sdk.AccAddress) bool { return false },
	}

	validator := keeper.NewAddressValidator(axelarnetK, bankK)
	assert.NotNil(t, validator)

	addr := nexus.CrossChainAddress{Chain: exported.Axelarnet, Address: "axelar1t66w8cazua870wu7t2hsffndmy2qy2v556ymndnczs83qpz2h45sq6lq9w"}
	assert.NoError(t, validator(ctx, addr))

	addr = nexus.CrossChainAddress{Chain: exported.Axelarnet, Address: "68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"}
	assert.NoError(t, validator(ctx, addr))

	addr = nexus.CrossChainAddress{Chain: exported.Axelarnet, Address: "0X68B93045FE7D8794A7CAF327E7F855CD6CD03BB8"}
	assert.NoError(t, validator(ctx, addr))

	addr = nexus.CrossChainAddress{Chain: exported.Axelarnet, Address: "0x" + strings.Repeat("0", 40)}
	assert.NoError(t, validator(ctx, addr))

	addr = nexus.CrossChainAddress{Chain: exported.Axelarnet, Address: strings.Repeat("0", 40)}
	assert.NoError(t, validator(ctx, addr))

	addr = nexus.CrossChainAddress{Chain: exported.Axelarnet, Address: ""}
	assert.Error(t, validator(ctx, addr))

	addr = nexus.CrossChainAddress{Chain: exported.Axelarnet, Address: "0x" + strings.Repeat("0", 41)}
	assert.Error(t, validator(ctx, addr))

	addr = nexus.CrossChainAddress{Chain: evm.Ethereum, Address: "0x6ZB93045fe7D8794a7cAF327e7f855CD6Cd03BB8"}
	assert.Error(t, validator(ctx, addr))
}
