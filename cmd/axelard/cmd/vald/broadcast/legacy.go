package broadcast

import (
	"fmt"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcast/legacy"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcast/types"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	sdkClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	abci "github.com/tendermint/tendermint/abci/types"
	tmLog "github.com/tendermint/tendermint/libs/log"
)

// Broadcaster submits transactions to a tendermint node
type LegacyBroadcasterImpl struct {
	Broadcaster
	client  legacy.LegacyClient
	signer  legacy.SignFn
	seqNo   uint64
	accNo   uint64
	chainID string
}

// NewLegacyBroadcaster returns a broadcaster to submit transactions to the blockchain with the legacy transaction data structures use by the REST endpoint.
// Only one instance of a broadcaster should be run for a given account, otherwise risk conflicting sequence numbers for submitted transactions.
func NewLegacyBroadcaster(signer legacy.SignFn, sdkCtx sdkClient.Context, client legacy.LegacyClient, conf broadcastTypes.ClientConfig, pipeline types.Pipeline, logger tmLog.Logger) (*LegacyBroadcasterImpl, error) {
	if conf.ChainID == "" {
		return nil, sdkerrors.Wrap(broadcastTypes.ErrInvalidChain, "chain ID required but not specified")
	}

	broadcaster := &LegacyBroadcasterImpl{
		signer:  signer,
		client:  client,
		seqNo:   0,
		chainID: conf.ChainID,
		Broadcaster: Broadcaster{
			logger:    logger,
			pipeline:  pipeline,
			ctx:       sdkCtx,
			txFactory: tx.Factory{},
		},
	}

	return broadcaster, nil
}

// Broadcast sends the passed messages to the network. This function in thread-safe.
func (b *LegacyBroadcasterImpl) BroadcastLegacyStdTx(tx legacytx.StdTx) (*sdk.TxResponse, error) {
	resChan := make(chan *sdk.TxResponse, 1)
	// serialize concurrent calls to broadcast
	if err := b.pipeline.Push(func() error {
		res, err := b.broadcastTx(tx)
		if err != nil {
			// reset account and sequence number in case they were the issue
			b.seqNo = 0
			b.accNo = 0
			return err
		}

		resChan <- res

		// broadcast has been successful, so increment sequence number
		b.seqNo++
		return nil
	}); err != nil {
		return nil, err
	}

	return <-resChan, nil
}

// broadcastTx signs a standard tx object and broadcasts it to the network
func (b *LegacyBroadcasterImpl) broadcastTx(stdTx legacytx.StdTx) (*sdk.TxResponse, error) {
	if len(stdTx.Msgs) == 0 {
		return nil, fmt.Errorf("call broadcast with at least one message")
	}

	// By convention the first signer of a tx pays the fees
	if len(stdTx.Msgs[0].GetSigners()) == 0 {
		return nil, fmt.Errorf("messages must have at least one signer")
	}

	stdSignMsg := legacytx.StdSignMsg{
		ChainID: b.chainID,
		Msgs:    stdTx.Msgs,
		Fee:     stdTx.Fee,
	}

	return b.signAndBroadcast(stdSignMsg)
}

func (b *LegacyBroadcasterImpl) signAndBroadcast(msg legacytx.StdSignMsg) (*sdk.TxResponse, error) {
	accNo, seqNo, err := b.updateAccountNumberSequence(msg.Msgs[0].GetSigners()[0])
	if err != nil {
		return nil, err
	}
	msg.AccountNumber = accNo
	msg.Sequence = seqNo

	tx, err := sign(b.signer, msg)
	if err != nil {
		return nil, err
	}

	b.logger.Debug(fmt.Sprintf("broadcasting %d messages from address: %.20s, acc no.: %d, seq no.: %d, chainId: %s",
		len(tx.Msgs), tx.Msgs[0].GetSigners()[0], msg.AccountNumber, msg.Sequence, msg.ChainID))

	res, err := b.client.BroadcastTxSync(tx)
	if err != nil {
		return nil, err
	}
	if res.Code != abci.CodeTypeOK {
		return nil, fmt.Errorf(res.RawLog)
	}

	// broadcast has been successful, so increment sequence number
	b.seqNo += 1
	return res, nil
}

func sign(sign legacy.SignFn, msg legacytx.StdSignMsg) (legacytx.StdTx, error) {
	var sigs []legacytx.StdSignature
	for i, m := range msg.Msgs {
		if len(m.GetSigners()) == 0 {
			return legacytx.StdTx{}, fmt.Errorf("signing failed: msg at idx [%d] without signers", i)
		}
		for _, s := range m.GetSigners() {
			sig, err := sign(s, msg)
			if err != nil {
				return legacytx.StdTx{}, err
			}
			sigs = append(sigs, sig)
		}
	}

	return legacytx.NewStdTx(msg.Msgs, msg.Fee, sigs, msg.Memo), nil
}

func (b *LegacyBroadcasterImpl) updateAccountNumberSequence(addr sdk.AccAddress) (uint64, uint64, error) {
	accNo, seqNo, err := b.client.GetAccountNumberSequence(addr)
	if err != nil {
		return 0, 0, err
	}
	if seqNo > b.seqNo {
		b.seqNo = seqNo
	}
	return accNo, b.seqNo, nil
}
