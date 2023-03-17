package axelarnet

import (
	"crypto/sha256"
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
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

// Message is attached in ICS20 packet memo field
type Message struct {
	DestinationChain   string `json:"destination_chain"`
	DestinationAddress string `json:"destination_address"`
	Payload            []byte `json:"payload"`
	Type               int64  `json:"type"`
}

func validateMessage(ctx sdk.Context, ibcK keeper.IBCKeeper, n types.Nexus, ibcPath string, msg Message, token keeper.Coin) error {
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
	if data.GetReceiver() != types.MessageSender.String() {
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

	if err := validateMessage(ctx, ibcK, n, path, msg, token); err != nil {
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
		err = handleMessage(ctx, n, sourceAddress, msg)
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

func handleMessage(ctx sdk.Context, n types.Nexus, sourceAddress nexus.CrossChainAddress, msg Message) error {
	destChain := funcs.MustOk(n.GetChain(ctx, nexus.ChainName(msg.DestinationChain)))

	recipient := nexus.CrossChainAddress{Chain: destChain, Address: msg.DestinationAddress}
	txHash := sha256.Sum256(ctx.TxBytes())
	m := nexus.NewGeneralMessage(
		n.GenerateMessageID(ctx, txHash[:]),
		sourceAddress,
		recipient,
		crypto.Keccak256Hash(msg.Payload).Bytes(),
		nexus.Approved,
		nil,
		txHash[:],
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

func handleMessageWithToken(ctx sdk.Context, n types.Nexus, b types.BankKeeper, sourceAddress nexus.CrossChainAddress, msg Message, asset keeper.Coin) error {
	destChain := funcs.MustOk(n.GetChain(ctx, nexus.ChainName(msg.DestinationChain)))

	if err := asset.Lock(b, types.MessageSender); err != nil {
		return err
	}

	recipient := nexus.CrossChainAddress{Chain: destChain, Address: msg.DestinationAddress}
	txHash := sha256.Sum256(ctx.TxBytes())
	m := nexus.NewGeneralMessage(
		n.GenerateMessageID(ctx, txHash[:]),
		sourceAddress,
		recipient,
		crypto.Keccak256Hash(msg.Payload).Bytes(),
		nexus.Approved,
		&asset.Coin,
		txHash[:],
	)

	events.Emit(ctx, &types.ContractCallWithTokenSubmitted{
		MessageID:        m.ID,
		Sender:           m.GetSourceAddress(),
		SourceChain:      m.GetSourceChain(),
		DestinationChain: m.GetDestinationChain(),
		ContractAddress:  m.GetDestinationAddress(),
		PayloadHash:      m.PayloadHash,
		Payload:          msg.Payload,
		Asset:            asset.Coin,
	})

	return n.SetNewMessage(ctx, m)
}

func handleTokenSent(ctx sdk.Context, n types.Nexus, b types.BankKeeper, sourceAddress nexus.CrossChainAddress, msg Message, asset keeper.Coin) error {
	destChain := funcs.MustOk(n.GetChain(ctx, nexus.ChainName(msg.DestinationChain)))
	crossChainAddr := nexus.CrossChainAddress{Chain: destChain, Address: msg.DestinationAddress}

	if err := asset.Lock(b, types.MessageSender); err != nil {
		return err
	}

	transferID, err := n.EnqueueTransfer(ctx, sourceAddress.Chain, crossChainAddr, asset.Coin)
	if err != nil {
		return err
	}

	events.Emit(ctx, &types.TokenSent{
		TransferID:         transferID,
		Sender:             sourceAddress.Address,
		SourceChain:        sourceAddress.Chain.Name,
		DestinationAddress: crossChainAddr.Address,
		DestinationChain:   crossChainAddr.Chain.Name,
		Asset:              asset.Coin,
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
