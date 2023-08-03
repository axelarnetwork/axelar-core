package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/bank/keeper"
	"github.com/axelarnetwork/axelar-core/x/bank/types/mock"
	. "github.com/axelarnetwork/utils/test"
)

func TestSendCoins(t *testing.T) {
	var (
		k          keeper.BankKeeper
		bankKeeper *mock.BankKeeperMock
	)

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	fromAddr := rand.AccAddr()
	toAddr := rand.AccAddr()
	amt := sdk.NewCoins(
		sdk.NewCoin("aaa", sdk.NewInt(rand.PosI64())),
		sdk.NewCoin("bbb", sdk.NewInt(rand.PosI64())),
		sdk.NewCoin("ccc", sdk.NewInt(rand.PosI64())),
	)

	Given("a bank keeper", func() {
		bankKeeper = &mock.BankKeeperMock{}
		k = keeper.NewBankKeeper(bankKeeper)
	}).
		Branch(
			When("from address is blocked", func() {
				bankKeeper.BlockedAddrFunc = func(addr sdk.AccAddress) bool {
					return addr.Equals(fromAddr)
				}
			}).
				Then("should return an error", func(t *testing.T) {
					err := k.SendCoins(ctx, fromAddr, toAddr, amt)

					assert.ErrorContains(t, err, fmt.Sprintf("%s is not allowed to send funds", fromAddr.String()))
				}),

			When("to address is blocked", func() {
				bankKeeper.BlockedAddrFunc = func(addr sdk.AccAddress) bool {
					return addr.Equals(toAddr)
				}
			}).
				Then("should return an error", func(t *testing.T) {
					err := k.SendCoins(ctx, fromAddr, toAddr, amt)

					assert.ErrorContains(t, err, fmt.Sprintf("%s is not allowed to receive funds", toAddr.String()))
				}),

			When("from/to address is not blocked", func() {
				bankKeeper.BlockedAddrFunc = func(addr sdk.AccAddress) bool {
					return !addr.Equals(fromAddr) && !addr.Equals(toAddr)
				}
			}).
				When("send coins fails", func() {
					bankKeeper.SendCoinsFunc = func(_ sdk.Context, _, _ sdk.AccAddress, _ sdk.Coins) error {
						return fmt.Errorf("send error")
					}
				}).
				Then("should return an error", func(t *testing.T) {
					err := k.SendCoins(ctx, fromAddr, toAddr, amt)

					assert.ErrorContains(t, err, "send error")
				}),

			When("from/to address is not blocked", func() {
				bankKeeper.BlockedAddrFunc = func(addr sdk.AccAddress) bool {
					return !addr.Equals(fromAddr) && !addr.Equals(toAddr)
				}
			}).
				When("send coins succeeds", func() {
					bankKeeper.SendCoinsFunc = func(_ sdk.Context, _, _ sdk.AccAddress, _ sdk.Coins) error {
						return nil
					}
				}).
				Then("should return an error", func(t *testing.T) {
					err := k.SendCoins(ctx, fromAddr, toAddr, amt)

					assert.NoError(t, err)
				}),
		).
		Run(t)
}

func TestSpendableBalance(t *testing.T) {
	var (
		k          keeper.BankKeeper
		bankKeeper *mock.BankKeeperMock
	)

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	addr := rand.AccAddr()
	spendableCoins := sdk.NewCoins(
		sdk.NewCoin("aaa", sdk.NewInt(rand.PosI64())),
		sdk.NewCoin("bbb", sdk.NewInt(rand.PosI64())),
		sdk.NewCoin("ccc", sdk.NewInt(rand.PosI64())),
	)

	Given("a bank keeper", func() {
		bankKeeper = &mock.BankKeeperMock{}
		k = keeper.NewBankKeeper(bankKeeper)
	}).
		Branch(
			When("denom not in spendable coins", func() {
				bankKeeper.SpendableCoinsFunc = func(sdk.Context, sdk.AccAddress) sdk.Coins {
					return spendableCoins
				}
			}).
				Then("should return zero", func(t *testing.T) {
					coin := k.SpendableBalance(ctx, addr, "ddd")
					assert.Equal(t, sdk.NewCoin("ddd", sdk.ZeroInt()), coin)
				}),

			When("denom in spendable coins", func() {
				bankKeeper.SpendableCoinsFunc = func(sdk.Context, sdk.AccAddress) sdk.Coins {
					return spendableCoins
				}
			}).
				Then("should return coin", func(t *testing.T) {
					coin := k.SpendableBalance(ctx, addr, spendableCoins[2].Denom)
					assert.Equal(t, spendableCoins[2], coin)
				}),
		).
		Run(t)
}
