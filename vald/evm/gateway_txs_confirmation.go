package evm

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	votetypes "github.com/axelarnetwork/axelar-core/x/vote/types"
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
	for i, result := range txReceipts {
		pollID := event.PollMappings[i].PollID
		txID := event.PollMappings[i].TxID

		logger := mgr.logger("chain", event.Chain, "poll_id", pollID.String(), "tx_id", txID.Hex())

		// only broadcast empty votes if the tx is not found or not finalized
		switch result.Err() {
		case nil:
			events := mgr.processGatewayTxLogs(event.Chain, event.GatewayAddress, result.Ok().Logs)
			logger.Infof("broadcasting vote %v", events)
			votes = append(votes, votetypes.NewVoteRequest(mgr.proxy, pollID, types.NewVoteEvents(event.Chain, events...)))
		case ErrNotFinalized:
			logger.Debug(fmt.Sprintf("transaction %s not finalized", txID.Hex()))
			logger.Infof("broadcasting empty vote due to error: %s", result.Err().Error())
			votes = append(votes, votetypes.NewVoteRequest(mgr.proxy, pollID, types.NewVoteEvents(event.Chain)))
		case ErrTxFailed:
			logger.Debug(fmt.Sprintf("transaction %s failed", txID.Hex()))
			logger.Infof("broadcasting empty vote due to error: %s", result.Err().Error())
			votes = append(votes, votetypes.NewVoteRequest(mgr.proxy, pollID, types.NewVoteEvents(event.Chain)))
		case ethereum.NotFound:
			logger.Debug(fmt.Sprintf("transaction receipt %s not found", txID.Hex()))
			logger.Infof("broadcasting empty vote due to error: %s", result.Err().Error())
			votes = append(votes, votetypes.NewVoteRequest(mgr.proxy, pollID, types.NewVoteEvents(event.Chain)))
		default:
			logger.Errorf("failed to get tx receipt: %s", result.Err().Error())
		}

	}

	_, err = mgr.broadcaster.Broadcast(context.TODO(), votes...)

	return err
}
