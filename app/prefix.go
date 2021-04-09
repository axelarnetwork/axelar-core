package app

import sdk "github.com/cosmos/cosmos-sdk/types"

// Bech32 prefixes
var (
	AccountAddressPrefix   = "axelar"
	AccountPubKeyPrefix    = AccountAddressPrefix + sdk.PrefixPublic
	ValidatorAddressPrefix = AccountAddressPrefix + sdk.PrefixValidator + sdk.PrefixOperator
	ValidatorPubKeyPrefix  = AccountAddressPrefix + sdk.PrefixValidator + sdk.PrefixOperator + sdk.PrefixPublic
	ConsNodeAddressPrefix  = AccountAddressPrefix + sdk.PrefixValidator + sdk.PrefixConsensus
	ConsNodePubKeyPrefix   = AccountAddressPrefix + sdk.PrefixValidator + sdk.PrefixConsensus + sdk.PrefixPublic
)

// SetConfig sets the prefix config for the bech32 encoding
func SetConfig() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(AccountAddressPrefix, AccountPubKeyPrefix)
	config.SetBech32PrefixForValidator(ValidatorAddressPrefix, ValidatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(ConsNodeAddressPrefix, ConsNodePubKeyPrefix)
	config.Seal()
}
