package keeper

import (
	"github.com/axelarnetwork/axelar-core/utils"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	axelarnetTypes "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmUtil "github.com/ethereum/go-ethereum/common"
	mathrand "math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

var chains = []exported.Chain{
	evm.Ethereum,
	{Name: "avalanche", Module: evmTypes.ModuleName, SupportsForeignAssets: true},
	{Name: "fantom", Module: evmTypes.ModuleName, SupportsForeignAssets: true},
	{Name: "moonbeam", Module: evmTypes.ModuleName, SupportsForeignAssets: true},
	{Name: "polygon", Module: evmTypes.ModuleName, SupportsForeignAssets: true},
	axelarnet.Axelarnet,
	{Name: "terra", Module: axelarnetTypes.ModuleName, SupportsForeignAssets: true},
}

func TestGetMigrationHandler_migrateLinkedAddressesKey(t *testing.T) {
	ctx, keeper := setup()

	// generate random linked addresses
	linkedAddresses := make([]types.LinkedAddresses, rand.I64Between(200, 1000))
	for i := 0; i < len(linkedAddresses); i++ {

		depositAddr, recipientAddr := makeRandAddressesForChain(chains[mathrand.Intn(len(chains))], chains[mathrand.Intn(len(chains))])
		linkedAddresses[i] = types.LinkedAddresses{
			DepositAddress:   depositAddr,
			RecipientAddress: recipientAddr,
		}
		// set old linked addresses
		keeper.getStore(ctx).Set(linkedAddressesPrefix.Append(utils.LowerCaseKey(linkedAddresses[i].DepositAddress.String())), &linkedAddresses[i])
	}

	// should get linked addresses by old key
	for i := 0; i < len(linkedAddresses); i++ {
		var actualLinkAddresses types.LinkedAddresses
		keeper.getStore(ctx).Get(linkedAddressesPrefix.Append(utils.LowerCaseKey(linkedAddresses[i].DepositAddress.String())), &actualLinkAddresses)
		assert.Equal(t, linkedAddresses[i], actualLinkAddresses)
	}

	// run migration
	handler := GetMigrationHandler(keeper)
	err := handler(ctx)
	assert.NoError(t, err)

	// should migrate to new linked addresses key
	for i := 0; i < len(linkedAddresses); i++ {
		var actualLinkAddresses types.LinkedAddresses
		actualLinkAddresses, ok := keeper.getLinkedAddresses(ctx, linkedAddresses[i].DepositAddress)
		assert.True(t, ok)
		assert.Equal(t, linkedAddresses[i], actualLinkAddresses)
	}
}

func TestGetMigrationHandler_deleteLatestDepositKey(t *testing.T) {
	ctx, keeper := setup()

	// generate random linked addresses
	linkedAddresses := make([]types.LinkedAddresses, rand.I64Between(200, 1000))
	for i := 0; i < len(linkedAddresses); i++ {

		depositAddr, recipientAddr := makeRandAddressesForChain(chains[mathrand.Intn(len(chains))], chains[mathrand.Intn(len(chains))])
		linkedAddresses[i] = types.LinkedAddresses{
			DepositAddress:   depositAddr,
			RecipientAddress: recipientAddr,
		}
		// set old latest deposit address
		keeper.getStore(ctx).Set(latestDepositAddressPrefix.
			AppendStr(linkedAddresses[i].DepositAddress.Chain.Name).
			Append(utils.LowerCaseKey(linkedAddresses[i].RecipientAddress.String())), &linkedAddresses[i].DepositAddress)
	}

	// should get latest deposit address by old key
	for i := 0; i < len(linkedAddresses); i++ {
		var actualDepositAddress exported.CrossChainAddress
		ok := keeper.getStore(ctx).Get(latestDepositAddressPrefix.
			AppendStr(linkedAddresses[i].DepositAddress.Chain.Name).
			Append(utils.LowerCaseKey(linkedAddresses[i].RecipientAddress.String())), &actualDepositAddress)
		assert.True(t, ok)
		assert.Equal(t, linkedAddresses[i].DepositAddress, actualDepositAddress)
	}

	// run migration
	handler := GetMigrationHandler(keeper)
	err := handler(ctx)
	assert.NoError(t, err)

	// should not get latest deposit address by old key
	for i := 0; i < len(linkedAddresses); i++ {
		var actualDepositAddress exported.CrossChainAddress
		ok := keeper.getStore(ctx).Get(latestDepositAddressPrefix.AppendStr(linkedAddresses[i].DepositAddress.Chain.Name).Append(utils.LowerCaseKey(linkedAddresses[i].RecipientAddress.String())), &actualDepositAddress)
		assert.False(t, ok)
	}
}

func makeRandAddressesForChain(origin, destination exported.Chain) (exported.CrossChainAddress, exported.CrossChainAddress) {
	var addr string

	switch origin.Module {
	case evmTypes.ModuleName:
		addr = genEvmAddr()
	case axelarnetTypes.ModuleName:
		addr = genCosmosAddr(origin.Name)
	default:
		panic("unexpected module for origin")
	}

	sender := exported.CrossChainAddress{
		Address: addr,
		Chain:   origin,
	}

	switch destination.Module {
	case evmTypes.ModuleName:
		addr = genEvmAddr()
	case axelarnetTypes.ModuleName:
		addr = genCosmosAddr(destination.Name)
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

func genCosmosAddr(chain string) string {
	prefix := ""
	switch strings.ToLower(chain) {
	case "axelarnet":
		prefix = "axelar"
	case "terra":
		prefix = "terra"
	default:
		prefix = ""
	}

	sdk.GetConfig().SetBech32PrefixForAccount(prefix, prefix)
	return rand.AccAddr().String()
}
