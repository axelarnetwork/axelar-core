package types

import (
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/axelarnetwork/axelar-core/utils"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// NewLinkedAddress creates a new address to make a deposit which can be transferred to another blockchain
func NewLinkedAddress(ctx sdk.Context, chain, symbol, recipientAddr string) sdk.AccAddress {
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

	if m.PortID == "" {
		return fmt.Errorf("port ID is empty")
	}

	if m.ChannelID == "" {
		return fmt.Errorf("channel ID is empty")
	}

	if m.Receiver == "" {
		return fmt.Errorf("receiver is empty")
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
	for _, chain := range chains {
		sort.Strings(chain.Assets)
	}
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
	if m.IBCPath == "" {
		return fmt.Errorf("IBC path is empty")
	}

	if len(m.Assets) == 0 {
		return fmt.Errorf("chain must contain assets")
	}

	if m.Name == "" {
		return fmt.Errorf("chain must have a name")
	}

	return nil
}
