package testutils

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported/mock"
	multisigTestutils "github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/multisig/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	votetestutils "github.com/axelarnetwork/axelar-core/x/vote/exported/testutils"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// RandomChains returns a random (valid) slice of chains for testing
func RandomChains(cdc codec.Codec) []types.GenesisState_Chain {
	chainCount := rand.I64Between(0, 20)
	var chains []types.GenesisState_Chain

	for i := int64(0); i < chainCount; i++ {
		chains = append(chains, RandomChain(cdc))
	}
	return chains
}

// RandomChain returns a random (valid) chain for testing
func RandomChain(cdc codec.Codec) types.GenesisState_Chain {
	eventCount := rand.I64Between(1, 100)
	events := make([]types.Event, eventCount)
	for i := 0; i < int(eventCount); i++ {
		events[i] = RandomGatewayEvent(types.EventConfirmed, types.EventCompleted, types.EventFailed)
	}

	chain := types.GenesisState_Chain{
		Params:              RandomParams(),
		Gateway:             RandomGateway(),
		CommandQueue:        RandomCommandQueue(cdc),
		CommandBatches:      RandomBatches(),
		Events:              events,
		ConfirmedEventQueue: getConfirmedEventQueue(cdc, events),
	}

	chain.Tokens = RandomTokens()

	confirmedTokens := getConfirmedTokens(chain.Tokens)
	if len(confirmedTokens) == 0 {
		return chain
	}

	chain.LegacyConfirmedDeposits = RandomDeposits()
	chain.LegacyBurnedDeposits = RandomDeposits()
	chain.ConfirmedDeposits = RandomDeposits()
	chain.BurnedDeposits = RandomDeposits()

	chain.BurnerInfos = RandomBurnerInfos(len(chain.LegacyConfirmedDeposits) + len(chain.LegacyBurnedDeposits) + len(chain.ConfirmedDeposits) + len(chain.BurnedDeposits))

	correctDepositsAndBurnerInfos(confirmedTokens, chain.LegacyConfirmedDeposits, chain.BurnerInfos)
	correctDepositsAndBurnerInfos(confirmedTokens, chain.LegacyBurnedDeposits, chain.BurnerInfos[len(chain.LegacyConfirmedDeposits):])
	correctDepositsAndBurnerInfos(confirmedTokens, chain.ConfirmedDeposits, chain.BurnerInfos[len(chain.LegacyConfirmedDeposits)+len(chain.LegacyBurnedDeposits):])
	correctDepositsAndBurnerInfos(confirmedTokens, chain.BurnedDeposits, chain.BurnerInfos[len(chain.LegacyConfirmedDeposits)+len(chain.LegacyBurnedDeposits)+len(chain.ConfirmedDeposits):])

	return chain
}

func correctDepositsAndBurnerInfos(confirmedTokens []types.ERC20TokenMetadata, deposits []types.ERC20Deposit, burnerInfos []types.BurnerInfo) {
	for i := range deposits {
		token := confirmedTokens[rand.I64Between(0, int64(len(confirmedTokens)))]

		deposits[i].Asset = token.Asset

		burnerInfos[i].TokenAddress = token.TokenAddress
		burnerInfos[i].Asset = token.Asset
		burnerInfos[i].Symbol = token.Details.Symbol

		burnerInfos[i].BurnerAddress = deposits[i].BurnerAddress
		burnerInfos[i].DestinationChain = deposits[i].DestinationChain
	}
}

func getConfirmedTokens(tokens []types.ERC20TokenMetadata) []types.ERC20TokenMetadata {
	var confirmedTokens []types.ERC20TokenMetadata
	for _, token := range tokens {
		if token.Status == types.Confirmed {
			confirmedTokens = append(confirmedTokens, token)
		}
	}
	return confirmedTokens
}

// RandomCommandQueue returns a random (valid) command queue state for testing
func RandomCommandQueue(cdc codec.Codec) utils.QueueState {
	qs := utils.QueueState{Items: make(map[string]utils.QueueState_Item)}
	queueName := "cmd_queue"
	queueLen := rand.I64Between(0, 20)
	commandPrefix := utils.KeyFromStr("command")

	for i := int64(0); i < queueLen; i++ {
		command := RandomCommand()

		qs.Items[fmt.Sprintf("%s_%d_%s", queueName, rand.PosI64(), command.ID.Hex())] = utils.QueueState_Item{
			Key:   commandPrefix.AppendStr(command.ID.Hex()).AsKey(),
			Value: cdc.MustMarshalLengthPrefixed(&command),
		}
	}

	return qs
}

func getConfirmedEventQueue(cdc codec.Codec, events []types.Event) utils.QueueState {
	qs := utils.QueueState{Items: make(map[string]utils.QueueState_Item)}
	queueName := "confirmed_event_queue"
	eventPrefix := utils.KeyFromStr("event")

	for _, event := range events {
		if event.Status != types.EventConfirmed {
			continue
		}

		qs.Items[fmt.Sprintf("%s_%s", queueName, event.GetID())] = utils.QueueState_Item{
			Key:   eventPrefix.AppendStr(string(event.GetID())).AsKey(),
			Value: cdc.MustMarshalLengthPrefixed(&event),
		}
	}

	return qs
}

// RandomNetworks returns a random (valid) slice of networks for testing
func RandomNetworks() []types.NetworkInfo {
	networkCount := rand.I64Between(1, 5)
	var networks []types.NetworkInfo

	for i := int64(0); i < networkCount; i++ {
		networks = append(networks, RandomNetwork())
	}
	return networks
}

// RandomNetwork returns a random (valid) network for testing
func RandomNetwork() types.NetworkInfo {
	return types.NetworkInfo{
		Name: randomNormalizedStr(5, 20),
		Id:   sdk.NewInt(rand.PosI64()),
	}
}

// RandomDeposits returns a random (valid) slice of deposits for testing
func RandomDeposits() []types.ERC20Deposit {
	depositCount := rand.I64Between(0, 20)
	var deposits []types.ERC20Deposit

	for i := int64(0); i < depositCount; i++ {
		deposits = append(deposits, RandomDeposit())
	}
	return deposits
}

// RandomDeposit returns a random (valid) deposit for testing
func RandomDeposit() types.ERC20Deposit {
	return types.ERC20Deposit{
		TxID:             RandomHash(),
		LogIndex:         uint64(rand.I64Between(0, 100)),
		Amount:           sdk.NewUint(uint64(rand.PosI64())),
		Asset:            rand.Denom(5, 10),
		DestinationChain: nexus.ChainName(randomNormalizedStr(5, 20)),
		BurnerAddress:    RandomAddress(),
	}
}

// RandomCommand returns a random (valid) command for testing
func RandomCommand() types.Command {
	commandType := rand.Of(
		slices.Filter(
			slices.TryCast[int32, types.CommandType](
				maps.Keys(types.CommandType_name)),
			func(t types.CommandType) bool { return t != types.COMMAND_TYPE_UNSPECIFIED })...)
	chainID := sdk.NewIntFromUint64(uint64(rand.PosI64()))
	asset := rand.Denom(5, 10)
	amount := big.NewInt(rand.PosI64())

	switch commandType {
	case types.COMMAND_TYPE_APPROVE_CONTRACT_CALL:
		return types.NewApproveContractCallCommand(chainID, multisigTestutils.KeyID(), nexustestutils.RandomChainName(), RandomHash(), uint64(rand.PosI64()), RandomEventContractCall())
	case types.COMMAND_TYPE_APPROVE_CONTRACT_CALL_WITH_MINT:
		return types.NewApproveContractCallWithMintCommand(chainID, multisigTestutils.KeyID(), nexustestutils.RandomChainName(), RandomHash(), uint64(rand.PosI64()), RandomEventContractCallWithToken(), sdk.NewUint(uint64(rand.PosI64())), asset)
	case types.COMMAND_TYPE_DEPLOY_TOKEN:
		return types.NewDeployTokenCommand(chainID, multisigTestutils.KeyID(), asset, RandomTokenDetails(), RandomAddress(), sdk.NewUint(uint64(rand.PosI64())))
	case types.COMMAND_TYPE_BURN_TOKEN:
		return types.NewBurnTokenCommand(chainID, multisigTestutils.KeyID(), rand.PosI64(), RandomBurnerInfo(), false)
	case types.COMMAND_TYPE_MINT_TOKEN:
		return types.NewMintTokenCommand(multisigTestutils.KeyID(), nexustestutils.RandomTransferID(), asset, common.Address(RandomAddress()), amount)
	case types.COMMAND_TYPE_TRANSFER_OPERATORSHIP:
		key := &mock.KeyMock{
			GetPubKeyFunc: func(valAddress sdk.ValAddress) (exported.PublicKey, bool) {
				return funcs.Must(btcec.NewPrivateKey()).PubKey().SerializeCompressed(), true
			},
			GetParticipantsFunc: func() []sdk.ValAddress {
				return slices.Expand2(rand.ValAddr, int(rand.I64Between(1, 10)))
			},
			GetWeightFunc:           func(sdk.ValAddress) sdk.Uint { return sdk.NewUint(uint64(rand.PosI64())) },
			GetMinPassingWeightFunc: func() sdk.Uint { return sdk.NewUint(uint64(rand.PosI64())) },
		}
		return types.NewMultisigTransferCommand(chainID, multisigTestutils.KeyID(), key)
	default:
		panic(fmt.Sprintf("unknown command type %s", commandType.String()))
	}
}

// RandomEventTransfer returns a random (valid) types.EventTransfer
func RandomEventTransfer() types.EventTransfer {
	return types.EventTransfer{
		To:     RandomAddress(),
		Amount: rand.UintBetween(sdk.OneUint(), sdk.NewUint(100000)),
	}
}

// RandomEventContractCallWithToken returns a random (valid) types.EventContractCallWithToken
func RandomEventContractCallWithToken() types.EventContractCallWithToken {
	return types.EventContractCallWithToken{
		Sender:           RandomAddress(),
		DestinationChain: nexustestutils.RandomChainName(),
		ContractAddress:  RandomAddress().Hex(),
		PayloadHash:      RandomHash(),
		Symbol:           rand.Denom(3, 5),
		Amount:           sdk.NewUint(uint64(rand.PosI64())),
	}
}

// RandomEventContractCall returns a random (valid) types.EventContractCall
func RandomEventContractCall() types.EventContractCall {
	return types.EventContractCall{
		Sender:           RandomAddress(),
		DestinationChain: nexustestutils.RandomChainName(),
		ContractAddress:  RandomAddress().Hex(),
		PayloadHash:      RandomHash(),
	}
}

// RandomBatches returns a random (valid) slice of command batches for testing
func RandomBatches() []types.CommandBatchMetadata {
	batchCount := rand.I64Between(0, 20)
	var batches []types.CommandBatchMetadata

	var prevBatch types.CommandBatchMetadata
	for i := int64(0); i < batchCount; i++ {
		batch := RandomBatch()
		if i < batchCount-1 {
			batch.Status = types.BatchSigned
			sig := testutils.MultiSig()
			batch.Signature = funcs.Must(codectypes.NewAnyWithValue(&sig))
		}
		batch.PrevBatchedCommandsID = prevBatch.ID

		batches = append(batches, batch)

		prevBatch = batch
	}

	return batches
}

// RandomBatch returns a random (valid) command batch for testing
func RandomBatch() types.CommandBatchMetadata {
	md := types.CommandBatchMetadata{
		ID:                    rand.Bytes(32),
		CommandIDs:            RandomCommandIDs(),
		Data:                  rand.Bytes(int(rand.I64Between(1, 1000))),
		SigHash:               RandomHash(),
		Status:                types.BatchedCommandsStatus(rand.I64Between(1, int64(len(types.BatchedCommandsStatus_name)))),
		KeyID:                 multisigTestutils.KeyID(),
		PrevBatchedCommandsID: rand.Bytes(32),
	}

	if md.Status == types.BatchSigned {
		sig := testutils.MultiSig()
		md.Signature = funcs.Must(codectypes.NewAnyWithValue(&sig))
	}
	return md
}

// RandomCommandIDs returns a random (valid) slice of command IDs for testing
func RandomCommandIDs() []types.CommandID {
	idCount := rand.I64Between(1, 20)
	var ids []types.CommandID

	for i := int64(0); i < idCount; i++ {
		ids = append(ids, RandomCommandID())
	}
	return ids
}

// RandomCommandID returns a random (valid) command ID for testing
func RandomCommandID() types.CommandID {
	return types.NewCommandID(rand.Bytes(int(rand.I64Between(1, 100))), sdk.NewInt(rand.PosI64()))
}

// RandomTokens returns a random (valid) slice of tokens for testing
func RandomTokens() []types.ERC20TokenMetadata {
	tokenCount := rand.I64Between(0, 20)
	var tokens []types.ERC20TokenMetadata

	for i := int64(0); i < tokenCount; i++ {
		tokens = append(tokens, RandomToken())
	}
	return tokens
}

// RandomToken returns a random (valid) token for testing
func RandomToken() types.ERC20TokenMetadata {
	bzBurnable, err := utils.HexDecode(types.Burnable)
	if err != nil {
		panic(err)
	}

	md := types.ERC20TokenMetadata{
		Asset:        rand.Denom(5, 20),
		ChainID:      sdk.NewInt(rand.PosI64()),
		Details:      RandomTokenDetails(),
		TokenAddress: RandomAddress(),
		TxHash:       RandomHash(),
		Status:       1 << rand.I64Between(0, int64(len(types.Status_name))),
		IsExternal:   rand.Bools(0.5).Next(),
		BurnerCode:   bzBurnable,
	}

	if md.IsExternal {
		md.BurnerCode = nil
	}
	return md
}

// RandomTokenDetails returns a random (valid) token details instance for testing
func RandomTokenDetails() types.TokenDetails {
	return types.TokenDetails{
		TokenName: randomNormalizedStr(5, 20),
		Symbol:    randomNormalizedStr(5, 20),
		Decimals:  uint8(rand.I64Between(1, 20)),
		Capacity:  sdk.NewInt(rand.PosI64()),
	}
}

// RandomGateway returns a random (valid) gateway for testing
func RandomGateway() types.Gateway {
	return types.Gateway{
		Address: RandomAddress(),
	}
}

// RandomBurnerInfos returns a random (valid) slice of burner infos for testing
func RandomBurnerInfos(count ...int) []types.BurnerInfo {
	burnerCount := rand.I64Between(0, 20)
	if len(count) == 1 {
		burnerCount = int64(count[0])
	}
	var infos []types.BurnerInfo

	for i := int64(0); i < burnerCount; i++ {
		infos = append(infos, RandomBurnerInfo())
	}
	return infos
}

// RandomBurnerInfo returns a random (valid) burner info instance for testing
func RandomBurnerInfo() types.BurnerInfo {
	return types.BurnerInfo{
		BurnerAddress:    RandomAddress(),
		TokenAddress:     RandomAddress(),
		DestinationChain: nexus.ChainName(randomNormalizedStr(5, 20)),
		Symbol:           randomNormalizedStr(5, 20),
		Asset:            rand.Denom(5, 20),
		Salt:             RandomHash(),
	}
}

// RandomParams returns a random (valid) params instance for testing
func RandomParams() types.Params {
	bzBurnable, err := utils.HexDecode(types.Burnable)
	if err != nil {
		panic(err)
	}

	nominator := rand.I64Between(1, 100)
	denominator := rand.I64Between(nominator, 101)
	params := types.Params{
		Chain:               nexus.ChainName(randomNormalizedStr(5, 20)),
		ConfirmationHeight:  uint64(rand.PosI64()),
		TokenCode:           rand.Bytes(int(rand.I64Between(10, 100))),
		Burnable:            bzBurnable,
		RevoteLockingPeriod: rand.PosI64(),
		Networks:            RandomNetworks(),
		VotingThreshold:     utils.NewThreshold(nominator, denominator),
		MinVoterCount:       rand.PosI64(),
		CommandsGasLimit:    uint32(rand.I64Between(0, 10000000)),
		EndBlockerLimit:     rand.PosI64(),
		TransferLimit:       uint64(rand.PosI64()),
	}

	params.Network = params.Networks[int(rand.I64Between(0, int64(len(params.Networks))))].Name

	return params
}

// RandomAddress returns a random (valid) address for testing
func RandomAddress() types.Address {
	return types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength)))
}

// RandomHash returns a random (valid) hash for testing
func RandomHash() types.Hash {
	return types.Hash(common.BytesToHash(rand.Bytes(common.HashLength)))
}

func randomNormalizedStr(min, max int) string {
	return strings.ReplaceAll(utils.NormalizeString(rand.StrBetween(min, max)), utils.DefaultDelimiter, "-")
}

// RandomGatewayEvent returns a random event for testing
func RandomGatewayEvent(statuses ...types.Event_Status) types.Event {
	if len(statuses) == 0 {
		statuses = []types.Event_Status{types.EventConfirmed, types.EventCompleted, types.EventFailed}
	}

	payload := rand.Bytes(int(rand.I64Between(1, 100)))

	return rand.Of(
		types.Event{
			Chain:  nexus.ChainName(randomNormalizedStr(5, 20)),
			TxID:   RandomHash(),
			Index:  uint64(rand.PosI64()),
			Status: rand.Of(statuses...),
			Event: &types.Event_ContractCall{
				ContractCall: &types.EventContractCall{
					Sender:           RandomAddress(),
					DestinationChain: nexus.ChainName(randomNormalizedStr(5, 20)),
					ContractAddress:  RandomAddress().Hex(),
					PayloadHash:      types.Hash(crypto.Keccak256Hash(payload)),
				},
			},
		},
		types.Event{
			Chain:  nexus.ChainName(randomNormalizedStr(5, 20)),
			TxID:   RandomHash(),
			Index:  uint64(rand.PosI64()),
			Status: rand.Of(statuses...),
			Event: &types.Event_ContractCallWithToken{
				ContractCallWithToken: &types.EventContractCallWithToken{
					Sender:           RandomAddress(),
					DestinationChain: nexus.ChainName(randomNormalizedStr(5, 20)),
					ContractAddress:  RandomAddress().Hex(),
					PayloadHash:      types.Hash(crypto.Keccak256Hash(payload)),
					Symbol:           randomNormalizedStr(5, 20),
					Amount:           sdk.NewUint(uint64(rand.PosI64())),
				},
			},
		},
		types.Event{
			Chain:  nexus.ChainName(randomNormalizedStr(5, 20)),
			TxID:   RandomHash(),
			Index:  uint64(rand.PosI64()),
			Status: rand.Of(statuses...),
			Event: &types.Event_TokenSent{
				TokenSent: &types.EventTokenSent{
					Sender:             RandomAddress(),
					DestinationChain:   nexus.ChainName(randomNormalizedStr(5, 20)),
					DestinationAddress: RandomAddress().Hex(),
					Symbol:             randomNormalizedStr(5, 20),
					Amount:             sdk.NewUint(uint64(rand.PosI64())),
				},
			},
		},
	)
}

// RandomConfirmDepositStarted generates a random ConfirmDepositStarted event struct
func RandomConfirmDepositStarted() types.ConfirmDepositStarted {
	return types.ConfirmDepositStarted{
		TxID:               RandomHash(),
		Chain:              nexustestutils.RandomChainName(),
		DepositAddress:     RandomAddress(),
		TokenAddress:       RandomAddress(),
		ConfirmationHeight: uint64(rand.PosI64()),
		PollParticipants:   votetestutils.RandomPollParticipants(),
	}
}
