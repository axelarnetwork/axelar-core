package axelarnet

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/axelarnetwork/axelar-core/utils/events"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

// Message is attached in ICS20 packet memo field
type Message struct {
	DestinationChain   string `json:"destination_chain"`
	DestinationAddress string `json:"destination_address"`
	Payload            []byte `json:"payload"`
	Type               int    `json:"type"`
}

func validateMessage(ctx sdk.Context, k keeper.Keeper, n types.Nexus, ibcPath string, msg Message, token sdk.Coin) error {
	chainName, ok := k.GetChainNameByIBCPath(ctx, ibcPath)
	if !ok {
		return fmt.Errorf("unrecognized IBC path %s", ibcPath)
	}
	chain := funcs.MustOk(n.GetChain(ctx, chainName))

	if !n.IsChainActivated(ctx, chain) {
		return fmt.Errorf("chain %s registered for IBC path %s is deactivated", chain.Name, ibcPath)
	}

	destChainName := nexus.ChainName(msg.DestinationChain)
	if err := destChainName.Validate(); err != nil {
		return err
	}

	destChain, ok := n.GetChain(ctx, destChainName)
	if !ok {
		return fmt.Errorf("unrecognized destination chain %s", destChain.Name)
	}

	if !n.IsChainActivated(ctx, destChain) {
		return fmt.Errorf("chain %s is deactivated", destChain.Name)
	}

	if err := n.ValidateAddress(ctx, nexus.CrossChainAddress{Chain: destChain, Address: msg.DestinationAddress}); err != nil {
		return err
	}

	switch msg.Type {
	case nexus.TypeGeneralMessage:
		return nil
	case nexus.TypeGeneralMessageWithToken, nexus.TypeSendToken:
		if !n.IsAssetRegistered(ctx, chain, token.GetDenom()) {
			return fmt.Errorf("asset %s is not registered on chain %s", token.GetDenom(), destChain.Name)
		}

		if !n.IsAssetRegistered(ctx, destChain, token.GetDenom()) {
			return fmt.Errorf("asset %s is not registered on chain %s", token.GetDenom(), destChain.Name)
		}
		return nil
	default:
		return fmt.Errorf("unrecognized message type")
	}
}

// OnRecvMessage handles general message from a cosmos chain
func OnRecvMessage(ctx sdk.Context, k keeper.Keeper, ibcK keeper.IBCKeeper, n types.Nexus, b types.BankKeeper, packet ibcexported.PacketI) ibcexported.Acknowledgement {
	ack := channeltypes.NewResultAcknowledgement([]byte{byte(1)})

	data, err := toICS20Packet(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// skip if packet not sent to Axelar message sender account
	if data.GetReceiver() != types.MessageSender.String() {
		return ack
	}

	var msg Message
	if err := json.Unmarshal([]byte(data.GetMemo()), &msg); err != nil {
		ackErr := sdkerrors.Wrapf(types.ErrGeneralMessage, "cannot unmarshal memo")
		return channeltypes.NewErrorAcknowledgement(ackErr)
	}

	// extract token from packet
	token := extractTokenFromPacketData(packet)

	path := types.NewIBCPath(packet.GetDestPort(), packet.GetDestChannel())

	if err := validateMessage(ctx, k, n, path, msg, token); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	sourceChain := funcs.MustOk(k.GetChainNameByIBCPath(ctx, path))
	sourceAddress := nexus.CrossChainAddress{
		Chain:   funcs.MustOk(n.GetChain(ctx, sourceChain)),
		Address: data.GetSender(),
	}

	switch msg.Type {
	case nexus.TypeGeneralMessage:
		err = handleMessage(ctx, n, sourceAddress, msg)
	case nexus.TypeGeneralMessageWithToken:
		err = handleMessageWithToken(ctx, n, b, ibcK, sourceAddress, msg, token)
	case nexus.TypeSendToken:
		err = handleTokenSent(ctx, n, b, ibcK, sourceAddress, msg, token)
	default:
		err = sdkerrors.Wrapf(types.ErrGeneralMessage, "unrecognized Message type")
	}

	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	return ack
}

func handleMessage(ctx sdk.Context, n types.Nexus, sourceAddress nexus.CrossChainAddress, msg Message) error {
	destChain := funcs.MustOk(n.GetChain(ctx, nexus.ChainName(msg.DestinationChain)))

	recipient := nexus.CrossChainAddress{Chain: destChain, Address: msg.DestinationAddress}
	m := nexus.NewGeneralMessage(
		n.GenerateMessageID(ctx, ctx.TxBytes()),
		sourceAddress,
		recipient,
		crypto.Keccak256Hash(msg.Payload).Bytes(),
		nexus.Sent,
		nil,
	)

	events.Emit(ctx, &types.ContractCallSubmitted{
		MessageID:        m.ID,
		Sender:           m.Sender.Address,
		SourceChain:      m.Sender.Chain.Name,
		DestinationChain: m.Recipient.Chain.Name,
		ContractAddress:  m.Recipient.Address,
		PayloadHash:      m.PayloadHash,
		Payload:          msg.Payload,
	})

	return n.SetNewMessage(ctx, m)
}

func handleMessageWithToken(ctx sdk.Context, n types.Nexus, b types.BankKeeper, ibcK keeper.IBCKeeper, sourceAddress nexus.CrossChainAddress, msg Message, token sdk.Coin) error {
	destChain := funcs.MustOk(n.GetChain(ctx, nexus.ChainName(msg.DestinationChain)))

	coin, err := keeper.NewCoin(ctx, ibcK, n, token)
	if err != nil {
		return err
	}

	if err = coin.Lock(b, types.MessageSender); err != nil {
		return err
	}

	recipient := nexus.CrossChainAddress{Chain: destChain, Address: msg.DestinationAddress}
	m := nexus.NewGeneralMessage(
		n.GenerateMessageID(ctx, ctx.TxBytes()),
		sourceAddress,
		recipient,
		crypto.Keccak256Hash(msg.Payload).Bytes(),
		nexus.Sent,
		&token,
	)

	events.Emit(ctx, &types.ContractCallWithTokenSubmitted{
		MessageID:        m.ID,
		Sender:           m.Sender.Address,
		SourceChain:      m.Sender.Chain.Name,
		DestinationChain: m.Recipient.Chain.Name,
		ContractAddress:  m.Recipient.Address,
		PayloadHash:      m.PayloadHash,
		Payload:          msg.Payload,
		Asset:            token,
	})

	return n.SetNewMessage(ctx, m)
}

func handleTokenSent(ctx sdk.Context, n types.Nexus, b types.BankKeeper, ibcK keeper.IBCKeeper, sourceAddress nexus.CrossChainAddress, msg Message, token sdk.Coin) error {
	destChain := funcs.MustOk(n.GetChain(ctx, nexus.ChainName(msg.DestinationChain)))
	crossChainAddr := nexus.CrossChainAddress{Chain: destChain, Address: msg.DestinationAddress}

	coin, err := keeper.NewCoin(ctx, ibcK, n, token)
	if err != nil {
		return err
	}

	if err = coin.Lock(b, types.MessageSender); err != nil {
		return err
	}

	transferID, err := n.EnqueueTransfer(ctx, sourceAddress.Chain, crossChainAddr, token)
	if err != nil {
		return err
	}

	ibcK.Logger(ctx).Debug(fmt.Sprintf("enqueued transfer for Message from chain %s", sourceAddress.Chain.Name),
		"chain", destChain.Name,
		"transferID", transferID.String(),
	)

	events.Emit(ctx, &types.TokenSent{
		TransferID:         transferID,
		SourceChain:        sourceAddress.Chain.Name,
		Sender:             sourceAddress.Address,
		DestinationChain:   nexus.ChainName(msg.DestinationChain),
		DestinationAddress: msg.DestinationAddress,
		Asset:              token,
	})

	return nil

}

// toICS20Packet unmarchal packet to ICS20 token packet
func toICS20Packet(packet ibcexported.PacketI) (ibctransfertypes.FungibleTokenPacketData, error) {
	var data ibctransfertypes.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return data, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "cannot unmarshal ICS-20 transfer packet data")
	}

	if err := data.ValidateBasic(); err != nil {
		return data, err
	}

	return data, nil
}

// extractTokenFromPacketData get token from ICS20 packet
// panic if unable to unmarshal packet data
func extractTokenFromPacketData(packet ibcexported.PacketI) sdk.Coin {
	var data ibctransfertypes.FungibleTokenPacketData
	funcs.MustNoErr(types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data))

	// parse the transfer amount
	amount := funcs.MustOk(sdk.NewIntFromString(data.Amount))

	var denom string
	if ibctransfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom) {
		// sender chain is not the source, un-escrow token

		// remove prefix added by sender chain
		voucherPrefix := ibctransfertypes.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
		unprefixedDenom := data.Denom[len(voucherPrefix):]

		// coin denomination used in sending from the escrow address
		denom = unprefixedDenom

		// the denomination used to send the coin is either
		// -the native denom
		// -the hash of the path if the denom is not native.
		denomTrace := ibctransfertypes.ParseDenomTrace(unprefixedDenom)
		if denomTrace.Path != "" {
			denom = denomTrace.IBCDenom()
		}
	} else {
		// sender chain is the source

		// since SendPacket did not prefix the denomination, we must prefix denomination here
		sourcePrefix := ibctransfertypes.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
		// NOTE: sourcePrefix contains the trailing "/"
		prefixedDenom := sourcePrefix + data.Denom

		// construct the denomination trace from the full raw denomination
		denomTrace := ibctransfertypes.ParseDenomTrace(prefixedDenom)

		denom = denomTrace.IBCDenom()
	}

	return sdk.NewCoin(denom, amount)
}
