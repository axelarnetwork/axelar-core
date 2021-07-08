package btc

import (
	"fmt"
	"strconv"

	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcaster/types"
	rpc3 "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/btc/rpc"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// Mgr manages all communication with Bitcoin
type Mgr struct {
	logger      log.Logger
	broadcaster types.Broadcaster
	rpc         rpc3.Client
	sender      sdk.AccAddress
	cdc         *codec.LegacyAmino
}

// NewMgr returns a new Mgr instance
func NewMgr(rpc rpc3.Client, broadcaster types.Broadcaster, sender sdk.AccAddress, logger log.Logger, cdc *codec.LegacyAmino) *Mgr {
	return &Mgr{
		rpc:         rpc,
		logger:      logger.With("listener", "btc"),
		broadcaster: broadcaster,
		sender:      sender,
		cdc:         cdc,
	}
}

// ProcessConfirmation votes on the correctness of a Bitcoin deposit
func (mgr *Mgr) ProcessConfirmation(attributes []sdk.Attribute) error {
	outPointInfo, confHeight, pollKey, err := parseConfirmationParams(mgr.cdc, attributes)
	if err != nil {
		return sdkerrors.Wrap(err, "Bitcoin transaction confirmation failed")
	}

	err = confirmTx(mgr.rpc, outPointInfo, confHeight)
	if err != nil {
		mgr.logger.Debug(sdkerrors.Wrap(err, "tx outpoint confirmation failed").Error())
	}
	msg := btc.NewVoteConfirmOutpointRequest(mgr.sender, pollKey, outPointInfo.GetOutPoint(), err == nil)

	mgr.logger.Debug(fmt.Sprintf("broadcasting vote %v for poll %s", msg.Confirmed, pollKey.String()))
	return mgr.broadcaster.Broadcast(msg)
}

func parseConfirmationParams(cdc *codec.LegacyAmino, attributes []sdk.Attribute) (outPoint btc.OutPointInfo, confHeight int64, pollKey vote.PollKey, err error) {
	var outPointFound, confHeightFound, pollKeyFound bool
	for _, attribute := range attributes {
		switch attribute.Key {
		case btc.AttributeKeyOutPointInfo:
			cdc.MustUnmarshalJSON([]byte(attribute.Value), &outPoint)
			outPointFound = true
		case btc.AttributeKeyConfHeight:
			h, err := strconv.Atoi(attribute.Value)
			if err != nil {
				return btc.OutPointInfo{}, 0, vote.PollKey{}, sdkerrors.Wrap(err, "could not parse confirmation height")
			}
			confHeight = int64(h)
			confHeightFound = true
		case btc.AttributeKeyPoll:
			cdc.MustUnmarshalJSON([]byte(attribute.Value), &pollKey)
			pollKeyFound = true
		default:
		}
	}
	if !outPointFound || !confHeightFound || !pollKeyFound {
		return btc.OutPointInfo{}, 0, vote.PollKey{}, fmt.Errorf("insufficient event attributes")
	}

	return outPoint, confHeight, pollKey, nil
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
