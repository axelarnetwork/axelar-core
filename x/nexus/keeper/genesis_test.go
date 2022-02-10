package keeper

import (
	"testing"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

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

func assertChainStatesEqual(t *testing.T, expected, actual *types.GenesisState) {
	assert.Equal(t, expected.Params, actual.Params)
	assert.Equal(t, expected.Nonce, actual.Nonce)
	assert.ElementsMatch(t, expected.Chains, actual.Chains)
	assert.ElementsMatch(t, expected.ChainStates, actual.ChainStates)
	assert.ElementsMatch(t, expected.LinkedAddresses, actual.LinkedAddresses)
	assert.ElementsMatch(t, expected.Transfers, actual.Transfers)
}

func TestExportGenesisInitGenesis(t *testing.T) {
	ctx, keeper := setup()

	getter := func (sdk.Context, string) (axelarnetTypes.CosmosChain, bool) {
		return axelarnetTypes.CosmosChain{Name: axelarnet.Axelarnet.Name, AddrPrefix: "axelar"}, true
	}

	keeper.InitGenesis(ctx, types.DefaultGenesisState())

	router := types.NewRouter()
	router.AddAddressValidator(evmTypes.ModuleName, evmkeeper.NewAddressValidator()).
		AddAddressValidator(axelarnetTypes.ModuleName, axelarnetkeeper.NewAddressValidator(getter))
	keeper.SetRouter(router)

	expected := types.DefaultGenesisState()

	keeper.SetChain(ctx, bitcoin.Bitcoin)
	keeper.RegisterAsset(ctx, bitcoin.Bitcoin, exported.NewAsset(bitcoin.NativeAsset, sdk.NewInt(1000000), true))
	keeper.RegisterAsset(ctx, evm.Ethereum, exported.NewAsset(axelarnet.NativeAsset, sdk.NewInt(1000000), false))

	expected.Chains = append(expected.Chains, bitcoin.Bitcoin)
	for _, chain := range expected.Chains {
		keeper.ActivateChain(ctx, chain)
	}

	linkedAddressesCount := rand.I64Between(100, 1000)
	expectedLinkedAddresses := make([]types.LinkedAddresses, linkedAddressesCount)
	for i := 0; i < int(linkedAddressesCount); i++ {
		depositAddress := getRandomAxelarnetAddress()
		recipientAddress := getRandomEthereumAddress()

		keeper.LinkAddresses(ctx, depositAddress, recipientAddress)
		expectedLinkedAddresses[i] = types.NewLinkedAddresses(depositAddress, recipientAddress)
	}
	expected.LinkedAddresses = expectedLinkedAddresses

	expected.Nonce = uint64(linkedAddressesCount)
	for i, linkedAddress := range expectedLinkedAddresses {
		depositAddress := linkedAddress.DepositAddress
		recipientAddress := linkedAddress.RecipientAddress
		asset := sdk.NewCoin(axelarnet.NativeAsset, sdk.NewInt(rand.PosI64()))

		keeper.EnqueueForTransfer(
			ctx,
			depositAddress,
			asset,
			sdk.ZeroDec(),
		)

		expectedTransfer := exported.NewPendingCrossChainTransfer(uint64(i), recipientAddress, asset)
		if rand.Bools(0.5).Next() {
			keeper.ArchivePendingTransfer(ctx, expectedTransfer)
			expectedTransfer.State = exported.Archived
		}

		expected.Transfers = append(expected.Transfers, expectedTransfer)
	}

	expected.ChainStates = []types.ChainState{
		{
			Chain:        axelarnet.Axelarnet,
			Assets:       []exported.Asset{exported.NewAsset(axelarnet.NativeAsset, sdk.NewInt(100000), true)},
			Activated:    true,
		},
		{
			Chain:     evm.Ethereum,
			Assets:    []exported.Asset{exported.NewAsset(axelarnet.NativeAsset, sdk.NewInt(1000000), false)},
			Activated: true,
		},
		{
			Chain:        bitcoin.Bitcoin,
			Assets:       []exported.Asset{exported.NewAsset(bitcoin.NativeAsset, sdk.NewInt(1000000), true)},
			Activated:    true,
		},
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
