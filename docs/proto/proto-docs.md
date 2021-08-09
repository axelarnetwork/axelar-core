<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [axelarnet/v1beta1/params.proto](#axelarnet/v1beta1/params.proto)
    - [Params](#axelarnet.v1beta1.Params)
  
- [axelarnet/v1beta1/genesis.proto](#axelarnet/v1beta1/genesis.proto)
    - [GenesisState](#axelarnet.v1beta1.GenesisState)
  
- [axelarnet/v1beta1/tx.proto](#axelarnet/v1beta1/tx.proto)
    - [AddCosmosBasedChainRequest](#axelarnet.v1beta1.AddCosmosBasedChainRequest)
    - [AddCosmosBasedChainResponse](#axelarnet.v1beta1.AddCosmosBasedChainResponse)
    - [ConfirmDepositRequest](#axelarnet.v1beta1.ConfirmDepositRequest)
    - [ConfirmDepositResponse](#axelarnet.v1beta1.ConfirmDepositResponse)
    - [ExecutePendingTransfersRequest](#axelarnet.v1beta1.ExecutePendingTransfersRequest)
    - [ExecutePendingTransfersResponse](#axelarnet.v1beta1.ExecutePendingTransfersResponse)
    - [LinkRequest](#axelarnet.v1beta1.LinkRequest)
    - [LinkResponse](#axelarnet.v1beta1.LinkResponse)
    - [RegisterIbcPathRequest](#axelarnet.v1beta1.RegisterIbcPathRequest)
    - [RegisterIbcPathResponse](#axelarnet.v1beta1.RegisterIbcPathResponse)
  
- [axelarnet/v1beta1/service.proto](#axelarnet/v1beta1/service.proto)
    - [MsgService](#axelarnet.v1beta1.MsgService)
  
- [tss/exported/v1beta1/types.proto](#tss/exported/v1beta1/types.proto)
    - [KeyRequirement](#tss.exported.v1beta1.KeyRequirement)
  
    - [KeyRole](#tss.exported.v1beta1.KeyRole)
    - [KeyShareDistributionPolicy](#tss.exported.v1beta1.KeyShareDistributionPolicy)
  
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
    - [TxStatus](#bitcoin.v1beta1.TxStatus)
  
- [utils/v1beta1/threshold.proto](#utils/v1beta1/threshold.proto)
    - [Threshold](#utils.v1beta1.Threshold)
  
- [bitcoin/v1beta1/params.proto](#bitcoin/v1beta1/params.proto)
    - [Params](#bitcoin.v1beta1.Params)
  
- [bitcoin/v1beta1/genesis.proto](#bitcoin/v1beta1/genesis.proto)
    - [GenesisState](#bitcoin.v1beta1.GenesisState)
  
- [bitcoin/v1beta1/query.proto](#bitcoin/v1beta1/query.proto)
    - [DepositQueryParams](#bitcoin.v1beta1.DepositQueryParams)
    - [QueryAddressResponse](#bitcoin.v1beta1.QueryAddressResponse)
    - [QueryTxResponse](#bitcoin.v1beta1.QueryTxResponse)
    - [QueryTxResponse.SigningInfo](#bitcoin.v1beta1.QueryTxResponse.SigningInfo)
  
- [vote/exported/v1beta1/types.proto](#vote/exported/v1beta1/types.proto)
    - [PollKey](#vote.exported.v1beta1.PollKey)
    - [PollMetadata](#vote.exported.v1beta1.PollMetadata)
  
    - [PollState](#vote.exported.v1beta1.PollState)
  
- [bitcoin/v1beta1/tx.proto](#bitcoin/v1beta1/tx.proto)
    - [ConfirmOutpointRequest](#bitcoin.v1beta1.ConfirmOutpointRequest)
    - [ConfirmOutpointResponse](#bitcoin.v1beta1.ConfirmOutpointResponse)
    - [CreateMasterTxRequest](#bitcoin.v1beta1.CreateMasterTxRequest)
    - [CreateMasterTxResponse](#bitcoin.v1beta1.CreateMasterTxResponse)
    - [CreatePendingTransfersTxRequest](#bitcoin.v1beta1.CreatePendingTransfersTxRequest)
    - [CreatePendingTransfersTxResponse](#bitcoin.v1beta1.CreatePendingTransfersTxResponse)
    - [LinkRequest](#bitcoin.v1beta1.LinkRequest)
    - [LinkResponse](#bitcoin.v1beta1.LinkResponse)
    - [RegisterExternalKeysRequest](#bitcoin.v1beta1.RegisterExternalKeysRequest)
    - [RegisterExternalKeysRequest.ExternalKey](#bitcoin.v1beta1.RegisterExternalKeysRequest.ExternalKey)
    - [RegisterExternalKeysResponse](#bitcoin.v1beta1.RegisterExternalKeysResponse)
    - [SignTxRequest](#bitcoin.v1beta1.SignTxRequest)
    - [SignTxResponse](#bitcoin.v1beta1.SignTxResponse)
    - [SubmitExternalSignatureRequest](#bitcoin.v1beta1.SubmitExternalSignatureRequest)
    - [SubmitExternalSignatureResponse](#bitcoin.v1beta1.SubmitExternalSignatureResponse)
    - [VoteConfirmOutpointRequest](#bitcoin.v1beta1.VoteConfirmOutpointRequest)
    - [VoteConfirmOutpointResponse](#bitcoin.v1beta1.VoteConfirmOutpointResponse)
  
- [bitcoin/v1beta1/service.proto](#bitcoin/v1beta1/service.proto)
    - [MsgService](#bitcoin.v1beta1.MsgService)
  
- [evm/v1beta1/types.proto](#evm/v1beta1/types.proto)
    - [BurnerInfo](#evm.v1beta1.BurnerInfo)
    - [ERC20Deposit](#evm.v1beta1.ERC20Deposit)
    - [ERC20TokenDeployment](#evm.v1beta1.ERC20TokenDeployment)
    - [NetworkInfo](#evm.v1beta1.NetworkInfo)
    - [TransferOwnership](#evm.v1beta1.TransferOwnership)
  
- [evm/v1beta1/params.proto](#evm/v1beta1/params.proto)
    - [Params](#evm.v1beta1.Params)
  
- [evm/v1beta1/genesis.proto](#evm/v1beta1/genesis.proto)
    - [GenesisState](#evm.v1beta1.GenesisState)
  
- [evm/v1beta1/query.proto](#evm/v1beta1/query.proto)
    - [DepositQueryParams](#evm.v1beta1.DepositQueryParams)
    - [QueryMasterAddressResponse](#evm.v1beta1.QueryMasterAddressResponse)
  
- [evm/v1beta1/tx.proto](#evm/v1beta1/tx.proto)
    - [AddChainRequest](#evm.v1beta1.AddChainRequest)
    - [AddChainResponse](#evm.v1beta1.AddChainResponse)
    - [ConfirmChainRequest](#evm.v1beta1.ConfirmChainRequest)
    - [ConfirmChainResponse](#evm.v1beta1.ConfirmChainResponse)
    - [ConfirmDepositRequest](#evm.v1beta1.ConfirmDepositRequest)
    - [ConfirmDepositResponse](#evm.v1beta1.ConfirmDepositResponse)
    - [ConfirmTokenRequest](#evm.v1beta1.ConfirmTokenRequest)
    - [ConfirmTokenResponse](#evm.v1beta1.ConfirmTokenResponse)
    - [ConfirmTransferOwnershipRequest](#evm.v1beta1.ConfirmTransferOwnershipRequest)
    - [ConfirmTransferOwnershipResponse](#evm.v1beta1.ConfirmTransferOwnershipResponse)
    - [LinkRequest](#evm.v1beta1.LinkRequest)
    - [LinkResponse](#evm.v1beta1.LinkResponse)
    - [SignBurnTokensRequest](#evm.v1beta1.SignBurnTokensRequest)
    - [SignBurnTokensResponse](#evm.v1beta1.SignBurnTokensResponse)
    - [SignDeployTokenRequest](#evm.v1beta1.SignDeployTokenRequest)
    - [SignDeployTokenResponse](#evm.v1beta1.SignDeployTokenResponse)
    - [SignPendingTransfersRequest](#evm.v1beta1.SignPendingTransfersRequest)
    - [SignPendingTransfersResponse](#evm.v1beta1.SignPendingTransfersResponse)
    - [SignTransferOwnershipRequest](#evm.v1beta1.SignTransferOwnershipRequest)
    - [SignTransferOwnershipResponse](#evm.v1beta1.SignTransferOwnershipResponse)
    - [SignTxRequest](#evm.v1beta1.SignTxRequest)
    - [SignTxResponse](#evm.v1beta1.SignTxResponse)
    - [VoteConfirmChainRequest](#evm.v1beta1.VoteConfirmChainRequest)
    - [VoteConfirmChainResponse](#evm.v1beta1.VoteConfirmChainResponse)
    - [VoteConfirmDepositRequest](#evm.v1beta1.VoteConfirmDepositRequest)
    - [VoteConfirmDepositResponse](#evm.v1beta1.VoteConfirmDepositResponse)
    - [VoteConfirmTokenRequest](#evm.v1beta1.VoteConfirmTokenRequest)
    - [VoteConfirmTokenResponse](#evm.v1beta1.VoteConfirmTokenResponse)
    - [VoteConfirmTransferOwnershipRequest](#evm.v1beta1.VoteConfirmTransferOwnershipRequest)
    - [VoteConfirmTransferOwnershipResponse](#evm.v1beta1.VoteConfirmTransferOwnershipResponse)
  
- [evm/v1beta1/service.proto](#evm/v1beta1/service.proto)
    - [MsgService](#evm.v1beta1.MsgService)
  
- [nexus/exported/v1beta1/types.proto](#nexus/exported/v1beta1/types.proto)
    - [Chain](#nexus.exported.v1beta1.Chain)
    - [CrossChainAddress](#nexus.exported.v1beta1.CrossChainAddress)
    - [CrossChainTransfer](#nexus.exported.v1beta1.CrossChainTransfer)
  
    - [TransferState](#nexus.exported.v1beta1.TransferState)
  
- [nexus/v1beta1/params.proto](#nexus/v1beta1/params.proto)
    - [Params](#nexus.v1beta1.Params)
  
- [nexus/v1beta1/genesis.proto](#nexus/v1beta1/genesis.proto)
    - [GenesisState](#nexus.v1beta1.GenesisState)
  
- [snapshot/exported/v1beta1/types.proto](#snapshot/exported/v1beta1/types.proto)
    - [Snapshot](#snapshot.exported.v1beta1.Snapshot)
    - [Validator](#snapshot.exported.v1beta1.Validator)
  
- [snapshot/v1beta1/params.proto](#snapshot/v1beta1/params.proto)
    - [Params](#snapshot.v1beta1.Params)
  
- [snapshot/v1beta1/genesis.proto](#snapshot/v1beta1/genesis.proto)
    - [GenesisState](#snapshot.v1beta1.GenesisState)
  
- [snapshot/v1beta1/tx.proto](#snapshot/v1beta1/tx.proto)
    - [DeactivateProxyRequest](#snapshot.v1beta1.DeactivateProxyRequest)
    - [DeactivateProxyResponse](#snapshot.v1beta1.DeactivateProxyResponse)
    - [RegisterProxyRequest](#snapshot.v1beta1.RegisterProxyRequest)
    - [RegisterProxyResponse](#snapshot.v1beta1.RegisterProxyResponse)
  
- [snapshot/v1beta1/service.proto](#snapshot/v1beta1/service.proto)
    - [MsgService](#snapshot.v1beta1.MsgService)
  
- [tss/tofnd/v1beta1/tofnd.proto](#tss/tofnd/v1beta1/tofnd.proto)
    - [KeyPresenceRequest](#tss.tofnd.v1beta1.KeyPresenceRequest)
    - [KeyPresenceResponse](#tss.tofnd.v1beta1.KeyPresenceResponse)
    - [KeygenInit](#tss.tofnd.v1beta1.KeygenInit)
    - [MessageIn](#tss.tofnd.v1beta1.MessageIn)
    - [MessageOut](#tss.tofnd.v1beta1.MessageOut)
    - [MessageOut.CriminalList](#tss.tofnd.v1beta1.MessageOut.CriminalList)
    - [MessageOut.CriminalList.Criminal](#tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal)
    - [MessageOut.KeygenResult](#tss.tofnd.v1beta1.MessageOut.KeygenResult)
    - [MessageOut.KeygenResult.KeygenOutput](#tss.tofnd.v1beta1.MessageOut.KeygenResult.KeygenOutput)
    - [MessageOut.SignResult](#tss.tofnd.v1beta1.MessageOut.SignResult)
    - [RecoverRequest](#tss.tofnd.v1beta1.RecoverRequest)
    - [RecoverResponse](#tss.tofnd.v1beta1.RecoverResponse)
    - [SignInit](#tss.tofnd.v1beta1.SignInit)
    - [TrafficIn](#tss.tofnd.v1beta1.TrafficIn)
    - [TrafficOut](#tss.tofnd.v1beta1.TrafficOut)
  
    - [KeyPresenceResponse.Response](#tss.tofnd.v1beta1.KeyPresenceResponse.Response)
    - [MessageOut.CriminalList.Criminal.CrimeType](#tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal.CrimeType)
    - [RecoverResponse.Response](#tss.tofnd.v1beta1.RecoverResponse.Response)
  
- [tss/v1beta1/params.proto](#tss/v1beta1/params.proto)
    - [Params](#tss.v1beta1.Params)
  
- [tss/v1beta1/genesis.proto](#tss/v1beta1/genesis.proto)
    - [GenesisState](#tss.v1beta1.GenesisState)
  
- [tss/v1beta1/query.proto](#tss/v1beta1/query.proto)
    - [QueryKeyResponse](#tss.v1beta1.QueryKeyResponse)
    - [QueryRecoveryResponse](#tss.v1beta1.QueryRecoveryResponse)
    - [QuerySigResponse](#tss.v1beta1.QuerySigResponse)
    - [Signature](#tss.v1beta1.Signature)
  
    - [VoteStatus](#tss.v1beta1.VoteStatus)
  
- [tss/v1beta1/tx.proto](#tss/v1beta1/tx.proto)
    - [ProcessKeygenTrafficRequest](#tss.v1beta1.ProcessKeygenTrafficRequest)
    - [ProcessKeygenTrafficResponse](#tss.v1beta1.ProcessKeygenTrafficResponse)
    - [ProcessSignTrafficRequest](#tss.v1beta1.ProcessSignTrafficRequest)
    - [ProcessSignTrafficResponse](#tss.v1beta1.ProcessSignTrafficResponse)
    - [RotateKeyRequest](#tss.v1beta1.RotateKeyRequest)
    - [RotateKeyResponse](#tss.v1beta1.RotateKeyResponse)
    - [StartKeygenRequest](#tss.v1beta1.StartKeygenRequest)
    - [StartKeygenResponse](#tss.v1beta1.StartKeygenResponse)
    - [VotePubKeyRequest](#tss.v1beta1.VotePubKeyRequest)
    - [VotePubKeyResponse](#tss.v1beta1.VotePubKeyResponse)
    - [VoteSigRequest](#tss.v1beta1.VoteSigRequest)
    - [VoteSigResponse](#tss.v1beta1.VoteSigResponse)
  
- [tss/v1beta1/service.proto](#tss/v1beta1/service.proto)
    - [MsgService](#tss.v1beta1.MsgService)
  
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
| `supported_chains` | [string](#string) | repeated |  |





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





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="axelarnet/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## axelarnet/v1beta1/tx.proto



<a name="axelarnet.v1beta1.AddCosmosBasedChainRequest"></a>

### AddCosmosBasedChainRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `name` | [string](#string) |  |  |
| `native_asset` | [string](#string) |  |  |






<a name="axelarnet.v1beta1.AddCosmosBasedChainResponse"></a>

### AddCosmosBasedChainResponse







<a name="axelarnet.v1beta1.ConfirmDepositRequest"></a>

### ConfirmDepositRequest
MsgConfirmDeposit represents a deposit confirmation message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `token` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `burner_address` | [bytes](#bytes) |  |  |






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






<a name="axelarnet.v1beta1.RegisterIbcPathRequest"></a>

### RegisterIbcPathRequest
RegisterIbcPathRequest represents a message to register a path for an asset


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `asset` | [string](#string) |  |  |
| `path` | [string](#string) |  |  |






<a name="axelarnet.v1beta1.RegisterIbcPathResponse"></a>

### RegisterIbcPathResponse






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
| `RegisterIbcPath` | [RegisterIbcPathRequest](#axelarnet.v1beta1.RegisterIbcPathRequest) | [RegisterIbcPathResponse](#axelarnet.v1beta1.RegisterIbcPathResponse) |  | POST|/axelar/axelarnet/register-ibc-path|
| `AddCosmosBasedChain` | [AddCosmosBasedChainRequest](#axelarnet.v1beta1.AddCosmosBasedChainRequest) | [AddCosmosBasedChainResponse](#axelarnet.v1beta1.AddCosmosBasedChainResponse) |  | POST|/axelar/axelarnet/add-cosmos-based-chain|

 <!-- end services -->



<a name="tss/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## tss/exported/v1beta1/types.proto



<a name="tss.exported.v1beta1.KeyRequirement"></a>

### KeyRequirement
KeyRequirement defines requirements for keys


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain_name` | [string](#string) |  |  |
| `key_role` | [KeyRole](#tss.exported.v1beta1.KeyRole) |  |  |
| `min_validator_subset_size` | [int64](#int64) |  |  |
| `key_share_distribution_policy` | [KeyShareDistributionPolicy](#tss.exported.v1beta1.KeyShareDistributionPolicy) |  |  |





 <!-- end messages -->


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


 <!-- end enums -->

 <!-- end HasExtensions -->

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
| `tx` | [bytes](#bytes) |  |  |
| `prev_signed_tx_hash` | [bytes](#bytes) |  |  |
| `confirmation_required` | [bool](#bool) |  |  |
| `anyone_can_spend_vout` | [uint32](#uint32) |  |  |






<a name="bitcoin.v1beta1.UnsignedTx"></a>

### UnsignedTx



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx` | [bytes](#bytes) |  |  |
| `info` | [UnsignedTx.Info](#bitcoin.v1beta1.UnsignedTx.Info) |  |  |
| `status` | [TxStatus](#bitcoin.v1beta1.TxStatus) |  |  |
| `confirmation_required` | [bool](#bool) |  |  |
| `anyone_can_spend_vout` | [uint32](#uint32) |  |  |
| `prev_aborted_key_id` | [string](#string) |  |  |






<a name="bitcoin.v1beta1.UnsignedTx.Info"></a>

### UnsignedTx.Info



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `assign_next_key` | [bool](#bool) |  |  |
| `next_key_id` | [string](#string) |  |  |
| `input_infos` | [UnsignedTx.Info.InputInfo](#bitcoin.v1beta1.UnsignedTx.Info.InputInfo) | repeated |  |






<a name="bitcoin.v1beta1.UnsignedTx.Info.InputInfo"></a>

### UnsignedTx.Info.InputInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `out_point_info` | [OutPointInfo](#bitcoin.v1beta1.OutPointInfo) |  |  |
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



<a name="bitcoin.v1beta1.TxStatus"></a>

### TxStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| TX_STATUS_UNSPECIFIED | 0 |  |
| TX_STATUS_CREATED | 1 |  |
| TX_STATUS_SIGNING | 2 |  |
| TX_STATUS_ABORTED | 3 |  |
| TX_STATUS_SIGNED | 4 |  |


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
| `master_address_lock_duration` | [int64](#int64) |  |  |
| `external_multisig_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  |  |





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
| `snapshot_seq_no` | [int64](#int64) |  |  |
| `expires_at` | [int64](#int64) |  |  |
| `result` | [google.protobuf.Any](#google.protobuf.Any) |  |  |
| `voting_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  |  |
| `state` | [PollState](#vote.exported.v1beta1.PollState) |  |  |





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






<a name="bitcoin.v1beta1.RegisterExternalKeysRequest"></a>

### RegisterExternalKeysRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `external_keys` | [RegisterExternalKeysRequest.ExternalKey](#bitcoin.v1beta1.RegisterExternalKeysRequest.ExternalKey) | repeated |  |






<a name="bitcoin.v1beta1.RegisterExternalKeysRequest.ExternalKey"></a>

### RegisterExternalKeysRequest.ExternalKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `pub_key` | [bytes](#bytes) |  |  |






<a name="bitcoin.v1beta1.RegisterExternalKeysResponse"></a>

### RegisterExternalKeysResponse







<a name="bitcoin.v1beta1.SignTxRequest"></a>

### SignTxRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `key_role` | [tss.exported.v1beta1.KeyRole](#tss.exported.v1beta1.KeyRole) |  |  |






<a name="bitcoin.v1beta1.SignTxResponse"></a>

### SignTxResponse







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
| `SignTx` | [SignTxRequest](#bitcoin.v1beta1.SignTxRequest) | [SignTxResponse](#bitcoin.v1beta1.SignTxResponse) |  | POST|/axelar/bitcoin/sign-tx|
| `RegisterExternalKeys` | [RegisterExternalKeysRequest](#bitcoin.v1beta1.RegisterExternalKeysRequest) | [RegisterExternalKeysResponse](#bitcoin.v1beta1.RegisterExternalKeysResponse) |  | POST|/axelar/bitcoin/register-external-key|
| `SubmitExternalSignature` | [SubmitExternalSignatureRequest](#bitcoin.v1beta1.SubmitExternalSignatureRequest) | [SubmitExternalSignatureResponse](#bitcoin.v1beta1.SubmitExternalSignatureResponse) |  | POST|/axelar/bitcoin/submit-external-signature|

 <!-- end services -->



<a name="evm/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evm/v1beta1/types.proto



<a name="evm.v1beta1.BurnerInfo"></a>

### BurnerInfo
BurnerInfo describes information required to burn token at an burner address
that is deposited by an user


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `token_address` | [bytes](#bytes) |  |  |
| `destination_chain` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `asset` | [string](#string) |  |  |
| `salt` | [bytes](#bytes) |  |  |






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






<a name="evm.v1beta1.ERC20TokenDeployment"></a>

### ERC20TokenDeployment
ERC20TokenDeployment describes information about an ERC20 token


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset` | [string](#string) |  |  |
| `token_address` | [string](#string) |  |  |






<a name="evm.v1beta1.NetworkInfo"></a>

### NetworkInfo
NetworkInfo describes information about a network


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `id` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.TransferOwnership"></a>

### TransferOwnership
TransferOwnership contains information for a transfer ownership


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `next_key_id` | [string](#string) |  |  |





 <!-- end messages -->

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
| `gateway` | [bytes](#bytes) |  |  |
| `token` | [bytes](#bytes) |  |  |
| `burnable` | [bytes](#bytes) |  |  |
| `revote_locking_period` | [int64](#int64) |  |  |
| `networks` | [NetworkInfo](#evm.v1beta1.NetworkInfo) | repeated |  |





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
| `params` | [Params](#evm.v1beta1.Params) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="evm/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evm/v1beta1/query.proto



<a name="evm.v1beta1.DepositQueryParams"></a>

### DepositQueryParams
DepositQueryParams describe the parameters used to query for an EVM
deposit address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `chain` | [string](#string) |  |  |






<a name="evm.v1beta1.QueryMasterAddressResponse"></a>

### QueryMasterAddressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |





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
| `native_asset` | [string](#string) |  |  |
| `key_requirement` | [tss.exported.v1beta1.KeyRequirement](#tss.exported.v1beta1.KeyRequirement) |  |  |
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







<a name="evm.v1beta1.ConfirmTokenRequest"></a>

### ConfirmTokenRequest
MsgConfirmToken represents a token deploy confirmation message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `origin_chain` | [string](#string) |  |  |






<a name="evm.v1beta1.ConfirmTokenResponse"></a>

### ConfirmTokenResponse







<a name="evm.v1beta1.ConfirmTransferOwnershipRequest"></a>

### ConfirmTransferOwnershipRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="evm.v1beta1.ConfirmTransferOwnershipResponse"></a>

### ConfirmTransferOwnershipResponse







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






<a name="evm.v1beta1.SignBurnTokensRequest"></a>

### SignBurnTokensRequest
MsgSignBurnTokens represents the message to sign commands to burn tokens with
AxelarGateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |






<a name="evm.v1beta1.SignBurnTokensResponse"></a>

### SignBurnTokensResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `command_id` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.SignDeployTokenRequest"></a>

### SignDeployTokenRequest
MsgSignDeployToken represents the message to sign a deploy token command for
AxelarGateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `origin_chain` | [string](#string) |  |  |
| `capacity` | [bytes](#bytes) |  |  |
| `decimals` | [uint32](#uint32) |  |  |
| `symbol` | [string](#string) |  |  |
| `token_name` | [string](#string) |  |  |






<a name="evm.v1beta1.SignDeployTokenResponse"></a>

### SignDeployTokenResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `command_id` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.SignPendingTransfersRequest"></a>

### SignPendingTransfersRequest
MsgSignPendingTransfers represents a message to trigger the signing of all
pending transfers


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |






<a name="evm.v1beta1.SignPendingTransfersResponse"></a>

### SignPendingTransfersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `command_id` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.SignTransferOwnershipRequest"></a>

### SignTransferOwnershipRequest
MsgSignDeployToken represents the message to sign a deploy token command for
AxelarGateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |






<a name="evm.v1beta1.SignTransferOwnershipResponse"></a>

### SignTransferOwnershipResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `command_id` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.SignTxRequest"></a>

### SignTxRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  | Tx is stored in serialized form because the amino codec cannot properly deserialize MsgSignTx otherwise |
| `tx` | [bytes](#bytes) |  |  |






<a name="evm.v1beta1.SignTxResponse"></a>

### SignTxResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [string](#string) |  |  |






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






<a name="evm.v1beta1.VoteConfirmTokenRequest"></a>

### VoteConfirmTokenRequest
MsgVoteConfirmToken represents a message that votes on a token deploy


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






<a name="evm.v1beta1.VoteConfirmTransferOwnershipRequest"></a>

### VoteConfirmTransferOwnershipRequest
MsgVoteConfirmDeposit represents a message that votes on a deposit


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `poll_key` | [vote.exported.v1beta1.PollKey](#vote.exported.v1beta1.PollKey) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `new_owner_address` | [bytes](#bytes) |  |  |
| `confirmed` | [bool](#bool) |  |  |






<a name="evm.v1beta1.VoteConfirmTransferOwnershipResponse"></a>

### VoteConfirmTransferOwnershipResponse



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
| `ConfirmToken` | [ConfirmTokenRequest](#evm.v1beta1.ConfirmTokenRequest) | [ConfirmTokenResponse](#evm.v1beta1.ConfirmTokenResponse) |  | POST|/axelar/evm/confirm-erc20-deploy|
| `ConfirmDeposit` | [ConfirmDepositRequest](#evm.v1beta1.ConfirmDepositRequest) | [ConfirmDepositResponse](#evm.v1beta1.ConfirmDepositResponse) |  | POST|/axelar/evm/confirm-erc20-deposit|
| `ConfirmTransferOwnership` | [ConfirmTransferOwnershipRequest](#evm.v1beta1.ConfirmTransferOwnershipRequest) | [ConfirmTransferOwnershipResponse](#evm.v1beta1.ConfirmTransferOwnershipResponse) |  | POST|/axelar/evm/confirm-transfer-ownership|
| `VoteConfirmChain` | [VoteConfirmChainRequest](#evm.v1beta1.VoteConfirmChainRequest) | [VoteConfirmChainResponse](#evm.v1beta1.VoteConfirmChainResponse) |  | ||
| `VoteConfirmDeposit` | [VoteConfirmDepositRequest](#evm.v1beta1.VoteConfirmDepositRequest) | [VoteConfirmDepositResponse](#evm.v1beta1.VoteConfirmDepositResponse) |  | ||
| `VoteConfirmToken` | [VoteConfirmTokenRequest](#evm.v1beta1.VoteConfirmTokenRequest) | [VoteConfirmTokenResponse](#evm.v1beta1.VoteConfirmTokenResponse) |  | ||
| `VoteConfirmTransferOwnership` | [VoteConfirmTransferOwnershipRequest](#evm.v1beta1.VoteConfirmTransferOwnershipRequest) | [VoteConfirmTransferOwnershipResponse](#evm.v1beta1.VoteConfirmTransferOwnershipResponse) |  | ||
| `SignDeployToken` | [SignDeployTokenRequest](#evm.v1beta1.SignDeployTokenRequest) | [SignDeployTokenResponse](#evm.v1beta1.SignDeployTokenResponse) |  | POST|/axelar/evm/sign-deploy-token|
| `SignBurnTokens` | [SignBurnTokensRequest](#evm.v1beta1.SignBurnTokensRequest) | [SignBurnTokensResponse](#evm.v1beta1.SignBurnTokensResponse) |  | POST|/axelar/evm/sign-burn|
| `SignTx` | [SignTxRequest](#evm.v1beta1.SignTxRequest) | [SignTxResponse](#evm.v1beta1.SignTxResponse) |  | POST|/axelar/evm/sign-tx|
| `SignPendingTransfers` | [SignPendingTransfersRequest](#evm.v1beta1.SignPendingTransfersRequest) | [SignPendingTransfersResponse](#evm.v1beta1.SignPendingTransfersResponse) |  | POST|/axelar/evm/sign-pending|
| `SignTransferOwnership` | [SignTransferOwnershipRequest](#evm.v1beta1.SignTransferOwnershipRequest) | [SignTransferOwnershipResponse](#evm.v1beta1.SignTransferOwnershipResponse) |  | ||
| `AddChain` | [AddChainRequest](#evm.v1beta1.AddChainRequest) | [AddChainResponse](#evm.v1beta1.AddChainResponse) |  | POST|/axelar/evm/add-chain|

 <!-- end services -->



<a name="nexus/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## nexus/exported/v1beta1/types.proto



<a name="nexus.exported.v1beta1.Chain"></a>

### Chain
Chain represents the properties of a registered blockchain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `native_asset` | [string](#string) |  |  |
| `supports_foreign_assets` | [bool](#bool) |  |  |






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



<a name="nexus/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## nexus/v1beta1/params.proto



<a name="nexus.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chains` | [nexus.exported.v1beta1.Chain](#nexus.exported.v1beta1.Chain) | repeated |  |





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






<a name="snapshot.exported.v1beta1.Validator"></a>

### Validator



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sdk_validator` | [google.protobuf.Any](#google.protobuf.Any) |  |  |
| `share_count` | [int64](#int64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="snapshot/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## snapshot/v1beta1/params.proto



<a name="snapshot.v1beta1.Params"></a>

### Params
Params represent the genesis parameters for the module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locking_period` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |





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
| `principal_addr` | [bytes](#bytes) |  |  |






<a name="snapshot.v1beta1.DeactivateProxyResponse"></a>

### DeactivateProxyResponse







<a name="snapshot.v1beta1.RegisterProxyRequest"></a>

### RegisterProxyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `principal_addr` | [bytes](#bytes) |  |  |
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



<a name="tss/tofnd/v1beta1/tofnd.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## tss/tofnd/v1beta1/tofnd.proto
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






<a name="tss.tofnd.v1beta1.KeygenInit"></a>

### KeygenInit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `new_key_uid` | [string](#string) |  |  |
| `party_uids` | [string](#string) | repeated |  |
| `party_share_counts` | [uint32](#uint32) | repeated |  |
| `my_party_index` | [int32](#int32) |  | parties[my_party_index] belongs to the server |
| `threshold` | [int32](#int32) |  |  |






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
| `data` | [MessageOut.KeygenResult.KeygenOutput](#tss.tofnd.v1beta1.MessageOut.KeygenResult.KeygenOutput) |  | Success response |
| `criminals` | [MessageOut.CriminalList](#tss.tofnd.v1beta1.MessageOut.CriminalList) |  | Faiilure response |






<a name="tss.tofnd.v1beta1.MessageOut.KeygenResult.KeygenOutput"></a>

### MessageOut.KeygenResult.KeygenOutput
Keygen's success response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pub_key` | [bytes](#bytes) |  | pub_key |
| `share_recovery_infos` | [bytes](#bytes) | repeated | recovery info |






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
| `share_recovery_infos` | [bytes](#bytes) | repeated |  |






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


<a name="tss.tofnd.v1beta1.KeyPresenceResponse.Response"></a>

### KeyPresenceResponse.Response


| Name | Number | Description |
| ---- | ------ | ----------- |
| RESPONSE_UNSPECIFIED | 0 |  |
| RESPONSE_PRESENT | 1 |  |
| RESPONSE_ABSENT | 2 |  |
| RESPONSE_FAIL | 3 |  |



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
| `locking_period` | [int64](#int64) |  | **Deprecated.**  |
| `min_keygen_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  | MinKeygenThreshold defines the minimum % of stake that must be online to authorize generation of a new key in the system. |
| `corruption_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  | CorruptionThreshold defines the corruption threshold with which we'll run keygen protocol. |
| `key_requirements` | [tss.exported.v1beta1.KeyRequirement](#tss.exported.v1beta1.KeyRequirement) | repeated | KeyRequirements defines the requirement of each key for each chain |
| `min_bond_fraction_per_share` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  | MinBondFractionPerShare defines the % of stake validators have to bond per key share |
| `suspend_duration_in_blocks` | [int64](#int64) |  | SuspendDurationInBlocks defines the number of blocks a validator is disallowed to participate in any TSS ceremony after committing a malicious behaviour during signing |
| `timeout_in_blocks` | [int64](#int64) |  | TimeoutInBlocks defines the timeout in blocks for signing and keygen |





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





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="tss/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## tss/v1beta1/query.proto



<a name="tss.v1beta1.QueryKeyResponse"></a>

### QueryKeyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `vote_status` | [VoteStatus](#tss.v1beta1.VoteStatus) |  |  |
| `role` | [tss.exported.v1beta1.KeyRole](#tss.exported.v1beta1.KeyRole) |  |  |






<a name="tss.v1beta1.QueryRecoveryResponse"></a>

### QueryRecoveryResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `party_uids` | [string](#string) | repeated |  |
| `party_share_counts` | [uint32](#uint32) | repeated |  |
| `threshold` | [int32](#int32) |  |  |
| `share_recovery_infos` | [bytes](#bytes) | repeated |  |






<a name="tss.v1beta1.QuerySigResponse"></a>

### QuerySigResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `vote_status` | [VoteStatus](#tss.v1beta1.VoteStatus) |  |  |
| `signature` | [Signature](#tss.v1beta1.Signature) |  |  |






<a name="tss.v1beta1.Signature"></a>

### Signature



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `r` | [bytes](#bytes) |  |  |
| `s` | [bytes](#bytes) |  |  |





 <!-- end messages -->


<a name="tss.v1beta1.VoteStatus"></a>

### VoteStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| VOTE_STATUS_UNSPECIFIED | 0 |  |
| VOTE_STATUS_PENDING | 1 |  |
| VOTE_STATUS_DECIDED | 2 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="tss/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## tss/v1beta1/tx.proto



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
| `new_key_id` | [string](#string) |  |  |
| `subset_size` | [int64](#int64) |  |  |
| `key_share_distribution_policy` | [tss.exported.v1beta1.KeyShareDistributionPolicy](#tss.exported.v1beta1.KeyShareDistributionPolicy) |  |  |






<a name="tss.v1beta1.StartKeygenResponse"></a>

### StartKeygenResponse







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
| `StartKeygen` | [StartKeygenRequest](#tss.v1beta1.StartKeygenRequest) | [StartKeygenResponse](#tss.v1beta1.StartKeygenResponse) |  | POST|/axelar/tss/startKeygen|
| `ProcessKeygenTraffic` | [ProcessKeygenTrafficRequest](#tss.v1beta1.ProcessKeygenTrafficRequest) | [ProcessKeygenTrafficResponse](#tss.v1beta1.ProcessKeygenTrafficResponse) |  | ||
| `RotateKey` | [RotateKeyRequest](#tss.v1beta1.RotateKeyRequest) | [RotateKeyResponse](#tss.v1beta1.RotateKeyResponse) |  | POST|/axelar/tss/assign/{chain}|
| `VotePubKey` | [VotePubKeyRequest](#tss.v1beta1.VotePubKeyRequest) | [VotePubKeyResponse](#tss.v1beta1.VotePubKeyResponse) |  | ||
| `ProcessSignTraffic` | [ProcessSignTrafficRequest](#tss.v1beta1.ProcessSignTrafficRequest) | [ProcessSignTrafficResponse](#tss.v1beta1.ProcessSignTrafficResponse) |  | ||
| `VoteSig` | [VoteSigRequest](#tss.v1beta1.VoteSigRequest) | [VoteSigResponse](#tss.v1beta1.VoteSigResponse) |  | ||

 <!-- end services -->



<a name="vote/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## vote/v1beta1/genesis.proto



<a name="vote.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `voting_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  |  |





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

