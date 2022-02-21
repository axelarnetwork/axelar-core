package keeper_test

import (
	mathrand "math/rand"
	"testing"

	"github.com/axelarnetwork/axelar-core/utils"

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
	linkedAddr   = 50
	terra        = nexus.Chain{Name: "terra", Module: axelarnettypes.ModuleName, SupportsForeignAssets: true}
	terraAssets  = []string{"uluna", "uusd"}
	avalanche    = nexus.Chain{Name: "avalanche", Module: evmtypes.ModuleName, SupportsForeignAssets: true}
	chains       = []nexus.Chain{evm.Ethereum, axelarnet.Axelarnet, terra, avalanche}
	assets       = append([]string{axelarnet.NativeAsset, "external-erc-20"}, terraAssets...)
	chainFeeInfo = nexus.NewFeeInfo(sdk.ZeroDec(), sdk.NewUint(50000), sdk.NewUint(5000000000))
	minAmount    = maxAmount / 2
)

func TestComputeTransferFee(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	repeated := 10
	k, ctx := setup(cfg)

	var (
		assetFees map[string]nexus.FeeInfo
	)

	for _, chain := range chains {
		k.SetChain(ctx, chain)
		for _, asset := range assets {
			k.RegisterFee(ctx, chain, asset, chainFeeInfo)
		}
	}

	err := nexus.NewFeeInfo(sdk.ZeroDec(), sdk.ZeroUint(), sdk.ZeroUint()).Validate()
	assert.Nil(t, err)

	err = nexus.NewFeeInfo(sdk.OneDec(), sdk.Uint(sdk.NewIntFromUint64(10000)), sdk.Uint(sdk.NewIntFromUint64(10000))).Validate()
	assert.Nil(t, err)

	// invalid fee
	err = nexus.NewFeeInfo(sdk.NewDecWithPrec(15, 1), sdk.ZeroUint(), sdk.ZeroUint()).Validate()
	assert.Error(t, err)

	// invalid fee
	err = nexus.NewFeeInfo(sdk.ZeroDec(), sdk.NewUint(10), sdk.NewUint(4)).Validate()
	assert.Error(t, err)

	Given("a keeper",
		func(t *testing.T) {
			k, ctx = setup(cfg)
			assetFees = make(map[string]nexus.FeeInfo)
		}).
		When("asset fees are registered",
			func(t *testing.T) {
				for _, chain := range chains {
					for _, asset := range assets {
						assetFees[chain.Name+"_"+asset] = randFee()
						k.RegisterFee(ctx, chain, asset, assetFees[chain.Name+"_"+asset])
					}
				}
			}).
		Then("transfer fees should be computed correctly",
			func(t *testing.T) {
				for _, sourceChain := range chains {
					for _, destinationChain := range chains {
						for _, asset := range assets {
							sourceChainFee, found := k.GetFeeInfo(ctx, sourceChain, asset)
							assert.True(t, found)

							destinationChainFee, found := k.GetFeeInfo(ctx, destinationChain, asset)
							assert.True(t, found)

							assetFee := assetFees[sourceChain.Name+"_"+asset]
							assert.Equal(t, sourceChainFee, assetFee)

							coin := sdk.NewCoin(asset, sdk.Int(randUint(0, uint64(maxAmount)*2)))
							amount := sdk.Uint(coin.Amount)

							fees := k.ComputeTransferFee(ctx, sourceChain, destinationChain, coin)

							baseFee := sourceChainFee.MinFee.Add(destinationChainFee.MinFee)

							if amount.LTE(baseFee) {
								assert.Equal(t, sdk.Uint(fees.Amount), baseFee)
							} else {
								assert.Less(t, fees.Amount.Int64(), coin.Amount.Int64())

								remaining := sdk.NewDecFromInt(coin.Amount.Sub(sdk.Int(baseFee)))
								sourceSurcharge := sdk.MinUint(sourceChainFee.MaxFee.Sub(sourceChainFee.MinFee), sdk.Uint(sourceChainFee.FeeRate.Mul(remaining).TruncateInt()))
								destinationSurcharge := sdk.MinUint(destinationChainFee.MaxFee.Sub(destinationChainFee.MinFee), sdk.Uint(destinationChainFee.FeeRate.Mul(remaining).TruncateInt()))
								total := baseFee.Add(sourceSurcharge).Add(destinationSurcharge)

								assert.Equal(t, sdk.Uint(fees.Amount), total)
							}
						}
					}
				}
			},
		).Run(t, repeated)
}

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
				_, err := k.EnqueueForTransfer(ctx, sender, makeRandAmount(randAsset()))
				assert.Error(t, err)
			},
		).Run(t, repeated)

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
				_, err := k.EnqueueForTransfer(ctx, sender, makeRandAmount(randAsset()))
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
		When("transfer amounts are smaller than min fee", func(t *testing.T) {
			for _, r := range recipients {
				asset := randAsset()
				feeInfo, ok := k.GetFeeInfo(ctx, r.Chain, asset)
				assert.True(t, ok)
				randAmt := sdk.NewCoin(randAsset(), sdk.NewInt(rand.I64Between(1, feeInfo.MinFee.BigInt().Int64()*2)))
				transfers = append(transfers, randAmt)
			}
		}).And().
			When("enqueue all transfers", func(t *testing.T) {
				for i, transfer := range transfers {
					_, err := k.EnqueueForTransfer(ctx, senders[i], transfer)
					assert.NoError(t, err)

					// count transfers
					c := expectedTransfers[recipients[i].Chain.Name]
					feeDue := sdk.ZeroInt()
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
						insufficientAmountTransfers := k.GetTransfersForChain(ctx, chain, nexus.InsufficientAmount)

						// total number of pending transfer match
						assert.Equal(t, 0, len(pendingTransfers))
						// total number of insufficient amount transfer match
						assert.Equal(t, len(transfers), len(insufficientAmountTransfers))
						// total fees match
						assert.Equal(t, expected.fees, k.GetTransferFees(ctx))

						// total transfer amount match
						total := sdk.Coins{}
						for _, transfer := range insufficientAmountTransfers {
							total = total.Add(sdk.NewCoin(transfer.Asset.Denom, transfer.Asset.Amount))
						}
						assert.Equal(t, expected.coins, total)
					}
				}),

		When("transfer amounts are greater than min amount", func(t *testing.T) {
			for _, r := range recipients {
				asset := randAsset()
				feeInfo, found := k.GetFeeInfo(ctx, r.Chain, asset)
				assert.True(t, found)
				transfers = append(transfers, makeRandAmount(asset).AddAmount(sdk.Int(feeInfo.MinFee)))
			}
		}).And().
			When("enqueue all transfers", func(t *testing.T) {
				for i, transfer := range transfers {
					_, err := k.EnqueueForTransfer(ctx, senders[i], transfer)
					assert.NoError(t, err)

					// count transfers
					c := expectedTransfers[recipients[i].Chain.Name]
					baseFee := chainFeeInfo.MinFee.MulUint64(2)
					feeDue := sdk.Int(baseFee)
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
						insufficientTransfers := k.GetTransfersForChain(ctx, chain, nexus.InsufficientAmount)

						// total number of insufficient amount transfer match
						assert.Equal(t, 0, len(insufficientTransfers))
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
				_, err = k.EnqueueForTransfer(ctx, sender, firstAmount)
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
				_, err := k.EnqueueForTransfer(ctx, sender, secondAmount)
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

					_, err = k.EnqueueForTransfer(ctx, s, makeAmountAboveMin(randAsset()))
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

	// register asset in ChainState
	for _, chain := range chains {
		k.SetChain(ctx, chain)
		for _, asset := range assets {
			isNative := false
			if chain.Name == axelarnet.Axelarnet.Name && asset == axelarnet.NativeAsset {
				isNative = true
			}
			if chain.Name == terra.Name && utils.IndexOf(terraAssets, asset) != -1 {
				isNative = true
			}

			k.RegisterAsset(ctx, chain, nexus.NewAsset(asset, isNative))
			k.RegisterFee(ctx, chain, asset, chainFeeInfo)
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
	return sdk.NewCoin(denom, sdk.NewInt(rand.I64Between(chainFeeInfo.MinFee.BigInt().Int64()*2, maxAmount)))
}

func randFee() nexus.FeeInfo {
	rate := sdk.NewDecWithPrec(sdk.Int(randUint(0, 100)).Int64(), 3)
	min := randUint(0, uint64(minAmount))
	max := randUint(min.Uint64(), uint64(maxAmount))
	return nexus.NewFeeInfo(rate, min, max)
}

func randUint(min uint64, max uint64) sdk.Uint {
	return sdk.NewUint(uint64(rand.I64Between(int64(min), int64(max))))
}
