package evm

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	sdkClient "github.com/cosmos/cosmos-sdk/client"
	sdkFlags "github.com/cosmos/cosmos-sdk/client/flags"

	tmEvents "github.com/axelarnetwork/tm-events/events"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	tmLog "github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcaster/types"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/parse"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// Smart contract event signatures
var (
	ERC20TransferSig        = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	ERC20TokenDeploymentSig = crypto.Keccak256Hash([]byte("TokenDeployed(string,address)"))
	TransferOwnershipSig    = crypto.Keccak256Hash([]byte("OwnershipTransferred(address,address)"))
	TransferOperatorshipSig = crypto.Keccak256Hash([]byte("OperatorshipTransferred(address,address)"))
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
	refundableMsg := axelarnet.NewRefundMsgRequest(mgr.cliCtx.FromAddress, msg)
	mgr.logger.Debug(fmt.Sprintf("broadcasting vote %v for poll %s", msg.Confirmed, pollKey.String()))
	_, err = mgr.broadcaster.Broadcast(mgr.cliCtx.WithBroadcastMode(sdkFlags.BroadcastBlock), refundableMsg)
	return err
}

// ProcessDepositConfirmation votes on the correctness of an EVM chain token deposit
func (mgr Mgr) ProcessDepositConfirmation(e tmEvents.Event) (err error) {
	chain, txID, amount, burnAddr, tokenAddr, confHeight, pollKey, err := parseDepositConfirmationParams(mgr.cdc, e.Attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "EVM deposit confirmation failed")
	}

	rpc, found := mgr.rpcs[strings.ToLower(chain)]
	if !found {
		return sdkerrors.Wrap(err, fmt.Sprintf("Unable to find an RPC for chain '%s'", chain))
	}

	confirmed := mgr.validate(rpc, txID, confHeight, func(txReceipt *geth.Receipt) bool {
		err = confirmERC20Deposit(txReceipt, amount, burnAddr, tokenAddr)
		if err != nil {
			mgr.logger.Debug(sdkerrors.Wrap(err, "deposit confirmation failed").Error())
			return false
		}
		return true
	})

	msg := evmTypes.NewVoteConfirmDepositRequest(mgr.cliCtx.FromAddress, chain, pollKey, txID, evmTypes.Address(burnAddr), confirmed)
	refundableMsg := axelarnet.NewRefundMsgRequest(mgr.cliCtx.FromAddress, msg)
	mgr.logger.Debug(fmt.Sprintf("broadcasting vote %v for poll %s", msg.Confirmed, pollKey.String()))
	_, err = mgr.broadcaster.Broadcast(mgr.cliCtx.WithBroadcastMode(sdkFlags.BroadcastBlock), refundableMsg)
	return err
}

// ProcessTokenConfirmation votes on the correctness of an EVM chain token deployment
func (mgr Mgr) ProcessTokenConfirmation(e tmEvents.Event) error {
	chain, txID, gatewayAddr, tokenAddr, asset, symbol, confHeight, pollKey, err := parseTokenConfirmationParams(mgr.cdc, e.Attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "EVM token deployment confirmation failed")
	}

	rpc, found := mgr.rpcs[strings.ToLower(chain)]
	if !found {
		return sdkerrors.Wrap(err, fmt.Sprintf("Unable to find an RPC for chain '%s'", chain))
	}

	confirmed := mgr.validate(rpc, txID, confHeight, func(txReceipt *geth.Receipt) bool {
		err = confirmERC20TokenDeployment(txReceipt, symbol, gatewayAddr, tokenAddr)
		if err != nil {
			mgr.logger.Debug(sdkerrors.Wrap(err, "token confirmation failed").Error())
			return false
		}
		return true
	})

	msg := evmTypes.NewVoteConfirmTokenRequest(mgr.cliCtx.FromAddress, chain, asset, pollKey, txID, confirmed)
	refundableMsg := axelarnet.NewRefundMsgRequest(mgr.cliCtx.FromAddress, msg)
	mgr.logger.Debug(fmt.Sprintf("broadcasting vote %v for poll %s", msg.Confirmed, pollKey.String()))
	_, err = mgr.broadcaster.Broadcast(mgr.cliCtx.WithBroadcastMode(sdkFlags.BroadcastBlock), refundableMsg)
	return err
}

// ProcessTransferOwnershipConfirmation votes on the correctness of an EVM chain transfer ownership
func (mgr Mgr) ProcessTransferOwnershipConfirmation(e tmEvents.Event) (err error) {
	chain, txID, transferKeyType, gatewayAddr, newOwnerAddr, confHeight, pollKey, err := parseTransferOwnershipConfirmationParams(mgr.cdc, e.Attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "EVM deposit confirmation failed")
	}

	rpc, found := mgr.rpcs[strings.ToLower(chain)]
	if !found {
		return sdkerrors.Wrap(err, fmt.Sprintf("Unable to find an RPC for chain '%s'", chain))
	}

	confirmed := mgr.validate(rpc, txID, confHeight, func(txReceipt *geth.Receipt) bool {
		if err = confirmTransferKey(txReceipt, transferKeyType, gatewayAddr, newOwnerAddr); err != nil {
			mgr.logger.Debug(sdkerrors.Wrap(err, "transfer ownership confirmation failed").Error())
			return false
		}

		return true
	})

	msg := evmTypes.NewVoteConfirmTransferKeyRequest(mgr.cliCtx.FromAddress, chain, pollKey, txID, transferKeyType, evmTypes.Address(newOwnerAddr), confirmed)
	refundableMsg := axelarnet.NewRefundMsgRequest(mgr.cliCtx.FromAddress, msg)
	mgr.logger.Debug(fmt.Sprintf("broadcasting vote %v for poll %s", msg.Confirmed, pollKey.String()))
	_, err = mgr.broadcaster.Broadcast(mgr.cliCtx.WithBroadcastMode(sdkFlags.BroadcastBlock), refundableMsg)
	return err
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
	amount sdk.Uint,
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
		{Key: evmTypes.AttributeKeyAmount, Map: func(s string) (interface{}, error) { return sdk.ParseUint(s) }},
		{Key: evmTypes.AttributeKeyBurnAddress, Map: func(s string) (interface{}, error) {
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
		return "", [32]byte{}, sdk.Uint{}, [20]byte{}, [20]byte{}, 0, vote.PollKey{}, err
	}

	return results[0].(string),
		results[1].(common.Hash),
		results[2].(sdk.Uint),
		results[3].(common.Address),
		results[4].(common.Address),
		results[5].(uint64),
		results[6].(vote.PollKey),
		nil
}

func parseTokenConfirmationParams(cdc *codec.LegacyAmino, attributes map[string]string) (
	chain string,
	txID common.Hash,
	gatewayAddr, tokenAddr common.Address,
	asset string,
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
		{Key: evmTypes.AttributeKeyAsset, Map: parse.IdentityMap},
		{Key: evmTypes.AttributeKeySymbol, Map: parse.IdentityMap},
		{Key: evmTypes.AttributeKeyConfHeight, Map: func(s string) (interface{}, error) { return strconv.ParseUint(s, 10, 64) }},
		{Key: evmTypes.AttributeKeyPoll, Map: func(s string) (interface{}, error) {
			cdc.MustUnmarshalJSON([]byte(s), &pollKey)
			return pollKey, nil
		}},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		return "", [32]byte{}, [20]byte{}, [20]byte{}, "", "", 0, vote.PollKey{}, err
	}

	return results[0].(string),
		results[1].(common.Hash),
		results[2].(common.Address),
		results[3].(common.Address),
		results[4].(string),
		results[5].(string),
		results[6].(uint64),
		results[7].(vote.PollKey),
		nil
}

func parseTransferOwnershipConfirmationParams(cdc *codec.LegacyAmino, attributes map[string]string) (
	chain string,
	txID common.Hash,
	transferKeyType evmTypes.TransferKeyType,
	gatewayAddr, newOwnerAddr common.Address,
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
		{Key: evmTypes.AttributeKeyGatewayAddress, Map: func(s string) (interface{}, error) {
			return common.HexToAddress(s), nil
		}},
		{Key: evmTypes.AttributeKeyAddress, Map: func(s string) (interface{}, error) {
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
		return "", [32]byte{}, 0, [20]byte{}, [20]byte{}, 0, vote.PollKey{}, err
	}

	return results[0].(string),
		results[1].(common.Hash),
		results[2].(evmTypes.TransferKeyType),
		results[3].(common.Address),
		results[4].(common.Address),
		results[5].(uint64),
		results[6].(vote.PollKey),
		nil
}

func (mgr Mgr) validate(rpc rpc.Client, txID common.Hash, confHeight uint64, validateLogs func(txReceipt *geth.Receipt) bool) bool {
	blockNumber, err := rpc.BlockNumber(context.Background())
	if err != nil {
		mgr.logger.Debug(sdkerrors.Wrap(err, "checking block number failed").Error())
		// TODO: this error is not the caller's fault, so we should implement a retry here instead of voting against
		return false
	}
	txReceipt, err := rpc.TransactionReceipt(context.Background(), txID)
	if err != nil {
		mgr.logger.Debug(sdkerrors.Wrap(err, "transaction receipt call failed").Error())
		return false
	}

	if !isTxFinalized(txReceipt, blockNumber, confHeight) {
		mgr.logger.Debug(fmt.Sprintf("transaction %s does not have enough confirmations yet", txReceipt.TxHash.String()))
		return false
	}
	if !isTxSuccessful(txReceipt) {
		mgr.logger.Debug(fmt.Sprintf("transaction %s failed", txReceipt.TxHash.String()))
		return false
	}
	return validateLogs(txReceipt)
}

func confirmERC20Deposit(txReceipt *geth.Receipt, amount sdk.Uint, burnAddr common.Address, tokenAddr common.Address) error {
	actualAmount := sdk.ZeroUint()
	for _, log := range txReceipt.Logs {
		/* Event is not related to the token */
		if log.Address != tokenAddr {
			continue
		}

		to, transferAmount, err := decodeERC20TransferEvent(log)
		/* Event is not an ERC20 transfer */
		if err != nil {
			continue
		}

		/* Transfer isn't sent to burner */
		if to != burnAddr {
			continue
		}

		actualAmount = actualAmount.Add(transferAmount)
	}

	if !actualAmount.Equal(amount) {
		return fmt.Errorf("given deposit amount: %d, actual amount: %d", amount.Uint64(), actualAmount.Uint64())
	}

	return nil
}

func confirmERC20TokenDeployment(txReceipt *geth.Receipt, expectedSymbol string, gatewayAddr, expectedAddr common.Address) error {
	for _, log := range txReceipt.Logs {
		// Event is not emitted by the axelar gateway
		if log.Address != gatewayAddr {
			continue
		}

		// Event is not for a ERC20 token deployment
		symbol, tokenAddr, err := decodeERC20TokenDeploymentEvent(log)
		if err != nil {
			continue
		}

		// Symbol does not match
		if symbol != expectedSymbol {
			continue
		}

		// token address does not match
		if tokenAddr != expectedAddr {
			continue
		}

		// if we reach this point, it means that the log matches what we want to verify,
		// so the function can return with no error
		return nil
	}

	return fmt.Errorf("failed to confirm token deployment for symbol '%s' at contract address '%s'", expectedSymbol, expectedAddr.String())
}

func confirmTransferKey(txReceipt *geth.Receipt, transferKeyType evmTypes.TransferKeyType, gatewayAddr, expectedNewAddr common.Address) (err error) {
	for i := len(txReceipt.Logs) - 1; i >= 0; i-- {
		log := txReceipt.Logs[i]
		// Event is not emitted by the axelar gateway
		if log.Address != gatewayAddr {
			continue
		}

		var actualNewAddr common.Address
		// There might be several transfer ownership/operatorship event. Only interest in the last one.
		switch transferKeyType {
		case evmTypes.Ownership:
			actualNewAddr, err = decodeTransferOwnershipEvent(log)
		case evmTypes.Operatorship:
			actualNewAddr, err = decodeTransferOperatorshipEvent(log)
		default:
			return fmt.Errorf("invalid transfer key type")
		}

		if err != nil {
			continue
		}

		// New addr does not match
		if actualNewAddr != expectedNewAddr {
			return fmt.Errorf("failed to confirm %s for new address '%s' at contract address '%s'", transferKeyType.SimpleString(), expectedNewAddr.String(), gatewayAddr.String())
		}

		// if we reach this point, it means that the log matches what we want to verify,
		// so the function can return with no error
		return nil
	}

	return fmt.Errorf("failed to confirm %s for new address '%s' at contract address '%s'", transferKeyType.SimpleString(), expectedNewAddr.String(), gatewayAddr.String())
}

func isTxFinalized(txReceipt *geth.Receipt, blockNumber uint64, confirmationHeight uint64) bool {
	return blockNumber-txReceipt.BlockNumber.Uint64()+1 >= confirmationHeight
}

func isTxSuccessful(txReceipt *geth.Receipt) bool {
	return txReceipt.Status == 1
}

func decodeERC20TransferEvent(log *geth.Log) (common.Address, sdk.Uint, error) {

	if len(log.Topics) != 3 || log.Topics[0] != ERC20TransferSig {
		return common.Address{}, sdk.Uint{}, fmt.Errorf("log is not an ERC20 transfer")
	}

	to := common.BytesToAddress(log.Topics[2][:])
	amount := new(big.Int)
	amount.SetBytes(log.Data[:32])

	return to, sdk.NewUintFromBigInt(amount), nil
}

func decodeERC20TokenDeploymentEvent(log *geth.Log) (string, common.Address, error) {
	if len(log.Topics) != 1 || log.Topics[0] != ERC20TokenDeploymentSig {
		return "", common.Address{}, fmt.Errorf("event is not for an ERC20 token deployment")
	}

	// Decode the data field
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return "", common.Address{}, err
	}
	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return "", common.Address{}, err
	}
	packedArgs := abi.Arguments{{Type: stringType}, {Type: addressType}}
	args, err := packedArgs.Unpack(log.Data)
	if err != nil {
		return "", common.Address{}, err
	}

	return args[0].(string), args[1].(common.Address), nil
}

func decodeTransferOwnershipEvent(log *geth.Log) (common.Address, error) {
	if len(log.Topics) != 3 || log.Topics[0] != TransferOwnershipSig {
		return common.Address{}, fmt.Errorf("event is not for a transfer owernship")
	}

	newOwnerAddr := common.BytesToAddress(log.Topics[2][:])

	return newOwnerAddr, nil
}

func decodeTransferOperatorshipEvent(log *geth.Log) (common.Address, error) {
	if len(log.Topics) != 3 || log.Topics[0] != TransferOperatorshipSig {
		return common.Address{}, fmt.Errorf("event is not for a transfer operatorship")
	}

	newOperatorAddr := common.BytesToAddress(log.Topics[2][:])

	return newOperatorAddr, nil
}
