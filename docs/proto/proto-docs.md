<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [axelarnetwork/axelarnet/v1beta1/params.proto](#axelarnetwork/axelarnet/v1beta1/params.proto)
    - [Params](#axelarnetwork.axelarnet.v1beta1.Params)
  
- [axelarnetwork/axelarnet/v1beta1/types.proto](#axelarnetwork/axelarnet/v1beta1/types.proto)
    - [Asset](#axelarnetwork.axelarnet.v1beta1.Asset)
    - [CosmosChain](#axelarnetwork.axelarnet.v1beta1.CosmosChain)
    - [IBCTransfer](#axelarnetwork.axelarnet.v1beta1.IBCTransfer)
  
- [axelarnetwork/axelarnet/v1beta1/genesis.proto](#axelarnetwork/axelarnet/v1beta1/genesis.proto)
    - [GenesisState](#axelarnetwork.axelarnet.v1beta1.GenesisState)
  
- [axelarnetwork/utils/v1beta1/threshold.proto](#axelarnetwork/utils/v1beta1/threshold.proto)
    - [Threshold](#axelarnetwork.utils.v1beta1.Threshold)
  
- [axelarnetwork/tss/exported/v1beta1/types.proto](#axelarnetwork/tss/exported/v1beta1/types.proto)
    - [Key](#axelarnetwork.tss.exported.v1beta1.Key)
    - [Key.ECDSAKey](#axelarnetwork.tss.exported.v1beta1.Key.ECDSAKey)
    - [Key.MultisigKey](#axelarnetwork.tss.exported.v1beta1.Key.MultisigKey)
    - [KeyRequirement](#axelarnetwork.tss.exported.v1beta1.KeyRequirement)
    - [SigKeyPair](#axelarnetwork.tss.exported.v1beta1.SigKeyPair)
    - [SignInfo](#axelarnetwork.tss.exported.v1beta1.SignInfo)
    - [Signature](#axelarnetwork.tss.exported.v1beta1.Signature)
    - [Signature.MultiSig](#axelarnetwork.tss.exported.v1beta1.Signature.MultiSig)
    - [Signature.SingleSig](#axelarnetwork.tss.exported.v1beta1.Signature.SingleSig)
  
    - [AckType](#axelarnetwork.tss.exported.v1beta1.AckType)
    - [KeyRole](#axelarnetwork.tss.exported.v1beta1.KeyRole)
    - [KeyShareDistributionPolicy](#axelarnetwork.tss.exported.v1beta1.KeyShareDistributionPolicy)
    - [KeyType](#axelarnetwork.tss.exported.v1beta1.KeyType)
    - [SigStatus](#axelarnetwork.tss.exported.v1beta1.SigStatus)
  
- [axelarnetwork/nexus/exported/v1beta1/types.proto](#axelarnetwork/nexus/exported/v1beta1/types.proto)
    - [Asset](#axelarnetwork.nexus.exported.v1beta1.Asset)
    - [Chain](#axelarnetwork.nexus.exported.v1beta1.Chain)
    - [CrossChainAddress](#axelarnetwork.nexus.exported.v1beta1.CrossChainAddress)
    - [CrossChainTransfer](#axelarnetwork.nexus.exported.v1beta1.CrossChainTransfer)
    - [FeeInfo](#axelarnetwork.nexus.exported.v1beta1.FeeInfo)
    - [TransferFee](#axelarnetwork.nexus.exported.v1beta1.TransferFee)
  
    - [TransferState](#axelarnetwork.nexus.exported.v1beta1.TransferState)
  
- [axelarnetwork/nexus/v1beta1/types.proto](#axelarnetwork/nexus/v1beta1/types.proto)
    - [ChainState](#axelarnetwork.nexus.v1beta1.ChainState)
    - [LinkedAddresses](#axelarnetwork.nexus.v1beta1.LinkedAddresses)
  
- [axelarnetwork/nexus/v1beta1/query.proto](#axelarnetwork/nexus/v1beta1/query.proto)
    - [AssetsRequest](#axelarnetwork.nexus.v1beta1.AssetsRequest)
    - [AssetsResponse](#axelarnetwork.nexus.v1beta1.AssetsResponse)
    - [ChainStateRequest](#axelarnetwork.nexus.v1beta1.ChainStateRequest)
    - [ChainStateResponse](#axelarnetwork.nexus.v1beta1.ChainStateResponse)
    - [ChainsByAssetRequest](#axelarnetwork.nexus.v1beta1.ChainsByAssetRequest)
    - [ChainsByAssetResponse](#axelarnetwork.nexus.v1beta1.ChainsByAssetResponse)
    - [ChainsRequest](#axelarnetwork.nexus.v1beta1.ChainsRequest)
    - [ChainsResponse](#axelarnetwork.nexus.v1beta1.ChainsResponse)
    - [FeeRequest](#axelarnetwork.nexus.v1beta1.FeeRequest)
    - [FeeResponse](#axelarnetwork.nexus.v1beta1.FeeResponse)
    - [LatestDepositAddressRequest](#axelarnetwork.nexus.v1beta1.LatestDepositAddressRequest)
    - [LatestDepositAddressResponse](#axelarnetwork.nexus.v1beta1.LatestDepositAddressResponse)
    - [QueryChainMaintainersResponse](#axelarnetwork.nexus.v1beta1.QueryChainMaintainersResponse)
    - [TransferFeeRequest](#axelarnetwork.nexus.v1beta1.TransferFeeRequest)
    - [TransferFeeResponse](#axelarnetwork.nexus.v1beta1.TransferFeeResponse)
    - [TransfersForChainRequest](#axelarnetwork.nexus.v1beta1.TransfersForChainRequest)
    - [TransfersForChainResponse](#axelarnetwork.nexus.v1beta1.TransfersForChainResponse)
  
- [axelarnetwork/axelarnet/v1beta1/query.proto](#axelarnetwork/axelarnet/v1beta1/query.proto)
    - [PendingIBCTransferCountRequest](#axelarnetwork.axelarnet.v1beta1.PendingIBCTransferCountRequest)
    - [PendingIBCTransferCountResponse](#axelarnetwork.axelarnet.v1beta1.PendingIBCTransferCountResponse)
    - [PendingIBCTransferCountResponse.TransfersByChainEntry](#axelarnetwork.axelarnet.v1beta1.PendingIBCTransferCountResponse.TransfersByChainEntry)
  
- [axelarnetwork/permission/exported/v1beta1/types.proto](#axelarnetwork/permission/exported/v1beta1/types.proto)
    - [Role](#axelarnetwork.permission.exported.v1beta1.Role)
  
    - [File-level Extensions](#axelarnetwork/permission/exported/v1beta1/types.proto-extensions)
  
- [axelarnetwork/axelarnet/v1beta1/tx.proto](#axelarnetwork/axelarnet/v1beta1/tx.proto)
    - [AddCosmosBasedChainRequest](#axelarnetwork.axelarnet.v1beta1.AddCosmosBasedChainRequest)
    - [AddCosmosBasedChainResponse](#axelarnetwork.axelarnet.v1beta1.AddCosmosBasedChainResponse)
    - [ConfirmDepositRequest](#axelarnetwork.axelarnet.v1beta1.ConfirmDepositRequest)
    - [ConfirmDepositResponse](#axelarnetwork.axelarnet.v1beta1.ConfirmDepositResponse)
    - [ExecutePendingTransfersRequest](#axelarnetwork.axelarnet.v1beta1.ExecutePendingTransfersRequest)
    - [ExecutePendingTransfersResponse](#axelarnetwork.axelarnet.v1beta1.ExecutePendingTransfersResponse)
    - [LinkRequest](#axelarnetwork.axelarnet.v1beta1.LinkRequest)
    - [LinkResponse](#axelarnetwork.axelarnet.v1beta1.LinkResponse)
    - [RegisterAssetRequest](#axelarnetwork.axelarnet.v1beta1.RegisterAssetRequest)
    - [RegisterAssetResponse](#axelarnetwork.axelarnet.v1beta1.RegisterAssetResponse)
    - [RegisterFeeCollectorRequest](#axelarnetwork.axelarnet.v1beta1.RegisterFeeCollectorRequest)
    - [RegisterFeeCollectorResponse](#axelarnetwork.axelarnet.v1beta1.RegisterFeeCollectorResponse)
    - [RegisterIBCPathRequest](#axelarnetwork.axelarnet.v1beta1.RegisterIBCPathRequest)
    - [RegisterIBCPathResponse](#axelarnetwork.axelarnet.v1beta1.RegisterIBCPathResponse)
    - [RouteIBCTransfersRequest](#axelarnetwork.axelarnet.v1beta1.RouteIBCTransfersRequest)
    - [RouteIBCTransfersResponse](#axelarnetwork.axelarnet.v1beta1.RouteIBCTransfersResponse)
  
- [axelarnetwork/axelarnet/v1beta1/service.proto](#axelarnetwork/axelarnet/v1beta1/service.proto)
    - [MsgService](#axelarnetwork.axelarnet.v1beta1.MsgService)
    - [QueryService](#axelarnetwork.axelarnet.v1beta1.QueryService)
  
- [axelarnetwork/bitcoin/v1beta1/types.proto](#axelarnetwork/bitcoin/v1beta1/types.proto)
    - [AddressInfo](#axelarnetwork.bitcoin.v1beta1.AddressInfo)
    - [AddressInfo.SpendingCondition](#axelarnetwork.bitcoin.v1beta1.AddressInfo.SpendingCondition)
    - [Network](#axelarnetwork.bitcoin.v1beta1.Network)
    - [OutPointInfo](#axelarnetwork.bitcoin.v1beta1.OutPointInfo)
    - [SignedTx](#axelarnetwork.bitcoin.v1beta1.SignedTx)
    - [UnsignedTx](#axelarnetwork.bitcoin.v1beta1.UnsignedTx)
    - [UnsignedTx.Info](#axelarnetwork.bitcoin.v1beta1.UnsignedTx.Info)
    - [UnsignedTx.Info.InputInfo](#axelarnetwork.bitcoin.v1beta1.UnsignedTx.Info.InputInfo)
    - [UnsignedTx.Info.InputInfo.SigRequirement](#axelarnetwork.bitcoin.v1beta1.UnsignedTx.Info.InputInfo.SigRequirement)
  
    - [AddressRole](#axelarnetwork.bitcoin.v1beta1.AddressRole)
    - [OutPointState](#axelarnetwork.bitcoin.v1beta1.OutPointState)
    - [TxStatus](#axelarnetwork.bitcoin.v1beta1.TxStatus)
    - [TxType](#axelarnetwork.bitcoin.v1beta1.TxType)
  
- [axelarnetwork/bitcoin/v1beta1/params.proto](#axelarnetwork/bitcoin/v1beta1/params.proto)
    - [Params](#axelarnetwork.bitcoin.v1beta1.Params)
  
- [axelarnetwork/bitcoin/v1beta1/genesis.proto](#axelarnetwork/bitcoin/v1beta1/genesis.proto)
    - [GenesisState](#axelarnetwork.bitcoin.v1beta1.GenesisState)
  
- [axelarnetwork/bitcoin/v1beta1/query.proto](#axelarnetwork/bitcoin/v1beta1/query.proto)
    - [DepositQueryParams](#axelarnetwork.bitcoin.v1beta1.DepositQueryParams)
    - [QueryAddressResponse](#axelarnetwork.bitcoin.v1beta1.QueryAddressResponse)
    - [QueryDepositStatusResponse](#axelarnetwork.bitcoin.v1beta1.QueryDepositStatusResponse)
    - [QueryTxResponse](#axelarnetwork.bitcoin.v1beta1.QueryTxResponse)
    - [QueryTxResponse.SigningInfo](#axelarnetwork.bitcoin.v1beta1.QueryTxResponse.SigningInfo)
  
- [axelarnetwork/snapshot/exported/v1beta1/types.proto](#axelarnetwork/snapshot/exported/v1beta1/types.proto)
    - [Snapshot](#axelarnetwork.snapshot.exported.v1beta1.Snapshot)
    - [Validator](#axelarnetwork.snapshot.exported.v1beta1.Validator)
  
    - [ValidatorIllegibility](#axelarnetwork.snapshot.exported.v1beta1.ValidatorIllegibility)
  
- [axelarnetwork/vote/exported/v1beta1/types.proto](#axelarnetwork/vote/exported/v1beta1/types.proto)
    - [PollKey](#axelarnetwork.vote.exported.v1beta1.PollKey)
    - [PollMetadata](#axelarnetwork.vote.exported.v1beta1.PollMetadata)
    - [Vote](#axelarnetwork.vote.exported.v1beta1.Vote)
    - [Voter](#axelarnetwork.vote.exported.v1beta1.Voter)
  
    - [PollState](#axelarnetwork.vote.exported.v1beta1.PollState)
  
- [axelarnetwork/bitcoin/v1beta1/tx.proto](#axelarnetwork/bitcoin/v1beta1/tx.proto)
    - [ConfirmOutpointRequest](#axelarnetwork.bitcoin.v1beta1.ConfirmOutpointRequest)
    - [ConfirmOutpointResponse](#axelarnetwork.bitcoin.v1beta1.ConfirmOutpointResponse)
    - [CreateMasterTxRequest](#axelarnetwork.bitcoin.v1beta1.CreateMasterTxRequest)
    - [CreateMasterTxResponse](#axelarnetwork.bitcoin.v1beta1.CreateMasterTxResponse)
    - [CreatePendingTransfersTxRequest](#axelarnetwork.bitcoin.v1beta1.CreatePendingTransfersTxRequest)
    - [CreatePendingTransfersTxResponse](#axelarnetwork.bitcoin.v1beta1.CreatePendingTransfersTxResponse)
    - [CreateRescueTxRequest](#axelarnetwork.bitcoin.v1beta1.CreateRescueTxRequest)
    - [CreateRescueTxResponse](#axelarnetwork.bitcoin.v1beta1.CreateRescueTxResponse)
    - [LinkRequest](#axelarnetwork.bitcoin.v1beta1.LinkRequest)
    - [LinkResponse](#axelarnetwork.bitcoin.v1beta1.LinkResponse)
    - [SignTxRequest](#axelarnetwork.bitcoin.v1beta1.SignTxRequest)
    - [SignTxResponse](#axelarnetwork.bitcoin.v1beta1.SignTxResponse)
    - [SubmitExternalSignatureRequest](#axelarnetwork.bitcoin.v1beta1.SubmitExternalSignatureRequest)
    - [SubmitExternalSignatureResponse](#axelarnetwork.bitcoin.v1beta1.SubmitExternalSignatureResponse)
    - [VoteConfirmOutpointRequest](#axelarnetwork.bitcoin.v1beta1.VoteConfirmOutpointRequest)
    - [VoteConfirmOutpointResponse](#axelarnetwork.bitcoin.v1beta1.VoteConfirmOutpointResponse)
  
- [axelarnetwork/bitcoin/v1beta1/service.proto](#axelarnetwork/bitcoin/v1beta1/service.proto)
    - [MsgService](#axelarnetwork.bitcoin.v1beta1.MsgService)
  
- [axelarnetwork/utils/v1beta1/queuer.proto](#axelarnetwork/utils/v1beta1/queuer.proto)
    - [QueueState](#axelarnetwork.utils.v1beta1.QueueState)
    - [QueueState.Item](#axelarnetwork.utils.v1beta1.QueueState.Item)
    - [QueueState.ItemsEntry](#axelarnetwork.utils.v1beta1.QueueState.ItemsEntry)
  
- [axelarnetwork/evm/v1beta1/types.proto](#axelarnetwork/evm/v1beta1/types.proto)
    - [Asset](#axelarnetwork.evm.v1beta1.Asset)
    - [BurnerInfo](#axelarnetwork.evm.v1beta1.BurnerInfo)
    - [Command](#axelarnetwork.evm.v1beta1.Command)
    - [CommandBatchMetadata](#axelarnetwork.evm.v1beta1.CommandBatchMetadata)
    - [ERC20Deposit](#axelarnetwork.evm.v1beta1.ERC20Deposit)
    - [ERC20TokenMetadata](#axelarnetwork.evm.v1beta1.ERC20TokenMetadata)
    - [Event](#axelarnetwork.evm.v1beta1.Event)
    - [EventContractCall](#axelarnetwork.evm.v1beta1.EventContractCall)
    - [EventContractCallWithToken](#axelarnetwork.evm.v1beta1.EventContractCallWithToken)
    - [EventMultisigOperatorshipTransferred](#axelarnetwork.evm.v1beta1.EventMultisigOperatorshipTransferred)
    - [EventMultisigOwnershipTransferred](#axelarnetwork.evm.v1beta1.EventMultisigOwnershipTransferred)
    - [EventSinglesigOperatorshipTransferred](#axelarnetwork.evm.v1beta1.EventSinglesigOperatorshipTransferred)
    - [EventSinglesigOwnershipTransferred](#axelarnetwork.evm.v1beta1.EventSinglesigOwnershipTransferred)
    - [EventTokenDeployed](#axelarnetwork.evm.v1beta1.EventTokenDeployed)
    - [EventTokenSent](#axelarnetwork.evm.v1beta1.EventTokenSent)
    - [EventTransfer](#axelarnetwork.evm.v1beta1.EventTransfer)
    - [Gateway](#axelarnetwork.evm.v1beta1.Gateway)
    - [NetworkInfo](#axelarnetwork.evm.v1beta1.NetworkInfo)
    - [SigMetadata](#axelarnetwork.evm.v1beta1.SigMetadata)
    - [TokenDetails](#axelarnetwork.evm.v1beta1.TokenDetails)
    - [TransactionMetadata](#axelarnetwork.evm.v1beta1.TransactionMetadata)
    - [TransferKey](#axelarnetwork.evm.v1beta1.TransferKey)
  
    - [BatchedCommandsStatus](#axelarnetwork.evm.v1beta1.BatchedCommandsStatus)
    - [DepositStatus](#axelarnetwork.evm.v1beta1.DepositStatus)
    - [Event.Status](#axelarnetwork.evm.v1beta1.Event.Status)
    - [Gateway.Status](#axelarnetwork.evm.v1beta1.Gateway.Status)
    - [SigType](#axelarnetwork.evm.v1beta1.SigType)
    - [Status](#axelarnetwork.evm.v1beta1.Status)
    - [TransferKeyType](#axelarnetwork.evm.v1beta1.TransferKeyType)
  
- [axelarnetwork/evm/v1beta1/params.proto](#axelarnetwork/evm/v1beta1/params.proto)
    - [Params](#axelarnetwork.evm.v1beta1.Params)
    - [PendingChain](#axelarnetwork.evm.v1beta1.PendingChain)
  
- [axelarnetwork/evm/v1beta1/genesis.proto](#axelarnetwork/evm/v1beta1/genesis.proto)
    - [GenesisState](#axelarnetwork.evm.v1beta1.GenesisState)
    - [GenesisState.Chain](#axelarnetwork.evm.v1beta1.GenesisState.Chain)
  
- [axelarnetwork/evm/v1beta1/query.proto](#axelarnetwork/evm/v1beta1/query.proto)
    - [BatchedCommandsRequest](#axelarnetwork.evm.v1beta1.BatchedCommandsRequest)
    - [BatchedCommandsResponse](#axelarnetwork.evm.v1beta1.BatchedCommandsResponse)
    - [BurnerInfoRequest](#axelarnetwork.evm.v1beta1.BurnerInfoRequest)
    - [BurnerInfoResponse](#axelarnetwork.evm.v1beta1.BurnerInfoResponse)
    - [BytecodeRequest](#axelarnetwork.evm.v1beta1.BytecodeRequest)
    - [BytecodeResponse](#axelarnetwork.evm.v1beta1.BytecodeResponse)
    - [ChainsRequest](#axelarnetwork.evm.v1beta1.ChainsRequest)
    - [ChainsResponse](#axelarnetwork.evm.v1beta1.ChainsResponse)
    - [ConfirmationHeightRequest](#axelarnetwork.evm.v1beta1.ConfirmationHeightRequest)
    - [ConfirmationHeightResponse](#axelarnetwork.evm.v1beta1.ConfirmationHeightResponse)
    - [DepositQueryParams](#axelarnetwork.evm.v1beta1.DepositQueryParams)
    - [DepositStateRequest](#axelarnetwork.evm.v1beta1.DepositStateRequest)
    - [DepositStateResponse](#axelarnetwork.evm.v1beta1.DepositStateResponse)
    - [EventRequest](#axelarnetwork.evm.v1beta1.EventRequest)
    - [EventResponse](#axelarnetwork.evm.v1beta1.EventResponse)
    - [GatewayAddressRequest](#axelarnetwork.evm.v1beta1.GatewayAddressRequest)
    - [GatewayAddressResponse](#axelarnetwork.evm.v1beta1.GatewayAddressResponse)
    - [KeyAddressRequest](#axelarnetwork.evm.v1beta1.KeyAddressRequest)
    - [KeyAddressResponse](#axelarnetwork.evm.v1beta1.KeyAddressResponse)
    - [KeyAddressResponse.MultisigAddresses](#axelarnetwork.evm.v1beta1.KeyAddressResponse.MultisigAddresses)
    - [KeyAddressResponse.ThresholdAddress](#axelarnetwork.evm.v1beta1.KeyAddressResponse.ThresholdAddress)
    - [PendingCommandsRequest](#axelarnetwork.evm.v1beta1.PendingCommandsRequest)
    - [PendingCommandsResponse](#axelarnetwork.evm.v1beta1.PendingCommandsResponse)
    - [QueryBurnerAddressResponse](#axelarnetwork.evm.v1beta1.QueryBurnerAddressResponse)
    - [QueryCommandResponse](#axelarnetwork.evm.v1beta1.QueryCommandResponse)
    - [QueryCommandResponse.ParamsEntry](#axelarnetwork.evm.v1beta1.QueryCommandResponse.ParamsEntry)
    - [QueryDepositStateParams](#axelarnetwork.evm.v1beta1.QueryDepositStateParams)
    - [QueryTokenAddressResponse](#axelarnetwork.evm.v1beta1.QueryTokenAddressResponse)
  
- [axelarnetwork/evm/v1beta1/tx.proto](#axelarnetwork/evm/v1beta1/tx.proto)
    - [AddChainRequest](#axelarnetwork.evm.v1beta1.AddChainRequest)
    - [AddChainResponse](#axelarnetwork.evm.v1beta1.AddChainResponse)
    - [ConfirmChainRequest](#axelarnetwork.evm.v1beta1.ConfirmChainRequest)
    - [ConfirmChainResponse](#axelarnetwork.evm.v1beta1.ConfirmChainResponse)
    - [ConfirmDepositRequest](#axelarnetwork.evm.v1beta1.ConfirmDepositRequest)
    - [ConfirmDepositResponse](#axelarnetwork.evm.v1beta1.ConfirmDepositResponse)
    - [ConfirmGatewayTxRequest](#axelarnetwork.evm.v1beta1.ConfirmGatewayTxRequest)
    - [ConfirmGatewayTxResponse](#axelarnetwork.evm.v1beta1.ConfirmGatewayTxResponse)
    - [ConfirmTokenRequest](#axelarnetwork.evm.v1beta1.ConfirmTokenRequest)
    - [ConfirmTokenResponse](#axelarnetwork.evm.v1beta1.ConfirmTokenResponse)
    - [ConfirmTransferKeyRequest](#axelarnetwork.evm.v1beta1.ConfirmTransferKeyRequest)
    - [ConfirmTransferKeyResponse](#axelarnetwork.evm.v1beta1.ConfirmTransferKeyResponse)
    - [CreateBurnTokensRequest](#axelarnetwork.evm.v1beta1.CreateBurnTokensRequest)
    - [CreateBurnTokensResponse](#axelarnetwork.evm.v1beta1.CreateBurnTokensResponse)
    - [CreateDeployTokenRequest](#axelarnetwork.evm.v1beta1.CreateDeployTokenRequest)
    - [CreateDeployTokenResponse](#axelarnetwork.evm.v1beta1.CreateDeployTokenResponse)
    - [CreatePendingTransfersRequest](#axelarnetwork.evm.v1beta1.CreatePendingTransfersRequest)
    - [CreatePendingTransfersResponse](#axelarnetwork.evm.v1beta1.CreatePendingTransfersResponse)
    - [CreateTransferOperatorshipRequest](#axelarnetwork.evm.v1beta1.CreateTransferOperatorshipRequest)
    - [CreateTransferOperatorshipResponse](#axelarnetwork.evm.v1beta1.CreateTransferOperatorshipResponse)
    - [CreateTransferOwnershipRequest](#axelarnetwork.evm.v1beta1.CreateTransferOwnershipRequest)
    - [CreateTransferOwnershipResponse](#axelarnetwork.evm.v1beta1.CreateTransferOwnershipResponse)
    - [LinkRequest](#axelarnetwork.evm.v1beta1.LinkRequest)
    - [LinkResponse](#axelarnetwork.evm.v1beta1.LinkResponse)
    - [SetGatewayRequest](#axelarnetwork.evm.v1beta1.SetGatewayRequest)
    - [SetGatewayResponse](#axelarnetwork.evm.v1beta1.SetGatewayResponse)
    - [SignCommandsRequest](#axelarnetwork.evm.v1beta1.SignCommandsRequest)
    - [SignCommandsResponse](#axelarnetwork.evm.v1beta1.SignCommandsResponse)
    - [VoteConfirmChainRequest](#axelarnetwork.evm.v1beta1.VoteConfirmChainRequest)
    - [VoteConfirmChainResponse](#axelarnetwork.evm.v1beta1.VoteConfirmChainResponse)
  
- [axelarnetwork/evm/v1beta1/service.proto](#axelarnetwork/evm/v1beta1/service.proto)
    - [MsgService](#axelarnetwork.evm.v1beta1.MsgService)
    - [QueryService](#axelarnetwork.evm.v1beta1.QueryService)
  
- [axelarnetwork/nexus/v1beta1/params.proto](#axelarnetwork/nexus/v1beta1/params.proto)
    - [Params](#axelarnetwork.nexus.v1beta1.Params)
  
- [axelarnetwork/nexus/v1beta1/genesis.proto](#axelarnetwork/nexus/v1beta1/genesis.proto)
    - [GenesisState](#axelarnetwork.nexus.v1beta1.GenesisState)
  
- [axelarnetwork/nexus/v1beta1/tx.proto](#axelarnetwork/nexus/v1beta1/tx.proto)
    - [ActivateChainRequest](#axelarnetwork.nexus.v1beta1.ActivateChainRequest)
    - [ActivateChainResponse](#axelarnetwork.nexus.v1beta1.ActivateChainResponse)
    - [DeactivateChainRequest](#axelarnetwork.nexus.v1beta1.DeactivateChainRequest)
    - [DeactivateChainResponse](#axelarnetwork.nexus.v1beta1.DeactivateChainResponse)
    - [DeregisterChainMaintainerRequest](#axelarnetwork.nexus.v1beta1.DeregisterChainMaintainerRequest)
    - [DeregisterChainMaintainerResponse](#axelarnetwork.nexus.v1beta1.DeregisterChainMaintainerResponse)
    - [RegisterAssetFeeRequest](#axelarnetwork.nexus.v1beta1.RegisterAssetFeeRequest)
    - [RegisterAssetFeeResponse](#axelarnetwork.nexus.v1beta1.RegisterAssetFeeResponse)
    - [RegisterChainMaintainerRequest](#axelarnetwork.nexus.v1beta1.RegisterChainMaintainerRequest)
    - [RegisterChainMaintainerResponse](#axelarnetwork.nexus.v1beta1.RegisterChainMaintainerResponse)
  
- [axelarnetwork/nexus/v1beta1/service.proto](#axelarnetwork/nexus/v1beta1/service.proto)
    - [MsgService](#axelarnetwork.nexus.v1beta1.MsgService)
    - [QueryService](#axelarnetwork.nexus.v1beta1.QueryService)
  
- [axelarnetwork/permission/v1beta1/types.proto](#axelarnetwork/permission/v1beta1/types.proto)
    - [GovAccount](#axelarnetwork.permission.v1beta1.GovAccount)
  
- [axelarnetwork/permission/v1beta1/params.proto](#axelarnetwork/permission/v1beta1/params.proto)
    - [Params](#axelarnetwork.permission.v1beta1.Params)
  
- [axelarnetwork/permission/v1beta1/genesis.proto](#axelarnetwork/permission/v1beta1/genesis.proto)
    - [GenesisState](#axelarnetwork.permission.v1beta1.GenesisState)
  
- [axelarnetwork/permission/v1beta1/query.proto](#axelarnetwork/permission/v1beta1/query.proto)
    - [QueryGovernanceKeyRequest](#axelarnetwork.permission.v1beta1.QueryGovernanceKeyRequest)
    - [QueryGovernanceKeyResponse](#axelarnetwork.permission.v1beta1.QueryGovernanceKeyResponse)
  
- [axelarnetwork/permission/v1beta1/tx.proto](#axelarnetwork/permission/v1beta1/tx.proto)
    - [DeregisterControllerRequest](#axelarnetwork.permission.v1beta1.DeregisterControllerRequest)
    - [DeregisterControllerResponse](#axelarnetwork.permission.v1beta1.DeregisterControllerResponse)
    - [RegisterControllerRequest](#axelarnetwork.permission.v1beta1.RegisterControllerRequest)
    - [RegisterControllerResponse](#axelarnetwork.permission.v1beta1.RegisterControllerResponse)
    - [UpdateGovernanceKeyRequest](#axelarnetwork.permission.v1beta1.UpdateGovernanceKeyRequest)
    - [UpdateGovernanceKeyResponse](#axelarnetwork.permission.v1beta1.UpdateGovernanceKeyResponse)
  
- [axelarnetwork/permission/v1beta1/service.proto](#axelarnetwork/permission/v1beta1/service.proto)
    - [Msg](#axelarnetwork.permission.v1beta1.Msg)
    - [Query](#axelarnetwork.permission.v1beta1.Query)
  
- [axelarnetwork/reward/v1beta1/params.proto](#axelarnetwork/reward/v1beta1/params.proto)
    - [Params](#axelarnetwork.reward.v1beta1.Params)
  
- [axelarnetwork/reward/v1beta1/types.proto](#axelarnetwork/reward/v1beta1/types.proto)
    - [Pool](#axelarnetwork.reward.v1beta1.Pool)
    - [Pool.Reward](#axelarnetwork.reward.v1beta1.Pool.Reward)
    - [Refund](#axelarnetwork.reward.v1beta1.Refund)
  
- [axelarnetwork/reward/v1beta1/genesis.proto](#axelarnetwork/reward/v1beta1/genesis.proto)
    - [GenesisState](#axelarnetwork.reward.v1beta1.GenesisState)
  
- [axelarnetwork/reward/v1beta1/tx.proto](#axelarnetwork/reward/v1beta1/tx.proto)
    - [RefundMsgRequest](#axelarnetwork.reward.v1beta1.RefundMsgRequest)
    - [RefundMsgResponse](#axelarnetwork.reward.v1beta1.RefundMsgResponse)
  
- [axelarnetwork/reward/v1beta1/service.proto](#axelarnetwork/reward/v1beta1/service.proto)
    - [MsgService](#axelarnetwork.reward.v1beta1.MsgService)
  
- [axelarnetwork/snapshot/v1beta1/params.proto](#axelarnetwork/snapshot/v1beta1/params.proto)
    - [Params](#axelarnetwork.snapshot.v1beta1.Params)
  
- [axelarnetwork/snapshot/v1beta1/types.proto](#axelarnetwork/snapshot/v1beta1/types.proto)
    - [ProxiedValidator](#axelarnetwork.snapshot.v1beta1.ProxiedValidator)
  
- [axelarnetwork/snapshot/v1beta1/genesis.proto](#axelarnetwork/snapshot/v1beta1/genesis.proto)
    - [GenesisState](#axelarnetwork.snapshot.v1beta1.GenesisState)
  
- [axelarnetwork/snapshot/v1beta1/query.proto](#axelarnetwork/snapshot/v1beta1/query.proto)
    - [QueryValidatorsResponse](#axelarnetwork.snapshot.v1beta1.QueryValidatorsResponse)
    - [QueryValidatorsResponse.TssIllegibilityInfo](#axelarnetwork.snapshot.v1beta1.QueryValidatorsResponse.TssIllegibilityInfo)
    - [QueryValidatorsResponse.Validator](#axelarnetwork.snapshot.v1beta1.QueryValidatorsResponse.Validator)
  
- [axelarnetwork/snapshot/v1beta1/tx.proto](#axelarnetwork/snapshot/v1beta1/tx.proto)
    - [DeactivateProxyRequest](#axelarnetwork.snapshot.v1beta1.DeactivateProxyRequest)
    - [DeactivateProxyResponse](#axelarnetwork.snapshot.v1beta1.DeactivateProxyResponse)
    - [RegisterProxyRequest](#axelarnetwork.snapshot.v1beta1.RegisterProxyRequest)
    - [RegisterProxyResponse](#axelarnetwork.snapshot.v1beta1.RegisterProxyResponse)
  
- [axelarnetwork/snapshot/v1beta1/service.proto](#axelarnetwork/snapshot/v1beta1/service.proto)
    - [MsgService](#axelarnetwork.snapshot.v1beta1.MsgService)
  
- [axelarnetwork/tss/tofnd/v1beta1/common.proto](#axelarnetwork/tss/tofnd/v1beta1/common.proto)
    - [KeyPresenceRequest](#axelarnetwork.tss.tofnd.v1beta1.KeyPresenceRequest)
    - [KeyPresenceResponse](#axelarnetwork.tss.tofnd.v1beta1.KeyPresenceResponse)
  
    - [KeyPresenceResponse.Response](#axelarnetwork.tss.tofnd.v1beta1.KeyPresenceResponse.Response)
  
- [axelarnetwork/tss/tofnd/v1beta1/multisig.proto](#axelarnetwork/tss/tofnd/v1beta1/multisig.proto)
    - [KeygenRequest](#axelarnetwork.tss.tofnd.v1beta1.KeygenRequest)
    - [KeygenResponse](#axelarnetwork.tss.tofnd.v1beta1.KeygenResponse)
    - [SignRequest](#axelarnetwork.tss.tofnd.v1beta1.SignRequest)
    - [SignResponse](#axelarnetwork.tss.tofnd.v1beta1.SignResponse)
  
- [axelarnetwork/tss/tofnd/v1beta1/tofnd.proto](#axelarnetwork/tss/tofnd/v1beta1/tofnd.proto)
    - [KeygenInit](#axelarnetwork.tss.tofnd.v1beta1.KeygenInit)
    - [KeygenOutput](#axelarnetwork.tss.tofnd.v1beta1.KeygenOutput)
    - [MessageIn](#axelarnetwork.tss.tofnd.v1beta1.MessageIn)
    - [MessageOut](#axelarnetwork.tss.tofnd.v1beta1.MessageOut)
    - [MessageOut.CriminalList](#axelarnetwork.tss.tofnd.v1beta1.MessageOut.CriminalList)
    - [MessageOut.CriminalList.Criminal](#axelarnetwork.tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal)
    - [MessageOut.KeygenResult](#axelarnetwork.tss.tofnd.v1beta1.MessageOut.KeygenResult)
    - [MessageOut.SignResult](#axelarnetwork.tss.tofnd.v1beta1.MessageOut.SignResult)
    - [RecoverRequest](#axelarnetwork.tss.tofnd.v1beta1.RecoverRequest)
    - [RecoverResponse](#axelarnetwork.tss.tofnd.v1beta1.RecoverResponse)
    - [SignInit](#axelarnetwork.tss.tofnd.v1beta1.SignInit)
    - [TrafficIn](#axelarnetwork.tss.tofnd.v1beta1.TrafficIn)
    - [TrafficOut](#axelarnetwork.tss.tofnd.v1beta1.TrafficOut)
  
    - [MessageOut.CriminalList.Criminal.CrimeType](#axelarnetwork.tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal.CrimeType)
    - [RecoverResponse.Response](#axelarnetwork.tss.tofnd.v1beta1.RecoverResponse.Response)
  
- [axelarnetwork/tss/v1beta1/params.proto](#axelarnetwork/tss/v1beta1/params.proto)
    - [Params](#axelarnetwork.tss.v1beta1.Params)
  
- [axelarnetwork/tss/v1beta1/types.proto](#axelarnetwork/tss/v1beta1/types.proto)
    - [ExternalKeys](#axelarnetwork.tss.v1beta1.ExternalKeys)
    - [KeyInfo](#axelarnetwork.tss.v1beta1.KeyInfo)
    - [KeyRecoveryInfo](#axelarnetwork.tss.v1beta1.KeyRecoveryInfo)
    - [KeyRecoveryInfo.PrivateEntry](#axelarnetwork.tss.v1beta1.KeyRecoveryInfo.PrivateEntry)
    - [KeygenVoteData](#axelarnetwork.tss.v1beta1.KeygenVoteData)
    - [MultisigInfo](#axelarnetwork.tss.v1beta1.MultisigInfo)
    - [MultisigInfo.Info](#axelarnetwork.tss.v1beta1.MultisigInfo.Info)
    - [ValidatorStatus](#axelarnetwork.tss.v1beta1.ValidatorStatus)
  
- [axelarnetwork/tss/v1beta1/genesis.proto](#axelarnetwork/tss/v1beta1/genesis.proto)
    - [GenesisState](#axelarnetwork.tss.v1beta1.GenesisState)
  
- [axelarnetwork/tss/v1beta1/query.proto](#axelarnetwork/tss/v1beta1/query.proto)
    - [AssignableKeyRequest](#axelarnetwork.tss.v1beta1.AssignableKeyRequest)
    - [AssignableKeyResponse](#axelarnetwork.tss.v1beta1.AssignableKeyResponse)
    - [NextKeyIDRequest](#axelarnetwork.tss.v1beta1.NextKeyIDRequest)
    - [NextKeyIDResponse](#axelarnetwork.tss.v1beta1.NextKeyIDResponse)
    - [QueryActiveOldKeysResponse](#axelarnetwork.tss.v1beta1.QueryActiveOldKeysResponse)
    - [QueryActiveOldKeysValidatorResponse](#axelarnetwork.tss.v1beta1.QueryActiveOldKeysValidatorResponse)
    - [QueryActiveOldKeysValidatorResponse.KeyInfo](#axelarnetwork.tss.v1beta1.QueryActiveOldKeysValidatorResponse.KeyInfo)
    - [QueryDeactivatedOperatorsResponse](#axelarnetwork.tss.v1beta1.QueryDeactivatedOperatorsResponse)
    - [QueryExternalKeyIDResponse](#axelarnetwork.tss.v1beta1.QueryExternalKeyIDResponse)
    - [QueryKeyResponse](#axelarnetwork.tss.v1beta1.QueryKeyResponse)
    - [QueryKeyResponse.ECDSAKey](#axelarnetwork.tss.v1beta1.QueryKeyResponse.ECDSAKey)
    - [QueryKeyResponse.Key](#axelarnetwork.tss.v1beta1.QueryKeyResponse.Key)
    - [QueryKeyResponse.MultisigKey](#axelarnetwork.tss.v1beta1.QueryKeyResponse.MultisigKey)
    - [QueryKeyShareResponse](#axelarnetwork.tss.v1beta1.QueryKeyShareResponse)
    - [QueryKeyShareResponse.ShareInfo](#axelarnetwork.tss.v1beta1.QueryKeyShareResponse.ShareInfo)
    - [QueryRecoveryResponse](#axelarnetwork.tss.v1beta1.QueryRecoveryResponse)
    - [QuerySignatureResponse](#axelarnetwork.tss.v1beta1.QuerySignatureResponse)
    - [QuerySignatureResponse.MultisigSignature](#axelarnetwork.tss.v1beta1.QuerySignatureResponse.MultisigSignature)
    - [QuerySignatureResponse.Signature](#axelarnetwork.tss.v1beta1.QuerySignatureResponse.Signature)
    - [QuerySignatureResponse.ThresholdSignature](#axelarnetwork.tss.v1beta1.QuerySignatureResponse.ThresholdSignature)
  
    - [VoteStatus](#axelarnetwork.tss.v1beta1.VoteStatus)
  
- [axelarnetwork/tss/v1beta1/tx.proto](#axelarnetwork/tss/v1beta1/tx.proto)
    - [HeartBeatRequest](#axelarnetwork.tss.v1beta1.HeartBeatRequest)
    - [HeartBeatResponse](#axelarnetwork.tss.v1beta1.HeartBeatResponse)
    - [ProcessKeygenTrafficRequest](#axelarnetwork.tss.v1beta1.ProcessKeygenTrafficRequest)
    - [ProcessKeygenTrafficResponse](#axelarnetwork.tss.v1beta1.ProcessKeygenTrafficResponse)
    - [ProcessSignTrafficRequest](#axelarnetwork.tss.v1beta1.ProcessSignTrafficRequest)
    - [ProcessSignTrafficResponse](#axelarnetwork.tss.v1beta1.ProcessSignTrafficResponse)
    - [RegisterExternalKeysRequest](#axelarnetwork.tss.v1beta1.RegisterExternalKeysRequest)
    - [RegisterExternalKeysRequest.ExternalKey](#axelarnetwork.tss.v1beta1.RegisterExternalKeysRequest.ExternalKey)
    - [RegisterExternalKeysResponse](#axelarnetwork.tss.v1beta1.RegisterExternalKeysResponse)
    - [RotateKeyRequest](#axelarnetwork.tss.v1beta1.RotateKeyRequest)
    - [RotateKeyResponse](#axelarnetwork.tss.v1beta1.RotateKeyResponse)
    - [StartKeygenRequest](#axelarnetwork.tss.v1beta1.StartKeygenRequest)
    - [StartKeygenResponse](#axelarnetwork.tss.v1beta1.StartKeygenResponse)
    - [SubmitMultisigPubKeysRequest](#axelarnetwork.tss.v1beta1.SubmitMultisigPubKeysRequest)
    - [SubmitMultisigPubKeysResponse](#axelarnetwork.tss.v1beta1.SubmitMultisigPubKeysResponse)
    - [SubmitMultisigSignaturesRequest](#axelarnetwork.tss.v1beta1.SubmitMultisigSignaturesRequest)
    - [SubmitMultisigSignaturesResponse](#axelarnetwork.tss.v1beta1.SubmitMultisigSignaturesResponse)
    - [VotePubKeyRequest](#axelarnetwork.tss.v1beta1.VotePubKeyRequest)
    - [VotePubKeyResponse](#axelarnetwork.tss.v1beta1.VotePubKeyResponse)
    - [VoteSigRequest](#axelarnetwork.tss.v1beta1.VoteSigRequest)
    - [VoteSigResponse](#axelarnetwork.tss.v1beta1.VoteSigResponse)
  
- [axelarnetwork/tss/v1beta1/service.proto](#axelarnetwork/tss/v1beta1/service.proto)
    - [MsgService](#axelarnetwork.tss.v1beta1.MsgService)
    - [QueryService](#axelarnetwork.tss.v1beta1.QueryService)
  
- [axelarnetwork/vote/v1beta1/params.proto](#axelarnetwork/vote/v1beta1/params.proto)
    - [Params](#axelarnetwork.vote.v1beta1.Params)
  
- [axelarnetwork/vote/v1beta1/genesis.proto](#axelarnetwork/vote/v1beta1/genesis.proto)
    - [GenesisState](#axelarnetwork.vote.v1beta1.GenesisState)
  
- [axelarnetwork/vote/v1beta1/tx.proto](#axelarnetwork/vote/v1beta1/tx.proto)
    - [VoteRequest](#axelarnetwork.vote.v1beta1.VoteRequest)
    - [VoteResponse](#axelarnetwork.vote.v1beta1.VoteResponse)
  
- [axelarnetwork/vote/v1beta1/service.proto](#axelarnetwork/vote/v1beta1/service.proto)
    - [MsgService](#axelarnetwork.vote.v1beta1.MsgService)
  
- [axelarnetwork/vote/v1beta1/types.proto](#axelarnetwork/vote/v1beta1/types.proto)
    - [TalliedVote](#axelarnetwork.vote.v1beta1.TalliedVote)
  
- [Scalar Value Types](#scalar-value-types)



<a name="axelarnetwork/axelarnet/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/axelarnet/v1beta1/params.proto



<a name="axelarnetwork.axelarnet.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `route_timeout_window` | [uint64](#uint64) |  | IBC packet route timeout window |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/axelarnet/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/axelarnet/v1beta1/types.proto



<a name="axelarnetwork.axelarnet.v1beta1.Asset"></a>

### Asset



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `min_amount` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.axelarnet.v1beta1.CosmosChain"></a>

### CosmosChain



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `ibc_path` | [string](#string) |  |  |
| `assets` | [Asset](#axelarnetwork.axelarnet.v1beta1.Asset) | repeated | **Deprecated.**  |
| `addr_prefix` | [string](#string) |  |  |






<a name="axelarnetwork.axelarnet.v1beta1.IBCTransfer"></a>

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



<a name="axelarnetwork/axelarnet/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/axelarnet/v1beta1/genesis.proto



<a name="axelarnetwork.axelarnet.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelarnetwork.axelarnet.v1beta1.Params) |  |  |
| `collector_address` | [bytes](#bytes) |  |  |
| `chains` | [CosmosChain](#axelarnetwork.axelarnet.v1beta1.CosmosChain) | repeated |  |
| `pending_transfers` | [IBCTransfer](#axelarnetwork.axelarnet.v1beta1.IBCTransfer) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/utils/v1beta1/threshold.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/utils/v1beta1/threshold.proto



<a name="axelarnetwork.utils.v1beta1.Threshold"></a>

### Threshold



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `numerator` | [int64](#int64) |  | split threshold into Numerator and denominator to avoid floating point errors down the line |
| `denominator` | [int64](#int64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/tss/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/tss/exported/v1beta1/types.proto



<a name="axelarnetwork.tss.exported.v1beta1.Key"></a>

### Key



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `role` | [KeyRole](#axelarnetwork.tss.exported.v1beta1.KeyRole) |  |  |
| `type` | [KeyType](#axelarnetwork.tss.exported.v1beta1.KeyType) |  |  |
| `ecdsa_key` | [Key.ECDSAKey](#axelarnetwork.tss.exported.v1beta1.Key.ECDSAKey) |  |  |
| `multisig_key` | [Key.MultisigKey](#axelarnetwork.tss.exported.v1beta1.Key.MultisigKey) |  |  |
| `rotated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `rotation_count` | [int64](#int64) |  |  |
| `chain` | [string](#string) |  |  |
| `snapshot_counter` | [int64](#int64) |  |  |






<a name="axelarnetwork.tss.exported.v1beta1.Key.ECDSAKey"></a>

### Key.ECDSAKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.tss.exported.v1beta1.Key.MultisigKey"></a>

### Key.MultisigKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `values` | [bytes](#bytes) | repeated |  |
| `threshold` | [int64](#int64) |  |  |






<a name="axelarnetwork.tss.exported.v1beta1.KeyRequirement"></a>

### KeyRequirement
KeyRequirement defines requirements for keys


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_role` | [KeyRole](#axelarnetwork.tss.exported.v1beta1.KeyRole) |  |  |
| `key_type` | [KeyType](#axelarnetwork.tss.exported.v1beta1.KeyType) |  |  |
| `min_keygen_threshold` | [axelarnetwork.utils.v1beta1.Threshold](#axelarnetwork.utils.v1beta1.Threshold) |  |  |
| `safety_threshold` | [axelarnetwork.utils.v1beta1.Threshold](#axelarnetwork.utils.v1beta1.Threshold) |  |  |
| `key_share_distribution_policy` | [KeyShareDistributionPolicy](#axelarnetwork.tss.exported.v1beta1.KeyShareDistributionPolicy) |  |  |
| `max_total_share_count` | [int64](#int64) |  |  |
| `min_total_share_count` | [int64](#int64) |  |  |
| `keygen_voting_threshold` | [axelarnetwork.utils.v1beta1.Threshold](#axelarnetwork.utils.v1beta1.Threshold) |  |  |
| `sign_voting_threshold` | [axelarnetwork.utils.v1beta1.Threshold](#axelarnetwork.utils.v1beta1.Threshold) |  |  |
| `keygen_timeout` | [int64](#int64) |  |  |
| `sign_timeout` | [int64](#int64) |  |  |






<a name="axelarnetwork.tss.exported.v1beta1.SigKeyPair"></a>

### SigKeyPair
PubKeyInfo holds a pubkey and a signature


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pub_key` | [bytes](#bytes) |  |  |
| `signature` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.tss.exported.v1beta1.SignInfo"></a>

### SignInfo
SignInfo holds information about a sign request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `sig_id` | [string](#string) |  |  |
| `msg` | [bytes](#bytes) |  |  |
| `snapshot_counter` | [int64](#int64) |  |  |
| `request_module` | [string](#string) |  |  |
| `metadata` | [string](#string) |  |  |






<a name="axelarnetwork.tss.exported.v1beta1.Signature"></a>

### Signature
Signature holds public key and ECDSA signature


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sig_id` | [string](#string) |  |  |
| `single_sig` | [Signature.SingleSig](#axelarnetwork.tss.exported.v1beta1.Signature.SingleSig) |  |  |
| `multi_sig` | [Signature.MultiSig](#axelarnetwork.tss.exported.v1beta1.Signature.MultiSig) |  |  |
| `sig_status` | [SigStatus](#axelarnetwork.tss.exported.v1beta1.SigStatus) |  |  |






<a name="axelarnetwork.tss.exported.v1beta1.Signature.MultiSig"></a>

### Signature.MultiSig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sig_key_pairs` | [SigKeyPair](#axelarnetwork.tss.exported.v1beta1.SigKeyPair) | repeated |  |






<a name="axelarnetwork.tss.exported.v1beta1.Signature.SingleSig"></a>

### Signature.SingleSig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sig_key_pair` | [SigKeyPair](#axelarnetwork.tss.exported.v1beta1.SigKeyPair) |  |  |





 <!-- end messages -->


<a name="axelarnetwork.tss.exported.v1beta1.AckType"></a>

### AckType


| Name | Number | Description |
| ---- | ------ | ----------- |
| ACK_TYPE_UNSPECIFIED | 0 |  |
| ACK_TYPE_KEYGEN | 1 |  |
| ACK_TYPE_SIGN | 2 |  |



<a name="axelarnetwork.tss.exported.v1beta1.KeyRole"></a>

### KeyRole


| Name | Number | Description |
| ---- | ------ | ----------- |
| KEY_ROLE_UNSPECIFIED | 0 |  |
| KEY_ROLE_MASTER_KEY | 1 |  |
| KEY_ROLE_SECONDARY_KEY | 2 |  |
| KEY_ROLE_EXTERNAL_KEY | 3 |  |



<a name="axelarnetwork.tss.exported.v1beta1.KeyShareDistributionPolicy"></a>

### KeyShareDistributionPolicy


| Name | Number | Description |
| ---- | ------ | ----------- |
| KEY_SHARE_DISTRIBUTION_POLICY_UNSPECIFIED | 0 |  |
| KEY_SHARE_DISTRIBUTION_POLICY_WEIGHTED_BY_STAKE | 1 |  |
| KEY_SHARE_DISTRIBUTION_POLICY_ONE_PER_VALIDATOR | 2 |  |



<a name="axelarnetwork.tss.exported.v1beta1.KeyType"></a>

### KeyType


| Name | Number | Description |
| ---- | ------ | ----------- |
| KEY_TYPE_UNSPECIFIED | 0 |  |
| KEY_TYPE_NONE | 1 |  |
| KEY_TYPE_THRESHOLD | 2 |  |
| KEY_TYPE_MULTISIG | 3 |  |



<a name="axelarnetwork.tss.exported.v1beta1.SigStatus"></a>

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



<a name="axelarnetwork/nexus/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/nexus/exported/v1beta1/types.proto



<a name="axelarnetwork.nexus.exported.v1beta1.Asset"></a>

### Asset



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `is_native_asset` | [bool](#bool) |  |  |






<a name="axelarnetwork.nexus.exported.v1beta1.Chain"></a>

### Chain
Chain represents the properties of a registered blockchain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `supports_foreign_assets` | [bool](#bool) |  |  |
| `key_type` | [axelarnetwork.tss.exported.v1beta1.KeyType](#axelarnetwork.tss.exported.v1beta1.KeyType) |  |  |
| `module` | [string](#string) |  |  |






<a name="axelarnetwork.nexus.exported.v1beta1.CrossChainAddress"></a>

### CrossChainAddress
CrossChainAddress represents a generalized address on any registered chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [Chain](#axelarnetwork.nexus.exported.v1beta1.Chain) |  |  |
| `address` | [string](#string) |  |  |






<a name="axelarnetwork.nexus.exported.v1beta1.CrossChainTransfer"></a>

### CrossChainTransfer
CrossChainTransfer represents a generalized transfer of some asset to a
registered blockchain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `recipient` | [CrossChainAddress](#axelarnetwork.nexus.exported.v1beta1.CrossChainAddress) |  |  |
| `asset` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `id` | [uint64](#uint64) |  |  |
| `state` | [TransferState](#axelarnetwork.nexus.exported.v1beta1.TransferState) |  |  |






<a name="axelarnetwork.nexus.exported.v1beta1.FeeInfo"></a>

### FeeInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `asset` | [string](#string) |  |  |
| `fee_rate` | [bytes](#bytes) |  |  |
| `min_fee` | [bytes](#bytes) |  |  |
| `max_fee` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.nexus.exported.v1beta1.TransferFee"></a>

### TransferFee
TransferFee represents accumulated fees generated by the network


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |





 <!-- end messages -->


<a name="axelarnetwork.nexus.exported.v1beta1.TransferState"></a>

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



<a name="axelarnetwork/nexus/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/nexus/v1beta1/types.proto



<a name="axelarnetwork.nexus.v1beta1.ChainState"></a>

### ChainState
ChainState represents the state of a registered blockchain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [axelarnetwork.nexus.exported.v1beta1.Chain](#axelarnetwork.nexus.exported.v1beta1.Chain) |  |  |
| `maintainers` | [bytes](#bytes) | repeated |  |
| `activated` | [bool](#bool) |  |  |
| `assets` | [axelarnetwork.nexus.exported.v1beta1.Asset](#axelarnetwork.nexus.exported.v1beta1.Asset) | repeated |  |






<a name="axelarnetwork.nexus.v1beta1.LinkedAddresses"></a>

### LinkedAddresses



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposit_address` | [axelarnetwork.nexus.exported.v1beta1.CrossChainAddress](#axelarnetwork.nexus.exported.v1beta1.CrossChainAddress) |  |  |
| `recipient_address` | [axelarnetwork.nexus.exported.v1beta1.CrossChainAddress](#axelarnetwork.nexus.exported.v1beta1.CrossChainAddress) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/nexus/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/nexus/v1beta1/query.proto



<a name="axelarnetwork.nexus.v1beta1.AssetsRequest"></a>

### AssetsRequest
AssetsRequest represents a message that queries the registered assets of a
chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |






<a name="axelarnetwork.nexus.v1beta1.AssetsResponse"></a>

### AssetsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `assets` | [string](#string) | repeated |  |






<a name="axelarnetwork.nexus.v1beta1.ChainStateRequest"></a>

### ChainStateRequest
ChainStateRequest represents a message that queries the state of a chain
registered on the network


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |






<a name="axelarnetwork.nexus.v1beta1.ChainStateResponse"></a>

### ChainStateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `state` | [ChainState](#axelarnetwork.nexus.v1beta1.ChainState) |  |  |






<a name="axelarnetwork.nexus.v1beta1.ChainsByAssetRequest"></a>

### ChainsByAssetRequest
ChainsByAssetRequest represents a message that queries the chains
that support an asset on the network


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset` | [string](#string) |  |  |






<a name="axelarnetwork.nexus.v1beta1.ChainsByAssetResponse"></a>

### ChainsByAssetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chains` | [string](#string) | repeated |  |






<a name="axelarnetwork.nexus.v1beta1.ChainsRequest"></a>

### ChainsRequest
ChainsRequest represents a message that queries the chains
registered on the network






<a name="axelarnetwork.nexus.v1beta1.ChainsResponse"></a>

### ChainsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chains` | [string](#string) | repeated |  |






<a name="axelarnetwork.nexus.v1beta1.FeeRequest"></a>

### FeeRequest
FeeRequest represents a message that queries the transfer fees associated
to an asset on a chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `asset` | [string](#string) |  |  |






<a name="axelarnetwork.nexus.v1beta1.FeeResponse"></a>

### FeeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `fee_info` | [axelarnetwork.nexus.exported.v1beta1.FeeInfo](#axelarnetwork.nexus.exported.v1beta1.FeeInfo) |  |  |






<a name="axelarnetwork.nexus.v1beta1.LatestDepositAddressRequest"></a>

### LatestDepositAddressRequest
LatestDepositAddressRequest represents a message that queries a deposit
address by recipient address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `recipient_addr` | [string](#string) |  |  |
| `recipient_chain` | [string](#string) |  |  |
| `deposit_chain` | [string](#string) |  |  |






<a name="axelarnetwork.nexus.v1beta1.LatestDepositAddressResponse"></a>

### LatestDepositAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposit_addr` | [string](#string) |  |  |






<a name="axelarnetwork.nexus.v1beta1.QueryChainMaintainersResponse"></a>

### QueryChainMaintainersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `maintainers` | [bytes](#bytes) | repeated |  |






<a name="axelarnetwork.nexus.v1beta1.TransferFeeRequest"></a>

### TransferFeeRequest
TransferFeeRequest represents a message that queries the fees charged by
the network for a cross-chain transfer


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source_chain` | [string](#string) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `amount` | [string](#string) |  |  |






<a name="axelarnetwork.nexus.v1beta1.TransferFeeResponse"></a>

### TransferFeeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="axelarnetwork.nexus.v1beta1.TransfersForChainRequest"></a>

### TransfersForChainRequest
TransfersForChainRequest represents a message that queries the
transfers for the specified chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `state` | [axelarnetwork.nexus.exported.v1beta1.TransferState](#axelarnetwork.nexus.exported.v1beta1.TransferState) |  |  |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |






<a name="axelarnetwork.nexus.v1beta1.TransfersForChainResponse"></a>

### TransfersForChainResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `transfers` | [axelarnetwork.nexus.exported.v1beta1.CrossChainTransfer](#axelarnetwork.nexus.exported.v1beta1.CrossChainTransfer) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/axelarnet/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/axelarnet/v1beta1/query.proto



<a name="axelarnetwork.axelarnet.v1beta1.PendingIBCTransferCountRequest"></a>

### PendingIBCTransferCountRequest







<a name="axelarnetwork.axelarnet.v1beta1.PendingIBCTransferCountResponse"></a>

### PendingIBCTransferCountResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `transfers_by_chain` | [PendingIBCTransferCountResponse.TransfersByChainEntry](#axelarnetwork.axelarnet.v1beta1.PendingIBCTransferCountResponse.TransfersByChainEntry) | repeated |  |






<a name="axelarnetwork.axelarnet.v1beta1.PendingIBCTransferCountResponse.TransfersByChainEntry"></a>

### PendingIBCTransferCountResponse.TransfersByChainEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [uint32](#uint32) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/permission/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/permission/exported/v1beta1/types.proto


 <!-- end messages -->


<a name="axelarnetwork.permission.exported.v1beta1.Role"></a>

### Role


| Name | Number | Description |
| ---- | ------ | ----------- |
| ROLE_UNSPECIFIED | 0 |  |
| ROLE_UNRESTRICTED | 1 |  |
| ROLE_CHAIN_MANAGEMENT | 2 |  |
| ROLE_ACCESS_CONTROL | 3 |  |


 <!-- end enums -->


<a name="axelarnetwork/permission/exported/v1beta1/types.proto-extensions"></a>

### File-level Extensions
| Extension | Type | Base | Number | Description |
| --------- | ---- | ---- | ------ | ----------- |
| `permission_role` | Role | .google.protobuf.MessageOptions | 50000 | 50000-99999 reserved for use withing individual organizations |

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/axelarnet/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/axelarnet/v1beta1/tx.proto



<a name="axelarnetwork.axelarnet.v1beta1.AddCosmosBasedChainRequest"></a>

### AddCosmosBasedChainRequest
MsgAddCosmosBasedChain represents a message to register a cosmos based chain
to nexus


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [axelarnetwork.nexus.exported.v1beta1.Chain](#axelarnetwork.nexus.exported.v1beta1.Chain) |  |  |
| `addr_prefix` | [string](#string) |  |  |
| `native_assets` | [axelarnetwork.nexus.exported.v1beta1.Asset](#axelarnetwork.nexus.exported.v1beta1.Asset) | repeated |  |






<a name="axelarnetwork.axelarnet.v1beta1.AddCosmosBasedChainResponse"></a>

### AddCosmosBasedChainResponse







<a name="axelarnetwork.axelarnet.v1beta1.ConfirmDepositRequest"></a>

### ConfirmDepositRequest
MsgConfirmDeposit represents a deposit confirmation message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `deposit_address` | [bytes](#bytes) |  |  |
| `denom` | [string](#string) |  |  |






<a name="axelarnetwork.axelarnet.v1beta1.ConfirmDepositResponse"></a>

### ConfirmDepositResponse







<a name="axelarnetwork.axelarnet.v1beta1.ExecutePendingTransfersRequest"></a>

### ExecutePendingTransfersRequest
MsgExecutePendingTransfers represents a message to trigger transfer all
pending transfers


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.axelarnet.v1beta1.ExecutePendingTransfersResponse"></a>

### ExecutePendingTransfersResponse







<a name="axelarnetwork.axelarnet.v1beta1.LinkRequest"></a>

### LinkRequest
MsgLink represents a message to link a cross-chain address to an Axelar
address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `recipient_addr` | [string](#string) |  |  |
| `recipient_chain` | [string](#string) |  |  |
| `asset` | [string](#string) |  |  |






<a name="axelarnetwork.axelarnet.v1beta1.LinkResponse"></a>

### LinkResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposit_addr` | [string](#string) |  |  |






<a name="axelarnetwork.axelarnet.v1beta1.RegisterAssetRequest"></a>

### RegisterAssetRequest
RegisterAssetRequest represents a message to register an asset to a cosmos
based chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `asset` | [axelarnetwork.nexus.exported.v1beta1.Asset](#axelarnetwork.nexus.exported.v1beta1.Asset) |  |  |






<a name="axelarnetwork.axelarnet.v1beta1.RegisterAssetResponse"></a>

### RegisterAssetResponse







<a name="axelarnetwork.axelarnet.v1beta1.RegisterFeeCollectorRequest"></a>

### RegisterFeeCollectorRequest
RegisterFeeCollectorRequest represents a message to register axelarnet fee
collector account


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `fee_collector` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.axelarnet.v1beta1.RegisterFeeCollectorResponse"></a>

### RegisterFeeCollectorResponse







<a name="axelarnetwork.axelarnet.v1beta1.RegisterIBCPathRequest"></a>

### RegisterIBCPathRequest
MSgRegisterIBCPath represents a message to register an IBC tracing path for
a cosmos chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `path` | [string](#string) |  |  |






<a name="axelarnetwork.axelarnet.v1beta1.RegisterIBCPathResponse"></a>

### RegisterIBCPathResponse







<a name="axelarnetwork.axelarnet.v1beta1.RouteIBCTransfersRequest"></a>

### RouteIBCTransfersRequest
RouteIBCTransfersRequest represents a message to route pending transfers to
cosmos based chains


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.axelarnet.v1beta1.RouteIBCTransfersResponse"></a>

### RouteIBCTransfersResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/axelarnet/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/axelarnet/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelarnetwork.axelarnet.v1beta1.MsgService"></a>

### MsgService
Msg defines the axelarnet Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Link` | [LinkRequest](#axelarnetwork.axelarnet.v1beta1.LinkRequest) | [LinkResponse](#axelarnetwork.axelarnet.v1beta1.LinkResponse) |  | POST|/axelar/axelarnet/link/{recipient_chain}|
| `ConfirmDeposit` | [ConfirmDepositRequest](#axelarnetwork.axelarnet.v1beta1.ConfirmDepositRequest) | [ConfirmDepositResponse](#axelarnetwork.axelarnet.v1beta1.ConfirmDepositResponse) |  | POST|/axelar/axelarnet/confirm-deposit|
| `ExecutePendingTransfers` | [ExecutePendingTransfersRequest](#axelarnetwork.axelarnet.v1beta1.ExecutePendingTransfersRequest) | [ExecutePendingTransfersResponse](#axelarnetwork.axelarnet.v1beta1.ExecutePendingTransfersResponse) |  | POST|/axelar/axelarnet/execute-pending-transfers|
| `RegisterIBCPath` | [RegisterIBCPathRequest](#axelarnetwork.axelarnet.v1beta1.RegisterIBCPathRequest) | [RegisterIBCPathResponse](#axelarnetwork.axelarnet.v1beta1.RegisterIBCPathResponse) |  | POST|/axelar/axelarnet/register-ibc-path|
| `AddCosmosBasedChain` | [AddCosmosBasedChainRequest](#axelarnetwork.axelarnet.v1beta1.AddCosmosBasedChainRequest) | [AddCosmosBasedChainResponse](#axelarnetwork.axelarnet.v1beta1.AddCosmosBasedChainResponse) |  | POST|/axelar/axelarnet/add-cosmos-based-chain|
| `RegisterAsset` | [RegisterAssetRequest](#axelarnetwork.axelarnet.v1beta1.RegisterAssetRequest) | [RegisterAssetResponse](#axelarnetwork.axelarnet.v1beta1.RegisterAssetResponse) |  | POST|/axelar/axelarnet/register-asset|
| `RouteIBCTransfers` | [RouteIBCTransfersRequest](#axelarnetwork.axelarnet.v1beta1.RouteIBCTransfersRequest) | [RouteIBCTransfersResponse](#axelarnetwork.axelarnet.v1beta1.RouteIBCTransfersResponse) |  | POST|/axelar/axelarnet/route-ibc-transfers|
| `RegisterFeeCollector` | [RegisterFeeCollectorRequest](#axelarnetwork.axelarnet.v1beta1.RegisterFeeCollectorRequest) | [RegisterFeeCollectorResponse](#axelarnetwork.axelarnet.v1beta1.RegisterFeeCollectorResponse) |  | POST|/axelar/axelarnet/register-fee-collector|


<a name="axelarnetwork.axelarnet.v1beta1.QueryService"></a>

### QueryService
QueryService defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `PendingIBCTransferCount` | [PendingIBCTransferCountRequest](#axelarnetwork.axelarnet.v1beta1.PendingIBCTransferCountRequest) | [PendingIBCTransferCountResponse](#axelarnetwork.axelarnet.v1beta1.PendingIBCTransferCountResponse) |  | GET|/axelar/axelarnet/v1beta1/ibc_transfer_count|

 <!-- end services -->



<a name="axelarnetwork/bitcoin/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/bitcoin/v1beta1/types.proto



<a name="axelarnetwork.bitcoin.v1beta1.AddressInfo"></a>

### AddressInfo
AddressInfo is a wrapper containing the Bitcoin P2WSH address, it's
corresponding script and the underlying key


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `role` | [AddressRole](#axelarnetwork.bitcoin.v1beta1.AddressRole) |  |  |
| `redeem_script` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `max_sig_count` | [uint32](#uint32) |  |  |
| `spending_condition` | [AddressInfo.SpendingCondition](#axelarnetwork.bitcoin.v1beta1.AddressInfo.SpendingCondition) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.AddressInfo.SpendingCondition"></a>

### AddressInfo.SpendingCondition



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `internal_key_ids` | [string](#string) | repeated | internal_key_ids lists the internal key IDs that one of which has to sign regardless of locktime |
| `external_key_ids` | [string](#string) | repeated | external_key_ids lists the external key IDs that external_multisig_threshold of which have to sign to spend before locktime if set |
| `external_multisig_threshold` | [int64](#int64) |  |  |
| `lock_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.Network"></a>

### Network



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.OutPointInfo"></a>

### OutPointInfo
OutPointInfo describes all the necessary information to confirm the outPoint
of a transaction


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `out_point` | [string](#string) |  |  |
| `amount` | [int64](#int64) |  |  |
| `address` | [string](#string) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.SignedTx"></a>

### SignedTx



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `type` | [TxType](#axelarnetwork.bitcoin.v1beta1.TxType) |  |  |
| `tx` | [bytes](#bytes) |  |  |
| `prev_signed_tx_hash` | [bytes](#bytes) |  |  |
| `confirmation_required` | [bool](#bool) |  |  |
| `anyone_can_spend_vout` | [uint32](#uint32) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.UnsignedTx"></a>

### UnsignedTx



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `type` | [TxType](#axelarnetwork.bitcoin.v1beta1.TxType) |  |  |
| `tx` | [bytes](#bytes) |  |  |
| `info` | [UnsignedTx.Info](#axelarnetwork.bitcoin.v1beta1.UnsignedTx.Info) |  |  |
| `status` | [TxStatus](#axelarnetwork.bitcoin.v1beta1.TxStatus) |  |  |
| `confirmation_required` | [bool](#bool) |  |  |
| `anyone_can_spend_vout` | [uint32](#uint32) |  |  |
| `prev_aborted_key_id` | [string](#string) |  |  |
| `internal_transfer_amount` | [int64](#int64) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.UnsignedTx.Info"></a>

### UnsignedTx.Info



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `rotate_key` | [bool](#bool) |  |  |
| `input_infos` | [UnsignedTx.Info.InputInfo](#axelarnetwork.bitcoin.v1beta1.UnsignedTx.Info.InputInfo) | repeated |  |






<a name="axelarnetwork.bitcoin.v1beta1.UnsignedTx.Info.InputInfo"></a>

### UnsignedTx.Info.InputInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sig_requirements` | [UnsignedTx.Info.InputInfo.SigRequirement](#axelarnetwork.bitcoin.v1beta1.UnsignedTx.Info.InputInfo.SigRequirement) | repeated |  |






<a name="axelarnetwork.bitcoin.v1beta1.UnsignedTx.Info.InputInfo.SigRequirement"></a>

### UnsignedTx.Info.InputInfo.SigRequirement



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `sig_hash` | [bytes](#bytes) |  |  |





 <!-- end messages -->


<a name="axelarnetwork.bitcoin.v1beta1.AddressRole"></a>

### AddressRole


| Name | Number | Description |
| ---- | ------ | ----------- |
| ADDRESS_ROLE_UNSPECIFIED | 0 |  |
| ADDRESS_ROLE_DEPOSIT | 1 |  |
| ADDRESS_ROLE_CONSOLIDATION | 2 |  |



<a name="axelarnetwork.bitcoin.v1beta1.OutPointState"></a>

### OutPointState


| Name | Number | Description |
| ---- | ------ | ----------- |
| OUT_POINT_STATE_UNSPECIFIED | 0 |  |
| OUT_POINT_STATE_PENDING | 1 |  |
| OUT_POINT_STATE_CONFIRMED | 2 |  |
| OUT_POINT_STATE_SPENT | 3 |  |



<a name="axelarnetwork.bitcoin.v1beta1.TxStatus"></a>

### TxStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| TX_STATUS_UNSPECIFIED | 0 |  |
| TX_STATUS_CREATED | 1 |  |
| TX_STATUS_SIGNING | 2 |  |
| TX_STATUS_ABORTED | 3 |  |
| TX_STATUS_SIGNED | 4 |  |



<a name="axelarnetwork.bitcoin.v1beta1.TxType"></a>

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



<a name="axelarnetwork/bitcoin/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/bitcoin/v1beta1/params.proto



<a name="axelarnetwork.bitcoin.v1beta1.Params"></a>

### Params



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `network` | [Network](#axelarnetwork.bitcoin.v1beta1.Network) |  |  |
| `confirmation_height` | [uint64](#uint64) |  |  |
| `revote_locking_period` | [int64](#int64) |  |  |
| `sig_check_interval` | [int64](#int64) |  |  |
| `min_output_amount` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  |  |
| `max_input_count` | [int64](#int64) |  |  |
| `max_secondary_output_amount` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  |  |
| `master_key_retention_period` | [int64](#int64) |  |  |
| `master_address_internal_key_lock_duration` | [int64](#int64) |  |  |
| `master_address_external_key_lock_duration` | [int64](#int64) |  |  |
| `voting_threshold` | [axelarnetwork.utils.v1beta1.Threshold](#axelarnetwork.utils.v1beta1.Threshold) |  |  |
| `min_voter_count` | [int64](#int64) |  |  |
| `max_tx_size` | [int64](#int64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/bitcoin/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/bitcoin/v1beta1/genesis.proto



<a name="axelarnetwork.bitcoin.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelarnetwork.bitcoin.v1beta1.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/bitcoin/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/bitcoin/v1beta1/query.proto



<a name="axelarnetwork.bitcoin.v1beta1.DepositQueryParams"></a>

### DepositQueryParams
DepositQueryParams describe the parameters used to query for a Bitcoin
deposit address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `chain` | [string](#string) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.QueryAddressResponse"></a>

### QueryAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.QueryDepositStatusResponse"></a>

### QueryDepositStatusResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |
| `status` | [OutPointState](#axelarnetwork.bitcoin.v1beta1.OutPointState) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.QueryTxResponse"></a>

### QueryTxResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx` | [string](#string) |  |  |
| `status` | [TxStatus](#axelarnetwork.bitcoin.v1beta1.TxStatus) |  |  |
| `confirmation_required` | [bool](#bool) |  |  |
| `prev_signed_tx_hash` | [string](#string) |  |  |
| `anyone_can_spend_vout` | [uint32](#uint32) |  |  |
| `signing_infos` | [QueryTxResponse.SigningInfo](#axelarnetwork.bitcoin.v1beta1.QueryTxResponse.SigningInfo) | repeated |  |






<a name="axelarnetwork.bitcoin.v1beta1.QueryTxResponse.SigningInfo"></a>

### QueryTxResponse.SigningInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `redeem_script` | [string](#string) |  |  |
| `amount` | [int64](#int64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/snapshot/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/snapshot/exported/v1beta1/types.proto



<a name="axelarnetwork.snapshot.exported.v1beta1.Snapshot"></a>

### Snapshot



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validators` | [Validator](#axelarnetwork.snapshot.exported.v1beta1.Validator) | repeated |  |
| `timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `height` | [int64](#int64) |  |  |
| `total_share_count` | [bytes](#bytes) |  |  |
| `counter` | [int64](#int64) |  |  |
| `key_share_distribution_policy` | [axelarnetwork.tss.exported.v1beta1.KeyShareDistributionPolicy](#axelarnetwork.tss.exported.v1beta1.KeyShareDistributionPolicy) |  |  |
| `corruption_threshold` | [int64](#int64) |  |  |






<a name="axelarnetwork.snapshot.exported.v1beta1.Validator"></a>

### Validator



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sdk_validator` | [google.protobuf.Any](#google.protobuf.Any) |  |  |
| `share_count` | [int64](#int64) |  |  |





 <!-- end messages -->


<a name="axelarnetwork.snapshot.exported.v1beta1.ValidatorIllegibility"></a>

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



<a name="axelarnetwork/vote/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/vote/exported/v1beta1/types.proto



<a name="axelarnetwork.vote.exported.v1beta1.PollKey"></a>

### PollKey
PollKey represents the key data for a poll


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |






<a name="axelarnetwork.vote.exported.v1beta1.PollMetadata"></a>

### PollMetadata
PollMetadata represents a poll with write-in voting, i.e. the result of the
vote can have any data type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [PollKey](#axelarnetwork.vote.exported.v1beta1.PollKey) |  |  |
| `expires_at` | [int64](#int64) |  |  |
| `result` | [google.protobuf.Any](#google.protobuf.Any) |  |  |
| `voting_threshold` | [axelarnetwork.utils.v1beta1.Threshold](#axelarnetwork.utils.v1beta1.Threshold) |  |  |
| `state` | [PollState](#axelarnetwork.vote.exported.v1beta1.PollState) |  |  |
| `min_voter_count` | [int64](#int64) |  |  |
| `voters` | [Voter](#axelarnetwork.vote.exported.v1beta1.Voter) | repeated |  |
| `total_voting_power` | [bytes](#bytes) |  |  |
| `reward_pool_name` | [string](#string) |  |  |






<a name="axelarnetwork.vote.exported.v1beta1.Vote"></a>

### Vote



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `results` | [google.protobuf.Any](#google.protobuf.Any) | repeated |  |






<a name="axelarnetwork.vote.exported.v1beta1.Voter"></a>

### Voter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validator` | [bytes](#bytes) |  |  |
| `voting_power` | [int64](#int64) |  |  |





 <!-- end messages -->


<a name="axelarnetwork.vote.exported.v1beta1.PollState"></a>

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



<a name="axelarnetwork/bitcoin/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/bitcoin/v1beta1/tx.proto



<a name="axelarnetwork.bitcoin.v1beta1.ConfirmOutpointRequest"></a>

### ConfirmOutpointRequest
MsgConfirmOutpoint represents a message to trigger the confirmation of a
Bitcoin outpoint


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `out_point_info` | [OutPointInfo](#axelarnetwork.bitcoin.v1beta1.OutPointInfo) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.ConfirmOutpointResponse"></a>

### ConfirmOutpointResponse







<a name="axelarnetwork.bitcoin.v1beta1.CreateMasterTxRequest"></a>

### CreateMasterTxRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `secondary_key_amount` | [int64](#int64) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.CreateMasterTxResponse"></a>

### CreateMasterTxResponse







<a name="axelarnetwork.bitcoin.v1beta1.CreatePendingTransfersTxRequest"></a>

### CreatePendingTransfersTxRequest
CreatePendingTransfersTxRequest represents a message to trigger the creation
of a secondary key consolidation transaction


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `master_key_amount` | [int64](#int64) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.CreatePendingTransfersTxResponse"></a>

### CreatePendingTransfersTxResponse







<a name="axelarnetwork.bitcoin.v1beta1.CreateRescueTxRequest"></a>

### CreateRescueTxRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.CreateRescueTxResponse"></a>

### CreateRescueTxResponse







<a name="axelarnetwork.bitcoin.v1beta1.LinkRequest"></a>

### LinkRequest
MsgLink represents a message to link a cross-chain address to a Bitcoin
address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `recipient_addr` | [string](#string) |  |  |
| `recipient_chain` | [string](#string) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.LinkResponse"></a>

### LinkResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposit_addr` | [string](#string) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.SignTxRequest"></a>

### SignTxRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `tx_type` | [TxType](#axelarnetwork.bitcoin.v1beta1.TxType) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.SignTxResponse"></a>

### SignTxResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `position` | [int64](#int64) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.SubmitExternalSignatureRequest"></a>

### SubmitExternalSignatureRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `signature` | [bytes](#bytes) |  |  |
| `sig_hash` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.SubmitExternalSignatureResponse"></a>

### SubmitExternalSignatureResponse







<a name="axelarnetwork.bitcoin.v1beta1.VoteConfirmOutpointRequest"></a>

### VoteConfirmOutpointRequest
MsgVoteConfirmOutpoint represents a message to that votes on an outpoint


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll_key` | [axelarnetwork.vote.exported.v1beta1.PollKey](#axelarnetwork.vote.exported.v1beta1.PollKey) |  |  |
| `out_point` | [string](#string) |  |  |
| `confirmed` | [bool](#bool) |  |  |






<a name="axelarnetwork.bitcoin.v1beta1.VoteConfirmOutpointResponse"></a>

### VoteConfirmOutpointResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `status` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/bitcoin/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/bitcoin/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelarnetwork.bitcoin.v1beta1.MsgService"></a>

### MsgService
Msg defines the bitcoin Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Link` | [LinkRequest](#axelarnetwork.bitcoin.v1beta1.LinkRequest) | [LinkResponse](#axelarnetwork.bitcoin.v1beta1.LinkResponse) |  | POST|/axelar/bitcoin/link/{recipient_chain}|
| `ConfirmOutpoint` | [ConfirmOutpointRequest](#axelarnetwork.bitcoin.v1beta1.ConfirmOutpointRequest) | [ConfirmOutpointResponse](#axelarnetwork.bitcoin.v1beta1.ConfirmOutpointResponse) |  | POST|/axelar/bitcoin/confirm|
| `VoteConfirmOutpoint` | [VoteConfirmOutpointRequest](#axelarnetwork.bitcoin.v1beta1.VoteConfirmOutpointRequest) | [VoteConfirmOutpointResponse](#axelarnetwork.bitcoin.v1beta1.VoteConfirmOutpointResponse) |  | ||
| `CreatePendingTransfersTx` | [CreatePendingTransfersTxRequest](#axelarnetwork.bitcoin.v1beta1.CreatePendingTransfersTxRequest) | [CreatePendingTransfersTxResponse](#axelarnetwork.bitcoin.v1beta1.CreatePendingTransfersTxResponse) |  | POST|/axelar/bitcoin/create-pending-transfers-tx|
| `CreateMasterTx` | [CreateMasterTxRequest](#axelarnetwork.bitcoin.v1beta1.CreateMasterTxRequest) | [CreateMasterTxResponse](#axelarnetwork.bitcoin.v1beta1.CreateMasterTxResponse) |  | POST|/axelar/bitcoin/create-master-tx|
| `CreateRescueTx` | [CreateRescueTxRequest](#axelarnetwork.bitcoin.v1beta1.CreateRescueTxRequest) | [CreateRescueTxResponse](#axelarnetwork.bitcoin.v1beta1.CreateRescueTxResponse) |  | POST|/axelar/bitcoin/create-rescue-tx|
| `SignTx` | [SignTxRequest](#axelarnetwork.bitcoin.v1beta1.SignTxRequest) | [SignTxResponse](#axelarnetwork.bitcoin.v1beta1.SignTxResponse) |  | POST|/axelar/bitcoin/sign-tx|
| `SubmitExternalSignature` | [SubmitExternalSignatureRequest](#axelarnetwork.bitcoin.v1beta1.SubmitExternalSignatureRequest) | [SubmitExternalSignatureResponse](#axelarnetwork.bitcoin.v1beta1.SubmitExternalSignatureResponse) |  | POST|/axelar/bitcoin/submit-external-signature|

 <!-- end services -->



<a name="axelarnetwork/utils/v1beta1/queuer.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/utils/v1beta1/queuer.proto



<a name="axelarnetwork.utils.v1beta1.QueueState"></a>

### QueueState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `items` | [QueueState.ItemsEntry](#axelarnetwork.utils.v1beta1.QueueState.ItemsEntry) | repeated |  |






<a name="axelarnetwork.utils.v1beta1.QueueState.Item"></a>

### QueueState.Item



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [bytes](#bytes) |  |  |
| `value` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.utils.v1beta1.QueueState.ItemsEntry"></a>

### QueueState.ItemsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [QueueState.Item](#axelarnetwork.utils.v1beta1.QueueState.Item) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/evm/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/evm/v1beta1/types.proto



<a name="axelarnetwork.evm.v1beta1.Asset"></a>

### Asset



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `name` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.BurnerInfo"></a>

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






<a name="axelarnetwork.evm.v1beta1.Command"></a>

### Command



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [bytes](#bytes) |  |  |
| `command` | [string](#string) |  |  |
| `params` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `max_gas_cost` | [uint32](#uint32) |  |  |






<a name="axelarnetwork.evm.v1beta1.CommandBatchMetadata"></a>

### CommandBatchMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [bytes](#bytes) |  |  |
| `command_ids` | [bytes](#bytes) | repeated |  |
| `data` | [bytes](#bytes) |  |  |
| `sig_hash` | [bytes](#bytes) |  |  |
| `status` | [BatchedCommandsStatus](#axelarnetwork.evm.v1beta1.BatchedCommandsStatus) |  |  |
| `key_id` | [string](#string) |  |  |
| `prev_batched_commands_id` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.ERC20Deposit"></a>

### ERC20Deposit
ERC20Deposit contains information for an ERC20 deposit


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `amount` | [bytes](#bytes) |  |  |
| `asset` | [string](#string) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `burner_address` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.ERC20TokenMetadata"></a>

### ERC20TokenMetadata
ERC20TokenMetadata describes information about an ERC20 token


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset` | [string](#string) |  |  |
| `chain_id` | [bytes](#bytes) |  |  |
| `details` | [TokenDetails](#axelarnetwork.evm.v1beta1.TokenDetails) |  |  |
| `token_address` | [string](#string) |  |  |
| `tx_hash` | [string](#string) |  |  |
| `status` | [Status](#axelarnetwork.evm.v1beta1.Status) |  |  |
| `is_external` | [bool](#bool) |  |  |
| `burner_code` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.Event"></a>

### Event



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `index` | [uint64](#uint64) |  |  |
| `status` | [Event.Status](#axelarnetwork.evm.v1beta1.Event.Status) |  |  |
| `token_sent` | [EventTokenSent](#axelarnetwork.evm.v1beta1.EventTokenSent) |  |  |
| `contract_call` | [EventContractCall](#axelarnetwork.evm.v1beta1.EventContractCall) |  |  |
| `contract_call_with_token` | [EventContractCallWithToken](#axelarnetwork.evm.v1beta1.EventContractCallWithToken) |  |  |
| `transfer` | [EventTransfer](#axelarnetwork.evm.v1beta1.EventTransfer) |  |  |
| `token_deployed` | [EventTokenDeployed](#axelarnetwork.evm.v1beta1.EventTokenDeployed) |  |  |
| `multisig_ownership_transferred` | [EventMultisigOwnershipTransferred](#axelarnetwork.evm.v1beta1.EventMultisigOwnershipTransferred) |  |  |
| `multisig_operatorship_transferred` | [EventMultisigOperatorshipTransferred](#axelarnetwork.evm.v1beta1.EventMultisigOperatorshipTransferred) |  |  |
| `singlesig_ownership_transferred` | [EventSinglesigOwnershipTransferred](#axelarnetwork.evm.v1beta1.EventSinglesigOwnershipTransferred) |  |  |
| `singlesig_operatorship_transferred` | [EventSinglesigOperatorshipTransferred](#axelarnetwork.evm.v1beta1.EventSinglesigOperatorshipTransferred) |  |  |






<a name="axelarnetwork.evm.v1beta1.EventContractCall"></a>

### EventContractCall



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `contract_address` | [string](#string) |  |  |
| `payload_hash` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.EventContractCallWithToken"></a>

### EventContractCallWithToken



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `contract_address` | [string](#string) |  |  |
| `payload_hash` | [bytes](#bytes) |  |  |
| `symbol` | [string](#string) |  |  |
| `amount` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.EventMultisigOperatorshipTransferred"></a>

### EventMultisigOperatorshipTransferred



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pre_operators` | [bytes](#bytes) | repeated |  |
| `prev_threshold` | [bytes](#bytes) |  |  |
| `new_operators` | [bytes](#bytes) | repeated |  |
| `new_threshold` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.EventMultisigOwnershipTransferred"></a>

### EventMultisigOwnershipTransferred



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pre_owners` | [bytes](#bytes) | repeated |  |
| `prev_threshold` | [bytes](#bytes) |  |  |
| `new_owners` | [bytes](#bytes) | repeated |  |
| `new_threshold` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.EventSinglesigOperatorshipTransferred"></a>

### EventSinglesigOperatorshipTransferred



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pre_operator` | [bytes](#bytes) |  |  |
| `new_operator` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.EventSinglesigOwnershipTransferred"></a>

### EventSinglesigOwnershipTransferred



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pre_owner` | [bytes](#bytes) |  |  |
| `new_owner` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.EventTokenDeployed"></a>

### EventTokenDeployed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `symbol` | [string](#string) |  |  |
| `token_address` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.EventTokenSent"></a>

### EventTokenSent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `destination_address` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `amount` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.EventTransfer"></a>

### EventTransfer



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to` | [bytes](#bytes) |  |  |
| `amount` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.Gateway"></a>

### Gateway



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [bytes](#bytes) |  |  |
| `status` | [Gateway.Status](#axelarnetwork.evm.v1beta1.Gateway.Status) |  | **Deprecated.**  |






<a name="axelarnetwork.evm.v1beta1.NetworkInfo"></a>

### NetworkInfo
NetworkInfo describes information about a network


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `id` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.SigMetadata"></a>

### SigMetadata
SigMetadata stores necessary information for external apps to map signature
results to evm relay transaction types


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `type` | [SigType](#axelarnetwork.evm.v1beta1.SigType) |  |  |
| `chain` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.TokenDetails"></a>

### TokenDetails



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `token_name` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `decimals` | [uint32](#uint32) |  |  |
| `capacity` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.TransactionMetadata"></a>

### TransactionMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `raw_tx` | [bytes](#bytes) |  |  |
| `pub_key` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.TransferKey"></a>

### TransferKey
TransferKey contains information for a transfer ownership or operatorship


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `type` | [TransferKeyType](#axelarnetwork.evm.v1beta1.TransferKeyType) |  |  |
| `next_key_id` | [string](#string) |  |  |





 <!-- end messages -->


<a name="axelarnetwork.evm.v1beta1.BatchedCommandsStatus"></a>

### BatchedCommandsStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| BATCHED_COMMANDS_STATUS_UNSPECIFIED | 0 |  |
| BATCHED_COMMANDS_STATUS_SIGNING | 1 |  |
| BATCHED_COMMANDS_STATUS_ABORTED | 2 |  |
| BATCHED_COMMANDS_STATUS_SIGNED | 3 |  |



<a name="axelarnetwork.evm.v1beta1.DepositStatus"></a>

### DepositStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| DEPOSIT_STATUS_UNSPECIFIED | 0 |  |
| DEPOSIT_STATUS_PENDING | 1 |  |
| DEPOSIT_STATUS_CONFIRMED | 2 |  |
| DEPOSIT_STATUS_BURNED | 3 |  |



<a name="axelarnetwork.evm.v1beta1.Event.Status"></a>

### Event.Status


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATUS_UNSPECIFIED | 0 |  |
| STATUS_CONFIRMED | 1 |  |
| STATUS_COMPLETED | 2 |  |
| STATUS_FAILED | 3 |  |



<a name="axelarnetwork.evm.v1beta1.Gateway.Status"></a>

### Gateway.Status


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATUS_UNSPECIFIED | 0 |  |
| STATUS_PENDING | 1 |  |
| STATUS_CONFIRMED | 2 |  |



<a name="axelarnetwork.evm.v1beta1.SigType"></a>

### SigType


| Name | Number | Description |
| ---- | ------ | ----------- |
| SIG_TYPE_UNSPECIFIED | 0 |  |
| SIG_TYPE_TX | 1 |  |
| SIG_TYPE_COMMAND | 2 |  |



<a name="axelarnetwork.evm.v1beta1.Status"></a>

### Status


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATUS_UNSPECIFIED | 0 | these enum values are used for bitwise operations, therefore they need to be powers of 2 |
| STATUS_INITIALIZED | 1 |  |
| STATUS_PENDING | 2 |  |
| STATUS_CONFIRMED | 4 |  |



<a name="axelarnetwork.evm.v1beta1.TransferKeyType"></a>

### TransferKeyType


| Name | Number | Description |
| ---- | ------ | ----------- |
| TRANSFER_KEY_TYPE_UNSPECIFIED | 0 |  |
| TRANSFER_KEY_TYPE_OWNERSHIP | 1 |  |
| TRANSFER_KEY_TYPE_OPERATORSHIP | 2 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/evm/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/evm/v1beta1/params.proto



<a name="axelarnetwork.evm.v1beta1.Params"></a>

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
| `networks` | [NetworkInfo](#axelarnetwork.evm.v1beta1.NetworkInfo) | repeated |  |
| `voting_threshold` | [axelarnetwork.utils.v1beta1.Threshold](#axelarnetwork.utils.v1beta1.Threshold) |  |  |
| `min_voter_count` | [int64](#int64) |  |  |
| `commands_gas_limit` | [uint32](#uint32) |  |  |






<a name="axelarnetwork.evm.v1beta1.PendingChain"></a>

### PendingChain



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelarnetwork.evm.v1beta1.Params) |  |  |
| `chain` | [axelarnetwork.nexus.exported.v1beta1.Chain](#axelarnetwork.nexus.exported.v1beta1.Chain) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/evm/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/evm/v1beta1/genesis.proto



<a name="axelarnetwork.evm.v1beta1.GenesisState"></a>

### GenesisState
GenesisState represents the genesis state


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chains` | [GenesisState.Chain](#axelarnetwork.evm.v1beta1.GenesisState.Chain) | repeated |  |






<a name="axelarnetwork.evm.v1beta1.GenesisState.Chain"></a>

### GenesisState.Chain



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelarnetwork.evm.v1beta1.Params) |  |  |
| `burner_infos` | [BurnerInfo](#axelarnetwork.evm.v1beta1.BurnerInfo) | repeated |  |
| `command_queue` | [axelarnetwork.utils.v1beta1.QueueState](#axelarnetwork.utils.v1beta1.QueueState) |  |  |
| `confirmed_deposits` | [ERC20Deposit](#axelarnetwork.evm.v1beta1.ERC20Deposit) | repeated |  |
| `burned_deposits` | [ERC20Deposit](#axelarnetwork.evm.v1beta1.ERC20Deposit) | repeated |  |
| `command_batches` | [CommandBatchMetadata](#axelarnetwork.evm.v1beta1.CommandBatchMetadata) | repeated |  |
| `gateway` | [Gateway](#axelarnetwork.evm.v1beta1.Gateway) |  |  |
| `tokens` | [ERC20TokenMetadata](#axelarnetwork.evm.v1beta1.ERC20TokenMetadata) | repeated |  |
| `events` | [Event](#axelarnetwork.evm.v1beta1.Event) | repeated |  |
| `confirmed_event_queue` | [axelarnetwork.utils.v1beta1.QueueState](#axelarnetwork.utils.v1beta1.QueueState) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/evm/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/evm/v1beta1/query.proto



<a name="axelarnetwork.evm.v1beta1.BatchedCommandsRequest"></a>

### BatchedCommandsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `id` | [string](#string) |  | id defines an optional id for the commandsbatch. If not specified the latest will be returned |






<a name="axelarnetwork.evm.v1beta1.BatchedCommandsResponse"></a>

### BatchedCommandsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `data` | [string](#string) |  |  |
| `status` | [BatchedCommandsStatus](#axelarnetwork.evm.v1beta1.BatchedCommandsStatus) |  |  |
| `key_id` | [string](#string) |  |  |
| `signature` | [string](#string) | repeated |  |
| `execute_data` | [string](#string) |  |  |
| `prev_batched_commands_id` | [string](#string) |  |  |
| `command_ids` | [string](#string) | repeated |  |






<a name="axelarnetwork.evm.v1beta1.BurnerInfoRequest"></a>

### BurnerInfoRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.BurnerInfoResponse"></a>

### BurnerInfoResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `burner_info` | [BurnerInfo](#axelarnetwork.evm.v1beta1.BurnerInfo) |  |  |






<a name="axelarnetwork.evm.v1beta1.BytecodeRequest"></a>

### BytecodeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `contract` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.BytecodeResponse"></a>

### BytecodeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `bytecode` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.ChainsRequest"></a>

### ChainsRequest







<a name="axelarnetwork.evm.v1beta1.ChainsResponse"></a>

### ChainsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chains` | [string](#string) | repeated |  |






<a name="axelarnetwork.evm.v1beta1.ConfirmationHeightRequest"></a>

### ConfirmationHeightRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.ConfirmationHeightResponse"></a>

### ConfirmationHeightResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `height` | [uint64](#uint64) |  |  |






<a name="axelarnetwork.evm.v1beta1.DepositQueryParams"></a>

### DepositQueryParams
DepositQueryParams describe the parameters used to query for an EVM
deposit address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `asset` | [string](#string) |  |  |
| `chain` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.DepositStateRequest"></a>

### DepositStateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `params` | [QueryDepositStateParams](#axelarnetwork.evm.v1beta1.QueryDepositStateParams) |  |  |






<a name="axelarnetwork.evm.v1beta1.DepositStateResponse"></a>

### DepositStateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `status` | [DepositStatus](#axelarnetwork.evm.v1beta1.DepositStatus) |  |  |






<a name="axelarnetwork.evm.v1beta1.EventRequest"></a>

### EventRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `event_id` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.EventResponse"></a>

### EventResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `event` | [Event](#axelarnetwork.evm.v1beta1.Event) |  |  |






<a name="axelarnetwork.evm.v1beta1.GatewayAddressRequest"></a>

### GatewayAddressRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.GatewayAddressResponse"></a>

### GatewayAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.KeyAddressRequest"></a>

### KeyAddressRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `role` | [int32](#int32) |  |  |
| `id` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.KeyAddressResponse"></a>

### KeyAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `multisig_addresses` | [KeyAddressResponse.MultisigAddresses](#axelarnetwork.evm.v1beta1.KeyAddressResponse.MultisigAddresses) |  |  |
| `threshold_address` | [KeyAddressResponse.ThresholdAddress](#axelarnetwork.evm.v1beta1.KeyAddressResponse.ThresholdAddress) |  |  |






<a name="axelarnetwork.evm.v1beta1.KeyAddressResponse.MultisigAddresses"></a>

### KeyAddressResponse.MultisigAddresses



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `addresses` | [string](#string) | repeated |  |
| `threshold` | [uint32](#uint32) |  |  |






<a name="axelarnetwork.evm.v1beta1.KeyAddressResponse.ThresholdAddress"></a>

### KeyAddressResponse.ThresholdAddress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.PendingCommandsRequest"></a>

### PendingCommandsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.PendingCommandsResponse"></a>

### PendingCommandsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `commands` | [QueryCommandResponse](#axelarnetwork.evm.v1beta1.QueryCommandResponse) | repeated |  |






<a name="axelarnetwork.evm.v1beta1.QueryBurnerAddressResponse"></a>

### QueryBurnerAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.QueryCommandResponse"></a>

### QueryCommandResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `type` | [string](#string) |  |  |
| `params` | [QueryCommandResponse.ParamsEntry](#axelarnetwork.evm.v1beta1.QueryCommandResponse.ParamsEntry) | repeated |  |
| `key_id` | [string](#string) |  |  |
| `max_gas_cost` | [uint32](#uint32) |  |  |






<a name="axelarnetwork.evm.v1beta1.QueryCommandResponse.ParamsEntry"></a>

### QueryCommandResponse.ParamsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.QueryDepositStateParams"></a>

### QueryDepositStateParams



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `burner_address` | [bytes](#bytes) |  |  |
| `amount` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.QueryTokenAddressResponse"></a>

### QueryTokenAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `confirmed` | [bool](#bool) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/evm/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/evm/v1beta1/tx.proto



<a name="axelarnetwork.evm.v1beta1.AddChainRequest"></a>

### AddChainRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `name` | [string](#string) |  |  |
| `key_type` | [axelarnetwork.tss.exported.v1beta1.KeyType](#axelarnetwork.tss.exported.v1beta1.KeyType) |  |  |
| `params` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.AddChainResponse"></a>

### AddChainResponse







<a name="axelarnetwork.evm.v1beta1.ConfirmChainRequest"></a>

### ConfirmChainRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `name` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.ConfirmChainResponse"></a>

### ConfirmChainResponse







<a name="axelarnetwork.evm.v1beta1.ConfirmDepositRequest"></a>

### ConfirmDepositRequest
MsgConfirmDeposit represents an erc20 deposit confirmation message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `amount` | [bytes](#bytes) |  | **Deprecated.**  |
| `burner_address` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.ConfirmDepositResponse"></a>

### ConfirmDepositResponse







<a name="axelarnetwork.evm.v1beta1.ConfirmGatewayTxRequest"></a>

### ConfirmGatewayTxRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.ConfirmGatewayTxResponse"></a>

### ConfirmGatewayTxResponse







<a name="axelarnetwork.evm.v1beta1.ConfirmTokenRequest"></a>

### ConfirmTokenRequest
MsgConfirmToken represents a token deploy confirmation message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `asset` | [Asset](#axelarnetwork.evm.v1beta1.Asset) |  |  |






<a name="axelarnetwork.evm.v1beta1.ConfirmTokenResponse"></a>

### ConfirmTokenResponse







<a name="axelarnetwork.evm.v1beta1.ConfirmTransferKeyRequest"></a>

### ConfirmTransferKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `transfer_type` | [TransferKeyType](#axelarnetwork.evm.v1beta1.TransferKeyType) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.ConfirmTransferKeyResponse"></a>

### ConfirmTransferKeyResponse







<a name="axelarnetwork.evm.v1beta1.CreateBurnTokensRequest"></a>

### CreateBurnTokensRequest
CreateBurnTokensRequest represents the message to create commands to burn
tokens with AxelarGateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.CreateBurnTokensResponse"></a>

### CreateBurnTokensResponse







<a name="axelarnetwork.evm.v1beta1.CreateDeployTokenRequest"></a>

### CreateDeployTokenRequest
CreateDeployTokenRequest represents the message to create a deploy token
command for AxelarGateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `asset` | [Asset](#axelarnetwork.evm.v1beta1.Asset) |  |  |
| `token_details` | [TokenDetails](#axelarnetwork.evm.v1beta1.TokenDetails) |  |  |
| `address` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.CreateDeployTokenResponse"></a>

### CreateDeployTokenResponse







<a name="axelarnetwork.evm.v1beta1.CreatePendingTransfersRequest"></a>

### CreatePendingTransfersRequest
CreatePendingTransfersRequest represents a message to trigger the creation of
commands handling all pending transfers


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.CreatePendingTransfersResponse"></a>

### CreatePendingTransfersResponse







<a name="axelarnetwork.evm.v1beta1.CreateTransferOperatorshipRequest"></a>

### CreateTransferOperatorshipRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.CreateTransferOperatorshipResponse"></a>

### CreateTransferOperatorshipResponse







<a name="axelarnetwork.evm.v1beta1.CreateTransferOwnershipRequest"></a>

### CreateTransferOwnershipRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.CreateTransferOwnershipResponse"></a>

### CreateTransferOwnershipResponse







<a name="axelarnetwork.evm.v1beta1.LinkRequest"></a>

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






<a name="axelarnetwork.evm.v1beta1.LinkResponse"></a>

### LinkResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposit_addr` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.SetGatewayRequest"></a>

### SetGatewayRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `address` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.evm.v1beta1.SetGatewayResponse"></a>

### SetGatewayResponse







<a name="axelarnetwork.evm.v1beta1.SignCommandsRequest"></a>

### SignCommandsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |






<a name="axelarnetwork.evm.v1beta1.SignCommandsResponse"></a>

### SignCommandsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `batched_commands_id` | [bytes](#bytes) |  |  |
| `command_count` | [uint32](#uint32) |  |  |






<a name="axelarnetwork.evm.v1beta1.VoteConfirmChainRequest"></a>

### VoteConfirmChainRequest
MsgVoteConfirmChain represents a message that votes on a new EVM chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `name` | [string](#string) |  |  |
| `poll_key` | [axelarnetwork.vote.exported.v1beta1.PollKey](#axelarnetwork.vote.exported.v1beta1.PollKey) |  |  |
| `confirmed` | [bool](#bool) |  |  |






<a name="axelarnetwork.evm.v1beta1.VoteConfirmChainResponse"></a>

### VoteConfirmChainResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/evm/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/evm/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelarnetwork.evm.v1beta1.MsgService"></a>

### MsgService
Msg defines the evm Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `SetGateway` | [SetGatewayRequest](#axelarnetwork.evm.v1beta1.SetGatewayRequest) | [SetGatewayResponse](#axelarnetwork.evm.v1beta1.SetGatewayResponse) |  | POST|/axelar/evm/set-gateway|
| `ConfirmGatewayTx` | [ConfirmGatewayTxRequest](#axelarnetwork.evm.v1beta1.ConfirmGatewayTxRequest) | [ConfirmGatewayTxResponse](#axelarnetwork.evm.v1beta1.ConfirmGatewayTxResponse) |  | POST|/axelar/evm/confirm-gateway-tx|
| `Link` | [LinkRequest](#axelarnetwork.evm.v1beta1.LinkRequest) | [LinkResponse](#axelarnetwork.evm.v1beta1.LinkResponse) |  | POST|/axelar/evm/link/{recipient_chain}|
| `ConfirmChain` | [ConfirmChainRequest](#axelarnetwork.evm.v1beta1.ConfirmChainRequest) | [ConfirmChainResponse](#axelarnetwork.evm.v1beta1.ConfirmChainResponse) |  | POST|/axelar/evm/confirm-chain|
| `ConfirmToken` | [ConfirmTokenRequest](#axelarnetwork.evm.v1beta1.ConfirmTokenRequest) | [ConfirmTokenResponse](#axelarnetwork.evm.v1beta1.ConfirmTokenResponse) |  | POST|/axelar/evm/confirm-erc20-deploy|
| `ConfirmDeposit` | [ConfirmDepositRequest](#axelarnetwork.evm.v1beta1.ConfirmDepositRequest) | [ConfirmDepositResponse](#axelarnetwork.evm.v1beta1.ConfirmDepositResponse) |  | POST|/axelar/evm/confirm-erc20-deposit|
| `ConfirmTransferKey` | [ConfirmTransferKeyRequest](#axelarnetwork.evm.v1beta1.ConfirmTransferKeyRequest) | [ConfirmTransferKeyResponse](#axelarnetwork.evm.v1beta1.ConfirmTransferKeyResponse) |  | POST|/axelar/evm/confirm-transfer-ownership|
| `VoteConfirmChain` | [VoteConfirmChainRequest](#axelarnetwork.evm.v1beta1.VoteConfirmChainRequest) | [VoteConfirmChainResponse](#axelarnetwork.evm.v1beta1.VoteConfirmChainResponse) |  | POST|/axelar/evm/vote-confirm-chain|
| `CreateDeployToken` | [CreateDeployTokenRequest](#axelarnetwork.evm.v1beta1.CreateDeployTokenRequest) | [CreateDeployTokenResponse](#axelarnetwork.evm.v1beta1.CreateDeployTokenResponse) |  | POST|/axelar/evm/create-deploy-token|
| `CreateBurnTokens` | [CreateBurnTokensRequest](#axelarnetwork.evm.v1beta1.CreateBurnTokensRequest) | [CreateBurnTokensResponse](#axelarnetwork.evm.v1beta1.CreateBurnTokensResponse) |  | POST|/axelar/evm/sign-burn|
| `CreatePendingTransfers` | [CreatePendingTransfersRequest](#axelarnetwork.evm.v1beta1.CreatePendingTransfersRequest) | [CreatePendingTransfersResponse](#axelarnetwork.evm.v1beta1.CreatePendingTransfersResponse) |  | POST|/axelar/evm/create-pending-transfers|
| `CreateTransferOwnership` | [CreateTransferOwnershipRequest](#axelarnetwork.evm.v1beta1.CreateTransferOwnershipRequest) | [CreateTransferOwnershipResponse](#axelarnetwork.evm.v1beta1.CreateTransferOwnershipResponse) |  | POST|/axelar/evm/create-transfer-ownership|
| `CreateTransferOperatorship` | [CreateTransferOperatorshipRequest](#axelarnetwork.evm.v1beta1.CreateTransferOperatorshipRequest) | [CreateTransferOperatorshipResponse](#axelarnetwork.evm.v1beta1.CreateTransferOperatorshipResponse) |  | POST|/axelar/evm/create-transfer-operatorship|
| `SignCommands` | [SignCommandsRequest](#axelarnetwork.evm.v1beta1.SignCommandsRequest) | [SignCommandsResponse](#axelarnetwork.evm.v1beta1.SignCommandsResponse) |  | POST|/axelar/evm/sign-commands|
| `AddChain` | [AddChainRequest](#axelarnetwork.evm.v1beta1.AddChainRequest) | [AddChainResponse](#axelarnetwork.evm.v1beta1.AddChainResponse) |  | POST|/axelar/evm/add-chain|


<a name="axelarnetwork.evm.v1beta1.QueryService"></a>

### QueryService
QueryService defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `BatchedCommands` | [BatchedCommandsRequest](#axelarnetwork.evm.v1beta1.BatchedCommandsRequest) | [BatchedCommandsResponse](#axelarnetwork.evm.v1beta1.BatchedCommandsResponse) | BatchedCommands queries the batched commands for a specified chain and BatchedCommandsID if no BatchedCommandsID is specified, then it returns the latest batched commands | GET|/evm/v1beta1/batched_commands|
| `BurnerInfo` | [BurnerInfoRequest](#axelarnetwork.evm.v1beta1.BurnerInfoRequest) | [BurnerInfoResponse](#axelarnetwork.evm.v1beta1.BurnerInfoResponse) | BurnerInfo queries the burner info for the specified address | GET|/evm/v1beta1/burner_info|
| `ConfirmationHeight` | [ConfirmationHeightRequest](#axelarnetwork.evm.v1beta1.ConfirmationHeightRequest) | [ConfirmationHeightResponse](#axelarnetwork.evm.v1beta1.ConfirmationHeightResponse) | ConfirmationHeight queries the confirmation height for the specified chain | GET|/evm/v1beta1/confirmation_height|
| `DepositState` | [DepositStateRequest](#axelarnetwork.evm.v1beta1.DepositStateRequest) | [DepositStateResponse](#axelarnetwork.evm.v1beta1.DepositStateResponse) | DepositState queries the state of the specified deposit | GET|/evm/v1beta1/deposit_state|
| `PendingCommands` | [PendingCommandsRequest](#axelarnetwork.evm.v1beta1.PendingCommandsRequest) | [PendingCommandsResponse](#axelarnetwork.evm.v1beta1.PendingCommandsResponse) | PendingCommands queries the pending commands for the specified chain | GET|/evm/v1beta1/pending_commands|
| `Chains` | [ChainsRequest](#axelarnetwork.evm.v1beta1.ChainsRequest) | [ChainsResponse](#axelarnetwork.evm.v1beta1.ChainsResponse) | Chains queries the available evm chains | GET|/evm/v1beta1/chains|
| `KeyAddress` | [KeyAddressRequest](#axelarnetwork.evm.v1beta1.KeyAddressRequest) | [KeyAddressResponse](#axelarnetwork.evm.v1beta1.KeyAddressResponse) | KeyAddress queries the address of key of a chain | GET|/evm/v1beta1/key_address|
| `GatewayAddress` | [GatewayAddressRequest](#axelarnetwork.evm.v1beta1.GatewayAddressRequest) | [GatewayAddressResponse](#axelarnetwork.evm.v1beta1.GatewayAddressResponse) | GatewayAddress queries the address of axelar gateway at the specified chain | GET|/evm/v1beta1/gateway_address|
| `Bytecode` | [BytecodeRequest](#axelarnetwork.evm.v1beta1.BytecodeRequest) | [BytecodeResponse](#axelarnetwork.evm.v1beta1.BytecodeResponse) | Bytecode queries the bytecode of a specified gateway at the specified chain | GET|/evm/v1beta1/bytecode|
| `Event` | [EventRequest](#axelarnetwork.evm.v1beta1.EventRequest) | [EventResponse](#axelarnetwork.evm.v1beta1.EventResponse) | Event queries an event at the specified chain | GET|/evm/v1beta1/event|

 <!-- end services -->



<a name="axelarnetwork/nexus/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/nexus/v1beta1/params.proto



<a name="axelarnetwork.nexus.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain_activation_threshold` | [axelarnetwork.utils.v1beta1.Threshold](#axelarnetwork.utils.v1beta1.Threshold) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/nexus/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/nexus/v1beta1/genesis.proto



<a name="axelarnetwork.nexus.v1beta1.GenesisState"></a>

### GenesisState
GenesisState represents the genesis state


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelarnetwork.nexus.v1beta1.Params) |  |  |
| `nonce` | [uint64](#uint64) |  |  |
| `chains` | [axelarnetwork.nexus.exported.v1beta1.Chain](#axelarnetwork.nexus.exported.v1beta1.Chain) | repeated |  |
| `chain_states` | [ChainState](#axelarnetwork.nexus.v1beta1.ChainState) | repeated |  |
| `linked_addresses` | [LinkedAddresses](#axelarnetwork.nexus.v1beta1.LinkedAddresses) | repeated |  |
| `transfers` | [axelarnetwork.nexus.exported.v1beta1.CrossChainTransfer](#axelarnetwork.nexus.exported.v1beta1.CrossChainTransfer) | repeated |  |
| `fee` | [axelarnetwork.nexus.exported.v1beta1.TransferFee](#axelarnetwork.nexus.exported.v1beta1.TransferFee) |  |  |
| `fee_infos` | [axelarnetwork.nexus.exported.v1beta1.FeeInfo](#axelarnetwork.nexus.exported.v1beta1.FeeInfo) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/nexus/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/nexus/v1beta1/tx.proto



<a name="axelarnetwork.nexus.v1beta1.ActivateChainRequest"></a>

### ActivateChainRequest
ActivateChainRequest represents a message to activate chains


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chains` | [string](#string) | repeated |  |






<a name="axelarnetwork.nexus.v1beta1.ActivateChainResponse"></a>

### ActivateChainResponse







<a name="axelarnetwork.nexus.v1beta1.DeactivateChainRequest"></a>

### DeactivateChainRequest
DeactivateChainRequest represents a message to deactivate chains


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chains` | [string](#string) | repeated |  |






<a name="axelarnetwork.nexus.v1beta1.DeactivateChainResponse"></a>

### DeactivateChainResponse







<a name="axelarnetwork.nexus.v1beta1.DeregisterChainMaintainerRequest"></a>

### DeregisterChainMaintainerRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chains` | [string](#string) | repeated |  |






<a name="axelarnetwork.nexus.v1beta1.DeregisterChainMaintainerResponse"></a>

### DeregisterChainMaintainerResponse







<a name="axelarnetwork.nexus.v1beta1.RegisterAssetFeeRequest"></a>

### RegisterAssetFeeRequest
RegisterAssetFeeRequest represents a message to register the transfer fee
info associated to an asset on a chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `fee_info` | [axelarnetwork.nexus.exported.v1beta1.FeeInfo](#axelarnetwork.nexus.exported.v1beta1.FeeInfo) |  |  |






<a name="axelarnetwork.nexus.v1beta1.RegisterAssetFeeResponse"></a>

### RegisterAssetFeeResponse







<a name="axelarnetwork.nexus.v1beta1.RegisterChainMaintainerRequest"></a>

### RegisterChainMaintainerRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chains` | [string](#string) | repeated |  |






<a name="axelarnetwork.nexus.v1beta1.RegisterChainMaintainerResponse"></a>

### RegisterChainMaintainerResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/nexus/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/nexus/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelarnetwork.nexus.v1beta1.MsgService"></a>

### MsgService
Msg defines the nexus Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterChainMaintainer` | [RegisterChainMaintainerRequest](#axelarnetwork.nexus.v1beta1.RegisterChainMaintainerRequest) | [RegisterChainMaintainerResponse](#axelarnetwork.nexus.v1beta1.RegisterChainMaintainerResponse) |  | POST|/axelar/nexus/registerChainMaintainer|
| `DeregisterChainMaintainer` | [DeregisterChainMaintainerRequest](#axelarnetwork.nexus.v1beta1.DeregisterChainMaintainerRequest) | [DeregisterChainMaintainerResponse](#axelarnetwork.nexus.v1beta1.DeregisterChainMaintainerResponse) |  | POST|/axelar/nexus/deregisterChainMaintainer|
| `ActivateChain` | [ActivateChainRequest](#axelarnetwork.nexus.v1beta1.ActivateChainRequest) | [ActivateChainResponse](#axelarnetwork.nexus.v1beta1.ActivateChainResponse) |  | POST|/axelar/nexus/activate-chain|
| `DeactivateChain` | [DeactivateChainRequest](#axelarnetwork.nexus.v1beta1.DeactivateChainRequest) | [DeactivateChainResponse](#axelarnetwork.nexus.v1beta1.DeactivateChainResponse) |  | POST|/axelar/nexus/deactivate-chain|
| `RegisterAssetFee` | [RegisterAssetFeeRequest](#axelarnetwork.nexus.v1beta1.RegisterAssetFeeRequest) | [RegisterAssetFeeResponse](#axelarnetwork.nexus.v1beta1.RegisterAssetFeeResponse) |  | POST|/axelar/axelarnet/register-asset-fee|


<a name="axelarnetwork.nexus.v1beta1.QueryService"></a>

### QueryService
QueryService defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `LatestDepositAddress` | [LatestDepositAddressRequest](#axelarnetwork.nexus.v1beta1.LatestDepositAddressRequest) | [LatestDepositAddressResponse](#axelarnetwork.nexus.v1beta1.LatestDepositAddressResponse) | LatestDepositAddress queries the a deposit address by recipient | GET|/nexus/v1beta1/latest_deposit_address/{recipient_chain}/{recipient_addr}|
| `TransfersForChain` | [TransfersForChainRequest](#axelarnetwork.nexus.v1beta1.TransfersForChainRequest) | [TransfersForChainResponse](#axelarnetwork.nexus.v1beta1.TransfersForChainResponse) | TransfersForChain queries transfers by chain | GET|/nexus/v1beta1/transfers_for_chain|
| `Fee` | [FeeRequest](#axelarnetwork.nexus.v1beta1.FeeRequest) | [FeeResponse](#axelarnetwork.nexus.v1beta1.FeeResponse) | Fee queries the fee info by chain and asset | GET|/axelar/nexus/v1beta1/fee|
| `TransferFee` | [TransferFeeRequest](#axelarnetwork.nexus.v1beta1.TransferFeeRequest) | [TransferFeeResponse](#axelarnetwork.nexus.v1beta1.TransferFeeResponse) | TransferFee queries the transfer fee by the source, destination chain, asset and amount | GET|/axelar/nexus/v1beta1/transfer_fee|
| `Chains` | [ChainsRequest](#axelarnetwork.nexus.v1beta1.ChainsRequest) | [ChainsResponse](#axelarnetwork.nexus.v1beta1.ChainsResponse) | Chains queries the chains registered on the network | GET|/axelar/nexus/v1beta1/chains|
| `Assets` | [AssetsRequest](#axelarnetwork.nexus.v1beta1.AssetsRequest) | [AssetsResponse](#axelarnetwork.nexus.v1beta1.AssetsResponse) | Assets queries the assets registered for a chain | GET|/axelar/nexus/v1beta1/assets/{chain}|
| `ChainState` | [ChainStateRequest](#axelarnetwork.nexus.v1beta1.ChainStateRequest) | [ChainStateResponse](#axelarnetwork.nexus.v1beta1.ChainStateResponse) | ChainState queries the state of a registered chain on the network | GET|/axelar/nexus/v1beta1/chain_state/{chain}|
| `ChainsByAsset` | [ChainsByAssetRequest](#axelarnetwork.nexus.v1beta1.ChainsByAssetRequest) | [ChainsByAssetResponse](#axelarnetwork.nexus.v1beta1.ChainsByAssetResponse) | ChainsByAsset queries the chains that support an asset on the network | GET|/axelar/nexus/v1beta1/chains_by_asset/{asset}|

 <!-- end services -->



<a name="axelarnetwork/permission/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/permission/v1beta1/types.proto



<a name="axelarnetwork.permission.v1beta1.GovAccount"></a>

### GovAccount



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [bytes](#bytes) |  |  |
| `role` | [axelarnetwork.permission.exported.v1beta1.Role](#axelarnetwork.permission.exported.v1beta1.Role) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/permission/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/permission/v1beta1/params.proto



<a name="axelarnetwork.permission.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/permission/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/permission/v1beta1/genesis.proto



<a name="axelarnetwork.permission.v1beta1.GenesisState"></a>

### GenesisState
GenesisState represents the genesis state


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelarnetwork.permission.v1beta1.Params) |  |  |
| `governance_key` | [cosmos.crypto.multisig.LegacyAminoPubKey](#cosmos.crypto.multisig.LegacyAminoPubKey) |  |  |
| `gov_accounts` | [GovAccount](#axelarnetwork.permission.v1beta1.GovAccount) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/permission/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/permission/v1beta1/query.proto



<a name="axelarnetwork.permission.v1beta1.QueryGovernanceKeyRequest"></a>

### QueryGovernanceKeyRequest
QueryGovernanceKeyRequest is the request type for the
Query/GovernanceKey RPC method






<a name="axelarnetwork.permission.v1beta1.QueryGovernanceKeyResponse"></a>

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



<a name="axelarnetwork/permission/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/permission/v1beta1/tx.proto



<a name="axelarnetwork.permission.v1beta1.DeregisterControllerRequest"></a>

### DeregisterControllerRequest
DeregisterController represents a message to deregister a controller account


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `controller` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.permission.v1beta1.DeregisterControllerResponse"></a>

### DeregisterControllerResponse







<a name="axelarnetwork.permission.v1beta1.RegisterControllerRequest"></a>

### RegisterControllerRequest
MsgRegisterController represents a message to register a controller account


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `controller` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.permission.v1beta1.RegisterControllerResponse"></a>

### RegisterControllerResponse







<a name="axelarnetwork.permission.v1beta1.UpdateGovernanceKeyRequest"></a>

### UpdateGovernanceKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `governance_key` | [cosmos.crypto.multisig.LegacyAminoPubKey](#cosmos.crypto.multisig.LegacyAminoPubKey) |  |  |






<a name="axelarnetwork.permission.v1beta1.UpdateGovernanceKeyResponse"></a>

### UpdateGovernanceKeyResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/permission/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/permission/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelarnetwork.permission.v1beta1.Msg"></a>

### Msg
Msg defines the gov Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterController` | [RegisterControllerRequest](#axelarnetwork.permission.v1beta1.RegisterControllerRequest) | [RegisterControllerResponse](#axelarnetwork.permission.v1beta1.RegisterControllerResponse) |  | ||
| `DeregisterController` | [DeregisterControllerRequest](#axelarnetwork.permission.v1beta1.DeregisterControllerRequest) | [DeregisterControllerResponse](#axelarnetwork.permission.v1beta1.DeregisterControllerResponse) |  | ||
| `UpdateGovernanceKey` | [UpdateGovernanceKeyRequest](#axelarnetwork.permission.v1beta1.UpdateGovernanceKeyRequest) | [UpdateGovernanceKeyResponse](#axelarnetwork.permission.v1beta1.UpdateGovernanceKeyResponse) |  | ||


<a name="axelarnetwork.permission.v1beta1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `GovernanceKey` | [QueryGovernanceKeyRequest](#axelarnetwork.permission.v1beta1.QueryGovernanceKeyRequest) | [QueryGovernanceKeyResponse](#axelarnetwork.permission.v1beta1.QueryGovernanceKeyResponse) | GovernanceKey returns multisig governance key | GET|/permission/v1beta1/governance_key|

 <!-- end services -->



<a name="axelarnetwork/reward/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/reward/v1beta1/params.proto



<a name="axelarnetwork.reward.v1beta1.Params"></a>

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



<a name="axelarnetwork/reward/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/reward/v1beta1/types.proto



<a name="axelarnetwork.reward.v1beta1.Pool"></a>

### Pool



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `rewards` | [Pool.Reward](#axelarnetwork.reward.v1beta1.Pool.Reward) | repeated |  |






<a name="axelarnetwork.reward.v1beta1.Pool.Reward"></a>

### Pool.Reward



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validator` | [bytes](#bytes) |  |  |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="axelarnetwork.reward.v1beta1.Refund"></a>

### Refund



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `payer` | [bytes](#bytes) |  |  |
| `fees` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/reward/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/reward/v1beta1/genesis.proto



<a name="axelarnetwork.reward.v1beta1.GenesisState"></a>

### GenesisState
GenesisState represents the genesis state


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelarnetwork.reward.v1beta1.Params) |  |  |
| `pools` | [Pool](#axelarnetwork.reward.v1beta1.Pool) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/reward/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/reward/v1beta1/tx.proto



<a name="axelarnetwork.reward.v1beta1.RefundMsgRequest"></a>

### RefundMsgRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `inner_message` | [google.protobuf.Any](#google.protobuf.Any) |  |  |






<a name="axelarnetwork.reward.v1beta1.RefundMsgResponse"></a>

### RefundMsgResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  |  |
| `log` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/reward/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/reward/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelarnetwork.reward.v1beta1.MsgService"></a>

### MsgService
Msg defines the axelarnet Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RefundMsg` | [RefundMsgRequest](#axelarnetwork.reward.v1beta1.RefundMsgRequest) | [RefundMsgResponse](#axelarnetwork.reward.v1beta1.RefundMsgResponse) |  | POST|/reward/refund-message|

 <!-- end services -->



<a name="axelarnetwork/snapshot/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/snapshot/v1beta1/params.proto



<a name="axelarnetwork.snapshot.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `min_proxy_balance` | [int64](#int64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/snapshot/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/snapshot/v1beta1/types.proto



<a name="axelarnetwork.snapshot.v1beta1.ProxiedValidator"></a>

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



<a name="axelarnetwork/snapshot/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/snapshot/v1beta1/genesis.proto



<a name="axelarnetwork.snapshot.v1beta1.GenesisState"></a>

### GenesisState
GenesisState represents the genesis state


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelarnetwork.snapshot.v1beta1.Params) |  |  |
| `snapshots` | [axelarnetwork.snapshot.exported.v1beta1.Snapshot](#axelarnetwork.snapshot.exported.v1beta1.Snapshot) | repeated |  |
| `proxied_validators` | [ProxiedValidator](#axelarnetwork.snapshot.v1beta1.ProxiedValidator) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/snapshot/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/snapshot/v1beta1/query.proto



<a name="axelarnetwork.snapshot.v1beta1.QueryValidatorsResponse"></a>

### QueryValidatorsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validators` | [QueryValidatorsResponse.Validator](#axelarnetwork.snapshot.v1beta1.QueryValidatorsResponse.Validator) | repeated |  |






<a name="axelarnetwork.snapshot.v1beta1.QueryValidatorsResponse.TssIllegibilityInfo"></a>

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






<a name="axelarnetwork.snapshot.v1beta1.QueryValidatorsResponse.Validator"></a>

### QueryValidatorsResponse.Validator



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `operator_address` | [string](#string) |  |  |
| `moniker` | [string](#string) |  |  |
| `tss_illegibility_info` | [QueryValidatorsResponse.TssIllegibilityInfo](#axelarnetwork.snapshot.v1beta1.QueryValidatorsResponse.TssIllegibilityInfo) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/snapshot/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/snapshot/v1beta1/tx.proto



<a name="axelarnetwork.snapshot.v1beta1.DeactivateProxyRequest"></a>

### DeactivateProxyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.snapshot.v1beta1.DeactivateProxyResponse"></a>

### DeactivateProxyResponse







<a name="axelarnetwork.snapshot.v1beta1.RegisterProxyRequest"></a>

### RegisterProxyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `proxy_addr` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.snapshot.v1beta1.RegisterProxyResponse"></a>

### RegisterProxyResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/snapshot/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/snapshot/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelarnetwork.snapshot.v1beta1.MsgService"></a>

### MsgService
Msg defines the snapshot Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterProxy` | [RegisterProxyRequest](#axelarnetwork.snapshot.v1beta1.RegisterProxyRequest) | [RegisterProxyResponse](#axelarnetwork.snapshot.v1beta1.RegisterProxyResponse) | RegisterProxy defines a method for registering a proxy account that can act in a validator account's stead. | POST|/axelar/snapshot/registerProxy/{proxy_addr}|
| `DeactivateProxy` | [DeactivateProxyRequest](#axelarnetwork.snapshot.v1beta1.DeactivateProxyRequest) | [DeactivateProxyResponse](#axelarnetwork.snapshot.v1beta1.DeactivateProxyResponse) | DeactivateProxy defines a method for deregistering a proxy account. | POST|/axelar/snapshot/deactivateProxy|

 <!-- end services -->



<a name="axelarnetwork/tss/tofnd/v1beta1/common.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/tss/tofnd/v1beta1/common.proto
File copied from golang tofnd with minor tweaks


<a name="axelarnetwork.tss.tofnd.v1beta1.KeyPresenceRequest"></a>

### KeyPresenceRequest
Key presence check types


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_uid` | [string](#string) |  |  |






<a name="axelarnetwork.tss.tofnd.v1beta1.KeyPresenceResponse"></a>

### KeyPresenceResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `response` | [KeyPresenceResponse.Response](#axelarnetwork.tss.tofnd.v1beta1.KeyPresenceResponse.Response) |  |  |





 <!-- end messages -->


<a name="axelarnetwork.tss.tofnd.v1beta1.KeyPresenceResponse.Response"></a>

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



<a name="axelarnetwork/tss/tofnd/v1beta1/multisig.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/tss/tofnd/v1beta1/multisig.proto
File copied from golang tofnd with minor tweaks


<a name="axelarnetwork.tss.tofnd.v1beta1.KeygenRequest"></a>

### KeygenRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_uid` | [string](#string) |  |  |
| `party_uid` | [string](#string) |  | used only for logging |






<a name="axelarnetwork.tss.tofnd.v1beta1.KeygenResponse"></a>

### KeygenResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pub_key` | [bytes](#bytes) |  | SEC1-encoded compressed curve point |
| `error` | [string](#string) |  | reply with an error message if keygen fails |






<a name="axelarnetwork.tss.tofnd.v1beta1.SignRequest"></a>

### SignRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_uid` | [string](#string) |  |  |
| `msg_to_sign` | [bytes](#bytes) |  | 32-byte pre-hashed message digest |
| `party_uid` | [string](#string) |  | used only for logging |






<a name="axelarnetwork.tss.tofnd.v1beta1.SignResponse"></a>

### SignResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signature` | [bytes](#bytes) |  | ASN.1 DER-encoded ECDSA signature |
| `error` | [string](#string) |  | reply with an error message if sign fails |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/tss/tofnd/v1beta1/tofnd.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/tss/tofnd/v1beta1/tofnd.proto
File copied from golang tofnd with minor tweaks


<a name="axelarnetwork.tss.tofnd.v1beta1.KeygenInit"></a>

### KeygenInit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `new_key_uid` | [string](#string) |  |  |
| `party_uids` | [string](#string) | repeated |  |
| `party_share_counts` | [uint32](#uint32) | repeated |  |
| `my_party_index` | [uint32](#uint32) |  | parties[my_party_index] belongs to the server |
| `threshold` | [uint32](#uint32) |  |  |






<a name="axelarnetwork.tss.tofnd.v1beta1.KeygenOutput"></a>

### KeygenOutput
Keygen's success response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pub_key` | [bytes](#bytes) |  | pub_key; common for all parties |
| `group_recover_info` | [bytes](#bytes) |  | recover info of all parties' shares; common for all parties |
| `private_recover_info` | [bytes](#bytes) |  | private recover info of this party's shares; unique for each party |






<a name="axelarnetwork.tss.tofnd.v1beta1.MessageIn"></a>

### MessageIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `keygen_init` | [KeygenInit](#axelarnetwork.tss.tofnd.v1beta1.KeygenInit) |  | first message only, Keygen |
| `sign_init` | [SignInit](#axelarnetwork.tss.tofnd.v1beta1.SignInit) |  | first message only, Sign |
| `traffic` | [TrafficIn](#axelarnetwork.tss.tofnd.v1beta1.TrafficIn) |  | all subsequent messages |
| `abort` | [bool](#bool) |  | abort the protocol, ignore the bool value |






<a name="axelarnetwork.tss.tofnd.v1beta1.MessageOut"></a>

### MessageOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `traffic` | [TrafficOut](#axelarnetwork.tss.tofnd.v1beta1.TrafficOut) |  | all but final message |
| `keygen_result` | [MessageOut.KeygenResult](#axelarnetwork.tss.tofnd.v1beta1.MessageOut.KeygenResult) |  | final message only, Keygen |
| `sign_result` | [MessageOut.SignResult](#axelarnetwork.tss.tofnd.v1beta1.MessageOut.SignResult) |  | final message only, Sign |
| `need_recover` | [bool](#bool) |  | issue recover from client |






<a name="axelarnetwork.tss.tofnd.v1beta1.MessageOut.CriminalList"></a>

### MessageOut.CriminalList
Keygen/Sign failure response message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `criminals` | [MessageOut.CriminalList.Criminal](#axelarnetwork.tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal) | repeated |  |






<a name="axelarnetwork.tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal"></a>

### MessageOut.CriminalList.Criminal



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `party_uid` | [string](#string) |  |  |
| `crime_type` | [MessageOut.CriminalList.Criminal.CrimeType](#axelarnetwork.tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal.CrimeType) |  |  |






<a name="axelarnetwork.tss.tofnd.v1beta1.MessageOut.KeygenResult"></a>

### MessageOut.KeygenResult
Keygen's response types


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [KeygenOutput](#axelarnetwork.tss.tofnd.v1beta1.KeygenOutput) |  | Success response |
| `criminals` | [MessageOut.CriminalList](#axelarnetwork.tss.tofnd.v1beta1.MessageOut.CriminalList) |  | Faiilure response |






<a name="axelarnetwork.tss.tofnd.v1beta1.MessageOut.SignResult"></a>

### MessageOut.SignResult
Sign's response types


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signature` | [bytes](#bytes) |  | Success response |
| `criminals` | [MessageOut.CriminalList](#axelarnetwork.tss.tofnd.v1beta1.MessageOut.CriminalList) |  | Failure response |






<a name="axelarnetwork.tss.tofnd.v1beta1.RecoverRequest"></a>

### RecoverRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `keygen_init` | [KeygenInit](#axelarnetwork.tss.tofnd.v1beta1.KeygenInit) |  |  |
| `keygen_output` | [KeygenOutput](#axelarnetwork.tss.tofnd.v1beta1.KeygenOutput) |  |  |






<a name="axelarnetwork.tss.tofnd.v1beta1.RecoverResponse"></a>

### RecoverResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `response` | [RecoverResponse.Response](#axelarnetwork.tss.tofnd.v1beta1.RecoverResponse.Response) |  |  |






<a name="axelarnetwork.tss.tofnd.v1beta1.SignInit"></a>

### SignInit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `new_sig_uid` | [string](#string) |  |  |
| `key_uid` | [string](#string) |  |  |
| `party_uids` | [string](#string) | repeated | TODO replace this with a subset of indices? |
| `message_to_sign` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.tss.tofnd.v1beta1.TrafficIn"></a>

### TrafficIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `from_party_uid` | [string](#string) |  |  |
| `payload` | [bytes](#bytes) |  |  |
| `is_broadcast` | [bool](#bool) |  |  |






<a name="axelarnetwork.tss.tofnd.v1beta1.TrafficOut"></a>

### TrafficOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to_party_uid` | [string](#string) |  |  |
| `payload` | [bytes](#bytes) |  |  |
| `is_broadcast` | [bool](#bool) |  |  |





 <!-- end messages -->


<a name="axelarnetwork.tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal.CrimeType"></a>

### MessageOut.CriminalList.Criminal.CrimeType


| Name | Number | Description |
| ---- | ------ | ----------- |
| CRIME_TYPE_UNSPECIFIED | 0 |  |
| CRIME_TYPE_NON_MALICIOUS | 1 |  |
| CRIME_TYPE_MALICIOUS | 2 |  |



<a name="axelarnetwork.tss.tofnd.v1beta1.RecoverResponse.Response"></a>

### RecoverResponse.Response


| Name | Number | Description |
| ---- | ------ | ----------- |
| RESPONSE_UNSPECIFIED | 0 |  |
| RESPONSE_SUCCESS | 1 |  |
| RESPONSE_FAIL | 2 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/tss/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/tss/v1beta1/params.proto



<a name="axelarnetwork.tss.v1beta1.Params"></a>

### Params
Params is the parameter set for this module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_requirements` | [axelarnetwork.tss.exported.v1beta1.KeyRequirement](#axelarnetwork.tss.exported.v1beta1.KeyRequirement) | repeated | KeyRequirements defines the requirement for each key role |
| `suspend_duration_in_blocks` | [int64](#int64) |  | SuspendDurationInBlocks defines the number of blocks a validator is disallowed to participate in any TSS ceremony after committing a malicious behaviour during signing |
| `heartbeat_period_in_blocks` | [int64](#int64) |  | HeartBeatPeriodInBlocks defines the time period in blocks for tss to emit the event asking validators to send their heartbeats |
| `max_missed_blocks_per_window` | [axelarnetwork.utils.v1beta1.Threshold](#axelarnetwork.utils.v1beta1.Threshold) |  |  |
| `unbonding_locking_key_rotation_count` | [int64](#int64) |  |  |
| `external_multisig_threshold` | [axelarnetwork.utils.v1beta1.Threshold](#axelarnetwork.utils.v1beta1.Threshold) |  |  |
| `max_sign_queue_size` | [int64](#int64) |  |  |
| `max_simultaneous_sign_shares` | [int64](#int64) |  |  |
| `tss_signed_blocks_window` | [int64](#int64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/tss/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/tss/v1beta1/types.proto



<a name="axelarnetwork.tss.v1beta1.ExternalKeys"></a>

### ExternalKeys



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `key_ids` | [string](#string) | repeated |  |






<a name="axelarnetwork.tss.v1beta1.KeyInfo"></a>

### KeyInfo
KeyInfo holds information about a key


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `key_role` | [axelarnetwork.tss.exported.v1beta1.KeyRole](#axelarnetwork.tss.exported.v1beta1.KeyRole) |  |  |
| `key_type` | [axelarnetwork.tss.exported.v1beta1.KeyType](#axelarnetwork.tss.exported.v1beta1.KeyType) |  |  |






<a name="axelarnetwork.tss.v1beta1.KeyRecoveryInfo"></a>

### KeyRecoveryInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `public` | [bytes](#bytes) |  |  |
| `private` | [KeyRecoveryInfo.PrivateEntry](#axelarnetwork.tss.v1beta1.KeyRecoveryInfo.PrivateEntry) | repeated |  |






<a name="axelarnetwork.tss.v1beta1.KeyRecoveryInfo.PrivateEntry"></a>

### KeyRecoveryInfo.PrivateEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.tss.v1beta1.KeygenVoteData"></a>

### KeygenVoteData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pub_key` | [bytes](#bytes) |  |  |
| `group_recovery_info` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.tss.v1beta1.MultisigInfo"></a>

### MultisigInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `timeout` | [int64](#int64) |  |  |
| `target_num` | [int64](#int64) |  |  |
| `infos` | [MultisigInfo.Info](#axelarnetwork.tss.v1beta1.MultisigInfo.Info) | repeated |  |






<a name="axelarnetwork.tss.v1beta1.MultisigInfo.Info"></a>

### MultisigInfo.Info



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `participant` | [bytes](#bytes) |  |  |
| `data` | [bytes](#bytes) | repeated |  |






<a name="axelarnetwork.tss.v1beta1.ValidatorStatus"></a>

### ValidatorStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validator` | [bytes](#bytes) |  |  |
| `suspended_until` | [uint64](#uint64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/tss/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/tss/v1beta1/genesis.proto



<a name="axelarnetwork.tss.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelarnetwork.tss.v1beta1.Params) |  |  |
| `key_recovery_infos` | [KeyRecoveryInfo](#axelarnetwork.tss.v1beta1.KeyRecoveryInfo) | repeated |  |
| `keys` | [axelarnetwork.tss.exported.v1beta1.Key](#axelarnetwork.tss.exported.v1beta1.Key) | repeated |  |
| `multisig_infos` | [MultisigInfo](#axelarnetwork.tss.v1beta1.MultisigInfo) | repeated |  |
| `external_keys` | [ExternalKeys](#axelarnetwork.tss.v1beta1.ExternalKeys) | repeated |  |
| `signatures` | [axelarnetwork.tss.exported.v1beta1.Signature](#axelarnetwork.tss.exported.v1beta1.Signature) | repeated |  |
| `validator_statuses` | [ValidatorStatus](#axelarnetwork.tss.v1beta1.ValidatorStatus) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/tss/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/tss/v1beta1/query.proto



<a name="axelarnetwork.tss.v1beta1.AssignableKeyRequest"></a>

### AssignableKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `key_role` | [axelarnetwork.tss.exported.v1beta1.KeyRole](#axelarnetwork.tss.exported.v1beta1.KeyRole) |  |  |






<a name="axelarnetwork.tss.v1beta1.AssignableKeyResponse"></a>

### AssignableKeyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `assignable` | [bool](#bool) |  |  |






<a name="axelarnetwork.tss.v1beta1.NextKeyIDRequest"></a>

### NextKeyIDRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `key_role` | [axelarnetwork.tss.exported.v1beta1.KeyRole](#axelarnetwork.tss.exported.v1beta1.KeyRole) |  |  |






<a name="axelarnetwork.tss.v1beta1.NextKeyIDResponse"></a>

### NextKeyIDResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |






<a name="axelarnetwork.tss.v1beta1.QueryActiveOldKeysResponse"></a>

### QueryActiveOldKeysResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_ids` | [string](#string) | repeated |  |






<a name="axelarnetwork.tss.v1beta1.QueryActiveOldKeysValidatorResponse"></a>

### QueryActiveOldKeysValidatorResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `keys_info` | [QueryActiveOldKeysValidatorResponse.KeyInfo](#axelarnetwork.tss.v1beta1.QueryActiveOldKeysValidatorResponse.KeyInfo) | repeated |  |






<a name="axelarnetwork.tss.v1beta1.QueryActiveOldKeysValidatorResponse.KeyInfo"></a>

### QueryActiveOldKeysValidatorResponse.KeyInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `chain` | [string](#string) |  |  |
| `role` | [int32](#int32) |  |  |






<a name="axelarnetwork.tss.v1beta1.QueryDeactivatedOperatorsResponse"></a>

### QueryDeactivatedOperatorsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `operator_addresses` | [string](#string) | repeated |  |






<a name="axelarnetwork.tss.v1beta1.QueryExternalKeyIDResponse"></a>

### QueryExternalKeyIDResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_ids` | [string](#string) | repeated |  |






<a name="axelarnetwork.tss.v1beta1.QueryKeyResponse"></a>

### QueryKeyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ecdsa_key` | [QueryKeyResponse.ECDSAKey](#axelarnetwork.tss.v1beta1.QueryKeyResponse.ECDSAKey) |  |  |
| `multisig_key` | [QueryKeyResponse.MultisigKey](#axelarnetwork.tss.v1beta1.QueryKeyResponse.MultisigKey) |  |  |
| `role` | [axelarnetwork.tss.exported.v1beta1.KeyRole](#axelarnetwork.tss.exported.v1beta1.KeyRole) |  |  |
| `rotated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="axelarnetwork.tss.v1beta1.QueryKeyResponse.ECDSAKey"></a>

### QueryKeyResponse.ECDSAKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `vote_status` | [VoteStatus](#axelarnetwork.tss.v1beta1.VoteStatus) |  |  |
| `key` | [QueryKeyResponse.Key](#axelarnetwork.tss.v1beta1.QueryKeyResponse.Key) |  |  |






<a name="axelarnetwork.tss.v1beta1.QueryKeyResponse.Key"></a>

### QueryKeyResponse.Key



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `x` | [string](#string) |  |  |
| `y` | [string](#string) |  |  |






<a name="axelarnetwork.tss.v1beta1.QueryKeyResponse.MultisigKey"></a>

### QueryKeyResponse.MultisigKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `threshold` | [int64](#int64) |  |  |
| `key` | [QueryKeyResponse.Key](#axelarnetwork.tss.v1beta1.QueryKeyResponse.Key) | repeated |  |






<a name="axelarnetwork.tss.v1beta1.QueryKeyShareResponse"></a>

### QueryKeyShareResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `share_infos` | [QueryKeyShareResponse.ShareInfo](#axelarnetwork.tss.v1beta1.QueryKeyShareResponse.ShareInfo) | repeated |  |






<a name="axelarnetwork.tss.v1beta1.QueryKeyShareResponse.ShareInfo"></a>

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






<a name="axelarnetwork.tss.v1beta1.QueryRecoveryResponse"></a>

### QueryRecoveryResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `party_uids` | [string](#string) | repeated |  |
| `party_share_counts` | [uint32](#uint32) | repeated |  |
| `threshold` | [uint32](#uint32) |  |  |
| `keygen_output` | [axelarnetwork.tss.tofnd.v1beta1.KeygenOutput](#axelarnetwork.tss.tofnd.v1beta1.KeygenOutput) |  |  |






<a name="axelarnetwork.tss.v1beta1.QuerySignatureResponse"></a>

### QuerySignatureResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `threshold_signature` | [QuerySignatureResponse.ThresholdSignature](#axelarnetwork.tss.v1beta1.QuerySignatureResponse.ThresholdSignature) |  |  |
| `multisig_signature` | [QuerySignatureResponse.MultisigSignature](#axelarnetwork.tss.v1beta1.QuerySignatureResponse.MultisigSignature) |  |  |






<a name="axelarnetwork.tss.v1beta1.QuerySignatureResponse.MultisigSignature"></a>

### QuerySignatureResponse.MultisigSignature



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sig_status` | [axelarnetwork.tss.exported.v1beta1.SigStatus](#axelarnetwork.tss.exported.v1beta1.SigStatus) |  |  |
| `signatures` | [QuerySignatureResponse.Signature](#axelarnetwork.tss.v1beta1.QuerySignatureResponse.Signature) | repeated |  |






<a name="axelarnetwork.tss.v1beta1.QuerySignatureResponse.Signature"></a>

### QuerySignatureResponse.Signature



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `r` | [string](#string) |  |  |
| `s` | [string](#string) |  |  |






<a name="axelarnetwork.tss.v1beta1.QuerySignatureResponse.ThresholdSignature"></a>

### QuerySignatureResponse.ThresholdSignature



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `vote_status` | [VoteStatus](#axelarnetwork.tss.v1beta1.VoteStatus) |  |  |
| `signature` | [QuerySignatureResponse.Signature](#axelarnetwork.tss.v1beta1.QuerySignatureResponse.Signature) |  |  |





 <!-- end messages -->


<a name="axelarnetwork.tss.v1beta1.VoteStatus"></a>

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



<a name="axelarnetwork/tss/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/tss/v1beta1/tx.proto



<a name="axelarnetwork.tss.v1beta1.HeartBeatRequest"></a>

### HeartBeatRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `key_ids` | [string](#string) | repeated |  |






<a name="axelarnetwork.tss.v1beta1.HeartBeatResponse"></a>

### HeartBeatResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `keygen_illegibility` | [int32](#int32) |  |  |
| `signing_illegibility` | [int32](#int32) |  |  |






<a name="axelarnetwork.tss.v1beta1.ProcessKeygenTrafficRequest"></a>

### ProcessKeygenTrafficRequest
ProcessKeygenTrafficRequest protocol message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `session_id` | [string](#string) |  |  |
| `payload` | [axelarnetwork.tss.tofnd.v1beta1.TrafficOut](#axelarnetwork.tss.tofnd.v1beta1.TrafficOut) |  |  |






<a name="axelarnetwork.tss.v1beta1.ProcessKeygenTrafficResponse"></a>

### ProcessKeygenTrafficResponse







<a name="axelarnetwork.tss.v1beta1.ProcessSignTrafficRequest"></a>

### ProcessSignTrafficRequest
ProcessSignTrafficRequest protocol message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `session_id` | [string](#string) |  |  |
| `payload` | [axelarnetwork.tss.tofnd.v1beta1.TrafficOut](#axelarnetwork.tss.tofnd.v1beta1.TrafficOut) |  |  |






<a name="axelarnetwork.tss.v1beta1.ProcessSignTrafficResponse"></a>

### ProcessSignTrafficResponse







<a name="axelarnetwork.tss.v1beta1.RegisterExternalKeysRequest"></a>

### RegisterExternalKeysRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `external_keys` | [RegisterExternalKeysRequest.ExternalKey](#axelarnetwork.tss.v1beta1.RegisterExternalKeysRequest.ExternalKey) | repeated |  |






<a name="axelarnetwork.tss.v1beta1.RegisterExternalKeysRequest.ExternalKey"></a>

### RegisterExternalKeysRequest.ExternalKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `pub_key` | [bytes](#bytes) |  |  |






<a name="axelarnetwork.tss.v1beta1.RegisterExternalKeysResponse"></a>

### RegisterExternalKeysResponse







<a name="axelarnetwork.tss.v1beta1.RotateKeyRequest"></a>

### RotateKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `key_role` | [axelarnetwork.tss.exported.v1beta1.KeyRole](#axelarnetwork.tss.exported.v1beta1.KeyRole) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="axelarnetwork.tss.v1beta1.RotateKeyResponse"></a>

### RotateKeyResponse







<a name="axelarnetwork.tss.v1beta1.StartKeygenRequest"></a>

### StartKeygenRequest
StartKeygenRequest indicate the start of keygen


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `key_info` | [KeyInfo](#axelarnetwork.tss.v1beta1.KeyInfo) |  |  |






<a name="axelarnetwork.tss.v1beta1.StartKeygenResponse"></a>

### StartKeygenResponse







<a name="axelarnetwork.tss.v1beta1.SubmitMultisigPubKeysRequest"></a>

### SubmitMultisigPubKeysRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `sig_key_pairs` | [axelarnetwork.tss.exported.v1beta1.SigKeyPair](#axelarnetwork.tss.exported.v1beta1.SigKeyPair) | repeated |  |






<a name="axelarnetwork.tss.v1beta1.SubmitMultisigPubKeysResponse"></a>

### SubmitMultisigPubKeysResponse







<a name="axelarnetwork.tss.v1beta1.SubmitMultisigSignaturesRequest"></a>

### SubmitMultisigSignaturesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `sig_id` | [string](#string) |  |  |
| `signatures` | [bytes](#bytes) | repeated |  |






<a name="axelarnetwork.tss.v1beta1.SubmitMultisigSignaturesResponse"></a>

### SubmitMultisigSignaturesResponse







<a name="axelarnetwork.tss.v1beta1.VotePubKeyRequest"></a>

### VotePubKeyRequest
VotePubKeyRequest represents the message to vote on a public key


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll_key` | [axelarnetwork.vote.exported.v1beta1.PollKey](#axelarnetwork.vote.exported.v1beta1.PollKey) |  |  |
| `result` | [axelarnetwork.tss.tofnd.v1beta1.MessageOut.KeygenResult](#axelarnetwork.tss.tofnd.v1beta1.MessageOut.KeygenResult) |  |  |






<a name="axelarnetwork.tss.v1beta1.VotePubKeyResponse"></a>

### VotePubKeyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |






<a name="axelarnetwork.tss.v1beta1.VoteSigRequest"></a>

### VoteSigRequest
VoteSigRequest represents a message to vote for a signature


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll_key` | [axelarnetwork.vote.exported.v1beta1.PollKey](#axelarnetwork.vote.exported.v1beta1.PollKey) |  |  |
| `result` | [axelarnetwork.tss.tofnd.v1beta1.MessageOut.SignResult](#axelarnetwork.tss.tofnd.v1beta1.MessageOut.SignResult) |  |  |






<a name="axelarnetwork.tss.v1beta1.VoteSigResponse"></a>

### VoteSigResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/tss/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/tss/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelarnetwork.tss.v1beta1.MsgService"></a>

### MsgService
Msg defines the tss Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterExternalKeys` | [RegisterExternalKeysRequest](#axelarnetwork.tss.v1beta1.RegisterExternalKeysRequest) | [RegisterExternalKeysResponse](#axelarnetwork.tss.v1beta1.RegisterExternalKeysResponse) |  | POST|/axelar/tss/register-external-key|
| `HeartBeat` | [HeartBeatRequest](#axelarnetwork.tss.v1beta1.HeartBeatRequest) | [HeartBeatResponse](#axelarnetwork.tss.v1beta1.HeartBeatResponse) |  | POST|/axelar/tss/heartbeat|
| `StartKeygen` | [StartKeygenRequest](#axelarnetwork.tss.v1beta1.StartKeygenRequest) | [StartKeygenResponse](#axelarnetwork.tss.v1beta1.StartKeygenResponse) |  | POST|/axelar/tss/startKeygen|
| `ProcessKeygenTraffic` | [ProcessKeygenTrafficRequest](#axelarnetwork.tss.v1beta1.ProcessKeygenTrafficRequest) | [ProcessKeygenTrafficResponse](#axelarnetwork.tss.v1beta1.ProcessKeygenTrafficResponse) |  | ||
| `RotateKey` | [RotateKeyRequest](#axelarnetwork.tss.v1beta1.RotateKeyRequest) | [RotateKeyResponse](#axelarnetwork.tss.v1beta1.RotateKeyResponse) |  | POST|/axelar/tss/assign/{chain}|
| `VotePubKey` | [VotePubKeyRequest](#axelarnetwork.tss.v1beta1.VotePubKeyRequest) | [VotePubKeyResponse](#axelarnetwork.tss.v1beta1.VotePubKeyResponse) |  | ||
| `ProcessSignTraffic` | [ProcessSignTrafficRequest](#axelarnetwork.tss.v1beta1.ProcessSignTrafficRequest) | [ProcessSignTrafficResponse](#axelarnetwork.tss.v1beta1.ProcessSignTrafficResponse) |  | ||
| `VoteSig` | [VoteSigRequest](#axelarnetwork.tss.v1beta1.VoteSigRequest) | [VoteSigResponse](#axelarnetwork.tss.v1beta1.VoteSigResponse) |  | ||
| `SubmitMultisigPubKeys` | [SubmitMultisigPubKeysRequest](#axelarnetwork.tss.v1beta1.SubmitMultisigPubKeysRequest) | [SubmitMultisigPubKeysResponse](#axelarnetwork.tss.v1beta1.SubmitMultisigPubKeysResponse) |  | ||
| `SubmitMultisigSignatures` | [SubmitMultisigSignaturesRequest](#axelarnetwork.tss.v1beta1.SubmitMultisigSignaturesRequest) | [SubmitMultisigSignaturesResponse](#axelarnetwork.tss.v1beta1.SubmitMultisigSignaturesResponse) |  | ||


<a name="axelarnetwork.tss.v1beta1.QueryService"></a>

### QueryService
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `NextKeyID` | [NextKeyIDRequest](#axelarnetwork.tss.v1beta1.NextKeyIDRequest) | [NextKeyIDResponse](#axelarnetwork.tss.v1beta1.NextKeyIDResponse) | NextKeyID returns the key ID assigned for the next rotation on a given chain and for the given key role | GET|/tss/v1beta1/next_key_id|
| `AssignableKey` | [AssignableKeyRequest](#axelarnetwork.tss.v1beta1.AssignableKeyRequest) | [AssignableKeyResponse](#axelarnetwork.tss.v1beta1.AssignableKeyResponse) | AssignableKey returns true if there is no assigned key for the next rotation on a given chain, and false otherwise | GET|/tss/v1beta1/assignable_key|

 <!-- end services -->



<a name="axelarnetwork/vote/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/vote/v1beta1/params.proto



<a name="axelarnetwork.vote.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `default_voting_threshold` | [axelarnetwork.utils.v1beta1.Threshold](#axelarnetwork.utils.v1beta1.Threshold) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/vote/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/vote/v1beta1/genesis.proto



<a name="axelarnetwork.vote.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelarnetwork.vote.v1beta1.Params) |  |  |
| `poll_metadatas` | [axelarnetwork.vote.exported.v1beta1.PollMetadata](#axelarnetwork.vote.exported.v1beta1.PollMetadata) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/vote/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/vote/v1beta1/tx.proto



<a name="axelarnetwork.vote.v1beta1.VoteRequest"></a>

### VoteRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll_key` | [axelarnetwork.vote.exported.v1beta1.PollKey](#axelarnetwork.vote.exported.v1beta1.PollKey) |  |  |
| `vote` | [axelarnetwork.vote.exported.v1beta1.Vote](#axelarnetwork.vote.exported.v1beta1.Vote) |  |  |






<a name="axelarnetwork.vote.v1beta1.VoteResponse"></a>

### VoteResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnetwork/vote/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/vote/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelarnetwork.vote.v1beta1.MsgService"></a>

### MsgService
Msg defines the vote Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Vote` | [VoteRequest](#axelarnetwork.vote.v1beta1.VoteRequest) | [VoteResponse](#axelarnetwork.vote.v1beta1.VoteResponse) |  | POST|/axelar/vote/vote|

 <!-- end services -->



<a name="axelarnetwork/vote/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnetwork/vote/v1beta1/types.proto



<a name="axelarnetwork.vote.v1beta1.TalliedVote"></a>

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

