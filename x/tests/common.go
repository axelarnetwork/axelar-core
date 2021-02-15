package tests

import (
	"context"

	tssd "github.com/axelarnetwork/tssd/pb"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	sdkExported "github.com/cosmos/cosmos-sdk/x/staking/exported"
	"google.golang.org/grpc"

	"github.com/axelarnetwork/axelar-core/testutils"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	btcMock "github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	ethMock "github.com/axelarnetwork/axelar-core/x/ethereum/types/mock"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/types/mock"
	tssdMock "github.com/axelarnetwork/axelar-core/x/tss/types/mock"
)

func randomSender2(validators2 []staking.Validator, validatorsCount int64) sdk.AccAddress {
	return sdk.AccAddress(validators2[testutils.RandIntBetween(0, validatorsCount)].OperatorAddress)
}
func createMocks2(validators2 *[]staking.Validator) testMocks2 {
	stakingKeeper := &snapMock.StakingKeeperMock{
		IterateLastValidatorsFunc: func(ctx sdk.Context, fn func(index int64, validator sdkExported.ValidatorI) (stop bool)) {
			for j, val := range *validators2 {
				if fn(int64(j), val) {
					break
				}
			}
		},
		GetLastTotalPowerFunc: func(ctx sdk.Context) sdk.Int {
			totalPower := sdk.ZeroInt()
			for _, val := range *validators2 {
				totalPower = totalPower.AddRaw(val.ConsensusPower())
			}
			return totalPower
		},
	}

	btcClient := &btcMock.RPCClientMock{
		SendRawTransactionFunc: func(tx *wire.MsgTx, _ bool) (*chainhash.Hash, error) {
			hash := tx.TxHash()
			return &hash, nil
		},
		NetworkFunc: func() btcTypes.Network { return btcTypes.Mainnet }}

	ethClient := &ethMock.RPCClientMock{
		// TODO add functions when needed
	}

	keygen := &tssdMock.TSSDKeyGenClientMock{}
	sign := &tssdMock.TSSDSignClientMock{}
	tssdClient := &tssdMock.TSSDClientMock{
		KeygenFunc: func(context.Context, ...grpc.CallOption) (tssd.GG18_KeygenClient, error) { return keygen, nil },
		SignFunc:   func(context.Context, ...grpc.CallOption) (tssd.GG18_SignClient, error) { return sign, nil },
	}
	return testMocks2{
		BTC:    btcClient,
		ETH:    ethClient,
		TSSD:   tssdClient,
		Keygen: keygen,
		Sign:   sign,
		Staker: stakingKeeper,
	}
}
