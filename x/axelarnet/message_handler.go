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
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

// Fee is used to pay relayer for executing cross chain message
type Fee struct {
	Amount    string `json:"amount"`
	Recipient string `json:"recipient"`
}

func (f Fee) validate(ctx sdk.Context, n types.Nexus, b types.BankKeeper, token sdk.Coin, msgType nexus.MessageType) error {
	amt, ok := sdk.NewIntFromString(f.Amount)
	if !ok || amt.LTE(sdk.ZeroInt()) {
		return fmt.Errorf("invalid fee amount")
	}

	afterFee := token.Amount.Sub(amt)
	switch msgType {
	case nexus.TypeGeneralMessage:
		if afterFee.LT(sdk.ZeroInt()) {
			return fmt.Errorf("fee amount is greater than transfer value")
		}
	case nexus.TypeGeneralMessageWithToken:
		if afterFee.LTE(sdk.ZeroInt()) {
			return fmt.Errorf("fee amount is greater or equal to transfer value")
		}
	default:
		return fmt.Errorf("unexpected message type for fee")
	}

	axelar := funcs.MustOk(n.GetChain(ctx, axelarnet.Axelarnet.GetName()))
	if !n.IsAssetRegistered(ctx, axelar, token.GetDenom()) {
		return fmt.Errorf("unregistered fee denom %s", token.GetDenom())
	}

	addr, err := sdk.AccAddressFromBech32(f.Recipient)
	if err != nil {
		return err
	}

	if b.BlockedAddr(addr) {
		return fmt.Errorf("fee recipient is a blocked address")
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
	// the acknowledgement is considered successful if it is a ResultAcknowledgement,
	// follow ibc transfer convention, put byte(1) in ResultAcknowledgement to indicate success.
	ack := channeltypes.NewResultAcknowledgement([]byte{byte(1)})

	data, err := types.ToICS20Packet(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// skip if packet not sent to Axelar message sender account
	if data.GetReceiver() != types.AxelarGMPAccount.String() {
		// Rate limit non-GMP IBC transfers
		// IBC receives are rate limited on the Incoming direction (tokens coming in to Axelar hub).
		if err := r.RateLimitPacket(ctx, packet, nexus.Incoming, types.NewIBCPath(packet.GetDestPort(), packet.GetDestChannel())); err != nil {
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
		// IBC receives are rate limited on the Incoming direction (tokens coming in to Axelar hub).
		if err := r.RateLimitPacket(ctx, packet, nexus.Incoming, types.NewIBCPath(packet.GetDestPort(), packet.GetDestChannel())); err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}
	}

	return ack
}

func validateMessage(ctx sdk.Context, ibcK keeper.IBCKeeper, n types.Nexus, b types.BankKeeper, ibcPath string, msg Message, token keeper.Coin) error {
	chainName, ok := ibcK.GetChainNameByIBCPath(ctx, ibcPath)
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

	// only allow sending messages to EVM chains
	if (msg.Type == nexus.TypeGeneralMessage || msg.Type == nexus.TypeGeneralMessageWithToken) &&
		!destChain.IsFrom(evmtypes.ModuleName) {
		return fmt.Errorf("destination chain is not an EVM chain")
	}

	if msg.Fee != nil {
		err := msg.Fee.validate(ctx, n, b, token.Coin, nexus.MessageType(msg.Type))
		if err != nil {
			return err
		}
	}

	switch msg.Type {
	case nexus.TypeGeneralMessage:
		return nil
	case nexus.TypeGeneralMessageWithToken, nexus.TypeSendToken:
		if !n.IsAssetRegistered(ctx, chain, token.GetDenom()) {
			return fmt.Errorf("asset %s is not registered on chain %s", token.GetDenom(), chain.Name)
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
	id := n.GenerateMessageID(ctx)

	fee, err := chargeFee(ctx, b, msg.Fee, funcs.Must(token.GetOriginalDenom()), id)
	if err != nil {
		return err
	}
	// subtract fee from transfer value
	token.Amount = token.Amount.Sub(fee)

	destChain := funcs.MustOk(n.GetChain(ctx, nexus.ChainName(msg.DestinationChain)))
	recipient := nexus.CrossChainAddress{Chain: destChain, Address: msg.DestinationAddress}
	m := nexus.NewGeneralMessage(
		id,
		sourceAddress,
		recipient,
		crypto.Keccak256Hash(msg.Payload).Bytes(),
		nexus.Approved,
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
	id := n.GenerateMessageID(ctx)

	fee, err := chargeFee(ctx, b, msg.Fee, funcs.Must(token.GetOriginalDenom()), id)
	if err != nil {
		return err
	}
	// subtract fee from transfer value
	token.Amount = token.Amount.Sub(fee)
	if err := token.Lock(b, types.AxelarGMPAccount); err != nil {
		return err
	}

	destChain := funcs.MustOk(n.GetChain(ctx, nexus.ChainName(msg.DestinationChain)))
	recipient := nexus.CrossChainAddress{Chain: destChain, Address: msg.DestinationAddress}
	m := nexus.NewGeneralMessage(
		id,
		sourceAddress,
		recipient,
		crypto.Keccak256Hash(msg.Payload).Bytes(),
		nexus.Approved,
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

// chargeFee transfers fee to the recipient, and returns fee amount
func chargeFee(ctx sdk.Context, b types.BankKeeper, fee *Fee, denom string, msgID string) (sdk.Int, error) {
	if fee == nil {
		return sdk.ZeroInt(), nil
	}

	coin := sdk.NewCoin(denom, funcs.MustOk(sdk.NewIntFromString(fee.Amount)))
	recipient := funcs.Must(sdk.AccAddressFromBech32(fee.Recipient))
	events.Emit(ctx, &types.FeePaid{
		MessageID: msgID,
		Recipient: recipient,
		Fee:       coin,
	})

	return coin.Amount, b.SendCoins(ctx, types.AxelarGMPAccount, recipient, sdk.NewCoins(coin))
}
