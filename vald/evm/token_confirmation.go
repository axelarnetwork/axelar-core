package evm

import (
	"bytes"
	"context"
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
)

// ProcessTokenConfirmation votes on the correctness of an EVM chain token deployment
func (mgr Mgr) ProcessTokenConfirmation(event *types.ConfirmTokenStarted) error {
	if !mgr.isParticipantOf(event.Participants) {
		mgr.logger("pollID", event.PollID).Debug("ignoring token confirmation poll: not a participant")
		return nil
	}

	var vote *voteTypes.VoteRequest

	txReceipt, err := mgr.GetTxReceiptIfFinalized(event.Chain, common.Hash(event.TxID), event.ConfirmationHeight)
	if err != nil {
		return err
	}
	if txReceipt.Err() != nil {
		vote = voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain))

		fmt.Printf("txReceipt.Err().Error() %#v\n", txReceipt.Err().Error())
		mgr.logger().Infof("broadcasting empty vote for poll %s: %s", event.PollID.String(), txReceipt.Err().Error())
	} else {
		events := mgr.processTokenConfirmationLogs(event, txReceipt.Ok().Logs)
		vote = voteTypes.NewVoteRequest(mgr.proxy, event.PollID, types.NewVoteEvents(event.Chain, events...))

		mgr.logger().Infof("broadcasting vote %v for poll %s", events, event.PollID.String())
	}

	_, err = mgr.broadcaster.Broadcast(context.TODO(), vote)

	return err
}

func (mgr Mgr) processTokenConfirmationLogs(event *types.ConfirmTokenStarted, logs []*geth.Log) []types.Event {
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

		return []types.Event{{
			Chain: event.Chain,
			TxID:  event.TxID,
			Index: uint64(i),
			Event: &types.Event_TokenDeployed{
				TokenDeployed: &erc20Event,
			},
		}}
	}

	return []types.Event{}
}
