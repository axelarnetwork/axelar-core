package keeper_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	evmUtil "github.com/ethereum/go-ethereum/common"
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
	"github.com/axelarnetwork/axelar-core/x/nexus/types/mock"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
)

const (
	addrMaxLength int   = 20
	maxAmount     int64 = 100000000000
	linkedAddr    int   = 50
)

var keeper nexusKeeper.Keeper
var feeRate = sdk.NewDecWithPrec(25, 5)

func init() {
	encCfg := app.MakeEncodingConfig()
	nexusSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("nexusKey"), sdk.NewKVStoreKey("tNexusKey"), "nexus")
	axelarnetKeeper := &mock.AxelarnetKeeperMock{
		GetFeeCollectorFunc: func(sdk.Context) (sdk.AccAddress, bool) { return rand.AccAddr(), true },
	}
	keeper = nexusKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("nexus"), nexusSubspace, axelarnetKeeper)

	nexusRouter := types.NewRouter()
	nexusRouter.AddRoute("evm", func(_ sdk.Context, addr nexus.CrossChainAddress) error {
		if !evmUtil.IsHexAddress(addr.Address) {
			return fmt.Errorf("not an hex address")
		}

		return nil
	}).AddRoute("bitcoin", func(ctx sdk.Context, addr nexus.CrossChainAddress) error {
		if _, err := btcutil.DecodeAddress(addr.Address, btcTypes.Testnet3.Params()); err != nil {
			return err
		}

		return nil
	}).AddRoute("axelarnet", func(ctx sdk.Context, addr nexus.CrossChainAddress) error {
		if _, err := sdk.GetFromBech32(addr.Address, "axelar"); err != nil {
			return err
		}

		return nil
	})
	keeper.SetRouter(nexusRouter)
}

func TestValidAddresses(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper.SetParams(ctx, types.DefaultParams())

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
}

func TestInvalidAddresses(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper.SetParams(ctx, types.DefaultParams())

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
}

func TestLinkNoForeignAssetSupport(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper.SetParams(ctx, types.DefaultParams())

	sender, recipient := makeRandAddressesForChain(btc.Bitcoin, evm.Ethereum)
	err := keeper.LinkAddresses(ctx, sender, recipient)
	assert.NoError(t, err)
	err = keeper.EnqueueForTransfer(ctx, sender, makeRandAmount(makeRandomDenom()), feeRate)
	assert.Error(t, err)
}

func TestLinkSuccess(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper.SetParams(ctx, types.DefaultParams())

	sender, recipient := makeRandAddressesForChain(btc.Bitcoin, evm.Ethereum)
	err := keeper.LinkAddresses(ctx, sender, recipient)
	assert.NoError(t, err)
	err = keeper.EnqueueForTransfer(ctx, sender, makeRandAmount(btcTypes.Satoshi), feeRate)
	assert.NoError(t, err)
	recp, ok := keeper.GetRecipient(ctx, sender)
	assert.True(t, ok)
	assert.Equal(t, recipient, recp)

	sender.Address = rand.Str(20)
	err = keeper.EnqueueForTransfer(ctx, sender, makeRandAmount(btcTypes.Satoshi), feeRate)
	assert.Error(t, err)
	recp, ok = keeper.GetRecipient(ctx, sender)
	assert.False(t, ok)
	assert.NotEqual(t, recipient, recp)
}

func TestPrepareNoLink(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper.SetParams(ctx, types.DefaultParams())

	sender, _ := makeRandAddressesForChain(btc.Bitcoin, evm.Ethereum)
	err := keeper.EnqueueForTransfer(ctx, sender, makeRandAmount(btcTypes.Satoshi), feeRate)
	assert.Error(t, err)
}

func TestPrepareSuccess(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper.SetParams(ctx, types.DefaultParams())

	amounts := make(map[exported.CrossChainAddress]sdk.Coin)
	for i := 0; i < linkedAddr; i++ {
		sender, recipient := makeRandAddressesForChain(btc.Bitcoin, evm.Ethereum)
		amounts[recipient] = makeRandAmount(btcTypes.Satoshi)
		err := keeper.LinkAddresses(ctx, sender, recipient)
		assert.NoError(t, err)
		err = keeper.EnqueueForTransfer(ctx, sender, amounts[recipient], feeRate)
		assert.NoError(t, err)
	}

	transfers := keeper.GetTransfersForChain(ctx, evm.Ethereum, exported.Pending)
	assert.Equal(t, len(transfers), len(amounts))
	assert.Equal(t, linkedAddr, len(transfers))

	count := 0
	for _, transfer := range transfers {
		amount, ok := amounts[transfer.Recipient]
		if ok {
			count++
			amount.Amount = amount.Amount.Sub(sdk.NewDecFromInt(amount.Amount).Mul(feeRate).TruncateInt())
			assert.Equal(t, transfer.Asset, amount)
		}
	}
	assert.Equal(t, linkedAddr, count)
}

func TestArchive(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper.SetParams(ctx, types.DefaultParams())

	for i := 0; i < linkedAddr; i++ {
		sender, recipient := makeRandAddressesForChain(btc.Bitcoin, evm.Ethereum)
		err := keeper.LinkAddresses(ctx, sender, recipient)
		assert.NoError(t, err)
		amount := makeRandAmount(btcTypes.Satoshi)
		err = keeper.EnqueueForTransfer(ctx, sender, amount, feeRate)
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
}

func TestTotalInvalid(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper.SetParams(ctx, types.DefaultParams())
	btcSender, btcRecipient := makeRandAddressesForChain(btc.Bitcoin, evm.Ethereum)
	err := keeper.LinkAddresses(ctx, btcSender, btcRecipient)
	assert.NoError(t, err)
	ethSender, ethRecipient := makeRandAddressesForChain(evm.Ethereum, btc.Bitcoin)
	err = keeper.LinkAddresses(ctx, ethSender, ethRecipient)
	assert.NoError(t, err)

	err = keeper.EnqueueForTransfer(ctx, btcSender, makeRandAmount(btcTypes.Satoshi), feeRate)
	assert.NoError(t, err)
	transfer := keeper.GetTransfersForChain(ctx, evm.Ethereum, exported.Pending)[0]
	keeper.ArchivePendingTransfer(ctx, transfer)
	total := transfer.Asset.Amount.Int64()
	amount := sdk.NewCoin(btcTypes.Satoshi, sdk.NewInt(total+rand.I64Between(1, 100000)))
	err = keeper.EnqueueForTransfer(ctx, ethSender, amount, feeRate)
	assert.Error(t, err)
}

func TestTotalSucess(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper.SetParams(ctx, types.DefaultParams())
	btcSender, btcRecipient := makeRandAddressesForChain(btc.Bitcoin, evm.Ethereum)
	err := keeper.LinkAddresses(ctx, btcSender, btcRecipient)
	assert.NoError(t, err)

	ethSender, ethRecipient := makeRandAddressesForChain(evm.Ethereum, btc.Bitcoin)
	err = keeper.LinkAddresses(ctx, ethSender, ethRecipient)
	assert.NoError(t, err)

	err = keeper.EnqueueForTransfer(ctx, btcSender, makeRandAmount(btcTypes.Satoshi), feeRate)
	assert.NoError(t, err)
	transfer := keeper.GetTransfersForChain(ctx, evm.Ethereum, exported.Pending)[0]
	keeper.ArchivePendingTransfer(ctx, transfer)
	total := transfer.Asset.Amount.Int64()
	amount := sdk.NewCoin(btcTypes.Satoshi, sdk.NewInt(rand.I64Between(1, total)))
	err = keeper.EnqueueForTransfer(ctx, ethSender, amount, feeRate)
	assert.NoError(t, err)
	amount = sdk.NewCoin(btcTypes.Satoshi, sdk.NewInt(total))
	err = keeper.EnqueueForTransfer(ctx, ethSender, amount, feeRate)
	assert.Error(t, err)
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
