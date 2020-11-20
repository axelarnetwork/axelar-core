package mock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
	"github.com/tendermint/tendermint/crypto"

	"github.com/axelarnetwork/axelar-core/x/voting/types"
)

var _ types.Staker = TestStaker{}
var _ exported.ValidatorI = TestValidator{}

type TestStaker struct {
	validators map[string]exported.ValidatorI
	totalPower int64
}

func NewTestStaker(validators ...staking.ValidatorI) TestStaker {
	staker := TestStaker{map[string]exported.ValidatorI{}, 0}

	for _, val := range validators {
		staker.validators[val.GetOperator().String()] = val
		staker.totalPower += val.GetConsensusPower()
	}
	return staker
}

func (s TestStaker) GetLastTotalPower(_ sdk.Context) (power sdk.Int) {
	return sdk.NewInt(s.totalPower)
}

func (s TestStaker) Validator(_ sdk.Context, address sdk.ValAddress) exported.ValidatorI {
	return s.validators[address.String()]
}

type TestValidator struct {
	power   int64
	address sdk.ValAddress
}

func NewTestValidator(addr sdk.ValAddress, votingPower int64) TestValidator {
	return TestValidator{
		power:   votingPower,
		address: addr,
	}
}

func (v TestValidator) IsJailed() bool {
	panic("implement me")
}

func (v TestValidator) GetMoniker() string {
	panic("implement me")
}

func (v TestValidator) GetStatus() sdk.BondStatus {
	panic("implement me")
}

func (v TestValidator) IsBonded() bool {
	panic("implement me")
}

func (v TestValidator) IsUnbonded() bool {
	panic("implement me")
}

func (v TestValidator) IsUnbonding() bool {
	panic("implement me")
}

func (v TestValidator) GetOperator() sdk.ValAddress {
	return v.address
}

func (v TestValidator) GetConsPubKey() crypto.PubKey {
	panic("implement me")
}

func (v TestValidator) GetConsAddr() sdk.ConsAddress {
	panic("implement me")
}

func (v TestValidator) GetTokens() sdk.Int {
	panic("implement me")
}

func (v TestValidator) GetBondedTokens() sdk.Int {
	panic("implement me")
}

func (v TestValidator) GetConsensusPower() int64 {
	return v.power
}

func (v TestValidator) GetCommission() sdk.Dec {
	panic("implement me")
}

func (v TestValidator) GetMinSelfDelegation() sdk.Int {
	panic("implement me")
}

func (v TestValidator) GetDelegatorShares() sdk.Dec {
	panic("implement me")
}

func (v TestValidator) TokensFromShares(_ sdk.Dec) sdk.Dec {
	panic("implement me")
}

func (v TestValidator) TokensFromSharesTruncated(_ sdk.Dec) sdk.Dec {
	panic("implement me")
}

func (v TestValidator) TokensFromSharesRoundUp(_ sdk.Dec) sdk.Dec {
	panic("implement me")
}

func (v TestValidator) SharesFromTokens(_ sdk.Int) (sdk.Dec, error) {
	panic("implement me")
}

func (v TestValidator) SharesFromTokensTruncated(_ sdk.Int) (sdk.Dec, error) {
	panic("implement me")
}
