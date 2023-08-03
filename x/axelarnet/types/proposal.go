package types

import (
	"encoding/hex"
	fmt "fmt"
	"strings"

	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/utils/slices"
)

const (
	ProposalTypeCallContracts string = "CallContracts"
)

func init() {
	gov.RegisterProposalType(ProposalTypeCallContracts)
	gov.RegisterProposalTypeCodec(&CallContractsProposal{}, "axelarnet/CallContractsProposal")
}

// ValidateBasic validates the contract call
func (c ContractCall) ValidateBasic() error {
	if err := c.Chain.Validate(); err != nil {
		return err
	}

	if err := utils.ValidateString(c.ContractAddress); err != nil {
		return err
	}

	if len(c.Payload) == 0 {
		return fmt.Errorf("payload cannot be empty")
	}

	return nil
}

// String returns a human readable string representation of the contract call
func (c ContractCall) String() string {
	return fmt.Sprintf("Chain: %s, Contract Address: %s, Payload: %s", c.Chain.String(), c.ContractAddress, hex.EncodeToString(c.Payload))
}

// Implements Proposal Interface
var _ gov.Content = &CallContractsProposal{}

// NewCallContractsProposal creates a new call contracts proposal
func NewCallContractsProposal(title, description string, contractCalls []ContractCall) gov.Content {
	return &CallContractsProposal{title, description, contractCalls}
}

// GetTitle returns the proposal title
func (p CallContractsProposal) GetTitle() string { return p.Title }

// GetDescription returns the proposal description
func (p CallContractsProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the proposal router key
func (p CallContractsProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the proposal type
func (p CallContractsProposal) ProposalType() string { return ProposalTypeCallContracts }

// ValidateBasic validates the proposal
func (p CallContractsProposal) ValidateBasic() error {
	if err := gov.ValidateAbstract(p); err != nil {
		return err
	}

	if len(p.ContractCalls) == 0 {
		return fmt.Errorf("no contract calls")
	}

	for _, contractCall := range p.ContractCalls {
		if err := contractCall.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}

// String returns a human readable string representation of the proposal
func (p CallContractsProposal) String() string {
	return fmt.Sprintf(`Call Contracts Proposal:
  Title:       %s
  Description: %s
  Contract Calls:
    - %s
`, p.Title, p.Description, strings.Join(slices.Map(p.ContractCalls, ContractCall.String), "\n    - "))
}
