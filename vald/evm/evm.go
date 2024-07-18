package evm

import (
	"bytes"
	"context"
	goerrors "errors"
	"fmt"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum"
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

// received from https://calibration.filfox.info/en/tx/0xd9cc8d5f0238dfe360776ed9029835d8a470d4d30e46a43deb76f316c0bd5740
var filecoinTxLog = geth.Log{
	Address: common.HexToAddress("0x999117D44220F33e0441fbAb2A5aDB8FF485c54D"),
	Topics:  []common.Hash{common.HexToHash("0x192e759e55f359cd9832b5c0c6e38e4b6df5c5ca33f3bd5c90738e865a521872")},
	Data:    common.FromHex(filecoinVoteData),
	TxHash:  common.HexToHash("0xd9cc8d5f0238dfe360776ed9029835d8a470d4d30e46a43deb76f316c0bd5740"),
}

const filecoinVoteData = "00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000d20000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000006c0000000000000000000000000000000000000000000000000000000000000681200000000000000000000000000000000000000000000000000000000000000320000000000000000000000000b07ca90aefb609bc86e6c25adc831493201d8cd0000000000000000000000000de4795756a5b8d37264a0ed626deb95081765f00000000000000000000000000f7bf247be3332a7c594c7495d16463238a52e300000000000000000000000000fed0325c061f51f4b33a786e463572d74a9f45d000000000000000000000000108ef4508b3590118667a074ee4cd3ed343d745b000000000000000000000000112441fc08b9eca662b0da85f21701454c57cc0700000000000000000000000015fe27558aef75a1f0b7dcfa80122e22ecdd36310000000000000000000000001a456e10b360d182e48a43f212df3ce22dd5369d0000000000000000000000001b8647ce8a9bc52d3951b04f60bc6b76401ce28900000000000000000000000022596cef717b5c0b55ab83d2a1233e73733e81930000000000000000000000002cc7ce2914c17a9f8fe43d879cd28482cf15bfab0000000000000000000000003402adfe6de6d92fc51d0cc225df2d0e37d8c7020000000000000000000000003bbda5020e5806d7be056425cf46011d60696f1700000000000000000000000045e6fd78f7350653d35c917a2d8e144c900a06c100000000000000000000000056c741e736ecd3ef34b1b7247e5d8fd9a62617260000000000000000000000005a99366e491a4d4d1cfec32161af738d6e709e900000000000000000000000005fb7430b50fce44d16797b208ba2df2456befeeb0000000000000000000000006111c9c9bf406d5fe020c4eec4146da31d9b810900000000000000000000000064815b222936a63f4c1c0ed0a1aa47d9fb10b3db0000000000000000000000006c462831facba3904f6ba799381ac50a889328d40000000000000000000000006fe12aab23ce6d1fee48d563595fef039934cebe000000000000000000000000700e2855bc181dd1612d98ae7edb1f89110687bb00000000000000000000000071e0ce7c4c4678729ce4eee2cdf4b2a4691a4ec200000000000000000000000075c0a1a74f637108cf0a2ed8f5ab36d6cb0f1bdb00000000000000000000000076b34508b24dc58cf66eade051286b864ed055340000000000000000000000007d56e2285a600e71306735134d20ad40158e79960000000000000000000000008894ce4c42626500a519091f05323702cadc1ebd000000000000000000000000890867d3d6a312258d763d020b8ab094686428dd0000000000000000000000008c186554f2e2cdbc4a2e5ae91642007ed7b2eaa30000000000000000000000008e089ef6c671a399f3fca5a9b7e0d8555025a9ce0000000000000000000000008eac7a0fa987c0b6a7380f856b110b0144868c8d0000000000000000000000008fb59b94094f63e676e9c84a04e044db0b793f8f0000000000000000000000009103e52100387bddd6d6a6aee17fcaab9af6c34e00000000000000000000000097900205513d9f8253e99148f73159f1e7d35906000000000000000000000000a3189dfecdea14c7e96f04791975083fb1ae124b000000000000000000000000a76b14b7b0e1a81ae617a8ccff22800568ddbe37000000000000000000000000aabb1f05377d014136ce5cfb1122776a761295ce000000000000000000000000b0bd759df8cbfdcbed919a5e7fdbee1df356a56b000000000000000000000000b5ce204ad598bdce0e308f8e492f928877978f63000000000000000000000000c2c63f7d45b1475f3206263e9e1fb54a5878adc0000000000000000000000000c62224017b3226499eaf0edd3997cc6d1c1b8e6e000000000000000000000000ca99aa36964787b075f29dbbb9df552f657ed214000000000000000000000000cf292af972143ad559fbb1b55636abd50db4221d000000000000000000000000d70183e31774c1fb2f51a5655ce24cbc25d2c746000000000000000000000000dd55430ce91ab6631dceeb94c8659569c22ff48f000000000000000000000000dd7e59cd4711ad2978374ec6c37ce8b85a045208000000000000000000000000ec28bb8c3993f7ff9672e0b7f93407382b4f601d000000000000000000000000f0a9d162e2895adf26a6420d0e86186d9ab020ef000000000000000000000000f89c6850e4d9ac11e1d4299c91aa41c67a09ca1c000000000000000000000000fe5771b268a4d924fd16dc99d293a2ca54e75c9e00000000000000000000000000000000000000000000000000000000000000320000000000000000000000000000000000000000000000000000000000000152000000000000000000000000000000000000000000000000000000000000014100000000000000000000000000000000000000000000000000000000000003e8000000000000000000000000000000000000000000000000000000000000008e000000000000000000000000000000000000000000000000000000000000009100000000000000000000000000000000000000000000000000000000000001c400000000000000000000000000000000000000000000000000000000000021af0000000000000000000000000000000000000000000000000000000000000094000000000000000000000000000000000000000000000000000000000000013e00000000000000000000000000000000000000000000000000000000000002db000000000000000000000000000000000000000000000000000000000000009e000000000000000000000000000000000000000000000000000000000000008d0000000000000000000000000000000000000000000000000000000000000013000000000000000000000000000000000000000000000000000000000000016a00000000000000000000000000000000000000000000000000000000000000ab00000000000000000000000000000000000000000000000000000000000000960000000000000000000000000000000000000000000000000000000000000014000000000000000000000000000000000000000000000000000000000000001100000000000000000000000000000000000000000000000000000000000000910000000000000000000000000000000000000000000000000000000000000094000000000000000000000000000000000000000000000000000000000000009b00000000000000000000000000000000000000000000000000000000000000b500000000000000000000000000000000000000000000000000000000000000a40000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000009600000000000000000000000000000000000000000000000000000000000002c800000000000000000000000000000000000000000000000000000000000000f400000000000000000000000000000000000000000000000000000000000001a800000000000000000000000000000000000000000000000000000000000000e700000000000000000000000000000000000000000000000000000000000003ea0000000000000000000000000000000000000000000000000000000000000099000000000000000000000000000000000000000000000000000000000000008d000000000000000000000000000000000000000000000000000000000000008d00000000000000000000000000000000000000000000000000000000000000a1000000000000000000000000000000000000000000000000000000000000031300000000000000000000000000000000000000000000000000000000000002d700000000000000000000000000000000000000000000000000000000000001e300000000000000000000000000000000000000000000000000000000000000db00000000000000000000000000000000000000000000000000000000000002660000000000000000000000000000000000000000000000000000000000002225000000000000000000000000000000000000000000000000000000000000227500000000000000000000000000000000000000000000000000000000000000b100000000000000000000000000000000000000000000000000000000000001cd000000000000000000000000000000000000000000000000000000000000009500000000000000000000000000000000000000000000000000000000000002c300000000000000000000000000000000000000000000000000000000000003b5000000000000000000000000000000000000000000000000000000000000008e000000000000000000000000000000000000000000000000000000000000023600000000000000000000000000000000000000000000000000000000000000cd00000000000000000000000000000000000000000000000000000000000002b9"

// ErrNotFinalized is returned when a transaction is not finalized
var ErrNotFinalized = goerrors.New("not finalized")

// ErrTxFailed is returned when a transaction has failed
var ErrTxFailed = goerrors.New("transaction failed")

// Mgr manages all communication with Ethereum
type Mgr struct {
	rpcs                      map[string]rpc.Client
	broadcaster               broadcast.Broadcaster
	validator                 sdk.ValAddress
	proxy                     sdk.AccAddress
	latestFinalizedBlockCache LatestFinalizedBlockCache
	chainID                   string
}

// NewMgr returns a new Mgr instance
func NewMgr(rpcs map[string]rpc.Client, broadcaster broadcast.Broadcaster, valAddr sdk.ValAddress, proxy sdk.AccAddress, chainID string, latestFinalizedBlockCache LatestFinalizedBlockCache) *Mgr {
	return &Mgr{
		rpcs:                      rpcs,
		proxy:                     proxy,
		broadcaster:               broadcaster,
		validator:                 valAddr,
		latestFinalizedBlockCache: latestFinalizedBlockCache,
		chainID:                   chainID,
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
	if !mgr.isParticipantOf(event.Participants) {
		mgr.logger("pollID", event.PollID).Debug("ignoring deposit confirmation poll: not a participant")
		return nil
	}

	txReceipt, err := mgr.GetTxReceiptIfFinalized(event.Chain, common.Hash(event.TxID), event.ConfirmationHeight)
	if err != nil {
		return err
	}
	if txReceipt == nil {
		mgr.logger().Infof("broadcasting empty vote for poll %s", event.PollID.String())
		_, err := mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain)))

		return err
	}

	var events []types.Event
	for i, log := range txReceipt.Logs {
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

	mgr.logger().Infof("broadcasting vote %v for poll %s", events, event.PollID.String())
	_, err = mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain, events...)))

	return err
}

// ProcessTokenConfirmation votes on the correctness of an EVM chain token deployment
func (mgr Mgr) ProcessTokenConfirmation(event *types.ConfirmTokenStarted) error {
	if !mgr.isParticipantOf(event.Participants) {
		mgr.logger("pollID", event.PollID).Debug("ignoring token confirmation poll: not a participant")
		return nil
	}

	txReceipt, err := mgr.GetTxReceiptIfFinalized(event.Chain, common.Hash(event.TxID), event.ConfirmationHeight)
	if err != nil {
		return err
	}
	if txReceipt == nil {
		mgr.logger().Infof("broadcasting empty vote for poll %s", event.PollID.String())
		_, err := mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain)))

		return err
	}

	var events []types.Event
	for i, log := range txReceipt.Logs {
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

	mgr.logger().Infof("broadcasting vote %v for poll %s", events, event.PollID.String())
	_, err = mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain, events...)))

	return err
}

// ProcessTransferKeyConfirmation votes on the correctness of an EVM chain key transfer
func (mgr Mgr) ProcessTransferKeyConfirmation(event *types.ConfirmKeyTransferStarted) error {
	if !mgr.isParticipantOf(event.Participants) {
		mgr.logger("pollID", event.PollID).Debug("ignoring key transfer confirmation poll: not a participant")
		return nil
	}

	if mgr.isStuckFileCoinKeyRotation(event) {
		events := mgr.prepareVoteForStuckFileCoinKeyRotation(event)
		voteRequest := voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain, events...))

		mgr.logger().Infof("broadcasting rescue vote %v for poll %s, to get filecoin key rotation unstuck", events, event.PollID.String())
		_, err := mgr.broadcaster.Broadcast(context.TODO(), voteRequest)
		return err
	}

	txReceipt, err := mgr.GetTxReceiptIfFinalized(event.Chain, common.Hash(event.TxID), event.ConfirmationHeight)
	if err != nil {
		return err
	}
	if txReceipt == nil {
		mgr.logger().Infof("broadcasting empty vote for poll %s", event.PollID.String())
		_, err := mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain)))

		return err
	}

	var events []types.Event
	for i := len(txReceipt.Logs) - 1; i >= 0; i-- {
		txlog := txReceipt.Logs[i]

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

	mgr.logger().Infof("broadcasting vote %v for poll %s", events, event.PollID.String())
	_, err = mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain, events...)))

	return err
}

func (mgr Mgr) isStuckFileCoinKeyRotation(event *types.ConfirmKeyTransferStarted) bool {
	if mgr.chainID != "axelar-testnet-lisbon-3" {
		return false
	}

	if event.Chain != "filecoin-2" {
		return false
	}

	if event.TxID != types.Hash(common.HexToHash("0xd9cc8d5f0238dfe360776ed9029835d8a470d4d30e46a43deb76f316c0bd5740")) {
		return false
	}

	if event.GatewayAddress != types.Address(common.HexToAddress("0x999117D44220F33e0441fbAb2A5aDB8FF485c54D")) {
		return false
	}

	return true
}

func (mgr Mgr) prepareVoteForStuckFileCoinKeyRotation(event *types.ConfirmKeyTransferStarted) []types.Event {
	multisigOperatorshipTransferred := funcs.Must(DecodeMultisigOperatorshipTransferredEvent(&filecoinTxLog))
	return []types.Event{{Chain: event.Chain,
		TxID:  event.TxID,
		Index: uint64(0),
		Event: &types.Event_MultisigOperatorshipTransferred{
			MultisigOperatorshipTransferred: &multisigOperatorshipTransferred,
		},
	}}
}

// ProcessGatewayTxConfirmation votes on the correctness of an EVM chain gateway's transactions
func (mgr Mgr) ProcessGatewayTxConfirmation(event *types.ConfirmGatewayTxStarted) error {
	if !mgr.isParticipantOf(event.Participants) {
		mgr.logger("pollID", event.PollID).Debug("ignoring gateway tx confirmation poll: not a participant")
		return nil
	}

	txReceipt, err := mgr.GetTxReceiptIfFinalized(event.Chain, common.Hash(event.TxID), event.ConfirmationHeight)
	if err != nil {
		return err
	}
	if txReceipt == nil {
		mgr.logger().Infof("broadcasting empty vote for poll %s", event.PollID.String())
		_, err := mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain)))

		return err
	}

	events := mgr.processGatewayTxLogs(event.Chain, event.GatewayAddress, txReceipt.Logs)
	mgr.logger().Infof("broadcasting vote %v for poll %s", events, event.PollID.String())
	_, err = mgr.broadcaster.Broadcast(context.TODO(), voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain, events...)))

	return err
}

// ProcessGatewayTxsConfirmation votes on the correctness of an EVM chain multiple gateway transactions
func (mgr Mgr) ProcessGatewayTxsConfirmation(event *types.ConfirmGatewayTxsStarted) error {
	if !mgr.isParticipantOf(event.Participants) {
		pollIDs := slices.Map(event.PollMappings, func(m types.PollMapping) vote.PollID { return m.PollID })
		mgr.logger("poll_ids", pollIDs).Debug("ignoring gateway txs confirmation poll: not a participant")
		return nil
	}

	txIDs := slices.Map(event.PollMappings, func(poll types.PollMapping) common.Hash { return common.Hash(poll.TxID) })
	txReceipts, err := mgr.GetTxReceiptsIfFinalized(event.Chain, txIDs, event.ConfirmationHeight)
	if err != nil {
		return err
	}

	var votes []sdk.Msg
	for i, result := range txReceipts {
		pollID := event.PollMappings[i].PollID
		txID := event.PollMappings[i].TxID

		logger := mgr.logger("chain", event.Chain, "poll_id", pollID.String(), "tx_id", txID.Hex())

		// only broadcast empty votes if the tx is not found or not finalized
		switch result.Err() {
		case nil:
			events := mgr.processGatewayTxLogs(event.Chain, event.GatewayAddress, result.Ok().Logs)
			logger.Infof("broadcasting vote %v", events)
			votes = append(votes, voteTypes.NewVoteRequest(mgr.proxy, pollID, types.NewVoteEvents(event.Chain, events...)))
		case ErrNotFinalized:
			logger.Debug(fmt.Sprintf("transaction %s not finalized", txID.Hex()))
			logger.Infof("broadcasting empty vote due to error: %s", result.Err().Error())
			votes = append(votes, voteTypes.NewVoteRequest(mgr.proxy, pollID, types.NewVoteEvents(event.Chain)))
		case ErrTxFailed:
			logger.Debug(fmt.Sprintf("transaction %s failed", txID.Hex()))
			logger.Infof("broadcasting empty vote due to error: %s", result.Err().Error())
			votes = append(votes, voteTypes.NewVoteRequest(mgr.proxy, pollID, types.NewVoteEvents(event.Chain)))
		case ethereum.NotFound:
			logger.Debug(fmt.Sprintf("transaction receipt %s not found", txID.Hex()))
			logger.Infof("broadcasting empty vote due to error: %s", result.Err().Error())
			votes = append(votes, voteTypes.NewVoteRequest(mgr.proxy, pollID, types.NewVoteEvents(event.Chain)))
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

func (mgr Mgr) GetTxReceiptIfFinalized(chain nexus.ChainName, txID common.Hash, confHeight uint64) (*geth.Receipt, error) {
	client, ok := mgr.rpcs[strings.ToLower(chain.String())]
	if !ok {
		return nil, fmt.Errorf("rpc client not found for chain %s", chain.String())
	}

	txReceipt, err := client.TransactionReceipt(context.Background(), txID)
	keyvals := []interface{}{"chain", chain.String(), "tx_id", txID.Hex()}
	logger := mgr.logger(keyvals...)
	if err == ethereum.NotFound {
		logger.Debug(fmt.Sprintf("transaction receipt %s not found", txID.Hex()))
		return nil, nil
	}
	if err != nil {
		return nil, sdkerrors.Wrap(errors.With(err, keyvals...), "failed getting transaction receipt")
	}

	if txReceipt.Status != geth.ReceiptStatusSuccessful {
		return nil, nil
	}

	isFinalized, err := mgr.isTxReceiptFinalized(chain, txReceipt, confHeight)
	if err != nil {
		return nil, sdkerrors.Wrapf(errors.With(err, keyvals...), "cannot determine if the transaction %s is finalized", txID.Hex())
	}
	if !isFinalized {
		logger.Debug(fmt.Sprintf("transaction %s in block %s not finalized", txID.Hex(), txReceipt.BlockNumber.String()))

		return nil, nil
	}

	return txReceipt, nil
}

// GetTxReceiptsIfFinalized retrieves receipts for provided transaction IDs, only if they're finalized.
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

	isFinalized := func(receipt *geth.Receipt) rs.Result[*geth.Receipt] {
		if receipt.Status != geth.ReceiptStatusSuccessful {
			return rs.FromErr[*geth.Receipt](ErrTxFailed)
		}

		isFinalized, err := mgr.isTxReceiptFinalized(chain, receipt, confHeight)
		if err != nil {
			return rs.FromErr[*geth.Receipt](sdkerrors.Wrapf(errors.With(err, "chain", chain.String()),
				"cannot determine if the transaction %s is finalized", receipt.TxHash.Hex()),
			)
		}

		if !isFinalized {
			return rs.FromErr[*geth.Receipt](ErrNotFinalized)
		}

		return rs.FromOk(receipt)
	}

	return slices.Map(results, func(r rpc.Result) rs.Result[*geth.Receipt] {
		return rs.Pipe(rs.Result[*geth.Receipt](r), isFinalized)
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
