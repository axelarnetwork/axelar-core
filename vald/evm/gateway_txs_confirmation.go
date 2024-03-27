package evm

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/slices"
)

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
	for i, txReceipt := range txReceipts {
		pollID := event.PollMappings[i].PollID
		txID := event.PollMappings[i].TxID

		logger := mgr.logger("chain", event.Chain, "poll_id", pollID.String(), "tx_id", txID.Hex())

		if txReceipt.Err() != nil {
			votes = append(votes, voteTypes.NewVoteRequest(mgr.proxy, pollID, types.NewVoteEvents(event.Chain)))

			logger.Infof("broadcasting empty vote for poll %s: %s", pollID.String(), txReceipt.Err().Error())
		} else {
			events := mgr.processGatewayTxLogs(event.Chain, event.GatewayAddress, txReceipt.Ok().Logs)
			votes = append(votes, voteTypes.NewVoteRequest(mgr.proxy, pollID, types.NewVoteEvents(event.Chain, events...)))

			logger.Infof("broadcasting vote %v for poll %s", events, pollID.String())
		}
	}

	_, err = mgr.broadcaster.Broadcast(context.TODO(), votes...)

	return err
}
