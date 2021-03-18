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
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// Mgr manages all communication with Bitcoin
type Mgr struct {
	myAddress   string
	Logger      log.Logger
	broadcaster broadcast.Broadcaster
	rpc         btc.RPCClient
}

// NewMgr returns a new Mgr instance
func NewMgr(rpc btc.RPCClient, myAddress string, broadcaster broadcast.Broadcaster, logger log.Logger) *Mgr {
	return &Mgr{
		rpc:         rpc,
		myAddress:   myAddress,
		Logger:      logger.With("listener", "btc"),
		broadcaster: broadcaster,
	}
}

// ProcessVerification votes on the correctness of a Bitcoin transaction
func (mgr *Mgr) ProcessVerification(attributes []sdk.Attribute) error {
	outPointInfo, confHeight, poll, err := parseVerificationStartParams(attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "failed Bitcoin transaction verification")
	}

	err = verifyTx(mgr.rpc, outPointInfo, confHeight)
	var v btc.MsgVoteVerifiedTx
	if err != nil {
		mgr.Logger.Debug(sdkerrors.Wrap(err, "verification failed").Error())
		v = btc.MsgVoteVerifiedTx{PollMeta: poll, VotingData: false}
	} else {
		v = btc.MsgVoteVerifiedTx{PollMeta: poll, VotingData: true}
	}
	mgr.Logger.Debug(fmt.Sprintf("broadcasting vote %v for poll %s", v.VotingData, poll.String()))
	return mgr.broadcaster.Broadcast(v)
}

func parseVerificationStartParams(attributes []sdk.Attribute) (outPoint btc.OutPointInfo, confHeight int64, poll vote.PollMeta, err error) {
	found := 0
	for _, attribute := range attributes {
		switch attribute.Key {
		case btc.AttributeKeyOutPoint:
			codec.Cdc.MustUnmarshalJSON([]byte(attribute.Value), &outPoint)
			found++
		case btc.AttributeKeyConfHeight:
			h, err := strconv.Atoi(attribute.Value)
			if err != nil {
				return btc.OutPointInfo{}, 0, vote.PollMeta{}, sdkerrors.Wrap(err, "could not parse confirmation height")
			}
			confHeight = int64(h)
			found++
		case btc.AttributeKeyPoll:
			codec.Cdc.MustUnmarshalJSON([]byte(attribute.Value), &poll)
			found++
		default:
		}
	}
	if found != 3 {
		return btc.OutPointInfo{}, 0, vote.PollMeta{}, fmt.Errorf("insufficient event attributes")
	}

	return outPoint, confHeight, poll, nil
}

func verifyTx(rpc btc.RPCClient, outPointInfo btc.OutPointInfo, requiredConfirmations int64) error {
	actualTxOut, err := rpc.GetTxOut(&outPointInfo.OutPoint.Hash, outPointInfo.OutPoint.Index, false)
	if actualTxOut == nil {
		return fmt.Errorf("tx {%s} not found", outPointInfo.OutPoint.String())
	}
	if err != nil {
		return sdkerrors.Wrap(err, "call to Bitcoin rpc failed")
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
		return fmt.Errorf("not enough confirmations yet")
	}

	return nil
}
