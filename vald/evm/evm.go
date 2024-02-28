package evm

import (
	"context"
	goerrors "errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"

	"github.com/axelarnetwork/axelar-core/sdk-utils/broadcast"
	"github.com/axelarnetwork/axelar-core/utils/errors"
	"github.com/axelarnetwork/axelar-core/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/log"
	rs "github.com/axelarnetwork/utils/monads/results"
	"github.com/axelarnetwork/utils/slices"
)

// ErrNotFinalized is returned when a transaction is not finalized
var ErrNotFinalized = goerrors.New("not finalized")

// ErrTxFailed is returned when a transaction has failed
var ErrTxFailed = goerrors.New("transaction failed")

// Mgr manages all communication with Ethereum
type Mgr struct {
	rpcs                      map[string]rpc.Client
	broadcaster               broadcast.Broadcaster
	validator                 sdk.ValAddress
	proxy                     sdk.AccAddress
	latestFinalizedBlockCache LatestFinalizedBlockCache
}

// NewMgr returns a new Mgr instance
func NewMgr(rpcs map[string]rpc.Client, broadcaster broadcast.Broadcaster, valAddr sdk.ValAddress, proxy sdk.AccAddress, latestFinalizedBlockCache LatestFinalizedBlockCache) *Mgr {
	return &Mgr{
		rpcs:                      rpcs,
		proxy:                     proxy,
		broadcaster:               broadcaster,
		validator:                 valAddr,
		latestFinalizedBlockCache: latestFinalizedBlockCache,
	}
}

func (mgr Mgr) logger(keyvals ...any) log.Logger {
	keyvals = append([]any{"listener", "evm"}, keyvals...)
	return log.WithKeyVals(keyvals...)
}

// ProcessNewChain notifies the operator that vald needs to be restarted/udpated for a new chain
func (mgr Mgr) ProcessNewChain(event *types.ChainAdded) (err error) {
	mgr.logger().Info(fmt.Sprintf("VALD needs to be updated and restarted for new chain %s", event.Chain.String()))
	return nil
}

func (mgr Mgr) isTxReceiptFinalized(chain nexus.ChainName, txReceipt *geth.Receipt, confHeight uint64) (bool, error) {
	client, ok := mgr.rpcs[strings.ToLower(chain.String())]
	if !ok {
		return false, fmt.Errorf("rpc client not found for chain %s", chain.String())
	}

	if mgr.latestFinalizedBlockCache.Get(chain).Cmp(txReceipt.BlockNumber) >= 0 {
		return true, nil
	}

	latestFinalizedBlockNumber, err := client.LatestFinalizedBlockNumber(context.Background(), confHeight)
	if err != nil {
		return false, err
	}

	mgr.latestFinalizedBlockCache.Set(chain, latestFinalizedBlockNumber)

	if latestFinalizedBlockNumber.Cmp(txReceipt.BlockNumber) < 0 {
		return false, nil
	}

	return true, nil
}

func (mgr Mgr) GetTxReceiptIfFinalized(chain nexus.ChainName, txID common.Hash, confHeight uint64) (*geth.Receipt, error) {
	client, ok := mgr.rpcs[strings.ToLower(chain.String())]
	if !ok {
		return nil, fmt.Errorf("rpc client not found for chain %s", chain.String())
	}

	txReceipt, err := client.TransactionReceipt(context.Background(), txID)
	keyvals := []interface{}{"chain", chain.String(), "tx_id", txID.Hex()}
	logger := mgr.logger(keyvals...)
	if err == ethereum.NotFound {
		logger.Debug(fmt.Sprintf("transaction receipt %s not found", txID.Hex()))
		return nil, nil
	}
	if err != nil {
		return nil, sdkerrors.Wrap(errors.With(err, keyvals...), "failed getting transaction receipt")
	}

	if txReceipt.Status != geth.ReceiptStatusSuccessful {
		return nil, nil
	}

	isFinalized, err := mgr.isTxReceiptFinalized(chain, txReceipt, confHeight)
	if err != nil {
		return nil, sdkerrors.Wrapf(errors.With(err, keyvals...), "cannot determine if the transaction %s is finalized", txID.Hex())
	}
	if !isFinalized {
		logger.Debug(fmt.Sprintf("transaction %s in block %s not finalized", txID.Hex(), txReceipt.BlockNumber.String()))

		return nil, nil
	}

	return txReceipt, nil
}

// GetTxReceiptsIfFinalized retrieves receipts for provided transaction IDs, only if they're finalized.
func (mgr Mgr) GetTxReceiptsIfFinalized(chain nexus.ChainName, txIDs []common.Hash, confHeight uint64) ([]rs.Result[*geth.Receipt], error) {
	client, ok := mgr.rpcs[strings.ToLower(chain.String())]
	if !ok {
		return nil, fmt.Errorf("rpc client not found for chain %s", chain.String())
	}

	results, err := client.TransactionReceipts(context.Background(), txIDs)
	if err != nil {
		return nil, sdkerrors.Wrapf(errors.With(err, "chain", chain.String(), "tx_ids", txIDs),
			"cannot get transaction receipts")
	}

	isFinalized := func(receipt *geth.Receipt) rs.Result[*geth.Receipt] {
		if receipt.Status != geth.ReceiptStatusSuccessful {
			return rs.FromErr[*geth.Receipt](ErrTxFailed)
		}

		isFinalized, err := mgr.isTxReceiptFinalized(chain, receipt, confHeight)
		if err != nil {
			return rs.FromErr[*geth.Receipt](sdkerrors.Wrapf(errors.With(err, "chain", chain.String()),
				"cannot determine if the transaction %s is finalized", receipt.TxHash.Hex()),
			)
		}

		if !isFinalized {
			return rs.FromErr[*geth.Receipt](ErrNotFinalized)
		}

		return rs.FromOk(receipt)
	}

	return slices.Map(results, func(r rpc.Result) rs.Result[*geth.Receipt] {
		return rs.Pipe(rs.Result[*geth.Receipt](r), isFinalized)
	}), nil
}

// isParticipantOf checks if the validator is in the poll participants list
func (mgr Mgr) isParticipantOf(participants []sdk.ValAddress) bool {
	return slices.Any(participants, func(v sdk.ValAddress) bool { return v.Equals(mgr.validator) })
}
