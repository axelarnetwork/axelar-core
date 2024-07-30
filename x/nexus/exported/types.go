package exported

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/utils/slices"
)

//go:generate moq -out ./mock/types.go -pkg mock . MaintainerState

// AddressValidator defines a function that implements address verification upon a request to link addresses
type AddressValidator func(ctx sdk.Context, address CrossChainAddress) error

type RoutingContext struct {
	Sender     sdk.AccAddress
	FeeGranter sdk.AccAddress
	Payload    []byte
}

// MessageRoute defines a function that implements message routing
type MessageRoute func(ctx sdk.Context, routingCtx RoutingContext, msg GeneralMessage) error

// TransferStateFromString converts a describing state string to the corresponding TransferState
func TransferStateFromString(s string) TransferState {
	state, ok := TransferState_value["TRANSFER_STATE_"+strings.ToUpper(s)]

	if !ok {
		return TRANSFER_STATE_UNSPECIFIED
	}

	return TransferState(state)
}

// Validate validates the TransferState
func (m TransferState) Validate() error {
	_, ok := TransferState_name[int32(m)]
	if !ok {
		return fmt.Errorf("unknown transfer state")
	}

	if m == TRANSFER_STATE_UNSPECIFIED {
		return fmt.Errorf("unspecified transfer state")
	}

	return nil
}

// Validate validates the CrossChainTransfer
func (m CrossChainTransfer) Validate() error {
	if err := m.Recipient.Validate(); err != nil {
		return err
	}

	if err := m.Asset.Validate(); err != nil {
		return err
	}

	if err := m.State.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate validates the CrossChainAddress
func (m CrossChainAddress) Validate() error {
	if err := m.Chain.Validate(); err != nil {
		return err
	}

	if err := utils.ValidateString(m.Address); err != nil {
		return sdkerrors.Wrap(err, "invalid address")
	}

	return nil
}

// TransferID represents the unique cross transfer identifier
type TransferID uint64

// String returns a string representation of TransferID
func (t TransferID) String() string {
	return strconv.FormatUint(uint64(t), 10)
}

// Bytes returns the byte array of TransferID
func (t TransferID) Bytes() []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(t))

	return bz
}

// NewPendingCrossChainTransfer returns a pending CrossChainTransfer
func NewPendingCrossChainTransfer(id uint64, recipient CrossChainAddress, asset sdk.Coin) CrossChainTransfer {
	return NewCrossChainTransfer(id, recipient, asset, Pending)
}

// NewCrossChainTransfer returns a CrossChainTransfer
func NewCrossChainTransfer(id uint64, recipient CrossChainAddress, asset sdk.Coin, state TransferState) CrossChainTransfer {
	return CrossChainTransfer{
		ID:        TransferID(id),
		Recipient: recipient,
		Asset:     asset,
		State:     state,
	}
}

// Validate performs a stateless check to ensure the Chain object has been initialized correctly
func (m Chain) Validate() error {
	if err := m.Name.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid chain name")
	}

	if err := m.KeyType.Validate(); err != nil {
		return err
	}

	if m.Module == "" {
		return fmt.Errorf("missing module name")
	}

	return nil
}

// GetName returns the chain name
func (m Chain) GetName() ChainName {
	return m.Name
}

// IsFrom returns true if the chain registered under the module
func (m Chain) IsFrom(module string) bool {
	return m.Module == module
}

// NewAsset returns an asset struct
func NewAsset(denom string, isNative bool) Asset {
	return Asset{Denom: utils.NormalizeString(denom), IsNativeAsset: isNative}
}

// Validate checks the stateless validity of the asset
func (m Asset) Validate() error {
	if err := sdk.ValidateDenom(m.Denom); err != nil {
		return sdkerrors.Wrap(err, "invalid denomination")
	}

	return nil
}

// NewFeeInfo returns a FeeInfo struct
func NewFeeInfo(chain ChainName, asset string, feeRate sdk.Dec, minFee sdk.Int, maxFee sdk.Int) FeeInfo {
	asset = utils.NormalizeString(asset)

	return FeeInfo{Chain: chain, Asset: asset, FeeRate: feeRate, MinFee: minFee, MaxFee: maxFee}
}

// ZeroFeeInfo returns a FeeInfo struct with zero fees
func ZeroFeeInfo(chain ChainName, asset string) FeeInfo {
	return NewFeeInfo(chain, asset, sdk.ZeroDec(), sdk.ZeroInt(), sdk.ZeroInt())
}

// Validate checks the stateless validity of fee info
func (m FeeInfo) Validate() error {
	if err := m.Chain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	if err := sdk.ValidateDenom(m.Asset); err != nil {
		return sdkerrors.Wrap(err, "invalid asset")
	}

	if m.MinFee.IsNegative() {
		return fmt.Errorf("min fee cannot be negative")
	}

	if m.MinFee.GT(m.MaxFee) {
		return fmt.Errorf("min fee should not be greater than max fee")
	}

	if m.FeeRate.IsNegative() {
		return fmt.Errorf("fee rate should not be negative")
	}

	if m.FeeRate.GT(sdk.OneDec()) {
		return fmt.Errorf("fee rate should not be greater than one")
	}

	if !m.FeeRate.IsZero() && m.MaxFee.IsZero() {
		return fmt.Errorf("fee rate is non zero while max fee is zero")
	}

	return nil
}

// ChainNameLengthMax bounds the max chain name length
const ChainNameLengthMax = 20

// ChainName ensures a correctly formatted EVM chain name
type ChainName string

// Validate returns an error, if the chain name is empty or too long
func (c ChainName) Validate() error {
	if err := utils.ValidateString(string(c)); err != nil {
		return sdkerrors.Wrap(err, "invalid chain name")
	}

	if len(c) > ChainNameLengthMax {
		return fmt.Errorf("chain name length %d is greater than %d", len(c), ChainNameLengthMax)
	}

	return nil
}

func (c ChainName) String() string {
	return string(c)
}

// Equals returns boolean for whether two chain names are case-insensitive equal
func (c ChainName) Equals(c2 ChainName) bool {
	return strings.EqualFold(c.String(), c2.String())
}

// MaintainerState allows to record status of chain maintainer
type MaintainerState interface {
	codec.ProtoMarshaler
	MarkMissingVote(missingVote bool)
	MarkIncorrectVote(incorrectVote bool)
	CountMissingVotes(window int) uint64
	CountIncorrectVotes(window int) uint64
	GetAddress() sdk.ValAddress
}

// ValidateBasic validates the transfer direction
func (m TransferDirection) ValidateBasic() error {
	switch m {
	case TransferDirectionFrom, TransferDirectionTo:
		return nil
	default:
		return fmt.Errorf("invalid transfer direction %s", m)
	}
}

// NewGeneralMessage returns a GeneralMessage struct with status set to approved
func NewGeneralMessage(id string, sender CrossChainAddress, recipient CrossChainAddress, payloadHash []byte, sourceTxID []byte, sourceTxIndex uint64, asset *sdk.Coin) GeneralMessage {
	return GeneralMessage{
		ID:            id,
		Sender:        sender,
		Recipient:     recipient,
		PayloadHash:   payloadHash,
		Status:        Approved,
		Asset:         asset,
		SourceTxID:    sourceTxID,
		SourceTxIndex: sourceTxIndex,
	}
}

// ValidateBasic validates the general message
func (m GeneralMessage) ValidateBasic() error {
	if err := utils.ValidateString(m.ID); err != nil {
		return sdkerrors.Wrap(err, "invalid general message id")
	}

	if err := m.Sender.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid source chain")
	}

	if err := m.Recipient.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid destination chain")
	}

	if m.Asset != nil {
		return m.Asset.Validate()
	}

	return nil
}

// Is returns true if status matches
func (m GeneralMessage) Is(status GeneralMessage_Status) bool {
	return m.Status == status
}

// Match returns true if hash of payload matches the expected
func (m GeneralMessage) Match(payload []byte) bool {
	return common.BytesToHash(m.PayloadHash) == crypto.Keccak256Hash(payload)
}

// GetSourceChain returns the source chain name
func (m GeneralMessage) GetSourceChain() ChainName {
	return m.Sender.Chain.Name
}

// GetSourceAddress returns the source address
func (m GeneralMessage) GetSourceAddress() string {
	return m.Sender.Address
}

// GetDestinationChain returns the destination chain name
func (m GeneralMessage) GetDestinationChain() ChainName {
	return m.Recipient.Chain.Name
}

// GetDestinationAddress returns the destination address
func (m GeneralMessage) GetDestinationAddress() string {
	return m.Recipient.Address
}

// MessageType on can be TypeGeneralMessage or TypeGeneralMessageWithToken
type MessageType int

const (
	// TypeUnrecognized means coin type is unrecognized
	TypeUnrecognized = iota
	// TypeGeneralMessage is a pure message
	TypeGeneralMessage
	// TypeGeneralMessageWithToken is a general message with token
	TypeGeneralMessageWithToken
	// TypeSendToken is a direct token transfer without link from a cosmos chain
	TypeSendToken
)

// Type returns the type of the message
func (m GeneralMessage) Type() MessageType {
	if m.Asset == nil {
		return TypeGeneralMessage
	}

	return TypeGeneralMessageWithToken
}

// FromGeneralMessage returns a WasmMessage from a GeneralMessage
func FromGeneralMessage(msg GeneralMessage) WasmMessage {
	return WasmMessage{
		SourceChain:        msg.GetSourceChain(),
		SourceAddress:      msg.GetSourceAddress(),
		DestinationChain:   msg.GetDestinationChain(),
		DestinationAddress: msg.GetDestinationAddress(),
		PayloadHash:        msg.PayloadHash,
		SourceTxID:         msg.SourceTxID,
		SourceTxIndex:      msg.SourceTxIndex,
		ID:                 msg.ID,
	}
}

var _ sdk.Msg = &WasmMessage{}

// ValidateBasic implements sdk.Msg
func (m WasmMessage) ValidateBasic() error {

	if err := utils.ValidateString(m.ID); err != nil {
		return sdkerrors.Wrap(err, "invalid wasm message id")
	}

	if err := m.SourceChain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid wasm message source chain name")
	}

	if err := utils.ValidateString(m.SourceAddress); err != nil {
		return sdkerrors.Wrap(err, "invalid wasm message source address")
	}

	if err := m.DestinationChain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid wasm message destination chain name")
	}

	if err := utils.ValidateString(m.DestinationAddress); err != nil {
		return sdkerrors.Wrap(err, "invalid wasm message destination address")
	}

	if len(m.PayloadHash) != 32 {
		return fmt.Errorf("invalid wasm message payload hash")
	}

	if len(m.SourceTxID) != 32 {
		return fmt.Errorf("invalid wasm message source tx id")
	}

	return nil
}

// GetSigners implements sdk.Msg. There is no signer for wasm generated messages, so this returns an empty slice.
func (m WasmMessage) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}

// WasmBytes is a wrapper around []byte that gets JSON marshalized as an array
// of numbers instead of base64-encoded string
type WasmBytes []byte

// MarshalJSON implements json.Marshaler
func (bz WasmBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(slices.Map(bz, func(b byte) uint16 { return uint16(b) }))
}

// UnmarshalJSON implements json.Unmarshaler
func (bz *WasmBytes) UnmarshalJSON(data []byte) error {
	var arr []uint16
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}

	*bz = slices.Map(arr, func(u uint16) byte { return byte(u) })

	return nil
}
