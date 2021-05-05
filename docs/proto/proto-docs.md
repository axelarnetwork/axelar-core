<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [bitcoin/v1beta1/types.proto](#bitcoin/v1beta1/types.proto)
    - [Network](#bitcoin.v1beta1.Network)
    - [OutPointInfo](#bitcoin.v1beta1.OutPointInfo)
  
- [bitcoin/v1beta1/params.proto](#bitcoin/v1beta1/params.proto)
    - [Params](#bitcoin.v1beta1.Params)
  
- [bitcoin/v1beta1/genesis.proto](#bitcoin/v1beta1/genesis.proto)
    - [GenesisState](#bitcoin.v1beta1.GenesisState)
  
- [bitcoin/v1beta1/query.proto](#bitcoin/v1beta1/query.proto)
    - [DepositQueryParams](#bitcoin.v1beta1.DepositQueryParams)
  
- [vote/exported/v1beta1/types.proto](#vote/exported/v1beta1/types.proto)
    - [PollMeta](#vote.exported.v1beta1.PollMeta)
  
- [bitcoin/v1beta1/tx.proto](#bitcoin/v1beta1/tx.proto)
    - [MsgConfirmOutpoint](#bitcoin.v1beta1.MsgConfirmOutpoint)
    - [MsgLink](#bitcoin.v1beta1.MsgLink)
    - [MsgSignPendingTransfers](#bitcoin.v1beta1.MsgSignPendingTransfers)
    - [MsgVoteConfirmOutpoint](#bitcoin.v1beta1.MsgVoteConfirmOutpoint)
  
- [broadcast/v1beta1/genesis.proto](#broadcast/v1beta1/genesis.proto)
    - [GenesisState](#broadcast.v1beta1.GenesisState)
  
- [broadcast/v1beta1/tx.proto](#broadcast/v1beta1/tx.proto)
    - [MsgRegisterProxy](#broadcast.v1beta1.MsgRegisterProxy)
  
- [ethereum/v1beta1/params.proto](#ethereum/v1beta1/params.proto)
    - [Params](#ethereum.v1beta1.Params)
  
- [ethereum/v1beta1/genesis.proto](#ethereum/v1beta1/genesis.proto)
    - [GenesisState](#ethereum.v1beta1.GenesisState)
  
- [ethereum/v1beta1/tx.proto](#ethereum/v1beta1/tx.proto)
    - [MsgConfirmDeposit](#ethereum.v1beta1.MsgConfirmDeposit)
    - [MsgConfirmToken](#ethereum.v1beta1.MsgConfirmToken)
    - [MsgLink](#ethereum.v1beta1.MsgLink)
    - [MsgSignBurnTokens](#ethereum.v1beta1.MsgSignBurnTokens)
    - [MsgSignDeployToken](#ethereum.v1beta1.MsgSignDeployToken)
    - [MsgSignPendingTransfers](#ethereum.v1beta1.MsgSignPendingTransfers)
    - [MsgSignTransferOwnership](#ethereum.v1beta1.MsgSignTransferOwnership)
    - [MsgSignTx](#ethereum.v1beta1.MsgSignTx)
    - [MsgVoteConfirmDeposit](#ethereum.v1beta1.MsgVoteConfirmDeposit)
    - [MsgVoteConfirmToken](#ethereum.v1beta1.MsgVoteConfirmToken)
  
- [nexus/exported/v1beta1/types.proto](#nexus/exported/v1beta1/types.proto)
    - [Chain](#nexus.exported.v1beta1.Chain)
  
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
    - [MsgAssignNextKey](#tss.v1beta1.MsgAssignNextKey)
    - [MsgDeregister](#tss.v1beta1.MsgDeregister)
    - [MsgKeygenStart](#tss.v1beta1.MsgKeygenStart)
    - [MsgKeygenTraffic](#tss.v1beta1.MsgKeygenTraffic)
    - [MsgRotateKey](#tss.v1beta1.MsgRotateKey)
    - [MsgSignTraffic](#tss.v1beta1.MsgSignTraffic)
    - [MsgVotePubKey](#tss.v1beta1.MsgVotePubKey)
    - [MsgVoteSig](#tss.v1beta1.MsgVoteSig)
  
- [vote/v1beta1/genesis.proto](#vote/v1beta1/genesis.proto)
    - [GenesisState](#vote.v1beta1.GenesisState)
  
- [Scalar Value Types](#scalar-value-types)



<a name="bitcoin/v1beta1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## bitcoin/v1beta1/types.proto



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



<a name="bitcoin.v1beta1.MsgConfirmOutpoint"></a>

### MsgConfirmOutpoint
MsgConfirmOutpoint represents a message to trigger the confirmation of a
Bitcoin outpoint


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `out_point_info` | [OutPointInfo](#bitcoin.v1beta1.OutPointInfo) |  |  |






<a name="bitcoin.v1beta1.MsgLink"></a>

### MsgLink
MsgLink represents a message to link a cross-chain address to a Bitcoin
address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `recipient_addr` | [string](#string) |  |  |
| `recipient_chain` | [string](#string) |  |  |






<a name="bitcoin.v1beta1.MsgSignPendingTransfers"></a>

### MsgSignPendingTransfers
MsgSignPendingTransfers represents a message to trigger the signing of a
consolidation transaction


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `fee` | [int64](#int64) |  |  |






<a name="bitcoin.v1beta1.MsgVoteConfirmOutpoint"></a>

### MsgVoteConfirmOutpoint
MsgVoteConfirmOutpoint represents a message to that votes on an outpoint


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll` | [vote.exported.v1beta1.PollMeta](#vote.exported.v1beta1.PollMeta) |  |  |
| `out_point` | [string](#string) |  |  |
| `confirmed` | [bool](#bool) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

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



<a name="broadcast.v1beta1.MsgRegisterProxy"></a>

### MsgRegisterProxy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `principal_addr` | [bytes](#bytes) |  |  |
| `proxy_addr` | [bytes](#bytes) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

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



<a name="ethereum.v1beta1.MsgConfirmDeposit"></a>

### MsgConfirmDeposit
MsgConfirmDeposit represents an erc20 deposit confirmation message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `tx_id` | [string](#string) |  |  |
| `amount` | [bytes](#bytes) |  |  |
| `burner_addr` | [string](#string) |  |  |






<a name="ethereum.v1beta1.MsgConfirmToken"></a>

### MsgConfirmToken
MsgConfirmToken represents a token deploy confirmation message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `tx_id` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |






<a name="ethereum.v1beta1.MsgLink"></a>

### MsgLink
MsgLink represents the message that links a cross chain address to a burner
address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `recipient_addr` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `recipient_chain` | [string](#string) |  |  |






<a name="ethereum.v1beta1.MsgSignBurnTokens"></a>

### MsgSignBurnTokens
MsgSignBurnTokens represents the message to sign commands to burn tokens with
AxelarGateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |






<a name="ethereum.v1beta1.MsgSignDeployToken"></a>

### MsgSignDeployToken
MsgSignDeployToken represents the message to sign a deploy token command for
AxelarGateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `capacity` | [bytes](#bytes) |  |  |
| `decimals` | [uint32](#uint32) |  |  |
| `symbol` | [string](#string) |  |  |
| `token_name` | [string](#string) |  |  |






<a name="ethereum.v1beta1.MsgSignPendingTransfers"></a>

### MsgSignPendingTransfers
MsgSignPendingTransfers represents a message to trigger the signing of all
pending transfers


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |






<a name="ethereum.v1beta1.MsgSignTransferOwnership"></a>

### MsgSignTransferOwnership
MsgSignDeployToken represents the message to sign a deploy token command for
AxelarGateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `new_owner` | [string](#string) |  |  |






<a name="ethereum.v1beta1.MsgSignTx"></a>

### MsgSignTx



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `tx` | [bytes](#bytes) |  | Tx is stored in serialized form because the amino codec cannot properly deserialize MsgSignTx otherwise |






<a name="ethereum.v1beta1.MsgVoteConfirmDeposit"></a>

### MsgVoteConfirmDeposit
MsgVoteConfirmDeposit represents a message that votes on a deposit


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll` | [vote.exported.v1beta1.PollMeta](#vote.exported.v1beta1.PollMeta) |  |  |
| `tx_id` | [string](#string) |  |  |
| `burn_addr` | [string](#string) |  |  |
| `confirmed` | [bool](#bool) |  |  |






<a name="ethereum.v1beta1.MsgVoteConfirmToken"></a>

### MsgVoteConfirmToken
MsgVoteConfirmToken represents a message that votes on a token deploy


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll` | [vote.exported.v1beta1.PollMeta](#vote.exported.v1beta1.PollMeta) |  |  |
| `tx_id` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `confirmed` | [bool](#bool) |  |  |





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
| `native_asset` | [string](#string) |  |  |
| `supports_foreign_assets` | [bool](#bool) |  |  |





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
| `locking_period` | [int64](#int64) |  | Deprecated |
| `min_keygen_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  | MinKeygenThreshold defines the minimum % of stake that must be online to authorize generation of a new key in the system. |
| `corruption_threshold` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  | CorruptionThreshold defines the corruption threshold with which we'll run keygen protocol. |
| `key_requirements` | [tss.exported.v1beta1.KeyRequirement](#tss.exported.v1beta1.KeyRequirement) | repeated | KeyRequirements defines the requirement of each key for each chain |
| `min_bond_fraction_per_share` | [utils.v1beta1.Threshold](#utils.v1beta1.Threshold) |  | MinBondFractionPerShare defines the % of stake validators have to bond per key share |





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



<a name="tss.v1beta1.MsgAssignNextKey"></a>

### MsgAssignNextKey
MsgAssignNextKey represents a message to assign a new key


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `key_id` | [string](#string) |  |  |
| `key_role` | [tss.exported.v1beta1.KeyRole](#tss.exported.v1beta1.KeyRole) |  |  |






<a name="tss.v1beta1.MsgDeregister"></a>

### MsgDeregister
MsgDeregister to deregister so that the validator will not participate in any
future keygen


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |






<a name="tss.v1beta1.MsgKeygenStart"></a>

### MsgKeygenStart
MsgKeygenStart indicate the start of keygen


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `new_key_id` | [string](#string) |  |  |
| `subset_size` | [int64](#int64) |  |  |
| `key_share_distribution_policy` | [tss.exported.v1beta1.KeyShareDistributionPolicy](#tss.exported.v1beta1.KeyShareDistributionPolicy) |  |  |






<a name="tss.v1beta1.MsgKeygenTraffic"></a>

### MsgKeygenTraffic
MsgKeygenTraffic protocol message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `session_id` | [string](#string) |  |  |
| `payload` | [tss.tofnd.v1beta1.TrafficOut](#tss.tofnd.v1beta1.TrafficOut) |  |  |






<a name="tss.v1beta1.MsgRotateKey"></a>

### MsgRotateKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `chain` | [string](#string) |  |  |
| `subset_size` | [int64](#int64) |  |  |
| `key_role` | [tss.exported.v1beta1.KeyRole](#tss.exported.v1beta1.KeyRole) |  |  |






<a name="tss.v1beta1.MsgSignTraffic"></a>

### MsgSignTraffic
MsgSignTraffic protocol message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `session_id` | [string](#string) |  |  |
| `payload` | [tss.tofnd.v1beta1.TrafficOut](#tss.tofnd.v1beta1.TrafficOut) |  |  |






<a name="tss.v1beta1.MsgVotePubKey"></a>

### MsgVotePubKey
MsgVotePubKey represents the message to vote on a public key


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll_meta` | [vote.exported.v1beta1.PollMeta](#vote.exported.v1beta1.PollMeta) |  |  |
| `pub_key_bytes` | [bytes](#bytes) |  | need to vote on the bytes instead of ecdsa.PublicKey, otherwise we lose the elliptic curve information |






<a name="tss.v1beta1.MsgVoteSig"></a>

### MsgVoteSig
MsgVoteSig represents a message to vote for a signature


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [bytes](#bytes) |  |  |
| `poll_meta` | [vote.exported.v1beta1.PollMeta](#vote.exported.v1beta1.PollMeta) |  |  |
| `sig_bytes` | [bytes](#bytes) |  | need to vote on the bytes instead of r, s, because Go cannot deserialize private fields using reflection (so *big.Int does not work) |





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

