package types

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	gethParams "github.com/ethereum/go-ethereum/params"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
)

// Parameter keys
var (
	KeyChain               = []byte("chain")
	KeyConfirmationHeight  = []byte("confirmationHeight")
	KeyNetwork             = []byte("network")
	KeyRevoteLockingPeriod = []byte("revoteLockingPeriod")
	KeyNetworks            = []byte("networks")
	KeyVotingThreshold     = []byte("votingThreshold")
	KeyGateway             = []byte("gateway")
	KeyToken               = []byte("token")
	KeyBurnable            = []byte("burneable")
	KeyMinVoterCount       = []byte("minVoterCount")
	KeyCommandsGasLimit    = []byte("commandsGasLimit")
)

// KeyTable returns a subspace.KeyTable that has registered all parameter types in this module's parameter set
func KeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns the module's parameter set initialized with default values
func DefaultParams() []Params {
	bzGateway, err := hex.DecodeString(gateway)
	if err != nil {
		panic(err)
	}
	bzToken, err := hex.DecodeString(token)
	if err != nil {
		panic(err)
	}
	bzBurnable, err := hex.DecodeString(burnable)
	if err != nil {
		panic(err)
	}

	return []Params{{
		Chain:               exported.Ethereum.Name,
		ConfirmationHeight:  1,
		Network:             Ganache,
		Gateway:             bzGateway,
		Token:               bzToken,
		Burnable:            bzBurnable,
		RevoteLockingPeriod: 50,
		Networks: []NetworkInfo{
			{
				Name: Mainnet,
				Id:   sdk.NewIntFromBigInt(gethParams.MainnetChainConfig.ChainID),
			},
			{
				Name: Ropsten,
				Id:   sdk.NewIntFromBigInt(gethParams.RopstenChainConfig.ChainID),
			},
			{
				Name: Rinkeby,
				Id:   sdk.NewIntFromBigInt(gethParams.RinkebyChainConfig.ChainID),
			},
			{
				Name: Goerli,
				Id:   sdk.NewIntFromBigInt(gethParams.GoerliChainConfig.ChainID),
			},
			{
				Name: Ganache,
				Id:   sdk.NewIntFromBigInt(gethParams.AllCliqueProtocolChanges.ChainID),
			},
		},
		VotingThreshold:  utils.Threshold{Numerator: 15, Denominator: 100},
		MinVoterCount:    1,
		CommandsGasLimit: 5000000,
	}}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of tss module's parameters.
func (m *Params) ParamSetPairs() params.ParamSetPairs {
	/*
		because the subspace package makes liberal use of pointers to set and get values from the store,
		this method needs to have a pointer receiver AND NewParamSetPair needs to receive the
		parameter values as pointer arguments, otherwise either the internal type reflection panics or the value will not be
		set on the correct Params data struct
	*/
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyChain, &m.Chain, validateChain),
		params.NewParamSetPair(KeyConfirmationHeight, &m.ConfirmationHeight, validateConfirmationHeight),
		params.NewParamSetPair(KeyNetwork, &m.Network, validateNetwork),
		params.NewParamSetPair(KeyGateway, &m.Gateway, validateBytes),
		params.NewParamSetPair(KeyToken, &m.Token, validateBytes),
		params.NewParamSetPair(KeyBurnable, &m.Burnable, validateBytes),
		params.NewParamSetPair(KeyRevoteLockingPeriod, &m.RevoteLockingPeriod, validateRevoteLockingPeriod),
		params.NewParamSetPair(KeyNetworks, &m.Networks, validateNetworks),
		params.NewParamSetPair(KeyVotingThreshold, &m.VotingThreshold, validateVotingThreshold),
		params.NewParamSetPair(KeyMinVoterCount, &m.MinVoterCount, validateMinVoterCount),
		params.NewParamSetPair(KeyCommandsGasLimit, &m.CommandsGasLimit, validateCommandsGasLimit),
	}
}

func validateChain(chain interface{}) error {
	c, ok := chain.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type for chain: %T", chain)
	}
	if c == "" {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "chain name cannot be an empty string")
	}
	return nil
}

func validateNetwork(network interface{}) error {
	n, ok := network.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type for network: %T", network)
	}
	if n == "" {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "network name cannot be an empty string")
	}
	return nil
}

func validateConfirmationHeight(height interface{}) error {
	h, ok := height.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type for confirmation height: %T", height)
	}
	if h < 1 {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "transaction confirmation height must be greater than 0")
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
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "revote lock period be greater than 0")
	}

	return nil
}

func validateNetworks(network interface{}) error {
	networks, ok := network.([]NetworkInfo)
	if !ok {
		return fmt.Errorf("invalid parameter type for networks: %T", network)
	}
	for _, n := range networks {
		if n.Name == "" {
			return sdkerrors.Wrap(types.ErrInvalidGenesis, "network name cannot be an empty string")
		}
	}

	return nil
}

func validateVotingThreshold(votingThreshold interface{}) error {
	val, ok := votingThreshold.(utils.Threshold)
	if !ok {
		return fmt.Errorf("invalid parameter type for VotingThreshold: %T", votingThreshold)
	}

	if val.Numerator <= 0 {
		return fmt.Errorf("threshold numerator must be a positive integer for VotingThreshold")
	}

	if val.Denominator <= 0 {
		return fmt.Errorf("threshold denominator must be a positive integer for VotingThreshold")
	}

	if val.Numerator > val.Denominator {
		return fmt.Errorf("threshold must be <=1 for VotingThreshold")
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

// Validate checks the validity of the values of the parameter set
func (m Params) Validate() error {
	if err := validateConfirmationHeight(m.ConfirmationHeight); err != nil {
		return err
	}

	if err := validateNetwork(m.Network); err != nil {
		return err
	}

	if err := validateRevoteLockingPeriod(m.RevoteLockingPeriod); err != nil {
		return err
	}

	if err := validateVotingThreshold(m.VotingThreshold); err != nil {
		return err
	}

	if err := validateMinVoterCount(m.MinVoterCount); err != nil {
		return err
	}

	if err := validateCommandsGasLimit(m.CommandsGasLimit); err != nil {
		return err
	}

	// ensure that the network is one of the supported ones
	for _, n := range m.Networks {
		if n.Name == m.Network {
			return nil
		}
	}

	return fmt.Errorf("'%s' not part of the network list", m.Network)
}
