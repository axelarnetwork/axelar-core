package keeper

import (
	"encoding/json"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

var _ wasmkeeper.Messenger = (*Messenger)(nil)

type message struct {
	SenderChain      exported.ChainName `json:"sender_chain"`
	SenderAddress    string             `json:"sender_address"`
	RecipientChain   exported.ChainName `json:"recipient_chain"`
	RecipientAddress string             `json:"recipient_address"`
	PayloadHash      []byte             `json:"payload_hash"`
	SourceTxID       []byte             `json:"source_tx_id"`
	SourceTxIndex    uint64             `json:"source_tx_index"`
}

type request = []message

type Messenger struct {
	types.Nexus
}

// NewMessenger returns a new Messenger
func NewMessenger(nexus types.Nexus) Messenger {
	return Messenger{nexus}
}

// DispatchMsg decodes the messages from the cosmowasm connection router and routes them to the nexus module if possible
func (m Messenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, _ string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, err error) {
	req := request{}
	if err := json.Unmarshal(msg.Custom, &req); err != nil {
		return nil, nil, sdkerrors.Wrap(wasmtypes.ErrUnknownMsg, err.Error())
	}

	connectionRouter := m.GetParams(ctx).ConnectionRouter

	if len(connectionRouter) == 0 {
		return nil, nil, fmt.Errorf("connection router is not set")
	}

	if !connectionRouter.Equals(contractAddr) {
		return nil, nil, fmt.Errorf("contract address %s is not the connection router", contractAddr)
	}

	// TODO: consider routing messages that can be routed instead of failing the
	// whole batch whenever one message fails and returning the ones that
	// succeeded/failed in the response
	// TODO: consider handling only one message at a time instead of a batch
	for _, msg := range req {
		recipientChain, ok := m.GetChain(ctx, msg.RecipientChain)
		if !ok {
			return nil, nil, fmt.Errorf("recipient chain %s is not a registered chain", msg.RecipientChain)
		}

		msgID, _, _ := m.GenerateMessageID(ctx)
		senderChain := exported.Chain{Name: msg.SenderChain, SupportsForeignAssets: false, KeyType: tss.None, Module: wasmtypes.ModuleName}
		sender := exported.CrossChainAddress{Chain: senderChain, Address: msg.SenderAddress}
		recipient := exported.CrossChainAddress{Chain: recipientChain, Address: msg.RecipientAddress}

		status := exported.Approved
		if !recipientChain.IsFrom(axelarnet.ModuleName) {
			status = exported.Processing
		}
		msg := exported.NewGeneralMessage(
			msgID,
			sender,
			recipient,
			msg.PayloadHash,
			status,
			msg.SourceTxID,
			msg.SourceTxIndex,
			nil,
		)

		if err := m.Nexus.SetNewMessageFromWasm(ctx, msg); err != nil {
			return nil, nil, err
		}

	}

	// TODO: return events
	return nil, nil, nil
}
