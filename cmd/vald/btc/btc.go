package btc

import (
	"fmt"
	"strconv"

	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	broadcast "github.com/axelarnetwork/axelar-core/cmd/vald/broadcast/types"
	rpc2 "github.com/axelarnetwork/axelar-core/cmd/vald/btc/rpc"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// Mgr manages all communication with Bitcoin
type Mgr struct {
	logger      log.Logger
	broadcaster broadcast.Broadcaster
	rpc         rpc2.Client
	sender      sdk.AccAddress
}

// NewMgr returns a new Mgr instance
func NewMgr(rpc rpc2.Client, broadcaster broadcast.Broadcaster, defaultSender sdk.AccAddress, logger log.Logger) *Mgr {
	return &Mgr{
		rpc:         rpc,
		logger:      logger.With("listener", "btc"),
		broadcaster: broadcaster,
		sender:      defaultSender,
	}
}

// ProcessConfirmation votes on the correctness of a Bitcoin deposit
func (mgr *Mgr) ProcessConfirmation(attributes []sdk.Attribute) error {
	outPointInfo, confHeight, poll, err := parseConfirmationParams(attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "Bitcoin transaction confirmation failed")
	}

	err = confirmTx(mgr.rpc, outPointInfo, confHeight)
	if err != nil {
		mgr.logger.Debug(sdkerrors.Wrap(err, "tx outpoint confirmation failed").Error())
	}
	msg := btc.MsgVoteConfirmOutpoint{
		Sender:    mgr.sender,
		Poll:      poll,
		Confirmed: err == nil,
		OutPoint:  *outPointInfo.OutPoint,
	}
	mgr.logger.Debug(fmt.Sprintf("broadcasting vote %v for poll %s", msg.Confirmed, poll.String()))
	return mgr.broadcaster.Broadcast(msg)
}

func parseConfirmationParams(attributes []sdk.Attribute) (outPoint btc.OutPointInfo, confHeight int64, poll vote.PollMeta, err error) {
	var outPointFound, confHeightFound, pollFound bool
	for _, attribute := range attributes {
		switch attribute.Key {
		case btc.AttributeKeyOutPointInfo:
			codec.Cdc.MustUnmarshalJSON([]byte(attribute.Value), &outPoint)
			outPointFound = true
		case btc.AttributeKeyConfHeight:
			h, err := strconv.Atoi(attribute.Value)
			if err != nil {
				return btc.OutPointInfo{}, 0, vote.PollMeta{}, sdkerrors.Wrap(err, "could not parse confirmation height")
			}
			confHeight = int64(h)
			confHeightFound = true
		case btc.AttributeKeyPoll:
			codec.Cdc.MustUnmarshalJSON([]byte(attribute.Value), &poll)
			pollFound = true
		default:
		}
	}
	if !outPointFound || !confHeightFound || !pollFound {
		return btc.OutPointInfo{}, 0, vote.PollMeta{}, fmt.Errorf("insufficient event attributes")
	}

	return outPoint, confHeight, poll, nil
}

func confirmTx(rpc rpc2.Client, outPointInfo btc.OutPointInfo, requiredConfirmations int64) error {
	actualTxOut, err := rpc.GetTxOut(&outPointInfo.OutPoint.Hash, outPointInfo.OutPoint.Index, false)
	if err != nil {
		return sdkerrors.Wrap(err, "call to Bitcoin rpc failed")
	}
	if actualTxOut == nil {
		return fmt.Errorf("tx {%s} not found", outPointInfo.OutPoint.String())
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
