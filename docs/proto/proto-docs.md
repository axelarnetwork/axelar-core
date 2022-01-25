<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [axelarnet/v1beta1/params.proto](#axelarnet/v1beta1/params.proto)
    - [Params](#axelarnet.v1beta1.Params)
  
- [axelarnet/v1beta1/types.proto](#axelarnet/v1beta1/types.proto)
    - [Asset](#axelarnet.v1beta1.Asset)
    - [CosmosChain](#axelarnet.v1beta1.CosmosChain)
    - [IBCTransfer](#axelarnet.v1beta1.IBCTransfer)
  
- [axelarnet/v1beta1/genesis.proto](#axelarnet/v1beta1/genesis.proto)
    - [GenesisState](#axelarnet.v1beta1.GenesisState)
  
- [utils/v1beta1/threshold.proto](#utils/v1beta1/threshold.proto)
    - [Threshold](#utils.v1beta1.Threshold)
  
- [tss/exported/v1beta1/types.proto](#tss/exported/v1beta1/types.proto)
    - [Key](#tss.exported.v1beta1.Key)
    - [Key.ECDSAKey](#tss.exported.v1beta1.Key.ECDSAKey)
    - [Key.MultisigKey](#tss.exported.v1beta1.Key.MultisigKey)
    - [KeyRequirement](#tss.exported.v1beta1.KeyRequirement)
    - [SigKeyPair](#tss.exported.v1beta1.SigKeyPair)
    - [SignInfo](#tss.exported.v1beta1.SignInfo)
    - [Signature](#tss.exported.v1beta1.Signature)
    - [Signature.MultiSig](#tss.exported.v1beta1.Signature.MultiSig)
    - [Signature.SingleSig](#tss.exported.v1beta1.Signature.SingleSig)
  
    - [AckType](#tss.exported.v1beta1.AckType)
    - [KeyRole](#tss.exported.v1beta1.KeyRole)
    - [KeyShareDistributionPolicy](#tss.exported.v1beta1.KeyShareDistributionPolicy)
    - [KeyType](#tss.exported.v1beta1.KeyType)
    - [SigStatus](#tss.exported.v1beta1.SigStatus)
  
- [nexus/exported/v1beta1/types.proto](#nexus/exported/v1beta1/types.proto)
    - [Asset](#nexus.exported.v1beta1.Asset)
    - [Chain](#nexus.exported.v1beta1.Chain)
    - [CrossChainAddress](#nexus.exported.v1beta1.CrossChainAddress)
    - [CrossChainTransfer](#nexus.exported.v1beta1.CrossChainTransfer)
    - [TransferFee](#nexus.exported.v1beta1.TransferFee)
  
    - [TransferState](#nexus.exported.v1beta1.TransferState)
  
- [axelarnet/v1beta1/tx.proto](#axelarnet/v1beta1/tx.proto)
    - [AddCosmosBasedChainRequest](#axelarnet.v1beta1.AddCosmosBasedChainRequest)
    - [AddCosmosBasedChainResponse](#axelarnet.v1beta1.AddCosmosBasedChainResponse)
    - [ConfirmDepositRequest](#axelarnet.v1beta1.ConfirmDepositRequest)
    - [ConfirmDepositResponse](#axelarnet.v1beta1.ConfirmDepositResponse)
    - [ExecutePendingTransfersRequest](#axelarnet.v1beta1.ExecutePendingTransfersRequest)
    - [ExecutePendingTransfersResponse](#axelarnet.v1beta1.ExecutePendingTransfersResponse)
    - [LinkRequest](#axelarnet.v1beta1.LinkRequest)
    - [LinkResponse](#axelarnet.v1beta1.LinkResponse)
    - [RegisterAssetRequest](#axelarnet.v1beta1.RegisterAssetRequest)
    - [RegisterAssetResponse](#axelarnet.v1beta1.RegisterAssetResponse)
    - [RegisterFeeCollectorRequest](#axelarnet.v1beta1.RegisterFeeCollectorRequest)
    - [RegisterFeeCollectorResponse](#axelarnet.v1beta1.RegisterFeeCollectorResponse)
    - [RegisterIBCPathRequest](#axelarnet.v1beta1.RegisterIBCPathRequest)
    - [RegisterIBCPathResponse](#axelarnet.v1beta1.RegisterIBCPathResponse)
    - [RouteIBCTransfersRequest](#axelarnet.v1beta1.RouteIBCTransfersRequest)
    - [RouteIBCTransfersResponse](#axelarnet.v1beta1.RouteIBCTransfersResponse)
  
- [axelarnet/v1beta1/service.proto](#axelarnet/v1beta1/service.proto)
    - [MsgService](#axelarnet.v1beta1.MsgService)
  
- [bitcoin/v1beta1/types.proto](#bitcoin/v1beta1/types.proto)
    - [AddressInfo](#bitcoin.v1beta1.AddressInfo)
    - [AddressInfo.SpendingCondition](#bitcoin.v1beta1.AddressInfo.SpendingCondition)
    - [Network](#bitcoin.v1beta1.Network)
    - [OutPointInfo](#bitcoin.v1beta1.OutPointInfo)
    - [SignedTx](#bitcoin.v1beta1.SignedTx)
    - [UnsignedTx](#bitcoin.v1beta1.UnsignedTx)
    - [UnsignedTx.Info](#bitcoin.v1beta1.UnsignedTx.Info)
    - [UnsignedTx.Info.InputInfo](#bitcoin.v1beta1.UnsignedTx.Info.InputInfo)
    - [UnsignedTx.Info.InputInfo.SigRequirement](#bitcoin.v1beta1.UnsignedTx.Info.InputInfo.SigRequirement)
  
    - [AddressRole](#bitcoin.v1beta1.AddressRole)
    - [OutPointState](#bitcoin.v1beta1.OutPointState)
    - [TxStatus](#bitcoin.v1beta1.TxStatus)
    - [TxType](#bitcoin.v1beta1.TxType)
  
- [bitcoin/v1beta1/params.proto](#bitcoin/v1beta1/params.proto)
    - [Params](#bitcoin.v1beta1.Params)
  
- [bitcoin/v1beta1/genesis.proto](#bitcoin/v1beta1/genesis.proto)
    - [GenesisState](#bitcoin.v1beta1.GenesisState)
  
- [bitcoin/v1beta1/query.proto](#bitcoin/v1beta1/query.proto)
    - [DepositQueryParams](#bitcoin.v1beta1.DepositQueryParams)
    - [QueryAddressResponse](#bitcoin.v1beta1.QueryAddressResponse)
    - [QueryDepositStatusResponse](#bitcoin.v1beta1.QueryDepositStatusResponse)
    - [QueryTxResponse](#bitcoin.v1beta1.QueryTxResponse)
    - [QueryTxResponse.SigningInfo](#bitcoin.v1beta1.QueryTxResponse.SigningInfo)
  
- [snapshot/exported/v1beta1/types.proto](#snapshot/exported/v1beta1/types.proto)
    - [Snapshot](#snapshot.exported.v1beta1.Snapshot)
    - [Validator](#snapshot.exported.v1beta1.Validator)
  
    - [ValidatorIllegibility](#snapshot.exported.v1beta1.ValidatorIllegibility)
  
- [vote/exported/v1beta1/types.proto](#vote/exported/v1beta1/types.proto)
    - [PollKey](#vote.exported.v1beta1.PollKey)
    - [PollMetadata](#vote.exported.v1beta1.PollMetadata)
    - [Voter](#vote.exported.v1beta1.Voter)
  
    - [PollState](#vote.exported.v1beta1.PollState)
  
- [bitcoin/v1beta1/tx.proto](#bitcoin/v1beta1/tx.proto)
    - [ConfirmOutpointRequest](#bitcoin.v1beta1.ConfirmOutpointRequest)
    - [ConfirmOutpointResponse](#bitcoin.v1beta1.ConfirmOutpointResponse)
    - [CreateMasterTxRequest](#bitcoin.v1beta1.CreateMasterTxRequest)
    - [CreateMasterTxResponse](#bitcoin.v1beta1.CreateMasterTxResponse)
    - [CreatePendingTransfersTxRequest](#bitcoin.v1beta1.CreatePendingTransfersTxRequest)
    - [CreatePendingTransfersTxResponse](#bitcoin.v1beta1.CreatePendingTransfersTxResponse)
    - [CreateRescueTxRequest](#bitcoin.v1beta1.CreateRescueTxRequest)
    - [CreateRescueTxResponse](#bitcoin.v1beta1.CreateRescueTxResponse)
    - [LinkRequest](#bitcoin.v1beta1.LinkRequest)
    - [LinkResponse](#bitcoin.v1beta1.LinkResponse)
    - [SignTxRequest](#bitcoin.v1beta1.SignTxRequest)
    - [SignTxResponse](#bitcoin.v1beta1.SignTxResponse)
    - [SubmitExternalSignatureRequest](#bitcoin.v1beta1.SubmitExternalSignatureRequest)
    - [SubmitExternalSignatureResponse](#bitcoin.v1beta1.SubmitExternalSignatureResponse)
    - [VoteConfirmOutpointRequest](#bitcoin.v1beta1.VoteConfirmOutpointRequest)
    - [VoteConfirmOutpointResponse](#bitcoin.v1beta1.VoteConfirmOutpointResponse)
  
- [bitcoin/v1beta1/service.proto](#bitcoin/v1beta1/service.proto)
    - [MsgService](#bitcoin.v1beta1.MsgService)
  
- [evm/v1beta1/types.proto](#evm/v1beta1/types.proto)
    - [Asset](#evm.v1beta1.Asset)
    - [BurnerInfo](#evm.v1beta1.BurnerInfo)
    - [Command](#evm.v1beta1.Command)
    - [CommandBatchMetadata](#evm.v1beta1.CommandBatchMetadata)
    - [ERC20Deposit](#evm.v1beta1.ERC20Deposit)
    - [ERC20TokenMetadata](#evm.v1beta1.ERC20TokenMetadata)
    - [Gateway](#evm.v1beta1.Gateway)
    - [NetworkInfo](#evm.v1beta1.NetworkInfo)
    - [SigMetadata](#evm.v1beta1.SigMetadata)
    - [TokenDetails](#evm.v1beta1.TokenDetails)
    - [TransactionMetadata](#evm.v1beta1.TransactionMetadata)
    - [TransferKey](#evm.v1beta1.TransferKey)
  
    - [BatchedCommandsStatus](#evm.v1beta1.BatchedCommandsStatus)
    - [DepositStatus](#evm.v1beta1.DepositStatus)
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
    - [GenesisState.Chain.CommandQueueEntry](#evm.v1beta1.GenesisState.Chain.CommandQueueEntry)
  
- [evm/v1beta1/query.proto](#evm/v1beta1/query.proto)
    - [BurnerInfoRequest](#evm.v1beta1.BurnerInfoRequest)
    - [BurnerInfoResponse](#evm.v1beta1.BurnerInfoResponse)
    - [ConfirmationHeightRequest](#evm.v1beta1.ConfirmationHeightRequest)
    - [ConfirmationHeightResponse](#evm.v1beta1.ConfirmationHeightResponse)
    - [DepositQueryParams](#evm.v1beta1.DepositQueryParams)
    - [DepositStateRequest](#evm.v1beta1.DepositStateRequest)
    - [DepositStateResponse](#evm.v1beta1.DepositStateResponse)
    - [QueryAddressResponse](#evm.v1beta1.QueryAddressResponse)
    - [QueryAddressResponse.MultisigAddresses](#evm.v1beta1.QueryAddressResponse.MultisigAddresses)
    - [QueryAddressResponse.ThresholdAddress](#evm.v1beta1.QueryAddressResponse.ThresholdAddress)
    - [QueryBatchedCommandsResponse](#evm.v1beta1.QueryBatchedCommandsResponse)
    - [QueryBurnerAddressResponse](#evm.v1beta1.QueryBurnerAddressResponse)
    - [QueryChainsResponse](#evm.v1beta1.QueryChainsResponse)
    - [QueryCommandResponse](#evm.v1beta1.QueryCommandResponse)
    - [QueryCommandResponse.ParamsEntry](#evm.v1beta1.QueryCommandResponse.ParamsEntry)
    - [QueryDepositStateParams](#evm.v1beta1.QueryDepositStateParams)
    - [QueryDepositStateResponse](#evm.v1beta1.QueryDepositStateResponse)
    - [QueryPendingCommandsResponse](#evm.v1beta1.QueryPendingCommandsResponse)
    - [QueryTokenAddressResponse](#evm.v1beta1.QueryTokenAddressResponse)
  
- [evm/v1beta1/tx.proto](#evm/v1beta1/tx.proto)
    - [AddChainRequest](#evm.v1beta1.AddChainRequest)
    - [AddChainResponse](#evm.v1beta1.AddChainResponse)
    - [ConfirmChainRequest](#evm.v1beta1.ConfirmChainRequest)
    - [ConfirmChainResponse](#evm.v1beta1.ConfirmChainResponse)
    - [ConfirmDepositRequest](#evm.v1beta1.ConfirmDepositRequest)
    - [ConfirmDepositResponse](#evm.v1beta1.ConfirmDepositResponse)
    - [ConfirmGatewayDeploymentRequest](#evm.v1beta1.ConfirmGatewayDeploymentRequest)
    - [ConfirmGatewayDeploymentResponse](#evm.v1beta1.ConfirmGatewayDeploymentResponse)
    - [ConfirmTokenRequest](#evm.v1beta1.ConfirmTokenRequest)
    - [ConfirmTokenResponse](#evm.v1beta1.ConfirmTokenResponse)
    - [ConfirmTransferKeyRequest](#evm.v1beta1.ConfirmTransferKeyRequest)
    - [ConfirmTransferKeyResponse](#evm.v1beta1.ConfirmTransferKeyResponse)
    - [CreateBurnTokensRequest](#evm.v1beta1.CreateBurnTokensRequest)
    - [CreateBurnTokensResponse](#evm.v1beta1.CreateBurnTokensResponse)
    - [CreateDeployTokenRequest](#evm.v1beta1.CreateDeployTokenRequest)
    - [CreateDeployTokenResponse](#evm.v1beta1.CreateDeployTokenResponse)
    - [CreatePendingTransfersRequest](#evm.v1beta1.CreatePendingTransfersRequest)
    - [CreatePendingTransfersResponse](#evm.v1beta1.CreatePendingTransfersResponse)
    - [CreateTransferOperatorshipRequest](#evm.v1beta1.CreateTransferOperatorshipRequest)
    - [CreateTransferOperatorshipResponse](#evm.v1beta1.CreateTransferOperatorshipResponse)
    - [CreateTransferOwnershipRequest](#evm.v1beta1.CreateTransferOwnershipRequest)
    - [CreateTransferOwnershipResponse](#evm.v1beta1.CreateTransferOwnershipResponse)
    - [LinkRequest](#evm.v1beta1.LinkRequest)
    - [LinkResponse](#evm.v1beta1.LinkResponse)
    - [SignCommandsRequest](#evm.v1beta1.SignCommandsRequest)
    - [SignCommandsResponse](#evm.v1beta1.SignCommandsResponse)
    - [VoteConfirmChainRequest](#evm.v1beta1.VoteConfirmChainRequest)
    - [VoteConfirmChainResponse](#evm.v1beta1.VoteConfirmChainResponse)
    - [VoteConfirmDepositRequest](#evm.v1beta1.VoteConfirmDepositRequest)
    - [VoteConfirmDepositResponse](#evm.v1beta1.VoteConfirmDepositResponse)
    - [VoteConfirmGatewayDeploymentRequest](#evm.v1beta1.VoteConfirmGatewayDeploymentRequest)
    - [VoteConfirmGatewayDeploymentResponse](#evm.v1beta1.VoteConfirmGatewayDeploymentResponse)
    - [VoteConfirmTokenRequest](#evm.v1beta1.VoteConfirmTokenRequest)
    - [VoteConfirmTokenResponse](#evm.v1beta1.VoteConfirmTokenResponse)
    - [VoteConfirmTransferKeyRequest](#evm.v1beta1.VoteConfirmTransferKeyRequest)
    - [VoteConfirmTransferKeyResponse](#evm.v1beta1.VoteConfirmTransferKeyResponse)
  
- [evm/v1beta1/service.proto](#evm/v1beta1/service.proto)
    - [MsgService](#evm.v1beta1.MsgService)
    - [QueryService](#evm.v1beta1.QueryService)
  
- [nexus/v1beta1/params.proto](#nexus/v1beta1/params.proto)
    - [Params](#nexus.v1beta1.Params)
  
- [nexus/v1beta1/types.proto](#nexus/v1beta1/types.proto)
    - [ChainState](#nexus.v1beta1.ChainState)
    - [LinkedAddresses](#nexus.v1beta1.LinkedAddresses)
  
- [nexus/v1beta1/genesis.proto](#nexus/v1beta1/genesis.proto)
    - [GenesisState](#nexus.v1beta1.GenesisState)
  
- [nexus/v1beta1/query.proto](#nexus/v1beta1/query.proto)
    - [LatestDepositAddressRequest](#nexus.v1beta1.LatestDepositAddressRequest)
    - [LatestDepositAddressResponse](#nexus.v1beta1.LatestDepositAddressResponse)
    - [QueryChainMaintainersResponse](#nexus.v1beta1.QueryChainMaintainersResponse)
    - [TransfersForChainRequest](#nexus.v1beta1.TransfersForChainRequest)
    - [TransfersForChainResponse](#nexus.v1beta1.TransfersForChainResponse)
  
- [nexus/v1beta1/tx.proto](#nexus/v1beta1/tx.proto)
    - [ActivateChainRequest](#nexus.v1beta1.ActivateChainRequest)
    - [ActivateChainResponse](#nexus.v1beta1.ActivateChainResponse)
    - [DeactivateChainRequest](#nexus.v1beta1.DeactivateChainRequest)
    - [DeactivateChainResponse](#nexus.v1beta1.DeactivateChainResponse)
    - [DeregisterChainMaintainerRequest](#nexus.v1beta1.DeregisterChainMaintainerRequest)
    - [DeregisterChainMaintainerResponse](#nexus.v1beta1.DeregisterChainMaintainerResponse)
    - [RegisterChainMaintainerRequest](#nexus.v1beta1.RegisterChainMaintainerRequest)
    - [RegisterChainMaintainerResponse](#nexus.v1beta1.RegisterChainMaintainerResponse)
  
- [nexus/v1beta1/service.proto](#nexus/v1beta1/service.proto)
    - [MsgService](#nexus.v1beta1.MsgService)
    - [QueryService](#nexus.v1beta1.QueryService)
  
- [permission/exported/v1beta1/types.proto](#permission/exported/v1beta1/types.proto)
    - [Role](#permission.exported.v1beta1.Role)
  
- [permission/v1beta1/types.proto](#permission/v1beta1/types.proto)
    - [GovAccount](#permission.v1beta1.GovAccount)
  
- [permission/v1beta1/params.proto](#permission/v1beta1/params.proto)
    - [Params](#permission.v1beta1.Params)
  
- [permission/v1beta1/genesis.proto](#permission/v1beta1/genesis.proto)
    - [GenesisState](#permission.v1beta1.GenesisState)
  
- [permission/v1beta1/query.proto](#permission/v1beta1/query.proto)
    - [QueryGovernanceKeyRequest](#permission.v1beta1.QueryGovernanceKeyRequest)
    - [QueryGovernanceKeyResponse](#permission.v1beta1.QueryGovernanceKeyResponse)
  
- [permission/v1beta1/tx.proto](#permission/v1beta1/tx.proto)
    - [DeregisterControllerRequest](#permission.v1beta1.DeregisterControllerRequest)
    - [DeregisterControllerResponse](#permission.v1beta1.DeregisterControllerResponse)
    - [RegisterControllerRequest](#permission.v1beta1.RegisterControllerRequest)
    - [RegisterControllerResponse](#permission.v1beta1.RegisterControllerResponse)
    - [UpdateGovernanceKeyRequest](#permission.v1beta1.UpdateGovernanceKeyRequest)
    - [UpdateGovernanceKeyResponse](#permission.v1beta1.UpdateGovernanceKeyResponse)
  
- [permission/v1beta1/service.proto](#permission/v1beta1/service.proto)
    - [Msg](#permission.v1beta1.Msg)
    - [Query](#permission.v1beta1.Query)
  
- [reward/v1beta1/params.proto](#reward/v1beta1/params.proto)
    - [Params](#reward.v1beta1.Params)
  
- [reward/v1beta1/types.proto](#reward/v1beta1/types.proto)
    - [Pool](#reward.v1beta1.Pool)
    - [Pool.Reward](#reward.v1beta1.Pool.Reward)
  
- [reward/v1beta1/genesis.proto](#reward/v1beta1/genesis.proto)
    - [GenesisState](#reward.v1beta1.GenesisState)
  
- [reward/v1beta1/tx.proto](#reward/v1beta1/tx.proto)
    - [RefundMsgRequest](#reward.v1beta1.RefundMsgRequest)
    - [RefundMsgResponse](#reward.v1beta1.RefundMsgResponse)
  
- [reward/v1beta1/service.proto](#reward/v1beta1/service.proto)
    - [MsgService](#reward.v1beta1.MsgService)
  
- [snapshot/v1beta1/params.proto](#snapshot/v1beta1/params.proto)
    - [Params](#snapshot.v1beta1.Params)
  
- [snapshot/v1beta1/types.proto](#snapshot/v1beta1/types.proto)
    - [ProxiedValidator](#snapshot.v1beta1.ProxiedValidator)
  
- [snapshot/v1beta1/genesis.proto](#snapshot/v1beta1/genesis.proto)
    - [GenesisState](#snapshot.v1beta1.GenesisState)
  
- [snapshot/v1beta1/query.proto](#snapshot/v1beta1/query.proto)
    - [QueryValidatorsResponse](#snapshot.v1beta1.QueryValidatorsResponse)
    - [QueryValidatorsResponse.TssIllegibilityInfo](#snapshot.v1beta1.QueryValidatorsResponse.TssIllegibilityInfo)
    - [QueryValidatorsResponse.Validator](#snapshot.v1beta1.QueryValidatorsResponse.Validator)
  
- [snapshot/v1beta1/tx.proto](#snapshot/v1beta1/tx.proto)
    - [DeactivateProxyRequest](#snapshot.v1beta1.DeactivateProxyRequest)
    - [DeactivateProxyResponse](#snapshot.v1beta1.DeactivateProxyResponse)
    - [RegisterProxyRequest](#snapshot.v1beta1.RegisterProxyRequest)
    - [RegisterProxyResponse](#snapshot.v1beta1.RegisterProxyResponse)
  
- [snapshot/v1beta1/service.proto](#snapshot/v1beta1/service.proto)
    - [MsgService](#snapshot.v1beta1.MsgService)
  
- [tss/tofnd/v1beta1/common.proto](#tss/tofnd/v1beta1/common.proto)
    - [KeyPresenceRequest](#tss.tofnd.v1beta1.KeyPresenceRequest)
    - [KeyPresenceResponse](#tss.tofnd.v1beta1.KeyPresenceResponse)
  
    - [KeyPresenceResponse.Response](#tss.tofnd.v1beta1.KeyPresenceResponse.Response)
  
- [tss/tofnd/v1beta1/multisig.proto](#tss/tofnd/v1beta1/multisig.proto)
    - [KeygenRequest](#tss.tofnd.v1beta1.KeygenRequest)
    - [KeygenResponse](#tss.tofnd.v1beta1.KeygenResponse)
    - [SignRequest](#tss.tofnd.v1beta1.SignRequest)
    - [SignResponse](#tss.tofnd.v1beta1.SignResponse)
  
- [tss/tofnd/v1beta1/tofnd.proto](#tss/tofnd/v1beta1/tofnd.proto)
    - [KeygenInit](#tss.tofnd.v1beta1.KeygenInit)
    - [KeygenOutput](#tss.tofnd.v1beta1.KeygenOutput)
    - [MessageIn](#tss.tofnd.v1beta1.MessageIn)
    - [MessageOut](#tss.tofnd.v1beta1.MessageOut)
    - [MessageOut.CriminalList](#tss.tofnd.v1beta1.MessageOut.CriminalList)
    - [MessageOut.CriminalList.Criminal](#tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal)
    - [MessageOut.KeygenResult](#tss.tofnd.v1beta1.MessageOut.KeygenResult)
    - [MessageOut.SignResult](#tss.tofnd.v1beta1.MessageOut.SignResult)
    - [RecoverRequest](#tss.tofnd.v1beta1.RecoverRequest)
    - [RecoverResponse](#tss.tofnd.v1beta1.RecoverResponse)
    - [SignInit](#tss.tofnd.v1beta1.SignInit)
    - [TrafficIn](#tss.tofnd.v1beta1.TrafficIn)
    - [TrafficOut](#tss.tofnd.v1beta1.TrafficOut)
  
    - [MessageOut.CriminalList.Criminal.CrimeType](#tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal.CrimeType)
    - [RecoverResponse.Response](#tss.tofnd.v1beta1.RecoverResponse.Response)
  
- [tss/v1beta1/params.proto](#tss/v1beta1/params.proto)
    - [Params](#tss.v1beta1.Params)
  
- [tss/v1beta1/types.proto](#tss/v1beta1/types.proto)
    - [ExternalKeys](#tss.v1beta1.ExternalKeys)
    - [KeyInfo](#tss.v1beta1.KeyInfo)
    - [KeyRecoveryInfo](#tss.v1beta1.KeyRecoveryInfo)
    - [KeyRecoveryInfo.PrivateEntry](#tss.v1beta1.KeyRecoveryInfo.PrivateEntry)
    - [KeygenVoteData](#tss.v1beta1.KeygenVoteData)
    - [MultisigInfo](#tss.v1beta1.MultisigInfo)
    - [MultisigInfo.Info](#tss.v1beta1.MultisigInfo.Info)
    - [ValidatorStatus](#tss.v1beta1.ValidatorStatus)
  
- [tss/v1beta1/genesis.proto](#tss/v1beta1/genesis.proto)
    - [GenesisState](#tss.v1beta1.GenesisState)
  
- [tss/v1beta1/query.proto](#tss/v1beta1/query.proto)
    - [QueryActiveOldKeysResponse](#tss.v1beta1.QueryActiveOldKeysResponse)
    - [QueryActiveOldKeysValidatorResponse](#tss.v1beta1.QueryActiveOldKeysValidatorResponse)
    - [QueryActiveOldKeysValidatorResponse.KeyInfo](#tss.v1beta1.QueryActiveOldKeysValidatorResponse.KeyInfo)
    - [QueryDeactivatedOperatorsResponse](#tss.v1beta1.QueryDeactivatedOperatorsResponse)
    - [QueryExternalKeyIDResponse](#tss.v1beta1.QueryExternalKeyIDResponse)
    - [QueryKeyResponse](#tss.v1beta1.QueryKeyResponse)
    - [QueryKeyResponse.ECDSAKey](#tss.v1beta1.QueryKeyResponse.ECDSAKey)
    - [QueryKeyResponse.Key](#tss.v1beta1.QueryKeyResponse.Key)
    - [QueryKeyResponse.MultisigKey](#tss.v1beta1.QueryKeyResponse.MultisigKey)
    - [QueryKeyShareResponse](#tss.v1beta1.QueryKeyShareResponse)
    - [QueryKeyShareResponse.ShareInfo](#tss.v1beta1.QueryKeyShareResponse.ShareInfo)
    - [QueryNextKeyIDRequest](#tss.v1beta1.QueryNextKeyIDRequest)
    - [QueryNextKeyIDResponse](#tss.v1beta1.QueryNextKeyIDResponse)
    - [QueryRecoveryResponse](#tss.v1beta1.QueryRecoveryResponse)
    - [QuerySignatureResponse](#tss.v1beta1.QuerySignatureResponse)
    - [QuerySignatureResponse.MultisigSignature](#tss.v1beta1.QuerySignatureResponse.MultisigSignature)
    - [QuerySignatureResponse.Signature](#tss.v1beta1.QuerySignatureResponse.Signature)
    - [QuerySignatureResponse.ThresholdSignature](#tss.v1beta1.QuerySignatureResponse.ThresholdSignature)
  
    - [VoteStatus](#tss.v1beta1.VoteStatus)
  
- [tss/v1beta1/tx.proto](#tss/v1beta1/tx.proto)
    - [HeartBeatRequest](#tss.v1beta1.HeartBeatRequest)
    - [HeartBeatResponse](#tss.v1beta1.HeartBeatResponse)
    - [ProcessKeygenTrafficRequest](#tss.v1beta1.ProcessKeygenTrafficRequest)
    - [ProcessKeygenTrafficResponse](#tss.v1beta1.ProcessKeygenTrafficResponse)
    - [ProcessSignTrafficRequest](#tss.v1beta1.ProcessSignTrafficRequest)
    - [ProcessSignTrafficResponse](#tss.v1beta1.ProcessSignTrafficResponse)
    - [RegisterExternalKeysRequest](#tss.v1beta1.RegisterExternalKeysRequest)
    - [RegisterExternalKeysRequest.ExternalKey](#tss.v1beta1.RegisterExternalKeysRequest.ExternalKey)
    - [RegisterExternalKeysResponse](#tss.v1beta1.RegisterExternalKeysResponse)
    - [RotateKeyRequest](#tss.v1beta1.RotateKeyRequest)
    - [RotateKeyResponse](#tss.v1beta1.RotateKeyResponse)
    - [StartKeygenRequest](#tss.v1beta1.StartKeygenRequest)
    - [StartKeygenResponse](#tss.v1beta1.StartKeygenResponse)
    - [SubmitMultisigPubKeysRequest](#tss.v1beta1.SubmitMultisigPubKeysRequest)
    - [SubmitMultisigPubKeysResponse](#tss.v1beta1.SubmitMultisigPubKeysResponse)
    - [SubmitMultisigSignaturesRequest](#tss.v1beta1.SubmitMultisigSignaturesRequest)
    - [SubmitMultisigSignaturesResponse](#tss.v1beta1.SubmitMultisigSignaturesResponse)
    - [VotePubKeyRequest](#tss.v1beta1.VotePubKeyRequest)
    - [VotePubKeyResponse](#tss.v1beta1.VotePubKeyResponse)
    - [VoteSigRequest](#tss.v1beta1.VoteSigRequest)
    - [VoteSigResponse](#tss.v1beta1.VoteSigResponse)
  
- [tss/v1beta1/service.proto](#tss/v1beta1/service.proto)
    - [MsgService](#tss.v1beta1.MsgService)
    - [Query](#tss.v1beta1.Query)
  
- [vote/v1beta1/params.proto](#vote/v1beta1/params.proto)
    - [Params](#vote.v1beta1.Params)
  
- [vote/v1beta1/genesis.proto](#vote/v1beta1/genesis.proto)
    - [GenesisState](#vote.v1beta1.GenesisState)
  
- [vote/v1beta1/types.proto](#vote/v1beta1/types.proto)
    - [TalliedVote](#vote.v1beta1.TalliedVote)
  
- [Scalar Value Types](#scalar-value-types)



<a name="axelarnet/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnet/v1beta1/params.proto



<a name="axelarnet.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `route_timeout_window` | [uint64](#uint64) |  | IBC packet route timeout window |
| `transaction_fee_rate` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnet/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnet/v1beta1/types.proto



<a name="axelarnet.v1beta1.Asset"></a>

### Asset



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `min_amount` | [bytes](#bytes) |  |  |






<a name="axelarnet.v1beta1.CosmosChain"></a>

### CosmosChain



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `ibc_path` | [string](#string) |  |  |
| `assets` | [Asset](#axelarnet.v1beta1.Asset) | repeated | **Deprecated.**  |
| `addr_prefix` | [string](#string) |  |  |






<a name="axelarnet.v1beta1.IBCTransfer"></a>

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



<a name="axelarnet/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnet/v1beta1/genesis.proto



<a name="axelarnet.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelarnet.v1beta1.Params) |  |  |
| `collector_address` | [bytes](#bytes) |  |  |
| `chains` | [CosmosChain](#axelarnet.v1beta1.CosmosChain) | repeated |  |
| `pending_transfers` | [IBCTransfer](#axelarnet.v1beta1.IBCTransfer) | repeated |  |





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



<a name="tss/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## tss/exported/v1beta1/types.proto



<a name="tss.exported.v1beta1.Key"></a>

### Key



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `role` | [KeyRole](#tss.exported.v1beta1.KeyRole) |  |  |
| `type` | [KeyType](#tss.exported.v1beta1.KeyType) |  |  |
| `ecdsa_key` | [Key.ECDSAKey](#tss.exported.v1beta1.Key.ECDSAKey) |  |  |
| `multisig_key` | [Key.MultisigKey](#tss.exported.v1beta1.Key.MultisigKey) |  |  |
| `rotated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `rotation_count` | [int64](#int64) |  |  |
| `chain` | [string](#string) |  |  |
| `snapshot_counter` | [int64](#int64) |  |  |






<a name="tss.exported.v1beta1.Key.ECDSAKey"></a>

### Key.ECDSAKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [bytes](#bytes) |  |  |






<a name="tss.exported.v1beta1.Key.MultisigKey"></a>

### Key.MultisigKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `values` | [bytes](#bytes) | repeated |  |
| `threshold` | [int64](#int64) |  |  |






<a name="tss.exported.v1beta1.KeyRequirement"></a>

### KeyRequirement
KeyRequirement defines requirements for keys


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_role` | [KeyRole](#tss.exported.v1beta1.KeyRole) |  |  |
| `key_type` | [KeyType](#tss.exported.v1beta1.KeyType) |  |  |
| `min_keygen_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  |  |
| `safety_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  |  |
| `key_share_distribution_policy` | [KeyShareDistributionPolicy](#tss.exported.v1beta1.KeyShareDistributionPolicy) |  |  |
| `max_total_share_count` | [int64](#int64) |  |  |
| `min_total_share_count` | [int64](#int64) |  |  |
| `keygen_voting_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  |  |
| `sign_voting_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  |  |
| `keygen_timeout` | [int64](#int64) |  |  |
| `sign_timeout` | [int64](#int64) |  |  |






<a name="tss.exported.v1beta1.SigKeyPair"></a>

### SigKeyPair
PubKeyInfo holds a pubkey and a signature


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pub_key` | [bytes](#bytes) |  |  |
| `signature` | [bytes](#bytes) |  |  |






<a name="tss.exported.v1beta1.SignInfo"></a>

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






<a name="tss.exported.v1beta1.Signature"></a>

### Signature
Signature holds public key and ECDSA signature


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sig_id` | [string](#string) |  |  |
| `single_sig` | [Signature.SingleSig](#tss.exported.v1beta1.Signature.SingleSig) |  |  |
| `multi_sig` | [Signature.MultiSig](#tss.exported.v1beta1.Signature.MultiSig) |  |  |
| `sig_status` | [SigStatus](#tss.exported.v1beta1.SigStatus) |  |  |






<a name="tss.exported.v1beta1.Signature.MultiSig"></a>

### Signature.MultiSig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sig_key_pairs` | [SigKeyPair](#tss.exported.v1beta1.SigKeyPair) | repeated |  |






<a name="tss.exported.v1beta1.Signature.SingleSig"></a>

### Signature.SingleSig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sig_key_pair` | [SigKeyPair](#tss.exported.v1beta1.SigKeyPair) |  |  |





 <!-- end messages -->


<a name="tss.exported.v1beta1.AckType"></a>

### AckType


| Name | Number | Description |
| ---- | ------ | ----------- |
| ACK_TYPE_UNSPECIFIED | 0 |  |
| ACK_TYPE_KEYGEN | 1 |  |
| ACK_TYPE_SIGN | 2 |  |



<a name="tss.exported.v1beta1.KeyRole"></a>

### KeyRole


| Name | Number | Description |
| ---- | ------ | ----------- |
| KEY_ROLE_UNSPECIFIED | 0 |  |
| KEY_ROLE_MASTER_KEY | 1 |  |
| KEY_ROLE_SECONDARY_KEY | 2 |  |
| KEY_ROLE_EXTERNAL_KEY | 3 |  |



<a name="tss.exported.v1beta1.KeyShareDistributionPolicy"></a>

### KeyShareDistributionPolicy


| Name | Number | Description |
| ---- | ------ | ----------- |
| KEY_SHARE_DISTRIBUTION_POLICY_UNSPECIFIED | 0 |  |
| KEY_SHARE_DISTRIBUTION_POLICY_WEIGHTED_BY_STAKE | 1 |  |
| KEY_SHARE_DISTRIBUTION_POLICY_ONE_PER_VALIDATOR | 2 |  |



<a name="tss.exported.v1beta1.KeyType"></a>

### KeyType


| Name | Number | Description |
| ---- | ------ | ----------- |
| KEY_TYPE_UNSPECIFIED | 0 |  |
| KEY_TYPE_NONE | 1 |  |
| KEY_TYPE_THRESHOLD | 2 |  |
| KEY_TYPE_MULTISIG | 3 |  |



<a name="tss.exported.v1beta1.SigStatus"></a>

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



<a name="nexus/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## nexus/exported/v1beta1/types.proto



<a name="nexus.exported.v1beta1.Asset"></a>

### Asset



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `min_amount` | [bytes](#bytes) |  |  |
| `is_native_asset` | [bool](#bool) |  |  |






<a name="nexus.exported.v1beta1.Chain"></a>

### Chain
Chain represents the properties of a registered blockchain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `supports_foreign_assets` | [bool](#bool) |  |  |
| `key_type` | [tss.exported.v1beta1.KeyType](#tss.exported.v1beta1.KeyType) |  |  |
| `module` | [string](#string) |  |  |






<a name="nexus.exported.v1beta1.CrossChainAddress"></a>

### CrossChainAddress
CrossChainAddress represents a generalized address on any registered chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [Chain](#nexus.exported.v1beta1.Chain) |  |  |
| `address` | [string](#string) |  |  |






<a name="nexus.exported.v1beta1.CrossChainTransfer"></a>

### CrossChainTransfer
CrossChainTransfer represents a generalized transfer of some asset to a
registered blockchain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `recipient` | [CrossChainAddress](#nexus.exported.v1beta1.CrossChainAddress) |  |  |
| `asset` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `id` | [uint64](#uint64) |  |  |
| `state` | [TransferState](#nexus.exported.v1beta1.TransferState) |  |  |






<a name="nexus.exported.v1beta1.TransferFee"></a>

### TransferFee
TransferFee represents accumulated fees generated by the network


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |





 <!-- end messages -->


<a name="nexus.exported.v1beta1.TransferState"></a>

### TransferState


| Name | Number | Description |
| ---- | ------ | ----------- |
| TRANSFER_STATE_UNSPECIFIED | 0 |  |
| TRANSFER_STATE_PENDING | 1 |  |
| TRANSFER_STATE_ARCHIVED | 2 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnet/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnet/v1beta1/tx.proto



<a name="axelarnet.v1beta1.AddCosmosBasedChainRequest"></a>

### AddCosmosBasedChainRequest
MsgAddCosmosBasedChain represents a message to register a cosmos based chain
to nexus


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [nexus.exported.v1beta1.Chain](#nexus.exported.v1beta1.Chain) |  |  |
| `addr_prefix` | [string](#string) |  |  |
| `native_assets` | [nexus.exported.v1beta1.Asset](#nexus.exported.v1beta1.Asset) | repeated |  |






<a name="axelarnet.v1beta1.AddCosmosBasedChainResponse"></a>

### AddCosmosBasedChainResponse







<a name="axelarnet.v1beta1.ConfirmDepositRequest"></a>

### ConfirmDepositRequest
MsgConfirmDeposit represents a deposit confirmation message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `token` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | **Deprecated.**  |
| `deposit_address` | [bytes](#bytes) |  |  |
| `denom` | [string](#string) |  |  |






<a name="axelarnet.v1beta1.ConfirmDepositResponse"></a>

### ConfirmDepositResponse







<a name="axelarnet.v1beta1.ExecutePendingTransfersRequest"></a>

### ExecutePendingTransfersRequest
MsgExecutePendingTransfers represents a message to trigger transfer all
pending transfers


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |






<a name="axelarnet.v1beta1.ExecutePendingTransfersResponse"></a>

### ExecutePendingTransfersResponse







<a name="axelarnet.v1beta1.LinkRequest"></a>

### LinkRequest
MsgLink represents a message to link a cross-chain address to an Axelar
address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `recipient_addr` | [string](#string) |  |  |
| `recipient_chain` | [string](#string) |  |  |
| `asset` | [string](#string) |  |  |






<a name="axelarnet.v1beta1.LinkResponse"></a>

### LinkResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposit_addr` | [string](#string) |  |  |






<a name="axelarnet.v1beta1.RegisterAssetRequest"></a>

### RegisterAssetRequest
RegisterAssetRequest represents a message to register an asset to a cosmos
based chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `asset` | [nexus.exported.v1beta1.Asset](#nexus.exported.v1beta1.Asset) |  |  |
| `is_native_asset` | [bool](#bool) |  |  |






<a name="axelarnet.v1beta1.RegisterAssetResponse"></a>

### RegisterAssetResponse







<a name="axelarnet.v1beta1.RegisterFeeCollectorRequest"></a>

### RegisterFeeCollectorRequest
RegisterFeeCollectorRequest represents a message to register axelarnet fee
collector account


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `fee_collector` | [bytes](#bytes) |  |  |






<a name="axelarnet.v1beta1.RegisterFeeCollectorResponse"></a>

### RegisterFeeCollectorResponse







<a name="axelarnet.v1beta1.RegisterIBCPathRequest"></a>

### RegisterIBCPathRequest
MSgRegisterIBCPath represents a message to register an IBC tracing path for
a cosmos chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `path` | [string](#string) |  |  |






<a name="axelarnet.v1beta1.RegisterIBCPathResponse"></a>

### RegisterIBCPathResponse







<a name="axelarnet.v1beta1.RouteIBCTransfersRequest"></a>

### RouteIBCTransfersRequest
RouteIBCTransfersRequest represents a message to route pending transfers to
cosmos based chains


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |






<a name="axelarnet.v1beta1.RouteIBCTransfersResponse"></a>

### RouteIBCTransfersResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnet/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnet/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelarnet.v1beta1.MsgService"></a>

### MsgService
Msg defines the axelarnet Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Link` | [LinkRequest](#axelarnet.v1beta1.LinkRequest) | [LinkResponse](#axelarnet.v1beta1.LinkResponse) |  | POST|/axelar/axelarnet/link/{recipient_chain}|
| `ConfirmDeposit` | [ConfirmDepositRequest](#axelarnet.v1beta1.ConfirmDepositRequest) | [ConfirmDepositResponse](#axelarnet.v1beta1.ConfirmDepositResponse) |  | POST|/axelar/axelarnet/confirm-deposit|
| `ExecutePendingTransfers` | [ExecutePendingTransfersRequest](#axelarnet.v1beta1.ExecutePendingTransfersRequest) | [ExecutePendingTransfersResponse](#axelarnet.v1beta1.ExecutePendingTransfersResponse) |  | POST|/axelar/axelarnet/execute-pending-transfers|
| `RegisterIBCPath` | [RegisterIBCPathRequest](#axelarnet.v1beta1.RegisterIBCPathRequest) | [RegisterIBCPathResponse](#axelarnet.v1beta1.RegisterIBCPathResponse) |  | POST|/axelar/axelarnet/register-ibc-path|
| `AddCosmosBasedChain` | [AddCosmosBasedChainRequest](#axelarnet.v1beta1.AddCosmosBasedChainRequest) | [AddCosmosBasedChainResponse](#axelarnet.v1beta1.AddCosmosBasedChainResponse) |  | POST|/axelar/axelarnet/add-cosmos-based-chain|
| `RegisterAsset` | [RegisterAssetRequest](#axelarnet.v1beta1.RegisterAssetRequest) | [RegisterAssetResponse](#axelarnet.v1beta1.RegisterAssetResponse) |  | POST|/axelar/axelarnet/register-asset|
| `RouteIBCTransfers` | [RouteIBCTransfersRequest](#axelarnet.v1beta1.RouteIBCTransfersRequest) | [RouteIBCTransfersResponse](#axelarnet.v1beta1.RouteIBCTransfersResponse) |  | POST|/axelar/axelarnet/route-ibc-transfers|
| `RegisterFeeCollector` | [RegisterFeeCollectorRequest](#axelarnet.v1beta1.RegisterFeeCollectorRequest) | [RegisterFeeCollectorResponse](#axelarnet.v1beta1.RegisterFeeCollectorResponse) |  | POST|/axelar/axelarnet/register-fee-collector|

 <!-- end services -->



<a name="bitcoin/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## bitcoin/v1beta1/types.proto



<a name="bitcoin.v1beta1.AddressInfo"></a>

### AddressInfo
AddressInfo is a wrapper containing the Bitcoin P2WSH address, it's
corresponding script and the underlying key


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `role` | [AddressRole](#bitcoin.v1beta1.AddressRole) |  |  |
| `redeem_script` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `max_sig_count` | [uint32](#uint32) |  |  |
| `spending_condition` | [AddressInfo.SpendingCondition](#bitcoin.v1beta1.AddressInfo.SpendingCondition) |  |  |






<a name="bitcoin.v1beta1.AddressInfo.SpendingCondition"></a>

### AddressInfo.SpendingCondition



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `internal_key_ids` | [string](#string) | repeated | internal_key_ids lists the internal key IDs that one of which has to sign regardless of locktime |
| `external_key_ids` | [string](#string) | repeated | external_key_ids lists the external key IDs that external_multisig_threshold of which have to sign to spend before locktime if set |
| `external_multisig_threshold` | [int64](#int64) |  |  |
| `lock_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="bitcoin.v1beta1.Network"></a>

### Network



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |






<a name="bitcoin.v1beta1.OutPointInfo"></a>

### OutPointInfo
OutPointInfo describes all the necessary information to confirm the outPoint
of a transaction


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `out_point` | [string](#string) |  |  |
| `amount` | [int64](#int64) |  |  |
| `address` | [string](#string) |  |  |






<a name="bitcoin.v1beta1.SignedTx"></a>

### SignedTx



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `type` | [TxType](#bitcoin.v1beta1.TxType) |  |  |
| `tx` | [bytes](#bytes) |  |  |
| `prev_signed_tx_hash` | [bytes](#bytes) |  |  |
| `confirmation_required` | [bool](#bool) |  |  |
| `anyone_can_spend_vout` | [uint32](#uint32) |  |  |






<a name="bitcoin.v1beta1.UnsignedTx"></a>

### UnsignedTx



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `type` | [TxType](#bitcoin.v1beta1.TxType) |  |  |
| `tx` | [bytes](#bytes) |  |  |
| `info` | [UnsignedTx.Info](#bitcoin.v1beta1.UnsignedTx.Info) |  |  |
| `status` | [TxStatus](#bitcoin.v1beta1.TxStatus) |  |  |
| `confirmation_required` | [bool](#bool) |  |  |
| `anyone_can_spend_vout` | [uint32](#uint32) |  |  |
| `prev_aborted_key_id` | [string](#string) |  |  |
| `internal_transfer_amount` | [int64](#int64) |  |  |






<a name="bitcoin.v1beta1.UnsignedTx.Info"></a>

### UnsignedTx.Info



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `rotate_key` | [bool](#bool) |  |  |
| `input_infos` | [UnsignedTx.Info.InputInfo](#bitcoin.v1beta1.UnsignedTx.Info.InputInfo) | repeated |  |






<a name="bitcoin.v1beta1.UnsignedTx.Info.InputInfo"></a>

### UnsignedTx.Info.InputInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sig_requirements` | [UnsignedTx.Info.InputInfo.SigRequirement](#bitcoin.v1beta1.UnsignedTx.Info.InputInfo.SigRequirement) | repeated |  |






<a name="bitcoin.v1beta1.UnsignedTx.Info.InputInfo.SigRequirement"></a>

### UnsignedTx.Info.InputInfo.SigRequirement



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `sig_hash` | [bytes](#bytes) |  |  |





 <!-- end messages -->


<a name="bitcoin.v1beta1.AddressRole"></a>

### AddressRole


| Name | Number | Description |
| ---- | ------ | ----------- |
| ADDRESS_ROLE_UNSPECIFIED | 0 |  |
| ADDRESS_ROLE_DEPOSIT | 1 |  |
| ADDRESS_ROLE_CONSOLIDATION | 2 |  |



<a name="bitcoin.v1beta1.OutPointState"></a>

### OutPointState


| Name | Number | Description |
| ---- | ------ | ----------- |
| OUT_POINT_STATE_UNSPECIFIED | 0 |  |
| OUT_POINT_STATE_PENDING | 1 |  |
| OUT_POINT_STATE_CONFIRMED | 2 |  |
| OUT_POINT_STATE_SPENT | 3 |  |



<a name="bitcoin.v1beta1.TxStatus"></a>

### TxStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| TX_STATUS_UNSPECIFIED | 0 |  |
| TX_STATUS_CREATED | 1 |  |
| TX_STATUS_SIGNING | 2 |  |
| TX_STATUS_ABORTED | 3 |  |
| TX_STATUS_SIGNED | 4 |  |



<a name="bitcoin.v1beta1.TxType"></a>

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



<a name="bitcoin/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## bitcoin/v1beta1/params.proto



<a name="bitcoin.v1beta1.Params"></a>

### Params



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `network` | [Network](#bitcoin.v1beta1.Network) |  |  |
| `confirmation_height` | [uint64](#uint64) |  |  |
| `revote_locking_period` | [int64](#int64) |  |  |
| `sig_check_interval` | [int64](#int64) |  |  |
| `min_output_amount` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  |  |
| `max_input_count` | [int64](#int64) |  |  |
| `max_secondary_output_amount` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  |  |
| `master_key_retention_period` | [int64](#int64) |  |  |
| `master_address_internal_key_lock_duration` | [int64](#int64) |  |  |
| `master_address_external_key_lock_duration` | [int64](#int64) |  |  |
| `voting_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  |  |
| `min_voter_count` | [int64](#int64) |  |  |
| `max_tx_size` | [int64](#int64) |  |  |
| `transaction_fee_rate` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="bitcoin/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## bitcoin/v1beta1/genesis.proto



<a name="bitcoin.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#bitcoin.v1beta1.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="bitcoin/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## bitcoin/v1beta1/query.proto



<a name="bitcoin.v1beta1.DepositQueryParams"></a>

### DepositQueryParams
DepositQueryParams describe the parameters used to query for a Bitcoin
deposit address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `chain` | [string](#string) |  |  |






<a name="bitcoin.v1beta1.QueryAddressResponse"></a>

### QueryAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="bitcoin.v1beta1.QueryDepositStatusResponse"></a>

### QueryDepositStatusResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |
| `status` | [OutPointState](#bitcoin.v1beta1.OutPointState) |  |  |






<a name="bitcoin.v1beta1.QueryTxResponse"></a>

### QueryTxResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx` | [string](#string) |  |  |
| `status` | [TxStatus](#bitcoin.v1beta1.TxStatus) |  |  |
| `confirmation_required` | [bool](#bool) |  |  |
| `prev_signed_tx_hash` | [string](#string) |  |  |
| `anyone_can_spend_vout` | [uint32](#uint32) |  |  |
| `signing_infos` | [QueryTxResponse.SigningInfo](#bitcoin.v1beta1.QueryTxResponse.SigningInfo) | repeated |  |






<a name="bitcoin.v1beta1.QueryTxResponse.SigningInfo"></a>

### QueryTxResponse.SigningInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `redeem_script` | [string](#string) |  |  |
| `amount` | [int64](#int64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="snapshot/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## snapshot/exported/v1beta1/types.proto



<a name="snapshot.exported.v1beta1.Snapshot"></a>

### Snapshot



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validators` | [Validator](#snapshot.exported.v1beta1.Validator) | repeated |  |
| `timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `height` | [int64](#int64) |  |  |
| `total_share_count` | [bytes](#bytes) |  |  |
| `counter` | [int64](#int64) |  |  |
| `key_share_distribution_policy` | [tss.exported.v1beta1.KeyShareDistributionPolicy](#tss.exported.v1beta1.KeyShareDistributionPolicy) |  |  |
| `corruption_threshold` | [int64](#int64) |  |  |






<a name="snapshot.exported.v1beta1.Validator"></a>

### Validator



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sdk_validator` | [google.protobuf.Any](#google.protobuf.Any) |  |  |
| `share_count` | [int64](#int64) |  |  |





 <!-- end messages -->


<a name="snapshot.exported.v1beta1.ValidatorIllegibility"></a>

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



<a name="vote/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## vote/exported/v1beta1/types.proto



<a name="vote.exported.v1beta1.PollKey"></a>

### PollKey
PollKey represents the key data for a poll


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |






<a name="vote.exported.v1beta1.PollMetadata"></a>

### PollMetadata
PollMetadata represents a poll with write-in voting, i.e. the result of the
vote can have any data type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [PollKey](#vote.exported.v1beta1.PollKey) |  |  |
| `expires_at` | [int64](#int64) |  |  |
| `result` | [google.protobuf.Any](#google.protobuf.Any) |  |  |
| `voting_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  |  |
| `state` | [PollState](#vote.exported.v1beta1.PollState) |  |  |
| `min_voter_count` | [int64](#int64) |  |  |
| `voters` | [Voter](#vote.exported.v1beta1.Voter) | repeated |  |
| `total_voting_power` | [bytes](#bytes) |  |  |
| `reward_pool_name` | [string](#string) |  |  |






<a name="vote.exported.v1beta1.Voter"></a>

### Voter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validator` | [bytes](#bytes) |  |  |
| `voting_power` | [int64](#int64) |  |  |





 <!-- end messages -->


<a name="vote.exported.v1beta1.PollState"></a>

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



<a name="bitcoin/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## bitcoin/v1beta1/tx.proto



<a name="bitcoin.v1beta1.ConfirmOutpointRequest"></a>

### ConfirmOutpointRequest
MsgConfirmOutpoint represents a message to trigger the confirmation of a
Bitcoin outpoint


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `out_point_info` | [OutPointInfo](#bitcoin.v1beta1.OutPointInfo) |  |  |






<a name="bitcoin.v1beta1.ConfirmOutpointResponse"></a>

### ConfirmOutpointResponse







<a name="bitcoin.v1beta1.CreateMasterTxRequest"></a>

### CreateMasterTxRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `secondary_key_amount` | [int64](#int64) |  |  |






<a name="bitcoin.v1beta1.CreateMasterTxResponse"></a>

### CreateMasterTxResponse







<a name="bitcoin.v1beta1.CreatePendingTransfersTxRequest"></a>

### CreatePendingTransfersTxRequest
CreatePendingTransfersTxRequest represents a message to trigger the creation
of a secondary key consolidation transaction


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `master_key_amount` | [int64](#int64) |  |  |






<a name="bitcoin.v1beta1.CreatePendingTransfersTxResponse"></a>

### CreatePendingTransfersTxResponse







<a name="bitcoin.v1beta1.CreateRescueTxRequest"></a>

### CreateRescueTxRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |






<a name="bitcoin.v1beta1.CreateRescueTxResponse"></a>

### CreateRescueTxResponse







<a name="bitcoin.v1beta1.LinkRequest"></a>

### LinkRequest
MsgLink represents a message to link a cross-chain address to a Bitcoin
address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `recipient_addr` | [string](#string) |  |  |
| `recipient_chain` | [string](#string) |  |  |






<a name="bitcoin.v1beta1.LinkResponse"></a>

### LinkResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposit_addr` | [string](#string) |  |  |






<a name="bitcoin.v1beta1.SignTxRequest"></a>

### SignTxRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `tx_type` | [TxType](#bitcoin.v1beta1.TxType) |  |  |






<a name="bitcoin.v1beta1.SignTxResponse"></a>

### SignTxResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `position` | [int64](#int64) |  |  |






<a name="bitcoin.v1beta1.SubmitExternalSignatureRequest"></a>

### SubmitExternalSignatureRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `signature` | [bytes](#bytes) |  |  |
| `sig_hash` | [bytes](#bytes) |  |  |






<a name="bitcoin.v1beta1.SubmitExternalSignatureResponse"></a>

### SubmitExternalSignatureResponse







<a name="bitcoin.v1beta1.VoteConfirmOutpointRequest"></a>

### VoteConfirmOutpointRequest
MsgVoteConfirmOutpoint represents a message to that votes on an outpoint


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll_key` | [vote.exported.v1beta1.PollKey](#vote.exported.v1beta1.PollKey) |  |  |
| `out_point` | [string](#string) |  |  |
| `confirmed` | [bool](#bool) |  |  |






<a name="bitcoin.v1beta1.VoteConfirmOutpointResponse"></a>

### VoteConfirmOutpointResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `status` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="bitcoin/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## bitcoin/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="bitcoin.v1beta1.MsgService"></a>

### MsgService
Msg defines the bitcoin Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Link` | [LinkRequest](#bitcoin.v1beta1.LinkRequest) | [LinkResponse](#bitcoin.v1beta1.LinkResponse) |  | POST|/axelar/bitcoin/link/{recipient_chain}|
| `ConfirmOutpoint` | [ConfirmOutpointRequest](#bitcoin.v1beta1.ConfirmOutpointRequest) | [ConfirmOutpointResponse](#bitcoin.v1beta1.ConfirmOutpointResponse) |  | POST|/axelar/bitcoin/confirm|
| `VoteConfirmOutpoint` | [VoteConfirmOutpointRequest](#bitcoin.v1beta1.VoteConfirmOutpointRequest) | [VoteConfirmOutpointResponse](#bitcoin.v1beta1.VoteConfirmOutpointResponse) |  | ||
| `CreatePendingTransfersTx` | [CreatePendingTransfersTxRequest](#bitcoin.v1beta1.CreatePendingTransfersTxRequest) | [CreatePendingTransfersTxResponse](#bitcoin.v1beta1.CreatePendingTransfersTxResponse) |  | POST|/axelar/bitcoin/create-pending-transfers-tx|
| `CreateMasterTx` | [CreateMasterTxRequest](#bitcoin.v1beta1.CreateMasterTxRequest) | [CreateMasterTxResponse](#bitcoin.v1beta1.CreateMasterTxResponse) |  | POST|/axelar/bitcoin/create-master-tx|
| `CreateRescueTx` | [CreateRescueTxRequest](#bitcoin.v1beta1.CreateRescueTxRequest) | [CreateRescueTxResponse](#bitcoin.v1beta1.CreateRescueTxResponse) |  | POST|/axelar/bitcoin/create-rescue-tx|
| `SignTx` | [SignTxRequest](#bitcoin.v1beta1.SignTxRequest) | [SignTxResponse](#bitcoin.v1beta1.SignTxResponse) |  | POST|/axelar/bitcoin/sign-tx|
| `SubmitExternalSignature` | [SubmitExternalSignatureRequest](#bitcoin.v1beta1.SubmitExternalSignatureRequest) | [SubmitExternalSignatureResponse](#bitcoin.v1beta1.SubmitExternalSignatureResponse) |  | POST|/axelar/bitcoin/submit-external-signature|

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
| `min_amount` | [bytes](#bytes) |  | **Deprecated.**  |
| `status` | [Status](#evm.v1beta1.Status) |  |  |
| `is_external` | [bool](#bool) |  |  |
| `burner_code` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.Gateway"></a>

### Gateway



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [bytes](#bytes) |  |  |
| `status` | [Gateway.Status](#evm.v1beta1.Gateway.Status) |  |  |






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
| `gateway_code` | [bytes](#bytes) |  |  |
| `token_code` | [bytes](#bytes) |  |  |
| `burnable` | [bytes](#bytes) |  |  |
| `revote_locking_period` | [int64](#int64) |  |  |
| `networks` | [NetworkInfo](#evm.v1beta1.NetworkInfo) | repeated |  |
| `voting_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  |  |
| `min_voter_count` | [int64](#int64) |  |  |
| `commands_gas_limit` | [uint32](#uint32) |  |  |
| `transaction_fee_rate` | [string](#string) |  |  |






<a name="evm.v1beta1.PendingChain"></a>

### PendingChain



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#evm.v1beta1.Params) |  |  |
| `chain` | [nexus.exported.v1beta1.Chain](#nexus.exported.v1beta1.Chain) |  |  |





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
| `command_queue` | [GenesisState.Chain.CommandQueueEntry](#evm.v1beta1.GenesisState.Chain.CommandQueueEntry) | repeated |  |
| `confirmed_deposits` | [ERC20Deposit](#evm.v1beta1.ERC20Deposit) | repeated |  |
| `burned_deposits` | [ERC20Deposit](#evm.v1beta1.ERC20Deposit) | repeated |  |
| `command_batches` | [CommandBatchMetadata](#evm.v1beta1.CommandBatchMetadata) | repeated |  |
| `gateway` | [Gateway](#evm.v1beta1.Gateway) |  |  |
| `tokens` | [ERC20TokenMetadata](#evm.v1beta1.ERC20TokenMetadata) | repeated |  |






<a name="evm.v1beta1.GenesisState.Chain.CommandQueueEntry"></a>

### GenesisState.Chain.CommandQueueEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [Command](#evm.v1beta1.Command) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="evm/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evm/v1beta1/query.proto



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






<a name="evm.v1beta1.QueryAddressResponse"></a>

### QueryAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `multisig_addresses` | [QueryAddressResponse.MultisigAddresses](#evm.v1beta1.QueryAddressResponse.MultisigAddresses) |  |  |
| `threshold_address` | [QueryAddressResponse.ThresholdAddress](#evm.v1beta1.QueryAddressResponse.ThresholdAddress) |  |  |






<a name="evm.v1beta1.QueryAddressResponse.MultisigAddresses"></a>

### QueryAddressResponse.MultisigAddresses



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `addresses` | [string](#string) | repeated |  |
| `threshold` | [uint32](#uint32) |  |  |






<a name="evm.v1beta1.QueryAddressResponse.ThresholdAddress"></a>

### QueryAddressResponse.ThresholdAddress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="evm.v1beta1.QueryBatchedCommandsResponse"></a>

### QueryBatchedCommandsResponse



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






<a name="evm.v1beta1.QueryBurnerAddressResponse"></a>

### QueryBurnerAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="evm.v1beta1.QueryChainsResponse"></a>

### QueryChainsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chains` | [string](#string) | repeated |  |






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






<a name="evm.v1beta1.QueryDepositStateResponse"></a>

### QueryDepositStateResponse
QueryDepositStateResponse is used by the legacy querier


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |
| `status` | [DepositStatus](#evm.v1beta1.DepositStatus) |  |  |






<a name="evm.v1beta1.QueryPendingCommandsResponse"></a>

### QueryPendingCommandsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `commands` | [QueryCommandResponse](#evm.v1beta1.QueryCommandResponse) | repeated |  |






<a name="evm.v1beta1.QueryTokenAddressResponse"></a>

### QueryTokenAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="evm/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evm/v1beta1/tx.proto



<a name="evm.v1beta1.AddChainRequest"></a>

### AddChainRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `name` | [string](#string) |  |  |
| `key_type` | [tss.exported.v1beta1.KeyType](#tss.exported.v1beta1.KeyType) |  |  |
| `params` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.AddChainResponse"></a>

### AddChainResponse







<a name="evm.v1beta1.ConfirmChainRequest"></a>

### ConfirmChainRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `name` | [string](#string) |  |  |






<a name="evm.v1beta1.ConfirmChainResponse"></a>

### ConfirmChainResponse







<a name="evm.v1beta1.ConfirmDepositRequest"></a>

### ConfirmDepositRequest
MsgConfirmDeposit represents an erc20 deposit confirmation message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `amount` | [bytes](#bytes) |  |  |
| `burner_address` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.ConfirmDepositResponse"></a>

### ConfirmDepositResponse







<a name="evm.v1beta1.ConfirmGatewayDeploymentRequest"></a>

### ConfirmGatewayDeploymentRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `address` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.ConfirmGatewayDeploymentResponse"></a>

### ConfirmGatewayDeploymentResponse







<a name="evm.v1beta1.ConfirmTokenRequest"></a>

### ConfirmTokenRequest
MsgConfirmToken represents a token deploy confirmation message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `asset` | [Asset](#evm.v1beta1.Asset) |  |  |






<a name="evm.v1beta1.ConfirmTokenResponse"></a>

### ConfirmTokenResponse







<a name="evm.v1beta1.ConfirmTransferKeyRequest"></a>

### ConfirmTransferKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `transfer_type` | [TransferKeyType](#evm.v1beta1.TransferKeyType) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="evm.v1beta1.ConfirmTransferKeyResponse"></a>

### ConfirmTransferKeyResponse







<a name="evm.v1beta1.CreateBurnTokensRequest"></a>

### CreateBurnTokensRequest
CreateBurnTokensRequest represents the message to create commands to burn
tokens with AxelarGateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |






<a name="evm.v1beta1.CreateBurnTokensResponse"></a>

### CreateBurnTokensResponse







<a name="evm.v1beta1.CreateDeployTokenRequest"></a>

### CreateDeployTokenRequest
CreateDeployTokenRequest represents the message to create a deploy token
command for AxelarGateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `asset` | [Asset](#evm.v1beta1.Asset) |  |  |
| `token_details` | [TokenDetails](#evm.v1beta1.TokenDetails) |  |  |
| `min_amount` | [bytes](#bytes) |  |  |
| `address` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.CreateDeployTokenResponse"></a>

### CreateDeployTokenResponse







<a name="evm.v1beta1.CreatePendingTransfersRequest"></a>

### CreatePendingTransfersRequest
CreatePendingTransfersRequest represents a message to trigger the creation of
commands handling all pending transfers


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |






<a name="evm.v1beta1.CreatePendingTransfersResponse"></a>

### CreatePendingTransfersResponse







<a name="evm.v1beta1.CreateTransferOperatorshipRequest"></a>

### CreateTransferOperatorshipRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="evm.v1beta1.CreateTransferOperatorshipResponse"></a>

### CreateTransferOperatorshipResponse







<a name="evm.v1beta1.CreateTransferOwnershipRequest"></a>

### CreateTransferOwnershipRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="evm.v1beta1.CreateTransferOwnershipResponse"></a>

### CreateTransferOwnershipResponse







<a name="evm.v1beta1.LinkRequest"></a>

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






<a name="evm.v1beta1.LinkResponse"></a>

### LinkResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposit_addr` | [string](#string) |  |  |






<a name="evm.v1beta1.SignCommandsRequest"></a>

### SignCommandsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |






<a name="evm.v1beta1.SignCommandsResponse"></a>

### SignCommandsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `batched_commands_id` | [bytes](#bytes) |  |  |
| `command_count` | [uint32](#uint32) |  |  |






<a name="evm.v1beta1.VoteConfirmChainRequest"></a>

### VoteConfirmChainRequest
MsgVoteConfirmChain represents a message that votes on a new EVM chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `name` | [string](#string) |  |  |
| `poll_key` | [vote.exported.v1beta1.PollKey](#vote.exported.v1beta1.PollKey) |  |  |
| `confirmed` | [bool](#bool) |  |  |






<a name="evm.v1beta1.VoteConfirmChainResponse"></a>

### VoteConfirmChainResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |






<a name="evm.v1beta1.VoteConfirmDepositRequest"></a>

### VoteConfirmDepositRequest
MsgVoteConfirmDeposit represents a message that votes on a deposit


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `poll_key` | [vote.exported.v1beta1.PollKey](#vote.exported.v1beta1.PollKey) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `burn_address` | [bytes](#bytes) |  |  |
| `confirmed` | [bool](#bool) |  |  |






<a name="evm.v1beta1.VoteConfirmDepositResponse"></a>

### VoteConfirmDepositResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |






<a name="evm.v1beta1.VoteConfirmGatewayDeploymentRequest"></a>

### VoteConfirmGatewayDeploymentRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll_key` | [vote.exported.v1beta1.PollKey](#vote.exported.v1beta1.PollKey) |  |  |
| `chain` | [string](#string) |  |  |
| `confirmed` | [bool](#bool) |  |  |






<a name="evm.v1beta1.VoteConfirmGatewayDeploymentResponse"></a>

### VoteConfirmGatewayDeploymentResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |






<a name="evm.v1beta1.VoteConfirmTokenRequest"></a>

### VoteConfirmTokenRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `poll_key` | [vote.exported.v1beta1.PollKey](#vote.exported.v1beta1.PollKey) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `asset` | [string](#string) |  |  |
| `confirmed` | [bool](#bool) |  |  |






<a name="evm.v1beta1.VoteConfirmTokenResponse"></a>

### VoteConfirmTokenResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |






<a name="evm.v1beta1.VoteConfirmTransferKeyRequest"></a>

### VoteConfirmTransferKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `poll_key` | [vote.exported.v1beta1.PollKey](#vote.exported.v1beta1.PollKey) |  |  |
| `confirmed` | [bool](#bool) |  |  |






<a name="evm.v1beta1.VoteConfirmTransferKeyResponse"></a>

### VoteConfirmTransferKeyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="evm/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evm/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="evm.v1beta1.MsgService"></a>

### MsgService
Msg defines the evm Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Link` | [LinkRequest](#evm.v1beta1.LinkRequest) | [LinkResponse](#evm.v1beta1.LinkResponse) |  | POST|/axelar/evm/link/{recipient_chain}|
| `ConfirmChain` | [ConfirmChainRequest](#evm.v1beta1.ConfirmChainRequest) | [ConfirmChainResponse](#evm.v1beta1.ConfirmChainResponse) |  | POST|/axelar/evm/confirm-chain|
| `ConfirmGatewayDeployment` | [ConfirmGatewayDeploymentRequest](#evm.v1beta1.ConfirmGatewayDeploymentRequest) | [ConfirmGatewayDeploymentResponse](#evm.v1beta1.ConfirmGatewayDeploymentResponse) |  | POST|/axelar/evm/confirm-gateway-deployment|
| `ConfirmToken` | [ConfirmTokenRequest](#evm.v1beta1.ConfirmTokenRequest) | [ConfirmTokenResponse](#evm.v1beta1.ConfirmTokenResponse) |  | POST|/axelar/evm/confirm-erc20-deploy|
| `ConfirmDeposit` | [ConfirmDepositRequest](#evm.v1beta1.ConfirmDepositRequest) | [ConfirmDepositResponse](#evm.v1beta1.ConfirmDepositResponse) |  | POST|/axelar/evm/confirm-erc20-deposit|
| `ConfirmTransferKey` | [ConfirmTransferKeyRequest](#evm.v1beta1.ConfirmTransferKeyRequest) | [ConfirmTransferKeyResponse](#evm.v1beta1.ConfirmTransferKeyResponse) |  | POST|/axelar/evm/confirm-transfer-ownership|
| `VoteConfirmChain` | [VoteConfirmChainRequest](#evm.v1beta1.VoteConfirmChainRequest) | [VoteConfirmChainResponse](#evm.v1beta1.VoteConfirmChainResponse) |  | POST|/axelar/evm/vote-confirm-chain|
| `VoteConfirmGatewayDeployment` | [VoteConfirmGatewayDeploymentRequest](#evm.v1beta1.VoteConfirmGatewayDeploymentRequest) | [VoteConfirmGatewayDeploymentResponse](#evm.v1beta1.VoteConfirmGatewayDeploymentResponse) |  | POST|/axelar/evm/vote-confirm-gateway-deployment|
| `VoteConfirmDeposit` | [VoteConfirmDepositRequest](#evm.v1beta1.VoteConfirmDepositRequest) | [VoteConfirmDepositResponse](#evm.v1beta1.VoteConfirmDepositResponse) |  | POST|/axelar/evm/vote-confirm-deposit|
| `VoteConfirmToken` | [VoteConfirmTokenRequest](#evm.v1beta1.VoteConfirmTokenRequest) | [VoteConfirmTokenResponse](#evm.v1beta1.VoteConfirmTokenResponse) |  | POST|/axelar/evm/vote-confirm-token|
| `VoteConfirmTransferKey` | [VoteConfirmTransferKeyRequest](#evm.v1beta1.VoteConfirmTransferKeyRequest) | [VoteConfirmTransferKeyResponse](#evm.v1beta1.VoteConfirmTransferKeyResponse) |  | POST|/axelar/evm/vote-confirm-transfer-key|
| `CreateDeployToken` | [CreateDeployTokenRequest](#evm.v1beta1.CreateDeployTokenRequest) | [CreateDeployTokenResponse](#evm.v1beta1.CreateDeployTokenResponse) |  | POST|/axelar/evm/create-deploy-token|
| `CreateBurnTokens` | [CreateBurnTokensRequest](#evm.v1beta1.CreateBurnTokensRequest) | [CreateBurnTokensResponse](#evm.v1beta1.CreateBurnTokensResponse) |  | POST|/axelar/evm/sign-burn|
| `CreatePendingTransfers` | [CreatePendingTransfersRequest](#evm.v1beta1.CreatePendingTransfersRequest) | [CreatePendingTransfersResponse](#evm.v1beta1.CreatePendingTransfersResponse) |  | POST|/axelar/evm/create-pending-transfers|
| `CreateTransferOwnership` | [CreateTransferOwnershipRequest](#evm.v1beta1.CreateTransferOwnershipRequest) | [CreateTransferOwnershipResponse](#evm.v1beta1.CreateTransferOwnershipResponse) |  | POST|/axelar/evm/create-transfer-ownership|
| `CreateTransferOperatorship` | [CreateTransferOperatorshipRequest](#evm.v1beta1.CreateTransferOperatorshipRequest) | [CreateTransferOperatorshipResponse](#evm.v1beta1.CreateTransferOperatorshipResponse) |  | POST|/axelar/evm/create-transfer-operatorship|
| `SignCommands` | [SignCommandsRequest](#evm.v1beta1.SignCommandsRequest) | [SignCommandsResponse](#evm.v1beta1.SignCommandsResponse) |  | POST|/axelar/evm/sign-commands|
| `AddChain` | [AddChainRequest](#evm.v1beta1.AddChainRequest) | [AddChainResponse](#evm.v1beta1.AddChainResponse) |  | POST|/axelar/evm/add-chain|


<a name="evm.v1beta1.QueryService"></a>

### QueryService
QueryService defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `BurnerInfo` | [BurnerInfoRequest](#evm.v1beta1.BurnerInfoRequest) | [BurnerInfoResponse](#evm.v1beta1.BurnerInfoResponse) | BurnerInfo queries the burner info for the specified address | GET|/evm/v1beta1/burner_info|
| `ConfirmationHeight` | [ConfirmationHeightRequest](#evm.v1beta1.ConfirmationHeightRequest) | [ConfirmationHeightResponse](#evm.v1beta1.ConfirmationHeightResponse) | ConfirmationHeight queries the confirmation height for the specified chain | GET|/evm/v1beta1/confirmation_height|
| `DepositState` | [DepositStateRequest](#evm.v1beta1.DepositStateRequest) | [DepositStateResponse](#evm.v1beta1.DepositStateResponse) | DepositState queries the state of the specified deposit | GET|/evm/v1beta1/deposit_state|

 <!-- end services -->



<a name="nexus/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## nexus/v1beta1/params.proto



<a name="nexus.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain_activation_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="nexus/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## nexus/v1beta1/types.proto



<a name="nexus.v1beta1.ChainState"></a>

### ChainState
ChainState represents the state of a registered blockchain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [nexus.exported.v1beta1.Chain](#nexus.exported.v1beta1.Chain) |  |  |
| `maintainers` | [bytes](#bytes) | repeated |  |
| `activated` | [bool](#bool) |  |  |
| `assets` | [nexus.exported.v1beta1.Asset](#nexus.exported.v1beta1.Asset) | repeated |  |
| `native_assets` | [string](#string) | repeated |  |






<a name="nexus.v1beta1.LinkedAddresses"></a>

### LinkedAddresses



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposit_address` | [nexus.exported.v1beta1.CrossChainAddress](#nexus.exported.v1beta1.CrossChainAddress) |  |  |
| `recipient_address` | [nexus.exported.v1beta1.CrossChainAddress](#nexus.exported.v1beta1.CrossChainAddress) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="nexus/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## nexus/v1beta1/genesis.proto



<a name="nexus.v1beta1.GenesisState"></a>

### GenesisState
GenesisState represents the genesis state


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#nexus.v1beta1.Params) |  |  |
| `nonce` | [uint64](#uint64) |  |  |
| `chains` | [nexus.exported.v1beta1.Chain](#nexus.exported.v1beta1.Chain) | repeated |  |
| `chain_states` | [ChainState](#nexus.v1beta1.ChainState) | repeated |  |
| `linked_addresses` | [LinkedAddresses](#nexus.v1beta1.LinkedAddresses) | repeated |  |
| `transfers` | [nexus.exported.v1beta1.CrossChainTransfer](#nexus.exported.v1beta1.CrossChainTransfer) | repeated |  |
| `fee` | [nexus.exported.v1beta1.TransferFee](#nexus.exported.v1beta1.TransferFee) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="nexus/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## nexus/v1beta1/query.proto



<a name="nexus.v1beta1.LatestDepositAddressRequest"></a>

### LatestDepositAddressRequest
LatestDepositAddressRequest represents a message that queries a deposit
address by recipient address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `recipient_addr` | [string](#string) |  |  |
| `recipient_chain` | [string](#string) |  |  |
| `deposit_chain` | [string](#string) |  |  |






<a name="nexus.v1beta1.LatestDepositAddressResponse"></a>

### LatestDepositAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposit_addr` | [string](#string) |  |  |






<a name="nexus.v1beta1.QueryChainMaintainersResponse"></a>

### QueryChainMaintainersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `maintainers` | [bytes](#bytes) | repeated |  |






<a name="nexus.v1beta1.TransfersForChainRequest"></a>

### TransfersForChainRequest
TransfersForChainRequest represents a message that queries the
transfers for the specified chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `state` | [nexus.exported.v1beta1.TransferState](#nexus.exported.v1beta1.TransferState) |  |  |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |






<a name="nexus.v1beta1.TransfersForChainResponse"></a>

### TransfersForChainResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `transfers` | [nexus.exported.v1beta1.CrossChainTransfer](#nexus.exported.v1beta1.CrossChainTransfer) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="nexus/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## nexus/v1beta1/tx.proto



<a name="nexus.v1beta1.ActivateChainRequest"></a>

### ActivateChainRequest
ActivateChainRequest represents a message to activate chains


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chains` | [string](#string) | repeated |  |






<a name="nexus.v1beta1.ActivateChainResponse"></a>

### ActivateChainResponse







<a name="nexus.v1beta1.DeactivateChainRequest"></a>

### DeactivateChainRequest
DeactivateChainRequest represents a message to deactivate chains


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chains` | [string](#string) | repeated |  |






<a name="nexus.v1beta1.DeactivateChainResponse"></a>

### DeactivateChainResponse







<a name="nexus.v1beta1.DeregisterChainMaintainerRequest"></a>

### DeregisterChainMaintainerRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chains` | [string](#string) | repeated |  |






<a name="nexus.v1beta1.DeregisterChainMaintainerResponse"></a>

### DeregisterChainMaintainerResponse







<a name="nexus.v1beta1.RegisterChainMaintainerRequest"></a>

### RegisterChainMaintainerRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chains` | [string](#string) | repeated |  |






<a name="nexus.v1beta1.RegisterChainMaintainerResponse"></a>

### RegisterChainMaintainerResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="nexus/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## nexus/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="nexus.v1beta1.MsgService"></a>

### MsgService
Msg defines the nexus Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterChainMaintainer` | [RegisterChainMaintainerRequest](#nexus.v1beta1.RegisterChainMaintainerRequest) | [RegisterChainMaintainerResponse](#nexus.v1beta1.RegisterChainMaintainerResponse) |  | POST|/axelar/nexus/registerChainMaintainer|
| `DeregisterChainMaintainer` | [DeregisterChainMaintainerRequest](#nexus.v1beta1.DeregisterChainMaintainerRequest) | [DeregisterChainMaintainerResponse](#nexus.v1beta1.DeregisterChainMaintainerResponse) |  | POST|/axelar/nexus/deregisterChainMaintainer|
| `ActivateChain` | [ActivateChainRequest](#nexus.v1beta1.ActivateChainRequest) | [ActivateChainResponse](#nexus.v1beta1.ActivateChainResponse) |  | POST|/axelar/nexus/registerChainMaintainer|
| `DeactivateChain` | [DeactivateChainRequest](#nexus.v1beta1.DeactivateChainRequest) | [DeactivateChainResponse](#nexus.v1beta1.DeactivateChainResponse) |  | ||


<a name="nexus.v1beta1.QueryService"></a>

### QueryService
QueryService defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `LatestDepositAddress` | [LatestDepositAddressRequest](#nexus.v1beta1.LatestDepositAddressRequest) | [LatestDepositAddressResponse](#nexus.v1beta1.LatestDepositAddressResponse) | LatestDepositAddress queries the a deposit address by recipient | GET|/nexus/v1beta1/latest_deposit_address/{recipient_chain}/{recipient_addr}|
| `TransfersForChain` | [TransfersForChainRequest](#nexus.v1beta1.TransfersForChainRequest) | [TransfersForChainResponse](#nexus.v1beta1.TransfersForChainResponse) | TransfersForChain queries transfers by chain | GET|/nexus/v1beta1/transfers_for_chain|

 <!-- end services -->



<a name="permission/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## permission/exported/v1beta1/types.proto


 <!-- end messages -->


<a name="permission.exported.v1beta1.Role"></a>

### Role


| Name | Number | Description |
| ---- | ------ | ----------- |
| ROLE_UNSPECIFIED | 0 |  |
| ROLE_ACCESS_CONTROL | 1 |  |
| ROLE_CHAIN_MANAGEMENT | 2 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="permission/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## permission/v1beta1/types.proto



<a name="permission.v1beta1.GovAccount"></a>

### GovAccount



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [bytes](#bytes) |  |  |
| `role` | [permission.exported.v1beta1.Role](#permission.exported.v1beta1.Role) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="permission/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## permission/v1beta1/params.proto



<a name="permission.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="permission/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## permission/v1beta1/genesis.proto



<a name="permission.v1beta1.GenesisState"></a>

### GenesisState
GenesisState represents the genesis state


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#permission.v1beta1.Params) |  |  |
| `governance_key` | [cosmos.crypto.multisig.LegacyAminoPubKey](#cosmos.crypto.multisig.LegacyAminoPubKey) |  |  |
| `gov_accounts` | [GovAccount](#permission.v1beta1.GovAccount) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="permission/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## permission/v1beta1/query.proto



<a name="permission.v1beta1.QueryGovernanceKeyRequest"></a>

### QueryGovernanceKeyRequest
QueryGovernanceKeyRequest is the request type for the
Query/GovernanceKey RPC method






<a name="permission.v1beta1.QueryGovernanceKeyResponse"></a>

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



<a name="permission/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## permission/v1beta1/tx.proto



<a name="permission.v1beta1.DeregisterControllerRequest"></a>

### DeregisterControllerRequest
DeregisterController represents a message to deregister a controller account


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `controller` | [bytes](#bytes) |  |  |






<a name="permission.v1beta1.DeregisterControllerResponse"></a>

### DeregisterControllerResponse







<a name="permission.v1beta1.RegisterControllerRequest"></a>

### RegisterControllerRequest
MsgRegisterController represents a message to register a controller account


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `controller` | [bytes](#bytes) |  |  |






<a name="permission.v1beta1.RegisterControllerResponse"></a>

### RegisterControllerResponse







<a name="permission.v1beta1.UpdateGovernanceKeyRequest"></a>

### UpdateGovernanceKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `governance_key` | [cosmos.crypto.multisig.LegacyAminoPubKey](#cosmos.crypto.multisig.LegacyAminoPubKey) |  |  |






<a name="permission.v1beta1.UpdateGovernanceKeyResponse"></a>

### UpdateGovernanceKeyResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="permission/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## permission/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="permission.v1beta1.Msg"></a>

### Msg
Msg defines the gov Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterController` | [RegisterControllerRequest](#permission.v1beta1.RegisterControllerRequest) | [RegisterControllerResponse](#permission.v1beta1.RegisterControllerResponse) |  | ||
| `DeregisterController` | [DeregisterControllerRequest](#permission.v1beta1.DeregisterControllerRequest) | [DeregisterControllerResponse](#permission.v1beta1.DeregisterControllerResponse) |  | ||
| `UpdateGovernanceKey` | [UpdateGovernanceKeyRequest](#permission.v1beta1.UpdateGovernanceKeyRequest) | [UpdateGovernanceKeyResponse](#permission.v1beta1.UpdateGovernanceKeyResponse) |  | ||


<a name="permission.v1beta1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `GovernanceKey` | [QueryGovernanceKeyRequest](#permission.v1beta1.QueryGovernanceKeyRequest) | [QueryGovernanceKeyResponse](#permission.v1beta1.QueryGovernanceKeyResponse) | GovernanceKey returns multisig governance key | GET|/permission/v1beta1/governance_key|

 <!-- end services -->



<a name="reward/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## reward/v1beta1/params.proto



<a name="reward.v1beta1.Params"></a>

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



<a name="reward/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## reward/v1beta1/types.proto



<a name="reward.v1beta1.Pool"></a>

### Pool



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `rewards` | [Pool.Reward](#reward.v1beta1.Pool.Reward) | repeated |  |






<a name="reward.v1beta1.Pool.Reward"></a>

### Pool.Reward



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validator` | [bytes](#bytes) |  |  |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="reward/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## reward/v1beta1/genesis.proto



<a name="reward.v1beta1.GenesisState"></a>

### GenesisState
GenesisState represents the genesis state


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#reward.v1beta1.Params) |  |  |
| `pools` | [Pool](#reward.v1beta1.Pool) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="reward/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## reward/v1beta1/tx.proto



<a name="reward.v1beta1.RefundMsgRequest"></a>

### RefundMsgRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `inner_message` | [google.protobuf.Any](#google.protobuf.Any) |  |  |






<a name="reward.v1beta1.RefundMsgResponse"></a>

### RefundMsgResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  |  |
| `log` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="reward/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## reward/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="reward.v1beta1.MsgService"></a>

### MsgService
Msg defines the axelarnet Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RefundMsg` | [RefundMsgRequest](#reward.v1beta1.RefundMsgRequest) | [RefundMsgResponse](#reward.v1beta1.RefundMsgResponse) |  | POST|/reward/refund-message|

 <!-- end services -->



<a name="snapshot/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## snapshot/v1beta1/params.proto



<a name="snapshot.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `min_proxy_balance` | [int64](#int64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="snapshot/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## snapshot/v1beta1/types.proto



<a name="snapshot.v1beta1.ProxiedValidator"></a>

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



<a name="snapshot/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## snapshot/v1beta1/genesis.proto



<a name="snapshot.v1beta1.GenesisState"></a>

### GenesisState
GenesisState represents the genesis state


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#snapshot.v1beta1.Params) |  |  |
| `snapshots` | [snapshot.exported.v1beta1.Snapshot](#snapshot.exported.v1beta1.Snapshot) | repeated |  |
| `proxied_validators` | [ProxiedValidator](#snapshot.v1beta1.ProxiedValidator) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="snapshot/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## snapshot/v1beta1/query.proto



<a name="snapshot.v1beta1.QueryValidatorsResponse"></a>

### QueryValidatorsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validators` | [QueryValidatorsResponse.Validator](#snapshot.v1beta1.QueryValidatorsResponse.Validator) | repeated |  |






<a name="snapshot.v1beta1.QueryValidatorsResponse.TssIllegibilityInfo"></a>

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






<a name="snapshot.v1beta1.QueryValidatorsResponse.Validator"></a>

### QueryValidatorsResponse.Validator



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `operator_address` | [string](#string) |  |  |
| `moniker` | [string](#string) |  |  |
| `tss_illegibility_info` | [QueryValidatorsResponse.TssIllegibilityInfo](#snapshot.v1beta1.QueryValidatorsResponse.TssIllegibilityInfo) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="snapshot/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## snapshot/v1beta1/tx.proto



<a name="snapshot.v1beta1.DeactivateProxyRequest"></a>

### DeactivateProxyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |






<a name="snapshot.v1beta1.DeactivateProxyResponse"></a>

### DeactivateProxyResponse







<a name="snapshot.v1beta1.RegisterProxyRequest"></a>

### RegisterProxyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `proxy_addr` | [bytes](#bytes) |  |  |






<a name="snapshot.v1beta1.RegisterProxyResponse"></a>

### RegisterProxyResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="snapshot/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## snapshot/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="snapshot.v1beta1.MsgService"></a>

### MsgService
Msg defines the snapshot Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterProxy` | [RegisterProxyRequest](#snapshot.v1beta1.RegisterProxyRequest) | [RegisterProxyResponse](#snapshot.v1beta1.RegisterProxyResponse) | RegisterProxy defines a method for registering a proxy account that can act in a validator account's stead. | POST|/axelar/snapshot/registerProxy/{proxy_addr}|
| `DeactivateProxy` | [DeactivateProxyRequest](#snapshot.v1beta1.DeactivateProxyRequest) | [DeactivateProxyResponse](#snapshot.v1beta1.DeactivateProxyResponse) | DeactivateProxy defines a method for deregistering a proxy account. | POST|/axelar/snapshot/deactivateProxy|

 <!-- end services -->



<a name="tss/tofnd/v1beta1/common.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## tss/tofnd/v1beta1/common.proto
File copied from golang tofnd with minor tweaks


<a name="tss.tofnd.v1beta1.KeyPresenceRequest"></a>

### KeyPresenceRequest
Key presence check types


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_uid` | [string](#string) |  |  |






<a name="tss.tofnd.v1beta1.KeyPresenceResponse"></a>

### KeyPresenceResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `response` | [KeyPresenceResponse.Response](#tss.tofnd.v1beta1.KeyPresenceResponse.Response) |  |  |





 <!-- end messages -->


<a name="tss.tofnd.v1beta1.KeyPresenceResponse.Response"></a>

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



<a name="tss/tofnd/v1beta1/multisig.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## tss/tofnd/v1beta1/multisig.proto
File copied from golang tofnd with minor tweaks


<a name="tss.tofnd.v1beta1.KeygenRequest"></a>

### KeygenRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_uid` | [string](#string) |  |  |
| `party_uid` | [string](#string) |  | used only for logging |






<a name="tss.tofnd.v1beta1.KeygenResponse"></a>

### KeygenResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pub_key` | [bytes](#bytes) |  | SEC1-encoded compressed curve point |
| `error` | [string](#string) |  | reply with an error message if keygen fails |






<a name="tss.tofnd.v1beta1.SignRequest"></a>

### SignRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_uid` | [string](#string) |  |  |
| `msg_to_sign` | [bytes](#bytes) |  | 32-byte pre-hashed message digest |
| `party_uid` | [string](#string) |  | used only for logging |






<a name="tss.tofnd.v1beta1.SignResponse"></a>

### SignResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signature` | [bytes](#bytes) |  | ASN.1 DER-encoded ECDSA signature |
| `error` | [string](#string) |  | reply with an error message if sign fails |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="tss/tofnd/v1beta1/tofnd.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## tss/tofnd/v1beta1/tofnd.proto
File copied from golang tofnd with minor tweaks


<a name="tss.tofnd.v1beta1.KeygenInit"></a>

### KeygenInit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `new_key_uid` | [string](#string) |  |  |
| `party_uids` | [string](#string) | repeated |  |
| `party_share_counts` | [uint32](#uint32) | repeated |  |
| `my_party_index` | [uint32](#uint32) |  | parties[my_party_index] belongs to the server |
| `threshold` | [uint32](#uint32) |  |  |






<a name="tss.tofnd.v1beta1.KeygenOutput"></a>

### KeygenOutput
Keygen's success response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pub_key` | [bytes](#bytes) |  | pub_key; common for all parties |
| `group_recover_info` | [bytes](#bytes) |  | recover info of all parties' shares; common for all parties |
| `private_recover_info` | [bytes](#bytes) |  | private recover info of this party's shares; unique for each party |






<a name="tss.tofnd.v1beta1.MessageIn"></a>

### MessageIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `keygen_init` | [KeygenInit](#tss.tofnd.v1beta1.KeygenInit) |  | first message only, Keygen |
| `sign_init` | [SignInit](#tss.tofnd.v1beta1.SignInit) |  | first message only, Sign |
| `traffic` | [TrafficIn](#tss.tofnd.v1beta1.TrafficIn) |  | all subsequent messages |
| `abort` | [bool](#bool) |  | abort the protocol, ignore the bool value |






<a name="tss.tofnd.v1beta1.MessageOut"></a>

### MessageOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `traffic` | [TrafficOut](#tss.tofnd.v1beta1.TrafficOut) |  | all but final message |
| `keygen_result` | [MessageOut.KeygenResult](#tss.tofnd.v1beta1.MessageOut.KeygenResult) |  | final message only, Keygen |
| `sign_result` | [MessageOut.SignResult](#tss.tofnd.v1beta1.MessageOut.SignResult) |  | final message only, Sign |
| `need_recover` | [bool](#bool) |  | issue recover from client |






<a name="tss.tofnd.v1beta1.MessageOut.CriminalList"></a>

### MessageOut.CriminalList
Keygen/Sign failure response message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `criminals` | [MessageOut.CriminalList.Criminal](#tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal) | repeated |  |






<a name="tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal"></a>

### MessageOut.CriminalList.Criminal



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `party_uid` | [string](#string) |  |  |
| `crime_type` | [MessageOut.CriminalList.Criminal.CrimeType](#tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal.CrimeType) |  |  |






<a name="tss.tofnd.v1beta1.MessageOut.KeygenResult"></a>

### MessageOut.KeygenResult
Keygen's response types


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [KeygenOutput](#tss.tofnd.v1beta1.KeygenOutput) |  | Success response |
| `criminals` | [MessageOut.CriminalList](#tss.tofnd.v1beta1.MessageOut.CriminalList) |  | Faiilure response |






<a name="tss.tofnd.v1beta1.MessageOut.SignResult"></a>

### MessageOut.SignResult
Sign's response types


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signature` | [bytes](#bytes) |  | Success response |
| `criminals` | [MessageOut.CriminalList](#tss.tofnd.v1beta1.MessageOut.CriminalList) |  | Failure response |






<a name="tss.tofnd.v1beta1.RecoverRequest"></a>

### RecoverRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `keygen_init` | [KeygenInit](#tss.tofnd.v1beta1.KeygenInit) |  |  |
| `keygen_output` | [KeygenOutput](#tss.tofnd.v1beta1.KeygenOutput) |  |  |






<a name="tss.tofnd.v1beta1.RecoverResponse"></a>

### RecoverResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `response` | [RecoverResponse.Response](#tss.tofnd.v1beta1.RecoverResponse.Response) |  |  |






<a name="tss.tofnd.v1beta1.SignInit"></a>

### SignInit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `new_sig_uid` | [string](#string) |  |  |
| `key_uid` | [string](#string) |  |  |
| `party_uids` | [string](#string) | repeated | TODO replace this with a subset of indices? |
| `message_to_sign` | [bytes](#bytes) |  |  |






<a name="tss.tofnd.v1beta1.TrafficIn"></a>

### TrafficIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `from_party_uid` | [string](#string) |  |  |
| `payload` | [bytes](#bytes) |  |  |
| `is_broadcast` | [bool](#bool) |  |  |






<a name="tss.tofnd.v1beta1.TrafficOut"></a>

### TrafficOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to_party_uid` | [string](#string) |  |  |
| `payload` | [bytes](#bytes) |  |  |
| `is_broadcast` | [bool](#bool) |  |  |





 <!-- end messages -->


<a name="tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal.CrimeType"></a>

### MessageOut.CriminalList.Criminal.CrimeType


| Name | Number | Description |
| ---- | ------ | ----------- |
| CRIME_TYPE_UNSPECIFIED | 0 |  |
| CRIME_TYPE_NON_MALICIOUS | 1 |  |
| CRIME_TYPE_MALICIOUS | 2 |  |



<a name="tss.tofnd.v1beta1.RecoverResponse.Response"></a>

### RecoverResponse.Response


| Name | Number | Description |
| ---- | ------ | ----------- |
| RESPONSE_UNSPECIFIED | 0 |  |
| RESPONSE_SUCCESS | 1 |  |
| RESPONSE_FAIL | 2 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="tss/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## tss/v1beta1/params.proto



<a name="tss.v1beta1.Params"></a>

### Params
Params is the parameter set for this module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_requirements` | [tss.exported.v1beta1.KeyRequirement](#tss.exported.v1beta1.KeyRequirement) | repeated | KeyRequirements defines the requirement for each key role |
| `suspend_duration_in_blocks` | [int64](#int64) |  | SuspendDurationInBlocks defines the number of blocks a validator is disallowed to participate in any TSS ceremony after committing a malicious behaviour during signing |
| `heartbeat_period_in_blocks` | [int64](#int64) |  | HeartBeatPeriodInBlocks defines the time period in blocks for tss to emit the event asking validators to send their heartbeats |
| `max_missed_blocks_per_window` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  |  |
| `unbonding_locking_key_rotation_count` | [int64](#int64) |  |  |
| `external_multisig_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  |  |
| `max_sign_queue_size` | [int64](#int64) |  |  |
| `max_simultaneous_sign_shares` | [int64](#int64) |  |  |
| `tss_signed_blocks_window` | [int64](#int64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="tss/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## tss/v1beta1/types.proto



<a name="tss.v1beta1.ExternalKeys"></a>

### ExternalKeys



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `key_ids` | [string](#string) | repeated |  |






<a name="tss.v1beta1.KeyInfo"></a>

### KeyInfo
KeyInfo holds information about a key


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `key_role` | [tss.exported.v1beta1.KeyRole](#tss.exported.v1beta1.KeyRole) |  |  |
| `key_type` | [tss.exported.v1beta1.KeyType](#tss.exported.v1beta1.KeyType) |  |  |






<a name="tss.v1beta1.KeyRecoveryInfo"></a>

### KeyRecoveryInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `public` | [bytes](#bytes) |  |  |
| `private` | [KeyRecoveryInfo.PrivateEntry](#tss.v1beta1.KeyRecoveryInfo.PrivateEntry) | repeated |  |






<a name="tss.v1beta1.KeyRecoveryInfo.PrivateEntry"></a>

### KeyRecoveryInfo.PrivateEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [bytes](#bytes) |  |  |






<a name="tss.v1beta1.KeygenVoteData"></a>

### KeygenVoteData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pub_key` | [bytes](#bytes) |  |  |
| `group_recovery_info` | [bytes](#bytes) |  |  |






<a name="tss.v1beta1.MultisigInfo"></a>

### MultisigInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `timeout` | [int64](#int64) |  |  |
| `target_num` | [int64](#int64) |  |  |
| `infos` | [MultisigInfo.Info](#tss.v1beta1.MultisigInfo.Info) | repeated |  |






<a name="tss.v1beta1.MultisigInfo.Info"></a>

### MultisigInfo.Info



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `participant` | [bytes](#bytes) |  |  |
| `data` | [bytes](#bytes) | repeated |  |






<a name="tss.v1beta1.ValidatorStatus"></a>

### ValidatorStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validator` | [bytes](#bytes) |  |  |
| `suspended_until` | [uint64](#uint64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="tss/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## tss/v1beta1/genesis.proto



<a name="tss.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#tss.v1beta1.Params) |  |  |
| `key_recovery_infos` | [KeyRecoveryInfo](#tss.v1beta1.KeyRecoveryInfo) | repeated |  |
| `keys` | [tss.exported.v1beta1.Key](#tss.exported.v1beta1.Key) | repeated |  |
| `multisig_infos` | [MultisigInfo](#tss.v1beta1.MultisigInfo) | repeated |  |
| `external_keys` | [ExternalKeys](#tss.v1beta1.ExternalKeys) | repeated |  |
| `signatures` | [tss.exported.v1beta1.Signature](#tss.exported.v1beta1.Signature) | repeated |  |
| `validator_statuses` | [ValidatorStatus](#tss.v1beta1.ValidatorStatus) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="tss/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## tss/v1beta1/query.proto



<a name="tss.v1beta1.QueryActiveOldKeysResponse"></a>

### QueryActiveOldKeysResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_ids` | [string](#string) | repeated |  |






<a name="tss.v1beta1.QueryActiveOldKeysValidatorResponse"></a>

### QueryActiveOldKeysValidatorResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `keys_info` | [QueryActiveOldKeysValidatorResponse.KeyInfo](#tss.v1beta1.QueryActiveOldKeysValidatorResponse.KeyInfo) | repeated |  |






<a name="tss.v1beta1.QueryActiveOldKeysValidatorResponse.KeyInfo"></a>

### QueryActiveOldKeysValidatorResponse.KeyInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `chain` | [string](#string) |  |  |
| `role` | [int32](#int32) |  |  |






<a name="tss.v1beta1.QueryDeactivatedOperatorsResponse"></a>

### QueryDeactivatedOperatorsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `operator_addresses` | [string](#string) | repeated |  |






<a name="tss.v1beta1.QueryExternalKeyIDResponse"></a>

### QueryExternalKeyIDResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_ids` | [string](#string) | repeated |  |






<a name="tss.v1beta1.QueryKeyResponse"></a>

### QueryKeyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ecdsa_key` | [QueryKeyResponse.ECDSAKey](#tss.v1beta1.QueryKeyResponse.ECDSAKey) |  |  |
| `multisig_key` | [QueryKeyResponse.MultisigKey](#tss.v1beta1.QueryKeyResponse.MultisigKey) |  |  |
| `role` | [tss.exported.v1beta1.KeyRole](#tss.exported.v1beta1.KeyRole) |  |  |
| `rotated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="tss.v1beta1.QueryKeyResponse.ECDSAKey"></a>

### QueryKeyResponse.ECDSAKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `vote_status` | [VoteStatus](#tss.v1beta1.VoteStatus) |  |  |
| `key` | [QueryKeyResponse.Key](#tss.v1beta1.QueryKeyResponse.Key) |  |  |






<a name="tss.v1beta1.QueryKeyResponse.Key"></a>

### QueryKeyResponse.Key



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `x` | [string](#string) |  |  |
| `y` | [string](#string) |  |  |






<a name="tss.v1beta1.QueryKeyResponse.MultisigKey"></a>

### QueryKeyResponse.MultisigKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `threshold` | [int64](#int64) |  |  |
| `key` | [QueryKeyResponse.Key](#tss.v1beta1.QueryKeyResponse.Key) | repeated |  |






<a name="tss.v1beta1.QueryKeyShareResponse"></a>

### QueryKeyShareResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `share_infos` | [QueryKeyShareResponse.ShareInfo](#tss.v1beta1.QueryKeyShareResponse.ShareInfo) | repeated |  |






<a name="tss.v1beta1.QueryKeyShareResponse.ShareInfo"></a>

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






<a name="tss.v1beta1.QueryNextKeyIDRequest"></a>

### QueryNextKeyIDRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `key_role` | [tss.exported.v1beta1.KeyRole](#tss.exported.v1beta1.KeyRole) |  |  |






<a name="tss.v1beta1.QueryNextKeyIDResponse"></a>

### QueryNextKeyIDResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |






<a name="tss.v1beta1.QueryRecoveryResponse"></a>

### QueryRecoveryResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `party_uids` | [string](#string) | repeated |  |
| `party_share_counts` | [uint32](#uint32) | repeated |  |
| `threshold` | [uint32](#uint32) |  |  |
| `keygen_output` | [tss.tofnd.v1beta1.KeygenOutput](#tss.tofnd.v1beta1.KeygenOutput) |  |  |






<a name="tss.v1beta1.QuerySignatureResponse"></a>

### QuerySignatureResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `threshold_signature` | [QuerySignatureResponse.ThresholdSignature](#tss.v1beta1.QuerySignatureResponse.ThresholdSignature) |  |  |
| `multisig_signature` | [QuerySignatureResponse.MultisigSignature](#tss.v1beta1.QuerySignatureResponse.MultisigSignature) |  |  |






<a name="tss.v1beta1.QuerySignatureResponse.MultisigSignature"></a>

### QuerySignatureResponse.MultisigSignature



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sig_status` | [tss.exported.v1beta1.SigStatus](#tss.exported.v1beta1.SigStatus) |  |  |
| `signatures` | [QuerySignatureResponse.Signature](#tss.v1beta1.QuerySignatureResponse.Signature) | repeated |  |






<a name="tss.v1beta1.QuerySignatureResponse.Signature"></a>

### QuerySignatureResponse.Signature



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `r` | [string](#string) |  |  |
| `s` | [string](#string) |  |  |






<a name="tss.v1beta1.QuerySignatureResponse.ThresholdSignature"></a>

### QuerySignatureResponse.ThresholdSignature



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `vote_status` | [VoteStatus](#tss.v1beta1.VoteStatus) |  |  |
| `signature` | [QuerySignatureResponse.Signature](#tss.v1beta1.QuerySignatureResponse.Signature) |  |  |





 <!-- end messages -->


<a name="tss.v1beta1.VoteStatus"></a>

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



<a name="tss/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## tss/v1beta1/tx.proto



<a name="tss.v1beta1.HeartBeatRequest"></a>

### HeartBeatRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `key_ids` | [string](#string) | repeated |  |






<a name="tss.v1beta1.HeartBeatResponse"></a>

### HeartBeatResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `keygen_illegibility` | [int32](#int32) |  |  |
| `signing_illegibility` | [int32](#int32) |  |  |






<a name="tss.v1beta1.ProcessKeygenTrafficRequest"></a>

### ProcessKeygenTrafficRequest
ProcessKeygenTrafficRequest protocol message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `session_id` | [string](#string) |  |  |
| `payload` | [tss.tofnd.v1beta1.TrafficOut](#tss.tofnd.v1beta1.TrafficOut) |  |  |






<a name="tss.v1beta1.ProcessKeygenTrafficResponse"></a>

### ProcessKeygenTrafficResponse







<a name="tss.v1beta1.ProcessSignTrafficRequest"></a>

### ProcessSignTrafficRequest
ProcessSignTrafficRequest protocol message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `session_id` | [string](#string) |  |  |
| `payload` | [tss.tofnd.v1beta1.TrafficOut](#tss.tofnd.v1beta1.TrafficOut) |  |  |






<a name="tss.v1beta1.ProcessSignTrafficResponse"></a>

### ProcessSignTrafficResponse







<a name="tss.v1beta1.RegisterExternalKeysRequest"></a>

### RegisterExternalKeysRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `external_keys` | [RegisterExternalKeysRequest.ExternalKey](#tss.v1beta1.RegisterExternalKeysRequest.ExternalKey) | repeated |  |






<a name="tss.v1beta1.RegisterExternalKeysRequest.ExternalKey"></a>

### RegisterExternalKeysRequest.ExternalKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `pub_key` | [bytes](#bytes) |  |  |






<a name="tss.v1beta1.RegisterExternalKeysResponse"></a>

### RegisterExternalKeysResponse







<a name="tss.v1beta1.RotateKeyRequest"></a>

### RotateKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `key_role` | [tss.exported.v1beta1.KeyRole](#tss.exported.v1beta1.KeyRole) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="tss.v1beta1.RotateKeyResponse"></a>

### RotateKeyResponse







<a name="tss.v1beta1.StartKeygenRequest"></a>

### StartKeygenRequest
StartKeygenRequest indicate the start of keygen


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `key_info` | [KeyInfo](#tss.v1beta1.KeyInfo) |  |  |






<a name="tss.v1beta1.StartKeygenResponse"></a>

### StartKeygenResponse







<a name="tss.v1beta1.SubmitMultisigPubKeysRequest"></a>

### SubmitMultisigPubKeysRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |
| `sig_key_pairs` | [tss.exported.v1beta1.SigKeyPair](#tss.exported.v1beta1.SigKeyPair) | repeated |  |






<a name="tss.v1beta1.SubmitMultisigPubKeysResponse"></a>

### SubmitMultisigPubKeysResponse







<a name="tss.v1beta1.SubmitMultisigSignaturesRequest"></a>

### SubmitMultisigSignaturesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `sig_id` | [string](#string) |  |  |
| `signatures` | [bytes](#bytes) | repeated |  |






<a name="tss.v1beta1.SubmitMultisigSignaturesResponse"></a>

### SubmitMultisigSignaturesResponse







<a name="tss.v1beta1.VotePubKeyRequest"></a>

### VotePubKeyRequest
VotePubKeyRequest represents the message to vote on a public key


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll_key` | [vote.exported.v1beta1.PollKey](#vote.exported.v1beta1.PollKey) |  |  |
| `result` | [tss.tofnd.v1beta1.MessageOut.KeygenResult](#tss.tofnd.v1beta1.MessageOut.KeygenResult) |  |  |






<a name="tss.v1beta1.VotePubKeyResponse"></a>

### VotePubKeyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |






<a name="tss.v1beta1.VoteSigRequest"></a>

### VoteSigRequest
VoteSigRequest represents a message to vote for a signature


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll_key` | [vote.exported.v1beta1.PollKey](#vote.exported.v1beta1.PollKey) |  |  |
| `result` | [tss.tofnd.v1beta1.MessageOut.SignResult](#tss.tofnd.v1beta1.MessageOut.SignResult) |  |  |






<a name="tss.v1beta1.VoteSigResponse"></a>

### VoteSigResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="tss/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## tss/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="tss.v1beta1.MsgService"></a>

### MsgService
Msg defines the tss Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterExternalKeys` | [RegisterExternalKeysRequest](#tss.v1beta1.RegisterExternalKeysRequest) | [RegisterExternalKeysResponse](#tss.v1beta1.RegisterExternalKeysResponse) |  | POST|/axelar/tss/register-external-key|
| `HeartBeat` | [HeartBeatRequest](#tss.v1beta1.HeartBeatRequest) | [HeartBeatResponse](#tss.v1beta1.HeartBeatResponse) |  | POST|/axelar/tss/heartbeat|
| `StartKeygen` | [StartKeygenRequest](#tss.v1beta1.StartKeygenRequest) | [StartKeygenResponse](#tss.v1beta1.StartKeygenResponse) |  | POST|/axelar/tss/startKeygen|
| `ProcessKeygenTraffic` | [ProcessKeygenTrafficRequest](#tss.v1beta1.ProcessKeygenTrafficRequest) | [ProcessKeygenTrafficResponse](#tss.v1beta1.ProcessKeygenTrafficResponse) |  | ||
| `RotateKey` | [RotateKeyRequest](#tss.v1beta1.RotateKeyRequest) | [RotateKeyResponse](#tss.v1beta1.RotateKeyResponse) |  | POST|/axelar/tss/assign/{chain}|
| `VotePubKey` | [VotePubKeyRequest](#tss.v1beta1.VotePubKeyRequest) | [VotePubKeyResponse](#tss.v1beta1.VotePubKeyResponse) |  | ||
| `ProcessSignTraffic` | [ProcessSignTrafficRequest](#tss.v1beta1.ProcessSignTrafficRequest) | [ProcessSignTrafficResponse](#tss.v1beta1.ProcessSignTrafficResponse) |  | ||
| `VoteSig` | [VoteSigRequest](#tss.v1beta1.VoteSigRequest) | [VoteSigResponse](#tss.v1beta1.VoteSigResponse) |  | ||
| `SubmitMultisigPubKeys` | [SubmitMultisigPubKeysRequest](#tss.v1beta1.SubmitMultisigPubKeysRequest) | [SubmitMultisigPubKeysResponse](#tss.v1beta1.SubmitMultisigPubKeysResponse) |  | ||
| `SubmitMultisigSignatures` | [SubmitMultisigSignaturesRequest](#tss.v1beta1.SubmitMultisigSignaturesRequest) | [SubmitMultisigSignaturesResponse](#tss.v1beta1.SubmitMultisigSignaturesResponse) |  | ||


<a name="tss.v1beta1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |

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

