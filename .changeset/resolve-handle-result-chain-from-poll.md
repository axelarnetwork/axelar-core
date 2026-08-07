---
'@axelar-network/axelar-core': patch
---

Resolve the source chain in the EVM vote handler's `HandleResult` from the poll metadata instead of the voter-supplied result, and reject a result whose chain doesn't match the poll's. `vote.VoteHandler.HandleResult` now takes the poll rather than the result.
