package keeper

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	"github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigtypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	tsstypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func setup() (sdk.Context, BaseKeeper) {
	encCfg := params.MakeEncodingConfig()

	encCfg.InterfaceRegistry.RegisterImplementations((*codec.ProtoMarshaler)(nil),
		&multisigtypes.MultiSig{},
	)

	paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"))
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper := NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("evm"), paramsK)

	for _, params := range types.DefaultParams() {
		keeper.ForChain(params.Chain).SetParams(ctx, params)
	}

	return ctx, keeper
}

func TestGetMigrationHandler(t *testing.T) {
	var (
		ctx     sdk.Context
		keeper  BaseKeeper
		handler func(ctx sdk.Context) error
	)

	evmChains := []nexus.Chain{exported.Ethereum}
	tokens := []types.ERC20TokenMetadata{
		{
			Asset: rand.NormalizedStr(5),
			Details: types.TokenDetails{
				TokenName: rand.NormalizedStr(5),
				Symbol:    rand.NormalizedStr(5),
				Decimals:  8,
				Capacity:  sdk.ZeroInt(),
			},
			Status:     types.Confirmed,
			IsExternal: true,
			BurnerCode: types.DefaultParams()[0].Burnable,
		},
		{
			Asset: rand.NormalizedStr(5),
			Details: types.TokenDetails{
				TokenName: rand.NormalizedStr(5),
				Symbol:    rand.NormalizedStr(5),
				Decimals:  8,
				Capacity:  sdk.ZeroInt(),
			},
			Status:     types.Pending,
			IsExternal: false,
			BurnerCode: types.DefaultParams()[0].Burnable,
		},
		{
			Asset: rand.NormalizedStr(5),
			Details: types.TokenDetails{
				TokenName: rand.NormalizedStr(5),
				Symbol:    rand.NormalizedStr(5),
				Decimals:  8,
				Capacity:  sdk.ZeroInt(),
			},
			Status:     types.Pending,
			IsExternal: true,
			BurnerCode: types.DefaultParams()[0].Burnable,
		},
	}

	givenMigrationHandler := Given("the migration handler", func() {
		ctx, keeper = setup()
		nexus := mock.NexusMock{
			GetChainsFunc: func(_ sdk.Context) []nexus.Chain {
				return evmChains
			},
		}

		handler = GetMigrationHandler(keeper, &nexus, &mock.SignerMock{}, &mock.MultisigKeeperMock{})
	})

	whenTokensAreSetup := givenMigrationHandler.
		When("tokens are setup for evm chains", func() {
			for _, chain := range evmChains {
				for _, token := range tokens {
					keeper.ForChain(chain.Name).(chainKeeper).setTokenMetadata(ctx, token)
				}
			}
		})

	whenTokensAreSetup.
		When("migration runs", func() {
			err := handler(ctx)
			assert.NoError(t, err)
		}).
		Then("should remove burner code for external tokens", func(t *testing.T) {
			for _, chain := range evmChains {
				ck := keeper.ForChain(chain.Name).(chainKeeper)

				for _, meta := range ck.getTokensMetadata(ctx) {
					if meta.IsExternal {
						assert.Nil(t, meta.BurnerCode)
					} else {
						assert.Equal(t, meta.BurnerCode, types.DefaultParams()[0].Burnable)
					}
				}
			}
		}).Run(t)

	givenMigrationHandler.
		When("EndBlockerLimit param is not set", func() {
			for _, chain := range evmChains {
				ck := keeper.ForChain(chain.Name).(chainKeeper)
				subspace, _ := ck.getSubspace(ctx)
				subspace.Set(ctx, types.KeyEndBlockerLimit, int64(0))
			}

		}).
		Then("should set EndBlockerLimit param", func(t *testing.T) {
			for _, chain := range evmChains {
				ck := keeper.ForChain(chain.Name).(chainKeeper)
				assert.Zero(t, ck.GetParams(ctx).EndBlockerLimit)
			}

			err := handler(ctx)
			assert.NoError(t, err)

			for _, chain := range evmChains {
				ck := keeper.ForChain(chain.Name).(chainKeeper)
				assert.Equal(t, types.DefaultParams()[0].EndBlockerLimit, ck.GetParams(ctx).EndBlockerLimit)
			}
		})
}

func TestMigrateCommandBatchSignature(t *testing.T) {
	var (
		ctx       sdk.Context
		keeper    BaseKeeper
		signer    *mock.SignerMock
		multisigK *mock.MultisigKeeperMock

		keys           map[multisig.KeyID]multisig.Key
		tssMultiSign   map[string]tsstypes.MultisigSignInfo
		commandBatches []types.CommandBatchMetadata
	)

	evmChains := []nexus.Chain{exported.Ethereum}

	givenSetup := Given("a context and keeper", func() {
		ctx, keeper = setup()
	}).
		Given("non-empty signed command batches", func() {
			for {
				if commandBatches = testutils.RandomBatches(); len(commandBatches) != 0 {
					break
				}
			}
			commandBatches = slices.Map(commandBatches, func(c types.CommandBatchMetadata) types.CommandBatchMetadata {
				c.Status = types.BatchSigned
				return c
			})
		}).
		Given("tss multisig sign info", func() {
			keys = make(map[multisig.KeyID]multisig.Key, len(commandBatches))
			tssMultiSign = make(map[string]tsstypes.MultisigSignInfo, len(commandBatches))

			slices.ForEach(commandBatches, func(batch types.CommandBatchMetadata) {
				participants := slices.Expand(func(int) sdk.ValAddress { return rand.ValAddr() }, int(rand.I64Between(5, 10)))

				pubKeys := make(map[string]multisig.PublicKey, len(participants))
				sigID := hex.EncodeToString(batch.ID)
				payloadHash := batch.SigHash.Bytes()

				var infos []*tsstypes.MultisigInfo_Info
				for _, p := range participants {
					sk := funcs.Must(btcec.NewPrivateKey(btcec.S256()))
					pubKeys[p.String()] = sk.PubKey().SerializeCompressed()

					sigKeyPair := tss.SigKeyPair{
						PubKey:    pubKeys[p.String()],
						Signature: funcs.Must(sk.Sign(payloadHash)).Serialize(),
					}

					infos = append(infos, &tsstypes.MultisigInfo_Info{Participant: p, Data: [][]byte{funcs.Must(sigKeyPair.Marshal())}})
				}

				if batch.KeyID != commandBatches[0].KeyID {
					keys[batch.KeyID] = &multisigtypes.Key{
						ID:      batch.KeyID,
						PubKeys: pubKeys,
					}
				}

				tssMultiSign[sigID] = tsstypes.MultisigSignInfo(&tsstypes.MultisigInfo{Infos: infos})
			})
		}).
		Given("mock keeps", func() {
			signer = &mock.SignerMock{
				GetMultisigSignInfoFunc: func(ctx sdk.Context, sigID string) (tsstypes.MultisigSignInfo, bool) {
					info, ok := tssMultiSign[sigID]
					return info, ok
				},
				GetKeyFunc: func(ctx sdk.Context, keyID tss.KeyID) (tss.Key, bool) {
					return tss.Key{Role: tss.SecondaryKey}, true
				},
			}

			multisigK = &mock.MultisigKeeperMock{
				GetKeyFunc: func(ctx sdk.Context, keyID multisig.KeyID) (multisig.Key, bool) {
					key, ok := keys[keyID]
					return key, ok
				},
			}
		})

	givenSetup.
		When("command batches are set", func() {
			slices.ForEach(evmChains, func(chain nexus.Chain) {
				ck := keeper.ForChain(chain.Name).(chainKeeper)
				ck.setLatestBatchMetadata(ctx, commandBatches[len(commandBatches)-1])
				slices.ForEach(commandBatches, func(c types.CommandBatchMetadata) {
					ck.setCommandBatchMetadata(ctx, c)
				})
			})
		}).
		Then("should migrate active key signature", func(t *testing.T) {
			slices.ForEach(evmChains, func(chain nexus.Chain) {
				ck := keeper.ForChain(chain.Name).(chainKeeper)
				err := migrateCommandBatchSignature(ctx, ck, signer, multisigK)
				assert.NoError(t, err)

				commandBatches2 := ck.getCommandBatchesMetadata(ctx)
				assert.Equal(t, len(commandBatches), len(commandBatches2))

				slices.ForEach(commandBatches2[1:], func(c types.CommandBatchMetadata) {
					// key for the first batch is not active
					if c.KeyID == commandBatches[0].KeyID {
						assert.Nil(t, c.Signature)
						return
					}

					assert.NotEmpty(t, c.Signature)

					signature := c.Signature.GetCachedValue().(codec.ProtoMarshaler).(*multisigtypes.MultiSig)
					assert.Equal(t, c.KeyID, signature.KeyID)
					assert.True(t, bytes.Equal(c.SigHash.Bytes(), signature.PayloadHash))
					assert.Equal(t, keys[c.KeyID].GetParticipants(), signature.GetParticipants())
				})
			})
		}).
		Run(t, 20)
}
