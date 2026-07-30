---
'@axelar-network/axelar-core': patch
---

Resolve the completed-poll chain from the poll metadata instead of the voter-supplied vote result in the EVM vote handler, so a completed poll whose result names an unregistered chain can no longer permanently stall the x/vote EndBlocker.
