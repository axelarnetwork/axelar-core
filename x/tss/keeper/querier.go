package keeper

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// Query paths
const (
	QuerySignature            = "signature"
	QueryKey                  = "key"
	QueryRecovery             = "recovery"
	QueryKeyID                = "key-id"
	QueryKeySharesByKeyID     = "key-share-id"
	QueryKeySharesByValidator = "key-share-validator"
	QueryDeactivated          = "deactivated"
)

// NewQuerier returns a new querier for the TSS module
func NewQuerier(k types.TSSKeeper, v types.Voter, s types.Snapshotter, staking types.StakingKeeper, n types.Nexus) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var res []byte
		var err error
		switch path[0] {
		case QuerySignature:
			res, err = querySignatureStatus(ctx, k, v, path[1])
		case QueryKey:
			res, err = queryKey(ctx, k, v, path[1])
		case QueryRecovery:
			res, err = queryRecovery(ctx, k, s, path[1], path[2])
		case QueryKeyID:
			res, err = queryKeyID(ctx, k, n, path[1], path[2])
		case QueryKeySharesByKeyID:
			res, err = queryKeySharesByKeyID(ctx, k, s, path[1])
		case QueryKeySharesByValidator:
			res, err = queryKeySharesByValidator(ctx, k, n, s, path[1])
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

func queryRecovery(ctx sdk.Context, k types.TSSKeeper, s types.Snapshotter, keyID string, addressStr string) ([]byte, error) {

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

	// TODO: get actual pubkey and groupinfo
	pubKey := []byte{1}
	if pubKey == nil {
		return nil, fmt.Errorf("could not obtain pubkey for key ID %s", keyID)
	}

	groupRecoverInfo := []byte{1}
	if pubKey == nil {
		return nil, fmt.Errorf("could not obtain group info for key ID %s", keyID)
	}

	privateRecoverInfo := k.GetPrivateRecoveryInfo(ctx, address, keyID)
	if pubKey == nil {
		return nil, fmt.Errorf("could not obtain private info for key ID %s", keyID)
	}

	resp := tssTypes.QueryRecoveryResponse{
		Threshold:        int32(snapshot.CorruptionThreshold),
		PartyUids:        participants,
		PartyShareCounts: participantShareCounts,
		KeygenOutput: &tssTypes.QueryRecoveryResponse_KeygenOutput{
			PubKey:             pubKey,
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

		return types.ModuleCdc.MarshalBinaryLengthPrefixed(&res)
	}

	var res types.QuerySignatureResponse
	pollMeta := voting.NewPollKey(types.ModuleName, sigID)

	if poll := v.GetPoll(ctx, pollMeta); poll.Is(voting.NonExistent) {
		res.VoteStatus = types.NotFound
	} else {
		res.VoteStatus = types.Pending
	}

	return types.ModuleCdc.MarshalBinaryLengthPrefixed(&res)
}

func queryKey(ctx sdk.Context, k types.TSSKeeper, v types.Voter, keyID string) ([]byte, error) {
	if key, ok := k.GetKey(ctx, keyID); ok {
		// poll was successful
		res := types.QueryKeyResponse{
			VoteStatus: types.Decided,
			Role:       key.Role,
		}

		return types.ModuleCdc.MarshalBinaryLengthPrefixed(&res)
	}

	var res types.QueryKeyResponse
	pollMeta := voting.NewPollKey(types.ModuleName, keyID)

	if poll := v.GetPoll(ctx, pollMeta); poll.Is(voting.NonExistent) {
		res.VoteStatus = types.NotFound
	} else {
		res.VoteStatus = types.Pending
	}

	return types.ModuleCdc.MarshalBinaryLengthPrefixed(&res)
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

func queryKeySharesByKeyID(ctx sdk.Context, k types.TSSKeeper, s types.Snapshotter, keyID string) ([]byte, error) {

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

	return types.ModuleCdc.MarshalBinaryLengthPrefixed(&resp)
}
