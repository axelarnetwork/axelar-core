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
	sender, recipient := makeRandAddresses()

	keeper.LinkAddresses(ctx, sender, recipient)
	result, ok := keeper.getRecipient(ctx, sender)
	assert.True(t, ok)
	assert.Equal(t, recipient, result)

	sender.Address = testutils.RandString(20)
	result, ok = keeper.getRecipient(ctx, sender)
	assert.False(t, ok)
	assert.Equal(t, exported.CrossChainAddress{}, result)
}

func TestPrepare(t *testing.T) {
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	sender, _ := makeRandAddresses()

	err := keeper.PrepareForTransfer(ctx, sender, makeRandAmount(makeRandomDenom()))
	assert.Error(t, err)
	chain := exported.Chain(testutils.RandIntBetween(0, exported.ConnectedChainCount))
	denom := makeRandomDenom()
	senders := make([]exported.CrossChainAddress, 0)

	for i := 0; i < linkedAddr; i++ {
		sender, recipient := makeRandAddressesForChain(chain)
		senders = append(senders, sender)
		keeper.LinkAddresses(ctx, sender, recipient)
		err = keeper.PrepareForTransfer(ctx, sender, makeRandAmount(denom))
		assert.NoError(t, err)
	}

	transfers := keeper.GetPendingTransfersForChain(ctx, chain)
	assert.Equal(t, linkedAddr, len(transfers))

	for _, sender := range senders {
		err = keeper.PrepareForTransfer(ctx, sender, sdk.NewInt64Coin(denom, 10))
		assert.NoError(t, err)
	}

	transfersUpdated := keeper.GetPendingTransfersForChain(ctx, chain)
	assert.Equal(t, linkedAddr, len(transfersUpdated))

	count := 0
	for _, transfer1 := range transfers {
		for _, transfer2 := range transfersUpdated {
			if transfer1.Recipient.Address == transfer2.Recipient.Address {
				count++
				assert.Equal(t, transfer2.Amount, transfer1.Amount.Add(sdk.NewInt64Coin(denom, 10)))
			}
		}
	}
	assert.Equal(t, linkedAddr, count)
}

func TestArchive(t *testing.T) {

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	chain := exported.Chain(testutils.RandIntBetween(0, exported.ConnectedChainCount))
	denom := makeRandomDenom()
	recipients := make([]exported.CrossChainAddress, 0)

	for i := 0; i < linkedAddr; i++ {
		sender, recipient := makeRandAddressesForChain(chain)
		recipients = append(recipients, recipient)
		keeper.LinkAddresses(ctx, sender, recipient)
		err := keeper.PrepareForTransfer(ctx, sender, makeRandAmount(denom))
		assert.NoError(t, err)
	}

	transfers := keeper.GetPendingTransfersForChain(ctx, chain)

	for _, recipient := range recipients {
		keeper.ArchivePendingTransfers(ctx, recipient)
	}

	archived := keeper.GetArchivedTransfersForChain(ctx, chain)
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

	return sdk.NewCoin(denom, sdk.NewInt(testutils.RandIntBetween(0, maxAmount)))
}

func makeRandAddresses() (exported.CrossChainAddress, exported.CrossChainAddress) {
	sender := exported.CrossChainAddress{
		Address: testutils.RandString(addrMaxLength),
		Chain:   exported.Chain(testutils.RandIntBetween(0, exported.ConnectedChainCount)),
	}
	recipient := exported.CrossChainAddress{
		Address: testutils.RandString(addrMaxLength),
		Chain:   exported.Chain(testutils.RandIntBetween(0, exported.ConnectedChainCount)),
	}

	return sender, recipient
}

func makeRandAddressesForChain(chain exported.Chain) (exported.CrossChainAddress, exported.CrossChainAddress) {
	sender := exported.CrossChainAddress{
		Address: testutils.RandString(addrMaxLength),
		Chain:   chain,
	}
	recipient := exported.CrossChainAddress{
		Address: testutils.RandString(addrMaxLength),
		Chain:   chain,
	}

	return sender, recipient
}
