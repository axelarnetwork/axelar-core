package keeper_test

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	tssTestUtils "github.com/axelarnetwork/axelar-core/x/tss/exported/testutils"
	tsstypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

func TestQueryDepositAddress(t *testing.T) {
	var (
		btcKeeper   *mock.BTCKeeperMock
		signer      *mock.SignerMock
		nexusKeeper *mock.NexusMock
		ctx         sdk.Context

		address string
	)

	externalKeyThreshold := tsstypes.DefaultParams().ExternalMultisigThreshold.Numerator
	externalKeyCount := tsstypes.DefaultParams().ExternalMultisigThreshold.Denominator
	externalKeys := make([]tss.Key, externalKeyCount)
	for i := 0; i < int(externalKeyCount); i++ {
		externalPrivKey, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			panic(err)
		}
		externalKeys[i] = tss.Key{
			ID:    tssTestUtils.RandKeyID(),
			Value: externalPrivKey.PublicKey,
			Role:  tss.ExternalKey,
		}
	}

	setup := func() {
		btcKeeper = &mock.BTCKeeperMock{}
		signer = &mock.SignerMock{}
		nexusKeeper = &mock.NexusMock{}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

		address = fmt.Sprintf("0x%s", hex.EncodeToString(rand.Bytes(20)))
	}

	t.Run("should return error if the given chain is unknown", testutils.Func(func(t *testing.T) {
		setup()

		params := types.DepositQueryParams{Chain: "unknown", Address: address}

		nexusKeeper.GetChainFunc = func(_ sdk.Context, _ string) (nexus.Chain, bool) {
			return nexus.Chain{}, false
		}

		_, err := keeper.QueryDepositAddress(ctx, btcKeeper, signer, nexusKeeper, types.ModuleCdc.MustMarshalLengthPrefixed(&params))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "recipient chain not found")
	}))

	t.Run("should return error if the secondary key for given chain is not available", testutils.Func(func(t *testing.T) {
		setup()

		params := types.DepositQueryParams{Chain: evm.Ethereum.Name, Address: address}

		nexusKeeper.GetChainFunc = func(_ sdk.Context, chain string) (nexus.Chain, bool) {
			if chain == params.Chain {
				return evm.Ethereum, true
			}
			return nexus.Chain{}, false
		}
		signer.GetCurrentKeyFunc = func(_ sdk.Context, _ nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
			if keyRole == tss.MasterKey {
				return tss.Key{}, true
			}
			return tss.Key{}, false
		}

		_, err := keeper.QueryDepositAddress(ctx, btcKeeper, signer, nexusKeeper, types.ModuleCdc.MustMarshalLengthPrefixed(&params))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secondary key not set")
	}))

	t.Run("should return error if the deposit address is not linked yet", testutils.Func(func(t *testing.T) {
		setup()

		secondaryPrivKey, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			panic(err)
		}
		now := time.Now()
		secondaryKey := tss.Key{
			ID:        tssTestUtils.RandKeyID(),
			Value:     secondaryPrivKey.PublicKey,
			Role:      tss.SecondaryKey,
			RotatedAt: &now,
		}
		params := types.DepositQueryParams{Chain: evm.Ethereum.Name, Address: address}

		nexusKeeper.GetChainFunc = func(_ sdk.Context, chain string) (nexus.Chain, bool) {
			if chain == params.Chain {
				return evm.Ethereum, true
			}
			return nexus.Chain{}, false
		}
		signer.GetCurrentKeyFunc = func(_ sdk.Context, _ nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
			if keyRole == tss.SecondaryKey {
				return secondaryKey, true
			}
			return tss.Key{}, false
		}
		signer.GetKeyFunc = func(ctx sdk.Context, keyID tss.KeyID) (tss.Key, bool) {
			for _, externalKey := range externalKeys {
				if keyID == externalKey.ID {
					return externalKey, true
				}
			}

			return tss.Key{}, false
		}
		signer.GetExternalMultisigThresholdFunc = func(ctx sdk.Context) utils.Threshold {
			return tsstypes.DefaultParams().ExternalMultisigThreshold
		}
		signer.GetExternalKeyIDsFunc = func(ctx sdk.Context, chain nexus.Chain) ([]tss.KeyID, bool) {
			externalKeyIDs := make([]tss.KeyID, len(externalKeys))
			for i := 0; i < len(externalKeyIDs); i++ {
				externalKeyIDs[i] = externalKeys[i].ID
			}

			return externalKeyIDs, true
		}
		btcKeeper.GetNetworkFunc = func(ctx sdk.Context) types.Network { return types.DefaultParams().Network }
		btcKeeper.GetMasterAddressExternalKeyLockDurationFunc = func(ctx sdk.Context) time.Duration { return types.DefaultParams().MasterAddressInternalKeyLockDuration }
		nexusKeeper.GetRecipientFunc = func(_ sdk.Context, _ nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
			return nexus.CrossChainAddress{}, false
		}

		_, err = keeper.QueryDepositAddress(ctx, btcKeeper, signer, nexusKeeper, types.ModuleCdc.MustMarshalLengthPrefixed(&params))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deposit address is not linked with recipient address")
	}))

	t.Run("should return the deposit address", testutils.Func(func(t *testing.T) {
		setup()

		secondaryKey := createRandomKey(tss.SecondaryKey, time.Now())
		params := types.DepositQueryParams{Chain: evm.Ethereum.Name, Address: address}

		nexusKeeper.GetChainFunc = func(_ sdk.Context, chain string) (nexus.Chain, bool) {
			if chain == params.Chain {
				return evm.Ethereum, true
			}
			return nexus.Chain{}, false
		}
		signer.GetCurrentKeyFunc = func(_ sdk.Context, _ nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
			if keyRole == tss.SecondaryKey {
				return secondaryKey, true
			}

			return tss.Key{}, false
		}
		signer.GetKeyFunc = func(ctx sdk.Context, keyID tss.KeyID) (tss.Key, bool) {
			for _, externalKey := range externalKeys {
				if keyID == externalKey.ID {
					return externalKey, true
				}
			}

			return tss.Key{}, false
		}
		btcKeeper.GetMasterAddressExternalKeyLockDurationFunc = func(ctx sdk.Context) time.Duration { return types.DefaultParams().MasterAddressExternalKeyLockDuration }
		btcKeeper.GetNetworkFunc = func(ctx sdk.Context) types.Network { return types.DefaultParams().Network }
		signer.GetExternalMultisigThresholdFunc = func(ctx sdk.Context) utils.Threshold {
			return tsstypes.DefaultParams().ExternalMultisigThreshold
		}
		signer.GetExternalKeyIDsFunc = func(ctx sdk.Context, chain nexus.Chain) ([]tss.KeyID, bool) {
			externalKeyIDs := make([]tss.KeyID, len(externalKeys))
			for i := 0; i < len(externalKeyIDs); i++ {
				externalKeyIDs[i] = externalKeys[i].ID
			}

			return externalKeyIDs, true
		}
		nexusKeeper.GetRecipientFunc = func(_ sdk.Context, _ nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
			return nexus.CrossChainAddress{}, true
		}

		expected := types.QueryAddressResponse{
			Address: types.NewDepositAddress(secondaryKey, externalKeyThreshold, externalKeys, secondaryKey.RotatedAt.Add(types.DefaultParams().MasterAddressExternalKeyLockDuration), nexus.CrossChainAddress{Chain: evm.Ethereum, Address: params.Address}, types.DefaultParams().Network).Address,
			KeyID:   secondaryKey.ID,
		}
		bz, err := keeper.QueryDepositAddress(ctx, btcKeeper, signer, nexusKeeper, types.ModuleCdc.MustMarshalLengthPrefixed(&params))

		var actual types.QueryAddressResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &actual)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	}))
}

func TestQueryConsolidationAddressByKeyID(t *testing.T) {
	var (
		btcKeeper *mock.BTCKeeperMock
		signer    *mock.SignerMock
		ctx       sdk.Context

		keyID tss.KeyID
	)

	setup := func() {
		btcKeeper = &mock.BTCKeeperMock{}
		signer = &mock.SignerMock{}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

		keyID = tssTestUtils.RandKeyID()
	}

	t.Run("should return error if the given key ID cannot be found", testutils.Func(func(t *testing.T) {
		setup()

		signer.GetKeyFunc = func(ctx sdk.Context, keyID tss.KeyID) (tss.Key, bool) { return tss.Key{}, false }

		_, err := keeper.QueryConsolidationAddressByKeyID(ctx, btcKeeper, signer, keyID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("no key with keyID %s found", keyID))
	}))

	t.Run("should return the master consolidation address if given key ID is for a master key", testutils.Func(func(t *testing.T) {
		setup()

		now := time.Now()
		rotationCount := rand.I64Between(100, 1000)
		oldMasterKeyRotationCount := rotationCount - types.DefaultParams().MasterKeyRetentionPeriod
		masterPrivKey, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			panic(err)
		}
		oldMasterPrivKey, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			panic(err)
		}
		masterKey := tss.Key{
			ID:        keyID,
			Value:     masterPrivKey.PublicKey,
			Role:      tss.MasterKey,
			RotatedAt: &now,
		}
		oldMasterKey := tss.Key{
			ID:    tssTestUtils.RandKeyID(),
			Value: oldMasterPrivKey.PublicKey,
			Role:  tss.MasterKey,
		}

		externalKeyCount := tsstypes.DefaultParams().ExternalMultisigThreshold.Denominator
		externalKeys := make([]tss.Key, externalKeyCount)
		for i := 0; i < int(externalKeyCount); i++ {
			externalPrivKey, err := btcec.NewPrivateKey(btcec.S256())
			if err != nil {
				panic(err)
			}
			externalKeys[i] = tss.Key{
				ID:    tssTestUtils.RandKeyID(),
				Value: externalPrivKey.PublicKey,
				Role:  tss.ExternalKey,
			}
		}

		btcKeeper.GetMasterKeyRetentionPeriodFunc = func(ctx sdk.Context) int64 { return types.DefaultParams().MasterKeyRetentionPeriod }
		btcKeeper.GetMasterAddressInternalKeyLockDurationFunc = func(ctx sdk.Context) time.Duration { return types.DefaultParams().MasterAddressInternalKeyLockDuration }
		btcKeeper.GetMasterAddressExternalKeyLockDurationFunc = func(ctx sdk.Context) time.Duration { return types.DefaultParams().MasterAddressExternalKeyLockDuration }
		signer.GetExternalMultisigThresholdFunc = func(ctx sdk.Context) utils.Threshold {
			return tsstypes.DefaultParams().ExternalMultisigThreshold
		}
		signer.GetExternalKeyIDsFunc = func(ctx sdk.Context, chain nexus.Chain) ([]tss.KeyID, bool) {
			externalKeyIDs := make([]tss.KeyID, len(externalKeys))
			for i := 0; i < len(externalKeyIDs); i++ {
				externalKeyIDs[i] = externalKeys[i].ID
			}

			return externalKeyIDs, true
		}
		btcKeeper.GetNetworkFunc = func(ctx sdk.Context) types.Network { return types.DefaultParams().Network }
		signer.GetKeyFunc = func(ctx sdk.Context, keyID tss.KeyID) (tss.Key, bool) {
			for _, externalKey := range externalKeys {
				if keyID == externalKey.ID {
					return externalKey, true
				}
			}

			if keyID == masterKey.ID {
				return masterKey, true
			}

			return tss.Key{}, false
		}
		signer.GetCurrentKeyFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
			if keyRole == tss.MasterKey {
				return masterKey, true
			}

			return tss.Key{}, false
		}
		signer.GetRotationCountFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) int64 { return rotationCount }
		signer.GetKeyByRotationCountFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole, rotationCount int64) (tss.Key, bool) {
			if rotationCount == oldMasterKeyRotationCount {
				return oldMasterKey, true
			}

			return tss.Key{}, false
		}

		expected := types.QueryAddressResponse{
			Address: types.NewMasterConsolidationAddress(masterKey, oldMasterKey, tsstypes.DefaultParams().ExternalMultisigThreshold.Numerator, externalKeys, now.Add(types.DefaultParams().MasterAddressInternalKeyLockDuration), now.Add(types.DefaultParams().MasterAddressExternalKeyLockDuration), types.DefaultParams().Network).Address,
			KeyID:   masterKey.ID,
		}

		bz, err := keeper.QueryConsolidationAddressByKeyID(ctx, btcKeeper, signer, keyID)
		assert.NoError(t, err)

		var actual types.QueryAddressResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &actual)
		assert.Equal(t, expected, actual)
	}))

	t.Run("should return the secondary consolidation address if given key ID is for a secondary key", testutils.Func(func(t *testing.T) {
		setup()

		secondaryPrivKey, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			panic(err)
		}
		secondaryKey := tss.Key{
			ID:    keyID,
			Value: secondaryPrivKey.PublicKey,
			Role:  tss.SecondaryKey,
		}

		btcKeeper.GetNetworkFunc = func(ctx sdk.Context) types.Network { return types.DefaultParams().Network }
		signer.GetKeyFunc = func(ctx sdk.Context, keyID tss.KeyID) (tss.Key, bool) {
			if keyID == secondaryKey.ID {
				return secondaryKey, true
			}

			return tss.Key{}, false
		}

		expected := types.QueryAddressResponse{
			Address: types.NewSecondaryConsolidationAddress(secondaryKey, types.DefaultParams().Network).Address,
			KeyID:   secondaryKey.ID,
		}

		bz, err := keeper.QueryConsolidationAddressByKeyID(ctx, btcKeeper, signer, keyID)

		var actual types.QueryAddressResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &actual)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	}))
}
