package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/utils/denom"
	"github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/balance/types"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
)

const (
	addrMaxLength int   = 20
	maxAmount     int64 = 100000000000
	linkedAddr    int   = 50
)

var keeper Keeper

func init() {
	cdc := testutils.Codec()
	balanceSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("balanceKey"), sdk.NewKVStoreKey("tbalanceKey"), "balance")
	keeper = NewKeeper(cdc, sdk.NewKVStoreKey("testKey"), balanceSubspace)
}

func TestLinkInvalidChain(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	keeper.SetParams(ctx, types.DefaultParams())

	sender, recipient := makeRandAddressesForChain(makeRandomChain(), exported.NONE)
	err := keeper.LinkAddresses(ctx, sender, recipient)
	assert.Error(t, err)
}

func TestLinkNoForeignAssetSupport(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	keeper.SetParams(ctx, types.DefaultParams())

	sender, recipient := makeRandAddressesForChain(exported.Bitcoin, exported.Ethereum)
	keeper.LinkAddresses(ctx, sender, recipient)
	err := keeper.EnqueueForTransfer(ctx, sender, makeRandAmount(makeRandomDenom()))
	assert.Error(t, err)
}

func TestLinkSuccess(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	keeper.SetParams(ctx, types.DefaultParams())

	sender, recipient := makeRandAddressesForChain(exported.Bitcoin, exported.Ethereum)
	keeper.LinkAddresses(ctx, sender, recipient)
	err := keeper.EnqueueForTransfer(ctx, sender, makeRandAmount(denom.Satoshi))
	assert.NoError(t, err)
	recp, ok := keeper.GetRecipient(ctx, sender)
	assert.True(t, ok)
	assert.Equal(t, recipient, recp)

	sender.Address = testutils.RandString(20)
	err = keeper.EnqueueForTransfer(ctx, sender, makeRandAmount(denom.Satoshi))
	assert.Error(t, err)
	recp, ok = keeper.GetRecipient(ctx, sender)
	assert.False(t, ok)
	assert.NotEqual(t, recipient, recp)
}

func TestPrepareNoLink(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	keeper.SetParams(ctx, types.DefaultParams())

	sender, _ := makeRandAddressesForChain(exported.Bitcoin, exported.Ethereum)
	err := keeper.EnqueueForTransfer(ctx, sender, makeRandAmount(denom.Satoshi))
	assert.Error(t, err)
}

func TestPrepareSuccess(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	keeper.SetParams(ctx, types.DefaultParams())

	amounts := make(map[exported.CrossChainAddress]sdk.Coin)
	for i := 0; i < linkedAddr; i++ {
		sender, recipient := makeRandAddressesForChain(exported.Bitcoin, exported.Ethereum)
		amounts[recipient] = makeRandAmount(denom.Satoshi)
		keeper.LinkAddresses(ctx, sender, recipient)
		err := keeper.EnqueueForTransfer(ctx, sender, amounts[recipient])
		assert.NoError(t, err)
	}

	transfers := keeper.GetPendingTransfersForChain(ctx, exported.Ethereum)
	assert.Equal(t, len(transfers), len(amounts))
	assert.Equal(t, linkedAddr, len(transfers))

	count := 0
	for _, transfer := range transfers {
		amount, ok := amounts[transfer.Recipient]
		if ok {
			count++
			assert.Equal(t, transfer.Amount, amount)
		}
	}
	assert.Equal(t, linkedAddr, count)
}

func TestArchive(t *testing.T) {

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	keeper.SetParams(ctx, types.DefaultParams())

	recipients := make([]exported.CrossChainAddress, 0)
	var total uint64 = 0

	for i := 0; i < linkedAddr; i++ {
		sender, recipient := makeRandAddressesForChain(exported.Bitcoin, exported.Ethereum)
		recipients = append(recipients, recipient)
		keeper.LinkAddresses(ctx, sender, recipient)
		amount := makeRandAmount(denom.Satoshi)
		err := keeper.EnqueueForTransfer(ctx, sender, amount)
		assert.NoError(t, err)
		total += amount.Amount.Uint64()
	}

	transfers := keeper.GetPendingTransfersForChain(ctx, exported.Ethereum)

	for _, transfer := range transfers {
		keeper.ArchivePendingTransfer(ctx, transfer)
	}

	archived := keeper.GetArchivedTransfersForChain(ctx, exported.Ethereum)
	assert.Equal(t, linkedAddr, len(archived))

	count := 0
	for _, archive := range archived {
		for _, transfer := range transfers {
			if transfer.Recipient.Address == archive.Recipient.Address {
				count++
				assert.Equal(t, archive.Amount, transfer.Amount)
			}
		}
	}
	assert.Equal(t, linkedAddr, count)
	assert.Equal(t, 0, len(keeper.GetPendingTransfersForChain(ctx, exported.Ethereum)))
}

func TestTotalInvalid(t *testing.T) {

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	keeper.SetParams(ctx, types.DefaultParams())
	btcSender, btcRecipient := makeRandAddressesForChain(exported.Bitcoin, exported.Ethereum)
	if err := keeper.LinkAddresses(ctx, btcSender, btcRecipient); err != nil {
		panic(err)
	}
	ethSender, ethRecipient := makeRandAddressesForChain(exported.Ethereum, exported.Bitcoin)
	if err := keeper.LinkAddresses(ctx, ethSender, ethRecipient); err != nil {
		panic(err)
	}

	err := keeper.EnqueueForTransfer(ctx, btcSender, makeRandAmount(denom.Satoshi))
	assert.NoError(t, err)
	transfer := keeper.GetPendingTransfersForChain(ctx, exported.Ethereum)[0]
	keeper.ArchivePendingTransfer(ctx, transfer)
	total := transfer.Amount.Amount.Int64()
	amount := sdk.NewCoin(denom.Satoshi, sdk.NewInt(total+testutils.RandIntBetween(1, 100000)))
	err = keeper.EnqueueForTransfer(ctx, ethSender, amount)
	assert.Error(t, err)
}

func TestTotalSucess(t *testing.T) {

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	keeper.SetParams(ctx, types.DefaultParams())
	btcSender, btcRecipient := makeRandAddressesForChain(exported.Bitcoin, exported.Ethereum)
	if err := keeper.LinkAddresses(ctx, btcSender, btcRecipient); err != nil {
		panic(err)
	}
	ethSender, ethRecipient := makeRandAddressesForChain(exported.Ethereum, exported.Bitcoin)
	if err := keeper.LinkAddresses(ctx, ethSender, ethRecipient); err != nil {
		panic(err)
	}

	err := keeper.EnqueueForTransfer(ctx, btcSender, makeRandAmount(denom.Satoshi))
	assert.NoError(t, err)
	transfer := keeper.GetPendingTransfersForChain(ctx, exported.Ethereum)[0]
	keeper.ArchivePendingTransfer(ctx, transfer)
	total := transfer.Amount.Amount.Int64()
	amount := sdk.NewCoin(denom.Satoshi, sdk.NewInt(testutils.RandIntBetween(1, total)))
	err = keeper.EnqueueForTransfer(ctx, ethSender, amount)
	assert.NoError(t, err)
	amount = sdk.NewCoin(denom.Satoshi, sdk.NewInt(total))
	err = keeper.EnqueueForTransfer(ctx, ethSender, amount)
	assert.Error(t, err)
}

func makeRandomDenom() string {

	alphabet := []rune("abcdefghijklmnopqrstuvwxyz")
	denom := ""
	denom = denom + string(alphabet[testutils.RandIntBetween(0, int64(len(alphabet)))])
	denom = denom + string(alphabet[testutils.RandIntBetween(0, int64(len(alphabet)))])
	denom = denom + string(alphabet[testutils.RandIntBetween(0, int64(len(alphabet)))])

	return denom
}

func makeRandAmount(denom string) sdk.Coin {

	return sdk.NewCoin(denom, sdk.NewInt(testutils.RandIntBetween(1, maxAmount)))
}

func makeRandomChain() exported.Chain {
	return exported.Chain(testutils.RandIntBetween(1, exported.ConnectedChainCount))
}

func makeRandAddressesForChain(origin, distination exported.Chain) (exported.CrossChainAddress, exported.CrossChainAddress) {
	sender := exported.CrossChainAddress{
		Address: testutils.RandString(addrMaxLength),
		Chain:   origin,
	}
	recipient := exported.CrossChainAddress{
		Address: testutils.RandString(addrMaxLength),
		Chain:   distination,
	}

	return sender, recipient
}
