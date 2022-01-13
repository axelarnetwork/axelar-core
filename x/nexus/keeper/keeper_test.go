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
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
)

const (
	maxAmount  int64 = 100000000000
	linkedAddr int   = 50
)

var keeper nexusKeeper.Keeper
var feeRate = sdk.NewDecWithPrec(25, 5)

func init() {
	encCfg := app.MakeEncodingConfig()
	nexusSubspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("nexusKey"), sdk.NewKVStoreKey("tNexusKey"), "nexus")
	keeper = nexusKeeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("nexus"), nexusSubspace)

	nexusRouter := types.NewRouter()
	nexusRouter.AddAddressValidator("evm", func(_ sdk.Context, addr nexus.CrossChainAddress) error {
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
		bz, err := sdk.GetFromBech32(addr.Address, "axelar")
		if err != nil {
			return err
		}

		err = sdk.VerifyAddressFormat(bz)
		if err != nil {
			return err
		}

		return nil
	})
	keeper.SetRouter(nexusRouter)
}

func TestLinkAddress(t *testing.T) {
	repeats := 20

	var ctx sdk.Context

	setup := func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper.SetParams(ctx, types.DefaultParams())
		keeper.ActivateChain(ctx, btc.Bitcoin)
		keeper.ActivateChain(ctx, evm.Ethereum)
		keeper.ActivateChain(ctx, axelarnet.Axelarnet)
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

func TestEnqueueForTransfer(t *testing.T) {
	repeats := 20

	var ctx sdk.Context

	setup := func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper.SetParams(ctx, types.DefaultParams())
		keeper.ActivateChain(ctx, btc.Bitcoin)
		keeper.ActivateChain(ctx, evm.Ethereum)
		keeper.ActivateChain(ctx, axelarnet.Axelarnet)
	}

	t.Run("should return error when no recipient linked to sender", testutils.Func(func(t *testing.T) {
		setup()
		sender, _ := makeRandAddressesForChain(btc.Bitcoin, evm.Ethereum)
		_, err := keeper.EnqueueForTransfer(ctx, sender, makeRandAmount(btcTypes.Satoshi), feeRate)
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should successfully enqueue transfer to linked recipient", testutils.Func(func(t *testing.T) {
		setup()
		amounts := make(map[exported.CrossChainAddress]sdk.Coin)
		for i := 0; i < linkedAddr; i++ {
			sender, recipient := makeRandAddressesForChain(btc.Bitcoin, evm.Ethereum)
			amounts[recipient] = makeRandAmount(btcTypes.Satoshi)
			err := keeper.LinkAddresses(ctx, sender, recipient)
			assert.NoError(t, err)
			_, err = keeper.EnqueueForTransfer(ctx, sender, amounts[recipient], feeRate)
			assert.NoError(t, err)
		}

		transfers := keeper.GetTransfersForChain(ctx, evm.Ethereum, exported.Pending)
		assert.Equal(t, len(transfers), len(amounts))
		assert.Equal(t, linkedAddr, len(transfers))

		count := 0
		for _, transfer := range transfers {
			if amount, ok := amounts[transfer.Recipient]; ok {
				count++
				amount = amount.SubAmount(sdk.NewDecFromInt(amount.Amount).Mul(feeRate).TruncateInt())
				assert.Equal(t, transfer.Asset, amount)
			}
		}
		assert.Equal(t, linkedAddr, count)
	}).Repeat(repeats))

	t.Run("should merge transfers to the same recipient", testutils.Func(func(t *testing.T) {
		setup()
		// merge transfers from same sender
		sender, recipient := makeRandAddressesForChain(btc.Bitcoin, evm.Ethereum)
		keeper.LinkAddresses(ctx, sender, recipient)
		firstAmount := makeRandAmount(btcTypes.Satoshi)
		_, err := keeper.EnqueueForTransfer(ctx, sender, firstAmount, feeRate)
		assert.NoError(t, err)
		recp, ok := keeper.GetRecipient(ctx, sender)
		assert.True(t, ok)
		assert.Equal(t, recipient, recp)
		transfers := keeper.GetTransfersForChain(ctx, evm.Ethereum, exported.Pending)
		assert.Len(t, transfers, 1)
		firstFeeDue := sdk.NewDecFromInt(firstAmount.Amount).Mul(feeRate).TruncateInt()
		assert.Equal(t, firstAmount.Amount.Sub(firstFeeDue), transfers[0].Asset.Amount)

		secondAmount := makeRandAmount(btcTypes.Satoshi)
		_, err = keeper.EnqueueForTransfer(ctx, sender, secondAmount, feeRate)
		assert.NoError(t, err)
		recp, ok = keeper.GetRecipient(ctx, sender)
		assert.True(t, ok)
		assert.Equal(t, recipient, recp)
		transfers = keeper.GetTransfersForChain(ctx, evm.Ethereum, exported.Pending)
		assert.Len(t, transfers, 1)
		secondFeeDue := sdk.NewDecFromInt(secondAmount.Amount).Mul(feeRate).TruncateInt()
		total := firstAmount.Amount.Sub(firstFeeDue).Add(secondAmount.Amount.Sub(secondFeeDue))
		assert.Equal(t, total, transfers[0].Asset.Amount)

		// new transfer from some other sender
		sender, recipient = makeRandAddressesForChain(btc.Bitcoin, evm.Ethereum)
		keeper.LinkAddresses(ctx, sender, recipient)
		_, err = keeper.EnqueueForTransfer(ctx, sender, makeRandAmount(btcTypes.Satoshi), feeRate)
		assert.NoError(t, err)
		recp, ok = keeper.GetRecipient(ctx, sender)
		assert.True(t, ok)
		assert.Equal(t, recipient, recp)
		assert.Len(t, keeper.GetTransfersForChain(ctx, evm.Ethereum, exported.Pending), 2)
	}).Repeat(repeats))

	t.Run("should archive pending transfers", testutils.Func(func(t *testing.T) {
		setup()
		for i := 0; i < linkedAddr; i++ {
			sender, recipient := makeRandAddressesForChain(btc.Bitcoin, evm.Ethereum)
			err := keeper.LinkAddresses(ctx, sender, recipient)
			assert.NoError(t, err)
			amount := makeRandAmount(btcTypes.Satoshi)
			_, err = keeper.EnqueueForTransfer(ctx, sender, amount, feeRate)
			assert.NoError(t, err)
		}

		transfers := keeper.GetTransfersForChain(ctx, evm.Ethereum, exported.Pending)

		for _, transfer := range transfers {
			keeper.ArchivePendingTransfer(ctx, transfer)
		}

		archived := keeper.GetTransfersForChain(ctx, evm.Ethereum, exported.Archived)
		assert.Equal(t, linkedAddr, len(archived))

		count := 0
		for _, archive := range archived {
			for _, transfer := range transfers {
				if transfer.Recipient.Address == archive.Recipient.Address {
					count++
					assert.Equal(t, archive.Asset, transfer.Asset)
				}
			}
		}
		assert.Equal(t, linkedAddr, count)
		assert.Equal(t, 0, len(keeper.GetTransfersForChain(ctx, evm.Ethereum, exported.Pending)))
	}).Repeat(repeats))
}

func TestSetChainGetChain_MixCaseChainName(t *testing.T) {
	chainName := strings.ToUpper(rand.StrBetween(5, 10)) + strings.ToLower(rand.StrBetween(5, 10))
	chain := exported.Chain{
		Name:                  chainName,
		NativeAsset:           rand.Str(3),
		SupportsForeignAssets: true,
		Module:                rand.Str(10),
	}

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
	chain := exported.Chain{
		Name:                  chainName,
		NativeAsset:           rand.Str(3),
		SupportsForeignAssets: true,
		Module:                rand.Str(10),
	}

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
	chain := exported.Chain{
		Name:                  chainName,
		NativeAsset:           rand.Str(3),
		SupportsForeignAssets: true,
		Module:                rand.Str(10),
	}

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper.SetChain(ctx, chain)

	actual, ok := keeper.GetChain(ctx, strings.ToUpper(chainName))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)

	actual, ok = keeper.GetChain(ctx, strings.ToLower(chainName))

	assert.True(t, ok)
	assert.Equal(t, chain, actual)
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
