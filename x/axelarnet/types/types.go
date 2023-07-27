package types

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	host "github.com/cosmos/ibc-go/v4/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v4/modules/core/exported"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// Log attribute keys
const (
	AttributeChain   = "chain"
	AttributeIBCPath = "ibcPath"
)

const (
	// DefaultRateLimitWindow is the default window for rate limits of assets on cosmos chains
	DefaultRateLimitWindow = 6 * time.Hour
)

const ZERO_X_PREFIX = "0x"

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

// ValidateBasic checks the stateless validity of the cosmos chain
func (m CosmosChain) ValidateBasic() error {
	if m.Name.Equals(exported.Axelarnet.Name) {
		if m.IBCPath != "" {
			return fmt.Errorf("IBC path should be empty for %s", exported.Axelarnet.Name)
		}
	} else {
		if err := ValidateIBCPath(m.IBCPath); err != nil {
			return err
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

// NewIBCTransfer creates a new pending IBC transfer
func NewIBCTransfer(sender sdk.AccAddress, receiver string, token sdk.Coin, portID string, channelID string, id nexus.TransferID) IBCTransfer {
	return IBCTransfer{
		Sender:    sender,
		Receiver:  receiver,
		Token:     token,
		PortID:    portID,
		ChannelID: channelID,
		ID:        id,
		Status:    TransferPending,
	}
}

// SetStatus sets the transfer status
func (m *IBCTransfer) SetStatus(status IBCTransfer_Status) error {
	switch status {
	case TransferCompleted, TransferFailed:
		// set from pending to completed or failed
		if m.Status != TransferPending {
			return fmt.Errorf("transfer %s is not pending", m.ID)
		}
	case TransferPending:
		// set from failed to pending
		if m.Status != TransferFailed {
			return fmt.Errorf("transfer %s is not failed", m.ID)
		}
	default:
		return fmt.Errorf("invalid status %s", status)
	}

	m.Status = status
	return nil
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

// CoinType on can be ICS20 token, native asset, or wrapped asset from external chains
type CoinType int

const (
	// Unrecognized means coin type is unrecognized
	Unrecognized = iota
	// Native means native token on Axelarnet
	Native = 1
	// ICS20 means coin from IBC chains
	ICS20 = 2
	// External means from external chains, such as EVM chains
	External = 3
)

// ValidateIBCPath validates direct IBC paths
func ValidateIBCPath(path string) error {
	if err := utils.ValidateString(path); err != nil {
		return sdkerrors.Wrap(err, "invalid IBC path")
	}

	pathValidator := host.NewPathValidator(func(path string) error {
		return nil
	})
	if err := pathValidator(path); err != nil {
		return sdkerrors.Wrap(err, "invalid IBC path")
	}

	// we only support direct IBC connections
	pathSplit := strings.Split(path, "/")
	if len(pathSplit) != 2 {
		return fmt.Errorf(fmt.Sprintf("invalid IBC path %s", path))
	}

	return nil
}

// NewIBCPath returns an IBC path for a given port and IBC channel
func NewIBCPath(port string, channel string) string {
	return fmt.Sprintf("%s/%s", port, channel)
}

// ToICS20Packet unmarshals IBC packet as ICS20 token packet
func ToICS20Packet(packet ibcexported.PacketI) (ibctransfertypes.FungibleTokenPacketData, error) {
	var data ibctransfertypes.FungibleTokenPacketData
	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return ibctransfertypes.FungibleTokenPacketData{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "cannot unmarshal ICS-20 transfer packet data")
	}

	if err := data.ValidateBasic(); err != nil {
		return ibctransfertypes.FungibleTokenPacketData{}, err
	}

	return data, nil
}

const (
	// NativeV1 is the payload version hex indicates send general message to native chain
	NativeV1 = "0x00000000"
	// CosmWasmV1 is the payload version hex indicates send general message to CosmWasm contract
	CosmWasmV1 = "0x00000001"
	// CosmWasmV2 indicates the payload is json encoded
	CosmWasmV2 = "0x00000002"
)

var (
	// AxelarGMPAccount account is the canonical general message sender
	AxelarGMPAccount = GetEscrowAddress(fmt.Sprintf("%s_%s", ModuleName, "gmp"))
)

// ValidateBasic returns an error if the given Fee is invalid; nil otherwise
func (f Fee) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(f.Recipient); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "fee recipient").Error())
	}

	if !f.Amount.IsValid() || !f.Amount.IsPositive() {
		return fmt.Errorf("invalid fee amount")
	}

	if f.RefundRecipient != nil {
		if err := sdk.VerifyAddressFormat(f.RefundRecipient); err != nil {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "fee refund recipient").Error())
		}
	}

	return nil
}

// CallContractProposalMinDeposits is a set of CallContractProposalMinDeposit, one per contract address
type CallContractProposalMinDeposits []CallContractProposalMinDeposit

type callContractProposalMinDepositsMap map[string]map[string]sdk.Coins

// Get returns the minimum deposit for the given chain and contract address
func (m callContractProposalMinDepositsMap) Get(chain nexus.ChainName, contractAddress string) sdk.Coins {
	c := strings.ToLower(chain.String())
	address := strings.ToLower(contractAddress)

	if _, ok := m[c]; !ok {
		return sdk.Coins{}
	}

	return m[c][address]
}

// ValidateBasic returns an error if the type is invalid
func (minDeposits CallContractProposalMinDeposits) ValidateBasic() error {
	chainContractAddressPairs := make(map[string]struct{})

	for _, minDeposit := range minDeposits {
		if err := minDeposit.ValidateBasic(); err != nil {
			return err
		}

		key := fmt.Sprintf("%s%s%s", strings.ToLower(minDeposit.Chain.String()), utils.DefaultDelimiter, strings.ToLower(minDeposit.ContractAddress))
		if _, ok := chainContractAddressPairs[key]; ok {
			return fmt.Errorf("duplicate chain and contract address pair found")
		}
		chainContractAddressPairs[key] = struct{}{}
	}

	return nil
}

// ToMap returns a map of chain name to contract address to min deposit
func (minDeposits CallContractProposalMinDeposits) ToMap(ctx sdk.Context, nexus Nexus) callContractProposalMinDepositsMap {
	minDepositsMap := make(callContractProposalMinDepositsMap)

	for _, minDeposit := range minDeposits {
		chain := strings.ToLower(minDeposit.Chain.String())
		contractAddress := strings.ToLower(minDeposit.ContractAddress)

		if _, ok := minDepositsMap[chain]; !ok {
			minDepositsMap[chain] = make(map[string]sdk.Coins)
		}

		minDepositsMap[chain][contractAddress] = minDeposit.MinDeposits

		// TODO: eventually, this is confusing and bad cuz cosmos addresses will also
		// show up here and be prefixed with 0x. Like the address validator, we should
		// also implement chain-specific address deserializer so that we just use the
		// actual bytes as map keys for this check instead of a string representation.
		if chain, ok := nexus.GetChain(ctx, minDeposit.Chain); !ok || !chain.IsFrom(evmtypes.ModuleName) {
			continue
		}
		if strings.HasPrefix(contractAddress, ZERO_X_PREFIX) {
			minDepositsMap[chain][strings.TrimPrefix(contractAddress, ZERO_X_PREFIX)] = minDeposit.MinDeposits
		} else {
			minDepositsMap[chain][fmt.Sprintf("%s%s", ZERO_X_PREFIX, contractAddress)] = minDeposit.MinDeposits
		}
	}

	return minDepositsMap
}

// ValidateBasic returns an error if the type is invalid
func (minDeposit CallContractProposalMinDeposit) ValidateBasic() error {
	if err := minDeposit.Chain.Validate(); err != nil {
		return err
	}

	if err := utils.ValidateString(minDeposit.ContractAddress); err != nil {
		return err
	}

	if minDeposit.MinDeposits.Empty() {
		return fmt.Errorf("min deposit cannot be empty")
	}

	if err := minDeposit.MinDeposits.Validate(); err != nil {
		return err
	}

	return nil
}
