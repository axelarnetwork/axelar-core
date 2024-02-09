package evm

import (
	"bytes"
	"context"
	goerrors "errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/axelarnetwork/axelar-core/sdk-utils/broadcast"
	"github.com/axelarnetwork/axelar-core/utils/errors"
	"github.com/axelarnetwork/axelar-core/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/log"
	rs "github.com/axelarnetwork/utils/monads/results"
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

// NotFinalizedError contains the block height of the transaction that is not finalized yet
type NotFinalizedError struct {
	BlockHeight uint64
}

func (e NotFinalizedError) Error() string {
	return "not finalized"
}

// FailedTransactionError contains the block height of the transaction that failed on the source chain
type FailedTransactionError struct {
	BlockHeight uint64
}

func (e FailedTransactionError) Error() string {
	return "failed on source chain"
}

// NotFoundError is a type-safe wrapper around ethereum.NotFound
type NotFoundError struct{}

func (e NotFoundError) Error() string {
	return ethereum.NotFound.Error()
}

// Mgr manages all communication with Ethereum
type Mgr struct {
	rpcs                      map[string]rpc.Client
	broadcaster               broadcast.Broadcaster
	validator                 sdk.ValAddress
	proxy                     sdk.AccAddress
	latestFinalizedBlockCache LatestFinalizedBlockCache
}

// NewMgr returns a new Mgr instance
func NewMgr(rpcs map[string]rpc.Client, broadcaster broadcast.Broadcaster, valAddr sdk.ValAddress, proxy sdk.AccAddress, latestFinalizedBlockCache LatestFinalizedBlockCache) *Mgr {
	return &Mgr{
		rpcs:                      rpcs,
		proxy:                     proxy,
		broadcaster:               broadcaster,
		validator:                 valAddr,
		latestFinalizedBlockCache: latestFinalizedBlockCache,
	}
}

func (mgr Mgr) logger(keyvals ...any) log.Logger {
	keyvals = append([]any{"listener", "evm"}, keyvals...)
	return log.WithKeyVals(keyvals...)
}

// ProcessNewChain notifies the operator that vald needs to be restarted/udpated for a new chain
func (mgr Mgr) ProcessNewChain(event *types.ChainAdded) (err error) {
	mgr.logger().Info(fmt.Sprintf("VALD needs to be updated and restarted for new chain %s", event.Chain.String()))
	return nil
}

// ProcessDepositConfirmation votes on the correctness of an EVM chain token deposit
func (mgr Mgr) ProcessDepositConfirmation(event *types.ConfirmDepositStarted) error {
	vc := voteContext{
		Participants:       event.Participants,
		PollMappings:       []types.PollMapping{{PollID: event.PollID, TxID: event.TxID}},
		Chain:              event.Chain,
		ConfirmationHeight: event.ConfirmationHeight,
		PollType:           "token deposit",
	}
	return mgr.vote(vc, func(logs []*geth.Log) []types.Event {
		return mgr.processTokenDepositLogs(logs, event)
	})
}

func (mgr Mgr) processTokenDepositLogs(logs []*geth.Log, event *types.ConfirmDepositStarted) []types.Event {
	var events []types.Event
	for i, log := range logs {
		if log.Topics[0] != ERC20TransferSig {
			continue
		}

		if !bytes.Equal(event.TokenAddress.Bytes(), log.Address.Bytes()) {
			continue
		}

		erc20Event, err := DecodeERC20TransferEvent(log)
		if err != nil {
			mgr.logger().Debug(sdkerrors.Wrap(err, "decode event Transfer failed").Error())
			continue
		}

		if erc20Event.To != event.DepositAddress {
			continue
		}

		if err := erc20Event.ValidateBasic(); err != nil {
			mgr.logger().Debug(sdkerrors.Wrap(err, "invalid event Transfer").Error())
			continue
		}

		events = append(events, types.Event{
			Chain: event.Chain,
			TxID:  event.TxID,
			Index: uint64(i),
			Event: &types.Event_Transfer{
				Transfer: &erc20Event,
			},
		})
	}
	return events
}

// ProcessTokenConfirmation votes on the correctness of an EVM chain token deployment
func (mgr Mgr) ProcessTokenConfirmation(event *types.ConfirmTokenStarted) error {
	vc := voteContext{
		Participants:       event.Participants,
		PollMappings:       []types.PollMapping{{PollID: event.PollID, TxID: event.TxID}},
		Chain:              event.Chain,
		ConfirmationHeight: event.ConfirmationHeight,
		PollType:           "token deployment confirmation",
	}
	return mgr.vote(vc, func(logs []*geth.Log) []types.Event {
		return mgr.processTokenConfirmationLogs(logs, event)
	})
}

func (mgr Mgr) processTokenConfirmationLogs(logs []*geth.Log, event *types.ConfirmTokenStarted) []types.Event {
	var events []types.Event
	for i, log := range logs {
		if log.Topics[0] != ERC20TokenDeploymentSig {
			continue
		}

		if !bytes.Equal(event.GatewayAddress.Bytes(), log.Address.Bytes()) {
			continue
		}

		erc20Event, err := DecodeERC20TokenDeploymentEvent(log)
		if err != nil {
			mgr.logger().Debug(sdkerrors.Wrap(err, "decode event TokenDeployed failed").Error())
			continue
		}

		if erc20Event.TokenAddress != event.TokenAddress || erc20Event.Symbol != event.TokenDetails.Symbol {
			continue
		}

		if err := erc20Event.ValidateBasic(); err != nil {
			mgr.logger().Debug(sdkerrors.Wrap(err, "invalid event ERC20TokenDeployment").Error())
			continue
		}

		events = append(events, types.Event{
			Chain: event.Chain,
			TxID:  event.TxID,
			Index: uint64(i),
			Event: &types.Event_TokenDeployed{
				TokenDeployed: &erc20Event,
			},
		})
		break
	}
	return events
}

// ProcessTransferKeyConfirmation votes on the correctness of an EVM chain key transfer
func (mgr Mgr) ProcessTransferKeyConfirmation(event *types.ConfirmKeyTransferStarted) error {
	vc := voteContext{
		Participants:       event.Participants,
		PollMappings:       []types.PollMapping{{PollID: event.PollID, TxID: event.TxID}},
		Chain:              event.Chain,
		ConfirmationHeight: event.ConfirmationHeight,
		PollType:           "key transfer confirmation",
	}
	return mgr.vote(vc, func(logs []*geth.Log) []types.Event {
		return mgr.processKeyTransferLogs(logs, event)
	})
}

func (mgr Mgr) processKeyTransferLogs(logs []*geth.Log, event *types.ConfirmKeyTransferStarted) []types.Event {

	var events []types.Event

	for i := len(logs) - 1; i >= 0; i-- {
		txlog := logs[i]

		if txlog.Topics[0] != MultisigTransferOperatorshipSig {
			continue
		}

		// Event is not emitted by the axelar gateway
		if txlog.Address != common.Address(event.GatewayAddress) {
			continue
		}

		transferOperatorshipEvent, err := DecodeMultisigOperatorshipTransferredEvent(txlog)
		if err != nil {
			mgr.logger().Debug(sdkerrors.Wrap(err, "failed decoding operatorship transferred event").Error())
			continue
		}

		if err := transferOperatorshipEvent.ValidateBasic(); err != nil {
			mgr.logger().Debug(sdkerrors.Wrap(err, "invalid event MultisigTransferOperatorship").Error())
			continue
		}

		events = append(events, types.Event{
			Chain: event.Chain,
			TxID:  event.TxID,
			Index: uint64(i),
			Event: &types.Event_MultisigOperatorshipTransferred{
				MultisigOperatorshipTransferred: &transferOperatorshipEvent,
			}})
		break
	}
	return events
}

// ProcessGatewayTxConfirmation votes on the correctness of an EVM chain gateway's transactions
func (mgr Mgr) ProcessGatewayTxConfirmation(event *types.ConfirmGatewayTxStarted) error {
	mappedEvent := &types.ConfirmGatewayTxsStarted{
		PollMappings:       []types.PollMapping{{PollID: event.PollID, TxID: event.TxID}},
		Chain:              event.Chain,
		GatewayAddress:     event.GatewayAddress,
		ConfirmationHeight: event.ConfirmationHeight,
		Participants:       event.Participants,
	}
	return mgr.ProcessGatewayTxsConfirmation(mappedEvent)
}

// ProcessGatewayTxsConfirmation votes on the correctness of an EVM chain multiple gateway transactions
func (mgr Mgr) ProcessGatewayTxsConfirmation(event *types.ConfirmGatewayTxsStarted) error {
	vc := voteContext{
		Participants:       event.Participants,
		PollMappings:       event.PollMappings,
		Chain:              event.Chain,
		ConfirmationHeight: event.ConfirmationHeight,
		PollType:           "gateway txs confirmation",
	}
	return mgr.vote(vc, func(logs []*geth.Log) []types.Event {
		return mgr.processGatewayTxLogs(event.Chain, event.GatewayAddress, logs)
	})
}

type voteContext struct {
	PollType           string
	Participants       []sdk.ValAddress
	PollMappings       []types.PollMapping
	Chain              nexus.ChainName
	ConfirmationHeight uint64
}

func (mgr Mgr) vote(voteContext voteContext, logsToVotes func(logs []*geth.Log) []types.Event) error {
	if !mgr.isParticipantOf(voteContext.Participants) {
		pollIDs := slices.Map(voteContext.PollMappings, func(m types.PollMapping) vote.PollID { return m.PollID })
		mgr.logger("poll_ids", pollIDs).Debug(fmt.Sprintf("ignoring %s poll: not a participant", voteContext.PollType))
		return nil
	}

	txIDs := slices.Map(voteContext.PollMappings, func(poll types.PollMapping) common.Hash { return common.Hash(poll.TxID) })
	txReceipts, err := mgr.GetTxReceiptsIfFinalized(voteContext.Chain, txIDs, voteContext.ConfirmationHeight)
	if err != nil {
		return err
	}

	var votes []sdk.Msg
	for i, result := range txReceipts {
		pollID := voteContext.PollMappings[i].PollID
		txID := voteContext.PollMappings[i].TxID

		logger := mgr.logger("chain", voteContext.Chain, "poll_id", pollID.String(), "tx_id", txID.Hex())

		// only broadcast empty votes if the tx is not found or not finalized
		switch err := result.Err().(type) {
		case nil:
			events := logsToVotes(result.Ok().Logs)
			logger.Infof("broadcasting vote %v", events)
			votes = append(votes, voteTypes.NewVoteRequest(mgr.proxy, pollID, types.NewVoteEvents(voteContext.Chain, events...)))
		case NotFinalizedError:
			logger.Debug(fmt.Sprintf("transaction %s in block %v not finalized", txID.Hex(), err.BlockHeight))
			logger.Infof("broadcasting empty vote due to error: %s", result.Err().Error())
			votes = append(votes, voteTypes.NewVoteRequest(mgr.proxy, pollID, types.NewVoteEvents(voteContext.Chain)))
		case FailedTransactionError:
			logger.Debug(fmt.Sprintf("transaction %s in block %v has failed status", txID.Hex(), err.BlockHeight))
			logger.Infof("broadcasting empty vote due to error: %s", result.Err().Error())
			votes = append(votes, voteTypes.NewVoteRequest(mgr.proxy, pollID, types.NewVoteEvents(voteContext.Chain)))
		case NotFoundError:
			logger.Debug(fmt.Sprintf("transaction receipt %s not found", txID.Hex()))
			logger.Infof("broadcasting empty vote due to error: %s", result.Err().Error())
			votes = append(votes, voteTypes.NewVoteRequest(mgr.proxy, pollID, types.NewVoteEvents(voteContext.Chain)))
		default:
			logger.Errorf("failed to get tx receipt: %s", result.Err().Error())
		}

	}

	_, err = mgr.broadcaster.Broadcast(context.TODO(), votes...)
	return err
}

func DecodeEventTokenSent(log *geth.Log) (types.EventTokenSent, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return types.EventTokenSent{}, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return types.EventTokenSent{}, err
	}

	arguments := abi.Arguments{
		{Type: stringType},
		{Type: stringType},
		{Type: stringType},
		{Type: uint256Type},
	}
	params, err := types.StrictDecode(arguments, log.Data)
	if err != nil {
		return types.EventTokenSent{}, err
	}

	return types.EventTokenSent{
		Sender:             types.Address(common.BytesToAddress(log.Topics[1].Bytes())),
		DestinationChain:   nexus.ChainName(params[0].(string)),
		DestinationAddress: params[1].(string),
		Symbol:             params[2].(string),
		Amount:             sdk.NewUintFromBigInt(params[3].(*big.Int)),
	}, nil
}

func DecodeEventContractCall(log *geth.Log) (types.EventContractCall, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return types.EventContractCall{}, err
	}

	bytesType, err := abi.NewType("bytes", "bytes", nil)
	if err != nil {
		return types.EventContractCall{}, err
	}

	arguments := abi.Arguments{
		{Type: stringType},
		{Type: stringType},
		{Type: bytesType},
	}
	params, err := types.StrictDecode(arguments, log.Data)
	if err != nil {
		return types.EventContractCall{}, err
	}

	return types.EventContractCall{
		Sender:           types.Address(common.BytesToAddress(log.Topics[1].Bytes())),
		DestinationChain: nexus.ChainName(params[0].(string)),
		ContractAddress:  params[1].(string),
		PayloadHash:      types.Hash(common.BytesToHash(log.Topics[2].Bytes())),
	}, nil
}

func DecodeEventContractCallWithToken(log *geth.Log) (types.EventContractCallWithToken, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return types.EventContractCallWithToken{}, err
	}

	bytesType, err := abi.NewType("bytes", "bytes", nil)
	if err != nil {
		return types.EventContractCallWithToken{}, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return types.EventContractCallWithToken{}, err
	}

	arguments := abi.Arguments{
		{Type: stringType},
		{Type: stringType},
		{Type: bytesType},
		{Type: stringType},
		{Type: uint256Type},
	}
	params, err := types.StrictDecode(arguments, log.Data)
	if err != nil {
		return types.EventContractCallWithToken{}, err
	}

	return types.EventContractCallWithToken{
		Sender:           types.Address(common.BytesToAddress(log.Topics[1].Bytes())),
		DestinationChain: nexus.ChainName(params[0].(string)),
		ContractAddress:  params[1].(string),
		PayloadHash:      types.Hash(common.BytesToHash(log.Topics[2].Bytes())),
		Symbol:           params[3].(string),
		Amount:           sdk.NewUintFromBigInt(params[4].(*big.Int)),
	}, nil
}

func (mgr Mgr) isTxReceiptFinalized(chain nexus.ChainName, txReceipt *geth.Receipt, confHeight uint64) (bool, error) {
	client, ok := mgr.rpcs[strings.ToLower(chain.String())]
	if !ok {
		return false, fmt.Errorf("rpc client not found for chain %s", chain.String())
	}

	if mgr.latestFinalizedBlockCache.Get(chain).Cmp(txReceipt.BlockNumber) >= 0 {
		return true, nil
	}

	latestFinalizedBlockNumber, err := client.LatestFinalizedBlockNumber(context.Background(), confHeight)
	if err != nil {
		return false, err
	}

	mgr.latestFinalizedBlockCache.Set(chain, latestFinalizedBlockNumber)

	if latestFinalizedBlockNumber.Cmp(txReceipt.BlockNumber) < 0 {
		return false, nil
	}

	return true, nil
}

// GetTxReceiptsIfFinalized retrieves receipts for provided transaction IDs, only if they're successful and finalized.
func (mgr Mgr) GetTxReceiptsIfFinalized(chain nexus.ChainName, txIDs []common.Hash, confHeight uint64) ([]rs.Result[*geth.Receipt], error) {
	client, ok := mgr.rpcs[strings.ToLower(chain.String())]
	if !ok {
		return nil, fmt.Errorf("rpc client not found for chain %s", chain.String())
	}

	results, err := client.TransactionReceipts(context.Background(), txIDs)
	if err != nil {
		return nil, sdkerrors.Wrapf(errors.With(err, "chain", chain.String(), "tx_ids", txIDs),
			"cannot get transaction receipts")
	}

	isFound := func(res rs.Result[*geth.Receipt]) rs.Result[*geth.Receipt] {
		if goerrors.Is(res.Err(), ethereum.NotFound) {
			return rs.FromErr[*geth.Receipt](NotFoundError{})
		}
		return res
	}

	isFinalized := func(receipt *geth.Receipt) rs.Result[*geth.Receipt] {
		if receipt.Status != geth.ReceiptStatusSuccessful {
			return rs.FromErr[*geth.Receipt](FailedTransactionError{BlockHeight: receipt.BlockNumber.Uint64()})
		}

		isFinalized, err := mgr.isTxReceiptFinalized(chain, receipt, confHeight)
		if err != nil {
			return rs.FromErr[*geth.Receipt](sdkerrors.Wrapf(errors.With(err, "chain", chain.String()),
				"cannot determine if the transaction %s is finalized", receipt.TxHash.Hex()),
			)
		}

		if !isFinalized {
			return rs.FromErr[*geth.Receipt](NotFinalizedError{BlockHeight: receipt.BlockNumber.Uint64()})
		}

		return rs.FromOk(receipt)
	}

	return slices.Map(results, func(r rpc.Result) rs.Result[*geth.Receipt] {
		res := isFound(rs.Result[*geth.Receipt](r))
		return rs.Pipe(res, isFinalized)
	}), nil
}

func DecodeERC20TransferEvent(log *geth.Log) (types.EventTransfer, error) {
	if len(log.Topics) != 3 || log.Topics[0] != ERC20TransferSig {
		return types.EventTransfer{}, fmt.Errorf("log is not an ERC20 transfer")
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return types.EventTransfer{}, err
	}

	to := common.BytesToAddress(log.Topics[2][:])

	arguments := abi.Arguments{
		{Type: uint256Type},
	}

	params, err := arguments.Unpack(log.Data)
	if err != nil {
		return types.EventTransfer{}, err
	}

	return types.EventTransfer{
		To:     types.Address(to),
		Amount: sdk.NewUintFromBigInt(params[0].(*big.Int)),
	}, nil
}

func DecodeERC20TokenDeploymentEvent(log *geth.Log) (types.EventTokenDeployed, error) {
	if len(log.Topics) != 1 || log.Topics[0] != ERC20TokenDeploymentSig {
		return types.EventTokenDeployed{}, fmt.Errorf("event is not for an ERC20 token deployment")
	}

	// Decode the data field
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return types.EventTokenDeployed{}, err
	}
	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return types.EventTokenDeployed{}, err
	}

	arguments := abi.Arguments{{Type: stringType}, {Type: addressType}}
	params, err := types.StrictDecode(arguments, log.Data)
	if err != nil {
		return types.EventTokenDeployed{}, err
	}

	return types.EventTokenDeployed{
		Symbol:       params[0].(string),
		TokenAddress: types.Address(params[1].(common.Address)),
	}, nil
}

func DecodeMultisigOperatorshipTransferredEvent(log *geth.Log) (types.EventMultisigOperatorshipTransferred, error) {
	if len(log.Topics) != 1 || log.Topics[0] != MultisigTransferOperatorshipSig {
		return types.EventMultisigOperatorshipTransferred{}, fmt.Errorf("event is not OperatorshipTransferred")
	}

	newAddresses, newWeights, newThreshold, err := unpackMultisigTransferKeyEvent(log)
	if err != nil {
		return types.EventMultisigOperatorshipTransferred{}, err
	}

	event := types.EventMultisigOperatorshipTransferred{
		NewOperators: slices.Map(newAddresses, func(addr common.Address) types.Address { return types.Address(addr) }),
		NewWeights:   slices.Map(newWeights, sdk.NewUintFromBigInt),
		NewThreshold: sdk.NewUintFromBigInt(newThreshold),
	}

	return event, nil
}

func unpackMultisigTransferKeyEvent(log *geth.Log) ([]common.Address, []*big.Int, *big.Int, error) {
	bytesType := funcs.Must(abi.NewType("bytes", "bytes", nil))
	newOperatorsData, err := types.StrictDecode(abi.Arguments{{Type: bytesType}}, log.Data)
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

// extract receipt processing from ProcessGatewayTxConfirmation, so that it can be used in ProcessGatewayTxsConfirmation
func (mgr Mgr) processGatewayTxLogs(chain nexus.ChainName, gatewayAddress types.Address, logs []*geth.Log) []types.Event {
	var events []types.Event
	for i, txlog := range logs {
		if !bytes.Equal(gatewayAddress.Bytes(), txlog.Address.Bytes()) {
			continue
		}

		switch txlog.Topics[0] {
		case ContractCallSig:
			gatewayEvent, err := DecodeEventContractCall(txlog)
			if err != nil {
				mgr.logger().Debug(sdkerrors.Wrap(err, "decode event ContractCall failed").Error())
				continue
			}

			if err := gatewayEvent.ValidateBasic(); err != nil {
				mgr.logger().Debug(sdkerrors.Wrap(err, "invalid event ContractCall").Error())
				continue
			}

			events = append(events, types.Event{
				Chain: chain,
				TxID:  types.Hash(txlog.TxHash),
				Index: uint64(i),
				Event: &types.Event_ContractCall{
					ContractCall: &gatewayEvent,
				},
			})
		case ContractCallWithTokenSig:
			gatewayEvent, err := DecodeEventContractCallWithToken(txlog)
			if err != nil {
				mgr.logger().Debug(sdkerrors.Wrap(err, "decode event ContractCallWithToken failed").Error())
				continue
			}

			if err := gatewayEvent.ValidateBasic(); err != nil {
				mgr.logger().Debug(sdkerrors.Wrap(err, "invalid event ContractCallWithToken").Error())
				continue
			}

			events = append(events, types.Event{
				Chain: chain,
				TxID:  types.Hash(txlog.TxHash),
				Index: uint64(i),
				Event: &types.Event_ContractCallWithToken{
					ContractCallWithToken: &gatewayEvent,
				},
			})
		case TokenSentSig:
			gatewayEvent, err := DecodeEventTokenSent(txlog)
			if err != nil {
				mgr.logger().Debug(sdkerrors.Wrap(err, "decode event TokenSent failed").Error())
			}

			if err := gatewayEvent.ValidateBasic(); err != nil {
				mgr.logger().Debug(sdkerrors.Wrap(err, "invalid event TokenSent").Error())
				continue
			}

			events = append(events, types.Event{
				Chain: chain,
				TxID:  types.Hash(txlog.TxHash),
				Index: uint64(i),
				Event: &types.Event_TokenSent{
					TokenSent: &gatewayEvent,
				},
			})
		default:
		}
	}

	return events
}

// isParticipantOf checks if the validator is in the poll participants list
func (mgr Mgr) isParticipantOf(participants []sdk.ValAddress) bool {
	return slices.Any(participants, func(v sdk.ValAddress) bool { return v.Equals(mgr.validator) })
}
