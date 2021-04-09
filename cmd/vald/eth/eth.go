package eth

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	tmLog "github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/cmd/vald/broadcast/types"
	ethRPC "github.com/axelarnetwork/axelar-core/cmd/vald/eth/rpc"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	ethTypes "github.com/axelarnetwork/axelar-core/x/ethereum/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// Smart contract event signatures
var (
	ERC20TransferSig    = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	ERC20TokenDeploySig = crypto.Keccak256Hash([]byte("TokenDeployed(string,address)"))
)

// Mgr manages all communication with Ethereum
type Mgr struct {
	logger      tmLog.Logger
	rpc         ethRPC.Client
	broadcaster types.Broadcaster
	sender      sdk.AccAddress
}

// NewMgr returns a new Mgr instance
func NewMgr(rpc ethRPC.Client, broadcaster types.Broadcaster, sender sdk.AccAddress, logger tmLog.Logger) *Mgr {
	return &Mgr{
		rpc:         rpc,
		broadcaster: broadcaster,
		sender:      sender,
		logger:      logger.With("listener", "eth"),
	}
}

// ProccessDepositConfirmation votes on the correctness of an Ethereum token deposit
func (mgr Mgr) ProccessDepositConfirmation(attributes []sdk.Attribute) (err error) {
	txID, amount, burnAddr, tokenAddr, confHeight, poll, err := parseDepositConfirmationParams(attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "Ethereum deposit confirmation failed")
	}

	confirmed := mgr.validate(txID, confHeight, func(txReceipt *geth.Receipt) bool {
		err = confirmERC20Deposit(txReceipt, amount, burnAddr, tokenAddr)
		if err != nil {
			mgr.logger.Debug(sdkerrors.Wrap(err, "deposit confirmation failed").Error())
			return false
		}
		return true
	})

	msg := ethTypes.MsgVoteConfirmDeposit{
		Sender:    mgr.sender,
		Poll:      poll,
		TxID:      txID.Hex(),
		BurnAddr:  burnAddr.Hex(),
		Confirmed: confirmed,
	}
	mgr.logger.Debug(fmt.Sprintf("broadcasting vote %v for poll %s", msg.Confirmed, poll.String()))
	return mgr.broadcaster.Broadcast(msg)
}

// ProccessTokenConfirmation votes on the correctness of an Ethereum token deployment
func (mgr Mgr) ProccessTokenConfirmation(attributes []sdk.Attribute) error {
	txID, gatewayAddr, tokenAddr, symbol, confHeight, poll, err := parseTokenConfirmationParams(attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "Ethereum token deployment confirmation failed")
	}

	confirmed := mgr.validate(txID, confHeight, func(txReceipt *geth.Receipt) bool {
		err = confirmERC20TokenDeploy(txReceipt, symbol, gatewayAddr, tokenAddr)
		if err != nil {
			mgr.logger.Debug(sdkerrors.Wrap(err, "token confirmation failed").Error())
			return false
		}
		return true
	})

	msg := ethTypes.MsgVoteConfirmToken{
		Sender:    mgr.sender,
		Poll:      poll,
		TxID:      txID.Hex(),
		Confirmed: confirmed,
		Symbol:    symbol,
	}
	mgr.logger.Debug(fmt.Sprintf("broadcasting vote %v for poll %s", msg.Confirmed, poll.String()))
	return mgr.broadcaster.Broadcast(msg)
}

func parseDepositConfirmationParams(attributes []sdk.Attribute) (
	txID common.Hash,
	amount sdk.Uint,
	burnAddr, tokenAddr common.Address,
	confHeight uint64,
	poll vote.PollMeta,
	err error,
) {
	var txIDFound, amountFound, burnAddrFound, tokenAddrFound, confHeightFound, pollFound bool
	for _, attribute := range attributes {
		switch attribute.Key {
		case ethTypes.AttributeKeyTxID:
			txID = common.HexToHash(attribute.Value)
			txIDFound = true
		case ethTypes.AttributeKeyAmount:
			amount, err = sdk.ParseUint(attribute.Value)
			if err != nil {
				return [32]byte{}, sdk.Uint{}, [20]byte{}, [20]byte{}, 0, vote.PollMeta{},
					sdkerrors.Wrap(err, "parsing transfer amount failed")
			}
			amountFound = true
		case ethTypes.AttributeKeyBurnAddress:
			burnAddr = common.HexToAddress(attribute.Value)
			burnAddrFound = true
		case ethTypes.AttributeKeyTokenAddress:
			tokenAddr = common.HexToAddress(attribute.Value)
			tokenAddrFound = true
		case ethTypes.AttributeKeyConfHeight:
			confHeight, err = strconv.ParseUint(attribute.Value, 10, 64)
			if err != nil {
				return common.Hash{}, sdk.Uint{}, common.Address{}, common.Address{}, 0, vote.PollMeta{},
					sdkerrors.Wrap(err, "parsing confirmation height failed")
			}
			confHeightFound = true
		case ethTypes.AttributeKeyPoll:
			codec.Cdc.MustUnmarshalJSON([]byte(attribute.Value), &poll)
			pollFound = true
		default:
		}
	}
	if !txIDFound || !amountFound || !burnAddrFound || !tokenAddrFound || !confHeightFound || !pollFound {
		return common.Hash{}, sdk.Uint{}, common.Address{}, common.Address{}, 0, vote.PollMeta{},
			fmt.Errorf("insufficient event attributes")
	}
	return txID, amount, burnAddr, tokenAddr, confHeight, poll, nil
}

func parseTokenConfirmationParams(attributes []sdk.Attribute) (
	txID common.Hash,
	gatewayAddr, tokenAddr common.Address,
	symbol string,
	confHeight uint64,
	poll vote.PollMeta,
	err error,
) {
	var txIDFound, gatewayAddrFound, tokenAddrFound, symbolFound, confHeightFound, pollFound bool
	for _, attribute := range attributes {
		switch attribute.Key {
		case ethTypes.AttributeKeyTxID:
			txID = common.HexToHash(attribute.Value)
			txIDFound = true
		case ethTypes.AttributeKeyGatewayAddress:
			gatewayAddr = common.HexToAddress(attribute.Value)
			gatewayAddrFound = true
		case ethTypes.AttributeKeyTokenAddress:
			tokenAddr = common.HexToAddress(attribute.Value)
			tokenAddrFound = true
		case ethTypes.AttributeKeySymbol:
			symbol = attribute.Value
			symbolFound = true
		case ethTypes.AttributeKeyConfHeight:
			h, err := strconv.Atoi(attribute.Value)
			if err != nil {
				return common.Hash{}, common.Address{}, common.Address{}, "", 0, vote.PollMeta{},
					sdkerrors.Wrap(err, "parsing confirmation height failed")
			}
			confHeight = uint64(h)
			confHeightFound = true
		case btc.AttributeKeyPoll:
			codec.Cdc.MustUnmarshalJSON([]byte(attribute.Value), &poll)
			pollFound = true
		default:
		}
	}
	if !txIDFound || !gatewayAddrFound || !tokenAddrFound || !symbolFound || !confHeightFound || !pollFound {
		return common.Hash{}, common.Address{}, common.Address{}, "", 0, vote.PollMeta{},
			fmt.Errorf("insufficient event attributes")
	}
	return txID, gatewayAddr, tokenAddr, symbol, confHeight, poll, nil
}

func (mgr Mgr) validate(txID common.Hash, confHeight uint64, validateLogs func(txReceipt *geth.Receipt) bool) bool {
	blockNumber, err := mgr.rpc.BlockNumber(context.Background())
	if err != nil {
		mgr.logger.Debug(sdkerrors.Wrap(err, "checking block number failed").Error())
		// TODO: this error is not the caller's fault, so we should implement a retry here instead of voting against
		return false
	}
	txReceipt, err := mgr.rpc.TransactionReceipt(context.Background(), txID)
	if err != nil {
		mgr.logger.Debug(sdkerrors.Wrap(err, "transaction receipt call failed").Error())
		return false
	}

	if !isTxFinalized(txReceipt, blockNumber, confHeight) {
		mgr.logger.Debug(fmt.Sprintf("transaction %s does not have enough confirmations yet", txReceipt.TxHash.String()))
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

func confirmERC20TokenDeploy(txReceipt *geth.Receipt, expectedSymbol string, gatewayAddr, expectedAddr common.Address) error {
	for _, log := range txReceipt.Logs {
		// Event is not emitted by the axelar gateway
		if log.Address != gatewayAddr {
			continue
		}

		// Event is not for a ERC20 token deployment
		symbol, tokenAddr, err := decodeERC20TokenDeployEvent(log)
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

func isTxFinalized(txReceipt *geth.Receipt, blockNumber uint64, confirmationHeight uint64) bool {
	return blockNumber-txReceipt.BlockNumber.Uint64()+1 >= confirmationHeight
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

func decodeERC20TokenDeployEvent(log *geth.Log) (string, common.Address, error) {
	if len(log.Topics) != 1 || log.Topics[0] != ERC20TokenDeploySig {
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
