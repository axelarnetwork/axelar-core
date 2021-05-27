<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [bitcoin/v1beta1/types.proto](#bitcoin/v1beta1/types.proto)
    - [AddressInfo](#bitcoin.v1beta1.AddressInfo)
    - [Network](#bitcoin.v1beta1.Network)
    - [OutPointInfo](#bitcoin.v1beta1.OutPointInfo)
  
    - [AddressRole](#bitcoin.v1beta1.AddressRole)
    - [SignState](#bitcoin.v1beta1.SignState)
  
- [bitcoin/v1beta1/params.proto](#bitcoin/v1beta1/params.proto)
    - [Params](#bitcoin.v1beta1.Params)
  
- [bitcoin/v1beta1/genesis.proto](#bitcoin/v1beta1/genesis.proto)
    - [GenesisState](#bitcoin.v1beta1.GenesisState)
  
- [bitcoin/v1beta1/query.proto](#bitcoin/v1beta1/query.proto)
    - [DepositQueryParams](#bitcoin.v1beta1.DepositQueryParams)
    - [QueryRawTxResponse](#bitcoin.v1beta1.QueryRawTxResponse)
  
- [vote/exported/v1beta1/types.proto](#vote/exported/v1beta1/types.proto)
    - [PollMeta](#vote.exported.v1beta1.PollMeta)
  
- [bitcoin/v1beta1/tx.proto](#bitcoin/v1beta1/tx.proto)
    - [ConfirmOutpointRequest](#bitcoin.v1beta1.ConfirmOutpointRequest)
    - [ConfirmOutpointResponse](#bitcoin.v1beta1.ConfirmOutpointResponse)
    - [LinkRequest](#bitcoin.v1beta1.LinkRequest)
    - [LinkResponse](#bitcoin.v1beta1.LinkResponse)
    - [SignPendingTransfersRequest](#bitcoin.v1beta1.SignPendingTransfersRequest)
    - [SignPendingTransfersResponse](#bitcoin.v1beta1.SignPendingTransfersResponse)
    - [VoteConfirmOutpointRequest](#bitcoin.v1beta1.VoteConfirmOutpointRequest)
    - [VoteConfirmOutpointResponse](#bitcoin.v1beta1.VoteConfirmOutpointResponse)
  
- [bitcoin/v1beta1/service.proto](#bitcoin/v1beta1/service.proto)
    - [MsgService](#bitcoin.v1beta1.MsgService)
  
- [broadcast/v1beta1/genesis.proto](#broadcast/v1beta1/genesis.proto)
    - [GenesisState](#broadcast.v1beta1.GenesisState)
  
- [broadcast/v1beta1/tx.proto](#broadcast/v1beta1/tx.proto)
    - [RegisterProxyRequest](#broadcast.v1beta1.RegisterProxyRequest)
    - [RegisterProxyResponse](#broadcast.v1beta1.RegisterProxyResponse)
  
- [broadcast/v1beta1/service.proto](#broadcast/v1beta1/service.proto)
    - [MsgService](#broadcast.v1beta1.MsgService)
  
- [ethereum/v1beta1/params.proto](#ethereum/v1beta1/params.proto)
    - [Params](#ethereum.v1beta1.Params)
  
- [ethereum/v1beta1/genesis.proto](#ethereum/v1beta1/genesis.proto)
    - [GenesisState](#ethereum.v1beta1.GenesisState)
  
- [ethereum/v1beta1/tx.proto](#ethereum/v1beta1/tx.proto)
    - [AddChainRequest](#ethereum.v1beta1.AddChainRequest)
    - [AddChainResponse](#ethereum.v1beta1.AddChainResponse)
    - [ConfirmDepositRequest](#ethereum.v1beta1.ConfirmDepositRequest)
    - [ConfirmDepositResponse](#ethereum.v1beta1.ConfirmDepositResponse)
    - [ConfirmTokenRequest](#ethereum.v1beta1.ConfirmTokenRequest)
    - [ConfirmTokenResponse](#ethereum.v1beta1.ConfirmTokenResponse)
    - [LinkRequest](#ethereum.v1beta1.LinkRequest)
    - [LinkResponse](#ethereum.v1beta1.LinkResponse)
    - [SignBurnTokensRequest](#ethereum.v1beta1.SignBurnTokensRequest)
    - [SignBurnTokensResponse](#ethereum.v1beta1.SignBurnTokensResponse)
    - [SignDeployTokenRequest](#ethereum.v1beta1.SignDeployTokenRequest)
    - [SignDeployTokenResponse](#ethereum.v1beta1.SignDeployTokenResponse)
    - [SignPendingTransfersRequest](#ethereum.v1beta1.SignPendingTransfersRequest)
    - [SignPendingTransfersResponse](#ethereum.v1beta1.SignPendingTransfersResponse)
    - [SignTransferOwnershipRequest](#ethereum.v1beta1.SignTransferOwnershipRequest)
    - [SignTransferOwnershipResponse](#ethereum.v1beta1.SignTransferOwnershipResponse)
    - [SignTxRequest](#ethereum.v1beta1.SignTxRequest)
    - [SignTxResponse](#ethereum.v1beta1.SignTxResponse)
    - [VoteConfirmDepositRequest](#ethereum.v1beta1.VoteConfirmDepositRequest)
    - [VoteConfirmDepositResponse](#ethereum.v1beta1.VoteConfirmDepositResponse)
    - [VoteConfirmTokenRequest](#ethereum.v1beta1.VoteConfirmTokenRequest)
    - [VoteConfirmTokenResponse](#ethereum.v1beta1.VoteConfirmTokenResponse)
  
- [ethereum/v1beta1/service.proto](#ethereum/v1beta1/service.proto)
    - [MsgService](#ethereum.v1beta1.MsgService)
  
- [ethereum/v1beta1/types.proto](#ethereum/v1beta1/types.proto)
    - [BurnerInfo](#ethereum.v1beta1.BurnerInfo)
    - [ERC20Deposit](#ethereum.v1beta1.ERC20Deposit)
    - [ERC20TokenDeployment](#ethereum.v1beta1.ERC20TokenDeployment)
  
- [nexus/exported/v1beta1/types.proto](#nexus/exported/v1beta1/types.proto)
    - [Chain](#nexus.exported.v1beta1.Chain)
    - [CrossChainAddress](#nexus.exported.v1beta1.CrossChainAddress)
    - [CrossChainTransfer](#nexus.exported.v1beta1.CrossChainTransfer)
  
- [nexus/v1beta1/params.proto](#nexus/v1beta1/params.proto)
    - [Params](#nexus.v1beta1.Params)
  
- [nexus/v1beta1/genesis.proto](#nexus/v1beta1/genesis.proto)
    - [GenesisState](#nexus.v1beta1.GenesisState)
  
- [snapshot/v1beta1/params.proto](#snapshot/v1beta1/params.proto)
    - [Params](#snapshot.v1beta1.Params)
  
- [snapshot/v1beta1/genesis.proto](#snapshot/v1beta1/genesis.proto)
    - [GenesisState](#snapshot.v1beta1.GenesisState)
  
- [tss/exported/v1beta1/types.proto](#tss/exported/v1beta1/types.proto)
    - [KeyRequirement](#tss.exported.v1beta1.KeyRequirement)
  
    - [KeyRole](#tss.exported.v1beta1.KeyRole)
    - [KeyShareDistributionPolicy](#tss.exported.v1beta1.KeyShareDistributionPolicy)
  
- [tss/tofnd/v1beta1/tofnd.proto](#tss/tofnd/v1beta1/tofnd.proto)
    - [KeygenInit](#tss.tofnd.v1beta1.KeygenInit)
    - [MessageIn](#tss.tofnd.v1beta1.MessageIn)
    - [MessageOut](#tss.tofnd.v1beta1.MessageOut)
    - [MessageOut.CriminalList](#tss.tofnd.v1beta1.MessageOut.CriminalList)
    - [MessageOut.CriminalList.Criminal](#tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal)
    - [MessageOut.SignResult](#tss.tofnd.v1beta1.MessageOut.SignResult)
    - [SignInit](#tss.tofnd.v1beta1.SignInit)
    - [TrafficIn](#tss.tofnd.v1beta1.TrafficIn)
    - [TrafficOut](#tss.tofnd.v1beta1.TrafficOut)
  
    - [MessageOut.CriminalList.Criminal.CrimeType](#tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal.CrimeType)
  
- [utils/v1beta1/threshold.proto](#utils/v1beta1/threshold.proto)
    - [Threshold](#utils.v1beta1.Threshold)
  
- [tss/v1beta1/params.proto](#tss/v1beta1/params.proto)
    - [Params](#tss.v1beta1.Params)
  
- [tss/v1beta1/genesis.proto](#tss/v1beta1/genesis.proto)
    - [GenesisState](#tss.v1beta1.GenesisState)
  
- [tss/v1beta1/tx.proto](#tss/v1beta1/tx.proto)
    - [AssignKeyRequest](#tss.v1beta1.AssignKeyRequest)
    - [AssignKeyResponse](#tss.v1beta1.AssignKeyResponse)
    - [DeregisterRequest](#tss.v1beta1.DeregisterRequest)
    - [DeregisterResponse](#tss.v1beta1.DeregisterResponse)
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
  
- [Scalar Value Types](#scalar-value-types)



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





 <!-- end messages -->


<a name="bitcoin.v1beta1.AddressRole"></a>

### AddressRole


| Name | Number | Description |
| ---- | ------ | ----------- |
| ADDRESS_ROLE_UNSPECIFIED | 0 |  |
| ADDRESS_ROLE_DEPOSIT | 1 |  |
| ADDRESS_ROLE_CONSOLIDATION | 2 |  |



<a name="bitcoin.v1beta1.SignState"></a>

### SignState


| Name | Number | Description |
| ---- | ------ | ----------- |
| SIGN_STATE_UNSPECIFIED | 0 |  |
| SIGN_STATE_SIGNING_PENDING_TRANSFERS | 1 |  |
| SIGN_STATE_SIGNED_NOT_CONFIRMED | 2 |  |
| SIGN_STATE_READY_TO_SIGN | 3 |  |


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
| `minimum_withdrawal_amount` | [int64](#int64) |  |  |





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






<a name="bitcoin.v1beta1.QueryRawTxResponse"></a>

### QueryRawTxResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `raw_tx` | [string](#string) |  |  |
| `state` | [SignState](#bitcoin.v1beta1.SignState) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="vote/exported/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## vote/exported/v1beta1/types.proto



<a name="vote.exported.v1beta1.PollMeta"></a>

### PollMeta
PollMeta represents the meta data for a poll


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |
| `nonce` | [int64](#int64) |  |  |





 <!-- end messages -->

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






<a name="bitcoin.v1beta1.SignPendingTransfersRequest"></a>

### SignPendingTransfersRequest
MsgSignPendingTransfers represents a message to trigger the signing of a
consolidation transaction


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `fee` | [int64](#int64) |  | **Deprecated.** TODO: Remove once c2d2 is ready to perform child-pay-for-parent for consolidation transactions |






<a name="bitcoin.v1beta1.SignPendingTransfersResponse"></a>

### SignPendingTransfersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `command_id` | [string](#string) |  |  |






<a name="bitcoin.v1beta1.VoteConfirmOutpointRequest"></a>

### VoteConfirmOutpointRequest
MsgVoteConfirmOutpoint represents a message to that votes on an outpoint


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll` | [vote.exported.v1beta1.PollMeta](#vote.exported.v1beta1.PollMeta) |  |  |
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
| `SignPendingTransfers` | [SignPendingTransfersRequest](#bitcoin.v1beta1.SignPendingTransfersRequest) | [SignPendingTransfersResponse](#bitcoin.v1beta1.SignPendingTransfersResponse) |  | POST|/axelar/bitcoin/sign|

 <!-- end services -->



<a name="broadcast/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## broadcast/v1beta1/genesis.proto



<a name="broadcast.v1beta1.GenesisState"></a>

### GenesisState






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="broadcast/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## broadcast/v1beta1/tx.proto



<a name="broadcast.v1beta1.RegisterProxyRequest"></a>

### RegisterProxyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `principal_addr` | [bytes](#bytes) |  |  |
| `proxy_addr` | [bytes](#bytes) |  |  |






<a name="broadcast.v1beta1.RegisterProxyResponse"></a>

### RegisterProxyResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="broadcast/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## broadcast/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="broadcast.v1beta1.MsgService"></a>

### MsgService
Msg defines the broadcast Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterProxy` | [RegisterProxyRequest](#broadcast.v1beta1.RegisterProxyRequest) | [RegisterProxyResponse](#broadcast.v1beta1.RegisterProxyResponse) | RegisterProxy defines a method for registering a proxy account that can act in a validator account's stead. | POST|/axelar/broadcast/registerProxy/{proxy_addr}|

 <!-- end services -->



<a name="ethereum/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ethereum/v1beta1/params.proto



<a name="ethereum.v1beta1.Params"></a>

### Params
Params is the parameter set for this module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `confirmation_height` | [uint64](#uint64) |  |  |
| `network` | [string](#string) |  |  |
| `gateway` | [bytes](#bytes) |  |  |
| `token` | [bytes](#bytes) |  |  |
| `burnable` | [bytes](#bytes) |  |  |
| `revote_locking_period` | [int64](#int64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="ethereum/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ethereum/v1beta1/genesis.proto



<a name="ethereum.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#ethereum.v1beta1.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="ethereum/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ethereum/v1beta1/tx.proto



<a name="ethereum.v1beta1.AddChainRequest"></a>

### AddChainRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `name` | [string](#string) |  |  |
| `native_asset` | [string](#string) |  |  |
| `supports_foreign` | [bool](#bool) |  |  |






<a name="ethereum.v1beta1.AddChainResponse"></a>

### AddChainResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |






<a name="ethereum.v1beta1.ConfirmDepositRequest"></a>

### ConfirmDepositRequest
MsgConfirmDeposit represents an erc20 deposit confirmation message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `amount` | [bytes](#bytes) |  |  |
| `burner_address` | [bytes](#bytes) |  |  |






<a name="ethereum.v1beta1.ConfirmDepositResponse"></a>

### ConfirmDepositResponse







<a name="ethereum.v1beta1.ConfirmTokenRequest"></a>

### ConfirmTokenRequest
MsgConfirmToken represents a token deploy confirmation message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `symbol` | [string](#string) |  |  |






<a name="ethereum.v1beta1.ConfirmTokenResponse"></a>

### ConfirmTokenResponse







<a name="ethereum.v1beta1.LinkRequest"></a>

### LinkRequest
MsgLink represents the message that links a cross chain address to a burner
address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `recipient_addr` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `recipient_chain` | [string](#string) |  |  |






<a name="ethereum.v1beta1.LinkResponse"></a>

### LinkResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposit_addr` | [string](#string) |  |  |






<a name="ethereum.v1beta1.SignBurnTokensRequest"></a>

### SignBurnTokensRequest
MsgSignBurnTokens represents the message to sign commands to burn tokens with
AxelarGateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |






<a name="ethereum.v1beta1.SignBurnTokensResponse"></a>

### SignBurnTokensResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `command_id` | [bytes](#bytes) |  |  |






<a name="ethereum.v1beta1.SignDeployTokenRequest"></a>

### SignDeployTokenRequest
MsgSignDeployToken represents the message to sign a deploy token command for
AxelarGateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `capacity` | [bytes](#bytes) |  |  |
| `decimals` | [uint32](#uint32) |  |  |
| `symbol` | [string](#string) |  |  |
| `token_name` | [string](#string) |  |  |






<a name="ethereum.v1beta1.SignDeployTokenResponse"></a>

### SignDeployTokenResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `command_id` | [bytes](#bytes) |  |  |






<a name="ethereum.v1beta1.SignPendingTransfersRequest"></a>

### SignPendingTransfersRequest
MsgSignPendingTransfers represents a message to trigger the signing of all
pending transfers


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |






<a name="ethereum.v1beta1.SignPendingTransfersResponse"></a>

### SignPendingTransfersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `command_id` | [bytes](#bytes) |  |  |






<a name="ethereum.v1beta1.SignTransferOwnershipRequest"></a>

### SignTransferOwnershipRequest
MsgSignDeployToken represents the message to sign a deploy token command for
AxelarGateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `new_owner` | [bytes](#bytes) |  |  |






<a name="ethereum.v1beta1.SignTransferOwnershipResponse"></a>

### SignTransferOwnershipResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `command_id` | [bytes](#bytes) |  |  |






<a name="ethereum.v1beta1.SignTxRequest"></a>

### SignTxRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `tx` | [bytes](#bytes) |  | Tx is stored in serialized form because the amino codec cannot properly deserialize MsgSignTx otherwise |






<a name="ethereum.v1beta1.SignTxResponse"></a>

### SignTxResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [string](#string) |  |  |






<a name="ethereum.v1beta1.VoteConfirmDepositRequest"></a>

### VoteConfirmDepositRequest
MsgVoteConfirmDeposit represents a message that votes on a deposit


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll` | [vote.exported.v1beta1.PollMeta](#vote.exported.v1beta1.PollMeta) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `burn_address` | [bytes](#bytes) |  |  |
| `confirmed` | [bool](#bool) |  |  |






<a name="ethereum.v1beta1.VoteConfirmDepositResponse"></a>

### VoteConfirmDepositResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |






<a name="ethereum.v1beta1.VoteConfirmTokenRequest"></a>

### VoteConfirmTokenRequest
MsgVoteConfirmToken represents a message that votes on a token deploy


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll` | [vote.exported.v1beta1.PollMeta](#vote.exported.v1beta1.PollMeta) |  |  |
| `tx_id` | [bytes](#bytes) |  |  |
| `symbol` | [string](#string) |  |  |
| `confirmed` | [bool](#bool) |  |  |






<a name="ethereum.v1beta1.VoteConfirmTokenResponse"></a>

### VoteConfirmTokenResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `log` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="ethereum/v1beta1/service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ethereum/v1beta1/service.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="ethereum.v1beta1.MsgService"></a>

### MsgService
Msg defines the ethereum Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Link` | [LinkRequest](#ethereum.v1beta1.LinkRequest) | [LinkResponse](#ethereum.v1beta1.LinkResponse) |  | POST|/axelar/ethereum/link/{recipient_chain}|
| `ConfirmToken` | [ConfirmTokenRequest](#ethereum.v1beta1.ConfirmTokenRequest) | [ConfirmTokenResponse](#ethereum.v1beta1.ConfirmTokenResponse) |  | POST|/axelar/ethereum/confirm-erc20-deploy/{symbol}|
| `ConfirmDeposit` | [ConfirmDepositRequest](#ethereum.v1beta1.ConfirmDepositRequest) | [ConfirmDepositResponse](#ethereum.v1beta1.ConfirmDepositResponse) |  | POST|/axelar/ethereum/confirm-erc20-deposit|
| `VoteConfirmDeposit` | [VoteConfirmDepositRequest](#ethereum.v1beta1.VoteConfirmDepositRequest) | [VoteConfirmDepositResponse](#ethereum.v1beta1.VoteConfirmDepositResponse) |  | ||
| `VoteConfirmToken` | [VoteConfirmTokenRequest](#ethereum.v1beta1.VoteConfirmTokenRequest) | [VoteConfirmTokenResponse](#ethereum.v1beta1.VoteConfirmTokenResponse) |  | ||
| `SignDeployToken` | [SignDeployTokenRequest](#ethereum.v1beta1.SignDeployTokenRequest) | [SignDeployTokenResponse](#ethereum.v1beta1.SignDeployTokenResponse) |  | POST|/axelar/ethereum/sign-deploy-token/{symbol}|
| `SignBurnTokens` | [SignBurnTokensRequest](#ethereum.v1beta1.SignBurnTokensRequest) | [SignBurnTokensResponse](#ethereum.v1beta1.SignBurnTokensResponse) |  | POST|/axelar/ethereum/sign-burn|
| `SignTx` | [SignTxRequest](#ethereum.v1beta1.SignTxRequest) | [SignTxResponse](#ethereum.v1beta1.SignTxResponse) |  | POST|/axelar/ethereum/sign-tx|
| `SignPendingTransfers` | [SignPendingTransfersRequest](#ethereum.v1beta1.SignPendingTransfersRequest) | [SignPendingTransfersResponse](#ethereum.v1beta1.SignPendingTransfersResponse) |  | POST|/axelar/ethereum/sign-pending|
| `SignTransferOwnership` | [SignTransferOwnershipRequest](#ethereum.v1beta1.SignTransferOwnershipRequest) | [SignTransferOwnershipResponse](#ethereum.v1beta1.SignTransferOwnershipResponse) |  | ||
| `AddChain` | [AddChainRequest](#ethereum.v1beta1.AddChainRequest) | [AddChainResponse](#ethereum.v1beta1.AddChainResponse) |  | POST|/axelar/ethereum/add-chain|

 <!-- end services -->



<a name="ethereum/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ethereum/v1beta1/types.proto



<a name="ethereum.v1beta1.BurnerInfo"></a>

### BurnerInfo
BurnerInfo describes information required to burn token at an burner address
that is deposited by an user


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `token_address` | [bytes](#bytes) |  |  |
| `symbol` | [string](#string) |  |  |
| `salt` | [bytes](#bytes) |  |  |






<a name="ethereum.v1beta1.ERC20Deposit"></a>

### ERC20Deposit
ERC20Deposit contains information for an ERC20 deposit


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [bytes](#bytes) |  |  |
| `amount` | [bytes](#bytes) |  |  |
| `symbol` | [string](#string) |  |  |
| `burner_address` | [bytes](#bytes) |  |  |






<a name="ethereum.v1beta1.ERC20TokenDeployment"></a>

### ERC20TokenDeployment
ERC20TokenDeployment describes information about an ERC20 token


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `symbol` | [string](#string) |  |  |
| `token_address` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

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
| `platform` | [string](#string) |  |  |
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
| `keygen_result` | [bytes](#bytes) |  | final message only, Keygen |
| `sign_result` | [MessageOut.SignResult](#tss.tofnd.v1beta1.MessageOut.SignResult) |  | final message only, Sign |






<a name="tss.tofnd.v1beta1.MessageOut.CriminalList"></a>

### MessageOut.CriminalList



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `criminals` | [MessageOut.CriminalList.Criminal](#tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal) | repeated |  |






<a name="tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal"></a>

### MessageOut.CriminalList.Criminal



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `party_uid` | [string](#string) |  |  |
| `crime_type` | [MessageOut.CriminalList.Criminal.CrimeType](#tss.tofnd.v1beta1.MessageOut.CriminalList.Criminal.CrimeType) |  |  |






<a name="tss.tofnd.v1beta1.MessageOut.SignResult"></a>

### MessageOut.SignResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signature` | [bytes](#bytes) |  |  |
| `criminals` | [MessageOut.CriminalList](#tss.tofnd.v1beta1.MessageOut.CriminalList) |  |  |






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



<a name="tss/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## tss/v1beta1/tx.proto



<a name="tss.v1beta1.AssignKeyRequest"></a>

### AssignKeyRequest
AssignKeyRequest represents a message to assign a new key


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |
| `key_role` | [tss.exported.v1beta1.KeyRole](#tss.exported.v1beta1.KeyRole) |  |  |






<a name="tss.v1beta1.AssignKeyResponse"></a>

### AssignKeyResponse







<a name="tss.v1beta1.DeregisterRequest"></a>

### DeregisterRequest
DeregisterRequest to deregister so that the validator will not participate in
any future keygen


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |






<a name="tss.v1beta1.DeregisterResponse"></a>

### DeregisterResponse







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
| `subset_size` | [int64](#int64) |  |  |
| `key_role` | [tss.exported.v1beta1.KeyRole](#tss.exported.v1beta1.KeyRole) |  |  |






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
| `poll_meta` | [vote.exported.v1beta1.PollMeta](#vote.exported.v1beta1.PollMeta) |  |  |
| `pub_key_bytes` | [bytes](#bytes) |  | need to vote on the bytes instead of ecdsa.PublicKey, otherwise we lose the elliptic curve information |






<a name="tss.v1beta1.VotePubKeyResponse"></a>

### VotePubKeyResponse







<a name="tss.v1beta1.VoteSigRequest"></a>

### VoteSigRequest
VoteSigRequest represents a message to vote for a signature


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll_meta` | [vote.exported.v1beta1.PollMeta](#vote.exported.v1beta1.PollMeta) |  |  |
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
| `AssignKey` | [AssignKeyRequest](#tss.v1beta1.AssignKeyRequest) | [AssignKeyResponse](#tss.v1beta1.AssignKeyResponse) |  | POST|/axelar/tss/assign/{chain}|
| `RotateKey` | [RotateKeyRequest](#tss.v1beta1.RotateKeyRequest) | [RotateKeyResponse](#tss.v1beta1.RotateKeyResponse) |  | POST|/axelar/tss/assign/{chain}|
| `VotePubKey` | [VotePubKeyRequest](#tss.v1beta1.VotePubKeyRequest) | [VotePubKeyResponse](#tss.v1beta1.VotePubKeyResponse) |  | ||
| `ProcessSignTraffic` | [ProcessSignTrafficRequest](#tss.v1beta1.ProcessSignTrafficRequest) | [ProcessSignTrafficResponse](#tss.v1beta1.ProcessSignTrafficResponse) |  | ||
| `VoteSig` | [VoteSigRequest](#tss.v1beta1.VoteSigRequest) | [VoteSigResponse](#tss.v1beta1.VoteSigResponse) |  | ||
| `Deregister` | [DeregisterRequest](#tss.v1beta1.DeregisterRequest) | [DeregisterResponse](#tss.v1beta1.DeregisterResponse) |  | ||

 <!-- end services -->



<a name="vote/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## vote/v1beta1/genesis.proto



<a name="vote.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `voting_interval` | [int64](#int64) |  |  |
| `voting_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  |  |





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

