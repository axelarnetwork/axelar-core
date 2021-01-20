package keeper

import (
	"testing"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/x/balance/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/stretchr/testify/assert"
)

const (
	addrMaxLength int   = 20
	denomLength   int   = 3
	maxAmount     int64 = 100000000000
	linkedAddr    int   = 50
)

var keeper Keeper

func init() {
	cdc := testutils.Codec()
	keeper = NewKeeper(cdc, sdk.NewKVStoreKey("testKey"))
}

func TestLink(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	sender, recipient := makeRandAddressesForChain(makeRandomChain(), makeRandomChain())

	keeper.LinkAddresses(ctx, sender, recipient)
	err := keeper.PrepareForTransfer(ctx, sender, makeRandAmount(makeRandomDenom()))
	assert.NoError(t, err)

	sender.Address = testutils.RandString(20)
	err = keeper.PrepareForTransfer(ctx, sender, makeRandAmount(makeRandomDenom()))
	assert.Error(t, err)
}

func TestPrepare(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	sender, _ := makeRandAddressesForChain(makeRandomChain(), makeRandomChain())

	err := keeper.PrepareForTransfer(ctx, sender, makeRandAmount(makeRandomDenom()))
	assert.Error(t, err)
	destination := makeRandomChain()
	amounts := make(map[exported.CrossChainAddress]sdk.Coin)

	for i := 0; i < linkedAddr; i++ {
		sender, recipient := makeRandAddressesForChain(makeRandomChain(), destination)
		amounts[recipient] = makeRandAmount(makeRandomDenom())
		keeper.LinkAddresses(ctx, sender, recipient)
		err = keeper.PrepareForTransfer(ctx, sender, amounts[recipient])
		assert.NoError(t, err)
	}

	transfers := keeper.GetPendingTransfersForChain(ctx, destination)
	assert.Equal(t, len(transfers), len(amounts))
	assert.Equal(t, linkedAddr, len(transfers))

	count := 0
	for _, x := range transfers {
		amount, ok := amounts[x.Recipient]
		if ok {
			count++
			assert.Equal(t, x.Amount, amount)
		}
	}
	assert.Equal(t, linkedAddr, count)
}

func TestArchive(t *testing.T) {

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	destination := makeRandomChain()
	denom := makeRandomDenom()
	recipients := make([]exported.CrossChainAddress, 0)

	for i := 0; i < linkedAddr; i++ {
		sender, recipient := makeRandAddressesForChain(makeRandomChain(), destination)
		recipients = append(recipients, recipient)
		keeper.LinkAddresses(ctx, sender, recipient)
		err := keeper.PrepareForTransfer(ctx, sender, makeRandAmount(denom))
		assert.NoError(t, err)
	}

	transfers := keeper.GetPendingTransfersForChain(ctx, destination)

	for _, transfer := range transfers {
		keeper.ArchivePendingTransfer(ctx, transfer)
	}

	archived := keeper.GetArchivedTransfersForChain(ctx, destination)
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
	assert.Equal(t, 0, len(keeper.GetPendingTransfersForChain(ctx, destination)))

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
