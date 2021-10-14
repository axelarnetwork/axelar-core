package keeper

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// Query paths
const (
	QuerySignature                = "signature"
	QueryKey                      = "key"
	QueryRecovery                 = "recovery"
	QueryKeyID                    = "key-id"
	QueryKeySharesByKeyID         = "key-share-id"
	QueryKeySharesByValidator     = "key-share-validator"
	QueryActiveOldKeys            = "active-old-keys"
	QueryActiveOldKeysByValidator = "active-old-keys-validator"
	QueryDeactivated              = "deactivated"
	QExternalKeyID                = "external-key-id"
)

// NewQuerier returns a new querier for the TSS module
func NewQuerier(k types.TSSKeeper, v types.Voter, s types.Snapshotter, staking types.StakingKeeper, n types.Nexus) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var res []byte
		var err error
		switch path[0] {
		case QExternalKeyID:
			res, err = QueryExternalKeyID(ctx, k, n, path[1])
		case QuerySignature:
			res, err = querySignatureStatus(ctx, k, v, path[1])
		case QueryKey:
			keyID := exported.KeyID(path[1])
			err = keyID.Validate()
			if err != nil {
				break
			}
			res, err = queryKeyStatus(ctx, k, v, keyID)
		case QueryRecovery:
			keyID := exported.KeyID(path[1])
			err = keyID.Validate()
			if err != nil {
				break
			}
			res, err = queryRecovery(ctx, k, s, keyID, path[2])
		case QueryKeyID:
			res, err = queryKeyID(ctx, k, n, path[1], path[2])
		case QueryKeySharesByKeyID:
			keyID := exported.KeyID(path[1])
			err = keyID.Validate()
			if err != nil {
				break
			}
			res, err = queryKeySharesByKeyID(ctx, k, s, keyID)
		case QueryKeySharesByValidator:
			res, err = queryKeySharesByValidator(ctx, k, n, s, path[1])
		case QueryActiveOldKeys:
			res, err = queryActiveOldKeyIDs(ctx, k, n, path[1], path[2])
		case QueryActiveOldKeysByValidator:
			res, err = queryActiveOldKeyIDsByValidator(ctx, k, n, s, path[1])
		case QueryDeactivated:
			res, err = queryDeactivatedOperator(ctx, k, s, staking)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown tss query endpoint: %s", path[0]))
		}

		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrTss, err.Error())
		}
		return res, nil
	}
}

func queryRecovery(ctx sdk.Context, k types.TSSKeeper, s types.Snapshotter, keyID exported.KeyID, addressStr string) ([]byte, error) {

	address, err := sdk.ValAddressFromBech32(addressStr)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to parse validator address")
	}

	counter, ok := k.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("could not obtain snapshot counter for key ID %s", keyID)
	}

	snapshot, ok := s.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("could not obtain snapshot for counter %d", counter)
	}

	participants := make([]string, 0, len(snapshot.Validators))
	participantShareCounts := make([]uint32, 0, len(snapshot.Validators))
	for _, validator := range snapshot.Validators {
		participants = append(participants, validator.GetSDKValidator().GetOperator().String())
		participantShareCounts = append(participantShareCounts, uint32(validator.ShareCount))
	}

	// get voted pub key
	pubKey, ok := k.GetKey(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("could not obtain pubkey for key ID %s", keyID)
	}

	// convert ecdsa pub key to bytes
	ecdsaPK := btcec.PublicKey(pubKey.Value)
	pubKeyBytes := ecdsaPK.SerializeCompressed()

	// get voted group recover info
	groupRecoverInfo := k.GetGroupRecoveryInfo(ctx, keyID)
	if groupRecoverInfo == nil {
		return nil, fmt.Errorf("could not obtain group info for key ID %s", keyID)
	}

	privateRecoverInfo := k.GetPrivateRecoveryInfo(ctx, address, keyID)
	if privateRecoverInfo == nil {
		return nil, fmt.Errorf("could not obtain private info for key ID %s", keyID)
	}

	resp := types.QueryRecoveryResponse{
		Threshold:        uint32(snapshot.CorruptionThreshold),
		PartyUids:        participants,
		PartyShareCounts: participantShareCounts,
		KeygenOutput: &tofnd.KeygenOutput{
			PubKey:             pubKeyBytes,
			GroupRecoverInfo:   groupRecoverInfo,
			PrivateRecoverInfo: privateRecoverInfo,
		},
	}

	return resp.Marshal()
}

func querySignatureStatus(ctx sdk.Context, k types.TSSKeeper, v types.Voter, sigID string) ([]byte, error) {
	if sig, status := k.GetSig(ctx, sigID); status == exported.SigStatus_Signed {
		// poll was successful
		res := types.QuerySignatureResponse{
			VoteStatus: types.Decided,
			Signature: &types.QuerySignatureResponse_Signature{
				R: hex.EncodeToString(sig.R.Bytes()),
				S: hex.EncodeToString(sig.S.Bytes()),
			},
		}

		return types.ModuleCdc.MarshalLengthPrefixed(&res)
	}

	var res types.QuerySignatureResponse
	pollMeta := voting.NewPollKey(types.ModuleName, sigID)

	if poll := v.GetPoll(ctx, pollMeta); poll.Is(voting.NonExistent) {
		res.VoteStatus = types.NotFound
	} else {
		res.VoteStatus = types.Pending
	}

	return types.ModuleCdc.MarshalLengthPrefixed(&res)
}

func queryKeyStatus(ctx sdk.Context, k types.TSSKeeper, v types.Voter, keyID exported.KeyID) ([]byte, error) {
	if key, ok := k.GetKey(ctx, keyID); ok {
		// poll was successful
		res := types.QueryKeyResponse{
			VoteStatus: types.Decided,
			Role:       key.Role,
			Key: &types.QueryKeyResponse_Key{
				X: hex.EncodeToString(key.Value.X.Bytes()),
				Y: hex.EncodeToString(key.Value.Y.Bytes()),
			},
			RotatedAt: key.RotatedAt,
		}

		return types.ModuleCdc.MarshalLengthPrefixed(&res)
	}

	var res types.QueryKeyResponse
	pollMeta := voting.NewPollKey(types.ModuleName, string(keyID))

	if poll := v.GetPoll(ctx, pollMeta); poll.Is(voting.NonExistent) {
		res.VoteStatus = types.NotFound
	} else {
		res.VoteStatus = types.Pending
	}

	return types.ModuleCdc.MarshalLengthPrefixed(&res)
}

// queryKeyID returns the keyID of the most recent key for a provided keyChain and keyRole
func queryKeyID(ctx sdk.Context, k types.TSSKeeper, n types.Nexus, keyChainStr string, keyRoleStr string) ([]byte, error) {
	keyChain, ok := n.GetChain(ctx, keyChainStr)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", keyChainStr)
	}

	keyRole, err := exported.KeyRoleFromSimpleStr(keyRoleStr)
	if err != nil {
		return nil, err
	}

	if keyRole == exported.ExternalKey {
		return nil, fmt.Errorf("use the chain specific query for %s to get external keyIDs", keyChainStr)
	}

	keyID, found := k.GetCurrentKeyID(ctx, keyChain, keyRole)
	if !found {
		return nil, fmt.Errorf("no key from chain %s role %s exists", keyChainStr, keyRoleStr)
	}

	return []byte(keyID), nil
}

func queryKeySharesByKeyID(ctx sdk.Context, k types.TSSKeeper, s types.Snapshotter, keyID exported.KeyID) ([]byte, error) {

	counter, ok := k.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("invalid keyID %s", keyID)
	}

	snapshot, ok := s.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("no snapshot found for counter number %d", counter)
	}

	var allShareInfos []types.QueryKeyShareResponse_ShareInfo
	for _, validator := range snapshot.Validators {

		thisShareInfo := types.QueryKeyShareResponse_ShareInfo{
			KeyID:               keyID,
			SnapshotBlockNumber: snapshot.Height,
			ValidatorAddress:    validator.GetSDKValidator().GetOperator().String(),
			NumValidatorShares:  validator.ShareCount,
			NumTotalShares:      snapshot.TotalShareCount.Int64(),
		}

		allShareInfos = append(allShareInfos, thisShareInfo)
	}

	keyShareInfos := types.QueryKeyShareResponse{
		ShareInfos: allShareInfos,
	}

	return keyShareInfos.Marshal()
}

func queryActiveOldKeyIDs(ctx sdk.Context, k types.TSSKeeper, n types.Nexus, chainName, roleStr string) ([]byte, error) {
	var queryResponse types.QueryActiveOldKeysResponse

	chain, ok := n.GetChain(ctx, chainName)
	if !ok {
		return nil, fmt.Errorf("could not find chain '%s'", chainName)
	}

	role, err := exported.KeyRoleFromSimpleStr(roleStr)
	if err != nil {
		return nil, err
	}

	keys, err := k.GetOldActiveKeys(ctx, chain, role)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		queryResponse.KeyIDs = append(queryResponse.KeyIDs, key.ID)
	}
	return queryResponse.Marshal()
}

func queryActiveOldKeyIDsByValidator(ctx sdk.Context, k types.TSSKeeper, n types.Nexus, s types.Snapshotter, targetValidatorAddr string) ([]byte, error) {
	var allKeys []types.QueryActiveOldKeysValidatorResponse_KeyInfo
	var queryResponse types.QueryActiveOldKeysValidatorResponse

	for _, chain := range n.GetChains(ctx) {
		for _, role := range exported.GetKeyRoles() {
			keys, err := k.GetOldActiveKeys(ctx, chain, role)
			if err != nil {
				return nil, err
			}

			for _, key := range keys {
				allKeys = append(allKeys, types.QueryActiveOldKeysValidatorResponse_KeyInfo{
					ID:    key.ID,
					Chain: chain.Name,
					Role:  role,
				})
			}
		}
	}

	for _, key := range allKeys {
		counter, ok := k.GetSnapshotCounterForKeyID(ctx, key.ID)
		if !ok {
			return nil, fmt.Errorf("could not get snapshot counter from keyID %s", key.ID)
		}

		snapshot, ok := s.GetSnapshot(ctx, counter)
		if !ok {
			return nil, fmt.Errorf("no snapshot found for counter number %d", counter)
		}

		for _, validator := range snapshot.Validators {
			validatorAddr := validator.GetSDKValidator().GetOperator().String()
			if validatorAddr == targetValidatorAddr {
				queryResponse.KeysInfo = append(queryResponse.KeysInfo, key)
				break
			}
		}
	}
	return queryResponse.Marshal()
}

func queryKeySharesByValidator(ctx sdk.Context, k types.TSSKeeper, n types.Nexus, s types.Snapshotter, targetValidatorAddr string) ([]byte, error) {

	var allShareInfos []types.QueryKeyShareResponse_ShareInfo

	for _, chain := range n.GetChains(ctx) {
		for _, keyRole := range exported.GetKeyRoles() {

			keyID, found := k.GetCurrentKeyID(ctx, chain, keyRole)

			if !found {
				continue
			}

			counter, ok := k.GetSnapshotCounterForKeyID(ctx, keyID)
			if !ok {
				return nil, fmt.Errorf("could not get snapshot counter from keyID %s", keyID)
			}

			snapshot, ok := s.GetSnapshot(ctx, counter)
			if !ok {
				return nil, fmt.Errorf("no snapshot found for counter number %d", counter)
			}

			for _, validator := range snapshot.Validators {

				validatorAddr := validator.GetSDKValidator().GetOperator().String()
				if validatorAddr == targetValidatorAddr {

					thisShareInfo := types.QueryKeyShareResponse_ShareInfo{
						KeyID:               keyID,
						KeyChain:            chain.Name,
						KeyRole:             keyRole.String(),
						SnapshotBlockNumber: snapshot.Height,
						ValidatorAddress:    validator.GetSDKValidator().GetOperator().String(),
						NumValidatorShares:  validator.ShareCount,
						NumTotalShares:      snapshot.TotalShareCount.Int64(),
					}
					allShareInfos = append(allShareInfos, thisShareInfo)
					break
				}
			}
		}
	}

	keyShareInfos := types.QueryKeyShareResponse{
		ShareInfos: allShareInfos,
	}

	return keyShareInfos.Marshal()
}

func queryDeactivatedOperator(ctx sdk.Context, k types.TSSKeeper, s types.Snapshotter, staking types.StakingKeeper) ([]byte, error) {

	var deactivatedValidators []string
	validatorIter := func(_ int64, validator stakingtypes.ValidatorI) (stop bool) {

		// this explicit type cast is necessary, because we need to call UnpackInterfaces() on the validator
		// and it is not exposed in the ValidatorI interface
		v, ok := validator.(stakingtypes.Validator)
		if !ok {
			k.Logger(ctx).Error(fmt.Sprintf("unexpected validator type: expected %T, got %T", stakingtypes.Validator{}, validator))
			return false
		}

		_, active := s.GetProxy(ctx, v.GetOperator())
		if !active {
			deactivatedValidators = append(deactivatedValidators, v.GetOperator().String())
		}

		return false
	}
	// IterateBondedValidatorsByPower(https://github.com/cosmos/cosmos-sdk/blob/7fc7b3f6ff82eb5ede52881778114f6b38bd7dfa/x/staking/keeper/alias_functions.go#L33) iterates validators by power in descending order
	staking.IterateBondedValidatorsByPower(ctx, validatorIter)

	resp := types.QueryDeactivatedOperatorsResponse{
		OperatorAddresses: deactivatedValidators,
	}

	return types.ModuleCdc.MarshalLengthPrefixed(&resp)
}

// QueryExternalKeyID returns the keyIDs of the current set of external keys for the given chain
func QueryExternalKeyID(ctx sdk.Context, k types.TSSKeeper, n types.Nexus, chainStr string) ([]byte, error) {
	chain, ok := n.GetChain(ctx, chainStr)
	if !ok {
		return nil, fmt.Errorf("unknown chain %s", chainStr)
	}

	externalKeyIDs, ok := k.GetExternalKeyIDs(ctx, chain)
	if !ok {
		return nil, fmt.Errorf("external keys not found")
	}

	resp := types.QueryExternalKeyIDResponse{
		KeyIDs: externalKeyIDs,
	}

	return types.ModuleCdc.MarshalLengthPrefixed(&resp)
}
