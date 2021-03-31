package types

import (
	"encoding/hex"
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// DefaultParamspace - default parameter namespace
const (
	DefaultParamspace = ModuleName
)

// Parameter keys
var (
	KeyConfirmationHeight  = []byte("confirmationHeight")
	KeyTokenDeploySig      = []byte("tokenDeploySig")
	KeyNetwork             = []byte("network")
	KeyRevoteLockingPeriod = []byte("RevoteLockingPeriod")

	KeyGateway  = []byte("gateway")
	KeyToken    = []byte("token")
	KeyBurnable = []byte("burneable")
)

// KeyTable returns a subspace.KeyTable that has registered all parameter types in this module's parameter set
func KeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

// Params is the parameter set for this module
type Params struct {
	ConfirmationHeight  uint64
	Network             Network
	Gateway             []byte
	Token               []byte
	Burnable            []byte
	TokenDeploySig      []byte
	RevoteLockingPeriod int64
}

// DefaultParams returns the module's parameter set initialized with default values
func DefaultParams() Params {
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

	return Params{
		ConfirmationHeight:  1,
		Network:             Ganache,
		Gateway:             bzGateway,
		Token:               bzToken,
		Burnable:            bzBurnable,
		TokenDeploySig:      ERC20TokenDeploySig.Bytes(),
		RevoteLockingPeriod: 50,
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of tss module's parameters.
func (p *Params) ParamSetPairs() subspace.ParamSetPairs {
	/*
		because the subspace package makes liberal use of pointers to set and get values from the store,
		this method needs to have a pointer receiver AND NewParamSetPair needs to receive the
		parameter values as pointer arguments, otherwise either the internal type reflection panics or the value will not be
		set on the correct Params data struct
	*/
	return subspace.ParamSetPairs{
		subspace.NewParamSetPair(KeyConfirmationHeight, &p.ConfirmationHeight, validateConfirmationHeight),
		subspace.NewParamSetPair(KeyNetwork, &p.Network, validateNetwork),
		subspace.NewParamSetPair(KeyGateway, &p.Gateway, validateBytes),
		subspace.NewParamSetPair(KeyToken, &p.Token, validateBytes),
		subspace.NewParamSetPair(KeyBurnable, &p.Burnable, validateBytes),
		subspace.NewParamSetPair(KeyTokenDeploySig, &p.TokenDeploySig, validateBytes),
		subspace.NewParamSetPair(KeyRevoteLockingPeriod, &p.RevoteLockingPeriod, validateRevoteLockingPeriod),
	}
}

func validateNetwork(network interface{}) error {
	n, ok := network.(Network)
	if !ok {
		return fmt.Errorf("invalid parameter type for network: %T", network)
	}
	return n.Validate()
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

// Validate checks the validity of the values of the parameter set
func (p Params) Validate() error {
	if err := validateConfirmationHeight(p.ConfirmationHeight); err != nil {
		return err
	}

	if err := validateNetwork(p.Network); err != nil {
		return err
	}

	if err := validateRevoteLockingPeriod(p.RevoteLockingPeriod); err != nil {
		return err
	}

	return nil
}
