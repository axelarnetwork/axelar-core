package btc

import (
	"fmt"
	"strconv"

	sdkClient "github.com/cosmos/cosmos-sdk/client"
	sdkFlags "github.com/cosmos/cosmos-sdk/client/flags"

	tmEvents "github.com/axelarnetwork/tm-events/events"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcaster/types"
	rpc3 "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/btc/rpc"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/parse"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// Mgr manages all communication with Bitcoin
type Mgr struct {
	cliCtx      sdkClient.Context
	logger      log.Logger
	broadcaster types.Broadcaster
	rpc         rpc3.Client
	cdc         *codec.LegacyAmino
}

// NewMgr returns a new Mgr instance
func NewMgr(rpc rpc3.Client, cliCtx sdkClient.Context, broadcaster types.Broadcaster, logger log.Logger, cdc *codec.LegacyAmino) *Mgr {
	return &Mgr{
		rpc:         rpc,
		cliCtx:      cliCtx,
		logger:      logger.With("listener", "btc"),
		broadcaster: broadcaster,
		cdc:         cdc,
	}
}

// ProcessConfirmation votes on the correctness of a Bitcoin deposit
func (mgr *Mgr) ProcessConfirmation(e tmEvents.Event) error {
	outPointInfo, confHeight, pollKey, err := parseConfirmationParams(mgr.cdc, e.Attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "Bitcoin transaction confirmation failed")
	}

	err = confirmTx(mgr.rpc, outPointInfo, confHeight)
	if err != nil {
		mgr.logger.Debug(sdkerrors.Wrap(err, "tx outpoint confirmation failed").Error())
	}
	msg := btc.NewVoteConfirmOutpointRequest(mgr.cliCtx.FromAddress, pollKey, outPointInfo.GetOutPoint(), err == nil)
	refundableMsg := axelarnet.NewRefundMsgRequest(mgr.cliCtx.FromAddress, msg)

	mgr.logger.Debug(fmt.Sprintf("broadcasting vote %v for poll %s", msg.Confirmed, pollKey.String()))
	_, err = mgr.broadcaster.Broadcast(mgr.cliCtx.WithBroadcastMode(sdkFlags.BroadcastBlock), refundableMsg)
	return err
}

func parseConfirmationParams(cdc *codec.LegacyAmino, attributes map[string]string) (outPoint btc.OutPointInfo, confHeight int64, pollKey vote.PollKey, err error) {
	parsers := []*parse.AttributeParser{
		{Key: btc.AttributeKeyOutPointInfo, Map: func(s string) (interface{}, error) {
			cdc.MustUnmarshalJSON([]byte(s), &outPoint)
			return outPoint, nil
		}},
		{Key: btc.AttributeKeyConfHeight, Map: func(s string) (interface{}, error) { return strconv.ParseInt(s, 10, 64) }},
		{Key: btc.AttributeKeyPoll, Map: func(s string) (interface{}, error) {
			cdc.MustUnmarshalJSON([]byte(s), &pollKey)
			return pollKey, nil
		}},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		return btc.OutPointInfo{}, 0, vote.PollKey{}, err
	}

	return results[0].(btc.OutPointInfo), results[1].(int64), results[2].(vote.PollKey), nil
}

func confirmTx(rpc rpc3.Client, outPointInfo btc.OutPointInfo, requiredConfirmations int64) error {
	outPoint := outPointInfo.GetOutPoint()
	actualTxOut, err := rpc.GetTxOut(&outPoint.Hash, outPoint.Index, false)
	if err != nil {
		return sdkerrors.Wrap(err, "call to Bitcoin rpc failed")
	}
	if actualTxOut == nil {
		return fmt.Errorf("tx {%s} not found", outPointInfo.OutPoint)
	}

	if len(actualTxOut.ScriptPubKey.Addresses) != 1 {
		return fmt.Errorf("deposit must be only spendable by a single address")
	}

	if actualTxOut.ScriptPubKey.Addresses[0] != outPointInfo.Address {
		return fmt.Errorf("expected destination address does not match actual destination address")
	}

	// if parse fails actual amount == 0, so the comparison will fail, so no need to handle error
	actualAmount, _ := btcutil.NewAmount(actualTxOut.Value)
	if actualAmount != outPointInfo.Amount {
		return fmt.Errorf("expected amount (%v) does not match actual amount (%v)", outPointInfo.Amount, actualAmount)
	}

	if actualTxOut.Confirmations < requiredConfirmations {
		return fmt.Errorf("not enough confirmations yet, expected at least %d, got %d", requiredConfirmations, actualTxOut.Confirmations)
	}

	return nil
}
