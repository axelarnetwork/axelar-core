package codec

import (
	"fmt"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	axelarnettypes "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexustypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
	permissiontypes "github.com/axelarnetwork/axelar-core/x/permission/types"
	rewardtypes "github.com/axelarnetwork/axelar-core/x/reward/types"
	snapshottypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	tsstypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	votetypes "github.com/axelarnetwork/axelar-core/x/vote/types"
)

type customRegistry interface {
	RegisterCustomTypeURL(iface interface{}, typeURL string, impl proto.Message)
}

// RegisterLegacyMsgInterfaces registers the msg codec before the package name
// refactor done in https://github.com/axelarnetwork/axelar-core/commit/2d5e35d7da4fb02ac55fb040fed420954d3be020
// to keep transaction query backwards compatible
func RegisterLegacyMsgInterfaces(registry cdctypes.InterfaceRegistry) {
	r, ok := registry.(customRegistry)
	if !ok {
		panic(fmt.Errorf("failed to convert registry type %T", registry))
	}

	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelar.evm.v1beta1.CreateTransferOwnershipRequest", &evmtypes.CreateTransferOwnershipRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelar.axelarnet.v1beta1.RegisterIBCPathRequest", &axelarnettypes.RegisterIBCPathRequest{})

	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelarnet.v1beta1.LinkRequest", &axelarnettypes.LinkRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelarnet.v1beta1.ConfirmDepositRequest", &axelarnettypes.ConfirmDepositRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelarnet.v1beta1.ExecutePendingTransfersRequest", &axelarnettypes.ExecutePendingTransfersRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelarnet.v1beta1.RegisterIBCPathRequest", &axelarnettypes.RegisterIBCPathRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelarnet.v1beta1.AddCosmosBasedChainRequest", &axelarnettypes.AddCosmosBasedChainRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelarnet.v1beta1.RegisterAssetRequest", &axelarnettypes.RegisterAssetRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelarnet.v1beta1.RouteIBCTransfersRequest", &axelarnettypes.RouteIBCTransfersRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelarnet.v1beta1.RegisterFeeCollectorRequest", &axelarnettypes.RegisterFeeCollectorRequest{})

	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/evm.v1beta1.LinkRequest", &evmtypes.LinkRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/evm.v1beta1.ConfirmTokenRequest", &evmtypes.ConfirmTokenRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/evm.v1beta1.ConfirmDepositRequest", &evmtypes.ConfirmDepositRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/evm.v1beta1.ConfirmTransferKeyRequest", &evmtypes.ConfirmTransferKeyRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/evm.v1beta1.CreatePendingTransfersRequest", &evmtypes.CreatePendingTransfersRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/evm.v1beta1.CreateDeployTokenRequest", &evmtypes.CreateDeployTokenRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/evm.v1beta1.CreateBurnTokensRequest", &evmtypes.CreateBurnTokensRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/evm.v1beta1.CreateTransferOwnershipRequest", &evmtypes.CreateTransferOwnershipRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/evm.v1beta1.CreateTransferOperatorshipRequest", &evmtypes.CreateTransferOperatorshipRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/evm.v1beta1.SignCommandsRequest", &evmtypes.SignCommandsRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/evm.v1beta1.AddChainRequest", &evmtypes.AddChainRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/evm.v1beta1.SetGatewayRequest", &evmtypes.SetGatewayRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/evm.v1beta1.ConfirmGatewayTxRequest", &evmtypes.ConfirmGatewayTxRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/evm.v1beta1.RetryFailedEventRequest", &evmtypes.RetryFailedEventRequest{})

	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/nexus.v1beta1.RegisterChainMaintainerRequest", &nexustypes.RegisterChainMaintainerRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/nexus.v1beta1.DeregisterChainMaintainerRequest", &nexustypes.DeregisterChainMaintainerRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/nexus.v1beta1.ActivateChainRequest", &nexustypes.ActivateChainRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/nexus.v1beta1.DeactivateChainRequest", &nexustypes.DeactivateChainRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/nexus.v1beta1.RegisterAssetFeeRequest", &nexustypes.RegisterAssetFeeRequest{})

	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/permission.v1beta1.UpdateGovernanceKeyRequest", &permissiontypes.UpdateGovernanceKeyRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/permission.v1beta1.RegisterControllerRequest", &permissiontypes.RegisterControllerRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/permission.v1beta1.DeregisterControllerRequest", &permissiontypes.DeregisterControllerRequest{})

	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/reward.v1beta1.RefundMsgRequest", &rewardtypes.RefundMsgRequest{})

	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/snapshot.v1beta1.RegisterProxyRequest", &snapshottypes.RegisterProxyRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/snapshot.v1beta1.DeactivateProxyRequest", &snapshottypes.DeactivateProxyRequest{})

	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/tss.v1beta1.HeartBeatRequest", &tsstypes.HeartBeatRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/tss.v1beta1.StartKeygenRequest", &tsstypes.StartKeygenRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/tss.v1beta1.ProcessKeygenTrafficRequest", &tsstypes.ProcessKeygenTrafficRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/tss.v1beta1.ProcessSignTrafficRequest", &tsstypes.ProcessSignTrafficRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/tss.v1beta1.RotateKeyRequest", &tsstypes.RotateKeyRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/tss.v1beta1.VoteSigRequest", &tsstypes.VoteSigRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/tss.v1beta1.VotePubKeyRequest", &tsstypes.VotePubKeyRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/tss.v1beta1.RegisterExternalKeysRequest", &tsstypes.RegisterExternalKeysRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/tss.v1beta1.SubmitMultisigPubKeysRequest", &tsstypes.SubmitMultisigPubKeysRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/tss.v1beta1.SubmitMultisigSignaturesRequest", &tsstypes.SubmitMultisigSignaturesRequest{})

	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelar.tss.v1beta1.StartKeygenRequest", &tsstypes.StartKeygenRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelar.tss.v1beta1.ProcessKeygenTrafficRequest", &tsstypes.ProcessKeygenTrafficRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelar.tss.v1beta1.ProcessSignTrafficRequest", &tsstypes.ProcessSignTrafficRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelar.tss.v1beta1.RotateKeyRequest", &tsstypes.RotateKeyRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelar.tss.v1beta1.VoteSigRequest", &tsstypes.VoteSigRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelar.tss.v1beta1.VotePubKeyRequest", &tsstypes.VotePubKeyRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelar.tss.v1beta1.RegisterExternalKeysRequest", &tsstypes.RegisterExternalKeysRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelar.tss.v1beta1.SubmitMultisigPubKeysRequest", &tsstypes.SubmitMultisigPubKeysRequest{})
	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/axelar.tss.v1beta1.SubmitMultisigSignaturesRequest", &tsstypes.SubmitMultisigSignaturesRequest{})

	r.RegisterCustomTypeURL((*sdk.Msg)(nil), "/vote.v1beta1.VoteRequest", &votetypes.VoteRequest{})
}
