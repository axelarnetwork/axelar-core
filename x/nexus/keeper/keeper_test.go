package keeper_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/btcsuite/btcutil"
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
	axelarnetTypes "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

const maxAmount int64 = 100000000000

var keeper nexusKeeper.Keeper
var feeRate = sdk.NewDecWithPrec(25, 5)

func addressValidator() types.Router {
	router := types.NewRouter()
	router.AddAddressValidator("evm", func(_ sdk.Context, addr nexus.CrossChainAddress) error {
		if !evmUtil.IsHexAddress(addr.Address) {
			return fmt.Errorf("not an hex address")
		}

		return nil
	}).AddAddressValidator("bitcoin", func(ctx sdk.Context, addr nexus.CrossChainAddress) error {
		if _, err := btcutil.DecodeAddress(addr.Address, btcTypes.Testnet3.Params()); err != nil {
			return err
		}

		return nil
	}).AddAddressValidator("axelarnet", func(ctx sdk.Context, addr nexus.CrossChainAddress) error {
		bz, err := sdk.GetFromBech32(addr.Address, getPrefixByAddress(addr.Address))
		if err != nil {
			return err
		}
		err = sdk.VerifyAddressFormat(bz)
		if err != nil {
			return err
		}

		return nil
	})

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

	setup := func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper.SetParams(ctx, types.DefaultParams())

		// set chain
		for _, chain := range []exported.Chain{btc.Bitcoin, evm.Ethereum, axelarnet.Axelarnet} {
			keeper.SetChain(ctx, chain)
			keeper.ActivateChain(ctx, chain)
		}

		// set native asset -> chain mapping
		_ = keeper.RegisterNativeAsset(ctx, btc.Bitcoin, btc.Satoshi)
		_ = keeper.RegisterNativeAsset(ctx, axelarnet.Axelarnet, axelarnet.Uaxl)
	}

	t.Run("should pass address validation", testutils.Func(func(t *testing.T) {
		setup()
		err := keeper.LinkAddresses(ctx,
			exported.CrossChainAddress{Chain: evm.Ethereum, Address: "0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"},
			exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: "axelar1t66w8cazua870wu7t2hsffndmy2qy2v556ymndnczs83qpz2h45sq6lq9w"},
		)

		assert.NoError(t, err)

		err = keeper.LinkAddresses(ctx,
			exported.CrossChainAddress{Chain: evm.Ethereum, Address: "0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"},
			exported.CrossChainAddress{Chain: btc.Bitcoin, Address: "bcrt1qjs8g7q8u0668l95zzxwqf2pnjnr005v2nasy7d32jrkd5cnmwmzsvx0c06"},
		)

		assert.NoError(t, err)
	}).Repeat(1))

	t.Run("should return error when link invalid addresses", testutils.Func(func(t *testing.T) {
		setup()
		err := keeper.LinkAddresses(ctx,
			exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: "0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"},
			exported.CrossChainAddress{Chain: evm.Ethereum, Address: "axelar1t66w8cazua870wu7t2hsffndmy2qy2v556ymndnczs83qpz2h45sq6lq9w"},
		)

		assert.Error(t, err)

		err = keeper.LinkAddresses(ctx,
			exported.CrossChainAddress{Chain: btc.Bitcoin, Address: "0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"},
			exported.CrossChainAddress{Chain: evm.Ethereum, Address: "bcrt1qjs8g7q8u0668l95zzxwqf2pnjnr005v2nasy7d32jrkd5cnmwmzsvx0c06"},
		)

		assert.Error(t, err)

		err = keeper.LinkAddresses(ctx,
			exported.CrossChainAddress{Chain: evm.Ethereum, Address: "0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"},
			exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: rand.StrBetween(10, 30)},
		)

		assert.Error(t, err)

		err = keeper.LinkAddresses(ctx,
			exported.CrossChainAddress{Chain: evm.Ethereum, Address: "0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"},
			exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: "terra1t66w8cazua870wu7t2hsffndmy2qy2v556ymndnczs83qpz2h45sq6lq9w"},
		)

		assert.Error(t, err)

		err = keeper.LinkAddresses(ctx,
			exported.CrossChainAddress{Chain: evm.Ethereum, Address: "0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"},
			exported.CrossChainAddress{Chain: btc.Bitcoin, Address: rand.StrBetween(10, 30)},
		)

		assert.Error(t, err)
	}).Repeat(1))

	t.Run("should return error when link chain which does not support foreign asset", testutils.Func(func(t *testing.T) {
		setup()
		sender, recipient := makeRandAddressesForChain(btc.Bitcoin, evm.Ethereum)
		err := keeper.LinkAddresses(ctx, sender, recipient)
		assert.NoError(t, err)
		_, err = keeper.EnqueueForTransfer(ctx, sender, makeRandAmount(makeRandomDenom()), feeRate)
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("successfully link", testutils.Func(func(t *testing.T) {
		setup()
		sender, recipient := makeRandAddressesForChain(btc.Bitcoin, evm.Ethereum)
		err := keeper.LinkAddresses(ctx, sender, recipient)
		assert.NoError(t, err)
		_, err = keeper.EnqueueForTransfer(ctx, sender, makeRandAmount(btcTypes.Satoshi), feeRate)
		assert.NoError(t, err)
		recp, ok := keeper.GetRecipient(ctx, sender)
		assert.True(t, ok)
		assert.Equal(t, recipient, recp)

		sender.Address = rand.Str(20)
		_, err = keeper.EnqueueForTransfer(ctx, sender, makeRandAmount(btcTypes.Satoshi), feeRate)
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

	actual, ok := keeper.GetChain(ctx, strings.ToUpper(chainName))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)

	actual, ok = keeper.GetChain(ctx, strings.ToLower(chainName))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)
}

func TestSetChainGetChain_UpperCaseChainName(t *testing.T) {
	chainName := strings.ToUpper(rand.StrBetween(5, 10))
	chain := makeRandomChain(chainName)

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper.SetChain(ctx, chain)

	actual, ok := keeper.GetChain(ctx, strings.ToUpper(chainName))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)

	actual, ok = keeper.GetChain(ctx, strings.ToLower(chainName))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)
}

func TestSetChainGetChain_LowerCaseChainName(t *testing.T) {
	chainName := strings.ToLower(rand.StrBetween(5, 10))
	chain := makeRandomChain(chainName)

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper.SetChain(ctx, chain)

	actual, ok := keeper.GetChain(ctx, strings.ToUpper(chainName))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)

	actual, ok = keeper.GetChain(ctx, strings.ToLower(chainName))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)
}

func makeRandomChain(chainName string) exported.Chain {
	return exported.Chain{
		Name:                  chainName,
		Module:                rand.Str(10),
		SupportsForeignAssets: true,
	}
}

func makeRandomDenom() string {
	d := rand.Strings(3, 3).WithAlphabet([]rune("abcdefghijklmnopqrstuvwxyz")).Take(1)
	return d[0]
}

func makeRandAmount(denom string) sdk.Coin {

	return sdk.NewCoin(denom, sdk.NewInt(rand.I64Between(1, maxAmount)))
}

func makeRandAddressesForChain(origin, destination exported.Chain) (exported.CrossChainAddress, exported.CrossChainAddress) {
	var addr string

	switch origin.Module {
	case btcTypes.ModuleName:
		addr = genBtcAddr()
	case evmTypes.ModuleName:
		addr = genEvmAddr()
	case axelarnetTypes.ModuleName:
		addr = genCosmosAddr(origin.Name)
	default:
		panic("unexpected module for origin")
	}

	sender := exported.CrossChainAddress{
		Address: addr,
		Chain:   origin,
	}

	switch destination.Module {
	case btcTypes.ModuleName:
		addr = genBtcAddr()
	case evmTypes.ModuleName:
		addr = genEvmAddr()
	case axelarnetTypes.ModuleName:
		addr = genCosmosAddr(destination.Name)
	default:
		panic("unexpected module for destination")
	}

	recipient := exported.CrossChainAddress{
		Address: addr,
		Chain:   destination,
	}

	return sender, recipient
}

func genEvmAddr() string {
	return evmUtil.BytesToAddress(rand.Bytes(evmUtil.AddressLength)).Hex()
}

func genBtcAddr() string {
	addr, err := btcutil.NewAddressWitnessScriptHash(rand.Bytes(32), btcTypes.Testnet3.Params())
	if err != nil {
		panic(err)
	}

	return addr.EncodeAddress()
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

func getPrefixByAddress(address string) string {
	switch {
	case strings.HasPrefix(address, "axelar"):
		return "axelar"
	case strings.HasPrefix(address, "terra"):
		return "terra"
	default:
		return ""
	}
}
