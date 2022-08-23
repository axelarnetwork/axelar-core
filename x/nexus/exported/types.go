package exported

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
)

// AddressValidator defines a function that implements address verification upon a request to link addresses
type AddressValidator func(ctx sdk.Context, address CrossChainAddress) error

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

// ChainState allows to record vote status of chain maintainers
type ChainState interface {
	codec.ProtoMarshaler
	MarkMissingVote(maintainer sdk.ValAddress, missingVote bool)
	MarkIncorrectVote(maintainer sdk.ValAddress, incorrectVote bool)
	ChainName() ChainName
}
