<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [axelar/axelarnet/v1beta1/events.proto](#axelar/axelarnet/v1beta1/events.proto)
    - [AxelarTransferCompleted](#axelar.axelarnet.v1beta1.AxelarTransferCompleted)
    - [FeeCollected](#axelar.axelarnet.v1beta1.FeeCollected)
    - [IBCTransferCompleted](#axelar.axelarnet.v1beta1.IBCTransferCompleted)
    - [IBCTransferFailed](#axelar.axelarnet.v1beta1.IBCTransferFailed)
    - [IBCTransferRetried](#axelar.axelarnet.v1beta1.IBCTransferRetried)
    - [IBCTransferSent](#axelar.axelarnet.v1beta1.IBCTransferSent)
  
- [axelar/axelarnet/v1beta1/params.proto](#axelar/axelarnet/v1beta1/params.proto)
    - [Params](#axelar.axelarnet.v1beta1.Params)
  
- [axelar/axelarnet/v1beta1/types.proto](#axelar/axelarnet/v1beta1/types.proto)
    - [Asset](#axelar.axelarnet.v1beta1.Asset)
    - [CosmosChain](#axelar.axelarnet.v1beta1.CosmosChain)
    - [IBCTransfer](#axelar.axelarnet.v1beta1.IBCTransfer)
  
    - [IBCTransfer.Status](#axelar.axelarnet.v1beta1.IBCTransfer.Status)
  
- [axelar/utils/v1beta1/queuer.proto](#axelar/utils/v1beta1/queuer.proto)
    - [QueueState](#axelar.utils.v1beta1.QueueState)
    - [QueueState.Item](#axelar.utils.v1beta1.QueueState.Item)
    - [QueueState.ItemsEntry](#axelar.utils.v1beta1.QueueState.ItemsEntry)
  
- [axelar/axelarnet/v1beta1/genesis.proto](#axelar/axelarnet/v1beta1/genesis.proto)
    - [GenesisState](#axelar.axelarnet.v1beta1.GenesisState)
  
- [axelar/utils/v1beta1/threshold.proto](#axelar/utils/v1beta1/threshold.proto)
    - [Threshold](#axelar.utils.v1beta1.Threshold)
  
- [axelar/tss/exported/v1beta1/types.proto](#axelar/tss/exported/v1beta1/types.proto)
    - [KeyRequirement](#axelar.tss.exported.v1beta1.KeyRequirement)
    - [SigKeyPair](#axelar.tss.exported.v1beta1.SigKeyPair)
  
    - [KeyRole](#axelar.tss.exported.v1beta1.KeyRole)
    - [KeyShareDistributionPolicy](#axelar.tss.exported.v1beta1.KeyShareDistributionPolicy)
    - [KeyType](#axelar.tss.exported.v1beta1.KeyType)
  
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
    - [RecipientAddressRequest](#axelar.nexus.v1beta1.RecipientAddressRequest)
    - [RecipientAddressResponse](#axelar.nexus.v1beta1.RecipientAddressResponse)
    - [TransferFeeRequest](#axelar.nexus.v1beta1.TransferFeeRequest)
    - [TransferFeeResponse](#axelar.nexus.v1beta1.TransferFeeResponse)
    - [TransfersForChainRequest](#axelar.nexus.v1beta1.TransfersForChainRequest)
    - [TransfersForChainResponse](#axelar.nexus.v1beta1.TransfersForChainResponse)
  
    - [ChainStatus](#axelar.nexus.v1beta1.ChainStatus)
  
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
    - [RetryIBCTransferRequest](#axelar.axelarnet.v1beta1.RetryIBCTransferRequest)
    - [RetryIBCTransferResponse](#axelar.axelarnet.v1beta1.RetryIBCTransferResponse)
    - [RouteIBCTransfersRequest](#axelar.axelarnet.v1beta1.RouteIBCTransfersRequest)
    - [RouteIBCTransfersResponse](#axelar.axelarnet.v1beta1.RouteIBCTransfersResponse)
  
- [axelar/axelarnet/v1beta1/service.proto](#axelar/axelarnet/v1beta1/service.proto)
    - [MsgService](#axelar.axelarnet.v1beta1.MsgService)
    - [QueryService](#axelar.axelarnet.v1beta1.QueryService)
  
- [axelar/snapshot/exported/v1beta1/types.proto](#axelar/snapshot/exported/v1beta1/types.proto)
    - [Participant](#axelar.snapshot.exported.v1beta1.Participant)
    - [Snapshot](#axelar.snapshot.exported.v1beta1.Snapshot)
    - [Snapshot.ParticipantsEntry](#axelar.snapshot.exported.v1beta1.Snapshot.ParticipantsEntry)
  
- [axelar/vote/exported/v1beta1/types.proto](#axelar/vote/exported/v1beta1/types.proto)
    - [PollKey](#axelar.vote.exported.v1beta1.PollKey)
    - [PollMetadata](#axelar.vote.exported.v1beta1.PollMetadata)
    - [PollParticipants](#axelar.vote.exported.v1beta1.PollParticipants)
  
    - [PollState](#axelar.vote.exported.v1beta1.PollState)
  
- [axelar/multisig/exported/v1beta1/types.proto](#axelar/multisig/exported/v1beta1/types.proto)
    - [KeyState](#axelar.multisig.exported.v1beta1.KeyState)
    - [MultisigState](#axelar.multisig.exported.v1beta1.MultisigState)
  
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
    - [EventTokenDeployed](#axelar.evm.v1beta1.EventTokenDeployed)
    - [EventTokenSent](#axelar.evm.v1beta1.EventTokenSent)
    - [EventTransfer](#axelar.evm.v1beta1.EventTransfer)
    - [Gateway](#axelar.evm.v1beta1.Gateway)
    - [NetworkInfo](#axelar.evm.v1beta1.NetworkInfo)
    - [PollMetadata](#axelar.evm.v1beta1.PollMetadata)
    - [SigMetadata](#axelar.evm.v1beta1.SigMetadata)
    - [TokenDetails](#axelar.evm.v1beta1.TokenDetails)
    - [TransactionMetadata](#axelar.evm.v1beta1.TransactionMetadata)
    - [TransferKey](#axelar.evm.v1beta1.TransferKey)
    - [VoteEvents](#axelar.evm.v1beta1.VoteEvents)
  
    - [BatchedCommandsStatus](#axelar.evm.v1beta1.BatchedCommandsStatus)
    - [DepositStatus](#axelar.evm.v1beta1.DepositStatus)
    - [Event.Status](#axelar.evm.v1beta1.Event.Status)
    - [SigType](#axelar.evm.v1beta1.SigType)
    - [Status](#axelar.evm.v1beta1.Status)
  
- [axelar/evm/v1beta1/events.proto](#axelar/evm/v1beta1/events.proto)
    - [BurnCommand](#axelar.evm.v1beta1.BurnCommand)
    - [ChainAdded](#axelar.evm.v1beta1.ChainAdded)
    - [CommandBatchAborted](#axelar.evm.v1beta1.CommandBatchAborted)
    - [CommandBatchSigned](#axelar.evm.v1beta1.CommandBatchSigned)
    - [ConfirmDepositStarted](#axelar.evm.v1beta1.ConfirmDepositStarted)
    - [ConfirmGatewayTxStarted](#axelar.evm.v1beta1.ConfirmGatewayTxStarted)
    - [ConfirmKeyTransferStarted](#axelar.evm.v1beta1.ConfirmKeyTransferStarted)
    - [ConfirmTokenStarted](#axelar.evm.v1beta1.ConfirmTokenStarted)
    - [ContractCallApproved](#axelar.evm.v1beta1.ContractCallApproved)
    - [ContractCallWithMintApproved](#axelar.evm.v1beta1.ContractCallWithMintApproved)
    - [EVMEventCompleted](#axelar.evm.v1beta1.EVMEventCompleted)
    - [EVMEventConfirmed](#axelar.evm.v1beta1.EVMEventConfirmed)
    - [EVMEventFailed](#axelar.evm.v1beta1.EVMEventFailed)
    - [EVMEventRetryFailed](#axelar.evm.v1beta1.EVMEventRetryFailed)
    - [MintCommand](#axelar.evm.v1beta1.MintCommand)
    - [NoEventsConfirmed](#axelar.evm.v1beta1.NoEventsConfirmed)
    - [PollCompleted](#axelar.evm.v1beta1.PollCompleted)
    - [PollExpired](#axelar.evm.v1beta1.PollExpired)
    - [PollFailed](#axelar.evm.v1beta1.PollFailed)
    - [TokenSent](#axelar.evm.v1beta1.TokenSent)
  
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
    - [ERC20TokensRequest](#axelar.evm.v1beta1.ERC20TokensRequest)
    - [ERC20TokensResponse](#axelar.evm.v1beta1.ERC20TokensResponse)
    - [ERC20TokensResponse.Token](#axelar.evm.v1beta1.ERC20TokensResponse.Token)
    - [EventRequest](#axelar.evm.v1beta1.EventRequest)
    - [EventResponse](#axelar.evm.v1beta1.EventResponse)
    - [GatewayAddressRequest](#axelar.evm.v1beta1.GatewayAddressRequest)
    - [GatewayAddressResponse](#axelar.evm.v1beta1.GatewayAddressResponse)
    - [KeyAddressRequest](#axelar.evm.v1beta1.KeyAddressRequest)
    - [KeyAddressResponse](#axelar.evm.v1beta1.KeyAddressResponse)
    - [KeyAddressResponse.WeightedAddress](#axelar.evm.v1beta1.KeyAddressResponse.WeightedAddress)
    - [PendingCommandsRequest](#axelar.evm.v1beta1.PendingCommandsRequest)
    - [PendingCommandsResponse](#axelar.evm.v1beta1.PendingCommandsResponse)
    - [Proof](#axelar.evm.v1beta1.Proof)
    - [QueryBurnerAddressResponse](#axelar.evm.v1beta1.QueryBurnerAddressResponse)
    - [QueryCommandResponse](#axelar.evm.v1beta1.QueryCommandResponse)
    - [QueryCommandResponse.ParamsEntry](#axelar.evm.v1beta1.QueryCommandResponse.ParamsEntry)
    - [QueryDepositStateParams](#axelar.evm.v1beta1.QueryDepositStateParams)
    - [QueryTokenAddressResponse](#axelar.evm.v1beta1.QueryTokenAddressResponse)
    - [TokenInfoRequest](#axelar.evm.v1beta1.TokenInfoRequest)
    - [TokenInfoResponse](#axelar.evm.v1beta1.TokenInfoResponse)
  
    - [TokenType](#axelar.evm.v1beta1.TokenType)
  
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
  
- [axelar/multisig/v1beta1/events.proto](#axelar/multisig/v1beta1/events.proto)
    - [KeyAssigned](#axelar.multisig.v1beta1.KeyAssigned)
    - [KeyRotated](#axelar.multisig.v1beta1.KeyRotated)
    - [KeygenCompleted](#axelar.multisig.v1beta1.KeygenCompleted)
    - [KeygenExpired](#axelar.multisig.v1beta1.KeygenExpired)
    - [KeygenStarted](#axelar.multisig.v1beta1.KeygenStarted)
    - [PubKeySubmitted](#axelar.multisig.v1beta1.PubKeySubmitted)
    - [SignatureSubmitted](#axelar.multisig.v1beta1.SignatureSubmitted)
    - [SigningCompleted](#axelar.multisig.v1beta1.SigningCompleted)
    - [SigningExpired](#axelar.multisig.v1beta1.SigningExpired)
    - [SigningStarted](#axelar.multisig.v1beta1.SigningStarted)
    - [SigningStarted.PubKeysEntry](#axelar.multisig.v1beta1.SigningStarted.PubKeysEntry)
  
- [axelar/multisig/v1beta1/params.proto](#axelar/multisig/v1beta1/params.proto)
    - [Params](#axelar.multisig.v1beta1.Params)
  
- [axelar/multisig/v1beta1/types.proto](#axelar/multisig/v1beta1/types.proto)
    - [Key](#axelar.multisig.v1beta1.Key)
    - [Key.PubKeysEntry](#axelar.multisig.v1beta1.Key.PubKeysEntry)
    - [KeyEpoch](#axelar.multisig.v1beta1.KeyEpoch)
    - [KeygenSession](#axelar.multisig.v1beta1.KeygenSession)
    - [KeygenSession.IsPubKeyReceivedEntry](#axelar.multisig.v1beta1.KeygenSession.IsPubKeyReceivedEntry)
    - [MultiSig](#axelar.multisig.v1beta1.MultiSig)
    - [MultiSig.SigsEntry](#axelar.multisig.v1beta1.MultiSig.SigsEntry)
    - [SigningSession](#axelar.multisig.v1beta1.SigningSession)
  
- [axelar/multisig/v1beta1/genesis.proto](#axelar/multisig/v1beta1/genesis.proto)
    - [GenesisState](#axelar.multisig.v1beta1.GenesisState)
  
- [axelar/multisig/v1beta1/query.proto](#axelar/multisig/v1beta1/query.proto)
    - [KeyIDRequest](#axelar.multisig.v1beta1.KeyIDRequest)
    - [KeyIDResponse](#axelar.multisig.v1beta1.KeyIDResponse)
    - [KeyRequest](#axelar.multisig.v1beta1.KeyRequest)
    - [KeyResponse](#axelar.multisig.v1beta1.KeyResponse)
    - [KeygenParticipant](#axelar.multisig.v1beta1.KeygenParticipant)
    - [KeygenSessionRequest](#axelar.multisig.v1beta1.KeygenSessionRequest)
    - [KeygenSessionResponse](#axelar.multisig.v1beta1.KeygenSessionResponse)
    - [NextKeyIDRequest](#axelar.multisig.v1beta1.NextKeyIDRequest)
    - [NextKeyIDResponse](#axelar.multisig.v1beta1.NextKeyIDResponse)
  
- [axelar/multisig/v1beta1/tx.proto](#axelar/multisig/v1beta1/tx.proto)
    - [RotateKeyRequest](#axelar.multisig.v1beta1.RotateKeyRequest)
    - [RotateKeyResponse](#axelar.multisig.v1beta1.RotateKeyResponse)
    - [StartKeygenRequest](#axelar.multisig.v1beta1.StartKeygenRequest)
    - [StartKeygenResponse](#axelar.multisig.v1beta1.StartKeygenResponse)
    - [SubmitPubKeyRequest](#axelar.multisig.v1beta1.SubmitPubKeyRequest)
    - [SubmitPubKeyResponse](#axelar.multisig.v1beta1.SubmitPubKeyResponse)
    - [SubmitSignatureRequest](#axelar.multisig.v1beta1.SubmitSignatureRequest)
    - [SubmitSignatureResponse](#axelar.multisig.v1beta1.SubmitSignatureResponse)
  
- [axelar/multisig/v1beta1/service.proto](#axelar/multisig/v1beta1/service.proto)
    - [MsgService](#axelar.multisig.v1beta1.MsgService)
    - [QueryService](#axelar.multisig.v1beta1.QueryService)
  
- [axelar/nexus/v1beta1/events.proto](#axelar/nexus/v1beta1/events.proto)
    - [FeeDeducted](#axelar.nexus.v1beta1.FeeDeducted)
    - [InsufficientFee](#axelar.nexus.v1beta1.InsufficientFee)
  
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
  
- [axelar/vote/v1beta1/events.proto](#axelar/vote/v1beta1/events.proto)
    - [Voted](#axelar.vote.v1beta1.Voted)
  
- [axelar/vote/v1beta1/params.proto](#axelar/vote/v1beta1/params.proto)
    - [Params](#axelar.vote.v1beta1.Params)
  
- [axelar/vote/v1beta1/genesis.proto](#axelar/vote/v1beta1/genesis.proto)
    - [GenesisState](#axelar.vote.v1beta1.GenesisState)
  
- [axelar/vote/v1beta1/types.proto](#axelar/vote/v1beta1/types.proto)
    - [TalliedVote](#axelar.vote.v1beta1.TalliedVote)
    - [TalliedVote.IsVoterLateEntry](#axelar.vote.v1beta1.TalliedVote.IsVoterLateEntry)
  
- [axelar/vote/v1beta1/tx.proto](#axelar/vote/v1beta1/tx.proto)
    - [VoteRequest](#axelar.vote.v1beta1.VoteRequest)
    - [VoteResponse](#axelar.vote.v1beta1.VoteResponse)
  
- [axelar/vote/v1beta1/service.proto](#axelar/vote/v1beta1/service.proto)
    - [MsgService](#axelar.vote.v1beta1.MsgService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="axelar/axelarnet/v1beta1/events.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/axelarnet/v1beta1/events.proto



<a name="axelar.axelarnet.v1beta1.AxelarTransferCompleted"></a>

### AxelarTransferCompleted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  |  |
| `receipient` | [string](#string) |  |  |
| `asset` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="axelar.axelarnet.v1beta1.FeeCollected"></a>

### FeeCollected



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collector` | [bytes](#bytes) |  |  |
| `fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="axelar.axelarnet.v1beta1.IBCTransferCompleted"></a>

### IBCTransferCompleted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  |  |
| `sequence` | [uint64](#uint64) |  |  |
| `port_id` | [string](#string) |  |  |
| `channel_id` | [string](#string) |  |  |






<a name="axelar.axelarnet.v1beta1.IBCTransferFailed"></a>

### IBCTransferFailed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  |  |
| `sequence` | [uint64](#uint64) |  |  |
| `port_id` | [string](#string) |  |  |
| `channel_id` | [string](#string) |  |  |






<a name="axelar.axelarnet.v1beta1.IBCTransferRetried"></a>

### IBCTransferRetried



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  |  |
| `receipient` | [string](#string) |  |  |
| `asset` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `sequence` | [uint64](#uint64) |  |  |
| `port_id` | [string](#string) |  |  |
| `channel_id` | [string](#string) |  |  |






<a name="axelar.axelarnet.v1beta1.IBCTransferSent"></a>

### IBCTransferSent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  |  |
| `receipient` | [string](#string) |  |  |
| `asset` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `sequence` | [uint64](#uint64) |  |  |
| `port_id` | [string](#string) |  |  |
| `channel_id` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/axelarnet/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/axelarnet/v1beta1/params.proto



<a name="axelar.axelarnet.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `route_timeout_window` | [uint64](#uint64) |  | IBC packet route timeout window |
| `transfer_limit` | [uint64](#uint64) |  |  |
| `end_blocker_limit` | [uint64](#uint64) |  |  |





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
| `sequence` | [uint64](#uint64) |  | **Deprecated.**  |
| `id` | [uint64](#uint64) |  |  |
| `status` | [IBCTransfer.Status](#axelar.axelarnet.v1beta1.IBCTransfer.Status) |  |  |





 <!-- end messages -->


<a name="axelar.axelarnet.v1beta1.IBCTransfer.Status"></a>

### IBCTransfer.Status


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATUS_UNSPECIFIED | 0 |  |
| STATUS_PENDING | 1 |  |
| STATUS_COMPLETED | 2 |  |
| STATUS_FAILED | 3 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

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
| `transfer_queue` | [axelar.utils.v1beta1.QueueState](#axelar.utils.v1beta1.QueueState) |  |  |
| `failed_transfers` | [IBCTransfer](#axelar.axelarnet.v1beta1.IBCTransfer) | repeated |  |





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





 <!-- end messages -->


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
| `activated` | [bool](#bool) |  |  |
| `assets` | [axelar.nexus.exported.v1beta1.Asset](#axelar.nexus.exported.v1beta1.Asset) | repeated |  |
| `maintainer_states` | [MaintainerState](#axelar.nexus.v1beta1.MaintainerState) | repeated | **Deprecated.**  |






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
| `chain` | [string](#string) |  |  |





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


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `status` | [ChainStatus](#axelar.nexus.v1beta1.ChainStatus) |  |  |






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






<a name="axelar.nexus.v1beta1.RecipientAddressRequest"></a>

### RecipientAddressRequest
RecipientAddressRequest represents a message that queries the registered
recipient address for a given deposit address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposit_addr` | [string](#string) |  |  |
| `deposit_chain` | [string](#string) |  |  |






<a name="axelar.nexus.v1beta1.RecipientAddressResponse"></a>

### RecipientAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `recipient_addr` | [string](#string) |  |  |
| `recipient_chain` | [string](#string) |  |  |






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


<a name="axelar.nexus.v1beta1.ChainStatus"></a>

### ChainStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| CHAIN_STATUS_UNSPECIFIED | 0 |  |
| CHAIN_STATUS_ACTIVATED | 1 |  |
| CHAIN_STATUS_DEACTIVATED | 2 |  |


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
| `chain` | [axelar.nexus.exported.v1beta1.Chain](#axelar.nexus.exported.v1beta1.Chain) |  | **Deprecated.** chain was deprecated in v0.27 |
| `addr_prefix` | [string](#string) |  |  |
| `native_assets` | [axelar.nexus.exported.v1beta1.Asset](#axelar.nexus.exported.v1beta1.Asset) | repeated |  |
| `cosmos_chain` | [string](#string) |  | TODO: Rename this to `chain` after v1beta1 -> v1 version bump |






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







<a name="axelar.axelarnet.v1beta1.RetryIBCTransferRequest"></a>

### RetryIBCTransferRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `id` | [uint64](#uint64) |  |  |






<a name="axelar.axelarnet.v1beta1.RetryIBCTransferResponse"></a>

### RetryIBCTransferResponse







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
| `RetryIBCTransfer` | [RetryIBCTransferRequest](#axelar.axelarnet.v1beta1.RetryIBCTransferRequest) | [RetryIBCTransferResponse](#axelar.axelarnet.v1beta1.RetryIBCTransferResponse) |  | POST|/axelar/axelarnet/retry_ibc_transfer|


<a name="axelar.axelarnet.v1beta1.QueryService"></a>

### QueryService
QueryService defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `PendingIBCTransferCount` | [PendingIBCTransferCountRequest](#axelar.axelarnet.v1beta1.PendingIBCTransferCountRequest) | [PendingIBCTransferCountResponse](#axelar.axelarnet.v1beta1.PendingIBCTransferCountResponse) | PendingIBCTransferCount queries the pending ibc transfers for all chains | GET|/axelar/axelarnet/v1beta1/ibc_transfer_count|

 <!-- end services -->



<a name="axelar/snapshot/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/snapshot/exported/v1beta1/types.proto



<a name="axelar.snapshot.exported.v1beta1.Participant"></a>

### Participant



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [bytes](#bytes) |  |  |
| `weight` | [bytes](#bytes) |  |  |






<a name="axelar.snapshot.exported.v1beta1.Snapshot"></a>

### Snapshot



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `height` | [int64](#int64) |  |  |
| `participants` | [Snapshot.ParticipantsEntry](#axelar.snapshot.exported.v1beta1.Snapshot.ParticipantsEntry) | repeated |  |
| `bonded_weight` | [bytes](#bytes) |  |  |






<a name="axelar.snapshot.exported.v1beta1.Snapshot.ParticipantsEntry"></a>

### Snapshot.ParticipantsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [Participant](#axelar.snapshot.exported.v1beta1.Participant) |  |  |





 <!-- end messages -->

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
| `expires_at` | [int64](#int64) |  |  |
| `result` | [google.protobuf.Any](#google.protobuf.Any) |  |  |
| `voting_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `state` | [PollState](#axelar.vote.exported.v1beta1.PollState) |  |  |
| `min_voter_count` | [int64](#int64) |  |  |
| `reward_pool_name` | [string](#string) |  |  |
| `grace_period` | [int64](#int64) |  |  |
| `completed_at` | [int64](#int64) |  |  |
| `id` | [uint64](#uint64) |  |  |
| `snapshot` | [axelar.snapshot.exported.v1beta1.Snapshot](#axelar.snapshot.exported.v1beta1.Snapshot) |  |  |
| `module` | [string](#string) |  |  |
| `module_metadata` | [google.protobuf.Any](#google.protobuf.Any) |  |  |






<a name="axelar.vote.exported.v1beta1.PollParticipants"></a>

### PollParticipants
PollParticipants should be embedded in poll events in other modules


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `poll_id` | [uint64](#uint64) |  |  |
| `participants` | [bytes](#bytes) | repeated |  |





 <!-- end messages -->


<a name="axelar.vote.exported.v1beta1.PollState"></a>

### PollState


| Name | Number | Description |
| ---- | ------ | ----------- |
| POLL_STATE_UNSPECIFIED | 0 |  |
| POLL_STATE_PENDING | 1 |  |
| POLL_STATE_COMPLETED | 2 |  |
| POLL_STATE_FAILED | 3 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/multisig/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/multisig/exported/v1beta1/types.proto


 <!-- end messages -->


<a name="axelar.multisig.exported.v1beta1.KeyState"></a>

### KeyState


| Name | Number | Description |
| ---- | ------ | ----------- |
| KEY_STATE_UNSPECIFIED | 0 |  |
| KEY_STATE_ASSIGNED | 1 |  |
| KEY_STATE_ACTIVE | 2 |  |



<a name="axelar.multisig.exported.v1beta1.MultisigState"></a>

### MultisigState


| Name | Number | Description |
| ---- | ------ | ----------- |
| MULTISIG_STATE_UNSPECIFIED | 0 |  |
| MULTISIG_STATE_PENDING | 1 |  |
| MULTISIG_STATE_COMPLETED | 2 |  |


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
| `signature` | [google.protobuf.Any](#google.protobuf.Any) |  |  |






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
| `multisig_ownership_transferred` | [EventMultisigOwnershipTransferred](#axelar.evm.v1beta1.EventMultisigOwnershipTransferred) |  | **Deprecated.**  |
| `multisig_operatorship_transferred` | [EventMultisigOperatorshipTransferred](#axelar.evm.v1beta1.EventMultisigOperatorshipTransferred) |  |  |






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
| `new_operators` | [bytes](#bytes) | repeated |  |
| `new_threshold` | [bytes](#bytes) |  |  |
| `new_weights` | [bytes](#bytes) | repeated |  |






<a name="axelar.evm.v1beta1.EventMultisigOwnershipTransferred"></a>

### EventMultisigOwnershipTransferred



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pre_owners` | [bytes](#bytes) | repeated |  |
| `prev_threshold` | [bytes](#bytes) |  |  |
| `new_owners` | [bytes](#bytes) | repeated |  |
| `new_threshold` | [bytes](#bytes) |  |  |






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






<a name="axelar.evm.v1beta1.NetworkInfo"></a>

### NetworkInfo
NetworkInfo describes information about a network


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `id` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.PollMetadata"></a>

### PollMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.SigMetadata"></a>

### SigMetadata
SigMetadata stores necessary information for external apps to map signature
results to evm relay transaction types


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `type` | [SigType](#axelar.evm.v1beta1.SigType) |  |  |
| `chain` | [string](#string) |  |  |
| `command_batch_id` | [bytes](#bytes) |  |  |






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
TransferKey contains information for a transfer operatorship


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
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


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/evm/v1beta1/events.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/evm/v1beta1/events.proto



<a name="axelar.evm.v1beta1.BurnCommand"></a>

### BurnCommand



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `command_id` | [bytes](#bytes) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `deposit_address` | [string](#string) |  |  |
| `asset` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.ChainAdded"></a>

### ChainAdded



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.CommandBatchAborted"></a>

### CommandBatchAborted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `command_batch_id` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.CommandBatchSigned"></a>

### CommandBatchSigned



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `command_batch_id` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.ConfirmDepositStarted"></a>

### ConfirmDepositStarted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `deposit_address` | [bytes](#bytes) |  |  |
| `token_address` | [bytes](#bytes) |  |  |
| `confirmation_height` | [uint64](#uint64) |  |  |
| `participants` | [axelar.vote.exported.v1beta1.PollParticipants](#axelar.vote.exported.v1beta1.PollParticipants) |  |  |
| `asset` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.ConfirmGatewayTxStarted"></a>

### ConfirmGatewayTxStarted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `gateway_address` | [bytes](#bytes) |  |  |
| `confirmation_height` | [uint64](#uint64) |  |  |
| `participants` | [axelar.vote.exported.v1beta1.PollParticipants](#axelar.vote.exported.v1beta1.PollParticipants) |  |  |






<a name="axelar.evm.v1beta1.ConfirmKeyTransferStarted"></a>

### ConfirmKeyTransferStarted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `gateway_address` | [bytes](#bytes) |  |  |
| `confirmation_height` | [uint64](#uint64) |  |  |
| `participants` | [axelar.vote.exported.v1beta1.PollParticipants](#axelar.vote.exported.v1beta1.PollParticipants) |  |  |






<a name="axelar.evm.v1beta1.ConfirmTokenStarted"></a>

### ConfirmTokenStarted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `gateway_address` | [bytes](#bytes) |  |  |
| `token_address` | [bytes](#bytes) |  |  |
| `token_details` | [TokenDetails](#axelar.evm.v1beta1.TokenDetails) |  |  |
| `confirmation_height` | [uint64](#uint64) |  |  |
| `participants` | [axelar.vote.exported.v1beta1.PollParticipants](#axelar.vote.exported.v1beta1.PollParticipants) |  |  |






<a name="axelar.evm.v1beta1.ContractCallApproved"></a>

### ContractCallApproved



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `event_id` | [string](#string) |  |  |
| `command_id` | [bytes](#bytes) |  |  |
| `sender` | [string](#string) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `contract_address` | [string](#string) |  |  |
| `payload_hash` | [bytes](#bytes) |  |  |






<a name="axelar.evm.v1beta1.ContractCallWithMintApproved"></a>

### ContractCallWithMintApproved



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `event_id` | [string](#string) |  |  |
| `command_id` | [bytes](#bytes) |  |  |
| `sender` | [string](#string) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `contract_address` | [string](#string) |  |  |
| `payload_hash` | [bytes](#bytes) |  |  |
| `asset` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="axelar.evm.v1beta1.EVMEventCompleted"></a>

### EVMEventCompleted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `event_id` | [string](#string) |  |  |
| `type` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.EVMEventConfirmed"></a>

### EVMEventConfirmed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `event_id` | [string](#string) |  |  |
| `type` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.EVMEventFailed"></a>

### EVMEventFailed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `event_id` | [string](#string) |  |  |
| `type` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.EVMEventRetryFailed"></a>

### EVMEventRetryFailed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `event_id` | [string](#string) |  |  |
| `type` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.MintCommand"></a>

### MintCommand



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `transfer_id` | [uint64](#uint64) |  |  |
| `command_id` | [bytes](#bytes) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `destination_address` | [string](#string) |  |  |
| `asset` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="axelar.evm.v1beta1.NoEventsConfirmed"></a>

### NoEventsConfirmed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `poll_id` | [uint64](#uint64) |  |  |






<a name="axelar.evm.v1beta1.PollCompleted"></a>

### PollCompleted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `poll_id` | [uint64](#uint64) |  |  |






<a name="axelar.evm.v1beta1.PollExpired"></a>

### PollExpired



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `poll_id` | [uint64](#uint64) |  |  |






<a name="axelar.evm.v1beta1.PollFailed"></a>

### PollFailed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `poll_id` | [uint64](#uint64) |  |  |






<a name="axelar.evm.v1beta1.TokenSent"></a>

### TokenSent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `event_id` | [string](#string) |  |  |
| `transfer_id` | [uint64](#uint64) |  |  |
| `sender` | [string](#string) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `destination_address` | [string](#string) |  |  |
| `asset` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |





 <!-- end messages -->

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
| `voting_grace_period` | [int64](#int64) |  |  |
| `end_blocker_limit` | [int64](#int64) |  |  |
| `transfer_limit` | [uint64](#uint64) |  |  |






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
| `execute_data` | [string](#string) |  |  |
| `prev_batched_commands_id` | [string](#string) |  |  |
| `command_ids` | [string](#string) | repeated |  |
| `proof` | [Proof](#axelar.evm.v1beta1.Proof) |  |  |






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






<a name="axelar.evm.v1beta1.ERC20TokensRequest"></a>

### ERC20TokensRequest
ERC20TokensRequest describes the chain for which the type of ERC20 tokens are
requested.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `type` | [TokenType](#axelar.evm.v1beta1.TokenType) |  |  |






<a name="axelar.evm.v1beta1.ERC20TokensResponse"></a>

### ERC20TokensResponse
ERC20TokensResponse describes the asset and symbol for all
ERC20 tokens requested for a chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tokens` | [ERC20TokensResponse.Token](#axelar.evm.v1beta1.ERC20TokensResponse.Token) | repeated |  |






<a name="axelar.evm.v1beta1.ERC20TokensResponse.Token"></a>

### ERC20TokensResponse.Token



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |






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
| `key_id` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.KeyAddressResponse"></a>

### KeyAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `addresses` | [KeyAddressResponse.WeightedAddress](#axelar.evm.v1beta1.KeyAddressResponse.WeightedAddress) | repeated |  |
| `threshold` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.KeyAddressResponse.WeightedAddress"></a>

### KeyAddressResponse.WeightedAddress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `weight` | [string](#string) |  |  |






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






<a name="axelar.evm.v1beta1.Proof"></a>

### Proof



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `addresses` | [string](#string) | repeated |  |
| `weights` | [string](#string) | repeated |  |
| `threshold` | [string](#string) |  |  |
| `signatures` | [string](#string) | repeated |  |






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






<a name="axelar.evm.v1beta1.QueryTokenAddressResponse"></a>

### QueryTokenAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `confirmed` | [bool](#bool) |  |  |






<a name="axelar.evm.v1beta1.TokenInfoRequest"></a>

### TokenInfoRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |
| `asset` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |






<a name="axelar.evm.v1beta1.TokenInfoResponse"></a>

### TokenInfoResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset` | [string](#string) |  |  |
| `details` | [TokenDetails](#axelar.evm.v1beta1.TokenDetails) |  |  |
| `address` | [string](#string) |  |  |
| `confirmed` | [bool](#bool) |  |  |
| `is_external` | [bool](#bool) |  |  |
| `burner_code_hash` | [string](#string) |  |  |





 <!-- end messages -->


<a name="axelar.evm.v1beta1.TokenType"></a>

### TokenType


| Name | Number | Description |
| ---- | ------ | ----------- |
| TOKEN_TYPE_UNSPECIFIED | 0 |  |
| TOKEN_TYPE_INTERNAL | 1 |  |
| TOKEN_TYPE_EXTERNAL | 2 |  |


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
| `key_type` | [axelar.tss.exported.v1beta1.KeyType](#axelar.tss.exported.v1beta1.KeyType) |  | **Deprecated.**  |
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
| `daily_mint_limit` | [string](#string) |  |  |






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
| `ERC20Tokens` | [ERC20TokensRequest](#axelar.evm.v1beta1.ERC20TokensRequest) | [ERC20TokensResponse](#axelar.evm.v1beta1.ERC20TokensResponse) | ERC20Tokens queries the ERC20 tokens registered for a chain | GET|/axelar/evm/v1beta1/erc20_tokens/{chain}|
| `TokenInfo` | [TokenInfoRequest](#axelar.evm.v1beta1.TokenInfoRequest) | [TokenInfoResponse](#axelar.evm.v1beta1.TokenInfoResponse) | TokenInfo queries the token info for a registered ERC20 Token | GET|/axelar/evm/v1beta1/token_info/{chain}|

 <!-- end services -->



<a name="axelar/multisig/v1beta1/events.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/multisig/v1beta1/events.proto



<a name="axelar.multisig.v1beta1.KeyAssigned"></a>

### KeyAssigned



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  |  |
| `chain` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="axelar.multisig.v1beta1.KeyRotated"></a>

### KeyRotated



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  |  |
| `chain` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="axelar.multisig.v1beta1.KeygenCompleted"></a>

### KeygenCompleted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="axelar.multisig.v1beta1.KeygenExpired"></a>

### KeygenExpired



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="axelar.multisig.v1beta1.KeygenStarted"></a>

### KeygenStarted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |
| `participants` | [bytes](#bytes) | repeated |  |






<a name="axelar.multisig.v1beta1.PubKeySubmitted"></a>

### PubKeySubmitted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |
| `participant` | [bytes](#bytes) |  |  |
| `pub_key` | [bytes](#bytes) |  |  |






<a name="axelar.multisig.v1beta1.SignatureSubmitted"></a>

### SignatureSubmitted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  |  |
| `sig_id` | [uint64](#uint64) |  |  |
| `participant` | [bytes](#bytes) |  |  |
| `signature` | [bytes](#bytes) |  |  |






<a name="axelar.multisig.v1beta1.SigningCompleted"></a>

### SigningCompleted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  |  |
| `sig_id` | [uint64](#uint64) |  |  |






<a name="axelar.multisig.v1beta1.SigningExpired"></a>

### SigningExpired



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  |  |
| `sig_id` | [uint64](#uint64) |  |  |






<a name="axelar.multisig.v1beta1.SigningStarted"></a>

### SigningStarted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  |  |
| `sig_id` | [uint64](#uint64) |  |  |
| `key_id` | [string](#string) |  |  |
| `pub_keys` | [SigningStarted.PubKeysEntry](#axelar.multisig.v1beta1.SigningStarted.PubKeysEntry) | repeated |  |
| `payload_hash` | [bytes](#bytes) |  |  |
| `requesting_module` | [string](#string) |  |  |






<a name="axelar.multisig.v1beta1.SigningStarted.PubKeysEntry"></a>

### SigningStarted.PubKeysEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [bytes](#bytes) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/multisig/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/multisig/v1beta1/params.proto



<a name="axelar.multisig.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `keygen_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `signing_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `keygen_timeout` | [int64](#int64) |  |  |
| `keygen_grace_period` | [int64](#int64) |  |  |
| `signing_timeout` | [int64](#int64) |  |  |
| `signing_grace_period` | [int64](#int64) |  |  |
| `active_epoch_count` | [uint64](#uint64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/multisig/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/multisig/v1beta1/types.proto



<a name="axelar.multisig.v1beta1.Key"></a>

### Key



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `snapshot` | [axelar.snapshot.exported.v1beta1.Snapshot](#axelar.snapshot.exported.v1beta1.Snapshot) |  |  |
| `pub_keys` | [Key.PubKeysEntry](#axelar.multisig.v1beta1.Key.PubKeysEntry) | repeated |  |
| `signing_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `state` | [axelar.multisig.exported.v1beta1.KeyState](#axelar.multisig.exported.v1beta1.KeyState) |  |  |






<a name="axelar.multisig.v1beta1.Key.PubKeysEntry"></a>

### Key.PubKeysEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [bytes](#bytes) |  |  |






<a name="axelar.multisig.v1beta1.KeyEpoch"></a>

### KeyEpoch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `epoch` | [uint64](#uint64) |  |  |
| `chain` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="axelar.multisig.v1beta1.KeygenSession"></a>

### KeygenSession



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [Key](#axelar.multisig.v1beta1.Key) |  |  |
| `state` | [axelar.multisig.exported.v1beta1.MultisigState](#axelar.multisig.exported.v1beta1.MultisigState) |  |  |
| `keygen_threshold` | [axelar.utils.v1beta1.Threshold](#axelar.utils.v1beta1.Threshold) |  |  |
| `expires_at` | [int64](#int64) |  |  |
| `completed_at` | [int64](#int64) |  |  |
| `is_pub_key_received` | [KeygenSession.IsPubKeyReceivedEntry](#axelar.multisig.v1beta1.KeygenSession.IsPubKeyReceivedEntry) | repeated |  |
| `grace_period` | [int64](#int64) |  |  |






<a name="axelar.multisig.v1beta1.KeygenSession.IsPubKeyReceivedEntry"></a>

### KeygenSession.IsPubKeyReceivedEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [bool](#bool) |  |  |






<a name="axelar.multisig.v1beta1.MultiSig"></a>

### MultiSig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `payload_hash` | [bytes](#bytes) |  |  |
| `sigs` | [MultiSig.SigsEntry](#axelar.multisig.v1beta1.MultiSig.SigsEntry) | repeated |  |






<a name="axelar.multisig.v1beta1.MultiSig.SigsEntry"></a>

### MultiSig.SigsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [bytes](#bytes) |  |  |






<a name="axelar.multisig.v1beta1.SigningSession"></a>

### SigningSession



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  |  |
| `multi_sig` | [MultiSig](#axelar.multisig.v1beta1.MultiSig) |  |  |
| `state` | [axelar.multisig.exported.v1beta1.MultisigState](#axelar.multisig.exported.v1beta1.MultisigState) |  |  |
| `key` | [Key](#axelar.multisig.v1beta1.Key) |  |  |
| `expires_at` | [int64](#int64) |  |  |
| `completed_at` | [int64](#int64) |  |  |
| `grace_period` | [int64](#int64) |  |  |
| `module` | [string](#string) |  |  |
| `module_metadata` | [google.protobuf.Any](#google.protobuf.Any) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/multisig/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/multisig/v1beta1/genesis.proto



<a name="axelar.multisig.v1beta1.GenesisState"></a>

### GenesisState
GenesisState represents the genesis state


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#axelar.multisig.v1beta1.Params) |  |  |
| `keygen_sessions` | [KeygenSession](#axelar.multisig.v1beta1.KeygenSession) | repeated |  |
| `signing_sessions` | [SigningSession](#axelar.multisig.v1beta1.SigningSession) | repeated |  |
| `keys` | [Key](#axelar.multisig.v1beta1.Key) | repeated |  |
| `key_epochs` | [KeyEpoch](#axelar.multisig.v1beta1.KeyEpoch) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/multisig/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/multisig/v1beta1/query.proto



<a name="axelar.multisig.v1beta1.KeyIDRequest"></a>

### KeyIDRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |






<a name="axelar.multisig.v1beta1.KeyIDResponse"></a>

### KeyIDResponse
KeyIDResponse contains the key ID of the key assigned to a given chain.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |






<a name="axelar.multisig.v1beta1.KeyRequest"></a>

### KeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |






<a name="axelar.multisig.v1beta1.KeyResponse"></a>

### KeyResponse
KeyResponse contains the key corresponding to a given key id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |
| `state` | [axelar.multisig.exported.v1beta1.KeyState](#axelar.multisig.exported.v1beta1.KeyState) |  |  |
| `started_at` | [int64](#int64) |  |  |
| `started_at_timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `threshold_weight` | [bytes](#bytes) |  |  |
| `bonded_weight` | [bytes](#bytes) |  |  |
| `participants` | [KeygenParticipant](#axelar.multisig.v1beta1.KeygenParticipant) | repeated | Keygen participants in descending order by weight |






<a name="axelar.multisig.v1beta1.KeygenParticipant"></a>

### KeygenParticipant



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `weight` | [bytes](#bytes) |  |  |
| `pub_key` | [string](#string) |  |  |






<a name="axelar.multisig.v1beta1.KeygenSessionRequest"></a>

### KeygenSessionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |






<a name="axelar.multisig.v1beta1.KeygenSessionResponse"></a>

### KeygenSessionResponse
KeygenSessionResponse contains the keygen session info for a given key ID.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `started_at` | [int64](#int64) |  |  |
| `started_at_timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `expires_at` | [int64](#int64) |  |  |
| `completed_at` | [int64](#int64) |  |  |
| `grace_period` | [int64](#int64) |  |  |
| `state` | [axelar.multisig.exported.v1beta1.MultisigState](#axelar.multisig.exported.v1beta1.MultisigState) |  |  |
| `keygen_threshold_weight` | [bytes](#bytes) |  |  |
| `signing_threshold_weight` | [bytes](#bytes) |  |  |
| `bonded_weight` | [bytes](#bytes) |  |  |
| `participants` | [KeygenParticipant](#axelar.multisig.v1beta1.KeygenParticipant) | repeated | Keygen candidates in descending order by weight |






<a name="axelar.multisig.v1beta1.NextKeyIDRequest"></a>

### NextKeyIDRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain` | [string](#string) |  |  |






<a name="axelar.multisig.v1beta1.NextKeyIDResponse"></a>

### NextKeyIDResponse
NextKeyIDResponse contains the key ID for the next rotation on the given
chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/multisig/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/multisig/v1beta1/tx.proto



<a name="axelar.multisig.v1beta1.RotateKeyRequest"></a>

### RotateKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="axelar.multisig.v1beta1.RotateKeyResponse"></a>

### RotateKeyResponse







<a name="axelar.multisig.v1beta1.StartKeygenRequest"></a>

### StartKeygenRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="axelar.multisig.v1beta1.StartKeygenResponse"></a>

### StartKeygenResponse







<a name="axelar.multisig.v1beta1.SubmitPubKeyRequest"></a>

### SubmitPubKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |
| `pub_key` | [bytes](#bytes) |  |  |
| `signature` | [bytes](#bytes) |  |  |






<a name="axelar.multisig.v1beta1.SubmitPubKeyResponse"></a>

### SubmitPubKeyResponse







<a name="axelar.multisig.v1beta1.SubmitSignatureRequest"></a>

### SubmitSignatureRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `sig_id` | [uint64](#uint64) |  |  |
| `signature` | [bytes](#bytes) |  |  |






<a name="axelar.multisig.v1beta1.SubmitSignatureResponse"></a>

### SubmitSignatureResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelar/multisig/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/multisig/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="axelar.multisig.v1beta1.MsgService"></a>

### MsgService
Msg defines the multisig Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `StartKeygen` | [StartKeygenRequest](#axelar.multisig.v1beta1.StartKeygenRequest) | [StartKeygenResponse](#axelar.multisig.v1beta1.StartKeygenResponse) |  | POST|/axelar/multisig/start_keygen|
| `SubmitPubKey` | [SubmitPubKeyRequest](#axelar.multisig.v1beta1.SubmitPubKeyRequest) | [SubmitPubKeyResponse](#axelar.multisig.v1beta1.SubmitPubKeyResponse) |  | POST|/axelar/multisig/submit_pub_key|
| `SubmitSignature` | [SubmitSignatureRequest](#axelar.multisig.v1beta1.SubmitSignatureRequest) | [SubmitSignatureResponse](#axelar.multisig.v1beta1.SubmitSignatureResponse) |  | POST|/axelar/multisig/submit_signature|
| `RotateKey` | [RotateKeyRequest](#axelar.multisig.v1beta1.RotateKeyRequest) | [RotateKeyResponse](#axelar.multisig.v1beta1.RotateKeyResponse) |  | POST|/axelar/multisig/rotate_key|


<a name="axelar.multisig.v1beta1.QueryService"></a>

### QueryService
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `KeyID` | [KeyIDRequest](#axelar.multisig.v1beta1.KeyIDRequest) | [KeyIDResponse](#axelar.multisig.v1beta1.KeyIDResponse) | KeyID returns the key ID of a key assigned to a given chain. If no key is assigned, it returns the grpc NOT_FOUND error. | GET|/axelar/multisig/v1beta1/key_id/{chain}|
| `NextKeyID` | [NextKeyIDRequest](#axelar.multisig.v1beta1.NextKeyIDRequest) | [NextKeyIDResponse](#axelar.multisig.v1beta1.NextKeyIDResponse) | NextKeyID returns the key ID assigned for the next rotation on a given chain. If no key rotation is in progress, it returns the grpc NOT_FOUND error. | GET|/axelar/multisig/v1beta1/next_key_id/{chain}|
| `Key` | [KeyRequest](#axelar.multisig.v1beta1.KeyRequest) | [KeyResponse](#axelar.multisig.v1beta1.KeyResponse) | Key returns the key corresponding to a given key ID. If no key is found, it returns the grpc NOT_FOUND error. | GET|/axelar/multisig/v1beta1/key|
| `KeygenSession` | [KeygenSessionRequest](#axelar.multisig.v1beta1.KeygenSessionRequest) | [KeygenSessionResponse](#axelar.multisig.v1beta1.KeygenSessionResponse) | KeygenSession returns the keygen session info for a given key ID. If no key is found, it returns the grpc NOT_FOUND error. | GET|/axelar/multisig/v1beta1/keygen_session|

 <!-- end services -->



<a name="axelar/nexus/v1beta1/events.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/nexus/v1beta1/events.proto



<a name="axelar.nexus.v1beta1.FeeDeducted"></a>

### FeeDeducted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `transfer_id` | [uint64](#uint64) |  |  |
| `recipient_chain` | [string](#string) |  |  |
| `recipient_address` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="axelar.nexus.v1beta1.InsufficientFee"></a>

### InsufficientFee



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `transfer_id` | [uint64](#uint64) |  |  |
| `recipient_chain` | [string](#string) |  |  |
| `recipient_address` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

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
| `RecipientAddress` | [RecipientAddressRequest](#axelar.nexus.v1beta1.RecipientAddressRequest) | [RecipientAddressResponse](#axelar.nexus.v1beta1.RecipientAddressResponse) | RecipientAddress queries the recipient address for a given deposit address | GET|/axelar/nexus/v1beta1/recipient_address/{deposit_chain}/{deposit_addr}|

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
| `key_mgmt_relative_inflation_rate` | [bytes](#bytes) |  |  |





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





 <!-- end messages -->

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
| `HeartBeat` | [HeartBeatRequest](#axelar.tss.v1beta1.HeartBeatRequest) | [HeartBeatResponse](#axelar.tss.v1beta1.HeartBeatResponse) |  | POST|/axelar/tss/heartbeat|


<a name="axelar.tss.v1beta1.QueryService"></a>

### QueryService
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |

 <!-- end services -->



<a name="axelar/vote/v1beta1/events.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelar/vote/v1beta1/events.proto



<a name="axelar.vote.v1beta1.Voted"></a>

### Voted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  |  |
| `action` | [string](#string) |  |  |
| `poll` | [string](#string) |  |  |
| `voter` | [string](#string) |  |  |
| `state` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

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
| `end_blocker_limit` | [int64](#int64) |  |  |





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
| `data` | [google.protobuf.Any](#google.protobuf.Any) |  |  |
| `poll_id` | [uint64](#uint64) |  |  |
| `is_voter_late` | [TalliedVote.IsVoterLateEntry](#axelar.vote.v1beta1.TalliedVote.IsVoterLateEntry) | repeated |  |






<a name="axelar.vote.v1beta1.TalliedVote.IsVoterLateEntry"></a>

### TalliedVote.IsVoterLateEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [bool](#bool) |  |  |





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
| `poll_id` | [uint64](#uint64) |  |  |
| `vote` | [google.protobuf.Any](#google.protobuf.Any) |  |  |






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

