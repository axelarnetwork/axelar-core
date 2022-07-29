package keeper

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/math"
	"github.com/axelarnetwork/utils/slices"
)

// SetSigRouter sets the sig router. It will panic if called more than once
func (k *Keeper) SetSigRouter(router types.SigRouter) {
	if k.sigRouter != nil {
		panic("router already set")
	}

	k.sigRouter = router

	// In order to avoid invalid or non-deterministic behavior, we seal the router immediately
	// to prevent additional handlers from being registered after the keeper is initialized.
	k.sigRouter.Seal()
}

// GetSigRouter returns the sig router. If no router was set, it returns a (sealed) router with no handlers
func (k Keeper) GetSigRouter() types.SigRouter {
	if k.sigRouter == nil {
		k.SetSigRouter(types.NewSigRouter())
	}

	return k.sigRouter
}

// GetSigningSessionsByExpiry returns all signing sessions that either expires at
// or goes out of the grace period at the given block height
func (k Keeper) GetSigningSessionsByExpiry(ctx sdk.Context, expiry int64) []types.SigningSession {
	var results []types.SigningSession

	iter := k.getStore(ctx).Iterator(expirySigningPrefix.Append(utils.KeyFromInt(expiry)))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var value gogoprototypes.UInt64Value
		iter.UnmarshalValue(&value)

		sigID := value.Value
		result, ok := k.getSigningSession(ctx, sigID)
		if !ok {
			panic(fmt.Errorf("signing session %d not found", sigID))
		}

		results = append(results, result)
	}

	return results
}

// Sign starts a signing session to sign the given payload's hash with the given
// key ID
func (k Keeper) Sign(ctx sdk.Context, keyID exported.KeyID, payloadHash exported.Hash, module string, moduleMetadata ...codec.ProtoMarshaler) error {
	if !k.GetSigRouter().HasHandler(module) {
		panic(fmt.Errorf("sig handler not registered for module %s", module))
	}

	key, ok := k.getKey(ctx, keyID)
	if !ok {
		return fmt.Errorf("key %s not found", keyID)
	}
	if key.State != exported.Active {
		return fmt.Errorf("key %s is not activated yet", keyID)
	}

	params := k.getParams(ctx)

	expiresAt := ctx.BlockHeight() + params.SigningTimeout
	signingSession := types.NewSigningSession(k.nextSigID(ctx), key, payloadHash, expiresAt, params.SigningGracePeriod, module, moduleMetadata...)
	if err := signingSession.ValidateBasic(); err != nil {
		return err
	}

	k.setSigningSession(ctx, signingSession)

	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(types.NewSigningStarted(signingSession.GetID(), key, payloadHash[:], module)))
	k.Logger(ctx).Info("signing session started",
		"sig_id", signingSession.GetID(),
		"key_id", key.GetID(),
		"participant_count", len(key.GetPubKeys()),
		"participants", strings.Join(slices.Map(key.GetParticipants(), sdk.ValAddress.String), ", "),
		"participants_weight", key.GetParticipantsWeight().String(),
		"bonded_weight", key.GetSnapshot().BondedWeight.String(),
		"signing_threshold", key.GetSigningThreshold().String(),
		"expires_at", expiresAt,
	)

	return nil
}

// DeleteSigningSession deletes the signing session with the given ID
func (k Keeper) DeleteSigningSession(ctx sdk.Context, id uint64) {
	signing, ok := k.getSigningSession(ctx, id)
	if !ok {
		return
	}

	k.getStore(ctx).Delete(getSigningSessionExpiryKey(signing))
	k.getStore(ctx).Delete(getSigningSessionKey(id))
}

func (k Keeper) setSigningSession(ctx sdk.Context, signing types.SigningSession) {
	// the deletion is necessary because we may update it to a different location depending on the current state of the session
	k.getStore(ctx).Delete(expirySigningPrefix.Append(utils.KeyFromInt(signing.ExpiresAt)).Append(utils.KeyFromInt(signing.GetID())))
	k.getStore(ctx).Set(getSigningSessionExpiryKey(signing), &gogoprototypes.UInt64Value{Value: signing.GetID()})

	k.getStore(ctx).Set(getSigningSessionKey(signing.GetID()), &signing)
}

func (k Keeper) getSigningSession(ctx sdk.Context, id uint64) (signing types.SigningSession, ok bool) {
	return signing, k.getStore(ctx).Get(getSigningSessionKey(id), &signing)
}

func (k Keeper) nextSigID(ctx sdk.Context) uint64 {
	var val gogoprototypes.UInt64Value
	k.getStore(ctx).Get(signingSessionCountKey, &val)
	defer k.getStore(ctx).Set(signingSessionCountKey, &gogoprototypes.UInt64Value{Value: val.Value + 1})

	return val.Value
}

func getSigningSessionExpiryKey(signing types.SigningSession) utils.Key {
	expiry := signing.ExpiresAt
	if signing.State == exported.Completed {
		expiry = math.Min(signing.ExpiresAt, signing.CompletedAt+signing.GracePeriod+1)
	}

	return expirySigningPrefix.Append(utils.KeyFromInt(expiry)).Append(utils.KeyFromInt(signing.GetID()))
}

func getSigningSessionKey(id uint64) utils.Key {
	return signingPrefix.Append(utils.KeyFromInt(id))
}
