package keeper

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"math/big"
	"strings"
	"testing"

	//elliptic "crypto/elliptic"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	tssTestUtils "github.com/axelarnetwork/axelar-core/x/tss/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/tss/types/mock"
)

func TestContractDeployment(t *testing.T) {
	kind := "ADMIN"
	addresses := []struct {
		ID      string
		Address common.Address
	}{
		{
			ID: "eth-external-1",
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("1e89a593d04185a9abb9a7753e6ee5bb70dd49a6551623c090264d8d1d678acf")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("bc0e9ee520cf11c7950dceef6f68e0739d715a4b10b8597be7d6ec9f3193d8e7")),
			}),
		},
		{
			ID: "eth-external-2",
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("d12dc3b0117d3551deaa344fbb97e3f76e557c098ea4f2bc507e77c76b87cf68")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("a3de08e2ead1cd2337883a4da4213b95b0283cae97915c8d44103a01362482b9")),
			}),
		},
		{
			ID: "eth-external-3",
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("d409ced993cfcac9146b1e40deb16d40c9c60690c812cabaac86da17254fcd4c")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("eaedbd4ff9f073469c69e9b479dbea80bdf7ab6864a637138f7160395245bc4f")),
			}),
		},
		{
			ID: "eth-external-4",
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("94635995ce17029517838e3941fff75decece54dd4df6ecc0d4d5d841741cd54")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("d6bf405c7d4be8fa7a5bae426910c5275c50907502909e8350a93a979bc899f9")),
			}),
		},
		{
			ID: "eth-external-5",
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("205a7a166d86f2cf194e20573ecf3bb76ae6d9b9c1f4b841795051a19d73480b")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("caad5c292dd34d02d4204d184218aeed73d68a269d72a4794e71e7048ddb859c")),
			}),
		},
		{
			ID: "eth-external-6",
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("138f7f8aa146075f217e4e7ba9421ac8428e6d7a42fb09bb79b5b81468872201")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("27aa517402511702ee8055708c8594a56f4407c407719b2ae9c90f9839588ae1")),
			}),
		},
		{
			ID: "eth-external-7",
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("bb59d6160ef139a20e15bfe11aef4cf94e61da887c754f79318d79b06c89d39a")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("f51601175e2f8d31aa04195e98b87d2e2b3d44a4119361917f60c5e5d485a797")),
			}),
		},
		{
			ID: "eth-external-6",
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("0721b8bb663e85d8e91db0b29b838b340a3ade52b12c3ad8cf9f0049fb2d3e74")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("d2b425ff05aa2643eeb34b69fec4175adce99a215e4d24a1f4b78c77bda9f4c6")),
			}),
		},
	}
	threshold := uint8(
		sdk.NewDec(int64(len(addresses))).
			MulInt64(4).
			QuoInt64(8).
			Ceil().
			RoundInt64(),
	)
	var addrSlice []string
	for _, address := range addresses {
		addrSlice = append(addrSlice, "\""+address.Address.Hex()+"\"")
	}
	t.Logf("\n\n%sS=\"[%s]\"\n%s=\"%d\"\n\n", kind, strings.Join(addrSlice, ","), kind+"_THRESHOLD", threshold)

	kind = "OWNER"
	threshold = 3
	addresses = []struct {
		ID      string
		Address common.Address
	}{
		{
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("cd30098c9f061a5b88e6ab825a4dcb48c0d9f8a71ba8df87ea774218078ac353")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("c392ce1a541d3c234396ce13825b88f593483aa987eb5bc70aaec9af735524ec")),
			}),
		},
		{
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("2d15786dd30b0d053bd1522b70f93e3c0b710528cfc6b89f5be88130fc6efadd")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("95091762cafbb31044d36ddb07c41f4a3362f8efe66e6c1e7aa9f002765bcdc5")),
			}),
		},
		{
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("81f81545d69c24f41dc70953ccc23aa37d2d2116610edc9f1f7f119de65b7537")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("a6e960805dae09541d24b1bed26d598252549c6afd97a6e34ef31326af7f6783")),
			}),
		},
		{
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("46c49e6e91d1fe7002035aec6c28d25f516225c114186589b8a4d200a56d74f1")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("7c691df309a76e603ca8323800ae1a10f58053d1a501fec0ff357a6c55e66196")),
			}),
		},
		{
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("56726c3f67c4a8d2c0275286c5f01b85dda596d2a64fbf75ad99be74dfbee6f0")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("96103ce5e955dff2d2c4201b04751c92d5c8c41b4cb70feea5ab66561c9cd75c")),
			}),
		},
	}

	addrSlice = make([]string, 0)
	for _, address := range addresses {
		addrSlice = append(addrSlice, "\""+address.Address.Hex()+"\"")
	}
	t.Logf("\n\n%sS=\"[%s]\"\n%s=\"%d\"\n\n", kind, strings.Join(addrSlice, ","), kind+"_THRESHOLD", threshold)

	kind = "OPERATOR"
	threshold = 3
	addresses = []struct {
		ID      string
		Address common.Address
	}{
		{
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("c2226dbf9ac2f6749e9a546134d274322e066953a37fe9654d153ef83c830d4b")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("24cbece48227194e67471180449b4edfb5f0298ddd13341926c9a0a3dbe9dddd")),
			}),
		},
		{
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("e0212453c984226136038498f738177c2b8a6a3e7078f6da1efc2c7834410dfa")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("9c5a154115dca4e6eaeb6035ba291e5c0ca932d307900c8493a7b2a96ab30734")),
			}),
		},
		{
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("8305954cbfb745acf1da3e27e70d891cb4e65e926d62d888173e003e88be61e1")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("b63ea563484a215953aea89aa6a5cea982bdfc9555cf50ad05d39a67d47685ff")),
			}),
		},
		{
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("1db697c955a2d0c878447acc1fabde289fc0c92bbaa404d66453a7f29ad1e4aa")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("63430f626fbaee1e8921a594befe912f3fb58eafa84d80ee277da595794c256b")),
			}),
		},
		{
			Address: crypto.PubkeyToAddress(ecdsa.PublicKey{
				X: big.NewInt(0).SetBytes(common.Hex2Bytes("41e442d79c21b66508bb25e026047b10cf7244714e9d597689a1fc2bf958ae9f")),
				Y: big.NewInt(0).SetBytes(common.Hex2Bytes("d53bc3493c952a217a87075f7f9b3f097eb560ee4511997eb6ee2e0a6a6d445c")),
			}),
		},
	}

	addrSlice = make([]string, 0)
	for _, address := range addresses {
		addrSlice = append(addrSlice, "\""+address.Address.Hex()+"\"")
	}
	t.Logf("\n\n%sS=\"[%s]\"\n%s=\"%d\"\n\n", kind, strings.Join(addrSlice, ","), kind+"_THRESHOLD", threshold)
}

func TestMsgServer_RotateKey(t *testing.T) {
	var (
		server    types.MsgServiceServer
		ctx       sdk.Context
		tssKeeper *mock.TSSKeeperMock
	)
	setup := func() {
		tssKeeper = &mock.TSSKeeperMock{
			RotateKeyFunc:     func(sdk.Context, nexus.Chain, exported.KeyRole) error { return nil },
			LoggerFunc:        func(ctx sdk.Context) log.Logger { return ctx.Logger() },
			AssignNextKeyFunc: func(sdk.Context, nexus.Chain, exported.KeyRole, exported.KeyID) error { return nil },
			AssertMatchesRequirementsFunc: func(ctx sdk.Context, snapshotter snapshot.Snapshotter, chain nexus.Chain, keyID exported.KeyID, keyRole exported.KeyRole) error {
				return nil
			},
		}
		snapshotter := &mock.SnapshotterMock{}
		staker := &mock.StakingKeeperMock{}
		voter := &mock.VoterMock{}
		nexusKeeper := &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain string) (nexus.Chain, bool) {
				return nexus.Chain{
					Name:                  chain,
					SupportsForeignAssets: true,
					Module:                rand.Str(10),
				}, true
			},
		}
		rewarder := &mock.RewarderMock{}
		server = NewMsgServerImpl(tssKeeper, snapshotter, staker, voter, nexusKeeper, rewarder)
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
	}
	repeats := 20
	t.Run("first key rotation", testutils.Func(func(t *testing.T) {
		setup()
		tssKeeper.GetCurrentKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) { return "", false }
		tssKeeper.GetNextKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) { return "", true }

		_, err := server.RotateKey(sdk.WrapSDKContext(ctx), &types.RotateKeyRequest{
			Sender:  rand.AccAddr(),
			Chain:   rand.StrBetween(5, 20),
			KeyRole: exported.KeyRole(rand.I64Between(1, 3)),
			KeyID:   tssTestUtils.RandKeyID(),
		})

		assert.NoError(t, err)
		assert.Len(t, tssKeeper.AssignNextKeyCalls(), 1)
		assert.Len(t, tssKeeper.RotateKeyCalls(), 1)
	}).Repeat(repeats))

	t.Run("next key is assigned", testutils.Func(func(t *testing.T) {
		setup()
		keyID := tssTestUtils.RandKeyID()
		tssKeeper.GetCurrentKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) {
			return tssTestUtils.RandKeyID(), true
		}
		tssKeeper.GetNextKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) { return keyID, true }

		_, err := server.RotateKey(sdk.WrapSDKContext(ctx), &types.RotateKeyRequest{
			Sender:  rand.AccAddr(),
			Chain:   rand.StrBetween(5, 20),
			KeyRole: exported.KeyRole(rand.I64Between(1, 3)),
			KeyID:   keyID,
		})

		assert.Error(t, err)
		assert.Len(t, tssKeeper.AssignNextKeyCalls(), 0)
		assert.Len(t, tssKeeper.RotateKeyCalls(), 0)
	}).Repeat(repeats))

	t.Run("no next key is assigned", testutils.Func(func(t *testing.T) {
		setup()
		tssKeeper.GetCurrentKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) {
			return tssTestUtils.RandKeyID(), true
		}
		tssKeeper.GetNextKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) { return "", false }

		_, err := server.RotateKey(sdk.WrapSDKContext(ctx), &types.RotateKeyRequest{
			Sender:  rand.AccAddr(),
			Chain:   rand.StrBetween(5, 20),
			KeyRole: exported.KeyRole(rand.I64Between(1, 3)),
			KeyID:   tssTestUtils.RandKeyID(),
		})

		assert.Error(t, err)
	}).Repeat(repeats))
}

func TestMsgServer_SubmitMultisigPubKey(t *testing.T) {
	var (
		server    types.MsgServiceServer
		ctx       sdk.Context
		tssKeeper *mock.TSSKeeperMock
		randSnap  snapshot.Snapshot
	)
	setup := func() {
		randSnap = randSnapshot()
		tssKeeper = &mock.TSSKeeperMock{
			LoggerFunc:                     func(ctx sdk.Context) log.Logger { return ctx.Logger() },
			GetSnapshotCounterForKeyIDFunc: func(sdk.Context, exported.KeyID) (int64, bool) { return rand.PosI64(), true },
			SubmitPubKeysFunc:              func(sdk.Context, exported.KeyID, sdk.ValAddress, ...[]byte) bool { return true },
			IsMultisigKeygenCompletedFunc:  func(sdk.Context, exported.KeyID) bool { return false },
			GetMultisigKeygenInfoFunc:      func(sdk.Context, exported.KeyID) (types.MultisigKeygenInfo, bool) { return &types.MultisigInfo{}, true },
			SetKeyFunc:                     func(ctx sdk.Context, key exported.Key) {},
		}
		snapshotter := &mock.SnapshotterMock{
			GetOperatorFunc: func(sdk.Context, sdk.AccAddress) sdk.ValAddress {
				return randSnap.Validators[0].GetSDKValidator().GetOperator()
			},
			GetSnapshotFunc: func(sdk.Context, int64) (snapshot.Snapshot, bool) { return randSnap, true },
		}
		staker := &mock.StakingKeeperMock{}
		voter := &mock.VoterMock{}
		nexusKeeper := &mock.NexusMock{}
		rewarder := &mock.RewarderMock{}
		server = NewMsgServerImpl(tssKeeper, snapshotter, staker, voter, nexusKeeper, rewarder)
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
	}
	repeats := 20
	t.Run("should return error when a validator submits for a completed multisig keygen", testutils.Func(func(t *testing.T) {
		setup()
		tssKeeper.IsMultisigKeygenCompletedFunc = func(sdk.Context, exported.KeyID) bool { return true }

		sigKeyPairs := newSigKeyPair(randSnap.Validators[0])

		_, err := server.SubmitMultisigPubKeys(sdk.WrapSDKContext(ctx), &types.SubmitMultisigPubKeysRequest{
			Sender:      rand.AccAddr(),
			KeyID:       tssTestUtils.RandKeyID(),
			SigKeyPairs: sigKeyPairs,
		})

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should return error when a validator submits incorrect number of pub keys", testutils.Func(func(t *testing.T) {
		setup()

		sigKeyPairs := newSigKeyPair(randSnap.Validators[0])
		sigKeyPairs = append(sigKeyPairs, sigKeyPairs[0])
		_, err := server.SubmitMultisigPubKeys(sdk.WrapSDKContext(ctx), &types.SubmitMultisigPubKeysRequest{
			Sender:      rand.AccAddr(),
			KeyID:       tssTestUtils.RandKeyID(),
			SigKeyPairs: sigKeyPairs,
		})

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should return error when a validator submits a duplicate pub key", testutils.Func(func(t *testing.T) {
		setup()
		tssKeeper.SubmitPubKeysFunc = func(sdk.Context, exported.KeyID, sdk.ValAddress, ...[]byte) bool { return false }

		sigKeyPairs := newSigKeyPair(randSnap.Validators[0])
		_, err := server.SubmitMultisigPubKeys(sdk.WrapSDKContext(ctx), &types.SubmitMultisigPubKeysRequest{
			Sender:      rand.AccAddr(),
			KeyID:       tssTestUtils.RandKeyID(),
			SigKeyPairs: sigKeyPairs,
		})

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should return error when a validator submits invalid pub key ", testutils.Func(func(t *testing.T) {
		setup()
		tssKeeper.SubmitPubKeysFunc = func(sdk.Context, exported.KeyID, sdk.ValAddress, ...[]byte) bool { return false }

		sigKeyPairs := newSigKeyPair(randSnap.Validators[0])
		idx := rand.I64Between(0, int64(len(sigKeyPairs)))
		sigKeyPairs[int(idx)].PubKey = rand.Bytes(33)

		_, err := server.SubmitMultisigPubKeys(sdk.WrapSDKContext(ctx), &types.SubmitMultisigPubKeysRequest{
			Sender:      rand.AccAddr(),
			KeyID:       tssTestUtils.RandKeyID(),
			SigKeyPairs: sigKeyPairs,
		})

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should submit pub keys when signature is valid", func(t *testing.T) {
		setup()
		pubKeyInfos := newSigKeyPair(randSnap.Validators[0])

		_, err := server.SubmitMultisigPubKeys(sdk.WrapSDKContext(ctx), &types.SubmitMultisigPubKeysRequest{
			Sender:      rand.AccAddr(),
			KeyID:       tssTestUtils.RandKeyID(),
			SigKeyPairs: pubKeyInfos,
		})

		assert.NoError(t, err)

	})
}

func TestMsgServer_SubmitMultisigSignatures(t *testing.T) {
	var (
		server    types.MsgServiceServer
		ctx       sdk.Context
		tssKeeper *mock.TSSKeeperMock
		randSnap  snapshot.Snapshot
	)
	setup := func() {
		randSnap = randSnapshot()
		signInfo := randSignInfo(randSnap)
		tssKeeper = &mock.TSSKeeperMock{
			LoggerFunc:        func(ctx sdk.Context) log.Logger { return ctx.Logger() },
			GetInfoForSigFunc: func(sdk.Context, string) (exported.SignInfo, bool) { return signInfo, true },
			GetMultisigPubKeysByValidatorFunc: func(sdk.Context, exported.KeyID, sdk.ValAddress) ([]ecdsa.PublicKey, bool) {
				return []ecdsa.PublicKey{}, true
			},
			SubmitSignaturesFunc:          func(sdk.Context, string, sdk.ValAddress, ...[]byte) bool { return true },
			IsMultisigKeygenCompletedFunc: func(sdk.Context, exported.KeyID) bool { return false },
			GetMultisigSignInfoFunc: func(sdk.Context, string) (types.MultisigSignInfo, bool) {
				return &types.MultisigInfo{TargetNum: rand.PosI64()}, true
			},
			GetSigFunc: func(sdk.Context, string) (exported.Signature, exported.SigStatus) {
				return exported.Signature{}, exported.SigStatus_Signing
			},
		}
		snapshotter := &mock.SnapshotterMock{
			GetOperatorFunc: func(sdk.Context, sdk.AccAddress) sdk.ValAddress {
				return randSnap.Validators[0].GetSDKValidator().GetOperator()
			},
			GetSnapshotFunc: func(sdk.Context, int64) (snapshot.Snapshot, bool) { return randSnap, true },
		}
		staker := &mock.StakingKeeperMock{}
		voter := &mock.VoterMock{}
		nexusKeeper := &mock.NexusMock{}
		rewarder := &mock.RewarderMock{}
		server = NewMsgServerImpl(tssKeeper, snapshotter, staker, voter, nexusKeeper, rewarder)
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
	}
	repeats := 20
	t.Run("should return error when cannot find pub keys", testutils.Func(func(t *testing.T) {
		setup()
		tssKeeper.GetMultisigPubKeysByValidatorFunc = func(sdk.Context, exported.KeyID, sdk.ValAddress) ([]ecdsa.PublicKey, bool) {
			return []ecdsa.PublicKey{}, false
		}

		sigs := randSignatures(randSnap.Validators[0])
		_, err := server.SubmitMultisigSignatures(sdk.WrapSDKContext(ctx), &types.SubmitMultisigSignaturesRequest{
			Sender:     rand.AccAddr(),
			SigID:      rand.StrBetween(5, 20),
			Signatures: sigs,
		})
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should return error when a validator submits incorrect number of signatures", testutils.Func(func(t *testing.T) {
		setup()

		sigs := randSignatures(randSnap.Validators[0])
		_, err := server.SubmitMultisigSignatures(sdk.WrapSDKContext(ctx), &types.SubmitMultisigSignaturesRequest{
			Sender:     rand.AccAddr(),
			SigID:      rand.StrBetween(5, 20),
			Signatures: sigs,
		})
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should return error when fail to verify signature ", testutils.Func(func(t *testing.T) {
		setup()

		sigs := randSignatures(randSnap.Validators[0])

		var randPubs []ecdsa.PublicKey
		for i := 0; i < len(sigs); i++ {
			privKey, _ := btcec.NewPrivateKey(btcec.S256())
			pk := btcec.PublicKey(privKey.PublicKey)
			pubkey := pk.ToECDSA()
			randPubs = append(randPubs, *pubkey)
		}
		tssKeeper.GetMultisigPubKeysByValidatorFunc = func(sdk.Context, exported.KeyID, sdk.ValAddress) ([]ecdsa.PublicKey, bool) { return randPubs, true }

		_, err := server.SubmitMultisigSignatures(sdk.WrapSDKContext(ctx), &types.SubmitMultisigSignaturesRequest{
			Sender:     rand.AccAddr(),
			SigID:      rand.StrBetween(5, 20),
			Signatures: sigs,
		})
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should submit valid signatures", testutils.Func(func(t *testing.T) {
		setup()

		var randPubs []ecdsa.PublicKey
		var sigs [][]byte
		for i := 0; i < len(sigs); i++ {
			privKey, _ := btcec.NewPrivateKey(btcec.S256())
			pk := btcec.PublicKey(privKey.PublicKey)
			pubkey := pk.ToECDSA()
			randPubs = append(randPubs, *pubkey)

			d := sha256.Sum256([]byte("message"))
			sig, _ := privKey.Sign(d[:])
			sigs = append(sigs, sig.Serialize())
		}
		tssKeeper.GetMultisigPubKeysByValidatorFunc = func(sdk.Context, exported.KeyID, sdk.ValAddress) ([]ecdsa.PublicKey, bool) { return randPubs, true }

		_, err := server.SubmitMultisigSignatures(sdk.WrapSDKContext(ctx), &types.SubmitMultisigSignaturesRequest{
			Sender:     rand.AccAddr(),
			SigID:      rand.StrBetween(5, 20),
			Signatures: sigs,
		})
		assert.NoError(t, err)
	}).Repeat(repeats))

}

func newSigKeyPair(validator snapshot.Validator) []exported.SigKeyPair {
	var sigKeyPairs []exported.SigKeyPair
	for i := int64(0); i < validator.ShareCount; i++ {
		privKey, _ := btcec.NewPrivateKey(btcec.S256())
		pk := btcec.PublicKey(privKey.PublicKey)
		d := sha256.Sum256([]byte(validator.GetSDKValidator().GetOperator().String()))
		sig, _ := privKey.Sign(d[:])
		sigKeyPairs = append(sigKeyPairs, exported.SigKeyPair{PubKey: pk.SerializeCompressed(), Signature: sig.Serialize()})

	}
	return sigKeyPairs
}

func randSignatures(validator snapshot.Validator) [][]byte {
	var sigs [][]byte
	for i := int64(0); i < validator.ShareCount; i++ {
		privKey, _ := btcec.NewPrivateKey(btcec.S256())
		d := sha256.Sum256([]byte("message"))
		sig, _ := privKey.Sign(d[:])
		sigs = append(sigs, sig.Serialize())

	}
	return sigs
}
