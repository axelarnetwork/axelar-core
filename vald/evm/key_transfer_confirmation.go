package evm

import (
	"context"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
)

// ProcessTransferKeyConfirmation votes on the correctness of an EVM chain key transfer
func (mgr Mgr) ProcessTransferKeyConfirmation(event *types.ConfirmKeyTransferStarted) error {
	if !mgr.isParticipantOf(event.Participants) {
		mgr.logger("pollID", event.PollID).Debug("ignoring key transfer confirmation poll: not a participant")
		return nil
	}

	var vote *voteTypes.VoteRequest

	txReceipt, err := mgr.GetTxReceiptIfFinalized(event.Chain, common.Hash(event.TxID), event.ConfirmationHeight)
	if err != nil {
		return err
	}
	if txReceipt.Err() != nil {
		vote = voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain))

		mgr.logger().Infof("broadcasting empty vote for poll %s: %s", event.PollID.String(), txReceipt.Err().Error())
	} else {
		events := mgr.processTransferKeyLogs(event, txReceipt.Ok().Logs)
		vote = voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain, events...))

		mgr.logger().Infof("broadcasting vote %v for poll %s", events, event.PollID.String())
	}

	_, err = mgr.broadcaster.Broadcast(context.TODO(), vote)

	return err
}

func (mgr Mgr) processTransferKeyLogs(event *types.ConfirmKeyTransferStarted, logs []*geth.Log) []types.Event {
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

		return []types.Event{{Chain: event.Chain,
			TxID:  event.TxID,
			Index: uint64(i),
			Event: &types.Event_MultisigOperatorshipTransferred{
				MultisigOperatorshipTransferred: &transferOperatorshipEvent,
			},
		}}
	}

	return []types.Event{}
}
