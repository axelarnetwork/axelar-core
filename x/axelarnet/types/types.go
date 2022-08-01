package types

import (
	"crypto/sha256"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	host "github.com/cosmos/ibc-go/v2/modules/core/24-host"
	ibc "github.com/cosmos/ibc-go/v2/modules/core/exported"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewLinkedAddress creates a new address to make a deposit which can be transferred to another blockchain
func NewLinkedAddress(ctx sdk.Context, chain nexus.ChainName, symbol, recipientAddr string) sdk.AccAddress {
	nonce := utils.GetNonce(ctx.HeaderHash(), ctx.BlockGasMeter())
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s_%s_%s_%x", chain, symbol, recipientAddr, nonce)))
	return hash[:address.Len]
}

// GetEscrowAddress creates an address for an ibc denomination
func GetEscrowAddress(denom string) sdk.AccAddress {
	hash := sha256.Sum256([]byte(denom))
	return hash[:address.Len]
}

// Validate checks the stateless validity of the transfer
func (m IBCTransfer) Validate() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return err
	}

	if err := utils.ValidateString(m.PortID); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}

	if err := utils.ValidateString(m.ChannelID); err != nil {
		return sdkerrors.Wrap(err, "invalid channel ID")
	}

	if err := utils.ValidateString(m.Receiver); err != nil {
		return sdkerrors.Wrap(err, "invalid receiver")
	}

	if err := m.Token.Validate(); err != nil {
		return err
	}

	return nil
}

type sortedChains []CosmosChain

func (s sortedChains) Len() int {
	return len(s)
}

func (s sortedChains) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s sortedChains) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// SortChains sorts the given slice
func SortChains(chains []CosmosChain) {
	sort.Stable(sortedChains(chains))
}

type sortedTransfers []IBCTransfer

func (s sortedTransfers) Len() int {
	return len(s)
}

func (s sortedTransfers) Less(i, j int) bool {
	return s[i].String() < s[j].String()
}

func (s sortedTransfers) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// SortTransfers sorts the given slice
func SortTransfers(transfers []IBCTransfer) {
	sort.Stable(sortedTransfers(transfers))
}

// Validate checks the stateless validity of the cosmos chain
func (m CosmosChain) Validate() error {
	if m.Name.Equals(exported.Axelarnet.Name) {
		if m.IBCPath != "" {
			return fmt.Errorf("IBC path should be empty for %s", exported.Axelarnet.Name)
		}
	} else {
		if err := utils.ValidateString(m.IBCPath); err != nil {
			return sdkerrors.Wrap(err, "invalid IBC path")
		}
	}

	if err := m.Name.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid name")
	}

	if err := utils.ValidateString(m.AddrPrefix); err != nil {
		return sdkerrors.Wrap(err, "invalid address prefix")
	}

	return nil
}

// NewIBCTransfer creates an IBC transfer
func NewIBCTransfer(sender sdk.AccAddress, receiver string, token sdk.Coin, portID string, channelID string) IBCTransfer {
	return IBCTransfer{
		Sender:    sender,
		Receiver:  receiver,
		Token:     token,
		PortID:    portID,
		ChannelID: channelID,
	}
}

// SetID set the ID for IBCTransfer
func (m *IBCTransfer) SetID(id nexus.TransferID) {
	m.ID = id
}

// ValidateBasic returns an error if the given IBCTransfer is invalid; nil otherwise
func (m IBCTransfer) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(err, "invalid transfer sender")
	}

	if err := utils.ValidateString(m.Receiver); err != nil {
		return sdkerrors.Wrap(err, "invalid transfer receiver")
	}

	if err := m.Token.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid token")
	}

	if err := host.PortIdentifierValidator(m.PortID); err != nil {
		return sdkerrors.Wrap(err, "invalid source port ID")
	}

	if err := host.ChannelIdentifierValidator(m.ChannelID); err != nil {
		return sdkerrors.Wrap(err, "invalid source channel ID")
	}

	return nil
}

// PacketToTransfer converts IBC packet to IBC transfer
func PacketToTransfer(packet ibc.PacketI) (IBCTransfer, error) {
	var data types.FungibleTokenPacketData
	if err := ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return IBCTransfer{}, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}
	sender, err := sdk.AccAddressFromBech32(data.Sender)
	if err != nil {
		return IBCTransfer{}, err
	}

	// parse the denomination from the full denom path
	trace := types.ParseDenomTrace(data.Denom)

	// parse the transfer amount
	transferAmount, ok := sdk.NewIntFromString(data.Amount)
	if !ok {
		return IBCTransfer{}, sdkerrors.Wrapf(types.ErrInvalidAmount, "unable to parse transfer amount (%s) into sdk.Int", data.Amount)
	}
	token := sdk.NewCoin(trace.IBCDenom(), transferAmount)

	return NewIBCTransfer(sender, data.Receiver, token, packet.GetSourcePort(), packet.GetSourceChannel()), nil
}
