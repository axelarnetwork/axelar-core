package keeper

import (
	"context"
	"fmt"
	"time"

	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

type Keeper struct {
	broadcaster   types.Broadcaster
	stakingKeeper types.Staker // needed only for `GetAllValidators`
	client        tssd.GG18Client
	keygenStream  tssd.GG18_KeygenClient // TODO support multiple concurrent sessions
	signStream    tssd.GG18_SignClient   // TODO support multiple concurrent sessions

	// TODO cruft for grpc; can we get rid of this?
	connection        *grpc.ClientConn
	context           context.Context
	contextCancelFunc context.CancelFunc
}

func NewKeeper(conf types.TssdConfig, logger log.Logger, broadcaster types.Broadcaster, staking types.Staker) (Keeper, error) {
	logger = prepareLogger(logger)

	// TODO don't start gRPC unless I'm a validator?
	// start a gRPC client
	tssdServerAddress := conf.Host + ":" + conf.Port
	logger.Info(fmt.Sprintf("initiate connection to tssd gRPC server: %s", tssdServerAddress))
	conn, err := grpc.Dial(tssdServerAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return Keeper{}, err
	}
	logger.Debug("successful connection to tssd gRPC server")
	client := tssd.NewGG18Client(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour) // TODO config file

	return Keeper{
		broadcaster:       broadcaster,
		stakingKeeper:     staking,
		client:            client,
		connection:        conn,
		context:           ctx,
		contextCancelFunc: cancel,
	}, nil
}

func prepareLogger(logger log.Logger) log.Logger {
	return logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return prepareLogger(ctx.Logger())
}

func (k Keeper) Close(logger log.Logger) error {
	logger = prepareLogger(logger)
	logger.Debug(fmt.Sprintf("initiate Close"))
	k.contextCancelFunc()
	if err := k.connection.Close(); err != nil {
		wrapErr := sdkerrors.Wrap(err, "failure to close connection to server")
		//goland:noinspection GoNilness
		logger.Error(wrapErr.Error())
		return wrapErr
	}
	logger.Debug(fmt.Sprintf("successful Close"))
	return nil
}
