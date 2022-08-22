package keeper_test

import (
	mathrand "math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
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
	linkedAddr  = 50
	terra       = nexus.Chain{Name: "terra", Module: axelarnettypes.ModuleName, SupportsForeignAssets: true}
	terraAssets = []string{"uluna", "uusd"}
	avalanche   = nexus.Chain{Name: "avalanche", Module: evmtypes.ModuleName, SupportsForeignAssets: true}
	chains      = []nexus.Chain{evm.Ethereum, axelarnet.Axelarnet, terra, avalanche}
	assets      = append([]string{axelarnet.NativeAsset, "external-erc-20"}, terraAssets...)
	minAmount   = maxAmount / 2
)

func TestUintIntConversion(t *testing.T) {
	maxUint := sdk.NewUintFromBigInt(math.MaxBig256)
	maxInt := sdk.Int(maxUint)

	// ensure max uint can be converted into int without overflow
	assert.True(t, maxInt.IsPositive())
	assert.Equal(t, maxInt.BigInt().BitLen(), 256)
	assert.Panics(t, func() { maxInt.AddRaw(1) })
	assert.Equal(t, maxUint, sdk.Uint(maxInt))
}

func TestComputeTransferFee(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	repeated := 10
	k, ctx := setup(cfg)

	var (
		assetFees map[string]nexus.FeeInfo
	)

	testChain := chains[0].Name
	testAsset := assets[0]

	err := nexus.NewFeeInfo(testChain, testAsset, sdk.ZeroDec(), sdk.ZeroInt(), sdk.ZeroInt()).Validate()
	assert.Nil(t, err)

	err = nexus.NewFeeInfo(testChain, testAsset, sdk.OneDec(), sdk.NewInt(10000), sdk.NewInt(10000)).Validate()
	assert.Nil(t, err)

	// invalid fee
	err = nexus.NewFeeInfo(testChain, testAsset, sdk.NewDecWithPrec(15, 1), sdk.ZeroInt(), sdk.ZeroInt()).Validate()
	assert.Error(t, err)

	// invalid fee
	err = nexus.NewFeeInfo(testChain, testAsset, sdk.ZeroDec(), sdk.NewInt(10), sdk.NewInt(4)).Validate()
	assert.Error(t, err)

	Given("a keeper",
		func() {
			k, ctx = setup(cfg)
			assetFees = make(map[string]nexus.FeeInfo)
		}).
		When("asset fees are registered",
			func() {
				for _, chain := range chains {
					for _, asset := range assets {
						assetFees[chain.Name.String()+"_"+asset] = randFee(chain.Name, asset)
						if err := k.RegisterFee(ctx, chain, assetFees[chain.Name.String()+"_"+asset]); err != nil {
							panic(err)
						}
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

							assetFee := assetFees[sourceChain.Name.String()+"_"+asset]
							assert.Equal(t, sourceChainFee, assetFee)

							coin := sdk.NewCoin(asset, randInt(0, maxAmount*2))
							amount := coin.Amount

							fees, err := k.ComputeTransferFee(ctx, sourceChain, destinationChain, coin)
							assert.Nil(t, err)

							minFee := sourceChainFee.MinFee.Add(destinationChainFee.MinFee)
							feeRate := sourceChainFee.FeeRate.Add(destinationChainFee.FeeRate)
							maxFee := sourceChainFee.MaxFee.Add(destinationChainFee.MaxFee)

							fee := sdk.NewDecFromInt(amount).Mul(feeRate).TruncateInt()
							fee = sdk.MaxInt(minFee, fee)
							fee = sdk.MinInt(maxFee, fee)

							assert.Equal(t, fees.Amount, fee)
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
		expectedTransfers map[nexus.ChainName]transferCounter
		asset             string
	)

	Given("a keeper",
		func() {
			k, ctx = setup(cfg)
		}).
		When("no recipient linked to sender",
			func() {
				sender, _ = makeRandAddresses(k, ctx)
			}).
		Then("enqueue transfer should return error",
			func(t *testing.T) {
				_, err := k.EnqueueForTransfer(ctx, sender, makeRandAmount(randAsset()))
				assert.Error(t, err)
			},
		).Run(t, repeated)

	addressError := Then("enqueue transfer should return error",
		func(t *testing.T) {
			_, err := k.EnqueueTransfer(ctx, sender.Chain, recipient, makeRandAmount(randAsset()))
			assert.Error(t, err)
		},
	)

	Given("a keeper",
		func() {
			k, ctx = setup(cfg)
		}).
		Branch(
			When("link invalid axelarnet address", func() {
				sender, _ = makeRandAddresses(k, ctx)
				recipient = nexus.CrossChainAddress{Chain: exported.Axelarnet, Address: "axelar1t66w8cazua870wu7t2hsffndmy2qy2v556ymndnczs83qpz2h45sq6lq9v"}
			}).
				Then2(addressError),

			When("link invalid terra address", func() {
				sender, _ = makeRandAddresses(k, ctx)
				recipient = nexus.CrossChainAddress{Chain: terra, Address: "terra18zhnqjv70v0d2f8v0s5lape0gr5ua94eqkk8ew"}
			}).
				Then2(addressError),

			When("link invalid EVM address", func() {
				sender, _ = makeRandAddresses(k, ctx)
				recipient = nexus.CrossChainAddress{Chain: evm.Ethereum, Address: rand.HexStr(41)}
			}).
				Then2(addressError),

			When("addresses are blocked", func() {
				sender, recipient = makeRandAddressesForChain(chains[mathrand.Intn(len(chains))], axelarnet.Axelarnet)
				bankK.BlockedAddrFunc = func(addr sdk.AccAddress) bool { return true }
			}).
				Then2(addressError)).
		Run(t, repeated)

	Given("a keeper",
		func() {
			k, ctx = setup(cfg)

			// clear start
			recipients = nil
			senders = nil
			transfers = nil
			expectedTransfers = nil
		}).
		When("senders and recipients are linked", func() {
			for i := 0; i < linkedAddr; i++ {
				s, r := makeRandAddresses(k, ctx)
				senders = append(senders, s)
				recipients = append(recipients, r)

				err := k.LinkAddresses(ctx, s, r)
				assert.NoError(t, err)
			}
		}).Branch(
		When("transfer amounts are smaller than min fee", func() {
			for _, r := range recipients {
				asset := randAsset()
				feeInfo, ok := k.GetFeeInfo(ctx, r.Chain, asset)
				assert.True(t, ok)
				randAmt := sdk.NewCoin(randAsset(), sdk.NewInt(rand.I64Between(1, feeInfo.MinFee.BigInt().Int64()*2)))
				transfers = append(transfers, randAmt)
			}
		}).
			When("enqueue all transfers", func() {
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
		When("transfer amounts are greater than min amount", func() {
			for i := 0; i < len(recipients); i++ {
				asset := randAsset()
				transfers = append(transfers, makeAmountAboveMin(asset))
			}
		}).
			When("enqueue all transfers", func() {
				for i, transfer := range transfers {
					_, err := k.EnqueueForTransfer(ctx, senders[i], transfer)
					assert.NoError(t, err)

					// count transfers
					c := expectedTransfers[recipients[i].Chain.Name]
					feeDue, err := k.ComputeTransferFee(ctx, senders[i].Chain, recipients[i].Chain, transfer)
					assert.Nil(t, err)
					c.fees.Add(feeDue)
					c.coins.Add(transfer.Sub(feeDue))
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
		func() {
			k, ctx = setup(cfg)
		}).
		When("enqueue transfer first time",
			func() {
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
			}).
		When("enqueue transfer second time",
			func() {
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
		func() {
			k, ctx = setup(cfg)
			expectedTransfers = nil
		}).
		When("enqueue transfers",
			func() {
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
			}).
		When("archive pending transfers",
			func() {
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

			if err := k.RegisterAsset(ctx, chain, nexus.NewAsset(asset, isNative)); err != nil {
				panic(err)
			}

			feeInfo := nexus.NewFeeInfo(chain.Name, asset, sdk.ZeroDec(), sdk.NewInt(minAmount), sdk.NewInt(maxAmount))
			if err := k.RegisterFee(ctx, chain, feeInfo); err != nil {
				panic(err)
			}
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
	return sdk.NewCoin(denom, sdk.NewInt(rand.I64Between(minAmount*2, maxAmount*2)))
}

func randFee(chain nexus.ChainName, asset string) nexus.FeeInfo {
	rate := sdk.NewDecWithPrec(sdk.Int(randInt(0, 100)).Int64(), 3)
	min := randInt(0, minAmount)
	max := randInt(min.Int64(), maxAmount)
	return nexus.NewFeeInfo(chain, asset, rate, min, max)
}

func randInt(min int64, max int64) sdk.Int {
	return sdk.NewInt(rand.I64Between(int64(min), int64(max)))
}
