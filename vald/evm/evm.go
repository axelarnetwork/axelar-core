package evm

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strings"

	sdkClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	tmLog "github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/sdk-utils/broadcast"
	"github.com/axelarnetwork/axelar-core/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// Smart contract event signatures
var (
	ERC20TransferSig                = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	ERC20TokenDeploymentSig         = crypto.Keccak256Hash([]byte("TokenDeployed(string,address)"))
	MultisigTransferOperatorshipSig = crypto.Keccak256Hash([]byte("OperatorshipTransferred(bytes)"))
	ContractCallSig                 = crypto.Keccak256Hash([]byte("ContractCall(address,string,string,bytes32,bytes)"))
	ContractCallWithTokenSig        = crypto.Keccak256Hash([]byte("ContractCallWithToken(address,string,string,bytes32,bytes,string,uint256)"))
	TokenSentSig                    = crypto.Keccak256Hash([]byte("TokenSent(address,string,string,string,uint256)"))
)

// Mgr manages all communication with Ethereum
type Mgr struct {
	cliCtx      sdkClient.Context
	logger      tmLog.Logger
	rpcs        map[string]rpc.Client
	broadcaster broadcast.Broadcaster
	cdc         *codec.LegacyAmino
	validator   sdk.ValAddress
}

// NewMgr returns a new Mgr instance
func NewMgr(rpcs map[string]rpc.Client, cliCtx sdkClient.Context, broadcaster broadcast.Broadcaster, logger tmLog.Logger, cdc *codec.LegacyAmino, valAddr sdk.ValAddress) *Mgr {
	return &Mgr{
		rpcs:        rpcs,
		cliCtx:      cliCtx,
		broadcaster: broadcaster,
		logger:      logger.With("listener", "evm"),
		cdc:         cdc,
		validator:   valAddr,
	}
}

// ProcessNewChain notifies the operator that vald needs to be restarted/udpated for a new chain
func (mgr Mgr) ProcessNewChain(event *types.ChainAdded) (err error) {
	mgr.logger.Info(fmt.Sprintf("VALD needs to be updated and restarted for new chain %s", event.Chain.String()))
	return nil
}

// ProcessDepositConfirmation votes on the correctness of an EVM chain token deposit
func (mgr Mgr) ProcessDepositConfirmation(event *evmTypes.ConfirmDepositStarted) error {
	if !slices.Any(event.Participants, func(v sdk.ValAddress) bool { return v.Equals(mgr.validator) }) {
		mgr.logger.Debug("ignoring deposit confirmation poll: not a participant", "pollID", event.PollID)
		return nil
	}

	rpcClient, found := mgr.rpcs[strings.ToLower(event.Chain.String())]
	if !found {
		return fmt.Errorf("unable to find an RPC for chain '%s'", event.Chain.String())
	}
	var events []evmTypes.Event
	_ = mgr.validate(rpcClient, common.Hash(event.TxID), event.ConfirmationHeight, func(_ *geth.Transaction, txReceipt *geth.Receipt) bool {
		for i, log := range txReceipt.Logs {
			switch log.Topics[0] {
			case ERC20TransferSig:
				if !bytes.Equal(event.TokenAddress.Bytes(), log.Address.Bytes()) {
					continue
				}

				erc20Event, err := decodeERC20TransferEvent(log)
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "decode event Transfer failed").Error())
					continue
				}

				if erc20Event.To != event.DepositAddress {
					continue
				}

				if err := erc20Event.ValidateBasic(); err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "invalid event Transfer").Error())
					continue
				}

				events = append(events, evmTypes.Event{
					Chain: event.Chain,
					TxID:  event.TxID,
					Index: uint64(i),
					Event: &evmTypes.Event_Transfer{
						Transfer: &erc20Event,
					},
				})
			default:
			}

		}
		return true
	})

	msg := voteTypes.NewVoteRequest(mgr.cliCtx.FromAddress, event.PollID, evmTypes.NewVoteEvents(event.Chain, events))
	mgr.logger.Info(fmt.Sprintf("broadcasting vote %v for poll %s", events, event.PollID.String()))
	_, err := mgr.broadcaster.Broadcast(context.TODO(), msg)
	return err
}

// ProcessTokenConfirmation votes on the correctness of an EVM chain token deployment
func (mgr Mgr) ProcessTokenConfirmation(event *evmTypes.ConfirmTokenStarted) error {
	if !slices.Any(event.Participants, func(v sdk.ValAddress) bool { return v.Equals(mgr.validator) }) {
		mgr.logger.Debug("ignoring token confirmation poll: not a participant", "pollID", event.PollID)
		return nil
	}

	rpcClient, found := mgr.rpcs[strings.ToLower(event.Chain.String())]
	if !found {
		return fmt.Errorf("unable to find an RPC for chain '%s'", event.Chain)
	}

	var events []evmTypes.Event
	_ = mgr.validate(rpcClient, common.Hash(event.TxID), event.ConfirmationHeight, func(_ *geth.Transaction, txReceipt *geth.Receipt) bool {
		for i, log := range txReceipt.Logs {
			if !bytes.Equal(event.GatewayAddress.Bytes(), log.Address.Bytes()) {
				continue
			}

			switch log.Topics[0] {
			case ERC20TokenDeploymentSig:
				erc20Event, err := decodeERC20TokenDeploymentEvent(log)
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "decode event TokenDeployed failed").Error())
					continue
				}
				if erc20Event.TokenAddress != event.TokenAddress || erc20Event.Symbol != event.TokenDetails.Symbol {
					continue
				}

				if err := erc20Event.ValidateBasic(); err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "invalid event ERC20TokenDeployment").Error())
					continue
				}

				events = append(events, evmTypes.Event{
					Chain: event.Chain,
					TxID:  event.TxID,
					Index: uint64(i),
					Event: &evmTypes.Event_TokenDeployed{
						TokenDeployed: &erc20Event,
					},
				})

				return true
			default:
			}
		}

		return true
	})

	msg := voteTypes.NewVoteRequest(mgr.cliCtx.FromAddress, event.PollID, evmTypes.NewVoteEvents(event.Chain, events))
	mgr.logger.Info(fmt.Sprintf("broadcasting vote %v for poll %s", events, event.PollID.String()))
	_, err := mgr.broadcaster.Broadcast(context.TODO(), msg)
	return err
}

// ProcessTransferKeyConfirmation votes on the correctness of an EVM chain key transfer
func (mgr Mgr) ProcessTransferKeyConfirmation(event *types.ConfirmKeyTransferStarted) error {
	if !slices.Any(event.Participants, func(v sdk.ValAddress) bool { return v.Equals(mgr.validator) }) {
		mgr.logger.Debug("ignoring key transfer confirmation poll: not a participant", "pollID", event.PollID)
		return nil
	}

	rpcClient, ok := mgr.rpcs[strings.ToLower(event.Chain.String())]
	if !ok {
		return fmt.Errorf("unable to find the RPC for chain %s", event.Chain)
	}

	var operatorshipTransferred evmTypes.Event
	ok = mgr.validate(rpcClient, common.Hash(event.TxID), event.ConfirmationHeight, func(_ *geth.Transaction, txReceipt *geth.Receipt) bool {
		for i := len(txReceipt.Logs) - 1; i >= 0; i-- {
			log := txReceipt.Logs[i]

			// Event is not emitted by the axelar gateway
			if log.Address != common.Address(event.GatewayAddress) {
				continue
			}

			if log.Topics[0] != MultisigTransferOperatorshipSig {
				continue
			}

			transferOperatorshipEvent, err := decodeMultisigOperatorshipTransferredEvent(log)
			if err != nil {
				mgr.logger.Debug(sdkerrors.Wrap(err, "failed decoding operatorship transferred event").Error())
				continue
			}

			if err := transferOperatorshipEvent.ValidateBasic(); err != nil {
				mgr.logger.Debug(sdkerrors.Wrap(err, "invalid event MultisigTransferOperatorship").Error())
				continue
			}

			operatorshipTransferred = evmTypes.Event{
				Chain: event.Chain,
				TxID:  event.TxID,
				Index: uint64(i),
				Event: &evmTypes.Event_MultisigOperatorshipTransferred{
					MultisigOperatorshipTransferred: &transferOperatorshipEvent,
				}}
			return true
		}

		return false
	})

	var evmEvents []evmTypes.Event
	if ok {
		evmEvents = append(evmEvents, operatorshipTransferred)
	}

	msg := voteTypes.NewVoteRequest(mgr.cliCtx.FromAddress, event.PollID, evmTypes.NewVoteEvents(event.Chain, evmEvents))
	mgr.logger.Info(fmt.Sprintf("broadcasting vote %v for poll %s", evmEvents, event.PollID.String()))
	_, err := mgr.broadcaster.Broadcast(context.TODO(), msg)

	return err
}

// ProcessGatewayTxConfirmation votes on the correctness of an EVM chain gateway's transactions
func (mgr Mgr) ProcessGatewayTxConfirmation(event *evmTypes.ConfirmGatewayTxStarted) error {
	if !slices.Any(event.Participants, func(v sdk.ValAddress) bool { return v.Equals(mgr.validator) }) {
		mgr.logger.Debug("ignoring gateway tx confirmation poll: not a participant", "pollID", event.PollID)
		return nil
	}

	rpcClient, found := mgr.rpcs[strings.ToLower(event.Chain.String())]
	if !found {
		return fmt.Errorf("unable to find an RPC for chain '%s'", event.Chain.String())
	}

	var events []evmTypes.Event
	_ = mgr.validate(rpcClient, common.Hash(event.TxID), event.ConfirmationHeight, func(_ *geth.Transaction, txReceipt *geth.Receipt) bool {
		for i, log := range txReceipt.Logs {
			if !bytes.Equal(event.GatewayAddress.Bytes(), log.Address.Bytes()) {
				continue
			}

			switch log.Topics[0] {
			case ContractCallSig:
				gatewayEvent, err := decodeEventContractCall(log)
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "decode event ContractCall failed").Error())

					return false
				}

				err = gatewayEvent.ValidateBasic()
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "invalid event ContractCall").Error())
					continue
				}

				events = append(events, evmTypes.Event{
					Chain: event.Chain,
					TxID:  event.TxID,
					Index: uint64(i),
					Event: &evmTypes.Event_ContractCall{
						ContractCall: &gatewayEvent,
					},
				})
			case ContractCallWithTokenSig:
				gatewayEvent, err := decodeEventContractCallWithToken(log)
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "decode event ContractCallWithToken failed").Error())

					return false
				}

				err = gatewayEvent.ValidateBasic()
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "invalid event ContractCallWithToken").Error())
					continue
				}

				events = append(events, evmTypes.Event{
					Chain: event.Chain,
					TxID:  event.TxID,
					Index: uint64(i),
					Event: &evmTypes.Event_ContractCallWithToken{
						ContractCallWithToken: &gatewayEvent,
					},
				})
			case TokenSentSig:
				gatewayEvent, err := decodeEventTokenSent(log)
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "decode event TokenSent failed").Error())
				}

				err = gatewayEvent.ValidateBasic()
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "invalid event TokenSent").Error())
					continue
				}

				events = append(events, evmTypes.Event{
					Chain: event.Chain,
					TxID:  event.TxID,
					Index: uint64(i),
					Event: &evmTypes.Event_TokenSent{
						TokenSent: &gatewayEvent,
					},
				})
			default:
			}
		}

		return true
	})

	msg := voteTypes.NewVoteRequest(mgr.cliCtx.FromAddress, event.PollID, evmTypes.NewVoteEvents(event.Chain, events))
	mgr.logger.Info(fmt.Sprintf("broadcasting vote %v for poll %s", events, event.PollID.String()))
	_, err := mgr.broadcaster.Broadcast(context.TODO(), msg)
	return err
}

func decodeEventTokenSent(log *geth.Log) (evmTypes.EventTokenSent, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return evmTypes.EventTokenSent{}, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return evmTypes.EventTokenSent{}, err
	}

	arguments := abi.Arguments{
		{Type: stringType},
		{Type: stringType},
		{Type: stringType},
		{Type: uint256Type},
	}
	params, err := evmTypes.StrictDecode(arguments, log.Data)
	if err != nil {
		return evmTypes.EventTokenSent{}, err
	}

	return evmTypes.EventTokenSent{
		Sender:             evmTypes.Address(common.BytesToAddress(log.Topics[1].Bytes())),
		DestinationChain:   nexus.ChainName(params[0].(string)),
		DestinationAddress: params[1].(string),
		Symbol:             params[2].(string),
		Amount:             sdk.NewUintFromBigInt(params[3].(*big.Int)),
	}, nil
}

func decodeEventContractCall(log *geth.Log) (evmTypes.EventContractCall, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return evmTypes.EventContractCall{}, err
	}

	bytesType, err := abi.NewType("bytes", "bytes", nil)
	if err != nil {
		return evmTypes.EventContractCall{}, err
	}

	arguments := abi.Arguments{
		{Type: stringType},
		{Type: stringType},
		{Type: bytesType},
	}
	params, err := evmTypes.StrictDecode(arguments, log.Data)
	if err != nil {
		return evmTypes.EventContractCall{}, err
	}

	return evmTypes.EventContractCall{
		Sender:           evmTypes.Address(common.BytesToAddress(log.Topics[1].Bytes())),
		DestinationChain: nexus.ChainName(params[0].(string)),
		ContractAddress:  params[1].(string),
		PayloadHash:      evmTypes.Hash(common.BytesToHash(log.Topics[2].Bytes())),
	}, nil
}

func decodeEventContractCallWithToken(log *geth.Log) (evmTypes.EventContractCallWithToken, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return evmTypes.EventContractCallWithToken{}, err
	}

	bytesType, err := abi.NewType("bytes", "bytes", nil)
	if err != nil {
		return evmTypes.EventContractCallWithToken{}, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return evmTypes.EventContractCallWithToken{}, err
	}

	arguments := abi.Arguments{
		{Type: stringType},
		{Type: stringType},
		{Type: bytesType},
		{Type: stringType},
		{Type: uint256Type},
	}
	params, err := evmTypes.StrictDecode(arguments, log.Data)
	if err != nil {
		return evmTypes.EventContractCallWithToken{}, err
	}

	return evmTypes.EventContractCallWithToken{
		Sender:           evmTypes.Address(common.BytesToAddress(log.Topics[1].Bytes())),
		DestinationChain: nexus.ChainName(params[0].(string)),
		ContractAddress:  params[1].(string),
		PayloadHash:      evmTypes.Hash(common.BytesToHash(log.Topics[2].Bytes())),
		Symbol:           params[3].(string),
		Amount:           sdk.NewUintFromBigInt(params[4].(*big.Int)),
	}, nil
}

func getLatestFinalizedBlockNumber(client rpc.Client, confHeight uint64) (*big.Int, error) {
	switch client := client.(type) {
	case rpc.MoonbeamClient:
		finalizedBlockHash, err := client.ChainGetFinalizedHead(context.Background())
		if err != nil {
			return nil, err
		}

		header, err := client.ChainGetHeader(context.Background(), finalizedBlockHash)
		if err != nil {
			return nil, err
		}

		return header.Number.ToInt(), nil
	case rpc.Eth2Client:
		// TODO: check error after the merge is settled on ethereum mainnet
		finalizedHeader, _ := client.FinalizedHeader(context.Background())
		if finalizedHeader != nil {
			return finalizedHeader.Number, nil
		}
	}

	blockNumber, err := client.BlockNumber(context.Background())
	if err != nil {
		return nil, err
	}

	return big.NewInt(int64(blockNumber - confHeight + 1)), nil
}

func (mgr Mgr) validate(client rpc.Client, txID common.Hash, confHeight uint64, validateTx func(tx *geth.Transaction, txReceipt *geth.Receipt) bool) bool {
	tx, _, err := client.TransactionByHash(context.Background(), txID)
	if err != nil {
		mgr.logger.Debug(sdkerrors.Wrap(err, "get transaction by hash call failed").Error())
		return false
	}

	txReceipt, err := client.TransactionReceipt(context.Background(), txID)
	if err != nil {
		mgr.logger.Debug(sdkerrors.Wrap(err, "get transaction receipt call failed").Error())
		return false
	}

	if !isTxSuccessful(txReceipt) {
		mgr.logger.Debug(fmt.Sprintf("transaction %s failed", txReceipt.TxHash.String()))
		return false
	}

	latestFinalizedBlockNumber, err := getLatestFinalizedBlockNumber(client, confHeight)
	if err != nil {
		mgr.logger.Debug(sdkerrors.Wrap(err, "get latest finalized block number failed").Error())
		return false
	}

	if latestFinalizedBlockNumber.Cmp(txReceipt.BlockNumber) < 0 {
		mgr.logger.Debug(fmt.Sprintf("transaction %s is not finalized yet", txReceipt.TxHash.String()))
		return false
	}

	txBlock, err := client.BlockByNumber(context.Background(), txReceipt.BlockNumber)
	if err != nil {
		mgr.logger.Debug(sdkerrors.Wrap(err, "get block by number call failed").Error())
		return false
	}

	txFound := false
	for _, t := range txBlock.Body().Transactions {
		if bytes.Equal(t.Hash().Bytes(), txReceipt.TxHash.Bytes()) {
			txFound = true
			break
		}
	}

	if !txFound {
		mgr.logger.Debug(fmt.Sprintf("transaction %s is not found in block number %d and hash %s", txReceipt.TxHash.String(), txBlock.NumberU64(), txBlock.Hash().String()))
		return false
	}

	return validateTx(tx, txReceipt)
}

func isTxSuccessful(txReceipt *geth.Receipt) bool {
	return txReceipt.Status == 1
}

func decodeERC20TransferEvent(log *geth.Log) (evmTypes.EventTransfer, error) {
	if len(log.Topics) != 3 || log.Topics[0] != ERC20TransferSig {
		return evmTypes.EventTransfer{}, fmt.Errorf("log is not an ERC20 transfer")
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return evmTypes.EventTransfer{}, err
	}

	to := common.BytesToAddress(log.Topics[2][:])

	arguments := abi.Arguments{
		{Type: uint256Type},
	}

	params, err := arguments.Unpack(log.Data)
	if err != nil {
		return evmTypes.EventTransfer{}, err
	}

	return evmTypes.EventTransfer{
		To:     evmTypes.Address(to),
		Amount: sdk.NewUintFromBigInt(params[0].(*big.Int)),
	}, nil
}

func decodeERC20TokenDeploymentEvent(log *geth.Log) (evmTypes.EventTokenDeployed, error) {
	if len(log.Topics) != 1 || log.Topics[0] != ERC20TokenDeploymentSig {
		return evmTypes.EventTokenDeployed{}, fmt.Errorf("event is not for an ERC20 token deployment")
	}

	// Decode the data field
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return evmTypes.EventTokenDeployed{}, err
	}
	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return evmTypes.EventTokenDeployed{}, err
	}

	arguments := abi.Arguments{{Type: stringType}, {Type: addressType}}
	params, err := evmTypes.StrictDecode(arguments, log.Data)
	if err != nil {
		return evmTypes.EventTokenDeployed{}, err
	}

	return evmTypes.EventTokenDeployed{
		Symbol:       params[0].(string),
		TokenAddress: evmTypes.Address(params[1].(common.Address)),
	}, nil
}

func decodeMultisigOperatorshipTransferredEvent(log *geth.Log) (evmTypes.EventMultisigOperatorshipTransferred, error) {
	if len(log.Topics) != 1 || log.Topics[0] != MultisigTransferOperatorshipSig {
		return evmTypes.EventMultisigOperatorshipTransferred{}, fmt.Errorf("event is not OperatorshipTransferred")
	}

	newAddresses, newWeights, newThreshold, err := unpackMultisigTransferKeyEvent(log)
	if err != nil {
		return evmTypes.EventMultisigOperatorshipTransferred{}, err
	}

	event := evmTypes.EventMultisigOperatorshipTransferred{
		NewOperators: slices.Map(newAddresses, func(addr common.Address) evmTypes.Address { return evmTypes.Address(addr) }),
		NewWeights:   slices.Map(newWeights, sdk.NewUintFromBigInt),
		NewThreshold: sdk.NewUintFromBigInt(newThreshold),
	}

	return event, nil
}

func unpackMultisigTransferKeyEvent(log *geth.Log) ([]common.Address, []*big.Int, *big.Int, error) {
	bytesType := funcs.Must(abi.NewType("bytes", "bytes", nil))
	newOperatorsData, err := evmTypes.StrictDecode(abi.Arguments{{Type: bytesType}}, log.Data)
	if err != nil {
		return nil, nil, nil, err
	}

	addressesType := funcs.Must(abi.NewType("address[]", "address[]", nil))
	uint256ArrayType := funcs.Must(abi.NewType("uint256[]", "uint256[]", nil))
	uint256Type := funcs.Must(abi.NewType("uint256", "uint256", nil))

	arguments := abi.Arguments{{Type: addressesType}, {Type: uint256ArrayType}, {Type: uint256Type}}
	params, err := types.StrictDecode(arguments, newOperatorsData[0].([]byte))
	if err != nil {
		return nil, nil, nil, err
	}

	return params[0].([]common.Address), params[1].([]*big.Int), params[2].(*big.Int), nil
}
