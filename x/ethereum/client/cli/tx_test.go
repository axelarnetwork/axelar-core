package cli

import (
	"math/big"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	ethParams "github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
)

func TestEthToWei_IsInteger(t *testing.T) {
	amount, _ := sdk.NewDecFromStr("3.2")
	eth := sdk.DecCoin{
		Denom:  "eth",
		Amount: amount,
	}
	wei := eth
	wei.Amount = eth.Amount.MulInt64(ethParams.Ether)

	assert.True(t, wei.Amount.IsInteger())
}

func TestGweiToWei_IsNotInteger(t *testing.T) {
	amount, _ := sdk.NewDecFromStr("3.0000000000002")
	gwei := sdk.DecCoin{
		Denom:  "gwei",
		Amount: amount,
	}
	wei := gwei
	wei.Amount = gwei.Amount.MulInt64(ethParams.GWei)

	assert.False(t, wei.Amount.IsInteger())
}

// TestMsgSignTx_CorrectCosmosSigning ensures that msgs containing ethereum transactions can be properly signed and decoded.
// Nesting an Ethereum Transaction type directly inside an sdk.Msg trip up the amino codec, so we need to encode it first.
func TestMsgSignTx_CorrectCosmosSigning(t *testing.T) {
	cdc := testutils.Codec()
	auth.RegisterCodec(cdc)
	accNumber := uint64(123)
	chainID := "test_chain"
	seq := uint64(9876)
	txBldr := auth.NewTxBuilder(
		utils.GetTxEncoder(cdc),
		accNumber,
		seq,
		0,
		1.2,
		false,
		chainID,
		"hello",
		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.ZeroInt())),
		sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(10000, sdk.Precision))},
	)
	keybase, err := keys.NewKeyring(sdk.KeyringServiceName(), keys.BackendTest, "./testdata/", os.Stdin)
	if err != nil {
		panic(err)
	}
	accountName := testutils.RandString(10)
	pw := testutils.RandString(20)
	info, _, err := keybase.CreateMnemonic(accountName, keys.English, pw, keys.Secp256k1)
	if err != nil {
		panic(err)
	}
	txBldr = txBldr.WithKeybase(keybase)

	ethTx := ethTypes.NewContractCreation(
		uint64(testutils.RandIntBetween(0, 10000000)),
		big.NewInt(testutils.RandIntBetween(0, 10000000)),
		uint64(testutils.RandIntBetween(0, 1000000)),
		big.NewInt(testutils.RandIntBetween(0, 10000000)),
		testutils.RandBytes(1000))
	json, err := ethTx.MarshalJSON()
	if err != nil {
		panic(err)
	}

	msg := types.MsgSignTx{
		Sender: info.GetAddress(),
		Tx:     json,
	}

	// msg := types.MsgVoteVerifiedTx{
	// 	Sender:     info.GetAddress(),
	// 	PollMeta:   exported.PollMeta{},
	// 	VotingData: false,
	// }
	bz, err := txBldr.BuildAndSign(accountName, pw, []sdk.Msg{msg})
	if err != nil {
		panic(err)
	}
	tx, err := auth.DefaultTxDecoder(cdc)(bz)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, json, tx.GetMsgs()[0].(types.MsgSignTx).Tx)

	stdmsg, err := txBldr.BuildSignMsg([]sdk.Msg{msg})
	if err != nil {
		panic(err)
	}

	assert.Equal(t, accNumber, stdmsg.AccountNumber)
	assert.Equal(t, chainID, stdmsg.ChainID)

	expectedSig, err := auth.MakeSignature(keybase, accountName, pw, stdmsg)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, expectedSig.Signature, tx.(ante.SigVerifiableTx).GetSignatures()[0])

	authSubspace := subspace.NewSubspace(cdc, sdk.NewKVStoreKey("store"), sdk.NewKVStoreKey("tstore"), "params")
	accountKeeper := auth.NewAccountKeeper(cdc, sdk.NewKVStoreKey("auth"), authSubspace, auth.ProtoBaseAccount)
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{
		ChainID: chainID,
		Height:  testutils.RandIntBetween(1, 1000000),
	}, false, log.TestingLogger())
	acc := accountKeeper.NewAccountWithAddress(ctx, info.GetAddress())
	err = acc.SetSequence(seq)
	if err != nil {
		panic(err)
	}
	err = acc.SetAccountNumber(accNumber)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, stdmsg.Bytes(), tx.(ante.SigVerifiableTx).GetSignBytes(ctx, acc))
	assert.Equal(t, accNumber, acc.GetAccountNumber())
	assert.Equal(t, seq, acc.GetSequence())
	assert.Equal(t, chainID, ctx.ChainID())

	err = acc.SetPubKey(info.GetPubKey())
	if err != nil {
		panic(err)
	}
	accountKeeper.SetAccount(ctx, acc)

	sigVer := ante.NewSigVerificationDecorator(accountKeeper)
	_, err = sigVer.AnteHandle(ctx, tx, false, func(sdk.Context, sdk.Tx, bool) (sdk.Context, error) { return ctx, nil })
	assert.NoError(t, err)
}
