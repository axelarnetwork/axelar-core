package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	axelarnetkeeper "github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	axelarnetTypes "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	bitcoin "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmkeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

func setup() (sdk.Context, Keeper) {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "nexus")

	keeper := NewKeeper(
		encodingConfig.Codec,
		sdk.NewKVStoreKey(types.StoreKey),
		subspace,
	)

	return ctx, keeper
}

func getRandomAxelarnetAddress() exported.CrossChainAddress {
	sdk.GetConfig().SetBech32PrefixForAccount("axelar", "axelar")
	return exported.CrossChainAddress{
		Chain:   axelarnet.Axelarnet,
		Address: rand.AccAddr().String(),
	}
}

func getRandomEthereumAddress() exported.CrossChainAddress {
	return exported.CrossChainAddress{
		Chain:   evm.Ethereum,
		Address: common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(),
	}
}

func randFee(chain string, asset string) exported.FeeInfo {
	rate := sdk.NewDecWithPrec(sdk.Int(randInt(0, 100)).Int64(), 3)
	min := randInt(0, 10)
	max := randInt(min.Int64()+1, 100)
	return exported.NewFeeInfo(chain, asset, rate, min, max)
}

func randInt(min int64, max int64) sdk.Int {
	return sdk.NewInt(rand.I64Between(int64(min), int64(max)))
}

func assertChainStatesEqual(t *testing.T, expected, actual *types.GenesisState) {
	assert.Equal(t, expected.Params, actual.Params)
	assert.Equal(t, expected.Nonce, actual.Nonce)
	assert.ElementsMatch(t, expected.Chains, actual.Chains)
	assert.ElementsMatch(t, expected.ChainStates, actual.ChainStates)
	assert.ElementsMatch(t, expected.LinkedAddresses, actual.LinkedAddresses)
	assert.ElementsMatch(t, expected.Transfers, actual.Transfers)
	assert.Equal(t, expected.Fee, actual.Fee)
	assert.ElementsMatch(t, expected.FeeInfos, actual.FeeInfos)
}

func TestExportGenesisInitGenesis(t *testing.T) {
	ctx, keeper := setup()

	getter := func(sdk.Context, string) (axelarnetTypes.CosmosChain, bool) {
		return axelarnetTypes.CosmosChain{Name: axelarnet.Axelarnet.Name, AddrPrefix: "axelar"}, true
	}

	keeper.InitGenesis(ctx, types.DefaultGenesisState())

	router := types.NewRouter()
	router.AddAddressValidator(evmTypes.ModuleName, evmkeeper.NewAddressValidator()).
		AddAddressValidator(axelarnetTypes.ModuleName, axelarnetkeeper.NewAddressValidator(getter))
	keeper.SetRouter(router)

	expected := types.DefaultGenesisState()

	if err := keeper.RegisterFee(ctx, axelarnet.Axelarnet, randFee(axelarnet.Axelarnet.Name, axelarnet.NativeAsset)); err != nil {
		panic(err)
	}

	keeper.SetChain(ctx, bitcoin.Bitcoin)
	if err := keeper.RegisterAsset(ctx, bitcoin.Bitcoin, exported.NewAsset(bitcoin.NativeAsset, true)); err != nil {
		panic(err)
	}
	if err := keeper.RegisterFee(ctx, bitcoin.Bitcoin, randFee(bitcoin.Bitcoin.Name, bitcoin.NativeAsset)); err != nil {
		panic(err)
	}

	if err := keeper.RegisterAsset(ctx, evm.Ethereum, exported.NewAsset(axelarnet.NativeAsset, false)); err != nil {
		panic(err)
	}
	if err := keeper.RegisterFee(ctx, evm.Ethereum, randFee(evm.Ethereum.Name, axelarnet.NativeAsset)); err != nil {
		panic(err)
	}

	expected.Chains = append(expected.Chains, bitcoin.Bitcoin)
	for _, chain := range expected.Chains {
		keeper.ActivateChain(ctx, chain)
	}

	linkedAddressesCount := rand.I64Between(100, 1000)
	expectedLinkedAddresses := make([]types.LinkedAddresses, linkedAddressesCount)
	for i := 0; i < int(linkedAddressesCount); i++ {
		depositAddress := getRandomAxelarnetAddress()
		recipientAddress := getRandomEthereumAddress()

		if err := keeper.LinkAddresses(ctx, depositAddress, recipientAddress); err != nil {
			panic(err)
		}
		expectedLinkedAddresses[i] = types.NewLinkedAddresses(depositAddress, recipientAddress)
	}
	expected.LinkedAddresses = expectedLinkedAddresses

	expected.Nonce = uint64(linkedAddressesCount)
	for i, linkedAddress := range expectedLinkedAddresses {
		depositAddress := linkedAddress.DepositAddress
		recipientAddress := linkedAddress.RecipientAddress

		_, minFee, maxFee, err := keeper.getCrossChainFees(ctx, depositAddress.Chain, recipientAddress.Chain, axelarnet.NativeAsset)
		assert.Nil(t, err)

		asset := sdk.NewCoin(axelarnet.NativeAsset, randInt(minFee.Int64()/2, maxFee.Int64()*2))
		fees, err := keeper.ComputeTransferFee(ctx, depositAddress.Chain, recipientAddress.Chain, asset)
		assert.Nil(t, err)

		_, err = keeper.EnqueueForTransfer(
			ctx,
			depositAddress,
			asset,
		)
		if err != nil {
			panic(err)
		}

		if asset.Amount.LTE(fees.Amount) {
			expectedTransfer := exported.NewCrossChainTransfer(uint64(i), recipientAddress, asset, exported.InsufficientAmount)
			expected.Transfers = append(expected.Transfers, expectedTransfer)
			continue
		}

		expectedTransfer := exported.NewPendingCrossChainTransfer(uint64(i), recipientAddress, asset.Sub(fees))
		if rand.Bools(0.5).Next() {
			keeper.ArchivePendingTransfer(ctx, expectedTransfer)
			expectedTransfer.State = exported.Archived
		}

		expected.Transfers = append(expected.Transfers, expectedTransfer)

		expected.Fee.Coins = expected.Fee.Coins.Add(fees)
	}

	expected.ChainStates = []types.ChainState{
		{
			Chain:     axelarnet.Axelarnet,
			Assets:    []exported.Asset{exported.NewAsset(axelarnet.NativeAsset, true)},
			Activated: true,
		},
		{
			Chain:     evm.Ethereum,
			Assets:    []exported.Asset{exported.NewAsset(axelarnet.NativeAsset, false)},
			Activated: true,
		},
		{
			Chain:     bitcoin.Bitcoin,
			Assets:    []exported.Asset{exported.NewAsset(bitcoin.NativeAsset, true)},
			Activated: true,
		},
	}

	for _, chainState := range expected.ChainStates {
		for _, asset := range chainState.Assets {
			feeInfo, found := keeper.GetFeeInfo(ctx, chainState.Chain, asset.Denom)
			if found {
				expected.FeeInfos = append(expected.FeeInfos, feeInfo)
			}
		}
	}

	actual := keeper.ExportGenesis(ctx)

	assert.NoError(t, actual.Validate())
	assertChainStatesEqual(t, expected, actual)

	ctx, keeper = setup()
	keeper.InitGenesis(ctx, expected)
	actual = keeper.ExportGenesis(ctx)

	assert.NoError(t, actual.Validate())
	assertChainStatesEqual(t, expected, actual)
}
