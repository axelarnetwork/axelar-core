package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/permission/exported"
	"github.com/axelarnetwork/axelar-core/x/permission/types"
)

var (
	governanceKey = utils.KeyFromStr("governance")
	accountPrefix = utils.KeyFromStr("account")
)

// Keeper provides access to all state changes regarding the gov module
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec
	params   params.Subspace
}

// NewKeeper returns a new reward keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramSpace params.Subspace) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		params:   paramSpace.WithKeyTable(types.KeyTable()),
	}
}

// GetParams gets the permission module's parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.params.GetParamSet(ctx, &p)
	return p
}

// setParams sets the permission module's parameters
func (k Keeper) setParams(ctx sdk.Context, p types.Params) {
	k.params.SetParamSet(ctx, &p)
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// setGovernanceKey sets the multisig governance key
func (k Keeper) setGovernanceKey(ctx sdk.Context, key multisig.LegacyAminoPubKey) {
	k.getStore(ctx).Set(governanceKey, &key)
}

// GetGovernanceKey gets the multisig governance key
func (k Keeper) GetGovernanceKey(ctx sdk.Context) (multisig.LegacyAminoPubKey, bool) {
	var key multisig.LegacyAminoPubKey
	ok := k.getStore(ctx).Get(governanceKey, &key)

	return key, ok
}

// GetRole returns the role of the given account address
func (k Keeper) GetRole(ctx sdk.Context, address sdk.AccAddress) exported.Role {
	if address.Empty() {
		return exported.ROLE_UNRESTRICTED
	}

	govAccount, ok := k.getGovAccount(ctx, address)
	if !ok {
		return exported.ROLE_UNRESTRICTED
	}

	return govAccount.Role
}

func (k Keeper) setGovAccount(ctx sdk.Context, govAccount types.GovAccount) {
	k.getStore(ctx).Set(accountPrefix.Append(utils.KeyFromBz(govAccount.Address)), &govAccount)
}

func (k Keeper) deleteGovAccount(ctx sdk.Context, address sdk.AccAddress) {
	k.getStore(ctx).Delete(accountPrefix.Append(utils.KeyFromBz(address)))
}

func (k Keeper) getGovAccount(ctx sdk.Context, address sdk.AccAddress) (govAccount types.GovAccount, ok bool) {
	return govAccount, k.getStore(ctx).Get(accountPrefix.Append(utils.KeyFromBz(address)), &govAccount)
}

func (k Keeper) getGovAccounts(ctx sdk.Context) []types.GovAccount {
	var accounts []types.GovAccount

	store := k.getStore(ctx)
	iter := store.Iterator(accountPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var account types.GovAccount
		iter.UnmarshalValue(&account)

		accounts = append(accounts, account)
	}

	return accounts
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
