package keeper

import (
	"fmt"

	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

// Query paths
const (
	QuerySigStatus 				= "sig-status"
	QueryKeyStatus 				= "key-status"
	QueryRecovery  				= "recovery"
	QueryKeyID	  				= "key-id"
	QueryKeySharesByKeyID		= "key-share-id"
	QueryKeySharesByValidator	= "key-share-validator"
)

// NewQuerier returns a new querier for the TSS module
func NewQuerier(k tssTypes.TSSKeeper, v tssTypes.Voter, s tssTypes.Snapshotter, n tssTypes.Nexus) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var res []byte
		var err error
		switch path[0] {
		case QuerySigStatus:
			res, err = querySigStatus(ctx, k, v, path[1])
		case QueryKeyStatus:
			res, err = queryKeygenStatus(ctx, k, v, path[1])
		case QueryRecovery:
			res, err = queryRecovery(ctx, k, s, path[1])
		case QueryKeyID:
			res, err = queryKeyID(ctx, k, n, path[1], path[2])
		case QueryKeySharesByKeyID:
			res, err = queryKeySharesByKeyID(ctx, k, s, path[1])
		case QueryKeySharesByValidator:
			res, err = queryKeySharesByValidator(ctx, k, n, s, path[1])
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown tss query endpoint: %s", path[0]))
		}

		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
		}
		return res, nil
	}
}

func queryRecovery(ctx sdk.Context, k tssTypes.TSSKeeper, s tssTypes.Snapshotter, keyID string) ([]byte, error) {
	counter, ok := k.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("could not obtain snapshot counter for key ID %s", keyID)
	}
	snapshot, ok := s.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("could not obtain snapshot for counter %d", counter)
	}

	threshold, found := k.GetCorruptionThreshold(ctx, keyID)
	if !found {
		return nil, fmt.Errorf("keyID %s has no corruption threshold defined", keyID)
	}

	participants := make([]string, 0, len(snapshot.Validators))
	participantShareCounts := make([]uint32, 0, len(snapshot.Validators))
	for _, validator := range snapshot.Validators {
		participants = append(participants, validator.GetSDKValidator().GetOperator().String())
		participantShareCounts = append(participantShareCounts, uint32(validator.ShareCount))
	}

	infos := k.GetAllRecoveryInfos(ctx, keyID)

	resp := tssTypes.QueryRecoveryResponse{
		Threshold:          int32(threshold),
		PartyUids:          participants,
		PartyShareCounts:   participantShareCounts,
		ShareRecoveryInfos: infos,
	}

	return resp.Marshal()
}

func querySigStatus(ctx sdk.Context, k tssTypes.TSSKeeper, v tssTypes.Voter, sigID string) ([]byte, error) {
	var resp tssTypes.QuerySigResponse
	if sig, ok := k.GetSig(ctx, sigID); ok {
		// poll was successful
		resp := tssTypes.QuerySigResponse{
			VoteStatus: tssTypes.Decided,
			Signature: &tssTypes.Signature{
				R: sig.R.Bytes(),
				S: sig.S.Bytes(),
			},
		}
		return resp.Marshal()
	}

	pollMeta := voting.NewPollKey(tssTypes.ModuleName, sigID)
	poll := v.GetPoll(ctx, pollMeta)

	if poll == nil {
		// poll either never existed or has been closed
		resp.VoteStatus = tssTypes.Unspecified
	} else {
		// poll still open, pending a decision
		resp.VoteStatus = tssTypes.Pending
	}

	return resp.Marshal()
}

func queryKeygenStatus(ctx sdk.Context, k tssTypes.TSSKeeper, v tssTypes.Voter, keyID string) ([]byte, error) {
	var resp tssTypes.QueryKeyResponse

	if key, ok := k.GetKey(ctx, keyID); ok {
		// poll was successful
		resp = tssTypes.QueryKeyResponse{
			VoteStatus: tssTypes.Decided,
			Role:       key.Role,
		}

		return resp.Marshal()
	}

	pollMeta := voting.NewPollKey(tssTypes.ModuleName, keyID)
	poll := v.GetPoll(ctx, pollMeta)
	if poll == nil {
		// poll either never existed or has been closed
		resp.VoteStatus = tssTypes.Unspecified
	} else {
		// poll still open, pending a decision
		resp.VoteStatus = tssTypes.Pending
	}

	return resp.Marshal()
}

// queryKeyID returns the keyID of the most recent key for a provided keyChain and keyRole
func queryKeyID(ctx sdk.Context, k tssTypes.TSSKeeper, n tssTypes.Nexus, keyChainStr string, keyRoleStr string) ([]byte, error) {
	keyChain, ok := n.GetChain(ctx, keyChainStr)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", keyChainStr)
	}

	keyRole, err := exported.KeyRoleFromSimpleStr(keyRoleStr)
	if err != nil {
		return nil, err
	}

	keyID, found := k.GetCurrentKeyID(ctx, keyChain, keyRole)
	if !found {
		return nil, fmt.Errorf("no key from chain %s role %s exists", keyChainStr, keyRoleStr)
	}

	return []byte(keyID), nil
}

func queryKeySharesByKeyID(ctx sdk.Context, k tssTypes.TSSKeeper, s tssTypes.Snapshotter, keyID string) ([]byte, error) {

	counter, ok := k.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("invalid keyID %s", keyID)
	}

	snapshot, ok := s.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("no snapshot found for counter number %d", counter)
	}

	
	var allShareInfos []tssTypes.QueryKeyShareResponse_ShareInfo
	for _, validator := range snapshot.Validators {

		thisShareInfo := tssTypes.QueryKeyShareResponse_ShareInfo {
			KeyID:					keyID,
			SnapshotBlockNumber:	snapshot.Height,
			ValidatorAddress:		validator.GetSDKValidator().GetOperator().String(),
			NumValidatorShares:		validator.ShareCount,
			NumTotalShares:			snapshot.TotalShareCount.Int64(),
		}

		allShareInfos = append(allShareInfos, thisShareInfo)
	}

	keyShareInfos := tssTypes.QueryKeyShareResponse {
		ShareInfos:		allShareInfos,
	}

	return keyShareInfos.Marshal()
}

func queryKeySharesByValidator(ctx sdk.Context, k tssTypes.TSSKeeper, n tssTypes.Nexus, s tssTypes.Snapshotter, targetValidatorAddr string) ([]byte, error) {

	var allShareInfos []tssTypes.QueryKeyShareResponse_ShareInfo

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
				if validatorAddr == targetValidatorAddr{

					thisShareInfo := tssTypes.QueryKeyShareResponse_ShareInfo {
						KeyID:					keyID,
						KeyChain:				chain.Name,
						KeyRole:				keyRole.String(),
						SnapshotBlockNumber:	snapshot.Height,
						ValidatorAddress:		validator.GetSDKValidator().GetOperator().String(),
						NumValidatorShares:		validator.ShareCount,
						NumTotalShares:			snapshot.TotalShareCount.Int64(),
					}
					allShareInfos = append(allShareInfos, thisShareInfo)
					break
				}
			}
		}
	}
	
	keyShareInfos := tssTypes.QueryKeyShareResponse {
		ShareInfos:		allShareInfos,
	}

	return keyShareInfos.Marshal()
}
