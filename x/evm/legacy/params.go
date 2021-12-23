package legacy

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
)

// Params wraps the current Params and implements ParamSetPairs for v0.13.x
type Params struct{ types.Params }

// ParamSetPairs implements ParamSetPairs for v0.13.x
func (m *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(types.KeyChain, &m.Chain, validateChain),
		params.NewParamSetPair(types.KeyConfirmationHeight, &m.ConfirmationHeight, validateConfirmationHeight),
		params.NewParamSetPair(types.KeyNetwork, &m.Network, validateNetwork),
		params.NewParamSetPair(types.KeyGateway, &m.GatewayCode, validateBytes),
		params.NewParamSetPair(types.KeyToken, &m.TokenCode, validateBytes),
		params.NewParamSetPair(types.KeyBurnable, &m.Burnable, validateBytes),
		params.NewParamSetPair(types.KeyRevoteLockingPeriod, &m.RevoteLockingPeriod, validateRevoteLockingPeriod),
		params.NewParamSetPair(types.KeyNetworks, &m.Networks, validateNetworks),
		params.NewParamSetPair(types.KeyVotingThreshold, &m.VotingThreshold, validateVotingThreshold),
		params.NewParamSetPair(types.KeyMinVoterCount, &m.MinVoterCount, validateMinVoterCount),
		params.NewParamSetPair(types.KeyCommandsGasLimit, &m.CommandsGasLimit, validateCommandsGasLimit),
		params.NewParamSetPair(types.KeyTransactionFeeRate, &m.TransactionFeeRate, validateTransactionFeeRate),
	}
}

func validateChain(chain interface{}) error {
	c, ok := chain.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type for chain: %T", chain)
	}
	if c == "" {
		return sdkerrors.Wrap(govtypes.ErrInvalidGenesis, "chain name cannot be an empty string")
	}
	return nil
}

func validateNetwork(network interface{}) error {
	n, ok := network.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type for network: %T", network)
	}
	if n == "" {
		return sdkerrors.Wrap(govtypes.ErrInvalidGenesis, "network name cannot be an empty string")
	}
	return nil
}

func validateConfirmationHeight(height interface{}) error {
	h, ok := height.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type for confirmation height: %T", height)
	}
	if h < 1 {
		return sdkerrors.Wrap(govtypes.ErrInvalidGenesis, "transaction confirmation height must be greater than 0")
	}
	return nil
}

func validateBytes(bytes interface{}) error {
	b, ok := bytes.([]byte)
	if !ok {
		return fmt.Errorf("invalid parameter type for byte slice: %T", bytes)
	}

	if len(b) == 0 {
		return fmt.Errorf("byte slice cannot be empty")
	}

	return nil
}

func validateRevoteLockingPeriod(RevoteLockingPeriod interface{}) error {
	r, ok := RevoteLockingPeriod.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type for revote lock period: %T", r)
	}

	if r <= 0 {
		return sdkerrors.Wrap(govtypes.ErrInvalidGenesis, "revote lock period be greater than 0")
	}

	return nil
}

func validateNetworks(network interface{}) error {
	networks, ok := network.([]types.NetworkInfo)
	if !ok {
		return fmt.Errorf("invalid parameter type for networks: %T", network)
	}
	for _, n := range networks {
		if n.Name == "" {
			return sdkerrors.Wrap(govtypes.ErrInvalidGenesis, "network name cannot be an empty string")
		}
	}

	return nil
}

func validateVotingThreshold(votingThreshold interface{}) error {
	val, ok := votingThreshold.(utils.Threshold)
	if !ok {
		return fmt.Errorf("invalid parameter type for VotingThreshold: %T", votingThreshold)
	}

	if err := val.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid VotingThreshold")
	}

	return nil
}

func validateMinVoterCount(minVoterCount interface{}) error {
	val, ok := minVoterCount.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type for MinVoterCount: %T", minVoterCount)
	}

	if val < 0 {
		return fmt.Errorf("min voter count must be >=0")
	}

	return nil
}

func validateCommandsGasLimit(commandsGasLimit interface{}) error {
	val, ok := commandsGasLimit.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type for commands gas limit: %T", commandsGasLimit)
	}

	if val <= 0 {
		return fmt.Errorf("commands gas limit must be >0")
	}

	return nil
}

func validateTransactionFeeRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("transaction fee rate must be positive: %s", v)
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("transaction fee rate %s must be <= 1", v)
	}

	return nil
}
