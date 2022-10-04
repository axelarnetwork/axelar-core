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
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

const maxAmount int64 = 100000000000

var keeper nexusKeeper.Keeper
var bankK *axelarnetmock.BankKeeperMock

func addressValidator() types.Router {
	axelarnetK := &axelarnetmock.BaseKeeperMock{
		GetCosmosChainByNameFunc: func(ctx sdk.Context, chain exported.ChainName) (axelarnetTypes.CosmosChain, bool) {
			var prefix string
			switch chain.String() {
			case "Axelarnet":
				prefix = "axelar"
			case "terra":
				prefix = "terra"
			default:
				panic("unknown chain")
			}
			return axelarnetTypes.CosmosChain{Name: chain, AddrPrefix: prefix}, true
		},
	}

	bankK = &axelarnetmock.BankKeeperMock{
		BlockedAddrFunc: func(addr sdk.AccAddress) bool { return false },
	}

	router := types.NewRouter()
	router.AddAddressValidator(evmTypes.ModuleName, evmkeeper.NewAddressValidator()).
		AddAddressValidator(axelarnetTypes.ModuleName, axelarnetkeeper.NewAddressValidator(axelarnetK, bankK))

	return router
}

func init() {
	encCfg := app.MakeEncodingConfig()
	nexusSubspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("nexusKey"), sdk.NewKVStoreKey("tNexusKey"), "nexus")
	keeper = nexusKeeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("nexus"), nexusSubspace)
	keeper.SetRouter(addressValidator())
}

func TestLinkAddress(t *testing.T) {
	repeats := 20

	var ctx sdk.Context
	terra := nexus.Chain{Name: nexus.ChainName("terra"), Module: axelarnetTypes.ModuleName, SupportsForeignAssets: true}
	evmAddr := exported.CrossChainAddress{Chain: evm.Ethereum, Address: "0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"}
	axelarAddr := exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: "axelar1t66w8cazua870wu7t2hsffndmy2qy2v556ymndnczs83qpz2h45sq6lq9w"}

	setup := func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper.SetParams(ctx, types.DefaultParams())

		// set chain
		for _, chain := range []exported.Chain{evm.Ethereum, axelarnet.Axelarnet, terra} {
			keeper.SetChain(ctx, chain)
			keeper.ActivateChain(ctx, chain)
		}

		bankK.BlockedAddrFunc = func(addr sdk.AccAddress) bool { return false }
	}

	t.Run("should pass address validation", testutils.Func(func(t *testing.T) {
		setup()
		err := keeper.LinkAddresses(ctx,
			evmAddr,
			exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: "axelar1t66w8cazua870wu7t2hsffndmy2qy2v556ymndnczs83qpz2h45sq6lq9w"},
		)
		assert.NoError(t, err)

		err = keeper.LinkAddresses(ctx,
			evmAddr,
			exported.CrossChainAddress{Chain: terra, Address: "terra18zhnqjv70v0d2f8v0s5lape0gr5ua94eqkk8ex"},
		)
		assert.NoError(t, err)

		err = keeper.LinkAddresses(ctx,
			exported.CrossChainAddress{Chain: evm.Ethereum, Address: "68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"},
			axelarAddr,
		)
		assert.NoError(t, err)
	}))

	t.Run("should return error when linking invalid addresses", testutils.Func(func(t *testing.T) {
		setup()

		err := keeper.LinkAddresses(ctx,
			exported.CrossChainAddress{Chain: evm.Ethereum, Address: "0xZ8B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"},
			axelarAddr,
		)
		assert.ErrorContains(t, err, "not an hex address")

		err = keeper.LinkAddresses(ctx,
			evmAddr,
			exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: rand.StrBetween(10, 30)},
		)
		assert.ErrorContains(t, err, "decoding bech32 failed")
	}))

	t.Run("should return error for blocked addresses", testutils.Func(func(t *testing.T) {
		setup()
		blockedAddr := rand.AccAddr()
		bankK.BlockedAddrFunc = func(addr sdk.AccAddress) bool { return addr.Equals(blockedAddr) }

		err := keeper.LinkAddresses(ctx,
			evmAddr,
			axelarAddr,
		)
		assert.NoError(t, err)

		err = keeper.LinkAddresses(ctx,
			exported.CrossChainAddress{Chain: evm.Ethereum, Address: "0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"},
			exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: blockedAddr.String()},
		)
		assert.ErrorContains(t, err, "is not allowed to receive")

		err = keeper.LinkAddresses(ctx,
			exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: blockedAddr.String()},
			evmAddr,
		)
		assert.ErrorContains(t, err, "is not allowed to receive")

		err = keeper.LinkAddresses(ctx,
			evmAddr,
			exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: rand.AccAddr().String()},
		)
		assert.NoError(t, err)
	}))

	t.Run("should return error when link chain does not support foreign asset", testutils.Func(func(t *testing.T) {
		setup()
		fromChain := nexus.Chain{
			Name:                  nexus.ChainName(rand.Str(5)),
			SupportsForeignAssets: false,
			Module:                evmTypes.ModuleName,
		}
		keeper.SetChain(ctx, fromChain)
		keeper.ActivateChain(ctx, fromChain)
		sender, recipient := makeRandAddressesForChain(fromChain, evm.Ethereum)
		err := keeper.LinkAddresses(ctx, sender, recipient)
		assert.NoError(t, err)
		_, err = keeper.EnqueueForTransfer(ctx, sender, makeRandAmount(makeRandomDenom()))
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("successfully link", testutils.Func(func(t *testing.T) {
		setup()
		sender, recipient := makeRandAddressesForChain(axelarnet.Axelarnet, evm.Ethereum)
		err := keeper.LinkAddresses(ctx, sender, recipient)
		assert.NoError(t, err)
		_, err = keeper.EnqueueForTransfer(ctx, sender, makeRandAmount(axelarnet.NativeAsset))
		assert.NoError(t, err)
		recp, ok := keeper.GetRecipient(ctx, sender)
		assert.True(t, ok)
		assert.Equal(t, recipient, recp)

		sender.Address = rand.Str(20)
		_, err = keeper.EnqueueForTransfer(ctx, sender, makeRandAmount(axelarnet.NativeAsset))
		assert.Error(t, err)
		recp, ok = keeper.GetRecipient(ctx, sender)
		assert.False(t, ok)
		assert.NotEqual(t, recipient, recp)
	}).Repeat(repeats))
}

func TestSetChainGetChain_MixCaseChainName(t *testing.T) {
	chainName := strings.ToUpper(rand.StrBetween(5, 10)) + strings.ToLower(rand.StrBetween(5, 10))
	chain := makeRandomChain(chainName)

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper.SetChain(ctx, chain)

	actual, ok := keeper.GetChain(ctx, exported.ChainName(strings.ToUpper(chainName)))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)

	actual, ok = keeper.GetChain(ctx, exported.ChainName(strings.ToLower(chainName)))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)
}

func TestSetChainGetChain_UpperCaseChainName(t *testing.T) {
	chainName := strings.ToUpper(rand.StrBetween(5, 10))
	chain := makeRandomChain(chainName)

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper.SetChain(ctx, chain)

	actual, ok := keeper.GetChain(ctx, exported.ChainName(strings.ToUpper(chainName)))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)

	actual, ok = keeper.GetChain(ctx, exported.ChainName(strings.ToLower(chainName)))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)
}

func TestSetChainGetChain_LowerCaseChainName(t *testing.T) {
	chainName := strings.ToLower(rand.StrBetween(5, 10))
	chain := makeRandomChain(chainName)

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper.SetChain(ctx, chain)

	actual, ok := keeper.GetChain(ctx, exported.ChainName(strings.ToUpper(chainName)))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)

	actual, ok = keeper.GetChain(ctx, exported.ChainName(strings.ToLower(chainName)))

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
	prefix := ""
	switch strings.ToLower(chain) {
	case "axelarnet":
		prefix = "axelar"
	case "terra":
		prefix = "terra"
	default:
		prefix = ""
	}

	sdk.GetConfig().SetBech32PrefixForAccount(prefix, prefix)
	return rand.AccAddr().String()
}
