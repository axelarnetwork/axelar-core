package keeper_test

import (
	"strings"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	store "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	evmUtil "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	axelarnetkeeper "github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	axelarnetTypes "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	axelarnetmock "github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmkeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

const maxAmount int64 = 100000000000

var k keeper.Keeper

func addressValidators() *types.AddressValidators {
	axelarnetK := &axelarnetmock.BaseKeeperMock{
		GetCosmosChainByNameFunc: func(ctx sdk.Context, chain exported.ChainName) (axelarnetTypes.CosmosChain, bool) {
			var prefix string
			switch chain {
			case axelarnet.Axelarnet.Name:
				prefix = "axelar"
			case "terra", "terra-2":
				prefix = "terra"
			default:
				prefix = strings.ToLower(chain.String())
			}
			return axelarnetTypes.CosmosChain{Name: chain, AddrPrefix: prefix}, true
		},
	}

	validators := types.NewAddressValidators()
	validators.AddAddressValidator(evmTypes.ModuleName, evmkeeper.NewAddressValidator()).
		AddAddressValidator(axelarnetTypes.ModuleName, axelarnetkeeper.NewAddressValidator(axelarnetK))

	validators.Seal()

	return validators
}

func init() {
	encCfg := app.MakeEncodingConfig()
	subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, store.NewKVStoreKey("nexusKey"), store.NewKVStoreKey("tNexusKey"), "nexus")
	k = keeper.NewKeeper(encCfg.Codec, store.NewKVStoreKey("nexus"), subspace)
	k.SetAddressValidators(addressValidators())
}

func TestValidateAddress(t *testing.T) {
	var ctx sdk.Context
	cfg := app.MakeEncodingConfig()
	k, ctx = setup(cfg, t)

	terra := exported.Chain{Name: exported.ChainName("terra"), Module: axelarnetTypes.ModuleName, SupportsForeignAssets: true}
	evmAddr := exported.CrossChainAddress{Chain: evm.Ethereum, Address: "0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"}
	axelarAddr := exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: "axelar1t66w8cazua870wu7t2hsffndmy2qy2v556ymndnczs83qpz2h45sq6lq9w"}

	t.Run("should pass address validation", testutils.Func(func(t *testing.T) {
		assert.NoError(t, k.ValidateAddress(ctx, evmAddr))
		assert.NoError(t, k.ValidateAddress(ctx, axelarAddr))
		assert.NoError(t, k.ValidateAddress(ctx, exported.CrossChainAddress{Chain: terra, Address: "terra18zhnqjv70v0d2f8v0s5lape0gr5ua94eqkk8ex"}))
		assert.NoError(t, k.ValidateAddress(ctx, exported.CrossChainAddress{Chain: evm.Ethereum, Address: "68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"}))
	}))

	t.Run("should return error when validating invalid addresses", testutils.Func(func(t *testing.T) {
		err := k.ValidateAddress(ctx, exported.CrossChainAddress{Chain: evm.Ethereum, Address: "0xZ8B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"})
		assert.ErrorContains(t, err, "not an hex address")

		err = k.ValidateAddress(ctx, exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: rand.StrBetween(10, 30)})
		assert.ErrorContains(t, err, "decoding bech32 failed")
	}))
}

func TestSetChainGetChain_MixCaseChainName(t *testing.T) {
	chainName := strings.ToUpper(rand.StrBetween(5, 10)) + strings.ToLower(rand.StrBetween(5, 10))
	chain := makeRandomChain(chainName)

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))
	k.SetChain(ctx, chain)

	actual, ok := k.GetChain(ctx, exported.ChainName(strings.ToUpper(chainName)))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)

	actual, ok = k.GetChain(ctx, exported.ChainName(strings.ToLower(chainName)))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)
}

func TestSetChainGetChain_UpperCaseChainName(t *testing.T) {
	chainName := strings.ToUpper(rand.StrBetween(5, 10))
	chain := makeRandomChain(chainName)

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))
	k.SetChain(ctx, chain)

	actual, ok := k.GetChain(ctx, exported.ChainName(strings.ToUpper(chainName)))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)

	actual, ok = k.GetChain(ctx, exported.ChainName(strings.ToLower(chainName)))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)
}

func TestSetChainGetChain_LowerCaseChainName(t *testing.T) {
	chainName := strings.ToLower(rand.StrBetween(5, 10))
	chain := makeRandomChain(chainName)

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))
	k.SetChain(ctx, chain)

	actual, ok := k.GetChain(ctx, exported.ChainName(strings.ToUpper(chainName)))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)

	actual, ok = k.GetChain(ctx, exported.ChainName(strings.ToLower(chainName)))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)
}

func makeRandomChain(chainName string) exported.Chain {
	return exported.Chain{
		Name:                  exported.ChainName(chainName),
		Module:                rand.Str(10),
		SupportsForeignAssets: true,
	}
}

func makeRandomDenom() string {
	d := rand.Strings(3, 4).WithAlphabet([]rune("abcdefghijklmnopqrstuvwxyz")).Take(1)
	return d[0]
}

func makeRandAmount(denom string) sdk.Coin {
	return sdk.NewCoin(denom, math.NewInt(rand.I64Between(1, maxAmount)))
}

func makeRandAddressesForChain(origin, destination exported.Chain) (exported.CrossChainAddress, exported.CrossChainAddress) {
	var addr string

	switch origin.Module {
	case evmTypes.ModuleName:
		addr = genEvmAddr()
	case axelarnetTypes.ModuleName:
		addr = genCosmosAddr(origin.Name.String())
	default:
		panic("unexpected module for origin")
	}

	sender := exported.CrossChainAddress{
		Address: addr,
		Chain:   origin,
	}

	switch destination.Module {
	case evmTypes.ModuleName:
		addr = genEvmAddr()
	case axelarnetTypes.ModuleName:
		addr = genCosmosAddr(destination.Name.String())
	default:
		panic("unexpected module for destination")
	}

	recipient := exported.CrossChainAddress{
		Address: addr,
		Chain:   destination,
	}

	// Reset bech32 prefix
	sdk.GetConfig().SetBech32PrefixForAccount("axelar", "axelar")

	return sender, recipient
}

func genEvmAddr() string {
	return evmUtil.BytesToAddress(rand.Bytes(evmUtil.AddressLength)).Hex()
}

func genCosmosAddr(chain string) string {
	defer sdk.GetConfig().SetBech32PrefixForAccount("axelar", "axelar")

	prefix := ""
	switch strings.ToLower(chain) {
	case "axelarnet":
		prefix = "axelar"
	case "terra", "terra-2":
		prefix = "terra"
	default:
		prefix = chain
	}

	sdk.GetConfig().SetBech32PrefixForAccount(prefix, prefix)
	return rand.AccAddr().String()
}
