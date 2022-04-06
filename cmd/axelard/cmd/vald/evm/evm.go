package evm

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"sort"
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

	tmEvents "github.com/axelarnetwork/tm-events/events"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcaster/types"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/parse"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
)

// Smart contract event signatures
var (
	ERC20TransferSig                 = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	ERC20TokenDeploymentSig          = crypto.Keccak256Hash([]byte("TokenDeployed(string,address)"))
	SinglesigTransferOwnershipSig    = crypto.Keccak256Hash([]byte("OwnershipTransferred(address,address)"))
	SinglesigTransferOperatorshipSig = crypto.Keccak256Hash([]byte("OperatorshipTransferred(address,address)"))
	MultisigTransferOwnershipSig     = crypto.Keccak256Hash([]byte("OwnershipTransferred(address[],uint256,address[],uint256)"))
	MultisigTransferOperatorshipSig  = crypto.Keccak256Hash([]byte("OperatorshipTransferred(address[],uint256,address[],uint256)"))
	ContractCallSig                  = crypto.Keccak256Hash([]byte("ContractCall(address,string,string,bytes32,bytes)"))
	ContractCallWithTokenSig         = crypto.Keccak256Hash([]byte("ContractCallWithToken(address,string,string,bytes32,bytes,string,uint256)"))
	TokenSentSig                     = crypto.Keccak256Hash([]byte("TokenSent(address,string,string,string,uint256)"))
)

// Mgr manages all communication with Ethereum
type Mgr struct {
	cliCtx      sdkClient.Context
	logger      tmLog.Logger
	rpcs        map[string]rpc.Client
	broadcaster types.Broadcaster
	cdc         *codec.LegacyAmino
}

// NewMgr returns a new Mgr instance
func NewMgr(rpcs map[string]rpc.Client, cliCtx sdkClient.Context, broadcaster types.Broadcaster, logger tmLog.Logger, cdc *codec.LegacyAmino) *Mgr {
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

	mgr.logger.Info(fmt.Sprintf("VALD needs to be updated and restarted for new chain %s with native asset %s", chain, nativeAsset))
	return nil
}

// ProcessChainConfirmation votes on the correctness of an EVM chain token deposit
func (mgr Mgr) ProcessChainConfirmation(e tmEvents.Event) (err error) {
	chain, pollKey, err := parseChainConfirmationParams(mgr.cdc, e.Attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "EVM chain confirmation failed")
	}

	_, confirmed := mgr.rpcs[strings.ToLower(chain)]

	msg := evmTypes.NewVoteConfirmChainRequest(mgr.cliCtx.FromAddress, chain, pollKey, confirmed)
	mgr.logger.Info(fmt.Sprintf("broadcasting vote %v for poll %s", msg.Confirmed, pollKey.String()))
	_, err = mgr.broadcaster.Broadcast(context.TODO(), msg)
	return err
}

// ProcessDepositConfirmation votes on the correctness of an EVM chain token deposit
func (mgr Mgr) ProcessDepositConfirmation(e tmEvents.Event) (err error) {
	chain, txID, burnAddr, tokenAddr, confHeight, pollKey, err := parseDepositConfirmationParams(mgr.cdc, e.Attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "EVM deposit confirmation failed")
	}

	rpc, found := mgr.rpcs[strings.ToLower(chain)]
	if !found {
		return sdkerrors.Wrap(err, fmt.Sprintf("Unable to find an RPC for chain '%s'", chain))
	}
	var events []evmTypes.Event
	confirmed := mgr.validate(rpc, txID, confHeight, func(_ *geth.Transaction, txReceipt *geth.Receipt) bool {
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

	v, err := packEvents(confirmed, events)
	if err != nil {
		return err
	}
	msg := voteTypes.NewVoteRequest(mgr.cliCtx.FromAddress, pollKey, v, chain)
	mgr.logger.Info(fmt.Sprintf("broadcasting vote %v for poll %s", events, pollKey.String()))
	_, err = mgr.broadcaster.Broadcast(context.TODO(), msg)
	return err
}

// ProcessTokenConfirmation votes on the correctness of an EVM chain token deployment
func (mgr Mgr) ProcessTokenConfirmation(e tmEvents.Event) error {
	chain, txID, gatewayAddr, tokenAddr, symbol, confHeight, pollKey, err := parseTokenConfirmationParams(mgr.cdc, e.Attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "EVM token deployment confirmation failed")
	}

	rpc, found := mgr.rpcs[strings.ToLower(chain)]
	if !found {
		return sdkerrors.Wrap(err, fmt.Sprintf("Unable to find an RPC for chain '%s'", chain))
	}

	var events []evmTypes.Event
	confirmed := mgr.validate(rpc, txID, confHeight, func(_ *geth.Transaction, txReceipt *geth.Receipt) bool {
		for i, log := range txReceipt.Logs {
			if !bytes.Equal(gatewayAddr.Bytes(), log.Address.Bytes()) {
				continue
			}

			switch log.Topics[0] {
			case ERC20TokenDeploymentSig:
				event, err := decodeERC20TokenDeploymentEvent(log)
				if err != nil {
					mgr.logger.Debug(sdkerrors.Wrap(err, "decode event TokenDeployed failed").Error())
					return false
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
			default:
			}
		}

		return true
	})

	v, err := packEvents(confirmed, events)
	if err != nil {
		return err
	}
	msg := voteTypes.NewVoteRequest(mgr.cliCtx.FromAddress, pollKey, v, chain)
	mgr.logger.Info(fmt.Sprintf("broadcasting vote %v for poll %s", events, pollKey.String()))
	_, err = mgr.broadcaster.Broadcast(context.TODO(), msg)
	return err
}

// ProcessTransferKeyConfirmation votes on the correctness of an EVM chain key transfer
func (mgr Mgr) ProcessTransferKeyConfirmation(e tmEvents.Event) (err error) {
	chain, txID, transferKeyType, keyType, gatewayAddr, newAddrs, threshold, confHeight, pollKey, err := parseTransferKeyConfirmationParams(mgr.cdc, e.Attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "EVM key transfer confirmation failed")
	}

	rpc, found := mgr.rpcs[strings.ToLower(chain)]
	if !found {
		return sdkerrors.Wrap(err, fmt.Sprintf("Unable to find an RPC for chain '%s'", chain))
	}

	confirmed := mgr.validate(rpc, txID, confHeight, func(_ *geth.Transaction, txReceipt *geth.Receipt) bool {
		switch keyType {
		case tss.Threshold:
			if err = confirmSinglesigTransferKey(txReceipt, transferKeyType, gatewayAddr, newAddrs[0]); err != nil {
				mgr.logger.Debug(sdkerrors.Wrapf(err, "%s key transfer confirmation failed", transferKeyType.SimpleString()).Error())
				return false
			}
		case tss.Multisig:
			if err = confirmMultisigTransferKey(txReceipt, transferKeyType, gatewayAddr, newAddrs, threshold); err != nil {
				mgr.logger.Debug(sdkerrors.Wrapf(err, "%s key transfer confirmation failed", transferKeyType.SimpleString()).Error())
				return false
			}
		default:
			mgr.logger.Error(fmt.Sprintf("unknown key type %s", keyType.SimpleString()))
			return false
		}

		return true
	})

	msg := evmTypes.NewVoteConfirmTransferKeyRequest(mgr.cliCtx.FromAddress, chain, pollKey, confirmed)
	mgr.logger.Info(fmt.Sprintf("broadcasting vote %v for poll %s", msg.Confirmed, pollKey.String()))
	_, err = mgr.broadcaster.Broadcast(context.TODO(), msg)
	return err
}

// ProcessGatewayTxConfirmation votes on the correctness of an EVM chain gateway's transactions
func (mgr Mgr) ProcessGatewayTxConfirmation(e tmEvents.Event) error {
	chain, gatewayAddress, txID, confHeight, pollKey, err := parseGatewayTxConfirmationParams(mgr.cdc, e.Attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "EVM gateway transaction confirmation failed")
	}

	rpc, found := mgr.rpcs[strings.ToLower(chain)]
	if !found {
		return sdkerrors.Wrap(err, fmt.Sprintf("Unable to find an RPC for chain '%s'", chain))
	}

	var events []evmTypes.Event
	confirmed := mgr.validate(rpc, txID, confHeight, func(_ *geth.Transaction, txReceipt *geth.Receipt) bool {
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

					return false
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

	var vote evmTypes.VoteConfirmGatewayTxRequest_Vote
	if confirmed {
		vote.Events = events
	}

	msg := evmTypes.NewVoteConfirmGatewayTxRequest(mgr.cliCtx.FromAddress, pollKey, vote)
	mgr.logger.Info(fmt.Sprintf("broadcasting vote %v for poll %s", vote.String(), pollKey.String()))
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
		DestinationChain:   params[0].(string),
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
		DestinationChain: params[0].(string),
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
		DestinationChain: params[0].(string),
		ContractAddress:  params[1].(string),
		PayloadHash:      evmTypes.Hash(common.BytesToHash(log.Topics[2].Bytes())),
		Symbol:           params[3].(string),
		Amount:           sdk.NewUintFromBigInt(params[4].(*big.Int)),
	}, nil
}

func parseGatewayTxConfirmationParams(cdc *codec.LegacyAmino, attributes map[string]string) (
	chain string,
	gatewayAddress common.Address,
	txID common.Hash,
	confHeight uint64,
	pollKey vote.PollKey,
	err error,
) {
	parsers := []*parse.AttributeParser{
		{Key: evmTypes.AttributeKeyChain, Map: parse.IdentityMap},
		{Key: evmTypes.AttributeKeyGatewayAddress, Map: func(s string) (interface{}, error) {
			return common.HexToAddress(s), nil
		}},
		{Key: evmTypes.AttributeKeyTxID, Map: func(s string) (interface{}, error) {
			return common.HexToHash(s), nil
		}},
		{Key: evmTypes.AttributeKeyConfHeight, Map: func(s string) (interface{}, error) { return strconv.ParseUint(s, 10, 64) }},
		{Key: evmTypes.AttributeKeyPoll, Map: func(s string) (interface{}, error) {
			cdc.MustUnmarshalJSON([]byte(s), &pollKey)

			return pollKey, nil
		}},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		return "", common.Address{}, common.Hash{}, 0, vote.PollKey{}, err
	}

	return results[0].(string),
		results[1].(common.Address),
		results[2].(common.Hash),
		results[3].(uint64),
		results[4].(vote.PollKey),
		nil
}

func parseNewChainParams(attributes map[string]string) (chain string, nativeAsset string, err error) {
	parsers := []*parse.AttributeParser{
		{Key: evmTypes.AttributeKeyChain, Map: parse.IdentityMap},
		{Key: evmTypes.AttributeKeyNativeAsset, Map: parse.IdentityMap},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		return "", "", err
	}

	return results[0].(string), results[1].(string), nil
}

func parseChainConfirmationParams(cdc *codec.LegacyAmino, attributes map[string]string) (
	chain string,
	pollKey vote.PollKey,
	err error,
) {
	parsers := []*parse.AttributeParser{
		{Key: evmTypes.AttributeKeyChain, Map: parse.IdentityMap},
		{Key: evmTypes.AttributeKeyPoll, Map: func(s string) (interface{}, error) {
			cdc.MustUnmarshalJSON([]byte(s), &pollKey)
			return pollKey, nil
		}},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		return "", vote.PollKey{}, err
	}

	return results[0].(string), results[1].(vote.PollKey), nil
}

func parseDepositConfirmationParams(cdc *codec.LegacyAmino, attributes map[string]string) (
	chain string,
	txID common.Hash,
	burnAddr, tokenAddr common.Address,
	confHeight uint64,
	pollKey vote.PollKey,
	err error,
) {
	parsers := []*parse.AttributeParser{
		{Key: evmTypes.AttributeKeyChain, Map: parse.IdentityMap},
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
			cdc.MustUnmarshalJSON([]byte(s), &pollKey)
			return pollKey, nil
		}},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		return "", [32]byte{}, [20]byte{}, [20]byte{}, 0, vote.PollKey{}, err
	}

	return results[0].(string),
		results[1].(common.Hash),
		results[2].(common.Address),
		results[3].(common.Address),
		results[4].(uint64),
		results[5].(vote.PollKey),
		nil
}

func parseTokenConfirmationParams(cdc *codec.LegacyAmino, attributes map[string]string) (
	chain string,
	txID common.Hash,
	gatewayAddr, tokenAddr common.Address,
	symbol string,
	confHeight uint64,
	pollKey vote.PollKey,
	err error,
) {
	parsers := []*parse.AttributeParser{
		{Key: evmTypes.AttributeKeyChain, Map: parse.IdentityMap},
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
			cdc.MustUnmarshalJSON([]byte(s), &pollKey)
			return pollKey, nil
		}},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		return "", [32]byte{}, [20]byte{}, [20]byte{}, "", 0, vote.PollKey{}, err
	}

	return results[0].(string),
		results[1].(common.Hash),
		results[2].(common.Address),
		results[3].(common.Address),
		results[4].(string),
		results[5].(uint64),
		results[6].(vote.PollKey),
		nil
}

func parseTransferKeyConfirmationParams(cdc *codec.LegacyAmino, attributes map[string]string) (
	chain string,
	txID common.Hash,
	transferKeyType evmTypes.TransferKeyType,
	keyType tss.KeyType,
	gatewayAddr common.Address,
	newAddrs []common.Address,
	threshold uint8,
	confHeight uint64,
	pollKey vote.PollKey,
	err error,
) {
	parsers := []*parse.AttributeParser{
		{Key: evmTypes.AttributeKeyChain, Map: parse.IdentityMap},
		{Key: evmTypes.AttributeKeyTxID, Map: func(s string) (interface{}, error) {
			return common.HexToHash(s), nil
		}},
		{Key: evmTypes.AttributeKeyTransferKeyType, Map: func(s string) (interface{}, error) {
			return evmTypes.TransferKeyTypeFromSimpleStr(s)
		}},
		{Key: evmTypes.AttributeKeyKeyType, Map: func(s string) (interface{}, error) {
			return tss.KeyTypeFromSimpleStr(s)
		}},
		{Key: evmTypes.AttributeKeyGatewayAddress, Map: func(s string) (interface{}, error) {
			return common.HexToAddress(s), nil
		}},
		{Key: evmTypes.AttributeKeyAddress, Map: func(s string) (interface{}, error) {
			addressStrs := strings.Split(s, ",")
			addresses := make([]common.Address, len(addressStrs))

			for i, addressStr := range addressStrs {
				addresses[i] = common.HexToAddress(addressStr)
			}

			return addresses, nil
		}},
		{Key: evmTypes.AttributeKeyThreshold, Map: func(s string) (interface{}, error) {
			if s == "" {
				return uint8(0), nil
			}

			threshold, err := strconv.ParseInt(s, 10, 8)
			if err != nil {
				return 0, err
			}

			return uint8(threshold), nil
		}},
		{Key: evmTypes.AttributeKeyConfHeight, Map: func(s string) (interface{}, error) { return strconv.ParseUint(s, 10, 64) }},
		{Key: evmTypes.AttributeKeyPoll, Map: func(s string) (interface{}, error) {
			cdc.MustUnmarshalJSON([]byte(s), &pollKey)
			return pollKey, nil
		}},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		return "", common.Hash{}, evmTypes.UnspecifiedTransferKeyType, tss.KEY_TYPE_UNSPECIFIED, common.Address{}, nil, 0, 0, vote.PollKey{}, err
	}

	return results[0].(string),
		results[1].(common.Hash),
		results[2].(evmTypes.TransferKeyType),
		results[3].(tss.KeyType),
		results[4].(common.Address),
		results[5].([]common.Address),
		results[6].(uint8),
		results[7].(uint64),
		results[8].(vote.PollKey),
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
	default:
		blockNumber, err := client.BlockNumber(context.Background())
		if err != nil {
			return nil, err
		}

		return big.NewInt(int64(blockNumber - confHeight + 1)), nil
	}
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

func confirmSinglesigTransferKey(txReceipt *geth.Receipt, transferKeyType evmTypes.TransferKeyType, gatewayAddr common.Address, expectedNewAddr common.Address) (err error) {
	for i := len(txReceipt.Logs) - 1; i >= 0; i-- {
		log := txReceipt.Logs[i]
		// Event is not emitted by the axelar gateway
		if log.Address != gatewayAddr {
			continue
		}

		// There might be several transfer ownership/operatorship event. Only interest in the last one.
		actualNewAddr, err := decodeSinglesigKeyTransferEvent(log, transferKeyType)
		if err != nil {
			continue
		}

		// New addr does not match
		if actualNewAddr != expectedNewAddr {
			break
		}

		// if we reach this point, it means that the log matches what we want to verify,
		// so the function can return with no error
		return nil
	}

	return fmt.Errorf("failed to confirm %s transfer for new address '%s' at contract address '%s'", transferKeyType.SimpleString(), expectedNewAddr.String(), gatewayAddr.String())
}

func addressesToHexes(addresses []common.Address) []string {
	hexes := make([]string, len(addresses))
	for i, address := range addresses {
		hexes[i] = address.Hex()
	}

	return hexes
}

func areAddressesEqual(addressesA, addressesB []common.Address) bool {
	if len(addressesA) != len(addressesB) {
		return false
	}

	hexesA := addressesToHexes(addressesA)
	sort.Strings(hexesA)
	hexesB := addressesToHexes(addressesB)
	sort.Strings(hexesB)

	for i, hexA := range hexesA {
		if hexA != hexesB[i] {
			return false
		}
	}

	return true
}

func confirmMultisigTransferKey(txReceipt *geth.Receipt, transferKeyType evmTypes.TransferKeyType, gatewayAddr common.Address, expectedNewAddrs []common.Address, expectedNewThreshold uint8) (err error) {
	for i := len(txReceipt.Logs) - 1; i >= 0; i-- {
		log := txReceipt.Logs[i]
		// Event is not emitted by the axelar gateway
		if log.Address != gatewayAddr {
			continue
		}

		// There might be several transfer ownership/operatorship event. Only interest in the last one.
		actualNewAddrs, actualNewThreshold, err := decodeMultisigKeyTransferEvent(log, transferKeyType)
		if err != nil {
			continue
		}

		// New addrs or threshold does not match
		if !areAddressesEqual(actualNewAddrs, expectedNewAddrs) || actualNewThreshold != expectedNewThreshold {
			break
		}

		// if we reach this point, it means that the log matches what we want to verify,
		// so the function can return with no error
		return nil
	}

	return fmt.Errorf("failed to confirm %s transfer for new addresses '%s' and threshold '%d' at contract address '%s'", transferKeyType.SimpleString(), strings.Join(addressesToHexes(expectedNewAddrs), ","), expectedNewThreshold, gatewayAddr.String())
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

func decodeSinglesigKeyTransferEvent(log *geth.Log, transferKeyType evmTypes.TransferKeyType) (common.Address, error) {
	var topic common.Hash
	switch transferKeyType {
	case evmTypes.Ownership:
		topic = SinglesigTransferOwnershipSig
	case evmTypes.Operatorship:
		topic = SinglesigTransferOperatorshipSig
	default:
		return common.Address{}, fmt.Errorf("unknown transfer key type %s", transferKeyType.SimpleString())
	}

	if len(log.Topics) != 3 || log.Topics[0] != topic {
		return common.Address{}, fmt.Errorf("event is not for a transfer singlesig key")
	}

	return common.BytesToAddress(log.Topics[2][:]), nil
}

func decodeMultisigKeyTransferEvent(log *geth.Log, transferKeyType evmTypes.TransferKeyType) ([]common.Address, uint8, error) {
	var topic common.Hash
	switch transferKeyType {
	case evmTypes.Ownership:
		topic = MultisigTransferOwnershipSig
	case evmTypes.Operatorship:
		topic = MultisigTransferOperatorshipSig
	default:
		return []common.Address{}, 0, fmt.Errorf("unknown transfer key type %s", transferKeyType.SimpleString())
	}

	if len(log.Topics) != 1 || log.Topics[0] != topic {
		return []common.Address{}, 0, fmt.Errorf("event is not for a transfer multisig key")
	}

	addressesType, err := abi.NewType("address[]", "address[]", nil)
	if err != nil {
		return []common.Address{}, 0, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return []common.Address{}, 0, err
	}

	arguments := abi.Arguments{{Type: addressesType}, {Type: uint256Type}, {Type: addressesType}, {Type: uint256Type}}
	params, err := evmTypes.StrictDecode(arguments, log.Data)
	if err != nil {
		return []common.Address{}, 0, err
	}

	if len(params) != 4 {
		return []common.Address{}, 0, fmt.Errorf("event is not for a transfer multisig key")
	}

	addresses, ok := params[2].([]common.Address)
	if !ok {
		return []common.Address{}, 0, fmt.Errorf("event is not for a transfer multisig key")
	}

	threshold, ok := params[3].(*big.Int)
	if !ok {
		return []common.Address{}, 0, fmt.Errorf("event is not for a transfer multisig key")
	}

	return addresses, uint8(threshold.Uint64()), nil
}

func packEvents(confirmed bool, events []evmTypes.Event) (vote.Vote, error) {
	var v vote.Vote
	if confirmed {
		eventsAny, err := evmTypes.PackEvents(events)
		if err != nil {
			return vote.Vote{}, sdkerrors.Wrap(err, "Pack events failed")
		}
		v.Results = eventsAny
	}
	return v, nil
}
