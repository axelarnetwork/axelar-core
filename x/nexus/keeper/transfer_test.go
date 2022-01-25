package keeper_test

import (
	mathrand "math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	axelarnettypes "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	. "github.com/axelarnetwork/utils/test"
)

var (
	linkedAddr = 50
	terra      = nexus.Chain{Name: "terra", Module: axelarnettypes.ModuleName, SupportsForeignAssets: true}
	avalanche  = nexus.Chain{Name: "avalanche", Module: evmtypes.ModuleName, SupportsForeignAssets: true}
	minAmount  = sdk.NewInt(10000000)
	chains     = []nexus.Chain{evm.Ethereum, axelarnet.Axelarnet, terra, avalanche}
	assets     = []string{"uaxl", "uusd", "uluna", "external-erc-20"}
)

func TestTransfer(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	repeated := 10

	type transferCounter struct {
		// total number of pending transfer
		count int
		// total transfer amount
		coins sdk.Coins
		fees  sdk.Coins
	}

	var (
		k   nexusKeeper.Keeper
		ctx sdk.Context

		sender    nexus.CrossChainAddress
		recipient nexus.CrossChainAddress

		senders    []nexus.CrossChainAddress
		recipients []nexus.CrossChainAddress
		transfers  []sdk.Coin
		// track total transfers, amounts and fees per chain
		expectedTransfers map[string]transferCounter
		asset             string
	)

	Given("a keeper",
		func(t *testing.T) {
			k, ctx = setup(cfg)
		}).
		When("no recipient linked to sender",
			func(t *testing.T) {
				sender, _ = makeRandAddresses(k, ctx)
			}).
		Then("enqueue transfer should return error",
			func(t *testing.T) {
				_, err := k.EnqueueForTransfer(ctx, sender, makeRandAmount(randAsset()), feeRate)
				assert.Error(t, err)
			},
		).Run(t, repeated)

	Given("a keeper",
		func(t *testing.T) {
			k, ctx = setup(cfg)

			// clear start
			recipients = nil
			senders = nil
			transfers = nil
			expectedTransfers = nil
		}).
		When("senders and recipients are linked", func(t *testing.T) {
			for i := 0; i < linkedAddr; i++ {
				s, r := makeRandAddresses(k, ctx)
				senders = append(senders, s)
				recipients = append(recipients, r)

				err := k.LinkAddresses(ctx, s, r)
				assert.NoError(t, err)
			}
		}).Branch(
		When("transfer amounts are smaller than min amount", func(t *testing.T) {
			for _, r := range recipients {
				asset := randAsset()
				min := k.GetMinAmount(ctx, r.Chain, asset)
				randAmt := sdk.NewCoin(randAsset(), sdk.NewInt(rand.I64Between(1, min.Int64())))
				transfers = append(transfers, randAmt)
			}
		}).And().
			When("enqueue all transfers", func(t *testing.T) {
				for i, transfer := range transfers {
					_, err := k.EnqueueForTransfer(ctx, senders[i], transfer, feeRate)
					assert.NoError(t, err)

					// count transfers
					c := expectedTransfers[recipients[i].Chain.Name]
					feeDue := sdk.NewDecFromInt(transfer.Amount).Mul(feeRate).TruncateInt()
					c.fees.Add(sdk.NewCoin(transfer.Denom, feeDue))
					c.coins.Add(sdk.NewCoin(transfer.Denom, transfer.Amount.Sub(feeDue)))
					c.count += 1
				}
			}).
			Then("return 0 pending transfers and collect fees",
				func(t *testing.T) {
					for chainName, expected := range expectedTransfers {
						chain, _ := k.GetChain(ctx, chainName)
						pendingTransfers := k.GetTransfersForChain(ctx, chain, nexus.Pending)

						// total number of pending transfer match
						assert.Equal(t, 0, len(pendingTransfers))
						// total fees match
						assert.Equal(t, expected.fees, k.GetTransferFees(ctx))
					}
				}),

		When("transfer amounts are greater than min amount", func(t *testing.T) {
			for _, r := range recipients {
				asset := randAsset()
				min := k.GetMinAmount(ctx, r.Chain, asset)
				transfers = append(transfers, makeRandAmount(asset).AddAmount(min))
			}
		}).And().
			When("enqueue all transfers", func(t *testing.T) {
				for i, transfer := range transfers {
					_, err := k.EnqueueForTransfer(ctx, senders[i], transfer, feeRate)
					assert.NoError(t, err)

					// count transfers
					c := expectedTransfers[recipients[i].Chain.Name]
					feeDue := sdk.NewDecFromInt(transfer.Amount).Mul(feeRate).TruncateInt()
					c.fees.Add(sdk.NewCoin(transfer.Denom, feeDue))
					c.coins.Add(sdk.NewCoin(transfer.Denom, transfer.Amount.Sub(feeDue)))
					c.count += 1
				}
			}).
			Then("return all pending transfers and collect fees",
				func(t *testing.T) {
					for chainName, expected := range expectedTransfers {
						chain, _ := k.GetChain(ctx, chainName)
						pendingTransfers := k.GetTransfersForChain(ctx, chain, nexus.Pending)

						// total number of pending transfer match
						assert.Equal(t, expected.count, len(pendingTransfers))
						// total amount match
						total := sdk.Coins{}
						for _, transfer := range pendingTransfers {
							total = total.Add(sdk.NewCoin(transfer.Asset.Denom, transfer.Asset.Amount))
						}
						// total transfer amount match
						assert.Equal(t, expected.coins, total)
						// total fees match
						assert.Equal(t, expected.fees, k.GetTransferFees(ctx))
					}
				}),
	).Run(t, repeated)

	Given("a keeper with registered assets",
		func(t *testing.T) {
			k, ctx = setup(cfg)
		}).
		When("enqueue transfer first time",
			func(t *testing.T) {
				sender, recipient = makeRandAddresses(k, ctx)
				err := k.LinkAddresses(ctx, sender, recipient)
				assert.NoError(t, err)

				asset = randAsset()
				firstAmount := makeAmountAboveMin(asset)
				_, err = k.EnqueueForTransfer(ctx, sender, firstAmount, feeRate)
				assert.NoError(t, err)

				actualRecipient, ok := k.GetRecipient(ctx, sender)
				assert.True(t, ok)
				assert.Equal(t, recipient, actualRecipient)

				actualTransfers := k.GetTransfersForChain(ctx, recipient.Chain, nexus.Pending)
				assert.Len(t, actualTransfers, 1)
			}).And().
		When("enqueue transfer second time",
			func(t *testing.T) {
				secondAmount := makeAmountAboveMin(asset)
				_, err := k.EnqueueForTransfer(ctx, sender, secondAmount, feeRate)
				assert.NoError(t, err)
			}).
		Then("should merge transfers to the same recipient",
			func(t *testing.T) {
				actualRecipient, ok := k.GetRecipient(ctx, sender)
				assert.True(t, ok)
				assert.Equal(t, recipient, actualRecipient)

				actualTransfers := k.GetTransfersForChain(ctx, recipient.Chain, nexus.Pending)
				assert.Len(t, actualTransfers, 1)
			},
		).Run(t, repeated)

	Given("a keeper with registered assets",
		func(t *testing.T) {
			k, ctx = setup(cfg)
			expectedTransfers = nil
		}).
		When("enqueue transfers",
			func(t *testing.T) {
				for i := 0; i < linkedAddr; i++ {
					s, r := makeRandAddresses(k, ctx)

					err := k.LinkAddresses(ctx, s, r)
					assert.NoError(t, err)

					_, err = k.EnqueueForTransfer(ctx, s, makeAmountAboveMin(randAsset()), feeRate)
					assert.NoError(t, err)

					c := expectedTransfers[recipients[i].Chain.Name]
					c.count += 1
				}

				for chainName, expected := range expectedTransfers {
					chain, _ := k.GetChain(ctx, chainName)
					assert.Equal(t, expected.count, len(k.GetTransfersForChain(ctx, chain, nexus.Pending)))
				}
			}).And().
		When("archive pending transfers",
			func(t *testing.T) {
				for chainName := range expectedTransfers {
					chain, _ := k.GetChain(ctx, chainName)
					for _, transfer := range k.GetTransfersForChain(ctx, chain, nexus.Pending) {
						k.ArchivePendingTransfer(ctx, transfer)
					}
				}

			}).
		Then("should return 0 pending transfer",
			func(t *testing.T) {
				for chainName, expected := range expectedTransfers {
					chain, _ := k.GetChain(ctx, chainName)
					pendingTransfers := k.GetTransfersForChain(ctx, chain, nexus.Pending)
					assert.Equal(t, 0, len(pendingTransfers))

					archivedTransfers := k.GetTransfersForChain(ctx, chain, nexus.Archived)
					assert.Equal(t, expected.count, len(archivedTransfers))
				}

			},
		).Run(t, repeated)
}

func setup(cfg params.EncodingConfig) (nexusKeeper.Keeper, sdk.Context) {
	subspace := paramstypes.NewSubspace(cfg.Codec, cfg.Amino, sdk.NewKVStoreKey("nexusKey"), sdk.NewKVStoreKey("tNexusKey"), "nexus")
	k := nexusKeeper.NewKeeper(cfg.Codec, sdk.NewKVStoreKey(types.StoreKey), subspace)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

	k.SetParams(ctx, types.DefaultParams())
	k.SetRouter(addressValidator())

	// register native asset
	_ = k.RegisterNativeAsset(ctx, axelarnet.Axelarnet, axelarnet.Uaxl)
	_ = k.RegisterNativeAsset(ctx, terra, "uusd")
	_ = k.RegisterNativeAsset(ctx, terra, "uluna")

	// register asset in ChainState
	for _, chain := range chains {
		k.SetChain(ctx, chain)
		for _, asset := range assets {
			k.RegisterAsset(ctx, chain, nexus.NewAsset(asset, minAmount))
		}
		k.ActivateChain(ctx, chain)
	}

	return k, ctx
}

func makeRandAddresses(k nexusKeeper.Keeper, ctx sdk.Context) (nexus.CrossChainAddress, nexus.CrossChainAddress) {
	chains := k.GetChains(ctx)
	return makeRandAddressesForChain(chains[mathrand.Intn(len(chains))], chains[mathrand.Intn(len(chains))])
}

func randAsset() string {
	return assets[mathrand.Intn(len(assets))]
}

func makeAmountAboveMin(denom string) sdk.Coin {
	return sdk.NewCoin(denom, sdk.NewInt(rand.I64Between(minAmount.Int64(), maxAmount)))
}
