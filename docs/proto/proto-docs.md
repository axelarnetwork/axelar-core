<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [axelar/axelarnet/v1beta1/params.proto](#axelar/axelarnet/v1beta1/params.proto)
    - [Params](#axelar.axelarnet.v1beta1.Params)
  
- [axelar/axelarnet/v1beta1/types.proto](#axelar/axelarnet/v1beta1/types.proto)
    - [Asset](#axelar.axelarnet.v1beta1.Asset)
    - [CosmosChain](#axelar.axelarnet.v1beta1.CosmosChain)
    - [IBCTransfer](#axelar.axelarnet.v1beta1.IBCTransfer)
  
- [axelar/axelarnet/v1beta1/genesis.proto](#axelar/axelarnet/v1beta1/genesis.proto)
    - [GenesisState](#axelar.axelarnet.v1beta1.GenesisState)
  
- [axelar/utils/v1beta1/threshold.proto](#axelar/utils/v1beta1/threshold.proto)
    - [Threshold](#axelar.utils.v1beta1.Threshold)
  
- [axelar/tss/exported/v1beta1/types.proto](#axelar/tss/exported/v1beta1/types.proto)
    - [Key](#axelar.tss.exported.v1beta1.Key)
    - [Key.ECDSAKey](#axelar.tss.exported.v1beta1.Key.ECDSAKey)
    - [Key.MultisigKey](#axelar.tss.exported.v1beta1.Key.MultisigKey)
    - [KeyRequirement](#axelar.tss.exported.v1beta1.KeyRequirement)
    - [SigKeyPair](#axelar.tss.exported.v1beta1.SigKeyPair)
    - [SignInfo](#axelar.tss.exported.v1beta1.SignInfo)
    - [Signature](#axelar.tss.exported.v1beta1.Signature)
    - [Signature.MultiSig](#axelar.tss.exported.v1beta1.Signature.MultiSig)
    - [Signature.SingleSig](#axelar.tss.exported.v1beta1.Signature.SingleSig)
  
    - [AckType](#axelar.tss.exported.v1beta1.AckType)
    - [KeyRole](#axelar.tss.exported.v1beta1.KeyRole)
    - [KeyShareDistributionPolicy](#axelar.tss.exported.v1beta1.KeyShareDistributionPolicy)
    - [KeyType](#axelar.tss.exported.v1beta1.KeyType)
    - [SigStatus](#axelar.tss.exported.v1beta1.SigStatus)
  
- [axelar/nexus/exported/v1beta1/types.proto](#axelar/nexus/exported/v1beta1/types.proto)
    - [Asset](#axelar.nexus.exported.v1beta1.Asset)
    - [Chain](#axelar.nexus.exported.v1beta1.Chain)
    - [CrossChainAddress](#axelar.nexus.exported.v1beta1.CrossChainAddress)
    - [CrossChainTransfer](#axelar.nexus.exported.v1beta1.CrossChainTransfer)
    - [FeeInfo](#axelar.nexus.exported.v1beta1.FeeInfo)
    - [TransferFee](#axelar.nexus.exported.v1beta1.TransferFee)
  
    - [TransferState](#axelar.nexus.exported.v1beta1.TransferState)
  
- [axelar/utils/v1beta1/bitmap.proto](#axelar/utils/v1beta1/bitmap.proto)
    - [Bitmap](#axelar.utils.v1beta1.Bitmap)
    - [CircularBuffer](#axelar.utils.v1beta1.CircularBuffer)
  
- [axelar/nexus/v1beta1/types.proto](#axelar/nexus/v1beta1/types.proto)
    - [ChainState](#axelar.nexus.v1beta1.ChainState)
    - [LinkedAddresses](#axelar.nexus.v1beta1.LinkedAddresses)
    - [MaintainerState](#axelar.nexus.v1beta1.MaintainerState)
  
- [axelar/nexus/v1beta1/query.proto](#axelar/nexus/v1beta1/query.proto)
    - [AssetsRequest](#axelar.nexus.v1beta1.AssetsRequest)
    - [AssetsResponse](#axelar.nexus.v1beta1.AssetsResponse)
    - [ChainStateRequest](#axelar.nexus.v1beta1.ChainStateRequest)
    - [ChainStateResponse](#axelar.nexus.v1beta1.ChainStateResponse)
    - [ChainsByAssetRequest](#axelar.nexus.v1beta1.ChainsByAssetRequest)
    - [ChainsByAssetResponse](#axelar.nexus.v1beta1.ChainsByAssetResponse)
    - [ChainsRequest](#axelar.nexus.v1beta1.ChainsRequest)
    - [ChainsResponse](#axelar.nexus.v1beta1.ChainsResponse)
    - [FeeInfoRequest](#axelar.nexus.v1beta1.FeeInfoRequest)
    - [FeeInfoResponse](#axelar.nexus.v1beta1.FeeInfoResponse)
    - [LatestDepositAddressRequest](#axelar.nexus.v1beta1.LatestDepositAddressRequest)
    - [LatestDepositAddressResponse](#axelar.nexus.v1beta1.LatestDepositAddressResponse)
    - [QueryChainMaintainersResponse](#axelar.nexus.v1beta1.QueryChainMaintainersResponse)
    - [TransferFeeRequest](#axelar.nexus.v1beta1.TransferFeeRequest)
    - [TransferFeeResponse](#axelar.nexus.v1beta1.TransferFeeResponse)
    - [TransfersForChainRequest](#axelar.nexus.v1beta1.TransfersForChainRequest)
    - [TransfersForChainResponse](#axelar.nexus.v1beta1.TransfersForChainResponse)
  
- [axelar/axelarnet/v1beta1/query.proto](#axelar/axelarnet/v1beta1/query.proto)
    - [PendingIBCTransferCountRequest](#axelar.axelarnet.v1beta1.PendingIBCTransferCountRequest)
    - [PendingIBCTransferCountResponse](#axelar.axelarnet.v1beta1.PendingIBCTransferCountResponse)
    - [PendingIBCTransferCountResponse.TransfersByChainEntry](#axelar.axelarnet.v1beta1.PendingIBCTransferCountResponse.TransfersByChainEntry)
  
- [axelar/permission/exported/v1beta1/types.proto](#axelar/permission/exported/v1beta1/types.proto)
    - [Role](#axelar.permission.exported.v1beta1.Role)
  
    - [File-level Extensions](#axelar/permission/exported/v1beta1/types.proto-extensions)
  
- [axelar/axelarnet/v1beta1/tx.proto](#axelar/axelarnet/v1beta1/tx.proto)
    - [AddCosmosBasedChainRequest](#axelar.axelarnet.v1beta1.AddCosmosBasedChainRequest)
    - [AddCosmosBasedChainResponse](#axelar.axelarnet.v1beta1.AddCosmosBasedChainResponse)
    - [ConfirmDepositRequest](#axelar.axelarnet.v1beta1.ConfirmDepositRequest)
    - [ConfirmDepositResponse](#axelar.axelarnet.v1beta1.ConfirmDepositResponse)
    - [ExecutePendingTransfersRequest](#axelar.axelarnet.v1beta1.ExecutePendingTransfersRequest)
    - [ExecutePendingTransfersResponse](#axelar.axelarnet.v1beta1.ExecutePendingTransfersResponse)
    - [LinkRequest](#axelar.axelarnet.v1beta1.LinkRequest)
    - [LinkResponse](#axelar.axelarnet.v1beta1.LinkResponse)
    - [RegisterAssetRequest](#axelar.axelarnet.v1beta1.RegisterAssetRequest)
    - [RegisterAssetResponse](#axelar.axelarnet.v1beta1.RegisterAssetResponse)
    - [RegisterFeeCollectorRequest](#axelar.axelarnet.v1beta1.RegisterFeeCollectorRequest)
    - [RegisterFeeCollectorResponse](#axelar.axelarnet.v1beta1.RegisterFeeCollectorResponse)
    - [RegisterIBCPathRequest](#axelar.axelarnet.v1beta1.RegisterIBCPathRequest)
    - [RegisterIBCPathResponse](#axelar.axelarnet.v1beta1.RegisterIBCPathResponse)
    - [RouteIBCTransfersRequest](#axelar.axelarnet.v1beta1.RouteIBCTransfersRequest)
    - [RouteIBCTransfersResponse](#axelar.axelarnet.v1beta1.RouteIBCTransfersResponse)
  
- [axelar/axelarnet/v1beta1/service.proto](#axelar/axelarnet/v1beta1/service.proto)
    - [MsgService](#axelar.axelarnet.v1beta1.MsgService)
    - [QueryService](#axelar.axelarnet.v1beta1.QueryService)
  
- [axelar/bitcoin/v1beta1/types.proto](#axelar/bitcoin/v1beta1/types.proto)
    - [AddressInfo](#axelar.bitcoin.v1beta1.AddressInfo)
    - [AddressInfo.SpendingCondition](#axelar.bitcoin.v1beta1.AddressInfo.SpendingCondition)
    - [Network](#axelar.bitcoin.v1beta1.Network)
    - [OutPointInfo](#axelar.bitcoin.v1beta1.OutPointInfo)
    - [SignedTx](#axelar.bitcoin.v1beta1.SignedTx)
    - [UnsignedTx](#axelar.bitcoin.v1beta1.UnsignedTx)
    - [UnsignedTx.Info](#axelar.bitcoin.v1beta1.UnsignedTx.Info)
    - [UnsignedTx.Info.InputInfo](#axelar.bitcoin.v1beta1.UnsignedTx.Info.InputInfo)
    - [UnsignedTx.Info.InputInfo.SigRequirement](#axelar.bitcoin.v1beta1.UnsignedTx.Info.InputInfo.SigRequirement)
  
    - [AddressRole](#axelar.bitcoin.v1beta1.AddressRole)
    - [OutPointState](#axelar.bitcoin.v1beta1.OutPointState)
    - [TxStatus](#axelar.bitcoin.v1beta1.TxStatus)
    - [TxType](#axelar.bitcoin.v1beta1.TxType)
  
- [axelar/bitcoin/v1beta1/params.proto](#axelar/bitcoin/v1beta1/params.proto)
    - [Params](#axelar.bitcoin.v1beta1.Params)
  
- [axelar/bitcoin/v1beta1/genesis.proto](#axelar/bitcoin/v1beta1/genesis.proto)
    - [GenesisState](#axelar.bitcoin.v1beta1.GenesisState)
  
- [axelar/bitcoin/v1beta1/query.proto](#axelar/bitcoin/v1beta1/query.proto)
    - [DepositQueryParams](#axelar.bitcoin.v1beta1.DepositQueryParams)
    - [QueryAddressResponse](#axelar.bitcoin.v1beta1.QueryAddressResponse)
    - [QueryDepositStatusResponse](#axelar.bitcoin.v1beta1.QueryDepositStatusResponse)
    - [QueryTxResponse](#axelar.bitcoin.v1beta1.QueryTxResponse)
    - [QueryTxResponse.SigningInfo](#axelar.bitcoin.v1beta1.QueryTxResponse.SigningInfo)
  
- [axelar/snapshot/exported/v1beta1/types.proto](#axelar/snapshot/exported/v1beta1/types.proto)
    - [Snapshot](#axelar.snapshot.exported.v1beta1.Snapshot)
    - [Validator](#axelar.snapshot.exported.v1beta1.Validator)
  
    - [ValidatorIllegibility](#axelar.snapshot.exported.v1beta1.ValidatorIllegibility)
  
- [axelar/vote/exported/v1beta1/types.proto](#axelar/vote/exported/v1beta1/types.proto)
    - [PollKey](#axelar.vote.exported.v1beta1.PollKey)
    - [PollMetadata](#axelar.vote.exported.v1beta1.PollMetadata)
    - [Vote](#axelar.vote.exported.v1beta1.Vote)
    - [Voter](#axelar.vote.exported.v1beta1.Voter)
  
    - [PollState](#axelar.vote.exported.v1beta1.PollState)
  
- [axelar/bitcoin/v1beta1/tx.proto](#axelar/bitcoin/v1beta1/tx.proto)
    - [ConfirmOutpointRequest](#axelar.bitcoin.v1beta1.ConfirmOutpointRequest)
    - [ConfirmOutpointResponse](#axelar.bitcoin.v1beta1.ConfirmOutpointResponse)
    - [CreateMasterTxRequest](#axelar.bitcoin.v1beta1.CreateMasterTxRequest)
    - [CreateMasterTxResponse](#axelar.bitcoin.v1beta1.CreateMasterTxResponse)
    - [CreatePendingTransfersTxRequest](#axelar.bitcoin.v1beta1.CreatePendingTransfersTxRequest)
    - [CreatePendingTransfersTxResponse](#axelar.bitcoin.v1beta1.CreatePendingTransfersTxResponse)
    - [CreateRescueTxRequest](#axelar.bitcoin.v1beta1.CreateRescueTxRequest)
    - [CreateRescueTxResponse](#axelar.bitcoin.v1beta1.CreateRescueTxResponse)
    - [LinkRequest](#axelar.bitcoin.v1beta1.LinkRequest)
    - [LinkResponse](#axelar.bitcoin.v1beta1.LinkResponse)
    - [SignTxRequest](#axelar.bitcoin.v1beta1.SignTxRequest)
    - [SignTxResponse](#axelar.bitcoin.v1beta1.SignTxResponse)
    - [SubmitExternalSignatureRequest](#axelar.bitcoin.v1beta1.SubmitExternalSignatureRequest)
    - [SubmitExternalSignatureResponse](#axelar.bitcoin.v1beta1.SubmitExternalSignatureResponse)
    - [VoteConfirmOutpointRequest](#axelar.bitcoin.v1beta1.VoteConfirmOutpointRequest)
    - [VoteConfirmOutpointResponse](#axelar.bitcoin.v1beta1.VoteConfirmOutpointResponse)
  
- [axelar/bitcoin/v1beta1/service.proto](#axelar/bitcoin/v1beta1/service.proto)
    - [MsgService](#axelar.bitcoin.v1beta1.MsgService)
  
- [axelar/utils/v1beta1/queuer.proto](#axelar/utils/v1beta1/queuer.proto)
    - [QueueState](#axelar.utils.v1beta1.QueueState)
    - [QueueState.Item](#axelar.utils.v1beta1.QueueState.Item)
    - [QueueState.ItemsEntry](#axelar.utils.v1beta1.QueueState.ItemsEntry)
  
- [axelar/evm/v1beta1/types.proto](#axelar/evm/v1beta1/types.proto)
    - [Asset](#axelar.evm.v1beta1.Asset)
    - [BurnerInfo](#axelar.evm.v1beta1.BurnerInfo)
    - [Command](#axelar.evm.v1beta1.Command)
    - [CommandBatchMetadata](#axelar.evm.v1beta1.CommandBatchMetadata)
    - [ERC20Deposit](#axelar.evm.v1beta1.ERC20Deposit)
    - [ERC20TokenMetadata](#axelar.evm.v1beta1.ERC20TokenMetadata)
    - [Event](#axelar.evm.v1beta1.Event)
    - [EventContractCall](#axelar.evm.v1beta1.EventContractCall)
    - [EventContractCallWithToken](#axelar.evm.v1beta1.EventContractCallWithToken)
    - [EventMultisigOperatorshipTransferred](#axelar.evm.v1beta1.EventMultisigOperatorshipTransferred)
    - [EventMultisigOwnershipTransferred](#axelar.evm.v1beta1.EventMultisigOwnershipTransferred)
    - [EventSinglesigOperatorshipTransferred](#axelar.evm.v1beta1.EventSinglesigOperatorshipTransferred)
    - [EventSinglesigOwnershipTransferred](#axelar.evm.v1beta1.EventSinglesigOwnershipTransferred)
    - [EventTokenDeployed](#axelar.evm.v1beta1.EventTokenDeployed)
    - [EventTokenSent](#axelar.evm.v1beta1.EventTokenSent)
    - [EventTransfer](#axelar.evm.v1beta1.EventTransfer)
    - [Gateway](#axelar.evm.v1beta1.Gateway)
    - [NetworkInfo](#axelar.evm.v1beta1.NetworkInfo)
    - [SigMetadata](#axelar.evm.v1beta1.SigMetadata)
    - [TokenDetails](#axelar.evm.v1beta1.TokenDetails)
    - [TransactionMetadata](#axelar.evm.v1beta1.TransactionMetadata)
    - [TransferKey](#axelar.evm.v1beta1.TransferKey)
    - [VoteEvents](#axelar.evm.v1beta1.VoteEvents)
  
    - [BatchedCommandsStatus](#axelar.evm.v1beta1.BatchedCommandsStatus)
    - [DepositStatus](#axelar.evm.v1beta1.DepositStatus)
    - [Event.Status](#axelar.evm.v1beta1.Event.Status)
    - [Gateway.Status](#axelar.evm.v1beta1.Gateway.Status)
    - [SigType](#axelar.evm.v1beta1.SigType)
    - [Status](#axelar.evm.v1beta1.Status)
    - [TransferKeyType](#axelar.evm.v1beta1.TransferKeyType)
  
- [axelar/evm/v1beta1/params.proto](#axelar/evm/v1beta1/params.proto)
    - [Params](#axelar.evm.v1beta1.Params)
    - [PendingChain](#axelar.evm.v1beta1.PendingChain)
  
- [axelar/evm/v1beta1/genesis.proto](#axelar/evm/v1beta1/genesis.proto)
    - [GenesisState](#axelar.evm.v1beta1.GenesisState)
    - [GenesisState.Chain](#axelar.evm.v1beta1.GenesisState.Chain)
  
- [axelar/evm/v1beta1/query.proto](#axelar/evm/v1beta1/query.proto)
    - [BatchedCommandsRequest](#axelar.evm.v1beta1.BatchedCommandsRequest)
    - [BatchedCommandsResponse](#axelar.evm.v1beta1.BatchedCommandsResponse)
    - [BurnerInfoRequest](#axelar.evm.v1beta1.BurnerInfoRequest)
    - [BurnerInfoResponse](#axelar.evm.v1beta1.BurnerInfoResponse)
    - [BytecodeRequest](#axelar.evm.v1beta1.BytecodeRequest)
    - [BytecodeResponse](#axelar.evm.v1beta1.BytecodeResponse)
    - [ChainsRequest](#axelar.evm.v1beta1.ChainsRequest)
    - [ChainsResponse](#axelar.evm.v1beta1.ChainsResponse)
    - [ConfirmationHeightRequest](#axelar.evm.v1beta1.ConfirmationHeightRequest)
    - [ConfirmationHeightResponse](#axelar.evm.v1beta1.ConfirmationHeightResponse)
    - [DepositQueryParams](#axelar.evm.v1beta1.DepositQueryParams)
    - [DepositStateRequest](#axelar.evm.v1beta1.DepositStateRequest)
    - [DepositStateResponse](#axelar.evm.v1beta1.DepositStateResponse)
    - [EventRequest](#axelar.evm.v1beta1.EventRequest)
    - [EventResponse](#axelar.evm.v1beta1.EventResponse)
    - [GatewayAddressRequest](#axelar.evm.v1beta1.GatewayAddressRequest)
    - [GatewayAddressResponse](#axelar.evm.v1beta1.GatewayAddressResponse)
    - [KeyAddressRequest](#axelar.evm.v1beta1.KeyAddressRequest)
    - [KeyAddressResponse](#axelar.evm.v1beta1.KeyAddressResponse)
    - [KeyAddressResponse.MultisigAddresses](#axelar.evm.v1beta1.KeyAddressResponse.MultisigAddresses)
    - [KeyAddressResponse.ThresholdAddress](#axelar.evm.v1beta1.KeyAddressResponse.ThresholdAddress)
    - [PendingCommandsRequest](#axelar.evm.v1beta1.PendingCommandsRequest)
    - [PendingCommandsResponse](#axelar.evm.v1beta1.PendingCommandsResponse)
    - [QueryBurnerAddressResponse](#axelar.evm.v1beta1.QueryBurnerAddressResponse)
    - [QueryCommandResponse](#axelar.evm.v1beta1.QueryCommandResponse)
    - [QueryCommandResponse.ParamsEntry](#axelar.evm.v1beta1.QueryCommandResponse.ParamsEntry)
    - [QueryDepositStateParams](#axelar.evm.v1beta1.QueryDepositStateParams)
    - [QueryTokenAddressResponse](#axelar.evm.v1beta1.QueryTokenAddressResponse)
  
- [axelar/evm/v1beta1/tx.proto](#axelar/evm/v1beta1/tx.proto)
    - [AddChainRequest](#axelar.evm.v1beta1.AddChainRequest)
    - [AddChainResponse](#axelar.evm.v1beta1.AddChainResponse)
    - [ConfirmDepositRequest](#axelar.evm.v1beta1.ConfirmDepositRequest)
    - [ConfirmDepositResponse](#axelar.evm.v1beta1.ConfirmDepositResponse)
    - [ConfirmGatewayTxRequest](#axelar.evm.v1beta1.ConfirmGatewayTxRequest)
    - [ConfirmGatewayTxResponse](#axelar.evm.v1beta1.ConfirmGatewayTxResponse)
    - [ConfirmTokenRequest](#axelar.evm.v1beta1.ConfirmTokenRequest)
    - [ConfirmTokenResponse](#axelar.evm.v1beta1.ConfirmTokenResponse)
    - [ConfirmTransferKeyRequest](#axelar.evm.v1beta1.ConfirmTransferKeyRequest)
    - [ConfirmTransferKeyResponse](#axelar.evm.v1beta1.ConfirmTransferKeyResponse)
    - [CreateBurnTokensRequest](#axelar.evm.v1beta1.CreateBurnTokensRequest)
    - [CreateBurnTokensResponse](#axelar.evm.v1beta1.CreateBurnTokensResponse)
    - [CreateDeployTokenRequest](#axelar.evm.v1beta1.CreateDeployTokenRequest)
    - [CreateDeployTokenResponse](#axelar.evm.v1beta1.CreateDeployTokenResponse)
    - [CreatePendingTransfersRequest](#axelar.evm.v1beta1.CreatePendingTransfersRequest)
    - [CreatePendingTransfersResponse](#axelar.evm.v1beta1.CreatePendingTransfersResponse)
    - [CreateTransferOperatorshipRequest](#axelar.evm.v1beta1.CreateTransferOperatorshipRequest)
    - [CreateTransferOperatorshipResponse](#axelar.evm.v1beta1.CreateTransferOperatorshipResponse)
    - [CreateTransferOwnershipRequest](#axelar.evm.v1beta1.CreateTransferOwnershipRequest)
    - [CreateTransferOwnershipResponse](#axelar.evm.v1beta1.CreateTransferOwnershipResponse)
    - [LinkRequest](#axelar.evm.v1beta1.LinkRequest)
    - [LinkResponse](#axelar.evm.v1beta1.LinkResponse)
    - [RetryFailedEventRequest](#axelar.evm.v1beta1.RetryFailedEventRequest)
    - [RetryFailedEventResponse](#axelar.evm.v1beta1.RetryFailedEventResponse)
    - [SetGatewayRequest](#axelar.evm.v1beta1.SetGatewayRequest)
    - [SetGatewayResponse](#axelar.evm.v1beta1.SetGatewayResponse)
    - [SignCommandsRequest](#axelar.evm.v1beta1.SignCommandsRequest)
    - [SignCommandsResponse](#axelar.evm.v1beta1.SignCommandsResponse)
  
- [axelar/evm/v1beta1/service.proto](#axelar/evm/v1beta1/service.proto)
    - [MsgService](#axelar.evm.v1beta1.MsgService)
    - [QueryService](#axelar.evm.v1beta1.QueryService)
  
- [axelar/nexus/v1beta1/params.proto](#axelar/nexus/v1beta1/params.proto)
    - [Params](#axelar.nexus.v1beta1.Params)
  
- [axelar/nexus/v1beta1/genesis.proto](#axelar/nexus/v1beta1/genesis.proto)
    - [GenesisState](#axelar.nexus.v1beta1.GenesisState)
  
- [axelar/nexus/v1beta1/tx.proto](#axelar/nexus/v1beta1/tx.proto)
    - [ActivateChainRequest](#axelar.nexus.v1beta1.ActivateChainRequest)
    - [ActivateChainResponse](#axelar.nexus.v1beta1.ActivateChainResponse)
    - [DeactivateChainRequest](#axelar.nexus.v1beta1.DeactivateChainRequest)
    - [DeactivateChainResponse](#axelar.nexus.v1beta1.DeactivateChainResponse)
    - [DeregisterChainMaintainerRequest](#axelar.nexus.v1beta1.DeregisterChainMaintainerRequest)
    - [DeregisterChainMaintainerResponse](#axelar.nexus.v1beta1.DeregisterChainMaintainerResponse)
    - [RegisterAssetFeeRequest](#axelar.nexus.v1beta1.RegisterAssetFeeRequest)
    - [RegisterAssetFeeResponse](#axelar.nexus.v1beta1.RegisterAssetFeeResponse)
    - [RegisterChainMaintainerRequest](#axelar.nexus.v1beta1.RegisterChainMaintainerRequest)
    - [RegisterChainMaintainerResponse](#axelar.nexus.v1beta1.RegisterChainMaintainerResponse)
  
- [axelar/nexus/v1beta1/service.proto](#axelar/nexus/v1beta1/service.proto)
    - [MsgService](#axelar.nexus.v1beta1.MsgService)
    - [QueryService](#axelar.nexus.v1beta1.QueryService)
  
- [axelar/permission/v1beta1/types.proto](#axelar/permission/v1beta1/types.proto)
    - [GovAccount](#axelar.permission.v1beta1.GovAccount)
  
- [axelar/permission/v1beta1/params.proto](#axelar/permission/v1beta1/params.proto)
    - [Params](#axelar.permission.v1beta1.Params)
  
- [axelar/permission/v1beta1/genesis.proto](#axelar/permission/v1beta1/genesis.proto)
    - [GenesisState](#axelar.permission.v1beta1.GenesisState)
  
- [axelar/permission/v1beta1/query.proto](#axelar/permission/v1beta1/query.proto)
    - [QueryGovernanceKeyRequest](#axelar.permission.v1beta1.QueryGovernanceKeyRequest)
    - [QueryGovernanceKeyResponse](#axelar.permission.v1beta1.QueryGovernanceKeyResponse)
  
- [axelar/permission/v1beta1/tx.proto](#axelar/permission/v1beta1/tx.proto)
    - [DeregisterControllerRequest](#axelar.permission.v1beta1.DeregisterControllerRequest)
    - [DeregisterControllerResponse](#axelar.permission.v1beta1.DeregisterControllerResponse)
    - [RegisterControllerRequest](#axelar.permission.v1beta1.RegisterControllerRequest)
    - [RegisterControllerResponse](#axelar.permission.v1beta1.RegisterControllerResponse)
    - [UpdateGovernanceKeyRequest](#axelar.permission.v1beta1.UpdateGovernanceKeyRequest)
    - [UpdateGovernanceKeyResponse](#axelar.permission.v1beta1.UpdateGovernanceKeyResponse)
  
- [axelar/permission/v1beta1/service.proto](#axelar/permission/v1beta1/service.proto)
    - [Msg](#axelar.permission.v1beta1.Msg)
    - [Query](#axelar.permission.v1beta1.Query)
  
- [axelar/reward/v1beta1/params.proto](#axelar/reward/v1beta1/params.proto)
    - [Params](#axelar.reward.v1beta1.Params)
  
- [axelar/reward/v1beta1/types.proto](#axelar/reward/v1beta1/types.proto)
    - [Pool](#axelar.reward.v1beta1.Pool)
    - [Pool.Reward](#axelar.reward.v1beta1.Pool.Reward)
    - [Refund](#axelar.reward.v1beta1.Refund)
  
- [axelar/reward/v1beta1/genesis.proto](#axelar/reward/v1beta1/genesis.proto)
    - [GenesisState](#axelar.reward.v1beta1.GenesisState)
  
- [axelar/reward/v1beta1/tx.proto](#axelar/reward/v1beta1/tx.proto)
    - [RefundMsgRequest](#axelar.reward.v1beta1.RefundMsgRequest)
    - [RefundMsgResponse](#axelar.reward.v1beta1.RefundMsgResponse)
  
- [axelar/reward/v1beta1/service.proto](#axelar/reward/v1beta1/service.proto)
    - [MsgService](#axelar.reward.v1beta1.MsgService)
  
- [axelar/snapshot/v1beta1/params.proto](#axelar/snapshot/v1beta1/params.proto)
    - [Params](#axelar.snapshot.v1beta1.Params)
  
- [axelar/snapshot/v1beta1/types.proto](#axelar/snapshot/v1beta1/types.proto)
    - [ProxiedValidator](#axelar.snapshot.v1beta1.ProxiedValidator)
  
- [axelar/snapshot/v1beta1/genesis.proto](#axelar/snapshot/v1beta1/genesis.proto)
    - [GenesisState](#axelar.snapshot.v1beta1.GenesisState)
  
- [axelar/snapshot/v1beta1/query.proto](#axelar/snapshot/v1beta1/query.proto)
    - [QueryValidatorsResponse](#axelar.snapshot.v1beta1.QueryValidatorsResponse)
    - [QueryValidatorsResponse.TssIllegibilityInfo](#axelar.snapshot.v1beta1.QueryValidatorsResponse.TssIllegibilityInfo)
    - [QueryValidatorsResponse.Validator](#axelar.snapshot.v1beta1.QueryValidatorsResponse.Validator)
  
- [axelar/snapshot/v1beta1/tx.proto](#axelar/snapshot/v1beta1/tx.proto)
    - [DeactivateProxyRequest](#axelar.snapshot.v1beta1.DeactivateProxyRequest)
    - [DeactivateProxyResponse](#axelar.snapshot.v1beta1.DeactivateProxyResponse)
    - [RegisterProxyRequest](#axelar.snapshot.v1beta1.RegisterProxyRequest)
    - [RegisterProxyResponse](#axelar.snapshot.v1beta1.RegisterProxyResponse)
  
- [axelar/snapshot/v1beta1/service.proto](#axelar/snapshot/v1beta1/service.proto)
    - [MsgService](#axelar.snapshot.v1beta1.MsgService)
  
- [axelar/tss/tofnd/v1beta1/common.proto](#axelar/tss/tofnd/v1beta1/common.proto)
    - [KeyPresenceRequest](#axelar.tss.tofnd.v1beta1.KeyPresenceRequest)
    - [KeyPresenceResponse](#axelar.tss.tofnd.v1beta1.KeyPresenceResponse)
  
    - [KeyPresenceResponse.Response](#axelar.tss.tofnd.v1beta1.KeyPresenceResponse.Response)
  
- [axelar/tss/tofnd/v1beta1/multisig.proto](#axelar/tss/tofnd/v1beta1/multisig.proto)
    - [KeygenRequest](#axelar.tss.tofnd.v1beta1.KeygenRequest)
    - [KeygenResponse](#axelar.tss.tofnd.v1beta1.KeygenResponse)
    - [SignRequest](#axelar.tss.tofnd.v1beta1.SignRequest)
    - [SignResponse](#axelar.tss.tofnd.v1beta1.SignResponse)
  
- [axelar/tss/tofnd/v1beta1/tofnd.proto](#axelar/tss/tofnd/v1beta1/tofnd.proto)
    - [KeygenInit](#axelar.tss.tofnd.v1beta1.KeygenInit)
    - [KeygenOutput](#axelar.tss.tofnd.v1beta1.KeygenOutput)
    - [MessageIn](#axelar.tss.tofnd.v1beta1.MessageIn)
    - [MessageOut](#axelar.tss.tofnd.v1beta1.MessageOut)
    - [MessageOut.CriminalList](#axelar.tss.tofnd.v1beta1.MessageOut.CriminalList)
    - [MessageOut.CriminalList.Criminal](#axelar.tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal)
    - [MessageOut.KeygenResult](#axelar.tss.tofnd.v1beta1.MessageOut.KeygenResult)
    - [MessageOut.SignResult](#axelar.tss.tofnd.v1beta1.MessageOut.SignResult)
    - [RecoverRequest](#axelar.tss.tofnd.v1beta1.RecoverRequest)
    - [RecoverResponse](#axelar.tss.tofnd.v1beta1.RecoverResponse)
    - [SignInit](#axelar.tss.tofnd.v1beta1.SignInit)
    - [TrafficIn](#axelar.tss.tofnd.v1beta1.TrafficIn)
    - [TrafficOut](#axelar.tss.tofnd.v1beta1.TrafficOut)
  
    - [MessageOut.CriminalList.Criminal.CrimeType](#axelar.tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal.CrimeType)
    - [RecoverResponse.Response](#axelar.tss.tofnd.v1beta1.RecoverResponse.Response)
  
- [axelar/tss/v1beta1/params.proto](#axelar/tss/v1beta1/params.proto)
    - [Params](#axelar.tss.v1beta1.Params)
  
- [axelar/tss/v1beta1/types.proto](#axelar/tss/v1beta1/types.proto)
    - [ExternalKeys](#axelar.tss.v1beta1.ExternalKeys)
    - [KeyInfo](#axelar.tss.v1beta1.KeyInfo)
    - [KeyRecoveryInfo](#axelar.tss.v1beta1.KeyRecoveryInfo)
    - [KeyRecoveryInfo.PrivateEntry](#axelar.tss.v1beta1.KeyRecoveryInfo.PrivateEntry)
    - [KeygenVoteData](#axelar.tss.v1beta1.KeygenVoteData)
    - [MultisigInfo](#axelar.tss.v1beta1.MultisigInfo)
    - [MultisigInfo.Info](#axelar.tss.v1beta1.MultisigInfo.Info)
    - [ValidatorStatus](#axelar.tss.v1beta1.ValidatorStatus)
  
- [axelar/tss/v1beta1/genesis.proto](#axelar/tss/v1beta1/genesis.proto)
    - [GenesisState](#axelar.tss.v1beta1.GenesisState)
  
- [axelar/tss/v1beta1/query.proto](#axelar/tss/v1beta1/query.proto)
    - [AssignableKeyRequest](#axelar.tss.v1beta1.AssignableKeyRequest)
    - [AssignableKeyResponse](#axelar.tss.v1beta1.AssignableKeyResponse)
    - [NextKeyIDRequest](#axelar.tss.v1beta1.NextKeyIDRequest)
    - [NextKeyIDResponse](#axelar.tss.v1beta1.NextKeyIDResponse)
    - [QueryActiveOldKeysResponse](#axelar.tss.v1beta1.QueryActiveOldKeysResponse)
    - [QueryActiveOldKeysValidatorResponse](#axelar.tss.v1beta1.QueryActiveOldKeysValidatorResponse)
    - [QueryActiveOldKeysValidatorResponse.KeyInfo](#axelar.tss.v1beta1.QueryActiveOldKeysValidatorResponse.KeyInfo)
    - [QueryDeactivatedOperatorsResponse](#axelar.tss.v1beta1.QueryDeactivatedOperatorsResponse)
    - [QueryExternalKeyIDResponse](#axelar.tss.v1beta1.QueryExternalKeyIDResponse)
    - [QueryKeyResponse](#axelar.tss.v1beta1.QueryKeyResponse)
    - [QueryKeyResponse.ECDSAKey](#axelar.tss.v1beta1.QueryKeyResponse.ECDSAKey)
    - [QueryKeyResponse.Key](#axelar.tss.v1beta1.QueryKeyResponse.Key)
    - [QueryKeyResponse.MultisigKey](#axelar.tss.v1beta1.QueryKeyResponse.MultisigKey)
    - [QueryKeyShareResponse](#axelar.tss.v1beta1.QueryKeyShareResponse)
    - [QueryKeyShareResponse.ShareInfo](#axelar.tss.v1beta1.QueryKeyShareResponse.ShareInfo)
    - [QueryRecoveryResponse](#axelar.tss.v1beta1.QueryRecoveryResponse)
    - [QuerySignatureResponse](#axelar.tss.v1beta1.QuerySignatureResponse)
    - [QuerySignatureResponse.MultisigSignature](#axelar.tss.v1beta1.QuerySignatureResponse.MultisigSignature)
    - [QuerySignatureResponse.Signature](#axelar.tss.v1beta1.QuerySignatureResponse.Signature)
    - [QuerySignatureResponse.ThresholdSignature](#axelar.tss.v1beta1.QuerySignatureResponse.ThresholdSignature)
    - [ValidatorMultisigKeysRequest](#axelar.tss.v1beta1.ValidatorMultisigKeysRequest)
    - [ValidatorMultisigKeysResponse](#axelar.tss.v1beta1.ValidatorMultisigKeysResponse)
    - [ValidatorMultisigKeysResponse.Keys](#axelar.tss.v1beta1.ValidatorMultisigKeysResponse.Keys)
    - [ValidatorMultisigKeysResponse.KeysEntry](#axelar.tss.v1beta1.ValidatorMultisigKeysResponse.KeysEntry)
  
    - [VoteStatus](#axelar.tss.v1beta1.VoteStatus)
  
- [axelar/tss/v1beta1/tx.proto](#axelar/tss/v1beta1/tx.proto)
    - [HeartBeatRequest](#axelar.tss.v1beta1.HeartBeatRequest)
    - [HeartBeatResponse](#axelar.tss.v1beta1.HeartBeatResponse)
    - [ProcessKeygenTrafficRequest](#axelar.tss.v1beta1.ProcessKeygenTrafficRequest)
    - [ProcessKeygenTrafficResponse](#axelar.tss.v1beta1.ProcessKeygenTrafficResponse)
    - [ProcessSignTrafficRequest](#axelar.tss.v1beta1.ProcessSignTrafficRequest)
    - [ProcessSignTrafficResponse](#axelar.tss.v1beta1.ProcessSignTrafficResponse)
    - [RegisterExternalKeysRequest](#axelar.tss.v1beta1.RegisterExternalKeysRequest)
    - [RegisterExternalKeysRequest.ExternalKey](#axelar.tss.v1beta1.RegisterExternalKeysRequest.ExternalKey)
    - [RegisterExternalKeysResponse](#axelar.tss.v1beta1.RegisterExternalKeysResponse)
    - [RotateKeyRequest](#axelar.tss.v1beta1.RotateKeyRequest)
    - [RotateKeyResponse](#axelar.tss.v1beta1.RotateKeyResponse)
    - [StartKeygenRequest](#axelar.tss.v1beta1.StartKeygenRequest)
    - [StartKeygenResponse](#axelar.tss.v1beta1.StartKeygenResponse)
    - [SubmitMultisigPubKeysRequest](#axelar.tss.v1beta1.SubmitMultisigPubKeysRequest)
    - [SubmitMultisigPubKeysResponse](#axelar.tss.v1beta1.SubmitMultisigPubKeysResponse)
    - [SubmitMultisigSignaturesRequest](#axelar.tss.v1beta1.SubmitMultisigSignaturesRequest)
    - [SubmitMultisigSignaturesResponse](#axelar.tss.v1beta1.SubmitMultisigSignaturesResponse)
    - [VotePubKeyRequest](#axelar.tss.v1beta1.VotePubKeyRequest)
    - [VotePubKeyResponse](#axelar.tss.v1beta1.VotePubKeyResponse)
    - [VoteSigRequest](#axelar.tss.v1beta1.VoteSigRequest)
    - [VoteSigResponse](#axelar.tss.v1beta1.VoteSigResponse)
  
- [axelar/tss/v1beta1/service.proto](#axelar/tss/v1beta1/service.proto)
    - [MsgService](#axelar.tss.v1beta1.MsgService)
    - [QueryService](#axelar.tss.v1beta1.QueryService)
  
- [axelar/vote/v1beta1/params.proto](#axelar/vote/v1beta1/params.proto)
    - [Params](#axelar.vote.v1beta1.Params)
  
- [axelar/vote/v1beta1/genesis.proto](#axelar/vote/v1beta1/genesis.proto)
    - [GenesisState](#axelar.vote.v1beta1.GenesisState)
  
- [axelar/vote/v1beta1/tx.proto](#axelar/vote/v1beta1/tx.proto)
    - [VoteRequest](#axelar.vote.v1beta1.VoteRequest)
    - [VoteResponse](#axelar.vote.v1beta1.VoteResponse)
  
- [axelar/vote/v1beta1/service.proto](#axelar/vote/v1beta1/service.proto)
    - [MsgService](#axelar.vote.v1beta1.MsgService)
  
- [axelar/vote/v1beta1/types.proto](#axelar/vote/v1beta1/types.proto)
    - [TalliedVote](#axelar.vote.v1beta1.TalliedVote)
  
- [utils/v1beta1/queuer.proto](#utils/v1beta1/queuer.proto)
    - [QueueState](#utils.v1beta1.QueueState)
    - [QueueState.Item](#utils.v1beta1.QueueState.Item)
    - [QueueState.ItemsEntry](#utils.v1beta1.QueueState.ItemsEntry)
  
- [utils/v1beta1/threshold.proto](#utils/v1beta1/threshold.proto)
    - [Threshold](#utils.v1beta1.Threshold)
  
- [evm/v1beta1/types.proto](#evm/v1beta1/types.proto)
    - [Asset](#evm.v1beta1.Asset)
    - [BurnerInfo](#evm.v1beta1.BurnerInfo)
    - [Command](#evm.v1beta1.Command)
    - [CommandBatchMetadata](#evm.v1beta1.CommandBatchMetadata)
    - [ERC20Deposit](#evm.v1beta1.ERC20Deposit)
    - [ERC20TokenMetadata](#evm.v1beta1.ERC20TokenMetadata)
    - [Event](#evm.v1beta1.Event)
    - [EventContractCall](#evm.v1beta1.EventContractCall)
    - [EventContractCallWithToken](#evm.v1beta1.EventContractCallWithToken)
    - [EventMultisigOperatorshipTransferred](#evm.v1beta1.EventMultisigOperatorshipTransferred)
    - [EventMultisigOwnershipTransferred](#evm.v1beta1.EventMultisigOwnershipTransferred)
    - [EventSinglesigOperatorshipTransferred](#evm.v1beta1.EventSinglesigOperatorshipTransferred)
    - [EventSinglesigOwnershipTransferred](#evm.v1beta1.EventSinglesigOwnershipTransferred)
    - [EventTokenDeployed](#evm.v1beta1.EventTokenDeployed)
    - [EventTokenSent](#evm.v1beta1.EventTokenSent)
    - [EventTransfer](#evm.v1beta1.EventTransfer)
    - [Gateway](#evm.v1beta1.Gateway)
    - [NetworkInfo](#evm.v1beta1.NetworkInfo)
    - [SigMetadata](#evm.v1beta1.SigMetadata)
    - [TokenDetails](#evm.v1beta1.TokenDetails)
    - [TransactionMetadata](#evm.v1beta1.TransactionMetadata)
    - [TransferKey](#evm.v1beta1.TransferKey)
  
    - [BatchedCommandsStatus](#evm.v1beta1.BatchedCommandsStatus)
    - [DepositStatus](#evm.v1beta1.DepositStatus)
    - [Event.Status](#evm.v1beta1.Event.Status)
    - [Gateway.Status](#evm.v1beta1.Gateway.Status)
    - [SigType](#evm.v1beta1.SigType)
    - [Status](#evm.v1beta1.Status)
    - [TransferKeyType](#evm.v1beta1.TransferKeyType)
  
- [evm/v1beta1/params.proto](#evm/v1beta1/params.proto)
    - [Params](#evm.v1beta1.Params)
    - [PendingChain](#evm.v1beta1.PendingChain)
  
- [evm/v1beta1/genesis.proto](#evm/v1beta1/genesis.proto)
    - [GenesisState](#evm.v1beta1.GenesisState)
    - [GenesisState.Chain](#evm.v1beta1.GenesisState.Chain)
  
- [evm/v1beta1/query.proto](#evm/v1beta1/query.proto)
    - [BatchedCommandsRequest](#evm.v1beta1.BatchedCommandsRequest)
    - [BatchedCommandsResponse](#evm.v1beta1.BatchedCommandsResponse)
    - [BurnerInfoRequest](#evm.v1beta1.BurnerInfoRequest)
    - [BurnerInfoResponse](#evm.v1beta1.BurnerInfoResponse)
    - [BytecodeRequest](#evm.v1beta1.BytecodeRequest)
    - [BytecodeResponse](#evm.v1beta1.BytecodeResponse)
    - [ChainsRequest](#evm.v1beta1.ChainsRequest)
    - [ChainsResponse](#evm.v1beta1.ChainsResponse)
    - [ConfirmationHeightRequest](#evm.v1beta1.ConfirmationHeightRequest)
    - [ConfirmationHeightResponse](#evm.v1beta1.ConfirmationHeightResponse)
    - [DepositQueryParams](#evm.v1beta1.DepositQueryParams)
    - [DepositStateRequest](#evm.v1beta1.DepositStateRequest)
    - [DepositStateResponse](#evm.v1beta1.DepositStateResponse)
    - [EventRequest](#evm.v1beta1.EventRequest)
    - [EventResponse](#evm.v1beta1.EventResponse)
    - [GatewayAddressRequest](#evm.v1beta1.GatewayAddressRequest)
    - [GatewayAddressResponse](#evm.v1beta1.GatewayAddressResponse)
    - [KeyAddressRequest](#evm.v1beta1.KeyAddressRequest)
    - [KeyAddressResponse](#evm.v1beta1.KeyAddressResponse)
    - [KeyAddressResponse.MultisigAddresses](#evm.v1beta1.KeyAddressResponse.MultisigAddresses)
    - [KeyAddressResponse.ThresholdAddress](#evm.v1beta1.KeyAddressResponse.ThresholdAddress)
    - [PendingCommandsRequest](#evm.v1beta1.PendingCommandsRequest)
    - [PendingCommandsResponse](#evm.v1beta1.PendingCommandsResponse)
    - [QueryBurnerAddressResponse](#evm.v1beta1.QueryBurnerAddressResponse)
    - [QueryCommandResponse](#evm.v1beta1.QueryCommandResponse)
    - [QueryCommandResponse.ParamsEntry](#evm.v1beta1.QueryCommandResponse.ParamsEntry)
    - [QueryDepositStateParams](#evm.v1beta1.QueryDepositStateParams)
    - [QueryTokenAddressResponse](#evm.v1beta1.QueryTokenAddressResponse)
  
- [vote/exported/v1beta1/types.proto](#vote/exported/v1beta1/types.proto)
    - [PollMetadata](#vote.exported.v1beta1.PollMetadata)
    - [Vote](#vote.exported.v1beta1.Vote)
  
- [vote/v1beta1/params.proto](#vote/v1beta1/params.proto)
    - [Params](#vote.v1beta1.Params)
  
- [vote/v1beta1/genesis.proto](#vote/v1beta1/genesis.proto)
    - [GenesisState](#vote.v1beta1.GenesisState)
  
- [vote/v1beta1/types.proto](#vote/v1beta1/types.proto)
    - [TalliedVote](#vote.v1beta1.TalliedVote)
  
- [Scalar Value Types](#scalar-value-types)



<a name="axelar/axelarnet/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/axelarnet/v1beta1/params.proto



<a name="axelar.axelarnet.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `route_timeout_window` | [uint64](#uint64) |  | IBC packet route timeout window |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/axelarnet/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/axelarnet/v1beta1/types.proto



<a name="axelar.axelarnet.v1beta1.Asset"></a>

### Asset



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `min_amount` | [bytes](#bytes) |  |  |






<a name="axelar.axelarnet.v1beta1.CosmosChain"></a>

### CosmosChain



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `ibc_path` | [string](#string) |  |  |
| `assets` | [Asset](#axelar.axelarnet.v1beta1.Asset) | repeated | **Deprecated.**  |
| `addr_prefix` | [string](#string) |  |  |






<a name="axelar.axelarnet.v1beta1.IBCTransfer"></a>

### IBCTransfer



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `receiver` | [string](#string) |  |  |
| `token` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `port_id` | [string](#string) |  |  |
| `channel_id` | [string](#string) |  |  |
| `sequence` | [uint64](#uint64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/axelarnet/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/axelarnet/v1beta1/genesis.proto



<a name="axelar.axelarnet.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelar.axelarnet.v1beta1.Params) |  |  |
| `collector_address` | [bytes](#bytes) |  |  |
| `chains` | [CosmosChain](#axelar.axelarnet.v1beta1.CosmosChain) | repeated |  |
| `pending_transfers` | [IBCTransfer](#axelar.axelarnet.v1beta1.IBCTransfer) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/utils/v1beta1/threshold.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/utils/v1beta1/threshold.proto



<a name="axelar.utils.v1beta1.Threshold"></a>

### Threshold



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `numerator` | [int64](#int64) |  | split threshold into Numerator and denominator to avoid floating point errors down the line |
| `denominator` | [int64](#int64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/tss/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/tss/exported/v1beta1/types.proto



<a name="axelar.tss.exported.v1beta1.Key"></a>

### Key



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `role` | [KeyRole](#axelar.tss.exported.v1beta1.KeyRole) |  |  |
| `type` | [KeyType](#axelar.tss.exported.v1beta1.KeyType) |  |  |
| `ecdsa_key` | [Key.ECDSAKey](#axelar.tss.exported.v1beta1.Key.ECDSAKey) |  |  |
| `multisig_key` | [Key.MultisigKey](#axelar.tss.exported.v1beta1.Key.MultisigKey) |  |  |
| `rotated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `rotation_count` | [int64](#int64) |  |  |
| `chain` | [string](#string) |  |  |
| `snapshot_counter` | [int64](#int64) |  |  |






<a name="axelar.tss.exported.v1beta1.Key.ECDSAKey"></a>

### Key.ECDSAKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [bytes](#bytes) |  |  |






<a name="axelar.tss.exported.v1beta1.Key.MultisigKey"></a>

### Key.MultisigKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `values` | [bytes](#bytes) | repeated |  |
| `threshold` | [int64](#int64) |  |  |






<a name="axelar.tss.exported.v1beta1.KeyRequirement"></a>

### KeyRequirement
KeyRequirement defines requirements for keys


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_role` | [KeyRole](#axelar.tss.exported.v1beta1.KeyRole) |  |  |
| `key_type` | [KeyType](#axelar.tss.exported.v1beta1.KeyType) |  |  |
| `min_keygen_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `safety_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `key_share_distribution_policy` | [KeyShareDistributionPolicy](#axelar.tss.exported.v1beta1.KeyShareDistributionPolicy) |  |  |
| `max_total_share_count` | [int64](#int64) |  |  |
| `min_total_share_count` | [int64](#int64) |  |  |
| `keygen_voting_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `sign_voting_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `keygen_timeout` | [int64](#int64) |  |  |
| `sign_timeout` | [int64](#int64) |  |  |






<a name="axelar.tss.exported.v1beta1.SigKeyPair"></a>

### SigKeyPair
PubKeyInfo holds a pubkey and a signature


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pub_key` | [bytes](#bytes) |  |  |
| `signature` | [bytes](#bytes) |  |  |






<a name="axelar.tss.exported.v1beta1.SignInfo"></a>

### SignInfo
SignInfo holds information about a sign request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `sig_id` | [string](#string) |  |  |
| `msg` | [bytes](#bytes) |  |  |
| `snapshot_counter` | [int64](#int64) |  |  |
| `request_module` | [string](#string) |  |  |
| `metadata` | [string](#string) |  | **Deprecated.**  |
| `module_metadata` | [google.protobuf.Any](#google.protobuf.Any) |  |  |






<a name="axelar.tss.exported.v1beta1.Signature"></a>

### Signature
Signature holds public key and ECDSA signature


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sig_id` | [string](#string) |  |  |
| `single_sig` | [Signature.SingleSig](#axelar.tss.exported.v1beta1.Signature.SingleSig) |  |  |
| `multi_sig` | [Signature.MultiSig](#axelar.tss.exported.v1beta1.Signature.MultiSig) |  |  |
| `sig_status` | [SigStatus](#axelar.tss.exported.v1beta1.SigStatus) |  |  |






<a name="axelar.tss.exported.v1beta1.Signature.MultiSig"></a>

### Signature.MultiSig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sig_key_pairs` | [SigKeyPair](#axelar.tss.exported.v1beta1.SigKeyPair) | repeated |  |






<a name="axelar.tss.exported.v1beta1.Signature.SingleSig"></a>

### Signature.SingleSig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sig_key_pair` | [SigKeyPair](#axelar.tss.exported.v1beta1.SigKeyPair) |  |  |





 <!-- end messages -->


<a name="axelar.tss.exported.v1beta1.AckType"></a>

### AckType


| Name | Number | Description |
| ---- | ------ | ----------- |
| ACK_TYPE_UNSPECIFIED | 0 |  |
| ACK_TYPE_KEYGEN | 1 |  |
| ACK_TYPE_SIGN | 2 |  |



<a name="axelar.tss.exported.v1beta1.KeyRole"></a>

### KeyRole


| Name | Number | Description |
| ---- | ------ | ----------- |
| KEY_ROLE_UNSPECIFIED | 0 |  |
| KEY_ROLE_MASTER_KEY | 1 |  |
| KEY_ROLE_SECONDARY_KEY | 2 |  |
| KEY_ROLE_EXTERNAL_KEY | 3 |  |



<a name="axelar.tss.exported.v1beta1.KeyShareDistributionPolicy"></a>

### KeyShareDistributionPolicy


| Name | Number | Description |
| ---- | ------ | ----------- |
| KEY_SHARE_DISTRIBUTION_POLICY_UNSPECIFIED | 0 |  |
| KEY_SHARE_DISTRIBUTION_POLICY_WEIGHTED_BY_STAKE | 1 |  |
| KEY_SHARE_DISTRIBUTION_POLICY_ONE_PER_VALIDATOR | 2 |  |



<a name="axelar.tss.exported.v1beta1.KeyType"></a>

### KeyType


| Name | Number | Description |
| ---- | ------ | ----------- |
| KEY_TYPE_UNSPECIFIED | 0 |  |
| KEY_TYPE_NONE | 1 |  |
| KEY_TYPE_THRESHOLD | 2 |  |
| KEY_TYPE_MULTISIG | 3 |  |



<a name="axelar.tss.exported.v1beta1.SigStatus"></a>

### SigStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| SIG_STATUS_UNSPECIFIED | 0 |  |
| SIG_STATUS_QUEUED | 1 |  |
| SIG_STATUS_SIGNING | 2 |  |
| SIG_STATUS_SIGNED | 3 |  |
| SIG_STATUS_ABORTED | 4 |  |
| SIG_STATUS_INVALID | 5 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/nexus/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/nexus/exported/v1beta1/types.proto



<a name="axelar.nexus.exported.v1beta1.Asset"></a>

### Asset



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `is_native_asset` | [bool](#bool) |  |  |






<a name="axelar.nexus.exported.v1beta1.Chain"></a>

### Chain
Chain represents the properties of a registered blockchain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `supports_foreign_assets` | [bool](#bool) |  |  |
| `key_type` | [axelar.tss.exported.v1beta1.KeyType](#axelar.tss.exported.v1beta1.KeyType) |  |  |
| `module` | [string](#string) |  |  |






<a name="axelar.nexus.exported.v1beta1.CrossChainAddress"></a>

### CrossChainAddress
CrossChainAddress represents a generalized address on any registered chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [Chain](#axelar.nexus.exported.v1beta1.Chain) |  |  |
| `address` | [string](#string) |  |  |






<a name="axelar.nexus.exported.v1beta1.CrossChainTransfer"></a>

### CrossChainTransfer
CrossChainTransfer represents a generalized transfer of some asset to a
registered blockchain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `recipient` | [CrossChainAddress](#axelar.nexus.exported.v1beta1.CrossChainAddress) |  |  |
| `asset` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `id` | [uint64](#uint64) |  |  |
| `state` | [TransferState](#axelar.nexus.exported.v1beta1.TransferState) |  |  |






<a name="axelar.nexus.exported.v1beta1.FeeInfo"></a>

### FeeInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `asset` | [string](#string) |  |  |
| `fee_rate` | [bytes](#bytes) |  |  |
| `min_fee` | [bytes](#bytes) |  |  |
| `max_fee` | [bytes](#bytes) |  |  |






<a name="axelar.nexus.exported.v1beta1.TransferFee"></a>

### TransferFee
TransferFee represents accumulated fees generated by the network


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |





 <!-- end messages -->


<a name="axelar.nexus.exported.v1beta1.TransferState"></a>

### TransferState


| Name | Number | Description |
| ---- | ------ | ----------- |
| TRANSFER_STATE_UNSPECIFIED | 0 |  |
| TRANSFER_STATE_PENDING | 1 |  |
| TRANSFER_STATE_ARCHIVED | 2 |  |
| TRANSFER_STATE_INSUFFICIENT_AMOUNT | 3 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/utils/v1beta1/bitmap.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/utils/v1beta1/bitmap.proto



<a name="axelar.utils.v1beta1.Bitmap"></a>

### Bitmap



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `true_count_cache` | [CircularBuffer](#axelar.utils.v1beta1.CircularBuffer) |  |  |






<a name="axelar.utils.v1beta1.CircularBuffer"></a>

### CircularBuffer



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `cumulative_value` | [uint64](#uint64) | repeated |  |
| `index` | [int32](#int32) |  |  |
| `max_size` | [int32](#int32) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/nexus/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/nexus/v1beta1/types.proto



<a name="axelar.nexus.v1beta1.ChainState"></a>

### ChainState
ChainState represents the state of a registered blockchain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [axelar.nexus.exported.v1beta1.Chain](#axelar.nexus.exported.v1beta1.Chain) |  |  |
| `maintainers` | [bytes](#bytes) | repeated | **Deprecated.**  |
| `activated` | [bool](#bool) |  |  |
| `assets` | [axelar.nexus.exported.v1beta1.Asset](#axelar.nexus.exported.v1beta1.Asset) | repeated |  |
| `maintainer_states` | [MaintainerState](#axelar.nexus.v1beta1.MaintainerState) | repeated |  |






<a name="axelar.nexus.v1beta1.LinkedAddresses"></a>

### LinkedAddresses



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposit_address` | [axelar.nexus.exported.v1beta1.CrossChainAddress](#axelar.nexus.exported.v1beta1.CrossChainAddress) |  |  |
| `recipient_address` | [axelar.nexus.exported.v1beta1.CrossChainAddress](#axelar.nexus.exported.v1beta1.CrossChainAddress) |  |  |






<a name="axelar.nexus.v1beta1.MaintainerState"></a>

### MaintainerState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [bytes](#bytes) |  |  |
| `missing_votes` | [axelar.utils.v1beta1.Bitmap](#axelar.utils.v1beta1.Bitmap) |  |  |
| `incorrect_votes` | [axelar.utils.v1beta1.Bitmap](#axelar.utils.v1beta1.Bitmap) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/nexus/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/nexus/v1beta1/query.proto



<a name="axelar.nexus.v1beta1.AssetsRequest"></a>

### AssetsRequest
AssetsRequest represents a message that queries the registered assets of a
chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |






<a name="axelar.nexus.v1beta1.AssetsResponse"></a>

### AssetsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `assets` | [string](#string) | repeated |  |






<a name="axelar.nexus.v1beta1.ChainStateRequest"></a>

### ChainStateRequest
ChainStateRequest represents a message that queries the state of a chain
registered on the network


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |






<a name="axelar.nexus.v1beta1.ChainStateResponse"></a>

### ChainStateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `state` | [ChainState](#axelar.nexus.v1beta1.ChainState) |  |  |






<a name="axelar.nexus.v1beta1.ChainsByAssetRequest"></a>

### ChainsByAssetRequest
ChainsByAssetRequest represents a message that queries the chains
that support an asset on the network


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset` | [string](#string) |  |  |






<a name="axelar.nexus.v1beta1.ChainsByAssetResponse"></a>

### ChainsByAssetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chains` | [string](#string) | repeated |  |






<a name="axelar.nexus.v1beta1.ChainsRequest"></a>

### ChainsRequest
ChainsRequest represents a message that queries the chains
registered on the network






<a name="axelar.nexus.v1beta1.ChainsResponse"></a>

### ChainsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chains` | [string](#string) | repeated |  |






<a name="axelar.nexus.v1beta1.FeeInfoRequest"></a>

### FeeInfoRequest
FeeInfoRequest represents a message that queries the transfer fees associated
to an asset on a chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `asset` | [string](#string) |  |  |






<a name="axelar.nexus.v1beta1.FeeInfoResponse"></a>

### FeeInfoResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `fee_info` | [axelar.nexus.exported.v1beta1.FeeInfo](#axelar.nexus.exported.v1beta1.FeeInfo) |  |  |






<a name="axelar.nexus.v1beta1.LatestDepositAddressRequest"></a>

### LatestDepositAddressRequest
LatestDepositAddressRequest represents a message that queries a deposit
address by recipient address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `recipient_addr` | [string](#string) |  |  |
| `recipient_chain` | [string](#string) |  |  |
| `deposit_chain` | [string](#string) |  |  |






<a name="axelar.nexus.v1beta1.LatestDepositAddressResponse"></a>

### LatestDepositAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposit_addr` | [string](#string) |  |  |






<a name="axelar.nexus.v1beta1.QueryChainMaintainersResponse"></a>

### QueryChainMaintainersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `maintainers` | [bytes](#bytes) | repeated |  |






<a name="axelar.nexus.v1beta1.TransferFeeRequest"></a>

### TransferFeeRequest
TransferFeeRequest represents a message that queries the fees charged by
the network for a cross-chain transfer


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source_chain` | [string](#string) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `amount` | [string](#string) |  |  |






<a name="axelar.nexus.v1beta1.TransferFeeResponse"></a>

### TransferFeeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="axelar.nexus.v1beta1.TransfersForChainRequest"></a>

### TransfersForChainRequest
TransfersForChainRequest represents a message that queries the
transfers for the specified chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `state` | [axelar.nexus.exported.v1beta1.TransferState](#axelar.nexus.exported.v1beta1.TransferState) |  |  |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |






<a name="axelar.nexus.v1beta1.TransfersForChainResponse"></a>

### TransfersForChainResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `transfers` | [axelar.nexus.exported.v1beta1.CrossChainTransfer](#axelar.nexus.exported.v1beta1.CrossChainTransfer) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/axelarnet/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/axelarnet/v1beta1/query.proto



<a name="axelar.axelarnet.v1beta1.PendingIBCTransferCountRequest"></a>

### PendingIBCTransferCountRequest







<a name="axelar.axelarnet.v1beta1.PendingIBCTransferCountResponse"></a>

### PendingIBCTransferCountResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `transfers_by_chain` | [PendingIBCTransferCountResponse.TransfersByChainEntry](#axelar.axelarnet.v1beta1.PendingIBCTransferCountResponse.TransfersByChainEntry) | repeated |  |






<a name="axelar.axelarnet.v1beta1.PendingIBCTransferCountResponse.TransfersByChainEntry"></a>

### PendingIBCTransferCountResponse.TransfersByChainEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [uint32](#uint32) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/permission/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/permission/exported/v1beta1/types.proto


 <!-- end messages -->


<a name="axelar.permission.exported.v1beta1.Role"></a>

### Role


| Name | Number | Description |
| ---- | ------ | ----------- |
| ROLE_UNSPECIFIED | 0 |  |
| ROLE_UNRESTRICTED | 1 |  |
| ROLE_CHAIN_MANAGEMENT | 2 |  |
| ROLE_ACCESS_CONTROL | 3 |  |


 <!-- end enums -->


<a name="axelar/permission/exported/v1beta1/types.proto-extensions"></a>

### File-level Extensions
| Extension | Type | Base | Number | Description |
| --------- | ---- | ---- | ------ | ----------- |
| `permission_role` | Role | .google.protobuf.MessageOptions | 50000 | 50000-99999 reserved for use withing individual organizations |

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/axelarnet/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/axelarnet/v1beta1/tx.proto



<a name="axelar.axelarnet.v1beta1.AddCosmosBasedChainRequest"></a>

### AddCosmosBasedChainRequest
MsgAddCosmosBasedChain represents a message to register a cosmos based chain
to nexus


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [axelar.nexus.exported.v1beta1.Chain](#axelar.nexus.exported.v1beta1.Chain) |  |  |
| `addr_prefix` | [string](#string) |  |  |
| `native_assets` | [axelar.nexus.exported.v1beta1.Asset](#axelar.nexus.exported.v1beta1.Asset) | repeated |  |






<a name="axelar.axelarnet.v1beta1.AddCosmosBasedChainResponse"></a>

### AddCosmosBasedChainResponse







<a name="axelar.axelarnet.v1beta1.ConfirmDepositRequest"></a>

### ConfirmDepositRequest
MsgConfirmDeposit represents a deposit confirmation message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `deposit_address` | [bytes](#bytes) |  |  |
| `denom` | [string](#string) |  |  |






<a name="axelar.axelarnet.v1beta1.ConfirmDepositResponse"></a>

### ConfirmDepositResponse







<a name="axelar.axelarnet.v1beta1.ExecutePendingTransfersRequest"></a>

### ExecutePendingTransfersRequest
MsgExecutePendingTransfers represents a message to trigger transfer all
pending transfers


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |






<a name="axelar.axelarnet.v1beta1.ExecutePendingTransfersResponse"></a>

### ExecutePendingTransfersResponse







<a name="axelar.axelarnet.v1beta1.LinkRequest"></a>

### LinkRequest
MsgLink represents a message to link a cross-chain address to an Axelar
address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `recipient_addr` | [string](#string) |  |  |
| `recipient_chain` | [string](#string) |  |  |
| `asset` | [string](#string) |  |  |






<a name="axelar.axelarnet.v1beta1.LinkResponse"></a>

### LinkResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposit_addr` | [string](#string) |  |  |






<a name="axelar.axelarnet.v1beta1.RegisterAssetRequest"></a>

### RegisterAssetRequest
RegisterAssetRequest represents a message to register an asset to a cosmos
based chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `asset` | [axelar.nexus.exported.v1beta1.Asset](#axelar.nexus.exported.v1beta1.Asset) |  |  |






<a name="axelar.axelarnet.v1beta1.RegisterAssetResponse"></a>

### RegisterAssetResponse







<a name="axelar.axelarnet.v1beta1.RegisterFeeCollectorRequest"></a>

### RegisterFeeCollectorRequest
RegisterFeeCollectorRequest represents a message to register axelarnet fee
collector account


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `fee_collector` | [bytes](#bytes) |  |  |






<a name="axelar.axelarnet.v1beta1.RegisterFeeCollectorResponse"></a>

### RegisterFeeCollectorResponse







<a name="axelar.axelarnet.v1beta1.RegisterIBCPathRequest"></a>

### RegisterIBCPathRequest
MSgRegisterIBCPath represents a message to register an IBC tracing path for
a cosmos chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `path` | [string](#string) |  |  |






<a name="axelar.axelarnet.v1beta1.RegisterIBCPathResponse"></a>

### RegisterIBCPathResponse







<a name="axelar.axelarnet.v1beta1.RouteIBCTransfersRequest"></a>

### RouteIBCTransfersRequest
RouteIBCTransfersRequest represents a message to route pending transfers to
cosmos based chains


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |






<a name="axelar.axelarnet.v1beta1.RouteIBCTransfersResponse"></a>

### RouteIBCTransfersResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/axelarnet/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/axelarnet/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelar.axelarnet.v1beta1.MsgService"></a>

### MsgService
Msg defines the axelarnet Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Link` | [LinkRequest](#axelar.axelarnet.v1beta1.LinkRequest) | [LinkResponse](#axelar.axelarnet.v1beta1.LinkResponse) |  | POST|/axelar/axelarnet/link|
| `ConfirmDeposit` | [ConfirmDepositRequest](#axelar.axelarnet.v1beta1.ConfirmDepositRequest) | [ConfirmDepositResponse](#axelar.axelarnet.v1beta1.ConfirmDepositResponse) |  | POST|/axelar/axelarnet/confirm_deposit|
| `ExecutePendingTransfers` | [ExecutePendingTransfersRequest](#axelar.axelarnet.v1beta1.ExecutePendingTransfersRequest) | [ExecutePendingTransfersResponse](#axelar.axelarnet.v1beta1.ExecutePendingTransfersResponse) |  | POST|/axelar/axelarnet/execute_pending_transfers|
| `RegisterIBCPath` | [RegisterIBCPathRequest](#axelar.axelarnet.v1beta1.RegisterIBCPathRequest) | [RegisterIBCPathResponse](#axelar.axelarnet.v1beta1.RegisterIBCPathResponse) |  | POST|/axelar/axelarnet/register_ibc_path|
| `AddCosmosBasedChain` | [AddCosmosBasedChainRequest](#axelar.axelarnet.v1beta1.AddCosmosBasedChainRequest) | [AddCosmosBasedChainResponse](#axelar.axelarnet.v1beta1.AddCosmosBasedChainResponse) |  | POST|/axelar/axelarnet/add_cosmos_based_chain|
| `RegisterAsset` | [RegisterAssetRequest](#axelar.axelarnet.v1beta1.RegisterAssetRequest) | [RegisterAssetResponse](#axelar.axelarnet.v1beta1.RegisterAssetResponse) |  | POST|/axelar/axelarnet/register_asset|
| `RouteIBCTransfers` | [RouteIBCTransfersRequest](#axelar.axelarnet.v1beta1.RouteIBCTransfersRequest) | [RouteIBCTransfersResponse](#axelar.axelarnet.v1beta1.RouteIBCTransfersResponse) |  | POST|/axelar/axelarnet/route_ibc_transfers|
| `RegisterFeeCollector` | [RegisterFeeCollectorRequest](#axelar.axelarnet.v1beta1.RegisterFeeCollectorRequest) | [RegisterFeeCollectorResponse](#axelar.axelarnet.v1beta1.RegisterFeeCollectorResponse) |  | POST|/axelar/axelarnet/register_fee_collector|


<a name="axelar.axelarnet.v1beta1.QueryService"></a>

### QueryService
QueryService defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `PendingIBCTransferCount` | [PendingIBCTransferCountRequest](#axelar.axelarnet.v1beta1.PendingIBCTransferCountRequest) | [PendingIBCTransferCountResponse](#axelar.axelarnet.v1beta1.PendingIBCTransferCountResponse) | PendingIBCTransferCount queries the pending ibc transfers for all chains | GET|/axelar/axelarnet/v1beta1/ibc_transfer_count|

 <!-- end services -->



<a name="axelar/bitcoin/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/bitcoin/v1beta1/types.proto



<a name="axelar.bitcoin.v1beta1.AddressInfo"></a>

### AddressInfo
AddressInfo is a wrapper containing the Bitcoin P2WSH address, it's
corresponding script and the underlying key


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `role` | [AddressRole](#axelar.bitcoin.v1beta1.AddressRole) |  |  |
| `redeem_script` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `max_sig_count` | [uint32](#uint32) |  |  |
| `spending_condition` | [AddressInfo.SpendingCondition](#axelar.bitcoin.v1beta1.AddressInfo.SpendingCondition) |  |  |






<a name="axelar.bitcoin.v1beta1.AddressInfo.SpendingCondition"></a>

### AddressInfo.SpendingCondition



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `internal_key_ids` | [string](#string) | repeated | internal_key_ids lists the internal key IDs that one of which has to sign regardless of locktime |
| `external_key_ids` | [string](#string) | repeated | external_key_ids lists the external key IDs that external_multisig_threshold of which have to sign to spend before locktime if set |
| `external_multisig_threshold` | [int64](#int64) |  |  |
| `lock_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="axelar.bitcoin.v1beta1.Network"></a>

### Network



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |






<a name="axelar.bitcoin.v1beta1.OutPointInfo"></a>

### OutPointInfo
OutPointInfo describes all the necessary information to confirm the outPoint
of a transaction


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `out_point` | [string](#string) |  |  |
| `amount` | [int64](#int64) |  |  |
| `address` | [string](#string) |  |  |






<a name="axelar.bitcoin.v1beta1.SignedTx"></a>

### SignedTx



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `type` | [TxType](#axelar.bitcoin.v1beta1.TxType) |  |  |
| `tx` | [bytes](#bytes) |  |  |
| `prev_signed_tx_hash` | [bytes](#bytes) |  |  |
| `confirmation_required` | [bool](#bool) |  |  |
| `anyone_can_spend_vout` | [uint32](#uint32) |  |  |






<a name="axelar.bitcoin.v1beta1.UnsignedTx"></a>

### UnsignedTx



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `type` | [TxType](#axelar.bitcoin.v1beta1.TxType) |  |  |
| `tx` | [bytes](#bytes) |  |  |
| `info` | [UnsignedTx.Info](#axelar.bitcoin.v1beta1.UnsignedTx.Info) |  |  |
| `status` | [TxStatus](#axelar.bitcoin.v1beta1.TxStatus) |  |  |
| `confirmation_required` | [bool](#bool) |  |  |
| `anyone_can_spend_vout` | [uint32](#uint32) |  |  |
| `prev_aborted_key_id` | [string](#string) |  |  |
| `internal_transfer_amount` | [int64](#int64) |  |  |






<a name="axelar.bitcoin.v1beta1.UnsignedTx.Info"></a>

### UnsignedTx.Info



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `rotate_key` | [bool](#bool) |  |  |
| `input_infos` | [UnsignedTx.Info.InputInfo](#axelar.bitcoin.v1beta1.UnsignedTx.Info.InputInfo) | repeated |  |






<a name="axelar.bitcoin.v1beta1.UnsignedTx.Info.InputInfo"></a>

### UnsignedTx.Info.InputInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sig_requirements` | [UnsignedTx.Info.InputInfo.SigRequirement](#axelar.bitcoin.v1beta1.UnsignedTx.Info.InputInfo.SigRequirement) | repeated |  |






<a name="axelar.bitcoin.v1beta1.UnsignedTx.Info.InputInfo.SigRequirement"></a>

### UnsignedTx.Info.InputInfo.SigRequirement



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `sig_hash` | [bytes](#bytes) |  |  |





 <!-- end messages -->


<a name="axelar.bitcoin.v1beta1.AddressRole"></a>

### AddressRole


| Name | Number | Description |
| ---- | ------ | ----------- |
| ADDRESS_ROLE_UNSPECIFIED | 0 |  |
| ADDRESS_ROLE_DEPOSIT | 1 |  |
| ADDRESS_ROLE_CONSOLIDATION | 2 |  |



<a name="axelar.bitcoin.v1beta1.OutPointState"></a>

### OutPointState


| Name | Number | Description |
| ---- | ------ | ----------- |
| OUT_POINT_STATE_UNSPECIFIED | 0 |  |
| OUT_POINT_STATE_PENDING | 1 |  |
| OUT_POINT_STATE_CONFIRMED | 2 |  |
| OUT_POINT_STATE_SPENT | 3 |  |



<a name="axelar.bitcoin.v1beta1.TxStatus"></a>

### TxStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| TX_STATUS_UNSPECIFIED | 0 |  |
| TX_STATUS_CREATED | 1 |  |
| TX_STATUS_SIGNING | 2 |  |
| TX_STATUS_ABORTED | 3 |  |
| TX_STATUS_SIGNED | 4 |  |



<a name="axelar.bitcoin.v1beta1.TxType"></a>

### TxType


| Name | Number | Description |
| ---- | ------ | ----------- |
| TX_TYPE_UNSPECIFIED | 0 |  |
| TX_TYPE_MASTER_CONSOLIDATION | 1 |  |
| TX_TYPE_SECONDARY_CONSOLIDATION | 2 |  |
| TX_TYPE_RESCUE | 3 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/bitcoin/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/bitcoin/v1beta1/params.proto



<a name="axelar.bitcoin.v1beta1.Params"></a>

### Params



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `network` | [Network](#axelar.bitcoin.v1beta1.Network) |  |  |
| `confirmation_height` | [uint64](#uint64) |  |  |
| `revote_locking_period` | [int64](#int64) |  |  |
| `sig_check_interval` | [int64](#int64) |  |  |
| `min_output_amount` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  |  |
| `max_input_count` | [int64](#int64) |  |  |
| `max_secondary_output_amount` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  |  |
| `master_key_retention_period` | [int64](#int64) |  |  |
| `master_address_internal_key_lock_duration` | [int64](#int64) |  |  |
| `master_address_external_key_lock_duration` | [int64](#int64) |  |  |
| `voting_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `min_voter_count` | [int64](#int64) |  |  |
| `max_tx_size` | [int64](#int64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/bitcoin/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/bitcoin/v1beta1/genesis.proto



<a name="axelar.bitcoin.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelar.bitcoin.v1beta1.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/bitcoin/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/bitcoin/v1beta1/query.proto



<a name="axelar.bitcoin.v1beta1.DepositQueryParams"></a>

### DepositQueryParams
DepositQueryParams describe the parameters used to query for a Bitcoin
deposit address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `chain` | [string](#string) |  |  |






<a name="axelar.bitcoin.v1beta1.QueryAddressResponse"></a>

### QueryAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="axelar.bitcoin.v1beta1.QueryDepositStatusResponse"></a>

### QueryDepositStatusResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |
| `status` | [OutPointState](#axelar.bitcoin.v1beta1.OutPointState) |  |  |






<a name="axelar.bitcoin.v1beta1.QueryTxResponse"></a>

### QueryTxResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx` | [string](#string) |  |  |
| `status` | [TxStatus](#axelar.bitcoin.v1beta1.TxStatus) |  |  |
| `confirmation_required` | [bool](#bool) |  |  |
| `prev_signed_tx_hash` | [string](#string) |  |  |
| `anyone_can_spend_vout` | [uint32](#uint32) |  |  |
| `signing_infos` | [QueryTxResponse.SigningInfo](#axelar.bitcoin.v1beta1.QueryTxResponse.SigningInfo) | repeated |  |






<a name="axelar.bitcoin.v1beta1.QueryTxResponse.SigningInfo"></a>

### QueryTxResponse.SigningInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `redeem_script` | [string](#string) |  |  |
| `amount` | [int64](#int64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/snapshot/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/snapshot/exported/v1beta1/types.proto



<a name="axelar.snapshot.exported.v1beta1.Snapshot"></a>

### Snapshot



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validators` | [Validator](#axelar.snapshot.exported.v1beta1.Validator) | repeated |  |
| `timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `height` | [int64](#int64) |  |  |
| `total_share_count` | [bytes](#bytes) |  |  |
| `counter` | [int64](#int64) |  |  |
| `key_share_distribution_policy` | [axelar.tss.exported.v1beta1.KeyShareDistributionPolicy](#axelar.tss.exported.v1beta1.KeyShareDistributionPolicy) |  |  |
| `corruption_threshold` | [int64](#int64) |  |  |






<a name="axelar.snapshot.exported.v1beta1.Validator"></a>

### Validator



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sdk_validator` | [google.protobuf.Any](#google.protobuf.Any) |  |  |
| `share_count` | [int64](#int64) |  |  |





 <!-- end messages -->


<a name="axelar.snapshot.exported.v1beta1.ValidatorIllegibility"></a>

### ValidatorIllegibility


| Name | Number | Description |
| ---- | ------ | ----------- |
| VALIDATOR_ILLEGIBILITY_UNSPECIFIED | 0 | these enum values are used for bitwise operations, therefore they need to be powers of 2 |
| VALIDATOR_ILLEGIBILITY_TOMBSTONED | 1 |  |
| VALIDATOR_ILLEGIBILITY_JAILED | 2 |  |
| VALIDATOR_ILLEGIBILITY_MISSED_TOO_MANY_BLOCKS | 4 |  |
| VALIDATOR_ILLEGIBILITY_NO_PROXY_REGISTERED | 8 |  |
| VALIDATOR_ILLEGIBILITY_TSS_SUSPENDED | 16 |  |
| VALIDATOR_ILLEGIBILITY_PROXY_INSUFICIENT_FUNDS | 32 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/vote/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/vote/exported/v1beta1/types.proto



<a name="axelar.vote.exported.v1beta1.PollKey"></a>

### PollKey
PollKey represents the key data for a poll


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |






<a name="axelar.vote.exported.v1beta1.PollMetadata"></a>

### PollMetadata
PollMetadata represents a poll with write-in voting, i.e. the result of the
vote can have any data type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [PollKey](#axelar.vote.exported.v1beta1.PollKey) |  |  |
| `expires_at` | [int64](#int64) |  |  |
| `result` | [google.protobuf.Any](#google.protobuf.Any) |  |  |
| `voting_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `state` | [PollState](#axelar.vote.exported.v1beta1.PollState) |  |  |
| `min_voter_count` | [int64](#int64) |  |  |
| `voters` | [Voter](#axelar.vote.exported.v1beta1.Voter) | repeated |  |
| `total_voting_power` | [bytes](#bytes) |  |  |
| `reward_pool_name` | [string](#string) |  |  |






<a name="axelar.vote.exported.v1beta1.Vote"></a>

### Vote



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `result` | [google.protobuf.Any](#google.protobuf.Any) |  |  |






<a name="axelar.vote.exported.v1beta1.Voter"></a>

### Voter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validator` | [bytes](#bytes) |  |  |
| `voting_power` | [int64](#int64) |  |  |





 <!-- end messages -->


<a name="axelar.vote.exported.v1beta1.PollState"></a>

### PollState


| Name | Number | Description |
| ---- | ------ | ----------- |
| POLL_STATE_UNSPECIFIED | 0 | these enum values are used for bitwise operations, therefore they need to be powers of 2 |
| POLL_STATE_PENDING | 1 |  |
| POLL_STATE_COMPLETED | 2 |  |
| POLL_STATE_FAILED | 4 |  |
| POLL_STATE_EXPIRED | 8 |  |
| POLL_STATE_ALLOW_OVERRIDE | 16 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/bitcoin/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/bitcoin/v1beta1/tx.proto



<a name="axelar.bitcoin.v1beta1.ConfirmOutpointRequest"></a>

### ConfirmOutpointRequest
MsgConfirmOutpoint represents a message to trigger the confirmation of a
Bitcoin outpoint


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `out_point_info` | [OutPointInfo](#axelar.bitcoin.v1beta1.OutPointInfo) |  |  |






<a name="axelar.bitcoin.v1beta1.ConfirmOutpointResponse"></a>

### ConfirmOutpointResponse







<a name="axelar.bitcoin.v1beta1.CreateMasterTxRequest"></a>

### CreateMasterTxRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `secondary_key_amount` | [int64](#int64) |  |  |






<a name="axelar.bitcoin.v1beta1.CreateMasterTxResponse"></a>

### CreateMasterTxResponse







<a name="axelar.bitcoin.v1beta1.CreatePendingTransfersTxRequest"></a>

### CreatePendingTransfersTxRequest
CreatePendingTransfersTxRequest represents a message to trigger the creation
of a secondary key consolidation transaction


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `master_key_amount` | [int64](#int64) |  |  |






<a name="axelar.bitcoin.v1beta1.CreatePendingTransfersTxResponse"></a>

### CreatePendingTransfersTxResponse







<a name="axelar.bitcoin.v1beta1.CreateRescueTxRequest"></a>

### CreateRescueTxRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |






<a name="axelar.bitcoin.v1beta1.CreateRescueTxResponse"></a>

### CreateRescueTxResponse







<a name="axelar.bitcoin.v1beta1.LinkRequest"></a>

### LinkRequest
MsgLink represents a message to link a cross-chain address to a Bitcoin
address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `recipient_addr` | [string](#string) |  |  |
| `recipient_chain` | [string](#string) |  |  |






<a name="axelar.bitcoin.v1beta1.LinkResponse"></a>

### LinkResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposit_addr` | [string](#string) |  |  |






<a name="axelar.bitcoin.v1beta1.SignTxRequest"></a>

### SignTxRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `tx_type` | [TxType](#axelar.bitcoin.v1beta1.TxType) |  |  |






<a name="axelar.bitcoin.v1beta1.SignTxResponse"></a>

### SignTxResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `position` | [int64](#int64) |  |  |






<a name="axelar.bitcoin.v1beta1.SubmitExternalSignatureRequest"></a>

### SubmitExternalSignatureRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `signature` | [bytes](#bytes) |  |  |
| `sig_hash` | [bytes](#bytes) |  |  |






<a name="axelar.bitcoin.v1beta1.SubmitExternalSignatureResponse"></a>

### SubmitExternalSignatureResponse







<a name="axelar.bitcoin.v1beta1.VoteConfirmOutpointRequest"></a>

### VoteConfirmOutpointRequest
MsgVoteConfirmOutpoint represents a message to that votes on an outpoint


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll_key` | [axelar.vote.exported.v1beta1.PollKey](#axelar.vote.exported.v1beta1.PollKey) |  |  |
| `out_point` | [string](#string) |  |  |
| `confirmed` | [bool](#bool) |  |  |






<a name="axelar.bitcoin.v1beta1.VoteConfirmOutpointResponse"></a>

### VoteConfirmOutpointResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `status` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/bitcoin/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/bitcoin/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelar.bitcoin.v1beta1.MsgService"></a>

### MsgService
Msg defines the bitcoin Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Link` | [LinkRequest](#axelar.bitcoin.v1beta1.LinkRequest) | [LinkResponse](#axelar.bitcoin.v1beta1.LinkResponse) |  | POST|/axelar/bitcoin/link|
| `ConfirmOutpoint` | [ConfirmOutpointRequest](#axelar.bitcoin.v1beta1.ConfirmOutpointRequest) | [ConfirmOutpointResponse](#axelar.bitcoin.v1beta1.ConfirmOutpointResponse) |  | POST|/axelar/bitcoin/confirm|
| `VoteConfirmOutpoint` | [VoteConfirmOutpointRequest](#axelar.bitcoin.v1beta1.VoteConfirmOutpointRequest) | [VoteConfirmOutpointResponse](#axelar.bitcoin.v1beta1.VoteConfirmOutpointResponse) |  | POST|/axelar/bitcoin/vote_confirm|
| `CreatePendingTransfersTx` | [CreatePendingTransfersTxRequest](#axelar.bitcoin.v1beta1.CreatePendingTransfersTxRequest) | [CreatePendingTransfersTxResponse](#axelar.bitcoin.v1beta1.CreatePendingTransfersTxResponse) |  | POST|/axelar/bitcoin/create_pending_transfers_tx|
| `CreateMasterTx` | [CreateMasterTxRequest](#axelar.bitcoin.v1beta1.CreateMasterTxRequest) | [CreateMasterTxResponse](#axelar.bitcoin.v1beta1.CreateMasterTxResponse) |  | POST|/axelar/bitcoin/create_master_tx|
| `CreateRescueTx` | [CreateRescueTxRequest](#axelar.bitcoin.v1beta1.CreateRescueTxRequest) | [CreateRescueTxResponse](#axelar.bitcoin.v1beta1.CreateRescueTxResponse) |  | POST|/axelar/bitcoin/create_rescue_tx|
| `SignTx` | [SignTxRequest](#axelar.bitcoin.v1beta1.SignTxRequest) | [SignTxResponse](#axelar.bitcoin.v1beta1.SignTxResponse) |  | POST|/axelar/bitcoin/sign_tx|
| `SubmitExternalSignature` | [SubmitExternalSignatureRequest](#axelar.bitcoin.v1beta1.SubmitExternalSignatureRequest) | [SubmitExternalSignatureResponse](#axelar.bitcoin.v1beta1.SubmitExternalSignatureResponse) |  | POST|/axelar/bitcoin/submit_external_signature|

 <!-- end services -->



<a name="axelar/utils/v1beta1/queuer.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/utils/v1beta1/queuer.proto



<a name="axelar.utils.v1beta1.QueueState"></a>

### QueueState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `items` | [QueueState.ItemsEntry](#axelar.utils.v1beta1.QueueState.ItemsEntry) | repeated |  |






<a name="axelar.utils.v1beta1.QueueState.Item"></a>

### QueueState.Item



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [bytes](#bytes) |  |  |
| `value` | [bytes](#bytes) |  |  |






<a name="axelar.utils.v1beta1.QueueState.ItemsEntry"></a>

### QueueState.ItemsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [QueueState.Item](#axelar.utils.v1beta1.QueueState.Item) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/evm/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/evm/v1beta1/types.proto



<a name="axelar.evm.v1beta1.Asset"></a>

### Asset



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `name` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.BurnerInfo"></a>

### BurnerInfo
BurnerInfo describes information required to burn token at an burner address
that is deposited by an user


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `burner_address` | [bytes](#bytes) |  |  |
| `token_address` | [bytes](#bytes) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `asset` | [string](#string) |  |  |
| `salt` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.Command"></a>

### Command



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [bytes](#bytes) |  |  |
| `command` | [string](#string) |  |  |
| `params` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `max_gas_cost` | [uint32](#uint32) |  |  |






<a name="axelar.evm.v1beta1.CommandBatchMetadata"></a>

### CommandBatchMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [bytes](#bytes) |  |  |
| `command_ids` | [bytes](#bytes) | repeated |  |
| `data` | [bytes](#bytes) |  |  |
| `sig_hash` | [bytes](#bytes) |  |  |
| `status` | [BatchedCommandsStatus](#axelar.evm.v1beta1.BatchedCommandsStatus) |  |  |
| `key_id` | [string](#string) |  |  |
| `prev_batched_commands_id` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.ERC20Deposit"></a>

### ERC20Deposit
ERC20Deposit contains information for an ERC20 deposit


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `amount` | [bytes](#bytes) |  |  |
| `asset` | [string](#string) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `burner_address` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.ERC20TokenMetadata"></a>

### ERC20TokenMetadata
ERC20TokenMetadata describes information about an ERC20 token


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset` | [string](#string) |  |  |
| `chain_id` | [bytes](#bytes) |  |  |
| `details` | [TokenDetails](#axelar.evm.v1beta1.TokenDetails) |  |  |
| `token_address` | [string](#string) |  |  |
| `tx_hash` | [string](#string) |  |  |
| `status` | [Status](#axelar.evm.v1beta1.Status) |  |  |
| `is_external` | [bool](#bool) |  |  |
| `burner_code` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.Event"></a>

### Event



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `index` | [uint64](#uint64) |  |  |
| `status` | [Event.Status](#axelar.evm.v1beta1.Event.Status) |  |  |
| `token_sent` | [EventTokenSent](#axelar.evm.v1beta1.EventTokenSent) |  |  |
| `contract_call` | [EventContractCall](#axelar.evm.v1beta1.EventContractCall) |  |  |
| `contract_call_with_token` | [EventContractCallWithToken](#axelar.evm.v1beta1.EventContractCallWithToken) |  |  |
| `transfer` | [EventTransfer](#axelar.evm.v1beta1.EventTransfer) |  |  |
| `token_deployed` | [EventTokenDeployed](#axelar.evm.v1beta1.EventTokenDeployed) |  |  |
| `multisig_ownership_transferred` | [EventMultisigOwnershipTransferred](#axelar.evm.v1beta1.EventMultisigOwnershipTransferred) |  |  |
| `multisig_operatorship_transferred` | [EventMultisigOperatorshipTransferred](#axelar.evm.v1beta1.EventMultisigOperatorshipTransferred) |  |  |
| `singlesig_ownership_transferred` | [EventSinglesigOwnershipTransferred](#axelar.evm.v1beta1.EventSinglesigOwnershipTransferred) |  |  |
| `singlesig_operatorship_transferred` | [EventSinglesigOperatorshipTransferred](#axelar.evm.v1beta1.EventSinglesigOperatorshipTransferred) |  |  |






<a name="axelar.evm.v1beta1.EventContractCall"></a>

### EventContractCall



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `contract_address` | [string](#string) |  |  |
| `payload_hash` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.EventContractCallWithToken"></a>

### EventContractCallWithToken



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `contract_address` | [string](#string) |  |  |
| `payload_hash` | [bytes](#bytes) |  |  |
| `symbol` | [string](#string) |  |  |
| `amount` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.EventMultisigOperatorshipTransferred"></a>

### EventMultisigOperatorshipTransferred



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pre_operators` | [bytes](#bytes) | repeated |  |
| `prev_threshold` | [bytes](#bytes) |  |  |
| `new_operators` | [bytes](#bytes) | repeated |  |
| `new_threshold` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.EventMultisigOwnershipTransferred"></a>

### EventMultisigOwnershipTransferred



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pre_owners` | [bytes](#bytes) | repeated |  |
| `prev_threshold` | [bytes](#bytes) |  |  |
| `new_owners` | [bytes](#bytes) | repeated |  |
| `new_threshold` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.EventSinglesigOperatorshipTransferred"></a>

### EventSinglesigOperatorshipTransferred



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pre_operator` | [bytes](#bytes) |  |  |
| `new_operator` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.EventSinglesigOwnershipTransferred"></a>

### EventSinglesigOwnershipTransferred



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pre_owner` | [bytes](#bytes) |  |  |
| `new_owner` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.EventTokenDeployed"></a>

### EventTokenDeployed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `symbol` | [string](#string) |  |  |
| `token_address` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.EventTokenSent"></a>

### EventTokenSent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `destination_address` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `amount` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.EventTransfer"></a>

### EventTransfer



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to` | [bytes](#bytes) |  |  |
| `amount` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.Gateway"></a>

### Gateway



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [bytes](#bytes) |  |  |
| `status` | [Gateway.Status](#axelar.evm.v1beta1.Gateway.Status) |  | **Deprecated.**  |






<a name="axelar.evm.v1beta1.NetworkInfo"></a>

### NetworkInfo
NetworkInfo describes information about a network


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `id` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.SigMetadata"></a>

### SigMetadata
SigMetadata stores necessary information for external apps to map signature
results to evm relay transaction types


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `type` | [SigType](#axelar.evm.v1beta1.SigType) |  |  |
| `chain` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.TokenDetails"></a>

### TokenDetails



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `token_name` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `decimals` | [uint32](#uint32) |  |  |
| `capacity` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.TransactionMetadata"></a>

### TransactionMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `raw_tx` | [bytes](#bytes) |  |  |
| `pub_key` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.TransferKey"></a>

### TransferKey
TransferKey contains information for a transfer ownership or operatorship


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `type` | [TransferKeyType](#axelar.evm.v1beta1.TransferKeyType) |  |  |
| `next_key_id` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.VoteEvents"></a>

### VoteEvents



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `events` | [Event](#axelar.evm.v1beta1.Event) | repeated |  |





 <!-- end messages -->


<a name="axelar.evm.v1beta1.BatchedCommandsStatus"></a>

### BatchedCommandsStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| BATCHED_COMMANDS_STATUS_UNSPECIFIED | 0 |  |
| BATCHED_COMMANDS_STATUS_SIGNING | 1 |  |
| BATCHED_COMMANDS_STATUS_ABORTED | 2 |  |
| BATCHED_COMMANDS_STATUS_SIGNED | 3 |  |



<a name="axelar.evm.v1beta1.DepositStatus"></a>

### DepositStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| DEPOSIT_STATUS_UNSPECIFIED | 0 |  |
| DEPOSIT_STATUS_PENDING | 1 |  |
| DEPOSIT_STATUS_CONFIRMED | 2 |  |
| DEPOSIT_STATUS_BURNED | 3 |  |



<a name="axelar.evm.v1beta1.Event.Status"></a>

### Event.Status


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATUS_UNSPECIFIED | 0 |  |
| STATUS_CONFIRMED | 1 |  |
| STATUS_COMPLETED | 2 |  |
| STATUS_FAILED | 3 |  |



<a name="axelar.evm.v1beta1.Gateway.Status"></a>

### Gateway.Status


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATUS_UNSPECIFIED | 0 |  |
| STATUS_PENDING | 1 |  |
| STATUS_CONFIRMED | 2 |  |



<a name="axelar.evm.v1beta1.SigType"></a>

### SigType


| Name | Number | Description |
| ---- | ------ | ----------- |
| SIG_TYPE_UNSPECIFIED | 0 |  |
| SIG_TYPE_TX | 1 |  |
| SIG_TYPE_COMMAND | 2 |  |



<a name="axelar.evm.v1beta1.Status"></a>

### Status


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATUS_UNSPECIFIED | 0 | these enum values are used for bitwise operations, therefore they need to be powers of 2 |
| STATUS_INITIALIZED | 1 |  |
| STATUS_PENDING | 2 |  |
| STATUS_CONFIRMED | 4 |  |



<a name="axelar.evm.v1beta1.TransferKeyType"></a>

### TransferKeyType


| Name | Number | Description |
| ---- | ------ | ----------- |
| TRANSFER_KEY_TYPE_UNSPECIFIED | 0 |  |
| TRANSFER_KEY_TYPE_OWNERSHIP | 1 |  |
| TRANSFER_KEY_TYPE_OPERATORSHIP | 2 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/evm/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/evm/v1beta1/params.proto



<a name="axelar.evm.v1beta1.Params"></a>

### Params
Params is the parameter set for this module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `confirmation_height` | [uint64](#uint64) |  |  |
| `network` | [string](#string) |  |  |
| `token_code` | [bytes](#bytes) |  |  |
| `burnable` | [bytes](#bytes) |  |  |
| `revote_locking_period` | [int64](#int64) |  |  |
| `networks` | [NetworkInfo](#axelar.evm.v1beta1.NetworkInfo) | repeated |  |
| `voting_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `min_voter_count` | [int64](#int64) |  |  |
| `commands_gas_limit` | [uint32](#uint32) |  |  |






<a name="axelar.evm.v1beta1.PendingChain"></a>

### PendingChain



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelar.evm.v1beta1.Params) |  |  |
| `chain` | [axelar.nexus.exported.v1beta1.Chain](#axelar.nexus.exported.v1beta1.Chain) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/evm/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/evm/v1beta1/genesis.proto



<a name="axelar.evm.v1beta1.GenesisState"></a>

### GenesisState
GenesisState represents the genesis state


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chains` | [GenesisState.Chain](#axelar.evm.v1beta1.GenesisState.Chain) | repeated |  |






<a name="axelar.evm.v1beta1.GenesisState.Chain"></a>

### GenesisState.Chain



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelar.evm.v1beta1.Params) |  |  |
| `burner_infos` | [BurnerInfo](#axelar.evm.v1beta1.BurnerInfo) | repeated |  |
| `command_queue` | [axelar.utils.v1beta1.QueueState](#axelar.utils.v1beta1.QueueState) |  |  |
| `confirmed_deposits` | [ERC20Deposit](#axelar.evm.v1beta1.ERC20Deposit) | repeated |  |
| `burned_deposits` | [ERC20Deposit](#axelar.evm.v1beta1.ERC20Deposit) | repeated |  |
| `command_batches` | [CommandBatchMetadata](#axelar.evm.v1beta1.CommandBatchMetadata) | repeated |  |
| `gateway` | [Gateway](#axelar.evm.v1beta1.Gateway) |  |  |
| `tokens` | [ERC20TokenMetadata](#axelar.evm.v1beta1.ERC20TokenMetadata) | repeated |  |
| `events` | [Event](#axelar.evm.v1beta1.Event) | repeated |  |
| `confirmed_event_queue` | [axelar.utils.v1beta1.QueueState](#axelar.utils.v1beta1.QueueState) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/evm/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/evm/v1beta1/query.proto



<a name="axelar.evm.v1beta1.BatchedCommandsRequest"></a>

### BatchedCommandsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `id` | [string](#string) |  | id defines an optional id for the commandsbatch. If not specified the latest will be returned |






<a name="axelar.evm.v1beta1.BatchedCommandsResponse"></a>

### BatchedCommandsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `data` | [string](#string) |  |  |
| `status` | [BatchedCommandsStatus](#axelar.evm.v1beta1.BatchedCommandsStatus) |  |  |
| `key_id` | [string](#string) |  |  |
| `signature` | [string](#string) | repeated |  |
| `execute_data` | [string](#string) |  |  |
| `prev_batched_commands_id` | [string](#string) |  |  |
| `command_ids` | [string](#string) | repeated |  |






<a name="axelar.evm.v1beta1.BurnerInfoRequest"></a>

### BurnerInfoRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.BurnerInfoResponse"></a>

### BurnerInfoResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `burner_info` | [BurnerInfo](#axelar.evm.v1beta1.BurnerInfo) |  |  |






<a name="axelar.evm.v1beta1.BytecodeRequest"></a>

### BytecodeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `contract` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.BytecodeResponse"></a>

### BytecodeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `bytecode` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.ChainsRequest"></a>

### ChainsRequest







<a name="axelar.evm.v1beta1.ChainsResponse"></a>

### ChainsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chains` | [string](#string) | repeated |  |






<a name="axelar.evm.v1beta1.ConfirmationHeightRequest"></a>

### ConfirmationHeightRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.ConfirmationHeightResponse"></a>

### ConfirmationHeightResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `height` | [uint64](#uint64) |  |  |






<a name="axelar.evm.v1beta1.DepositQueryParams"></a>

### DepositQueryParams
DepositQueryParams describe the parameters used to query for an EVM
deposit address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `asset` | [string](#string) |  |  |
| `chain` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.DepositStateRequest"></a>

### DepositStateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `params` | [QueryDepositStateParams](#axelar.evm.v1beta1.QueryDepositStateParams) |  |  |






<a name="axelar.evm.v1beta1.DepositStateResponse"></a>

### DepositStateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `status` | [DepositStatus](#axelar.evm.v1beta1.DepositStatus) |  |  |






<a name="axelar.evm.v1beta1.EventRequest"></a>

### EventRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `event_id` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.EventResponse"></a>

### EventResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `event` | [Event](#axelar.evm.v1beta1.Event) |  |  |






<a name="axelar.evm.v1beta1.GatewayAddressRequest"></a>

### GatewayAddressRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.GatewayAddressResponse"></a>

### GatewayAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.KeyAddressRequest"></a>

### KeyAddressRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `role` | [int32](#int32) |  |  |
| `id` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.KeyAddressResponse"></a>

### KeyAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `multisig_addresses` | [KeyAddressResponse.MultisigAddresses](#axelar.evm.v1beta1.KeyAddressResponse.MultisigAddresses) |  |  |
| `threshold_address` | [KeyAddressResponse.ThresholdAddress](#axelar.evm.v1beta1.KeyAddressResponse.ThresholdAddress) |  |  |






<a name="axelar.evm.v1beta1.KeyAddressResponse.MultisigAddresses"></a>

### KeyAddressResponse.MultisigAddresses



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `addresses` | [string](#string) | repeated |  |
| `threshold` | [uint32](#uint32) |  |  |






<a name="axelar.evm.v1beta1.KeyAddressResponse.ThresholdAddress"></a>

### KeyAddressResponse.ThresholdAddress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.PendingCommandsRequest"></a>

### PendingCommandsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.PendingCommandsResponse"></a>

### PendingCommandsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `commands` | [QueryCommandResponse](#axelar.evm.v1beta1.QueryCommandResponse) | repeated |  |






<a name="axelar.evm.v1beta1.QueryBurnerAddressResponse"></a>

### QueryBurnerAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.QueryCommandResponse"></a>

### QueryCommandResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `type` | [string](#string) |  |  |
| `params` | [QueryCommandResponse.ParamsEntry](#axelar.evm.v1beta1.QueryCommandResponse.ParamsEntry) | repeated |  |
| `key_id` | [string](#string) |  |  |
| `max_gas_cost` | [uint32](#uint32) |  |  |






<a name="axelar.evm.v1beta1.QueryCommandResponse.ParamsEntry"></a>

### QueryCommandResponse.ParamsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.QueryDepositStateParams"></a>

### QueryDepositStateParams



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `burner_address` | [bytes](#bytes) |  |  |
| `amount` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.QueryTokenAddressResponse"></a>

### QueryTokenAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `confirmed` | [bool](#bool) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/evm/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/evm/v1beta1/tx.proto



<a name="axelar.evm.v1beta1.AddChainRequest"></a>

### AddChainRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `name` | [string](#string) |  |  |
| `key_type` | [axelar.tss.exported.v1beta1.KeyType](#axelar.tss.exported.v1beta1.KeyType) |  |  |
| `params` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.AddChainResponse"></a>

### AddChainResponse







<a name="axelar.evm.v1beta1.ConfirmDepositRequest"></a>

### ConfirmDepositRequest
MsgConfirmDeposit represents an erc20 deposit confirmation message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `amount` | [bytes](#bytes) |  | **Deprecated.**  |
| `burner_address` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.ConfirmDepositResponse"></a>

### ConfirmDepositResponse







<a name="axelar.evm.v1beta1.ConfirmGatewayTxRequest"></a>

### ConfirmGatewayTxRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.ConfirmGatewayTxResponse"></a>

### ConfirmGatewayTxResponse







<a name="axelar.evm.v1beta1.ConfirmTokenRequest"></a>

### ConfirmTokenRequest
MsgConfirmToken represents a token deploy confirmation message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `asset` | [Asset](#axelar.evm.v1beta1.Asset) |  |  |






<a name="axelar.evm.v1beta1.ConfirmTokenResponse"></a>

### ConfirmTokenResponse







<a name="axelar.evm.v1beta1.ConfirmTransferKeyRequest"></a>

### ConfirmTransferKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `transfer_type` | [TransferKeyType](#axelar.evm.v1beta1.TransferKeyType) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.ConfirmTransferKeyResponse"></a>

### ConfirmTransferKeyResponse







<a name="axelar.evm.v1beta1.CreateBurnTokensRequest"></a>

### CreateBurnTokensRequest
CreateBurnTokensRequest represents the message to create commands to burn
tokens with AxelarGateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.CreateBurnTokensResponse"></a>

### CreateBurnTokensResponse







<a name="axelar.evm.v1beta1.CreateDeployTokenRequest"></a>

### CreateDeployTokenRequest
CreateDeployTokenRequest represents the message to create a deploy token
command for AxelarGateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `asset` | [Asset](#axelar.evm.v1beta1.Asset) |  |  |
| `token_details` | [TokenDetails](#axelar.evm.v1beta1.TokenDetails) |  |  |
| `address` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.CreateDeployTokenResponse"></a>

### CreateDeployTokenResponse







<a name="axelar.evm.v1beta1.CreatePendingTransfersRequest"></a>

### CreatePendingTransfersRequest
CreatePendingTransfersRequest represents a message to trigger the creation of
commands handling all pending transfers


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.CreatePendingTransfersResponse"></a>

### CreatePendingTransfersResponse







<a name="axelar.evm.v1beta1.CreateTransferOperatorshipRequest"></a>

### CreateTransferOperatorshipRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.CreateTransferOperatorshipResponse"></a>

### CreateTransferOperatorshipResponse







<a name="axelar.evm.v1beta1.CreateTransferOwnershipRequest"></a>

### CreateTransferOwnershipRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.CreateTransferOwnershipResponse"></a>

### CreateTransferOwnershipResponse







<a name="axelar.evm.v1beta1.LinkRequest"></a>

### LinkRequest
MsgLink represents the message that links a cross chain address to a burner
address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `recipient_addr` | [string](#string) |  |  |
| `asset` | [string](#string) |  |  |
| `recipient_chain` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.LinkResponse"></a>

### LinkResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposit_addr` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.RetryFailedEventRequest"></a>

### RetryFailedEventRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `event_id` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.RetryFailedEventResponse"></a>

### RetryFailedEventResponse







<a name="axelar.evm.v1beta1.SetGatewayRequest"></a>

### SetGatewayRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `address` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.SetGatewayResponse"></a>

### SetGatewayResponse







<a name="axelar.evm.v1beta1.SignCommandsRequest"></a>

### SignCommandsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.SignCommandsResponse"></a>

### SignCommandsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `batched_commands_id` | [bytes](#bytes) |  |  |
| `command_count` | [uint32](#uint32) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/evm/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/evm/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelar.evm.v1beta1.MsgService"></a>

### MsgService
Msg defines the evm Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `SetGateway` | [SetGatewayRequest](#axelar.evm.v1beta1.SetGatewayRequest) | [SetGatewayResponse](#axelar.evm.v1beta1.SetGatewayResponse) |  | POST|/axelar/evm/set_gateway|
| `ConfirmGatewayTx` | [ConfirmGatewayTxRequest](#axelar.evm.v1beta1.ConfirmGatewayTxRequest) | [ConfirmGatewayTxResponse](#axelar.evm.v1beta1.ConfirmGatewayTxResponse) |  | POST|/axelar/evm/confirm_gateway_tx|
| `Link` | [LinkRequest](#axelar.evm.v1beta1.LinkRequest) | [LinkResponse](#axelar.evm.v1beta1.LinkResponse) |  | POST|/axelar/evm/link|
| `ConfirmToken` | [ConfirmTokenRequest](#axelar.evm.v1beta1.ConfirmTokenRequest) | [ConfirmTokenResponse](#axelar.evm.v1beta1.ConfirmTokenResponse) |  | POST|/axelar/evm/confirm_token|
| `ConfirmDeposit` | [ConfirmDepositRequest](#axelar.evm.v1beta1.ConfirmDepositRequest) | [ConfirmDepositResponse](#axelar.evm.v1beta1.ConfirmDepositResponse) |  | POST|/axelar/evm/confirm_deposit|
| `ConfirmTransferKey` | [ConfirmTransferKeyRequest](#axelar.evm.v1beta1.ConfirmTransferKeyRequest) | [ConfirmTransferKeyResponse](#axelar.evm.v1beta1.ConfirmTransferKeyResponse) |  | POST|/axelar/evm/confirm_transfer_key|
| `CreateDeployToken` | [CreateDeployTokenRequest](#axelar.evm.v1beta1.CreateDeployTokenRequest) | [CreateDeployTokenResponse](#axelar.evm.v1beta1.CreateDeployTokenResponse) |  | POST|/axelar/evm/create_deploy_token|
| `CreateBurnTokens` | [CreateBurnTokensRequest](#axelar.evm.v1beta1.CreateBurnTokensRequest) | [CreateBurnTokensResponse](#axelar.evm.v1beta1.CreateBurnTokensResponse) |  | POST|/axelar/evm/create_burn_tokens|
| `CreatePendingTransfers` | [CreatePendingTransfersRequest](#axelar.evm.v1beta1.CreatePendingTransfersRequest) | [CreatePendingTransfersResponse](#axelar.evm.v1beta1.CreatePendingTransfersResponse) |  | POST|/axelar/evm/create_pending_transfers|
| `CreateTransferOwnership` | [CreateTransferOwnershipRequest](#axelar.evm.v1beta1.CreateTransferOwnershipRequest) | [CreateTransferOwnershipResponse](#axelar.evm.v1beta1.CreateTransferOwnershipResponse) |  | POST|/axelar/evm/create_transfer_ownership|
| `CreateTransferOperatorship` | [CreateTransferOperatorshipRequest](#axelar.evm.v1beta1.CreateTransferOperatorshipRequest) | [CreateTransferOperatorshipResponse](#axelar.evm.v1beta1.CreateTransferOperatorshipResponse) |  | POST|/axelar/evm/create_transfer_operatorship|
| `SignCommands` | [SignCommandsRequest](#axelar.evm.v1beta1.SignCommandsRequest) | [SignCommandsResponse](#axelar.evm.v1beta1.SignCommandsResponse) |  | POST|/axelar/evm/sign_commands|
| `AddChain` | [AddChainRequest](#axelar.evm.v1beta1.AddChainRequest) | [AddChainResponse](#axelar.evm.v1beta1.AddChainResponse) |  | POST|/axelar/evm/add_chain|
| `RetryFailedEvent` | [RetryFailedEventRequest](#axelar.evm.v1beta1.RetryFailedEventRequest) | [RetryFailedEventResponse](#axelar.evm.v1beta1.RetryFailedEventResponse) |  | POST|/axelar/evm/retry-failed-event|


<a name="axelar.evm.v1beta1.QueryService"></a>

### QueryService
QueryService defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `BatchedCommands` | [BatchedCommandsRequest](#axelar.evm.v1beta1.BatchedCommandsRequest) | [BatchedCommandsResponse](#axelar.evm.v1beta1.BatchedCommandsResponse) | BatchedCommands queries the batched commands for a specified chain and BatchedCommandsID if no BatchedCommandsID is specified, then it returns the latest batched commands | GET|/axelar/evm/v1beta1/batched_commands/{chain}/{id}|
| `BurnerInfo` | [BurnerInfoRequest](#axelar.evm.v1beta1.BurnerInfoRequest) | [BurnerInfoResponse](#axelar.evm.v1beta1.BurnerInfoResponse) | BurnerInfo queries the burner info for the specified address | GET|/axelar/evm/v1beta1/burner_info|
| `ConfirmationHeight` | [ConfirmationHeightRequest](#axelar.evm.v1beta1.ConfirmationHeightRequest) | [ConfirmationHeightResponse](#axelar.evm.v1beta1.ConfirmationHeightResponse) | ConfirmationHeight queries the confirmation height for the specified chain | GET|/axelar/evm/v1beta1/confirmation_height/{chain}|
| `DepositState` | [DepositStateRequest](#axelar.evm.v1beta1.DepositStateRequest) | [DepositStateResponse](#axelar.evm.v1beta1.DepositStateResponse) | DepositState queries the state of the specified deposit | GET|/axelar/evm/v1beta1/deposit_state|
| `PendingCommands` | [PendingCommandsRequest](#axelar.evm.v1beta1.PendingCommandsRequest) | [PendingCommandsResponse](#axelar.evm.v1beta1.PendingCommandsResponse) | PendingCommands queries the pending commands for the specified chain | GET|/axelar/evm/v1beta1/pending_commands/{chain}|
| `Chains` | [ChainsRequest](#axelar.evm.v1beta1.ChainsRequest) | [ChainsResponse](#axelar.evm.v1beta1.ChainsResponse) | Chains queries the available evm chains | GET|/axelar/evm/v1beta1/chains|
| `KeyAddress` | [KeyAddressRequest](#axelar.evm.v1beta1.KeyAddressRequest) | [KeyAddressResponse](#axelar.evm.v1beta1.KeyAddressResponse) | KeyAddress queries the address of key of a chain | GET|/axelar/evm/v1beta1/key_address/{chain}|
| `GatewayAddress` | [GatewayAddressRequest](#axelar.evm.v1beta1.GatewayAddressRequest) | [GatewayAddressResponse](#axelar.evm.v1beta1.GatewayAddressResponse) | GatewayAddress queries the address of axelar gateway at the specified chain | GET|/axelar/evm/v1beta1/gateway_address/{chain}|
| `Bytecode` | [BytecodeRequest](#axelar.evm.v1beta1.BytecodeRequest) | [BytecodeResponse](#axelar.evm.v1beta1.BytecodeResponse) | Bytecode queries the bytecode of a specified gateway at the specified chain | GET|/axelar/evm/v1beta1/bytecode/{chain}/{contract}|
| `Event` | [EventRequest](#axelar.evm.v1beta1.EventRequest) | [EventResponse](#axelar.evm.v1beta1.EventResponse) | Event queries an event at the specified chain | GET|/axelar/evm/v1beta1/event/{chain}/{event_id}|

 <!-- end services -->



<a name="axelar/nexus/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/nexus/v1beta1/params.proto



<a name="axelar.nexus.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain_activation_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `chain_maintainer_missing_vote_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `chain_maintainer_incorrect_vote_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `chain_maintainer_check_window` | [int32](#int32) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/nexus/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/nexus/v1beta1/genesis.proto



<a name="axelar.nexus.v1beta1.GenesisState"></a>

### GenesisState
GenesisState represents the genesis state


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelar.nexus.v1beta1.Params) |  |  |
| `nonce` | [uint64](#uint64) |  |  |
| `chains` | [axelar.nexus.exported.v1beta1.Chain](#axelar.nexus.exported.v1beta1.Chain) | repeated |  |
| `chain_states` | [ChainState](#axelar.nexus.v1beta1.ChainState) | repeated |  |
| `linked_addresses` | [LinkedAddresses](#axelar.nexus.v1beta1.LinkedAddresses) | repeated |  |
| `transfers` | [axelar.nexus.exported.v1beta1.CrossChainTransfer](#axelar.nexus.exported.v1beta1.CrossChainTransfer) | repeated |  |
| `fee` | [axelar.nexus.exported.v1beta1.TransferFee](#axelar.nexus.exported.v1beta1.TransferFee) |  |  |
| `fee_infos` | [axelar.nexus.exported.v1beta1.FeeInfo](#axelar.nexus.exported.v1beta1.FeeInfo) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/nexus/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/nexus/v1beta1/tx.proto



<a name="axelar.nexus.v1beta1.ActivateChainRequest"></a>

### ActivateChainRequest
ActivateChainRequest represents a message to activate chains


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chains` | [string](#string) | repeated |  |






<a name="axelar.nexus.v1beta1.ActivateChainResponse"></a>

### ActivateChainResponse







<a name="axelar.nexus.v1beta1.DeactivateChainRequest"></a>

### DeactivateChainRequest
DeactivateChainRequest represents a message to deactivate chains


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chains` | [string](#string) | repeated |  |






<a name="axelar.nexus.v1beta1.DeactivateChainResponse"></a>

### DeactivateChainResponse







<a name="axelar.nexus.v1beta1.DeregisterChainMaintainerRequest"></a>

### DeregisterChainMaintainerRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chains` | [string](#string) | repeated |  |






<a name="axelar.nexus.v1beta1.DeregisterChainMaintainerResponse"></a>

### DeregisterChainMaintainerResponse







<a name="axelar.nexus.v1beta1.RegisterAssetFeeRequest"></a>

### RegisterAssetFeeRequest
RegisterAssetFeeRequest represents a message to register the transfer fee
info associated to an asset on a chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `fee_info` | [axelar.nexus.exported.v1beta1.FeeInfo](#axelar.nexus.exported.v1beta1.FeeInfo) |  |  |






<a name="axelar.nexus.v1beta1.RegisterAssetFeeResponse"></a>

### RegisterAssetFeeResponse







<a name="axelar.nexus.v1beta1.RegisterChainMaintainerRequest"></a>

### RegisterChainMaintainerRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chains` | [string](#string) | repeated |  |






<a name="axelar.nexus.v1beta1.RegisterChainMaintainerResponse"></a>

### RegisterChainMaintainerResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/nexus/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/nexus/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelar.nexus.v1beta1.MsgService"></a>

### MsgService
Msg defines the nexus Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterChainMaintainer` | [RegisterChainMaintainerRequest](#axelar.nexus.v1beta1.RegisterChainMaintainerRequest) | [RegisterChainMaintainerResponse](#axelar.nexus.v1beta1.RegisterChainMaintainerResponse) |  | POST|/axelar/nexus/register_chain_maintainer|
| `DeregisterChainMaintainer` | [DeregisterChainMaintainerRequest](#axelar.nexus.v1beta1.DeregisterChainMaintainerRequest) | [DeregisterChainMaintainerResponse](#axelar.nexus.v1beta1.DeregisterChainMaintainerResponse) |  | POST|/axelar/nexus/deregister_chain_maintainer|
| `ActivateChain` | [ActivateChainRequest](#axelar.nexus.v1beta1.ActivateChainRequest) | [ActivateChainResponse](#axelar.nexus.v1beta1.ActivateChainResponse) |  | POST|/axelar/nexus/activate_chain|
| `DeactivateChain` | [DeactivateChainRequest](#axelar.nexus.v1beta1.DeactivateChainRequest) | [DeactivateChainResponse](#axelar.nexus.v1beta1.DeactivateChainResponse) |  | POST|/axelar/nexus/deactivate_chain|
| `RegisterAssetFee` | [RegisterAssetFeeRequest](#axelar.nexus.v1beta1.RegisterAssetFeeRequest) | [RegisterAssetFeeResponse](#axelar.nexus.v1beta1.RegisterAssetFeeResponse) |  | POST|/axelar/nexus/register_asset_fee|


<a name="axelar.nexus.v1beta1.QueryService"></a>

### QueryService
QueryService defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `LatestDepositAddress` | [LatestDepositAddressRequest](#axelar.nexus.v1beta1.LatestDepositAddressRequest) | [LatestDepositAddressResponse](#axelar.nexus.v1beta1.LatestDepositAddressResponse) | LatestDepositAddress queries the a deposit address by recipient | GET|/axelar/nexus/v1beta1/latest_deposit_address/{recipient_addr}/{recipient_chain}/{deposit_chain}|
| `TransfersForChain` | [TransfersForChainRequest](#axelar.nexus.v1beta1.TransfersForChainRequest) | [TransfersForChainResponse](#axelar.nexus.v1beta1.TransfersForChainResponse) | TransfersForChain queries transfers by chain | GET|/axelar/nexus/v1beta1/transfers_for_chain/{chain}/{state}|
| `FeeInfo` | [FeeInfoRequest](#axelar.nexus.v1beta1.FeeInfoRequest) | [FeeInfoResponse](#axelar.nexus.v1beta1.FeeInfoResponse) | FeeInfo queries the fee info by chain and asset | GET|/axelar/nexus/v1beta1/fee_info/{chain}/{asset}GET|/axelar/nexus/v1beta1/fee|
| `TransferFee` | [TransferFeeRequest](#axelar.nexus.v1beta1.TransferFeeRequest) | [TransferFeeResponse](#axelar.nexus.v1beta1.TransferFeeResponse) | TransferFee queries the transfer fee by the source, destination chain, and amount. If amount is 0, the min fee is returned | GET|/axelar/nexus/v1beta1/transfer_fee/{source_chain}/{destination_chain}/{amount}GET|/axelar/nexus/v1beta1/transfer_fee|
| `Chains` | [ChainsRequest](#axelar.nexus.v1beta1.ChainsRequest) | [ChainsResponse](#axelar.nexus.v1beta1.ChainsResponse) | Chains queries the chains registered on the network | GET|/axelar/nexus/v1beta1/chains|
| `Assets` | [AssetsRequest](#axelar.nexus.v1beta1.AssetsRequest) | [AssetsResponse](#axelar.nexus.v1beta1.AssetsResponse) | Assets queries the assets registered for a chain | GET|/axelar/nexus/v1beta1/assets/{chain}|
| `ChainState` | [ChainStateRequest](#axelar.nexus.v1beta1.ChainStateRequest) | [ChainStateResponse](#axelar.nexus.v1beta1.ChainStateResponse) | ChainState queries the state of a registered chain on the network | GET|/axelar/nexus/v1beta1/chain_state/{chain}|
| `ChainsByAsset` | [ChainsByAssetRequest](#axelar.nexus.v1beta1.ChainsByAssetRequest) | [ChainsByAssetResponse](#axelar.nexus.v1beta1.ChainsByAssetResponse) | ChainsByAsset queries the chains that support an asset on the network | GET|/axelar/nexus/v1beta1/chains_by_asset/{asset}|

 <!-- end services -->



<a name="axelar/permission/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/permission/v1beta1/types.proto



<a name="axelar.permission.v1beta1.GovAccount"></a>

### GovAccount



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [bytes](#bytes) |  |  |
| `role` | [axelar.permission.exported.v1beta1.Role](#axelar.permission.exported.v1beta1.Role) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/permission/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/permission/v1beta1/params.proto



<a name="axelar.permission.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/permission/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/permission/v1beta1/genesis.proto



<a name="axelar.permission.v1beta1.GenesisState"></a>

### GenesisState
GenesisState represents the genesis state


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelar.permission.v1beta1.Params) |  |  |
| `governance_key` | [cosmos.crypto.multisig.LegacyAminoPubKey](#cosmos.crypto.multisig.LegacyAminoPubKey) |  |  |
| `gov_accounts` | [GovAccount](#axelar.permission.v1beta1.GovAccount) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/permission/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/permission/v1beta1/query.proto



<a name="axelar.permission.v1beta1.QueryGovernanceKeyRequest"></a>

### QueryGovernanceKeyRequest
QueryGovernanceKeyRequest is the request type for the
Query/GovernanceKey RPC method






<a name="axelar.permission.v1beta1.QueryGovernanceKeyResponse"></a>

### QueryGovernanceKeyResponse
QueryGovernanceKeyResponse is the response type for the
Query/GovernanceKey RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `governance_key` | [cosmos.crypto.multisig.LegacyAminoPubKey](#cosmos.crypto.multisig.LegacyAminoPubKey) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/permission/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/permission/v1beta1/tx.proto



<a name="axelar.permission.v1beta1.DeregisterControllerRequest"></a>

### DeregisterControllerRequest
DeregisterController represents a message to deregister a controller account


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `controller` | [bytes](#bytes) |  |  |






<a name="axelar.permission.v1beta1.DeregisterControllerResponse"></a>

### DeregisterControllerResponse







<a name="axelar.permission.v1beta1.RegisterControllerRequest"></a>

### RegisterControllerRequest
MsgRegisterController represents a message to register a controller account


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `controller` | [bytes](#bytes) |  |  |






<a name="axelar.permission.v1beta1.RegisterControllerResponse"></a>

### RegisterControllerResponse







<a name="axelar.permission.v1beta1.UpdateGovernanceKeyRequest"></a>

### UpdateGovernanceKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `governance_key` | [cosmos.crypto.multisig.LegacyAminoPubKey](#cosmos.crypto.multisig.LegacyAminoPubKey) |  |  |






<a name="axelar.permission.v1beta1.UpdateGovernanceKeyResponse"></a>

### UpdateGovernanceKeyResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/permission/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/permission/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelar.permission.v1beta1.Msg"></a>

### Msg
Msg defines the gov Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterController` | [RegisterControllerRequest](#axelar.permission.v1beta1.RegisterControllerRequest) | [RegisterControllerResponse](#axelar.permission.v1beta1.RegisterControllerResponse) |  | POST|/axelar/permission/register_controller|
| `DeregisterController` | [DeregisterControllerRequest](#axelar.permission.v1beta1.DeregisterControllerRequest) | [DeregisterControllerResponse](#axelar.permission.v1beta1.DeregisterControllerResponse) |  | POST|/axelar/permission/deregister_controller|
| `UpdateGovernanceKey` | [UpdateGovernanceKeyRequest](#axelar.permission.v1beta1.UpdateGovernanceKeyRequest) | [UpdateGovernanceKeyResponse](#axelar.permission.v1beta1.UpdateGovernanceKeyResponse) |  | POST|/axelar/permission/update_governance_key|


<a name="axelar.permission.v1beta1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `GovernanceKey` | [QueryGovernanceKeyRequest](#axelar.permission.v1beta1.QueryGovernanceKeyRequest) | [QueryGovernanceKeyResponse](#axelar.permission.v1beta1.QueryGovernanceKeyResponse) | GovernanceKey returns the multisig governance key | GET|/axelar/permission/v1beta1/governance_key|

 <!-- end services -->



<a name="axelar/reward/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/reward/v1beta1/params.proto



<a name="axelar.reward.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `external_chain_voting_inflation_rate` | [bytes](#bytes) |  |  |
| `tss_relative_inflation_rate` | [bytes](#bytes) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/reward/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/reward/v1beta1/types.proto



<a name="axelar.reward.v1beta1.Pool"></a>

### Pool



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `rewards` | [Pool.Reward](#axelar.reward.v1beta1.Pool.Reward) | repeated |  |






<a name="axelar.reward.v1beta1.Pool.Reward"></a>

### Pool.Reward



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validator` | [bytes](#bytes) |  |  |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="axelar.reward.v1beta1.Refund"></a>

### Refund



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `payer` | [bytes](#bytes) |  |  |
| `fees` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/reward/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/reward/v1beta1/genesis.proto



<a name="axelar.reward.v1beta1.GenesisState"></a>

### GenesisState
GenesisState represents the genesis state


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelar.reward.v1beta1.Params) |  |  |
| `pools` | [Pool](#axelar.reward.v1beta1.Pool) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/reward/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/reward/v1beta1/tx.proto



<a name="axelar.reward.v1beta1.RefundMsgRequest"></a>

### RefundMsgRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `inner_message` | [google.protobuf.Any](#google.protobuf.Any) |  |  |






<a name="axelar.reward.v1beta1.RefundMsgResponse"></a>

### RefundMsgResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  |  |
| `log` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/reward/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/reward/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelar.reward.v1beta1.MsgService"></a>

### MsgService
Msg defines the axelarnet Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RefundMsg` | [RefundMsgRequest](#axelar.reward.v1beta1.RefundMsgRequest) | [RefundMsgResponse](#axelar.reward.v1beta1.RefundMsgResponse) |  | POST|/axelar/reward/refund_message|

 <!-- end services -->



<a name="axelar/snapshot/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/snapshot/v1beta1/params.proto



<a name="axelar.snapshot.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `min_proxy_balance` | [int64](#int64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/snapshot/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/snapshot/v1beta1/types.proto



<a name="axelar.snapshot.v1beta1.ProxiedValidator"></a>

### ProxiedValidator



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validator` | [bytes](#bytes) |  |  |
| `proxy` | [bytes](#bytes) |  |  |
| `active` | [bool](#bool) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/snapshot/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/snapshot/v1beta1/genesis.proto



<a name="axelar.snapshot.v1beta1.GenesisState"></a>

### GenesisState
GenesisState represents the genesis state


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelar.snapshot.v1beta1.Params) |  |  |
| `snapshots` | [axelar.snapshot.exported.v1beta1.Snapshot](#axelar.snapshot.exported.v1beta1.Snapshot) | repeated |  |
| `proxied_validators` | [ProxiedValidator](#axelar.snapshot.v1beta1.ProxiedValidator) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/snapshot/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/snapshot/v1beta1/query.proto



<a name="axelar.snapshot.v1beta1.QueryValidatorsResponse"></a>

### QueryValidatorsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validators` | [QueryValidatorsResponse.Validator](#axelar.snapshot.v1beta1.QueryValidatorsResponse.Validator) | repeated |  |






<a name="axelar.snapshot.v1beta1.QueryValidatorsResponse.TssIllegibilityInfo"></a>

### QueryValidatorsResponse.TssIllegibilityInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tombstoned` | [bool](#bool) |  |  |
| `jailed` | [bool](#bool) |  |  |
| `missed_too_many_blocks` | [bool](#bool) |  |  |
| `no_proxy_registered` | [bool](#bool) |  |  |
| `tss_suspended` | [bool](#bool) |  |  |
| `proxy_insuficient_funds` | [bool](#bool) |  |  |
| `stale_tss_heartbeat` | [bool](#bool) |  |  |






<a name="axelar.snapshot.v1beta1.QueryValidatorsResponse.Validator"></a>

### QueryValidatorsResponse.Validator



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `operator_address` | [string](#string) |  |  |
| `moniker` | [string](#string) |  |  |
| `tss_illegibility_info` | [QueryValidatorsResponse.TssIllegibilityInfo](#axelar.snapshot.v1beta1.QueryValidatorsResponse.TssIllegibilityInfo) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/snapshot/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/snapshot/v1beta1/tx.proto



<a name="axelar.snapshot.v1beta1.DeactivateProxyRequest"></a>

### DeactivateProxyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |






<a name="axelar.snapshot.v1beta1.DeactivateProxyResponse"></a>

### DeactivateProxyResponse







<a name="axelar.snapshot.v1beta1.RegisterProxyRequest"></a>

### RegisterProxyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `proxy_addr` | [bytes](#bytes) |  |  |






<a name="axelar.snapshot.v1beta1.RegisterProxyResponse"></a>

### RegisterProxyResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/snapshot/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/snapshot/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelar.snapshot.v1beta1.MsgService"></a>

### MsgService
Msg defines the snapshot Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterProxy` | [RegisterProxyRequest](#axelar.snapshot.v1beta1.RegisterProxyRequest) | [RegisterProxyResponse](#axelar.snapshot.v1beta1.RegisterProxyResponse) | RegisterProxy defines a method for registering a proxy account that can act in a validator account's stead. | POST|/axelar/snapshot/register_proxy|
| `DeactivateProxy` | [DeactivateProxyRequest](#axelar.snapshot.v1beta1.DeactivateProxyRequest) | [DeactivateProxyResponse](#axelar.snapshot.v1beta1.DeactivateProxyResponse) | DeactivateProxy defines a method for deregistering a proxy account. | POST|/axelar/snapshot/deactivate_proxy|

 <!-- end services -->



<a name="axelar/tss/tofnd/v1beta1/common.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/tss/tofnd/v1beta1/common.proto
File copied from golang tofnd with minor tweaks


<a name="axelar.tss.tofnd.v1beta1.KeyPresenceRequest"></a>

### KeyPresenceRequest
Key presence check types


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_uid` | [string](#string) |  |  |
| `pub_key` | [bytes](#bytes) |  | SEC1-encoded compressed pub key bytes to find the right |






<a name="axelar.tss.tofnd.v1beta1.KeyPresenceResponse"></a>

### KeyPresenceResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `response` | [KeyPresenceResponse.Response](#axelar.tss.tofnd.v1beta1.KeyPresenceResponse.Response) |  |  |





 <!-- end messages -->


<a name="axelar.tss.tofnd.v1beta1.KeyPresenceResponse.Response"></a>

### KeyPresenceResponse.Response


| Name | Number | Description |
| ---- | ------ | ----------- |
| RESPONSE_UNSPECIFIED | 0 |  |
| RESPONSE_PRESENT | 1 |  |
| RESPONSE_ABSENT | 2 |  |
| RESPONSE_FAIL | 3 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/tss/tofnd/v1beta1/multisig.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/tss/tofnd/v1beta1/multisig.proto
File copied from golang tofnd with minor tweaks


<a name="axelar.tss.tofnd.v1beta1.KeygenRequest"></a>

### KeygenRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_uid` | [string](#string) |  |  |
| `party_uid` | [string](#string) |  | used only for logging |






<a name="axelar.tss.tofnd.v1beta1.KeygenResponse"></a>

### KeygenResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pub_key` | [bytes](#bytes) |  | SEC1-encoded compressed curve point |
| `error` | [string](#string) |  | reply with an error message if keygen fails |






<a name="axelar.tss.tofnd.v1beta1.SignRequest"></a>

### SignRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_uid` | [string](#string) |  |  |
| `msg_to_sign` | [bytes](#bytes) |  | 32-byte pre-hashed message digest |
| `party_uid` | [string](#string) |  | used only for logging |
| `pub_key` | [bytes](#bytes) |  | SEC1-encoded compressed pub key bytes to find the right |






<a name="axelar.tss.tofnd.v1beta1.SignResponse"></a>

### SignResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signature` | [bytes](#bytes) |  | ASN.1 DER-encoded ECDSA signature |
| `error` | [string](#string) |  | reply with an error message if sign fails |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/tss/tofnd/v1beta1/tofnd.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/tss/tofnd/v1beta1/tofnd.proto
File copied from golang tofnd with minor tweaks


<a name="axelar.tss.tofnd.v1beta1.KeygenInit"></a>

### KeygenInit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `new_key_uid` | [string](#string) |  |  |
| `party_uids` | [string](#string) | repeated |  |
| `party_share_counts` | [uint32](#uint32) | repeated |  |
| `my_party_index` | [uint32](#uint32) |  | parties[my_party_index] belongs to the server |
| `threshold` | [uint32](#uint32) |  |  |






<a name="axelar.tss.tofnd.v1beta1.KeygenOutput"></a>

### KeygenOutput
Keygen's success response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pub_key` | [bytes](#bytes) |  | pub_key; common for all parties |
| `group_recover_info` | [bytes](#bytes) |  | recover info of all parties' shares; common for all parties |
| `private_recover_info` | [bytes](#bytes) |  | private recover info of this party's shares; unique for each party |






<a name="axelar.tss.tofnd.v1beta1.MessageIn"></a>

### MessageIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `keygen_init` | [KeygenInit](#axelar.tss.tofnd.v1beta1.KeygenInit) |  | first message only, Keygen |
| `sign_init` | [SignInit](#axelar.tss.tofnd.v1beta1.SignInit) |  | first message only, Sign |
| `traffic` | [TrafficIn](#axelar.tss.tofnd.v1beta1.TrafficIn) |  | all subsequent messages |
| `abort` | [bool](#bool) |  | abort the protocol, ignore the bool value |






<a name="axelar.tss.tofnd.v1beta1.MessageOut"></a>

### MessageOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `traffic` | [TrafficOut](#axelar.tss.tofnd.v1beta1.TrafficOut) |  | all but final message |
| `keygen_result` | [MessageOut.KeygenResult](#axelar.tss.tofnd.v1beta1.MessageOut.KeygenResult) |  | final message only, Keygen |
| `sign_result` | [MessageOut.SignResult](#axelar.tss.tofnd.v1beta1.MessageOut.SignResult) |  | final message only, Sign |
| `need_recover` | [bool](#bool) |  | issue recover from client |






<a name="axelar.tss.tofnd.v1beta1.MessageOut.CriminalList"></a>

### MessageOut.CriminalList
Keygen/Sign failure response message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `criminals` | [MessageOut.CriminalList.Criminal](#axelar.tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal) | repeated |  |






<a name="axelar.tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal"></a>

### MessageOut.CriminalList.Criminal



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `party_uid` | [string](#string) |  |  |
| `crime_type` | [MessageOut.CriminalList.Criminal.CrimeType](#axelar.tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal.CrimeType) |  |  |






<a name="axelar.tss.tofnd.v1beta1.MessageOut.KeygenResult"></a>

### MessageOut.KeygenResult
Keygen's response types


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [KeygenOutput](#axelar.tss.tofnd.v1beta1.KeygenOutput) |  | Success response |
| `criminals` | [MessageOut.CriminalList](#axelar.tss.tofnd.v1beta1.MessageOut.CriminalList) |  | Faiilure response |






<a name="axelar.tss.tofnd.v1beta1.MessageOut.SignResult"></a>

### MessageOut.SignResult
Sign's response types


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signature` | [bytes](#bytes) |  | Success response |
| `criminals` | [MessageOut.CriminalList](#axelar.tss.tofnd.v1beta1.MessageOut.CriminalList) |  | Failure response |






<a name="axelar.tss.tofnd.v1beta1.RecoverRequest"></a>

### RecoverRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `keygen_init` | [KeygenInit](#axelar.tss.tofnd.v1beta1.KeygenInit) |  |  |
| `keygen_output` | [KeygenOutput](#axelar.tss.tofnd.v1beta1.KeygenOutput) |  |  |






<a name="axelar.tss.tofnd.v1beta1.RecoverResponse"></a>

### RecoverResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `response` | [RecoverResponse.Response](#axelar.tss.tofnd.v1beta1.RecoverResponse.Response) |  |  |






<a name="axelar.tss.tofnd.v1beta1.SignInit"></a>

### SignInit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `new_sig_uid` | [string](#string) |  |  |
| `key_uid` | [string](#string) |  |  |
| `party_uids` | [string](#string) | repeated | TODO replace this with a subset of indices? |
| `message_to_sign` | [bytes](#bytes) |  |  |






<a name="axelar.tss.tofnd.v1beta1.TrafficIn"></a>

### TrafficIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `from_party_uid` | [string](#string) |  |  |
| `payload` | [bytes](#bytes) |  |  |
| `is_broadcast` | [bool](#bool) |  |  |






<a name="axelar.tss.tofnd.v1beta1.TrafficOut"></a>

### TrafficOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to_party_uid` | [string](#string) |  |  |
| `payload` | [bytes](#bytes) |  |  |
| `is_broadcast` | [bool](#bool) |  |  |





 <!-- end messages -->


<a name="axelar.tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal.CrimeType"></a>

### MessageOut.CriminalList.Criminal.CrimeType


| Name | Number | Description |
| ---- | ------ | ----------- |
| CRIME_TYPE_UNSPECIFIED | 0 |  |
| CRIME_TYPE_NON_MALICIOUS | 1 |  |
| CRIME_TYPE_MALICIOUS | 2 |  |



<a name="axelar.tss.tofnd.v1beta1.RecoverResponse.Response"></a>

### RecoverResponse.Response


| Name | Number | Description |
| ---- | ------ | ----------- |
| RESPONSE_UNSPECIFIED | 0 |  |
| RESPONSE_SUCCESS | 1 |  |
| RESPONSE_FAIL | 2 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/tss/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/tss/v1beta1/params.proto



<a name="axelar.tss.v1beta1.Params"></a>

### Params
Params is the parameter set for this module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_requirements` | [axelar.tss.exported.v1beta1.KeyRequirement](#axelar.tss.exported.v1beta1.KeyRequirement) | repeated | KeyRequirements defines the requirement for each key role |
| `suspend_duration_in_blocks` | [int64](#int64) |  | SuspendDurationInBlocks defines the number of blocks a validator is disallowed to participate in any TSS ceremony after committing a malicious behaviour during signing |
| `heartbeat_period_in_blocks` | [int64](#int64) |  | HeartBeatPeriodInBlocks defines the time period in blocks for tss to emit the event asking validators to send their heartbeats |
| `max_missed_blocks_per_window` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `unbonding_locking_key_rotation_count` | [int64](#int64) |  |  |
| `external_multisig_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `max_sign_queue_size` | [int64](#int64) |  |  |
| `max_simultaneous_sign_shares` | [int64](#int64) |  |  |
| `tss_signed_blocks_window` | [int64](#int64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/tss/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/tss/v1beta1/types.proto



<a name="axelar.tss.v1beta1.ExternalKeys"></a>

### ExternalKeys



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `key_ids` | [string](#string) | repeated |  |






<a name="axelar.tss.v1beta1.KeyInfo"></a>

### KeyInfo
KeyInfo holds information about a key


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `key_role` | [axelar.tss.exported.v1beta1.KeyRole](#axelar.tss.exported.v1beta1.KeyRole) |  |  |
| `key_type` | [axelar.tss.exported.v1beta1.KeyType](#axelar.tss.exported.v1beta1.KeyType) |  |  |






<a name="axelar.tss.v1beta1.KeyRecoveryInfo"></a>

### KeyRecoveryInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `public` | [bytes](#bytes) |  |  |
| `private` | [KeyRecoveryInfo.PrivateEntry](#axelar.tss.v1beta1.KeyRecoveryInfo.PrivateEntry) | repeated |  |






<a name="axelar.tss.v1beta1.KeyRecoveryInfo.PrivateEntry"></a>

### KeyRecoveryInfo.PrivateEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [bytes](#bytes) |  |  |






<a name="axelar.tss.v1beta1.KeygenVoteData"></a>

### KeygenVoteData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pub_key` | [bytes](#bytes) |  |  |
| `group_recovery_info` | [bytes](#bytes) |  |  |






<a name="axelar.tss.v1beta1.MultisigInfo"></a>

### MultisigInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `timeout` | [int64](#int64) |  |  |
| `target_num` | [int64](#int64) |  |  |
| `infos` | [MultisigInfo.Info](#axelar.tss.v1beta1.MultisigInfo.Info) | repeated |  |






<a name="axelar.tss.v1beta1.MultisigInfo.Info"></a>

### MultisigInfo.Info



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `participant` | [bytes](#bytes) |  |  |
| `data` | [bytes](#bytes) | repeated |  |






<a name="axelar.tss.v1beta1.ValidatorStatus"></a>

### ValidatorStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validator` | [bytes](#bytes) |  |  |
| `suspended_until` | [uint64](#uint64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/tss/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/tss/v1beta1/genesis.proto



<a name="axelar.tss.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelar.tss.v1beta1.Params) |  |  |
| `key_recovery_infos` | [KeyRecoveryInfo](#axelar.tss.v1beta1.KeyRecoveryInfo) | repeated |  |
| `keys` | [axelar.tss.exported.v1beta1.Key](#axelar.tss.exported.v1beta1.Key) | repeated |  |
| `multisig_infos` | [MultisigInfo](#axelar.tss.v1beta1.MultisigInfo) | repeated |  |
| `external_keys` | [ExternalKeys](#axelar.tss.v1beta1.ExternalKeys) | repeated |  |
| `signatures` | [axelar.tss.exported.v1beta1.Signature](#axelar.tss.exported.v1beta1.Signature) | repeated |  |
| `validator_statuses` | [ValidatorStatus](#axelar.tss.v1beta1.ValidatorStatus) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/tss/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/tss/v1beta1/query.proto



<a name="axelar.tss.v1beta1.AssignableKeyRequest"></a>

### AssignableKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `key_role` | [axelar.tss.exported.v1beta1.KeyRole](#axelar.tss.exported.v1beta1.KeyRole) |  |  |






<a name="axelar.tss.v1beta1.AssignableKeyResponse"></a>

### AssignableKeyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `assignable` | [bool](#bool) |  |  |






<a name="axelar.tss.v1beta1.NextKeyIDRequest"></a>

### NextKeyIDRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `key_role` | [axelar.tss.exported.v1beta1.KeyRole](#axelar.tss.exported.v1beta1.KeyRole) |  |  |






<a name="axelar.tss.v1beta1.NextKeyIDResponse"></a>

### NextKeyIDResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |






<a name="axelar.tss.v1beta1.QueryActiveOldKeysResponse"></a>

### QueryActiveOldKeysResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_ids` | [string](#string) | repeated |  |






<a name="axelar.tss.v1beta1.QueryActiveOldKeysValidatorResponse"></a>

### QueryActiveOldKeysValidatorResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `keys_info` | [QueryActiveOldKeysValidatorResponse.KeyInfo](#axelar.tss.v1beta1.QueryActiveOldKeysValidatorResponse.KeyInfo) | repeated |  |






<a name="axelar.tss.v1beta1.QueryActiveOldKeysValidatorResponse.KeyInfo"></a>

### QueryActiveOldKeysValidatorResponse.KeyInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `chain` | [string](#string) |  |  |
| `role` | [int32](#int32) |  |  |






<a name="axelar.tss.v1beta1.QueryDeactivatedOperatorsResponse"></a>

### QueryDeactivatedOperatorsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `operator_addresses` | [string](#string) | repeated |  |






<a name="axelar.tss.v1beta1.QueryExternalKeyIDResponse"></a>

### QueryExternalKeyIDResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_ids` | [string](#string) | repeated |  |






<a name="axelar.tss.v1beta1.QueryKeyResponse"></a>

### QueryKeyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ecdsa_key` | [QueryKeyResponse.ECDSAKey](#axelar.tss.v1beta1.QueryKeyResponse.ECDSAKey) |  |  |
| `multisig_key` | [QueryKeyResponse.MultisigKey](#axelar.tss.v1beta1.QueryKeyResponse.MultisigKey) |  |  |
| `role` | [axelar.tss.exported.v1beta1.KeyRole](#axelar.tss.exported.v1beta1.KeyRole) |  |  |
| `rotated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="axelar.tss.v1beta1.QueryKeyResponse.ECDSAKey"></a>

### QueryKeyResponse.ECDSAKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `vote_status` | [VoteStatus](#axelar.tss.v1beta1.VoteStatus) |  |  |
| `key` | [QueryKeyResponse.Key](#axelar.tss.v1beta1.QueryKeyResponse.Key) |  |  |






<a name="axelar.tss.v1beta1.QueryKeyResponse.Key"></a>

### QueryKeyResponse.Key



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `x` | [string](#string) |  |  |
| `y` | [string](#string) |  |  |






<a name="axelar.tss.v1beta1.QueryKeyResponse.MultisigKey"></a>

### QueryKeyResponse.MultisigKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `threshold` | [int64](#int64) |  |  |
| `key` | [QueryKeyResponse.Key](#axelar.tss.v1beta1.QueryKeyResponse.Key) | repeated |  |






<a name="axelar.tss.v1beta1.QueryKeyShareResponse"></a>

### QueryKeyShareResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `share_infos` | [QueryKeyShareResponse.ShareInfo](#axelar.tss.v1beta1.QueryKeyShareResponse.ShareInfo) | repeated |  |






<a name="axelar.tss.v1beta1.QueryKeyShareResponse.ShareInfo"></a>

### QueryKeyShareResponse.ShareInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `key_chain` | [string](#string) |  |  |
| `key_role` | [string](#string) |  |  |
| `snapshot_block_number` | [int64](#int64) |  |  |
| `validator_address` | [string](#string) |  |  |
| `num_validator_shares` | [int64](#int64) |  |  |
| `num_total_shares` | [int64](#int64) |  |  |






<a name="axelar.tss.v1beta1.QueryRecoveryResponse"></a>

### QueryRecoveryResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `party_uids` | [string](#string) | repeated |  |
| `party_share_counts` | [uint32](#uint32) | repeated |  |
| `threshold` | [uint32](#uint32) |  |  |
| `keygen_output` | [axelar.tss.tofnd.v1beta1.KeygenOutput](#axelar.tss.tofnd.v1beta1.KeygenOutput) |  |  |






<a name="axelar.tss.v1beta1.QuerySignatureResponse"></a>

### QuerySignatureResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `threshold_signature` | [QuerySignatureResponse.ThresholdSignature](#axelar.tss.v1beta1.QuerySignatureResponse.ThresholdSignature) |  |  |
| `multisig_signature` | [QuerySignatureResponse.MultisigSignature](#axelar.tss.v1beta1.QuerySignatureResponse.MultisigSignature) |  |  |






<a name="axelar.tss.v1beta1.QuerySignatureResponse.MultisigSignature"></a>

### QuerySignatureResponse.MultisigSignature



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sig_status` | [axelar.tss.exported.v1beta1.SigStatus](#axelar.tss.exported.v1beta1.SigStatus) |  |  |
| `signatures` | [QuerySignatureResponse.Signature](#axelar.tss.v1beta1.QuerySignatureResponse.Signature) | repeated |  |






<a name="axelar.tss.v1beta1.QuerySignatureResponse.Signature"></a>

### QuerySignatureResponse.Signature



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `r` | [string](#string) |  |  |
| `s` | [string](#string) |  |  |






<a name="axelar.tss.v1beta1.QuerySignatureResponse.ThresholdSignature"></a>

### QuerySignatureResponse.ThresholdSignature



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `vote_status` | [VoteStatus](#axelar.tss.v1beta1.VoteStatus) |  |  |
| `signature` | [QuerySignatureResponse.Signature](#axelar.tss.v1beta1.QuerySignatureResponse.Signature) |  |  |






<a name="axelar.tss.v1beta1.ValidatorMultisigKeysRequest"></a>

### ValidatorMultisigKeysRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="axelar.tss.v1beta1.ValidatorMultisigKeysResponse"></a>

### ValidatorMultisigKeysResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `keys` | [ValidatorMultisigKeysResponse.KeysEntry](#axelar.tss.v1beta1.ValidatorMultisigKeysResponse.KeysEntry) | repeated |  |






<a name="axelar.tss.v1beta1.ValidatorMultisigKeysResponse.Keys"></a>

### ValidatorMultisigKeysResponse.Keys



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `keys` | [bytes](#bytes) | repeated |  |






<a name="axelar.tss.v1beta1.ValidatorMultisigKeysResponse.KeysEntry"></a>

### ValidatorMultisigKeysResponse.KeysEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [ValidatorMultisigKeysResponse.Keys](#axelar.tss.v1beta1.ValidatorMultisigKeysResponse.Keys) |  |  |





 <!-- end messages -->


<a name="axelar.tss.v1beta1.VoteStatus"></a>

### VoteStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| VOTE_STATUS_UNSPECIFIED | 0 |  |
| VOTE_STATUS_NOT_FOUND | 1 |  |
| VOTE_STATUS_PENDING | 2 |  |
| VOTE_STATUS_DECIDED | 3 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/tss/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/tss/v1beta1/tx.proto



<a name="axelar.tss.v1beta1.HeartBeatRequest"></a>

### HeartBeatRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `key_ids` | [string](#string) | repeated |  |






<a name="axelar.tss.v1beta1.HeartBeatResponse"></a>

### HeartBeatResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `keygen_illegibility` | [int32](#int32) |  |  |
| `signing_illegibility` | [int32](#int32) |  |  |






<a name="axelar.tss.v1beta1.ProcessKeygenTrafficRequest"></a>

### ProcessKeygenTrafficRequest
ProcessKeygenTrafficRequest protocol message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `session_id` | [string](#string) |  |  |
| `payload` | [axelar.tss.tofnd.v1beta1.TrafficOut](#axelar.tss.tofnd.v1beta1.TrafficOut) |  |  |






<a name="axelar.tss.v1beta1.ProcessKeygenTrafficResponse"></a>

### ProcessKeygenTrafficResponse







<a name="axelar.tss.v1beta1.ProcessSignTrafficRequest"></a>

### ProcessSignTrafficRequest
ProcessSignTrafficRequest protocol message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `session_id` | [string](#string) |  |  |
| `payload` | [axelar.tss.tofnd.v1beta1.TrafficOut](#axelar.tss.tofnd.v1beta1.TrafficOut) |  |  |






<a name="axelar.tss.v1beta1.ProcessSignTrafficResponse"></a>

### ProcessSignTrafficResponse







<a name="axelar.tss.v1beta1.RegisterExternalKeysRequest"></a>

### RegisterExternalKeysRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `external_keys` | [RegisterExternalKeysRequest.ExternalKey](#axelar.tss.v1beta1.RegisterExternalKeysRequest.ExternalKey) | repeated |  |






<a name="axelar.tss.v1beta1.RegisterExternalKeysRequest.ExternalKey"></a>

### RegisterExternalKeysRequest.ExternalKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `pub_key` | [bytes](#bytes) |  |  |






<a name="axelar.tss.v1beta1.RegisterExternalKeysResponse"></a>

### RegisterExternalKeysResponse







<a name="axelar.tss.v1beta1.RotateKeyRequest"></a>

### RotateKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `key_role` | [axelar.tss.exported.v1beta1.KeyRole](#axelar.tss.exported.v1beta1.KeyRole) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="axelar.tss.v1beta1.RotateKeyResponse"></a>

### RotateKeyResponse







<a name="axelar.tss.v1beta1.StartKeygenRequest"></a>

### StartKeygenRequest
StartKeygenRequest indicate the start of keygen


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `key_info` | [KeyInfo](#axelar.tss.v1beta1.KeyInfo) |  |  |






<a name="axelar.tss.v1beta1.StartKeygenResponse"></a>

### StartKeygenResponse







<a name="axelar.tss.v1beta1.SubmitMultisigPubKeysRequest"></a>

### SubmitMultisigPubKeysRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `sig_key_pairs` | [axelar.tss.exported.v1beta1.SigKeyPair](#axelar.tss.exported.v1beta1.SigKeyPair) | repeated |  |






<a name="axelar.tss.v1beta1.SubmitMultisigPubKeysResponse"></a>

### SubmitMultisigPubKeysResponse







<a name="axelar.tss.v1beta1.SubmitMultisigSignaturesRequest"></a>

### SubmitMultisigSignaturesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `sig_id` | [string](#string) |  |  |
| `signatures` | [bytes](#bytes) | repeated |  |






<a name="axelar.tss.v1beta1.SubmitMultisigSignaturesResponse"></a>

### SubmitMultisigSignaturesResponse







<a name="axelar.tss.v1beta1.VotePubKeyRequest"></a>

### VotePubKeyRequest
VotePubKeyRequest represents the message to vote on a public key


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll_key` | [axelar.vote.exported.v1beta1.PollKey](#axelar.vote.exported.v1beta1.PollKey) |  |  |
| `result` | [axelar.tss.tofnd.v1beta1.MessageOut.KeygenResult](#axelar.tss.tofnd.v1beta1.MessageOut.KeygenResult) |  |  |






<a name="axelar.tss.v1beta1.VotePubKeyResponse"></a>

### VotePubKeyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |






<a name="axelar.tss.v1beta1.VoteSigRequest"></a>

### VoteSigRequest
VoteSigRequest represents a message to vote for a signature


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll_key` | [axelar.vote.exported.v1beta1.PollKey](#axelar.vote.exported.v1beta1.PollKey) |  |  |
| `result` | [axelar.tss.tofnd.v1beta1.MessageOut.SignResult](#axelar.tss.tofnd.v1beta1.MessageOut.SignResult) |  |  |






<a name="axelar.tss.v1beta1.VoteSigResponse"></a>

### VoteSigResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/tss/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/tss/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelar.tss.v1beta1.MsgService"></a>

### MsgService
Msg defines the tss Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterExternalKeys` | [RegisterExternalKeysRequest](#axelar.tss.v1beta1.RegisterExternalKeysRequest) | [RegisterExternalKeysResponse](#axelar.tss.v1beta1.RegisterExternalKeysResponse) |  | POST|/axelar/tss/register_external_keys|
| `HeartBeat` | [HeartBeatRequest](#axelar.tss.v1beta1.HeartBeatRequest) | [HeartBeatResponse](#axelar.tss.v1beta1.HeartBeatResponse) |  | POST|/axelar/tss/heartbeat|
| `StartKeygen` | [StartKeygenRequest](#axelar.tss.v1beta1.StartKeygenRequest) | [StartKeygenResponse](#axelar.tss.v1beta1.StartKeygenResponse) |  | POST|/axelar/tss/start_keygen|
| `ProcessKeygenTraffic` | [ProcessKeygenTrafficRequest](#axelar.tss.v1beta1.ProcessKeygenTrafficRequest) | [ProcessKeygenTrafficResponse](#axelar.tss.v1beta1.ProcessKeygenTrafficResponse) |  | POST|/axelar/tss/process_keygen_traffic|
| `RotateKey` | [RotateKeyRequest](#axelar.tss.v1beta1.RotateKeyRequest) | [RotateKeyResponse](#axelar.tss.v1beta1.RotateKeyResponse) |  | POST|/axelar/tss/rotate_key|
| `VotePubKey` | [VotePubKeyRequest](#axelar.tss.v1beta1.VotePubKeyRequest) | [VotePubKeyResponse](#axelar.tss.v1beta1.VotePubKeyResponse) |  | POST|/axelar/tss/vote_pub_key|
| `ProcessSignTraffic` | [ProcessSignTrafficRequest](#axelar.tss.v1beta1.ProcessSignTrafficRequest) | [ProcessSignTrafficResponse](#axelar.tss.v1beta1.ProcessSignTrafficResponse) |  | POST|/axelar/tss/process_sign_traffic|
| `VoteSig` | [VoteSigRequest](#axelar.tss.v1beta1.VoteSigRequest) | [VoteSigResponse](#axelar.tss.v1beta1.VoteSigResponse) |  | POST|/axelar/tss/vote_sig|
| `SubmitMultisigPubKeys` | [SubmitMultisigPubKeysRequest](#axelar.tss.v1beta1.SubmitMultisigPubKeysRequest) | [SubmitMultisigPubKeysResponse](#axelar.tss.v1beta1.SubmitMultisigPubKeysResponse) |  | POST|/axelar/tss/submit_multisig_pub_keys|
| `SubmitMultisigSignatures` | [SubmitMultisigSignaturesRequest](#axelar.tss.v1beta1.SubmitMultisigSignaturesRequest) | [SubmitMultisigSignaturesResponse](#axelar.tss.v1beta1.SubmitMultisigSignaturesResponse) |  | POST|/axelar/tss/submit_multisig_signatures|


<a name="axelar.tss.v1beta1.QueryService"></a>

### QueryService
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `NextKeyID` | [NextKeyIDRequest](#axelar.tss.v1beta1.NextKeyIDRequest) | [NextKeyIDResponse](#axelar.tss.v1beta1.NextKeyIDResponse) | NextKeyID returns the key ID assigned for the next rotation on a given chain and for the given key role | GET|/axelar/tss/v1beta1/next_key_id/{chain}/{key_role}|
| `AssignableKey` | [AssignableKeyRequest](#axelar.tss.v1beta1.AssignableKeyRequest) | [AssignableKeyResponse](#axelar.tss.v1beta1.AssignableKeyResponse) | AssignableKey returns true if there is no assigned key for the next rotation on a given chain, and false otherwise | GET|/axelar/tss/v1beta1/assignable_key/{chain}/{key_role}|
| `ValidatorMultisigKeys` | [ValidatorMultisigKeysRequest](#axelar.tss.v1beta1.ValidatorMultisigKeysRequest) | [ValidatorMultisigKeysResponse](#axelar.tss.v1beta1.ValidatorMultisigKeysResponse) | ValidatorMultisigKeys returns the validator's multisig pubkeys corresponding to each active key ID | GET|/axelar/tss/v1beta1/validator_multisig_keys/{address}|

 <!-- end services -->



<a name="axelar/vote/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/vote/v1beta1/params.proto



<a name="axelar.vote.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `default_voting_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/vote/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/vote/v1beta1/genesis.proto



<a name="axelar.vote.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelar.vote.v1beta1.Params) |  |  |
| `poll_metadatas` | [axelar.vote.exported.v1beta1.PollMetadata](#axelar.vote.exported.v1beta1.PollMetadata) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/vote/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/vote/v1beta1/tx.proto



<a name="axelar.vote.v1beta1.VoteRequest"></a>

### VoteRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll_key` | [axelar.vote.exported.v1beta1.PollKey](#axelar.vote.exported.v1beta1.PollKey) |  |  |
| `vote` | [axelar.vote.exported.v1beta1.Vote](#axelar.vote.exported.v1beta1.Vote) |  |  |






<a name="axelar.vote.v1beta1.VoteResponse"></a>

### VoteResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/vote/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/vote/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelar.vote.v1beta1.MsgService"></a>

### MsgService
Msg defines the vote Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Vote` | [VoteRequest](#axelar.vote.v1beta1.VoteRequest) | [VoteResponse](#axelar.vote.v1beta1.VoteResponse) |  | POST|/axelar/vote/vote|

 <!-- end services -->



<a name="axelar/vote/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/vote/v1beta1/types.proto



<a name="axelar.vote.v1beta1.TalliedVote"></a>

### TalliedVote
TalliedVote represents a vote for a poll with the accumulated stake of all
validators voting for the same data


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tally` | [bytes](#bytes) |  |  |
| `voters` | [bytes](#bytes) | repeated |  |
| `data` | [google.protobuf.Any](#google.protobuf.Any) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="utils/v1beta1/queuer.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## utils/v1beta1/queuer.proto



<a name="utils.v1beta1.QueueState"></a>

### QueueState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `items` | [QueueState.ItemsEntry](#utils.v1beta1.QueueState.ItemsEntry) | repeated |  |






<a name="utils.v1beta1.QueueState.Item"></a>

### QueueState.Item



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [bytes](#bytes) |  |  |
| `value` | [bytes](#bytes) |  |  |






<a name="utils.v1beta1.QueueState.ItemsEntry"></a>

### QueueState.ItemsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [QueueState.Item](#utils.v1beta1.QueueState.Item) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="utils/v1beta1/threshold.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## utils/v1beta1/threshold.proto



<a name="utils.v1beta1.Threshold"></a>

### Threshold



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `numerator` | [int64](#int64) |  | split threshold into Numerator and denominator to avoid floating point errors down the line |
| `denominator` | [int64](#int64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="evm/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evm/v1beta1/types.proto



<a name="evm.v1beta1.Asset"></a>

### Asset



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `name` | [string](#string) |  |  |






<a name="evm.v1beta1.BurnerInfo"></a>

### BurnerInfo
BurnerInfo describes information required to burn token at an burner address
that is deposited by an user


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `burner_address` | [bytes](#bytes) |  |  |
| `token_address` | [bytes](#bytes) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `asset` | [string](#string) |  |  |
| `salt` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.Command"></a>

### Command



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [bytes](#bytes) |  |  |
| `command` | [string](#string) |  |  |
| `params` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `max_gas_cost` | [uint32](#uint32) |  |  |






<a name="evm.v1beta1.CommandBatchMetadata"></a>

### CommandBatchMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [bytes](#bytes) |  |  |
| `command_ids` | [bytes](#bytes) | repeated |  |
| `data` | [bytes](#bytes) |  |  |
| `sig_hash` | [bytes](#bytes) |  |  |
| `status` | [BatchedCommandsStatus](#evm.v1beta1.BatchedCommandsStatus) |  |  |
| `key_id` | [string](#string) |  |  |
| `prev_batched_commands_id` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.ERC20Deposit"></a>

### ERC20Deposit
ERC20Deposit contains information for an ERC20 deposit


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `amount` | [bytes](#bytes) |  |  |
| `asset` | [string](#string) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `burner_address` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.ERC20TokenMetadata"></a>

### ERC20TokenMetadata
ERC20TokenMetadata describes information about an ERC20 token


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset` | [string](#string) |  |  |
| `chain_id` | [bytes](#bytes) |  |  |
| `details` | [TokenDetails](#evm.v1beta1.TokenDetails) |  |  |
| `token_address` | [string](#string) |  |  |
| `tx_hash` | [string](#string) |  |  |
| `status` | [Status](#evm.v1beta1.Status) |  |  |
| `is_external` | [bool](#bool) |  |  |
| `burner_code` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.Event"></a>

### Event



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `index` | [uint64](#uint64) |  |  |
| `status` | [Event.Status](#evm.v1beta1.Event.Status) |  |  |
| `token_sent` | [EventTokenSent](#evm.v1beta1.EventTokenSent) |  |  |
| `contract_call` | [EventContractCall](#evm.v1beta1.EventContractCall) |  |  |
| `contract_call_with_token` | [EventContractCallWithToken](#evm.v1beta1.EventContractCallWithToken) |  |  |
| `transfer` | [EventTransfer](#evm.v1beta1.EventTransfer) |  |  |
| `token_deployed` | [EventTokenDeployed](#evm.v1beta1.EventTokenDeployed) |  |  |
| `multisig_ownership_transferred` | [EventMultisigOwnershipTransferred](#evm.v1beta1.EventMultisigOwnershipTransferred) |  |  |
| `multisig_operatorship_transferred` | [EventMultisigOperatorshipTransferred](#evm.v1beta1.EventMultisigOperatorshipTransferred) |  |  |
| `singlesig_ownership_transferred` | [EventSinglesigOwnershipTransferred](#evm.v1beta1.EventSinglesigOwnershipTransferred) |  |  |
| `singlesig_operatorship_transferred` | [EventSinglesigOperatorshipTransferred](#evm.v1beta1.EventSinglesigOperatorshipTransferred) |  |  |






<a name="evm.v1beta1.EventContractCall"></a>

### EventContractCall



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `contract_address` | [string](#string) |  |  |
| `payload_hash` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.EventContractCallWithToken"></a>

### EventContractCallWithToken



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `contract_address` | [string](#string) |  |  |
| `payload_hash` | [bytes](#bytes) |  |  |
| `symbol` | [string](#string) |  |  |
| `amount` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.EventMultisigOperatorshipTransferred"></a>

### EventMultisigOperatorshipTransferred



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pre_operators` | [bytes](#bytes) | repeated |  |
| `prev_threshold` | [bytes](#bytes) |  |  |
| `new_operators` | [bytes](#bytes) | repeated |  |
| `new_threshold` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.EventMultisigOwnershipTransferred"></a>

### EventMultisigOwnershipTransferred



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pre_owners` | [bytes](#bytes) | repeated |  |
| `prev_threshold` | [bytes](#bytes) |  |  |
| `new_owners` | [bytes](#bytes) | repeated |  |
| `new_threshold` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.EventSinglesigOperatorshipTransferred"></a>

### EventSinglesigOperatorshipTransferred



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pre_operator` | [bytes](#bytes) |  |  |
| `new_operator` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.EventSinglesigOwnershipTransferred"></a>

### EventSinglesigOwnershipTransferred



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pre_owner` | [bytes](#bytes) |  |  |
| `new_owner` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.EventTokenDeployed"></a>

### EventTokenDeployed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `symbol` | [string](#string) |  |  |
| `token_address` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.EventTokenSent"></a>

### EventTokenSent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `destination_address` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `amount` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.EventTransfer"></a>

### EventTransfer



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to` | [bytes](#bytes) |  |  |
| `amount` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.Gateway"></a>

### Gateway



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [bytes](#bytes) |  |  |
| `status` | [Gateway.Status](#evm.v1beta1.Gateway.Status) |  | **Deprecated.**  |






<a name="evm.v1beta1.NetworkInfo"></a>

### NetworkInfo
NetworkInfo describes information about a network


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `id` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.SigMetadata"></a>

### SigMetadata
SigMetadata stores necessary information for external apps to map signature
results to evm relay transaction types


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `type` | [SigType](#evm.v1beta1.SigType) |  |  |
| `chain` | [string](#string) |  |  |






<a name="evm.v1beta1.TokenDetails"></a>

### TokenDetails



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `token_name` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `decimals` | [uint32](#uint32) |  |  |
| `capacity` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.TransactionMetadata"></a>

### TransactionMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `raw_tx` | [bytes](#bytes) |  |  |
| `pub_key` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.TransferKey"></a>

### TransferKey
TransferKey contains information for a transfer ownership or operatorship


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `type` | [TransferKeyType](#evm.v1beta1.TransferKeyType) |  |  |
| `next_key_id` | [string](#string) |  |  |





 <!-- end messages -->


<a name="evm.v1beta1.BatchedCommandsStatus"></a>

### BatchedCommandsStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| BATCHED_COMMANDS_STATUS_UNSPECIFIED | 0 |  |
| BATCHED_COMMANDS_STATUS_SIGNING | 1 |  |
| BATCHED_COMMANDS_STATUS_ABORTED | 2 |  |
| BATCHED_COMMANDS_STATUS_SIGNED | 3 |  |



<a name="evm.v1beta1.DepositStatus"></a>

### DepositStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| DEPOSIT_STATUS_UNSPECIFIED | 0 |  |
| DEPOSIT_STATUS_PENDING | 1 |  |
| DEPOSIT_STATUS_CONFIRMED | 2 |  |
| DEPOSIT_STATUS_BURNED | 3 |  |



<a name="evm.v1beta1.Event.Status"></a>

### Event.Status


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATUS_UNSPECIFIED | 0 |  |
| STATUS_CONFIRMED | 1 |  |
| STATUS_COMPLETED | 2 |  |
| STATUS_FAILED | 3 |  |



<a name="evm.v1beta1.Gateway.Status"></a>

### Gateway.Status


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATUS_UNSPECIFIED | 0 |  |
| STATUS_PENDING | 1 |  |
| STATUS_CONFIRMED | 2 |  |



<a name="evm.v1beta1.SigType"></a>

### SigType


| Name | Number | Description |
| ---- | ------ | ----------- |
| SIG_TYPE_UNSPECIFIED | 0 |  |
| SIG_TYPE_TX | 1 |  |
| SIG_TYPE_COMMAND | 2 |  |



<a name="evm.v1beta1.Status"></a>

### Status


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATUS_UNSPECIFIED | 0 | these enum values are used for bitwise operations, therefore they need to be powers of 2 |
| STATUS_INITIALIZED | 1 |  |
| STATUS_PENDING | 2 |  |
| STATUS_CONFIRMED | 4 |  |



<a name="evm.v1beta1.TransferKeyType"></a>

### TransferKeyType


| Name | Number | Description |
| ---- | ------ | ----------- |
| TRANSFER_KEY_TYPE_UNSPECIFIED | 0 |  |
| TRANSFER_KEY_TYPE_OWNERSHIP | 1 |  |
| TRANSFER_KEY_TYPE_OPERATORSHIP | 2 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="evm/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evm/v1beta1/params.proto



<a name="evm.v1beta1.Params"></a>

### Params
Params is the parameter set for this module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `confirmation_height` | [uint64](#uint64) |  |  |
| `network` | [string](#string) |  |  |
| `token_code` | [bytes](#bytes) |  |  |
| `burnable` | [bytes](#bytes) |  |  |
| `revote_locking_period` | [int64](#int64) |  |  |
| `networks` | [NetworkInfo](#evm.v1beta1.NetworkInfo) | repeated |  |
| `voting_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  |  |
| `min_voter_count` | [int64](#int64) |  |  |
| `commands_gas_limit` | [uint32](#uint32) |  |  |






<a name="evm.v1beta1.PendingChain"></a>

### PendingChain



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#evm.v1beta1.Params) |  |  |
| `chain` | [axelar.nexus.exported.v1beta1.Chain](#axelar.nexus.exported.v1beta1.Chain) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="evm/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evm/v1beta1/genesis.proto



<a name="evm.v1beta1.GenesisState"></a>

### GenesisState
GenesisState represents the genesis state


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chains` | [GenesisState.Chain](#evm.v1beta1.GenesisState.Chain) | repeated |  |






<a name="evm.v1beta1.GenesisState.Chain"></a>

### GenesisState.Chain



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#evm.v1beta1.Params) |  |  |
| `burner_infos` | [BurnerInfo](#evm.v1beta1.BurnerInfo) | repeated |  |
| `command_queue` | [utils.v1beta1.QueueState](#utils.v1beta1.QueueState) |  |  |
| `confirmed_deposits` | [ERC20Deposit](#evm.v1beta1.ERC20Deposit) | repeated |  |
| `burned_deposits` | [ERC20Deposit](#evm.v1beta1.ERC20Deposit) | repeated |  |
| `command_batches` | [CommandBatchMetadata](#evm.v1beta1.CommandBatchMetadata) | repeated |  |
| `gateway` | [Gateway](#evm.v1beta1.Gateway) |  |  |
| `tokens` | [ERC20TokenMetadata](#evm.v1beta1.ERC20TokenMetadata) | repeated |  |
| `events` | [Event](#evm.v1beta1.Event) | repeated |  |
| `confirmed_event_queue` | [utils.v1beta1.QueueState](#utils.v1beta1.QueueState) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="evm/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evm/v1beta1/query.proto



<a name="evm.v1beta1.BatchedCommandsRequest"></a>

### BatchedCommandsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `id` | [string](#string) |  | id defines an optional id for the commandsbatch. If not specified the latest will be returned |






<a name="evm.v1beta1.BatchedCommandsResponse"></a>

### BatchedCommandsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `data` | [string](#string) |  |  |
| `status` | [BatchedCommandsStatus](#evm.v1beta1.BatchedCommandsStatus) |  |  |
| `key_id` | [string](#string) |  |  |
| `signature` | [string](#string) | repeated |  |
| `execute_data` | [string](#string) |  |  |
| `prev_batched_commands_id` | [string](#string) |  |  |
| `command_ids` | [string](#string) | repeated |  |






<a name="evm.v1beta1.BurnerInfoRequest"></a>

### BurnerInfoRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.BurnerInfoResponse"></a>

### BurnerInfoResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `burner_info` | [BurnerInfo](#evm.v1beta1.BurnerInfo) |  |  |






<a name="evm.v1beta1.BytecodeRequest"></a>

### BytecodeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `contract` | [string](#string) |  |  |






<a name="evm.v1beta1.BytecodeResponse"></a>

### BytecodeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `bytecode` | [string](#string) |  |  |






<a name="evm.v1beta1.ChainsRequest"></a>

### ChainsRequest







<a name="evm.v1beta1.ChainsResponse"></a>

### ChainsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chains` | [string](#string) | repeated |  |






<a name="evm.v1beta1.ConfirmationHeightRequest"></a>

### ConfirmationHeightRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |






<a name="evm.v1beta1.ConfirmationHeightResponse"></a>

### ConfirmationHeightResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `height` | [uint64](#uint64) |  |  |






<a name="evm.v1beta1.DepositQueryParams"></a>

### DepositQueryParams
DepositQueryParams describe the parameters used to query for an EVM
deposit address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `asset` | [string](#string) |  |  |
| `chain` | [string](#string) |  |  |






<a name="evm.v1beta1.DepositStateRequest"></a>

### DepositStateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `params` | [QueryDepositStateParams](#evm.v1beta1.QueryDepositStateParams) |  |  |






<a name="evm.v1beta1.DepositStateResponse"></a>

### DepositStateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `status` | [DepositStatus](#evm.v1beta1.DepositStatus) |  |  |






<a name="evm.v1beta1.EventRequest"></a>

### EventRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `event_id` | [string](#string) |  |  |






<a name="evm.v1beta1.EventResponse"></a>

### EventResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `event` | [Event](#evm.v1beta1.Event) |  |  |






<a name="evm.v1beta1.GatewayAddressRequest"></a>

### GatewayAddressRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |






<a name="evm.v1beta1.GatewayAddressResponse"></a>

### GatewayAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="evm.v1beta1.KeyAddressRequest"></a>

### KeyAddressRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `role` | [int32](#int32) |  |  |
| `id` | [string](#string) |  |  |






<a name="evm.v1beta1.KeyAddressResponse"></a>

### KeyAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `multisig_addresses` | [KeyAddressResponse.MultisigAddresses](#evm.v1beta1.KeyAddressResponse.MultisigAddresses) |  |  |
| `threshold_address` | [KeyAddressResponse.ThresholdAddress](#evm.v1beta1.KeyAddressResponse.ThresholdAddress) |  |  |






<a name="evm.v1beta1.KeyAddressResponse.MultisigAddresses"></a>

### KeyAddressResponse.MultisigAddresses



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `addresses` | [string](#string) | repeated |  |
| `threshold` | [uint32](#uint32) |  |  |






<a name="evm.v1beta1.KeyAddressResponse.ThresholdAddress"></a>

### KeyAddressResponse.ThresholdAddress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="evm.v1beta1.PendingCommandsRequest"></a>

### PendingCommandsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |






<a name="evm.v1beta1.PendingCommandsResponse"></a>

### PendingCommandsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `commands` | [QueryCommandResponse](#evm.v1beta1.QueryCommandResponse) | repeated |  |






<a name="evm.v1beta1.QueryBurnerAddressResponse"></a>

### QueryBurnerAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="evm.v1beta1.QueryCommandResponse"></a>

### QueryCommandResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `type` | [string](#string) |  |  |
| `params` | [QueryCommandResponse.ParamsEntry](#evm.v1beta1.QueryCommandResponse.ParamsEntry) | repeated |  |
| `key_id` | [string](#string) |  |  |
| `max_gas_cost` | [uint32](#uint32) |  |  |






<a name="evm.v1beta1.QueryCommandResponse.ParamsEntry"></a>

### QueryCommandResponse.ParamsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [string](#string) |  |  |






<a name="evm.v1beta1.QueryDepositStateParams"></a>

### QueryDepositStateParams



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `burner_address` | [bytes](#bytes) |  |  |
| `amount` | [string](#string) |  |  |






<a name="evm.v1beta1.QueryTokenAddressResponse"></a>

### QueryTokenAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `confirmed` | [bool](#bool) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="vote/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## vote/exported/v1beta1/types.proto



<a name="vote.exported.v1beta1.PollMetadata"></a>

### PollMetadata
PollMetadata represents a poll with write-in voting, i.e. the result of the
vote can have any data type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [axelar.vote.exported.v1beta1.PollKey](#axelar.vote.exported.v1beta1.PollKey) |  |  |
| `expires_at` | [int64](#int64) |  |  |
| `result` | [google.protobuf.Any](#google.protobuf.Any) |  |  |
| `voting_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `state` | [axelar.vote.exported.v1beta1.PollState](#axelar.vote.exported.v1beta1.PollState) |  |  |
| `min_voter_count` | [int64](#int64) |  |  |
| `voters` | [axelar.vote.exported.v1beta1.Voter](#axelar.vote.exported.v1beta1.Voter) | repeated |  |
| `total_voting_power` | [bytes](#bytes) |  |  |
| `reward_pool_name` | [string](#string) |  |  |






<a name="vote.exported.v1beta1.Vote"></a>

### Vote



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `results` | [google.protobuf.Any](#google.protobuf.Any) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="vote/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## vote/v1beta1/params.proto



<a name="vote.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `default_voting_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="vote/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## vote/v1beta1/genesis.proto



<a name="vote.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#vote.v1beta1.Params) |  |  |
| `poll_metadatas` | [vote.exported.v1beta1.PollMetadata](#vote.exported.v1beta1.PollMetadata) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="vote/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## vote/v1beta1/types.proto



<a name="vote.v1beta1.TalliedVote"></a>

### TalliedVote
TalliedVote represents a vote for a poll with the accumulated stake of all
validators voting for the same data


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tally` | [bytes](#bytes) |  |  |
| `voters` | [bytes](#bytes) | repeated |  |
| `data` | [google.protobuf.Any](#google.protobuf.Any) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

