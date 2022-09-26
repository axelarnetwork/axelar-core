package types

import (
	fmt "fmt"
	"reflect"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/slices"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// EventID ensures a correctly formatted event ID
type EventID string

func newEventID(txID Hash, index uint64) EventID {
	return EventID(fmt.Sprintf("%s-%d", txID.Hex(), index))
}

// Validate returns an error, if the event ID is not in format of txID-index
func (id EventID) Validate() error {
	if err := utils.ValidateString(string(id)); err != nil {
		return err
	}

	arr := strings.Split(string(id), "-")
	if len(arr) != 2 {
		return fmt.Errorf("event ID should be in foramt of txID-index")
	}

	bz, err := hexutil.Decode(arr[0])
	if err != nil {
		return sdkerrors.Wrap(err, "invalid tx hash hex encoding")
	}

	if len(bz) != common.HashLength {
		return fmt.Errorf("invalid tx hash length")
	}

	_, err = strconv.ParseInt(arr[1], 10, 64)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid index")
	}

	return nil
}

// GetID returns an unique ID for the event
func (m Event) GetID() EventID {
	return newEventID(m.TxID, m.Index)
}

// GetEventType returns the type forz the event
func (m Event) GetEventType() string {
	return getType(m.GetEvent())
}

func getType(val interface{}) string {
	t := reflect.TypeOf(val)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// ValidateBasic returns an error if the event is invalid
func (m Event) ValidateBasic() error {
	if err := m.Chain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid source chain")
	}

	if m.TxID.IsZero() {
		return fmt.Errorf("invalid tx id")
	}

	switch event := m.GetEvent().(type) {
	case *Event_ContractCall:
		if event.ContractCall == nil {
			return fmt.Errorf("missing event ContractCall")
		}

		if err := event.ContractCall.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "invalid event ContractCall")
		}
	case *Event_ContractCallWithToken:
		if event.ContractCallWithToken == nil {
			return fmt.Errorf("missing event ContractCallWithToken")
		}

		if err := event.ContractCallWithToken.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "invalid event ContractCallWithToken")
		}
	case *Event_TokenSent:
		if event.TokenSent == nil {
			return fmt.Errorf("missing event TokenSent")
		}

		if err := event.TokenSent.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "invalid event TokenSent")
		}
	case *Event_Transfer:
		if event.Transfer == nil {
			return fmt.Errorf("missing event Transfer")
		}

		if err := event.Transfer.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "invalid event Transfer")
		}
	case *Event_TokenDeployed:
		if event.TokenDeployed == nil {
			return fmt.Errorf("missing event TokenDeployed")
		}

		if err := event.TokenDeployed.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "invalid event TokenDeployed")
		}
	case *Event_MultisigOperatorshipTransferred:
		if event.MultisigOperatorshipTransferred == nil {
			return fmt.Errorf("missing event MultisigOperatorshipTransferred")
		}

		if err := event.MultisigOperatorshipTransferred.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "invalid event MultisigOperatorshipTransferred")
		}
	default:
		return fmt.Errorf("unknown type of event")
	}

	return nil
}

// ValidateBasic returns an error if the event token sent is invalid
func (m EventTokenSent) ValidateBasic() error {
	if m.Sender.IsZeroAddress() {
		return fmt.Errorf("invalid sender")
	}

	if err := m.DestinationChain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid destination chain")
	}

	if err := utils.ValidateString(m.DestinationAddress); err != nil {
		return sdkerrors.Wrap(err, "invalid destination address")
	}

	if err := utils.ValidateString(m.Symbol); err != nil {
		return sdkerrors.Wrap(err, "invalid symbol")
	}

	if m.Amount.IsZero() {
		return fmt.Errorf("invalid amount")
	}

	return nil
}

// ValidateBasic returns an error if the event contract call is invalid
func (m EventContractCall) ValidateBasic() error {
	if m.Sender.IsZeroAddress() {
		return fmt.Errorf("invalid sender")
	}

	if err := m.DestinationChain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid destination chain")
	}

	if !common.IsHexAddress(m.ContractAddress) {
		return fmt.Errorf("invalid contract address")
	}

	if m.PayloadHash.IsZero() {
		return fmt.Errorf("invalid payload hash")
	}

	return nil
}

// ValidateBasic returns an error if the event contract call with token is invalid
func (m EventContractCallWithToken) ValidateBasic() error {
	if m.Sender.IsZeroAddress() {
		return fmt.Errorf("invalid sender")
	}

	if err := m.DestinationChain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid destination chain")
	}

	if !common.IsHexAddress(m.ContractAddress) {
		return fmt.Errorf("invalid contract address")
	}

	if m.PayloadHash.IsZero() {
		return fmt.Errorf("invalid payload hash")
	}

	if err := utils.ValidateString(m.Symbol); err != nil {
		return sdkerrors.Wrap(err, "invalid symbol")
	}

	if m.Amount.IsZero() {
		return fmt.Errorf("invalid amount")
	}

	return nil
}

// ValidateBasic returns an error if the event transfer is invalid
func (m EventTransfer) ValidateBasic() error {
	if m.To.IsZeroAddress() {
		return fmt.Errorf("invalid sender")
	}

	if m.Amount.IsZero() {
		return fmt.Errorf("invalid amount")
	}

	return nil
}

// ValidateBasic returns an error if the event token deployed is invalid
func (m EventTokenDeployed) ValidateBasic() error {
	if m.TokenAddress.IsZeroAddress() {
		return fmt.Errorf("invalid sender")
	}

	if err := utils.ValidateString(m.Symbol); err != nil {
		return sdkerrors.Wrap(err, "invalid symbol")
	}

	return nil
}

// ValidateBasic returns an error if the event multisig operatorship transferred is invalid
func (m EventMultisigOperatorshipTransferred) ValidateBasic() error {
	if slices.Any(m.NewOperators, Address.IsZeroAddress) {
		return fmt.Errorf("invalid new operators")
	}

	if len(m.NewOperators) != len(m.NewWeights) {
		return fmt.Errorf("length of new operators does not match new weights")
	}

	totalWeight := sdk.ZeroUint()
	slices.ForEach(m.NewWeights, func(w sdk.Uint) { totalWeight = totalWeight.Add(w) })

	if m.NewThreshold.IsZero() || m.NewThreshold.GT(totalWeight) {
		return fmt.Errorf("invalid new threshold")
	}

	return nil
}

// NewVoteEvents is the constructor for vote events
func NewVoteEvents(chain nexus.ChainName, events ...Event) *VoteEvents {
	return &VoteEvents{
		Chain:  chain,
		Events: events,
	}
}

// ValidateBasic does stateless validation of the object
func (m VoteEvents) ValidateBasic() error {
	if err := m.Chain.Validate(); err != nil {
		return err
	}

	for _, event := range m.Events {
		if err := event.ValidateBasic(); err != nil {
			return err
		}

		if event.Chain != m.Chain {
			return fmt.Errorf("events are not from the same source chain")
		}
	}

	return nil
}
