package evm

import (
	"bytes"
	"fmt"

	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(_ sdk.Context, _ abci.RequestBeginBlock, _ types.BaseKeeper) {}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock, baseKeeper types.BaseKeeper, signer types.Signer, voter types.InitPoller, snapshotter types.Snapshotter, nexus types.Nexus) []abci.ValidatorUpdate {
	txs := baseKeeper.GetScheduledUnsignedTxs(ctx)
	if len(txs) > 0 {
		baseKeeper.Logger(ctx).Info(fmt.Sprintf("processing %d unsigned commands", len(txs)))
	}
	for _, tx := range txs {
		processScheduledTx(ctx, tx, baseKeeper, signer, voter, snapshotter, nexus)
	}
	baseKeeper.DeleteScheduledTxs(ctx)

	cmds := baseKeeper.GetScheduledUnsignedCommands(ctx)
	if len(cmds) > 0 {
		baseKeeper.Logger(ctx).Info(fmt.Sprintf("processing %d unsigned commands", len(cmds)))
	}
	for _, cmd := range cmds {
		processScheduledCommand(ctx, cmd, baseKeeper, signer, voter, snapshotter)
	}
	baseKeeper.DeleteScheduledCommands(ctx)

	return nil
}

func processScheduledTx(
	ctx sdk.Context,
	tx types.ScheduledUnsignedTx,
	baseKeeper types.BaseKeeper,
	signer types.Signer,
	voter types.InitPoller,
	snapshotter types.Snapshotter,
	nexus types.Nexus) {

	keeper := baseKeeper.ForChain(ctx, tx.Chain)
	chain, ok := nexus.GetChain(ctx, tx.Chain)
	if !ok {
		baseKeeper.Logger(ctx).Error(fmt.Sprintf("%s is not a registered chain", tx.Chain))
		return
	}

	snapshot, found := snapshotter.GetSnapshot(ctx, tx.SignInfo.SnapshotCounter)
	if !found {
		baseKeeper.Logger(ctx).Error(fmt.Sprintf("could not find snapshot for counter %d", tx.SignInfo.SnapshotCounter))
		return
	}

	baseKeeper.Logger(ctx).Info(fmt.Sprintf("storing raw tx %s", tx.TxID))
	byteCodes, ok := keeper.GetGatewayByteCodes(ctx)
	if !ok {
		baseKeeper.Logger(ctx).Error(fmt.Sprintf("Could not retrieve gateway bytecodes for chain %s", tx.Chain))
		return
	}

	unsignedTx := keeper.GetUnsignedTx(ctx, tx.TxID)
	if unsignedTx == nil {
		baseKeeper.Logger(ctx).Error(fmt.Sprintf("Could not retrieve unsigned TX '%s' for chain %s", tx.TxID, tx.Chain))
		return
	}

	err := signer.StartSign(ctx, voter, tx.SignInfo.KeyID, tx.SignInfo.SigID, tx.SignInfo.Msg, snapshot)
	if err != nil {
		baseKeeper.Logger(ctx).Error(fmt.Sprintf("error while starting sign for sig ID %s: %s", tx.SignInfo.SigID, err.Error()))
		return
	}

	// if this is the transaction that is deploying Axelar Gateway, calculate and save address
	if unsignedTx.To() == nil && bytes.Equal(unsignedTx.Data(), byteCodes) {

		pub, ok := signer.GetCurrentKey(ctx, chain, tss.MasterKey)
		if !ok {
			baseKeeper.Logger(ctx).Error(fmt.Sprintf("no master key for chain %s found", chain.Name))
			return
		}

		addr := crypto.CreateAddress(crypto.PubkeyToAddress(pub.Value), unsignedTx.Nonce())
		keeper.SetGatewayAddress(ctx, addr)

		telemetry.NewLabel("eth_factory_addr", addr.String())
	}
}

func processScheduledCommand(
	ctx sdk.Context,
	cmd types.ScheduledUnsignedCommand,
	baseKeeper types.BaseKeeper,
	signer types.Signer,
	voter types.InitPoller,
	snapshotter types.Snapshotter) {

	keeper := baseKeeper.ForChain(ctx, cmd.Chain)

	snapshot, found := snapshotter.GetSnapshot(ctx, cmd.SignInfo.SnapshotCounter)
	if !found {
		baseKeeper.Logger(ctx).Error(fmt.Sprintf("could not find snapshot for counter %d", cmd.SignInfo.SnapshotCounter))
		return
	}

	err := signer.StartSign(ctx, voter, cmd.SignInfo.KeyID, cmd.SignInfo.SigID, cmd.SignInfo.Msg, snapshot)
	if err != nil {
		baseKeeper.Logger(ctx).Error(fmt.Sprintf("error while starting sign for sig ID %s: %s", cmd.SignInfo.SigID, err.Error()))
		return
	}

	var commandID types.CommandID
	copy(commandID[:], cmd.CommandID)
	keeper.SetCommandData(ctx, commandID, cmd.CommandData)
}
