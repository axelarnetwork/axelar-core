package types

import (
	"fmt"
	"time"

	"github.com/axelarnetwork/axelar-core/utils"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	dustLimit = 546
)

// Parameter keys
var (
	KeyConfirmationHeight                   = []byte("confirmationHeight")
	KeyNetwork                              = []byte("network")
	KeyRevoteLockingPeriod                  = []byte("revoteLockingPeriod")
	KeySigCheckInterval                     = []byte("sigCheckInterval")
	KeyMinOutputAmount                      = []byte("minOutputAmount")
	KeyMaxInputCount                        = []byte("maxInputCount")
	KeyMaxSecondaryOutputAmount             = []byte("maxSecondaryOutputAmount")
	KeyMasterKeyRetentionPeriod             = []byte("masterKeyRetentionPeriod")
	KeyMasterAddressInternalKeyLockDuration = []byte("masterAddressInternalKeyLockDuration")
	KeyMasterAddressExternalKeyLockDuration = []byte("masterAddressExternalKeyLockDuration")
	KeyVotingThreshold                      = []byte("votingThreshold")
	KeyMinVoterCount                        = []byte("minVoterCount")
	KeyMaxTxSize                            = []byte("maxTxSize")
)

// KeyTable returns a subspace.KeyTable that has registered all parameter types in this module's parameter set
func KeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns the module's parameter set initialized with default values
func DefaultParams() Params {
	return Params{
		ConfirmationHeight:                   1,
		Network:                              Network{Name: Regtest.Name},
		RevoteLockingPeriod:                  50,
		SigCheckInterval:                     10,
		MinOutputAmount:                      sdktypes.NewDecCoin(Satoshi, sdktypes.NewInt(1000)),
		MaxInputCount:                        50,
		MaxSecondaryOutputAmount:             sdktypes.NewDecCoin(Bitcoin, sdktypes.NewInt(300)),
		MasterKeyRetentionPeriod:             8,
		MasterAddressInternalKeyLockDuration: 14 * 24 * time.Hour, // 14 days
		MasterAddressExternalKeyLockDuration: 28 * 24 * time.Hour, // 28 days
		VotingThreshold:                      utils.Threshold{Numerator: 15, Denominator: 100},
		MinVoterCount:                        1,
		MaxTxSize:                            1024 * 1024 / 3, // 1/3 MiB
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of tss module's parameters.
func (m *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	/*
		because the paramtypes package makes liberal use of pointers to set and get values from the store,
		this method needs to have a pointer receiver AND NewParamSetPair needs to receive the
		parameter values as pointer arguments, otherwise either the internal type reflection panics or the value will not be
		set on the correct Params data struct
	*/
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyConfirmationHeight, &m.ConfirmationHeight, validateConfirmationHeight),
		paramtypes.NewParamSetPair(KeyNetwork, &m.Network, validateNetwork),
		paramtypes.NewParamSetPair(KeyRevoteLockingPeriod, &m.RevoteLockingPeriod, validateRevoteLockingPeriod),
		paramtypes.NewParamSetPair(KeySigCheckInterval, &m.SigCheckInterval, validateSigCheckInterval),
		paramtypes.NewParamSetPair(KeyMinOutputAmount, &m.MinOutputAmount, validateMinOutputAmount),
		paramtypes.NewParamSetPair(KeyMaxInputCount, &m.MaxInputCount, validateMaxInputCount),
		paramtypes.NewParamSetPair(KeyMaxSecondaryOutputAmount, &m.MaxSecondaryOutputAmount, validateMaxSecondaryOutputAmount),
		paramtypes.NewParamSetPair(KeyMasterKeyRetentionPeriod, &m.MasterKeyRetentionPeriod, validateMasterKeyRetentionPeriod),
		paramtypes.NewParamSetPair(KeyMasterAddressInternalKeyLockDuration, &m.MasterAddressInternalKeyLockDuration, validateMasterAddressLockDuration),
		paramtypes.NewParamSetPair(KeyMasterAddressExternalKeyLockDuration, &m.MasterAddressExternalKeyLockDuration, validateMasterAddressLockDuration),
		paramtypes.NewParamSetPair(KeyVotingThreshold, &m.VotingThreshold, validateVotingThreshold),
		paramtypes.NewParamSetPair(KeyMinVoterCount, &m.MinVoterCount, validateMinVoterCount),
		paramtypes.NewParamSetPair(KeyMaxTxSize, &m.MaxTxSize, validateMaxTxSize),
	}
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

func validateNetwork(network interface{}) error {
	n, ok := network.(Network)
	if !ok {
		return sdkerrors.Wrapf(types.ErrInvalidGenesis, "invalid parameter type for network: %T", network)
	}
	return n.Validate()
}

func validateRevoteLockingPeriod(period interface{}) error {
	r, ok := period.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type for revote lock period: %T", r)
	}

	if r <= 0 {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "revote lock period must be greater than 0")
	}

	return nil
}

func validateMinOutputAmount(amount interface{}) error {
	coin, ok := amount.(sdktypes.DecCoin)
	if !ok {
		return fmt.Errorf("invalid parameter type for min output amount: %T", coin)
	}

	satoshi, err := ToSatoshiCoin(coin)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidGenesis, "invalid min output amount with error %s", err.Error())
	}

	if satoshi.Amount.LT(sdktypes.NewInt(dustLimit)) {
		return sdkerrors.Wrapf(types.ErrInvalidGenesis, "min output amount has to be greater than %d", dustLimit)
	}

	return nil
}

func validateSigCheckInterval(interval interface{}) error {
	i, ok := interval.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type for signature check interval: %T", i)
	}

	if i <= 0 {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "signature check interval must be greater than 0")
	}

	return nil
}

func validateMaxInputCount(maxInputCount interface{}) error {
	m, ok := maxInputCount.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type for max input count: %T", maxInputCount)
	}

	if m <= 0 {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "max input count must be greater than 0")
	}

	return nil
}

func validateMaxSecondaryOutputAmount(amount interface{}) error {
	coin, ok := amount.(sdktypes.DecCoin)
	if !ok {
		return fmt.Errorf("invalid parameter type for max secondary output amount: %T", coin)
	}

	satoshi, err := ToSatoshiCoin(coin)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidGenesis, "invalid max secondary output amount with error %s", err.Error())
	}

	if satoshi.Amount.LT(sdktypes.NewInt(dustLimit)) {
		return sdkerrors.Wrapf(types.ErrInvalidGenesis, "max secondary output amount has to be greater than %d", dustLimit)
	}

	return nil
}

func validateMasterKeyRetentionPeriod(masterKeyRetentionPeriod interface{}) error {
	m, ok := masterKeyRetentionPeriod.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type for master key retention period: %T", masterKeyRetentionPeriod)
	}

	if m <= 0 {
		return fmt.Errorf("master key retention period has to be greater than 0")
	}

	return nil
}

func validateMasterAddressLockDuration(masterAddressInternalKeyLockDuration interface{}) error {
	m, ok := masterAddressInternalKeyLockDuration.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type for master address lock duration: %T", masterAddressInternalKeyLockDuration)
	}

	if m <= 0 {
		return fmt.Errorf("master address lock duration has to be greater than 0")
	}

	return nil
}

func validateMasterAddressLockDurations(masterAddressInternalKeyLockDuration time.Duration, masterAddressExternalKeyLockDuration time.Duration) error {
	if masterAddressExternalKeyLockDuration <= masterAddressInternalKeyLockDuration {
		return fmt.Errorf("master address external-key lock duration must be greater than master address internal-key lock duration")
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

func validateMaxTxSize(maxTxSize interface{}) error {
	val, ok := maxTxSize.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type for MaxTxSize: %T", maxTxSize)
	}

	if val <= 0 {
		return fmt.Errorf("max tx size must be >0")
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
	if err := validateSigCheckInterval(m.SigCheckInterval); err != nil {
		return err
	}

	if err := validateMinOutputAmount(m.MinOutputAmount); err != nil {
		return err
	}

	if err := validateMaxInputCount(m.MaxInputCount); err != nil {
		return err
	}

	if err := validateMaxSecondaryOutputAmount(m.MaxSecondaryOutputAmount); err != nil {
		return err
	}

	if err := validateMasterKeyRetentionPeriod(m.MasterKeyRetentionPeriod); err != nil {
		return err
	}

	if err := validateVotingThreshold(m.VotingThreshold); err != nil {
		return err
	}

	if err := validateMinVoterCount(m.MinVoterCount); err != nil {
		return err
	}

	if err := validateMaxTxSize(m.MaxTxSize); err != nil {
		return err
	}

	if err := validateMasterAddressLockDuration(m.MasterAddressInternalKeyLockDuration); err != nil {
		return err
	}

	if err := validateMasterAddressLockDuration(m.MasterAddressExternalKeyLockDuration); err != nil {
		return err
	}

	if err := validateMasterAddressLockDurations(m.MasterAddressInternalKeyLockDuration, m.MasterAddressExternalKeyLockDuration); err != nil {
		return err
	}

	return nil
}
