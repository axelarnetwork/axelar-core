package keeper_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	evmUtil "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

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
	subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("nexusKey"), sdk.NewKVStoreKey("tNexusKey"), "nexus")
	k = keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("nexus"), subspace)
	k.SetAddressValidators(addressValidators())
}

func TestLinkAddress(t *testing.T) {
	repeats := 20

	var ctx sdk.Context
	cfg := app.MakeEncodingConfig()
	k, ctx = setup(cfg)

	terra := exported.Chain{Name: exported.ChainName("terra"), Module: axelarnetTypes.ModuleName, SupportsForeignAssets: true}
	evmAddr := exported.CrossChainAddress{Chain: evm.Ethereum, Address: "0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"}
	axelarAddr := exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: "axelar1t66w8cazua870wu7t2hsffndmy2qy2v556ymndnczs83qpz2h45sq6lq9w"}

	t.Run("should pass address validation", testutils.Func(func(t *testing.T) {
		err := k.LinkAddresses(ctx,
			evmAddr,
			exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: "axelar1t66w8cazua870wu7t2hsffndmy2qy2v556ymndnczs83qpz2h45sq6lq9w"},
		)
		assert.NoError(t, err)

		err = k.LinkAddresses(ctx,
			evmAddr,
			exported.CrossChainAddress{Chain: terra, Address: "terra18zhnqjv70v0d2f8v0s5lape0gr5ua94eqkk8ex"},
		)
		assert.NoError(t, err)

		err = k.LinkAddresses(ctx,
			exported.CrossChainAddress{Chain: evm.Ethereum, Address: "68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"},
			axelarAddr,
		)
		assert.NoError(t, err)
	}))

	t.Run("should return error when linking invalid addresses", testutils.Func(func(t *testing.T) {
		err := k.LinkAddresses(ctx,
			exported.CrossChainAddress{Chain: evm.Ethereum, Address: "0xZ8B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"},
			axelarAddr,
		)
		assert.ErrorContains(t, err, "not an hex address")

		err = k.LinkAddresses(ctx,
			evmAddr,
			exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: rand.StrBetween(10, 30)},
		)
		assert.ErrorContains(t, err, "decoding bech32 failed")
	}))

	t.Run("should return error when link chain does not support foreign asset", testutils.Func(func(t *testing.T) {
		fromChain := exported.Chain{
			Name:                  exported.ChainName(rand.Str(5)),
			SupportsForeignAssets: false,
			Module:                evmTypes.ModuleName,
		}
		k.SetChain(ctx, fromChain)
		k.ActivateChain(ctx, fromChain)
		sender, recipient := makeRandAddressesForChain(fromChain, evm.Ethereum)
		err := k.LinkAddresses(ctx, sender, recipient)
		assert.NoError(t, err)
		_, err = k.EnqueueForTransfer(ctx, sender, makeRandAmount(makeRandomDenom()))
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("successfully link", testutils.Func(func(t *testing.T) {
		sender, recipient := makeRandAddressesForChain(axelarnet.Axelarnet, evm.Ethereum)
		err := k.LinkAddresses(ctx, sender, recipient)
		assert.NoError(t, err)
		_, err = k.EnqueueForTransfer(ctx, sender, makeRandAmount(axelarnet.NativeAsset))
		assert.NoError(t, err)
		recp, ok := k.GetRecipient(ctx, sender)
		assert.True(t, ok)
		assert.Equal(t, recipient, recp)

		sender.Address = rand.Str(20)
		_, err = k.EnqueueForTransfer(ctx, sender, makeRandAmount(axelarnet.NativeAsset))
		assert.Error(t, err)
		recp, ok = k.GetRecipient(ctx, sender)
		assert.False(t, ok)
		assert.NotEqual(t, recipient, recp)
	}).Repeat(repeats))
}

func TestSetChainGetChain_MixCaseChainName(t *testing.T) {
	chainName := strings.ToUpper(rand.StrBetween(5, 10)) + strings.ToLower(rand.StrBetween(5, 10))
	chain := makeRandomChain(chainName)

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
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

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
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

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
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
	return sdk.NewCoin(denom, sdk.NewInt(rand.I64Between(1, maxAmount)))
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
