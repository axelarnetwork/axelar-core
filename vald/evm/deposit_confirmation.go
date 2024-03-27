package evm

import (
	"bytes"
	"context"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
)

// ProcessDepositConfirmation votes on the correctness of an EVM chain token deposit
func (mgr Mgr) ProcessDepositConfirmation(event *types.ConfirmDepositStarted) error {
	if !mgr.isParticipantOf(event.Participants) {
		mgr.logger("pollID", event.PollID).Debug("ignoring deposit confirmation poll: not a participant")
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
		events := mgr.processDepositConfirmationLogs(event, txReceipt.Ok().Logs)
		vote = voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain, events...))

		mgr.logger().Infof("broadcasting vote %v for poll %s", events, event.PollID.String())
	}

	_, err = mgr.broadcaster.Broadcast(context.TODO(), vote)

	return err
}

func (mgr Mgr) processDepositConfirmationLogs(event *types.ConfirmDepositStarted, logs []*geth.Log) []types.Event {
	events := make([]types.Event, 0, len(logs))

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
