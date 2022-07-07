package evm

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strconv"
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

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcast"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/parse"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	tmEvents "github.com/axelarnetwork/tm-events/events"
	"github.com/axelarnetwork/utils/slices"
)

// Smart contract event signatures
var (
	ERC20TransferSig                 = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	ERC20TokenDeploymentSig          = crypto.Keccak256Hash([]byte("TokenDeployed(string,address)"))
	SinglesigTransferOperatorshipSig = crypto.Keccak256Hash([]byte("OperatorshipTransferred(address,address)"))
	MultisigTransferOperatorshipSig  = crypto.Keccak256Hash([]byte("OperatorshipTransferred(bytes)"))
	ContractCallSig                  = crypto.Keccak256Hash([]byte("ContractCall(address,string,string,bytes32,bytes)"))
	ContractCallWithTokenSig         = crypto.Keccak256Hash([]byte("ContractCallWithToken(address,string,string,bytes32,bytes,string,uint256)"))
	TokenSentSig                     = crypto.Keccak256Hash([]byte("TokenSent(address,string,string,string,uint256)"))
)

// Mgr manages all communication with Ethereum
type Mgr struct {
	cliCtx      sdkClient.Context
	logger      tmLog.Logger
	rpcs        map[string]rpc.Client
	broadcaster broadcast.Broadcaster
	cdc         *codec.LegacyAmino
}

// NewMgr returns a new Mgr instance
func NewMgr(rpcs map[string]rpc.Client, cliCtx sdkClient.Context, broadcaster broadcast.Broadcaster, logger tmLog.Logger, cdc *codec.LegacyAmino) *Mgr {
	return &Mgr{
		rpcs:        rpcs,
		cliCtx:      cliCtx,
		broadcaster: broadcaster,
		logger:      logger.With("listener", "evm"),
		cdc:         cdc,
	}
}

// ProcessNewChain notifies the operator that vald needs to be restarted/udpated for a new chain
func (mgr Mgr) ProcessNewChain(e tmEvents.Event) (err error) {
	chain, nativeAsset, err := parseNewChainParams(e.Attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "Invalid update event")
	}

	mgr.logger.Info(fmt.Sprintf("VALD needs to be updated and restarted for new chain %s with native asset %s", chain.String(), nativeAsset))
	return nil
}

// ProcessDepositConfirmation votes on the correctness of an EVM chain token deposit
func (mgr Mgr) ProcessDepositConfirmation(e tmEvents.Event) (err error) {
	chain, txID, burnAddr, tokenAddr, confHeight, pollID, err := parseDepositConfirmationParams(mgr.cdc, e.Attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "EVM deposit confirmation failed")
	}

	rpc, found := mgr.rpcs[strings.ToLower(chain.String())]
	if !found {
		return sdkerrors.Wrap(err, fmt.Sprintf("Unable to find an RPC for chain '%s'", chain))
	}
	var events []evmTypes.Event
	_ = mgr.validate(rpc, txID, confHeight, func(_ *geth.Transaction, txReceipt *geth.Receipt) bool {
		for i, log := range txReceipt.Logs {
			switch log.Topics[0] {
			case ERC20TransferSig:
				if !bytes.Equal(tokenAddr.Bytes(), log.Address.Bytes()) {
					continue
				}

				event, err := decodeERC20TransferEvent(log)
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "decode event Transfer failed").Error())
					continue
				}

				if event.To != evmTypes.Address(burnAddr) {
					continue
				}

				events = append(events, evmTypes.Event{
					Chain: chain,
					TxId:  evmTypes.Hash(txID),
					Index: uint64(i),
					Event: &evmTypes.Event_Transfer{
						Transfer: &event,
					},
				})
			default:
			}

		}
		return true
	})

	msg := voteTypes.NewVoteRequest(mgr.cliCtx.FromAddress, pollID, evmTypes.NewVoteEvents(chain, events))
	mgr.logger.Info(fmt.Sprintf("broadcasting vote %v for poll %s", events, pollID.String()))
	_, err = mgr.broadcaster.Broadcast(context.TODO(), msg)
	return err
}

// ProcessTokenConfirmation votes on the correctness of an EVM chain token deployment
func (mgr Mgr) ProcessTokenConfirmation(e tmEvents.Event) error {
	chain, txID, gatewayAddr, tokenAddr, symbol, confHeight, pollID, err := parseTokenConfirmationParams(mgr.cdc, e.Attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "EVM token deployment confirmation failed")
	}

	rpc, found := mgr.rpcs[strings.ToLower(chain.String())]
	if !found {
		return sdkerrors.Wrap(err, fmt.Sprintf("Unable to find an RPC for chain '%s'", chain))
	}

	var events []evmTypes.Event
	_ = mgr.validate(rpc, txID, confHeight, func(_ *geth.Transaction, txReceipt *geth.Receipt) bool {
		for i, log := range txReceipt.Logs {
			if !bytes.Equal(gatewayAddr.Bytes(), log.Address.Bytes()) {
				continue
			}

			switch log.Topics[0] {
			case ERC20TokenDeploymentSig:
				event, err := decodeERC20TokenDeploymentEvent(log)
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "decode event TokenDeployed failed").Error())
					continue
				}
				if event.TokenAddress != evmTypes.Address(tokenAddr) || event.Symbol != symbol {
					continue
				}
				events = append(events, evmTypes.Event{
					Chain: chain,
					TxId:  evmTypes.Hash(txID),
					Index: uint64(i),
					Event: &evmTypes.Event_TokenDeployed{
						TokenDeployed: &event,
					},
				})

				return true
			default:
			}
		}

		return true
	})

	msg := voteTypes.NewVoteRequest(mgr.cliCtx.FromAddress, pollID, evmTypes.NewVoteEvents(chain, events))
	mgr.logger.Info(fmt.Sprintf("broadcasting vote %v for poll %s", events, pollID.String()))
	_, err = mgr.broadcaster.Broadcast(context.TODO(), msg)
	return err
}

// ProcessTransferKeyConfirmation votes on the correctness of an EVM chain key transfer
func (mgr Mgr) ProcessTransferKeyConfirmation(e tmEvents.Event) (err error) {
	chain, txID, keyType, gatewayAddr, confHeight, pollID, err := parseTransferKeyConfirmationParams(mgr.cdc, e.Attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "EVM key transfer confirmation failed")
	}

	rpc, found := mgr.rpcs[strings.ToLower(chain.String())]
	if !found {
		return sdkerrors.Wrap(err, fmt.Sprintf("Unable to find an RPC for chain '%s'", chain))
	}

	var events []evmTypes.Event
	_ = mgr.validate(rpc, txID, confHeight, func(_ *geth.Transaction, txReceipt *geth.Receipt) bool {
		for i := len(txReceipt.Logs) - 1; i >= 0; i-- {
			log := txReceipt.Logs[i]

			// Event is not emitted by the axelar gateway
			if log.Address != gatewayAddr {
				continue
			}

			switch {
			case keyType == tss.Threshold && log.Topics[0] == SinglesigTransferOperatorshipSig:
				event, err := decodeSinglesigOperatorshipTransferredEvent(log)
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "key transfer confirmation failed").Error())
					continue
				}

				events = append(events, evmTypes.Event{
					Chain: chain,
					TxId:  evmTypes.Hash(txID),
					Index: uint64(i),
					Event: &evmTypes.Event_SinglesigOperatorshipTransferred{
						SinglesigOperatorshipTransferred: &event,
					},
				})
			case keyType == tss.Multisig && log.Topics[0] == MultisigTransferOperatorshipSig:
				event, err := decodeMultisigOperatorshipTransferredEvent(log)
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "key transfer confirmation failed").Error())
					continue
				}

				events = append(events, evmTypes.Event{
					Chain: chain,
					TxId:  evmTypes.Hash(txID),
					Index: uint64(i),
					Event: &evmTypes.Event_MultisigOperatorshipTransferred{
						MultisigOperatorshipTransferred: &event,
					},
				})
			default:
			}

			// There might be several transfer ownership/operatorship event. Only interest in the last one.
			if len(events) != 0 {
				break
			}
		}

		return true
	})

	msg := voteTypes.NewVoteRequest(mgr.cliCtx.FromAddress, pollID, evmTypes.NewVoteEvents(chain, events))
	mgr.logger.Info(fmt.Sprintf("broadcasting vote %v for poll %s", events, pollID.String()))
	_, err = mgr.broadcaster.Broadcast(context.TODO(), msg)
	return err
}

// ProcessGatewayTxConfirmation votes on the correctness of an EVM chain gateway's transactions
func (mgr Mgr) ProcessGatewayTxConfirmation(e tmEvents.Event) error {
	chain, gatewayAddress, txID, confHeight, pollID, err := parseGatewayTxConfirmationParams(mgr.cdc, e.Attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "EVM gateway transaction confirmation failed")
	}

	rpc, found := mgr.rpcs[strings.ToLower(chain.String())]
	if !found {
		return sdkerrors.Wrap(err, fmt.Sprintf("Unable to find an RPC for chain '%s'", chain))
	}

	var events []evmTypes.Event
	_ = mgr.validate(rpc, txID, confHeight, func(_ *geth.Transaction, txReceipt *geth.Receipt) bool {
		for i, log := range txReceipt.Logs {
			if !bytes.Equal(gatewayAddress.Bytes(), log.Address.Bytes()) {
				continue
			}

			switch log.Topics[0] {
			case ContractCallSig:
				event, err := decodeEventContractCall(log)
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "decode event ContractCall failed").Error())

					return false
				}

				err = event.ValidateBasic()
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "invalid event ContractCall").Error())
					continue
				}

				events = append(events, evmTypes.Event{
					Chain: chain,
					TxId:  evmTypes.Hash(txID),
					Index: uint64(i),
					Event: &evmTypes.Event_ContractCall{
						ContractCall: &event,
					},
				})
			case ContractCallWithTokenSig:
				event, err := decodeEventContractCallWithToken(log)
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "decode event ContractCallWithToken failed").Error())

					return false
				}

				err = event.ValidateBasic()
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "invalid event ContractCallWithToken").Error())
					continue
				}

				events = append(events, evmTypes.Event{
					Chain: chain,
					TxId:  evmTypes.Hash(txID),
					Index: uint64(i),
					Event: &evmTypes.Event_ContractCallWithToken{
						ContractCallWithToken: &event,
					},
				})
			case TokenSentSig:
				event, err := decodeEventTokenSent(log)
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "decode event TokenSent failed").Error())
				}

				err = event.ValidateBasic()
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "invalid event TokenSent").Error())
					continue
				}

				events = append(events, evmTypes.Event{
					Chain: chain,
					TxId:  evmTypes.Hash(txID),
					Index: uint64(i),
					Event: &evmTypes.Event_TokenSent{
						TokenSent: &event,
					},
				})
			default:
			}
		}

		return true
	})

	msg := voteTypes.NewVoteRequest(mgr.cliCtx.FromAddress, pollID, evmTypes.NewVoteEvents(chain, events))
	mgr.logger.Info(fmt.Sprintf("broadcasting vote %v for poll %s", events, pollID.String()))
	_, err = mgr.broadcaster.Broadcast(context.TODO(), msg)
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

func parseGatewayTxConfirmationParams(cdc *codec.LegacyAmino, attributes map[string]string) (
	chain nexus.ChainName,
	gatewayAddress common.Address,
	txID common.Hash,
	confHeight uint64,
	pollID vote.PollID,
	err error,
) {
	parsers := []*parse.AttributeParser{
		{Key: evmTypes.AttributeKeyChain, Map: func(s string) (interface{}, error) {
			return nexus.ChainName(s), nil
		}},
		{Key: evmTypes.AttributeKeyGatewayAddress, Map: func(s string) (interface{}, error) {
			return common.HexToAddress(s), nil
		}},
		{Key: evmTypes.AttributeKeyTxID, Map: func(s string) (interface{}, error) {
			return common.HexToHash(s), nil
		}},
		{Key: evmTypes.AttributeKeyConfHeight, Map: func(s string) (interface{}, error) { return strconv.ParseUint(s, 10, 64) }},
		{Key: evmTypes.AttributeKeyPoll, Map: func(s string) (interface{}, error) {
			id, err := strconv.ParseUint(s, 10, 64)
			return vote.PollID(id), err
		}},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		return "", common.Address{}, common.Hash{}, 0, 0, err
	}

	return results[0].(nexus.ChainName),
		results[1].(common.Address),
		results[2].(common.Hash),
		results[3].(uint64),
		results[4].(vote.PollID),
		nil
}

func parseNewChainParams(attributes map[string]string) (chain nexus.ChainName, nativeAsset string, err error) {
	parsers := []*parse.AttributeParser{
		{Key: evmTypes.AttributeKeyChain, Map: func(s string) (interface{}, error) {
			return nexus.ChainName(s), nil
		}},
		{Key: evmTypes.AttributeKeyNativeAsset, Map: parse.IdentityMap},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		return "", "", err
	}

	return results[0].(nexus.ChainName), results[1].(string), nil
}

func parseDepositConfirmationParams(cdc *codec.LegacyAmino, attributes map[string]string) (
	chain nexus.ChainName,
	txID common.Hash,
	burnAddr, tokenAddr common.Address,
	confHeight uint64,
	pollID vote.PollID,
	err error,
) {
	parsers := []*parse.AttributeParser{
		{Key: evmTypes.AttributeKeyChain, Map: func(s string) (interface{}, error) {
			return nexus.ChainName(s), nil
		}},
		{Key: evmTypes.AttributeKeyTxID, Map: func(s string) (interface{}, error) {
			return common.HexToHash(s), nil
		}},
		{Key: evmTypes.AttributeKeyDepositAddress, Map: func(s string) (interface{}, error) {
			return common.HexToAddress(s), nil
		}},
		{Key: evmTypes.AttributeKeyTokenAddress, Map: func(s string) (interface{}, error) {
			return common.HexToAddress(s), nil
		}},
		{Key: evmTypes.AttributeKeyConfHeight, Map: func(s string) (interface{}, error) { return strconv.ParseUint(s, 10, 64) }},
		{Key: evmTypes.AttributeKeyPoll, Map: func(s string) (interface{}, error) {
			id, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return vote.PollID(0), err
			}

			return vote.PollID(id), nil
		}},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		return "", [32]byte{}, [20]byte{}, [20]byte{}, 0, 0, err
	}

	return results[0].(nexus.ChainName),
		results[1].(common.Hash),
		results[2].(common.Address),
		results[3].(common.Address),
		results[4].(uint64),
		results[5].(vote.PollID),
		nil
}

func parseTokenConfirmationParams(cdc *codec.LegacyAmino, attributes map[string]string) (
	chain nexus.ChainName,
	txID common.Hash,
	gatewayAddr, tokenAddr common.Address,
	symbol string,
	confHeight uint64,
	pollID vote.PollID,
	err error,
) {
	parsers := []*parse.AttributeParser{
		{Key: evmTypes.AttributeKeyChain, Map: func(s string) (interface{}, error) {
			return nexus.ChainName(s), nil
		}},
		{Key: evmTypes.AttributeKeyTxID, Map: func(s string) (interface{}, error) {
			return common.HexToHash(s), nil
		}},
		{Key: evmTypes.AttributeKeyGatewayAddress, Map: func(s string) (interface{}, error) {
			return common.HexToAddress(s), nil
		}},
		{Key: evmTypes.AttributeKeyTokenAddress, Map: func(s string) (interface{}, error) {
			return common.HexToAddress(s), nil
		}},
		{Key: evmTypes.AttributeKeySymbol, Map: parse.IdentityMap},
		{Key: evmTypes.AttributeKeyConfHeight, Map: func(s string) (interface{}, error) { return strconv.ParseUint(s, 10, 64) }},
		{Key: evmTypes.AttributeKeyPoll, Map: func(s string) (interface{}, error) {
			id, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return vote.PollID(0), err
			}

			return vote.PollID(id), nil
		}},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		return "", [32]byte{}, [20]byte{}, [20]byte{}, "", 0, 0, err
	}

	return results[0].(nexus.ChainName),
		results[1].(common.Hash),
		results[2].(common.Address),
		results[3].(common.Address),
		results[4].(string),
		results[5].(uint64),
		results[6].(vote.PollID),
		nil
}

func parseTransferKeyConfirmationParams(cdc *codec.LegacyAmino, attributes map[string]string) (
	chain nexus.ChainName,
	txID common.Hash,
	keyType tss.KeyType,
	gatewayAddr common.Address,
	confHeight uint64,
	pollID vote.PollID,
	err error,
) {
	parsers := []*parse.AttributeParser{
		{Key: evmTypes.AttributeKeyChain, Map: func(s string) (interface{}, error) {
			return nexus.ChainName(s), nil
		}},
		{Key: evmTypes.AttributeKeyTxID, Map: func(s string) (interface{}, error) {
			return common.HexToHash(s), nil
		}},
		{Key: evmTypes.AttributeKeyKeyType, Map: func(s string) (interface{}, error) {
			return tss.KeyTypeFromSimpleStr(s)
		}},
		{Key: evmTypes.AttributeKeyGatewayAddress, Map: func(s string) (interface{}, error) {
			return common.HexToAddress(s), nil
		}},
		{Key: evmTypes.AttributeKeyConfHeight, Map: func(s string) (interface{}, error) { return strconv.ParseUint(s, 10, 64) }},
		{Key: evmTypes.AttributeKeyPoll, Map: func(s string) (interface{}, error) {
			id, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return vote.PollID(0), err
			}

			return vote.PollID(id), nil
		}},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		return "", common.Hash{}, tss.KEY_TYPE_UNSPECIFIED, common.Address{}, 0, 0, err
	}

	return results[0].(nexus.ChainName),
		results[1].(common.Hash),
		results[2].(tss.KeyType),
		results[3].(common.Address),
		results[4].(uint64),
		results[5].(vote.PollID),
		nil
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
	for _, tx := range txBlock.Body().Transactions {
		if bytes.Equal(tx.Hash().Bytes(), txReceipt.TxHash.Bytes()) {
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

func decodeSinglesigOperatorshipTransferredEvent(log *geth.Log) (evmTypes.EventSinglesigOperatorshipTransferred, error) {
	if len(log.Topics) != 3 || log.Topics[0] != SinglesigTransferOperatorshipSig {
		return evmTypes.EventSinglesigOperatorshipTransferred{}, fmt.Errorf("event is not for a transfer singlesig key")
	}

	return evmTypes.EventSinglesigOperatorshipTransferred{
		PreOperator: evmTypes.Address(common.BytesToAddress(log.Topics[1][:])),
		NewOperator: evmTypes.Address(common.BytesToAddress(log.Topics[2][:])),
	}, nil
}

func decodeMultisigOperatorshipTransferredEvent(log *geth.Log) (evmTypes.EventMultisigOperatorshipTransferred, error) {
	if len(log.Topics) != 1 || log.Topics[0] != MultisigTransferOperatorshipSig {
		return evmTypes.EventMultisigOperatorshipTransferred{}, fmt.Errorf("event is not a MultisigTransferOwnershipSig")
	}

	newAddresses, newThreshold, err := unpackMultisigTransferKeyEvent(log)
	if err != nil {
		return evmTypes.EventMultisigOperatorshipTransferred{}, err
	}

	return evmTypes.EventMultisigOperatorshipTransferred{
		NewOperators: slices.Map(newAddresses, func(addr common.Address) evmTypes.Address { return evmTypes.Address(addr) }),
		NewThreshold: sdk.NewUintFromBigInt(newThreshold),
	}, nil
}

func unpackMultisigTransferKeyEvent(log *geth.Log) ([]common.Address, *big.Int, error) {
	bytesType, err := abi.NewType("bytes", "bytes", nil)
	if err != nil {
		return []common.Address{}, &big.Int{}, err
	}

	operatorData, err := evmTypes.StrictDecode(abi.Arguments{{Type: bytesType}}, log.Data)
	if err != nil {
		return []common.Address{}, &big.Int{}, err
	}

	addressesType, err := abi.NewType("address[]", "address[]", nil)
	if err != nil {
		return []common.Address{}, &big.Int{}, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return []common.Address{}, &big.Int{}, err
	}

	params, err := evmTypes.StrictDecode(abi.Arguments{{Type: addressesType}, {Type: uint256Type}}, operatorData[0].([]byte))
	if err != nil {
		return []common.Address{}, &big.Int{}, err
	}

	return params[0].([]common.Address), params[1].(*big.Int), nil
}
