package axelarnet

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/events"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

// Fee is used to pay relayer for executing cross chain message
type Fee struct {
	Amount          string  `json:"amount"`
	Recipient       string  `json:"recipient"`
	RefundRecipient *string `json:"refund_recipient"`
}

// ValidateBasic validates the fee
func (f Fee) ValidateBasic() error {
	amount, ok := sdk.NewIntFromString(f.Amount)
	if !ok || !amount.IsPositive() {
		return fmt.Errorf("invalid fee amount")
	}

	if _, err := sdk.AccAddressFromBech32(f.Recipient); err != nil {
		return err
	}

	if f.RefundRecipient != nil {
		// The refund recipient is an address of the chain where the request is originally initiated from,
		// and therefore we cannot validate much of it.
		if err := utils.ValidateString(*f.RefundRecipient); err != nil {
			return err
		}
	}

	return nil
}

func validateFee(ctx sdk.Context, n types.Nexus, b types.BankKeeper, token sdk.Coin, msgType nexus.MessageType, sourceChain nexus.Chain, fee Fee) error {
	if err := fee.ValidateBasic(); err != nil {
		return err
	}

	afterFee := token.Amount.Sub(funcs.MustOk(sdk.NewIntFromString(fee.Amount)))
	switch msgType {
	case nexus.TypeGeneralMessage:
		if afterFee.IsNegative() {
			return fmt.Errorf("fee amount is greater than transfer value")
		}
	case nexus.TypeGeneralMessageWithToken:
		if !afterFee.IsPositive() {
			return fmt.Errorf("transfer amount must be non-zero after fees are deducted")
		}
	default:
		return fmt.Errorf("unexpected message type for fee")
	}

	axelar := funcs.MustOk(n.GetChain(ctx, axelarnet.Axelarnet.GetName()))
	if !n.IsAssetRegistered(ctx, axelar, token.GetDenom()) {
		return fmt.Errorf("unregistered fee denom %s", token.GetDenom())
	}

	if fee.RefundRecipient != nil {
		if err := n.ValidateAddress(ctx, nexus.CrossChainAddress{Chain: sourceChain, Address: *fee.RefundRecipient}); err != nil {
			return err
		}
	}

	return nil
}

// Message is attached in ICS20 packet memo field
type Message struct {
	DestinationChain   string `json:"destination_chain"`
	DestinationAddress string `json:"destination_address"`
	Payload            []byte `json:"payload"`
	Type               int64  `json:"type"`
	Fee                *Fee   `json:"fee"` // Optional
}

// OnRecvMessage handles general message from a cosmos chain
func OnRecvMessage(ctx sdk.Context, k keeper.Keeper, ibcK keeper.IBCKeeper, n types.Nexus, b types.BankKeeper, r RateLimiter, packet ibcexported.PacketI) ibcexported.Acknowledgement {
	// The acknowledgement is considered successful if it is a ResultAcknowledgement,
	// follow ibc transfer convention, put byte(1) in ResultAcknowledgement to indicate success.
	ack := channeltypes.NewResultAcknowledgement([]byte{byte(1)})

	data, err := types.ToICS20Packet(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	err = validateReceiver(data.GetReceiver())
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// Skip if packet not sent to Axelar message sender account.
	if data.GetReceiver() != types.AxelarGMPAccount.String() {
		// Rate limit non-GMP IBC transfers
		// IBC receives are rate limited on the from direction (tokens coming from the source chain).
		if err := r.RateLimitPacket(ctx, packet, nexus.TransferDirectionFrom, types.NewIBCPath(packet.GetDestPort(), packet.GetDestChannel())); err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}

		return ack
	}

	var msg Message
	if err := json.Unmarshal([]byte(data.GetMemo()), &msg); err != nil {
		return channeltypes.NewErrorAcknowledgement(sdkerrors.Wrapf(types.ErrGeneralMessage, "cannot unmarshal memo"))
	}

	// extract token from packet
	token, err := extractTokenFromPacketData(ctx, ibcK, n, packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	path := types.NewIBCPath(packet.GetDestPort(), packet.GetDestChannel())

	if err := validateMessage(ctx, ibcK, n, b, path, msg, token); err != nil {
		ibcK.Logger(ctx).Debug(fmt.Sprintf("failed validating message: %s", err.Error()),
			"src_channel", packet.GetSourceChannel(),
			"dest_channel", packet.GetDestChannel(),
			"sequence", packet.GetSequence(),
		)
		return channeltypes.NewErrorAcknowledgement(err)
	}

	sourceChain := funcs.MustOk(k.GetChainNameByIBCPath(ctx, path))
	sourceAddress := nexus.CrossChainAddress{
		Chain:   funcs.MustOk(n.GetChain(ctx, sourceChain)),
		Address: data.GetSender(),
	}

	rateLimitPacket := true

	switch msg.Type {
	case nexus.TypeGeneralMessage:
		err = handleMessage(ctx, n, b, sourceAddress, msg, token)
	case nexus.TypeGeneralMessageWithToken:
		err = handleMessageWithToken(ctx, n, b, sourceAddress, msg, token)
	case nexus.TypeSendToken:
		// Send token is already rate limited in nexus.EnqueueTransfer
		rateLimitPacket = false
		err = handleTokenSent(ctx, n, b, sourceAddress, msg, token)
	default:
		err = sdkerrors.Wrapf(types.ErrGeneralMessage, "unrecognized Message type")
	}

	if err != nil {
		ibcK.Logger(ctx).Debug(fmt.Sprintf("failed handling message: %s", err.Error()),
			"chain", sourceChain.String(),
			"src_channel", packet.GetSourceChannel(),
			"dest_channel", packet.GetDestChannel(),
			"sequence", packet.GetSequence(),
		)
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if rateLimitPacket {
		// IBC receives are rate limited on the from direction (tokens coming from the source chain).
		if err := r.RateLimitPacket(ctx, packet, nexus.TransferDirectionFrom, types.NewIBCPath(packet.GetDestPort(), packet.GetDestChannel())); err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}
	}

	return ack
}

func validateMessage(ctx sdk.Context, ibcK keeper.IBCKeeper, n types.Nexus, b types.BankKeeper, ibcPath string, msg Message, token keeper.Coin) error {
	srcChainName, ok := ibcK.GetChainNameByIBCPath(ctx, ibcPath)
	if !ok {
		return fmt.Errorf("unrecognized IBC path %s", ibcPath)
	}
	srcChain := funcs.MustOk(n.GetChain(ctx, srcChainName))

	if !n.IsChainActivated(ctx, srcChain) {
		return fmt.Errorf("chain %s registered for IBC path %s is deactivated", srcChain.Name, ibcPath)
	}

	destChainName := nexus.ChainName(msg.DestinationChain)
	if err := destChainName.Validate(); err != nil {
		return err
	}

	destChain, ok := n.GetChain(ctx, destChainName)
	if !ok {
		return fmt.Errorf("unrecognized destination chain %s", destChainName)
	}

	if !n.IsChainActivated(ctx, destChain) {
		return fmt.Errorf("chain %s is deactivated", destChain.Name)
	}

	if err := n.ValidateAddress(ctx, nexus.CrossChainAddress{Chain: destChain, Address: msg.DestinationAddress}); err != nil {
		return err
	}

	if msg.Fee != nil {
		err := validateFee(ctx, n, b, token.Coin, nexus.MessageType(msg.Type), srcChain, *msg.Fee)
		if err != nil {
			return err
		}
	}

	switch msg.Type {
	case nexus.TypeGeneralMessage:
		return nil
	case nexus.TypeGeneralMessageWithToken, nexus.TypeSendToken:
		if !n.IsAssetRegistered(ctx, srcChain, token.GetDenom()) {
			return fmt.Errorf("asset %s is not registered on chain %s", token.GetDenom(), srcChain.Name)
		}

		if !n.IsAssetRegistered(ctx, destChain, token.GetDenom()) {
			return fmt.Errorf("asset %s is not registered on chain %s", token.GetDenom(), destChain.Name)
		}
		return nil
	default:
		return fmt.Errorf("unrecognized message type")
	}
}

func handleMessage(ctx sdk.Context, n types.Nexus, b types.BankKeeper, sourceAddress nexus.CrossChainAddress, msg Message, token keeper.Coin) error {
	id, txID, nonce := n.GenerateMessageID(ctx)

	// ignore token for call contract
	_, err := deductFee(ctx, b, msg.Fee, token, id)
	if err != nil {
		return err
	}

	destChain := funcs.MustOk(n.GetChain(ctx, nexus.ChainName(msg.DestinationChain)))
	recipient := nexus.CrossChainAddress{Chain: destChain, Address: msg.DestinationAddress}
	m := nexus.NewGeneralMessage(
		id,
		sourceAddress,
		recipient,
		crypto.Keccak256Hash(msg.Payload).Bytes(),
		txID,
		nonce,
		nil,
	)

	events.Emit(ctx, &types.ContractCallSubmitted{
		MessageID:        m.ID,
		Sender:           m.GetSourceAddress(),
		SourceChain:      m.GetSourceChain(),
		DestinationChain: m.GetDestinationChain(),
		ContractAddress:  m.GetDestinationAddress(),
		PayloadHash:      m.PayloadHash,
		Payload:          msg.Payload,
	})

	return n.SetNewMessage(ctx, m)
}

func handleMessageWithToken(ctx sdk.Context, n types.Nexus, b types.BankKeeper, sourceAddress nexus.CrossChainAddress, msg Message, token keeper.Coin) error {
	id, txID, nonce := n.GenerateMessageID(ctx)

	token, err := deductFee(ctx, b, msg.Fee, token, id)
	if err != nil {
		return err
	}

	if err = token.Lock(b, types.AxelarGMPAccount); err != nil {
		return err
	}

	destChain := funcs.MustOk(n.GetChain(ctx, nexus.ChainName(msg.DestinationChain)))
	recipient := nexus.CrossChainAddress{Chain: destChain, Address: msg.DestinationAddress}
	m := nexus.NewGeneralMessage(
		id,
		sourceAddress,
		recipient,
		crypto.Keccak256Hash(msg.Payload).Bytes(),
		txID,
		nonce,
		&token.Coin,
	)

	events.Emit(ctx, &types.ContractCallWithTokenSubmitted{
		MessageID:        m.ID,
		Sender:           m.GetSourceAddress(),
		SourceChain:      m.GetSourceChain(),
		DestinationChain: m.GetDestinationChain(),
		ContractAddress:  m.GetDestinationAddress(),
		PayloadHash:      m.PayloadHash,
		Payload:          msg.Payload,
		Asset:            token.Coin,
	})

	return n.SetNewMessage(ctx, m)
}

func handleTokenSent(ctx sdk.Context, n types.Nexus, b types.BankKeeper, sourceAddress nexus.CrossChainAddress, msg Message, token keeper.Coin) error {
	destChain := funcs.MustOk(n.GetChain(ctx, nexus.ChainName(msg.DestinationChain)))
	crossChainAddr := nexus.CrossChainAddress{Chain: destChain, Address: msg.DestinationAddress}

	if err := token.Lock(b, types.AxelarGMPAccount); err != nil {
		return err
	}

	transferID, err := n.EnqueueTransfer(ctx, sourceAddress.Chain, crossChainAddr, token.Coin)
	if err != nil {
		return err
	}

	events.Emit(ctx, &types.TokenSent{
		TransferID:         transferID,
		Sender:             sourceAddress.Address,
		SourceChain:        sourceAddress.Chain.Name,
		DestinationAddress: crossChainAddr.Address,
		DestinationChain:   crossChainAddr.Chain.Name,
		Asset:              token.Coin,
	})

	return nil

}

// extractTokenFromPacketData get normalized token from ICS20 packet
// panic if unable to unmarshal packet data
func extractTokenFromPacketData(ctx sdk.Context, ibcK keeper.IBCKeeper, n types.Nexus, packet ibcexported.PacketI) (keeper.Coin, error) {
	data := funcs.Must(types.ToICS20Packet(packet))

	// parse the transfer amount
	amount := funcs.MustOk(sdk.NewIntFromString(data.Amount))

	var denom string
	if ibctransfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom) {
		// sender chain is not the source, un-escrow token

		// remove prefix added by sender chain
		icsPrefix := ibctransfertypes.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
		unprefixedDenom := data.Denom[len(icsPrefix):]

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

	return keeper.NewCoin(ctx, ibcK, n, sdk.NewCoin(denom, amount))
}

// deductFee pays the fee and returns the updated transfer amount with the fee deducted
func deductFee(ctx sdk.Context, b types.BankKeeper, fee *Fee, token keeper.Coin, msgID string) (keeper.Coin, error) {
	if fee == nil {
		return token, nil
	}

	feeAmount := funcs.MustOk(sdk.NewIntFromString(fee.Amount))
	coin := sdk.NewCoin(funcs.Must(token.GetOriginalDenom()), feeAmount)
	recipient := funcs.Must(sdk.AccAddressFromBech32(fee.Recipient))

	feePaidEvent := types.FeePaid{
		MessageID: msgID,
		Recipient: recipient,
		Fee:       coin,
		Asset:     token.GetDenom(),
	}
	if fee.RefundRecipient != nil {
		feePaidEvent.RefundRecipient = *fee.RefundRecipient
	}
	events.Emit(ctx, &feePaidEvent)

	// subtract fee from transfer value
	token.Amount = token.Amount.Sub(feeAmount)

	return token, b.SendCoins(ctx, types.AxelarGMPAccount, recipient, sdk.NewCoins(coin))
}

// validateReceiver rejects uppercase GMP account address
func validateReceiver(receiver string) error {
	if strings.ToUpper(receiver) == receiver && types.AxelarGMPAccount.Equals(funcs.Must(sdk.AccAddressFromBech32(receiver))) {
		return fmt.Errorf("uppercase GMP account address is not allowed")
	}

	return nil
}
