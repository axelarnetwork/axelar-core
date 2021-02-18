package types

import (
	"encoding/hex"
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/ethereum/go-ethereum/crypto"
)

// Default parameter namespace
const (
	DefaultParamspace = ModuleName
)

var (
	KeyConfirmationHeight = []byte("confirmationHeight")
	KeyNetwork            = []byte("network")
	KeyBurnable           = []byte("burneable")
	KeyToken              = []byte("token")
	KeyTransferSig        = []byte("transfersig")

	// ERC20TokenDeploySig is the signature of the ERC20 transfer method
	ERC20TokenDeploySig = "TokenDeployed(string,address)"
)

func KeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

type Params struct {
	ConfirmationHeight uint64
	Network            Network
	Token              []byte
	Burnable           []byte
	TransferSig        []byte
}

func DefaultParams() Params {

	bzToken, err := hex.DecodeString(token)
	if err != nil {
		panic(err)
	}
	bzBurnable, err := hex.DecodeString(burnable)
	if err != nil {
		panic(err)
	}

	transferSig := crypto.Keccak256Hash([]byte(ERC20TokenDeploySig)).Bytes()

	return Params{
		ConfirmationHeight: 1,
		Network:            Ganache,
		Token:              bzToken,
		Burnable:           bzBurnable,
		TransferSig:        transferSig,
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
		subspace.NewParamSetPair(KeyToken, &p.Token, validateBytes),
		subspace.NewParamSetPair(KeyBurnable, &p.Burnable, validateBytes),
		subspace.NewParamSetPair(KeyTransferSig, &p.TransferSig, validateBytes),
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
	if h < 0 {
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

func (p Params) Validate() error {
	if err := validateConfirmationHeight(p.ConfirmationHeight); err != nil {
		return err
	}
	if err := validateNetwork(p.Network); err != nil {
		return err
	}
	return nil
}
